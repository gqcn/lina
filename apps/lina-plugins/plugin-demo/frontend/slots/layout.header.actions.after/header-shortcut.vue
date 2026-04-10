<script lang="ts">
import { pluginSlotKeys } from '#/plugins/plugin-slots';

export const pluginSlotMeta = {
  order: 0,
  slotKey: pluginSlotKeys.layoutHeaderActionsAfter,
};

export default {};
</script>

<template>
  <div class="rounded-full border border-violet-200 bg-violet-50 px-3 py-1 text-xs font-medium text-violet-700">
    plugin-demo 头部后置扩展
  </div>
</template>
