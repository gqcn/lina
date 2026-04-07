<script lang="ts">
import { pluginSlotKeys } from '#/plugins/plugin-slots';

export const pluginSlotMeta = {
  order: 0,
  slotKey: pluginSlotKeys.dashboardWorkspaceBefore,
};
</script>

<template>
  <div class="rounded-3xl border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-700">
    plugin-demo 正在通过 `dashboard.workspace.before` 在工作台顶部插入横幅内容。
  </div>
</template>
