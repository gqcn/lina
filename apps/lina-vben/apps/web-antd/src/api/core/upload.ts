import type { AxiosRequestConfig } from '@vben/request';

import { requestClient } from '#/api/request';

/**
 * Axios upload progress event type
 */
export type AxiosProgressEvent = AxiosRequestConfig['onUploadProgress'];

/**
 * Upload result returned by the server
 */
export interface UploadResult {
  id: number;
  name: string;
  original: string;
  url: string;
  suffix: string;
  size: number;
}

/**
 * Upload a single file via the unified file upload API
 * @param file File to upload
 * @param options Upload options
 */
export function uploadApi(
  file: Blob | File,
  options?: {
    onUploadProgress?: AxiosProgressEvent;
    signal?: AbortSignal;
  },
) {
  const { onUploadProgress, signal } = options ?? {};
  const formData = new FormData();
  formData.append('file', file);
  return requestClient.post<UploadResult>('/file/upload', formData, {
    headers: { 'Content-Type': 'multipart/form-data' },
    onUploadProgress,
    signal,
    timeout: 60_000,
  });
}

/**
 * Upload API function type
 */
export type UploadApiFn = typeof uploadApi;
