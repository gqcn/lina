<script setup lang="ts">
import { computed } from 'vue';
import { useRoute } from 'vue-router';

import { Result as AResult } from 'ant-design-vue';

import { getPluginPageByRoute } from '#/plugins/page-registry';

const route = useRoute();
const currentRoutePath = computed(() => route.path.replace(/^\//, ''));
const pageEntry = computed(() => getPluginPageByRoute(currentRoutePath.value));
</script>

<template>
  <component :is="pageEntry.component" v-if="pageEntry" />
  <a-result
    v-else
    status="404"
    title="插件页面未找到"
    sub-title="当前路由没有匹配到已注册的源码插件前端页面。"
  />
</template>
