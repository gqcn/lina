import type {
  PluginListParams,
  PluginRuntimeState,
  SystemPlugin,
} from './model';

import { requestClient } from '#/api/request';

/** 插件列表 */
export async function pluginList(params?: PluginListParams) {
  const res = await requestClient.get<{ list: SystemPlugin[]; total: number }>(
    '/plugins',
    { params },
  );
  return { items: res.list, total: res.total };
}

/** 公共插件运行时状态 */
export async function pluginRuntimeList() {
  const res = await requestClient.get<{ list: PluginRuntimeState[] }>(
    '/plugins/runtime',
  );
  return res.list;
}

/** 同步源码插件 */
export function pluginSync() {
  return requestClient.post<{ total: number }>('/plugins/sync');
}

/** 安装插件 */
export function pluginInstall(pluginId: string) {
  return requestClient.post(`/plugins/${pluginId}/install`);
}

/** 启用插件 */
export function pluginEnable(pluginId: string) {
  return requestClient.put(`/plugins/${pluginId}/enable`);
}

/** 禁用插件 */
export function pluginDisable(pluginId: string) {
  return requestClient.put(`/plugins/${pluginId}/disable`);
}

/** 卸载插件 */
export function pluginUninstall(pluginId: string) {
  return requestClient.delete(`/plugins/${pluginId}`);
}
