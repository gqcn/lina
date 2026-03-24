import type { ConfigListParams, SysConfig } from './model';

import { requestClient } from '#/api/request';

/** 参数设置列表 */
export async function configList(params?: ConfigListParams) {
  const res = await requestClient.get<{ list: SysConfig[]; total: number }>(
    '/config',
    { params },
  );
  // VXE-Grid proxy expects { items, total } format
  return { items: res.list, total: res.total };
}

/** 新增参数设置 */
export function configAdd(data: Partial<SysConfig>) {
  return requestClient.post('/config', data);
}

/** 更新参数设置 */
export function configUpdate(id: number, data: Partial<SysConfig>) {
  return requestClient.put(`/config/${id}`, data);
}

/** 删除参数设置 */
export function configDelete(id: number) {
  return requestClient.delete(`/config/${id}`);
}

/** 获取参数设置详情 */
export function configInfo(id: number) {
  return requestClient.get<SysConfig>(`/config/${id}`);
}

/** 导出参数设置 */
export function configExport(params?: ConfigListParams) {
  return requestClient.download<Blob>('/config/export', { params });
}
