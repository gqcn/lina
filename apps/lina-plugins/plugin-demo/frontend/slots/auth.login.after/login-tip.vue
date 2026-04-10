<script lang="ts">
import { pluginSlotKeys } from '#/plugins/plugin-slots';

export const pluginSlotMeta = {
  order: 0,
  slotKey: pluginSlotKeys.authLoginAfter,
};

export default {};
</script>

<template>
  <div class="rounded-2xl border border-emerald-200 bg-emerald-50 p-3 text-sm text-emerald-700">
    plugin-demo 已向登录页公开区注册扩展内容，用于验证 `auth.login.after` 插槽。
  </div>
</template>
