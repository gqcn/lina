<script lang="ts">
export const pluginPageMeta = {
  routePath: 'plugin-demo-sidebar-entry',
  title: '插件示例',
};
</script>

<script setup lang="ts">
import {
  Card as ACard,
  TypographyParagraph as ATypographyParagraph,
  TypographyTitle as ATypographyTitle,
} from 'ant-design-vue';

import { Page } from '#/plugins/runtime';
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
    </a-card>
  </Page>
</template>
