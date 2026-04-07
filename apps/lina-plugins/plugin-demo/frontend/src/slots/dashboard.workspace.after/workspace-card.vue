<script lang="ts">
import { pluginSlotKeys } from '#/plugins/plugin-slots';

export const pluginSlotMeta = {
  order: 0,
  slotKey: pluginSlotKeys.dashboardWorkspaceAfter,
};
</script>

<template>
  <div class="rounded-2xl border border-dashed border-blue-200 bg-blue-50 p-4">
    <div class="text-sm font-semibold text-blue-700">
      插件示例工作台卡片
    </div>
    <p class="mt-2 text-sm text-blue-600">
      当前工作台卡片来自 plugin-demo 的真实前端 Slot 文件，用于验证源码插件启用后可在宿主工作台区域挂载简洁的扩展内容。
    </p>
  </div>
</template>
