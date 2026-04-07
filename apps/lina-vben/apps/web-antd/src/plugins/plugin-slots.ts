export const pluginSlotKeys = {
  dashboardWorkspaceAfter: 'dashboard.workspace.after',
  layoutUserDropdownAfter: 'layout.user-dropdown.after',
} as const;

export type PluginSlotKey =
  (typeof pluginSlotKeys)[keyof typeof pluginSlotKeys];

export interface PublishedPluginSlot {
  description: string;
  hostLocation: string;
  key: PluginSlotKey;
}

const publishedPluginSlotKeySet = new Set<PluginSlotKey>(
  Object.values(pluginSlotKeys),
);

export const publishedPluginSlots: PublishedPluginSlot[] = [
  {
    description: '工作台主内容区底部扩展区域，适合挂载插件卡片或统计块。',
    hostLocation: 'dashboard.workspace',
    key: pluginSlotKeys.dashboardWorkspaceAfter,
  },
  {
    description: '右上角用户菜单左侧扩展区域，适合挂载轻量入口或状态提示。',
    hostLocation: 'layout.user-dropdown',
    key: pluginSlotKeys.layoutUserDropdownAfter,
  },
];

export function isPluginSlotKey(value: string): value is PluginSlotKey {
  return publishedPluginSlotKeySet.has(value as PluginSlotKey);
}
