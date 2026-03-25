import type { DictType, DictTypeListParams } from './dict-type-model';

import { requestClient } from '#/api/request';

/** 字典类型列表 */
export async function dictTypeList(params?: DictTypeListParams) {
  const res = await requestClient.get<{ list: DictType[]; total: number }>(
    '/dict/type',
    { params },
  );
  // VXE-Grid proxy expects { items, total } format
  return { items: res.list, total: res.total };
}

/** 新增字典类型 */
export function dictTypeAdd(data: Partial<DictType>) {
  return requestClient.post('/dict/type', data);
}

/** 更新字典类型 */
export function dictTypeUpdate(id: number, data: Partial<DictType>) {
  return requestClient.put(`/dict/type/${id}`, data);
}

/** 删除字典类型 */
export function dictTypeDelete(id: number) {
  return requestClient.delete(`/dict/type/${id}`);
}

/** 获取字典类型详情 */
export function dictTypeInfo(id: number) {
  return requestClient.get<DictType>(`/dict/type/${id}`);
}

/** 导出字典类型 */
export function dictTypeExport(params?: DictTypeListParams) {
  return requestClient.download<Blob>('/dict/type/export', { params });
}

/** 获取字典类型选项列表 */
export async function dictTypeOptions() {
  const res = await requestClient.get<{ list: DictType[] }>(
    '/dict/type/options',
  );
  return res.list;
}

/** 导入字典类型 */
export function dictTypeImport(file: File, updateSupport?: boolean) {
  const formData = new FormData();
  formData.append('file', file);
  if (updateSupport) {
    formData.append('updateSupport', '1');
  }
  return requestClient.post<{
    success: number;
    fail: number;
    failList: Array<{ row: number; reason: string }>;
  }>('/dict/type/import', formData);
}

/** 下载字典类型导入模板 */
export function dictTypeImportTemplate() {
  return requestClient.download<Blob>('/dict/type/import-template');
}
