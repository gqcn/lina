<script lang="ts">
import { pluginSlotKeys } from '#/plugins/plugin-slots';

export const pluginSlotMeta = {
  order: 0,
  slotKey: pluginSlotKeys.crudTableAfter,
};
</script>

<template>
  <div class="rounded-2xl border border-slate-200 bg-slate-50 p-4 text-sm text-slate-600">
    plugin-demo 正在复用通用 CRUD 表格下方扩展位，用于验证 `crud.table.after` 会自动作用于基于 `useVbenVxeGrid` 的页面。
  </div>
</template>
