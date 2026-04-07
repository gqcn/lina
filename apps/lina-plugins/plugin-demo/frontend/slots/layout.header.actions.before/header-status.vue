<script lang="ts">
import { pluginSlotKeys } from '#/plugins/plugin-slots';

export const pluginSlotMeta = {
  order: 0,
  slotKey: pluginSlotKeys.layoutHeaderActionsBefore,
};
</script>

<template>
  <div class="rounded-full border border-rose-200 bg-rose-50 px-3 py-1 text-xs font-medium text-rose-700">
    plugin-demo 头部前置扩展
  </div>
</template>
