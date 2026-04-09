<script setup lang="ts">
import type { RouteLocationNormalizedLoaded } from 'vue-router';

import { computed, onBeforeUnmount, ref, shallowRef, watch } from 'vue';
import { useRoute } from 'vue-router';

import { Result as AResult, Spin as ASpin } from 'ant-design-vue';

import { getPluginPageByRoute } from '#/plugins/page-registry';

const runtimeEmbeddedMountMode = 'embedded-mount';
const runtimeEmbeddedHostTestId = 'plugin-runtime-embedded-host';
const runtimeEmbeddedSourceQueryKey = 'embeddedSrc';
const runtimeEmbeddedAccessModeQueryKey = 'pluginAccessMode';

type RuntimeEmbeddedRouteQuery = Record<string, string>;

type RuntimeEmbeddedMountContext = {
  assetURL: string;
  baseURL: string;
  container: HTMLElement;
  query: RuntimeEmbeddedRouteQuery;
  route: RouteLocationNormalizedLoaded;
  routePath: string;
  title: string;
};

type RuntimeEmbeddedMountInstance = {
  unmount?: (context: RuntimeEmbeddedMountContext) => Promise<void> | void;
  update?: (context: RuntimeEmbeddedMountContext) => Promise<void> | void;
};

type RuntimeEmbeddedMountResult =
  | RuntimeEmbeddedMountInstance
  | ((context: RuntimeEmbeddedMountContext) => Promise<void> | void)
  | null
  | undefined;

type RuntimeEmbeddedMountFunction = (
  context: RuntimeEmbeddedMountContext,
) => Promise<RuntimeEmbeddedMountResult> | RuntimeEmbeddedMountResult;

type RuntimeEmbeddedModule = {
  mount?: RuntimeEmbeddedMountFunction;
  unmount?: (context: RuntimeEmbeddedMountContext) => Promise<void> | void;
  update?: (context: RuntimeEmbeddedMountContext) => Promise<void> | void;
};

type MountedRuntimeEmbeddedModule = {
  context: RuntimeEmbeddedMountContext;
  instance: null | RuntimeEmbeddedMountInstance;
  module: RuntimeEmbeddedModule;
};

