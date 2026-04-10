<script lang="ts">
import { pluginSlotKeys } from '#/plugins/plugin-slots';

export const pluginSlotMeta = {
  order: 0,
  slotKey: pluginSlotKeys.crudToolbarAfter,
};

export default {};
</script>

<template>
  <div class="rounded-full border border-sky-200 bg-sky-50 px-3 py-1 text-xs font-medium text-sky-700">
    plugin-demo CRUD 扩展
  </div>
</template>
