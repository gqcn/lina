<script lang="ts">
export const pluginPageMeta = {
  routePath: 'plugin-demo-sidebar-entry',
  title: '插件示例',
};
</script>

<script setup lang="ts">
import { onMounted, ref } from 'vue';

import {
  Card as ACard,
  Descriptions as ADescriptions,
  DescriptionsItem as ADescriptionsItem,
  Tag as ATag,
  TypographyParagraph as ATypographyParagraph,
  TypographyTitle as ATypographyTitle,
} from 'ant-design-vue';

import { requestClient } from '#/api/request';
import { Page } from '#/plugins/runtime';

interface PluginSummary {
  callbackModes: string[];
  cronJobName: string;
  cronPattern: string;
  cronPrimaryAware: boolean;
  extensionPoints: string[];
  generatedAt: string;
  message: string;
  pluginId: string;
}

const summary = ref<null | PluginSummary>(null);

async function loadSummary() {
  summary.value = await requestClient.get<PluginSummary>('/plugins/plugin-demo/summary');
}

onMounted(() => {
  void loadSummary();
});
</script>

<template>
  <Page :auto-content-height="true">
    <a-card :bordered="false" class="border border-slate-200">
      <a-typography-title :level="3" class="!mb-3">
        插件示例已生效
      </a-typography-title>
      <a-typography-paragraph class="!mb-0 text-slate-600">
        当前页面来自 plugin-demo 的左侧菜单入口，用于验证源码插件可以向宿主左侧导航插入页面，并在后台主内容区正常打开。
      </a-typography-paragraph>
      <a-descriptions
        v-if="summary"
        :column="1"
        class="mt-6 rounded-xl bg-slate-50 p-4"
        size="small"
      >
        <a-descriptions-item label="插件标识">
          {{ summary.pluginId }}
        </a-descriptions-item>
        <a-descriptions-item label="后端说明">
          {{ summary.message }}
        </a-descriptions-item>
        <a-descriptions-item label="后端扩展点">
          <div class="flex flex-wrap gap-2">
            <a-tag
              v-for="item in summary.extensionPoints"
              :key="item"
              color="geekblue"
            >
              {{ item }}
            </a-tag>
          </div>
        </a-descriptions-item>
        <a-descriptions-item label="回调模式">
          <div class="flex flex-wrap gap-2">
            <a-tag v-for="item in summary.callbackModes" :key="item" color="blue">
              {{ item }}
            </a-tag>
          </div>
        </a-descriptions-item>
        <a-descriptions-item label="定时任务">
          <div class="flex flex-wrap gap-2">
            <a-tag color="cyan">
              {{ summary.cronJobName }}
            </a-tag>
            <a-tag color="gold">
              {{ summary.cronPattern }}
            </a-tag>
            <a-tag :color="summary.cronPrimaryAware ? 'green' : 'default'">
              {{ summary.cronPrimaryAware ? '支持主节点识别' : '无主节点识别' }}
            </a-tag>
          </div>
        </a-descriptions-item>
        <a-descriptions-item label="生成时间">
          {{ summary.generatedAt }}
        </a-descriptions-item>
      </a-descriptions>
    </a-card>
  </Page>
</template>