const route = useRoute();
const currentRoutePath = computed(() => route.path.replace(/^\//, ''));
const pageEntry = computed(() => getPluginPageByRoute(currentRoutePath.value));
const runtimeEmbeddedHost = ref<HTMLElement>();
const runtimeEmbeddedLoading = ref(false);
const runtimeEmbeddedError = ref('');
const mountedRuntimeEmbeddedModule =
  shallowRef<MountedRuntimeEmbeddedModule | null>(null);

let runtimeEmbeddedMountToken = 0;

const normalizedRouteQuery = computed<RuntimeEmbeddedRouteQuery>(() => {
  const mergedQuery = {
    ...((route.meta.query ?? {}) as Record<string, unknown>),
    ...(route.query as Record<string, unknown>),
  };

  const query: RuntimeEmbeddedRouteQuery = {};
  for (const [key, value] of Object.entries(mergedQuery)) {
    if (Array.isArray(value)) {
      const firstValue = value.at(0);
      if (firstValue != null) {
        query[key] = String(firstValue);
      }
      continue;
    }
    if (value != null) {
      query[key] = String(value);
    }
  }
  return query;
});

const runtimeEmbeddedSource = computed(() => {
  return (
    normalizedRouteQuery.value[runtimeEmbeddedSourceQueryKey]?.trim() ?? ''
  );
});

const isRuntimeEmbeddedMountMode = computed(() => {
  return (
    normalizedRouteQuery.value[runtimeEmbeddedAccessModeQueryKey] ===
      runtimeEmbeddedMountMode && !!runtimeEmbeddedSource.value
  );
});

function toAbsoluteRuntimeEmbeddedAssetURL(source: string) {
  return new URL(source, window.location.origin).toString();
}

function normalizeRuntimeEmbeddedMountResult(
  result: RuntimeEmbeddedMountResult,
): null | RuntimeEmbeddedMountInstance {
  if (!result) {
    return null;
  }
  if (typeof result === 'function') {
    return {
      unmount: result,
    };
  }
  return result;
}

function resolveRuntimeEmbeddedModule(
  candidate: unknown,
): RuntimeEmbeddedModule {
  const moduleCandidate = candidate as Record<string, unknown> | undefined;
  const defaultExport =
    (moduleCandidate?.default as Record<string, unknown> | undefined) ?? {};
  const defaultMount =
    typeof moduleCandidate?.default === 'function'
      ? (moduleCandidate.default as RuntimeEmbeddedMountFunction)
      : (defaultExport.mount as RuntimeEmbeddedMountFunction | undefined);

  return {
    mount:
      (moduleCandidate?.mount as RuntimeEmbeddedMountFunction | undefined) ??
      defaultMount,
    unmount:
      (moduleCandidate?.unmount as RuntimeEmbeddedModule['unmount']) ??
      (defaultExport.unmount as RuntimeEmbeddedModule['unmount']),
    update:
      (moduleCandidate?.update as RuntimeEmbeddedModule['update']) ??
      (defaultExport.update as RuntimeEmbeddedModule['update']),
  };
}

function buildRuntimeEmbeddedMountContext(
  assetURL: string,
): RuntimeEmbeddedMountContext {
  const container = runtimeEmbeddedHost.value;
  if (!container) {
    throw new Error('Runtime embedded mount container is not ready.');
  }

  return {
    assetURL,
    baseURL: assetURL.slice(0, assetURL.lastIndexOf('/') + 1),
    container,
    query: normalizedRouteQuery.value,
    route,
    routePath: currentRoutePath.value,
    title: String(route.meta.title ?? currentRoutePath.value),
  };
}

async function cleanupMountedRuntimeEmbeddedModule() {
  const mounted = mountedRuntimeEmbeddedModule.value;
  mountedRuntimeEmbeddedModule.value = null;

  if (!mounted) {
    runtimeEmbeddedHost.value?.replaceChildren();
    return;
  }

  try {
    if (mounted.instance?.unmount) {
      await mounted.instance.unmount(mounted.context);
    } else if (mounted.module.unmount) {
      await mounted.module.unmount(mounted.context);
    }
  } finally {
    mounted.context.container.replaceChildren();
  }
}

async function mountRuntimeEmbeddedModule() {
  const hostElement = runtimeEmbeddedHost.value;
  runtimeEmbeddedMountToken += 1;
  const currentMountToken = runtimeEmbeddedMountToken;

  await cleanupMountedRuntimeEmbeddedModule();

  if (!hostElement || !isRuntimeEmbeddedMountMode.value) {
    runtimeEmbeddedLoading.value = false;
    runtimeEmbeddedError.value = '';
    return;
  }

  runtimeEmbeddedLoading.value = true;
  runtimeEmbeddedError.value = '';

  try {
    const assetURL = toAbsoluteRuntimeEmbeddedAssetURL(
      runtimeEmbeddedSource.value,
    );

    // Runtime embedded modules are delivered as hosted ESM assets. The host
    // imports them lazily so the plugin can use its own frontend stack while
    // still being mounted inside the Lina content container.
    const importedModule = await import(/* @vite-ignore */ assetURL);
    if (currentMountToken !== runtimeEmbeddedMountToken) {
      return;
    }

    const runtimeEmbeddedModule = resolveRuntimeEmbeddedModule(importedModule);
    if (!runtimeEmbeddedModule.mount) {
      throw new Error(
        'Runtime embedded entry must export a mount(context) function.',
      );
    }

    const mountContext = buildRuntimeEmbeddedMountContext(assetURL);
    const mountResult = await runtimeEmbeddedModule.mount(mountContext);
    if (currentMountToken !== runtimeEmbeddedMountToken) {
      return;
    }

    mountedRuntimeEmbeddedModule.value = {
      context: mountContext,
      instance: normalizeRuntimeEmbeddedMountResult(mountResult),
      module: runtimeEmbeddedModule,
    };
  } catch (error) {
    runtimeEmbeddedError.value =
      error instanceof Error
        ? error.message
        : 'Runtime embedded plugin mount failed.';
    runtimeEmbeddedHost.value?.replaceChildren();
  } finally {
    if (currentMountToken === runtimeEmbeddedMountToken) {
      runtimeEmbeddedLoading.value = false;
    }
  }
}

watch(
  [() => route.fullPath, isRuntimeEmbeddedMountMode, runtimeEmbeddedHost],
  async () => {
    if (pageEntry.value) {
      runtimeEmbeddedError.value = '';
      runtimeEmbeddedLoading.value = false;
      await cleanupMountedRuntimeEmbeddedModule();
      return;
    }
    await mountRuntimeEmbeddedModule();
  },
  { immediate: true },
);

onBeforeUnmount(() => {
  runtimeEmbeddedMountToken += 1;
  void cleanupMountedRuntimeEmbeddedModule();
});
</script>

<template>
  <component :is="pageEntry.component" v-if="pageEntry" />
  <section v-else-if="isRuntimeEmbeddedMountMode" class="runtime-embedded-page">
    <div class="runtime-embedded-page__body">
      <div
        :data-testid="runtimeEmbeddedHostTestId"
        class="runtime-embedded-page__host"
        ref="runtimeEmbeddedHost"
      />

      <div class="runtime-embedded-page__overlay" v-if="runtimeEmbeddedLoading">
        <a-spin size="large" />
      </div>

      <div
        class="runtime-embedded-page__overlay"
        v-else-if="runtimeEmbeddedError"
      >
        <a-result
          status="error"
          title="Runtime plugin mount failed"
          :sub-title="runtimeEmbeddedError"
        />
      </div>
    </div>
  </section>
  <a-result
    v-else
    status="404"
    title="插件页面未找到"
    sub-title="当前路由没有匹配到已注册的源码插件前端页面，也没有声明可用的 runtime 内嵌挂载入口。"
  />
</template>

<style scoped>
.runtime-embedded-page {
  height: 100%;
  min-height: 460px;
}

.runtime-embedded-page__body {
  position: relative;
  height: 100%;
  min-height: 460px;
  border-radius: 20px;
  background: transparent;
  overflow: hidden;
}

.runtime-embedded-page__host {
  height: 100%;
  min-height: 460px;
}

.runtime-embedded-page__overlay {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgb(255 255 255 / 88%);
}
</style>
