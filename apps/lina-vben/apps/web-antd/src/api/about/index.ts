import { requestClient } from '#/api/request';

export interface SystemInfoResult {
  goVersion: string;
  gfVersion: string;
  os: string;
  arch: string;
  dbVersion: string;
  startTime: string;
  runDuration: string;
}

/** 获取系统运行信息 */
export function getSystemInfo() {
  return requestClient.get<SystemInfoResult>('/system/info');
}
