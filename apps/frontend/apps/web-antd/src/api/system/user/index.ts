import { requestClient } from '#/api/request';

export interface SysUser {
  id: number;
  username: string;
  nickname: string;
  email: string;
  phone: string;
  avatar: string;
  status: number;
  remark: string;
  loginDate: string;
  createdAt: string;
  updatedAt: string;
}

export interface UserListParams {
  pageNum?: number;
  pageSize?: number;
  username?: string;
  nickname?: string;
  status?: number;
  phone?: string;
  beginTime?: string;
  endTime?: string;
  orderBy?: string;
  orderDirection?: string;
}

export interface UserListResult {
  list: SysUser[];
  total: number;
}

export interface UserCreateParams {
  username: string;
  password: string;
  nickname?: string;
  email?: string;
  phone?: string;
  status?: number;
  remark?: string;
}

export interface UserUpdateParams {
  id: number;
  username?: string;
  password?: string;
  nickname?: string;
  email?: string;
  phone?: string;
  status?: number;
  remark?: string;
}

/** 用户列表 */
export async function userList(params?: UserListParams) {
  const res = await requestClient.get<UserListResult>('/user', { params });
  // VXE-Grid proxy expects { items, total } format
  return { items: res.list, total: res.total };
}

/** 创建用户 */
export function userAdd(data: UserCreateParams) {
  return requestClient.post('/user', data);
}

/** 更新用户 */
export function userUpdate(data: UserUpdateParams) {
  return requestClient.put(`/user/${data.id}`, data);
}

/** 删除用户 */
export function userDelete(id: number) {
  return requestClient.delete(`/user/${id}`);
}

/** 获取用户详情 */
export function userInfo(id: number) {
  return requestClient.get<SysUser>(`/user/${id}`);
}

/** 修改用户状态 */
export function userStatusChange(id: number, status: number) {
  return requestClient.put(`/user/${id}/status`, { status });
}

/** 获取当前用户信息 */
export function getProfile() {
  return requestClient.get<SysUser>('/user/profile');
}

/** 更新当前用户信息 */
export function updateProfile(data: {
  nickname?: string;
  email?: string;
  phone?: string;
  password?: string;
}) {
  return requestClient.put('/user/profile', data);
}

/** 导出用户列表为 Excel */
export function userExport(params?: { ids?: number[] }) {
  return requestClient.download<Blob>('/user/export', {
    params,
  });
}

/** 导入用户 */
export function userImport(file: File, updateSupport?: boolean) {
  const formData = new FormData();
  formData.append('file', file);
  if (updateSupport) {
    formData.append('updateSupport', '1');
  }
  return requestClient.post<{
    success: number;
    fail: number;
    failList: Array<{ row: number; reason: string }>;
  }>('/user/import', formData);
}

/** 下载导入模板 */
export function userImportTemplate() {
  return requestClient.download<Blob>('/user/import-template');
}

/** 重置用户密码 */
export function userResetPassword(id: number, password: string) {
  return requestClient.put(`/user/${id}/reset-password`, { password });
}

/** 上传头像 */
export function userUpdateAvatar(fileCallback: {
  file: Blob;
  filename: string;
}) {
  let { file } = fileCallback;
  const { filename } = fileCallback;
  const uniqueName = filename || `${Date.now()}_${Math.random().toString(36).slice(2, 10)}.png`;
  file = new File([file], uniqueName);
  return requestClient.post(
    '/user/profile/avatar',
    { avatarfile: file },
    { headers: { 'Content-Type': 'multipart/form-data' } },
  );
}
