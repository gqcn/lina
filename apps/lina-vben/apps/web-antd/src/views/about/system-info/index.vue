<script setup lang="ts">
import { h, onMounted, ref } from 'vue';

import { Page } from '@vben/common-ui';

import { getSystemInfo } from '#/api/about';

import {
  BACKEND_COMPONENTS,
  FRONTEND_COMPONENTS,
  PROJECT_INFO,
} from '../config';

defineOptions({ name: 'SystemInfo' });

interface DescriptionItem {
  content: any;
  title: string;
}

const renderLink = (href: string, text: string) =>
  h(
    'a',
    { href, target: '_blank', class: 'vben-link' },
    { default: () => text },
  );

// 关于项目
const projectItems: DescriptionItem[] = [
  { title: '项目名称', content: PROJECT_INFO.name },
  { title: '项目描述', content: PROJECT_INFO.description },
  { title: '版本号', content: PROJECT_INFO.version },
  { title: '开源许可', content: PROJECT_INFO.license },
  {
    title: '项目主页',
    content: renderLink(PROJECT_INFO.homepage, '点击查看'),
  },
];

// 基本信息（后端 API 数据）
const runtimeItems = ref<DescriptionItem[]>([]);
const loading = ref(true);

onMounted(async () => {
  try {
    const info = await getSystemInfo();
    runtimeItems.value = [
      { title: 'Go 版本', content: info.goVersion },
      { title: 'GoFrame 版本', content: info.gfVersion },
      { title: '操作系统', content: `${info.os}/${info.arch}` },
      { title: '数据库版本', content: `MySQL ${info.dbVersion}` },
      { title: '启动时间', content: info.startTime },
      { title: '运行时长', content: info.runDuration },
    ];
  } finally {
    loading.value = false;
  }
});

// 后端组件
const backendItems: DescriptionItem[] = BACKEND_COMPONENTS.map((item) => ({
  title: item.name,
  content: h('div', [
    h('span', { class: 'text-foreground/80' }, item.version),
    h('span', { class: 'mx-2 text-foreground/30' }, '|'),
    renderLink(item.url, item.description),
  ]),
}));

// 前端组件
const frontendItems: DescriptionItem[] = FRONTEND_COMPONENTS.map((item) => ({
  title: item.name,
  content: h('div', [
    h('span', { class: 'text-foreground/80' }, item.version),
    h('span', { class: 'mx-2 text-foreground/30' }, '|'),
    renderLink(item.url, item.description),
  ]),
}));
</script>

<template>
  <Page title="系统信息">
    <template #description>
      <p class="mt-3 text-sm/6 text-foreground">
        <a :href="PROJECT_INFO.homepage" class="vben-link" target="_blank">
          {{ PROJECT_INFO.name }}
        </a>
        {{ PROJECT_INFO.description }}
      </p>
    </template>

    <!-- 关于项目 -->
    <div class="card-box p-5">
      <h5 class="text-lg text-foreground">关于项目</h5>
      <div class="mt-4">
        <dl class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4">
          <template v-for="item in projectItems" :key="item.title">
            <div
              class="border-t border-border px-4 py-6 sm:col-span-1 sm:px-0"
            >
              <dt class="text-sm/6 font-medium text-foreground">
                {{ item.title }}
              </dt>
              <dd class="mt-1 text-sm/6 text-foreground sm:mt-2">
                <component
                  :is="item.content"
                  v-if="typeof item.content === 'object'"
                />
                <span v-else>{{ item.content }}</span>
              </dd>
            </div>
          </template>
        </dl>
      </div>
    </div>

    <!-- 基本信息 -->
    <div class="card-box mt-6 p-5">
      <h5 class="text-lg text-foreground">基本信息</h5>
      <div class="mt-4">
        <dl
          v-if="!loading"
          class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4"
        >
          <template v-for="item in runtimeItems" :key="item.title">
            <div
              class="border-t border-border px-4 py-6 sm:col-span-1 sm:px-0"
            >
              <dt class="text-sm/6 font-medium text-foreground">
                {{ item.title }}
              </dt>
              <dd class="mt-1 text-sm/6 text-foreground sm:mt-2">
                <component
                  :is="item.content"
                  v-if="typeof item.content === 'object'"
                />
                <span v-else>{{ item.content }}</span>
              </dd>
            </div>
          </template>
        </dl>
        <div v-else class="py-8 text-center text-foreground/60">加载中...</div>
      </div>
    </div>

    <!-- 后端组件 -->
    <div class="card-box mt-6 p-5">
      <h5 class="text-lg text-foreground">后端组件</h5>
      <div class="mt-4">
        <dl class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4">
          <template v-for="item in backendItems" :key="item.title">
            <div
              class="border-t border-border px-4 py-3 sm:col-span-1 sm:px-0"
            >
              <dt class="text-sm text-foreground">
                {{ item.title }}
              </dt>
              <dd class="mt-1 text-sm text-foreground/80 sm:mt-2">
                <component
                  :is="item.content"
                  v-if="typeof item.content === 'object'"
                />
                <span v-else>{{ item.content }}</span>
              </dd>
            </div>
          </template>
        </dl>
      </div>
    </div>

    <!-- 前端组件 -->
    <div class="card-box mt-6 p-5">
      <h5 class="text-lg text-foreground">前端组件</h5>
      <div class="mt-4">
        <dl class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4">
          <template v-for="item in frontendItems" :key="item.title">
            <div
              class="border-t border-border px-4 py-3 sm:col-span-1 sm:px-0"
            >
              <dt class="text-sm text-foreground">
                {{ item.title }}
              </dt>
              <dd class="mt-1 text-sm text-foreground/80 sm:mt-2">
                <component
                  :is="item.content"
                  v-if="typeof item.content === 'object'"
                />
                <span v-else>{{ item.content }}</span>
              </dd>
            </div>
          </template>
        </dl>
      </div>
    </div>
  </Page>
</template>
