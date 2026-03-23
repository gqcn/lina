import { requestClient } from '#/api/request';

export namespace JobApi {
  export interface Job {
    id: number;
    name: string;
    group: string;
    command: string;
    cronExpr: string;
    description?: string;
    status: number;
    singleton: number;
    maxTimes: number;
    execTimes: number;
    isSystem: number;
    createBy?: string;
    createTime?: string;
    updateBy?: string;
    updateTime?: string;
    remark?: string;
  }

  export interface JobLog {
    id: number;
    jobId: number;
    jobName: string;
    jobGroup: string;
    command: string;
    status: number;
    startTime: string;
    endTime?: string;
    duration?: number;
    errorMsg?: string;
    createTime?: string;
  }

  export interface ListParams {
    name?: string;
    group?: string;
    status?: number;
    page: number;
    pageSize: number;
  }

  export interface LogListParams {
    jobName?: string;
    status?: number;
    startTime?: string;
    endTime?: string;
    page: number;
    pageSize: number;
  }
}

export const jobApi = {
  list: (params: JobApi.ListParams) =>
    requestClient.get<{ items: JobApi.Job[]; total: number }>('/job/list', {
      params,
    }),

  create: (data: Partial<JobApi.Job>) => requestClient.post('/job/create', data),

  update: (data: Partial<JobApi.Job>) => requestClient.put('/job/update', data),

  delete: (ids: number[]) => requestClient.delete('/job/delete', { data: { ids } }),

  updateStatus: (id: number, status: number) =>
    requestClient.put('/job/status', { id, status }),

  run: (id: number) => requestClient.post('/job/run', { id }),

  logList: (params: JobApi.LogListParams) =>
    requestClient.get<{ items: JobApi.JobLog[]; total: number }>(
      '/job/log/list',
      { params },
    ),
};
