export type PluginType = 'dynamic' | 'source' | string;

export interface PluginListParams {
  pageNum?: number;
  pageSize?: number;
  id?: string;
  installed?: number;
  name?: string;
  status?: number;
  type?: PluginType;
}

export interface SystemPlugin {
  id: string;
  name: string;
  version: string;
  type: PluginType;
  description: string;
  installed: number;
  installedAt: string;
  enabled: number;
  statusKey: string;
  updatedAt: string;
}

export interface PluginDynamicState {
  id: string;
  installed: number;
  enabled: number;
  version: string;
  generation: number;
  statusKey: string;
}

export interface PluginUploadDynamicResult {
  id: string;
  name: string;
  version: string;
  type: PluginType;
  runtimeKind: string;
  runtimeAbi: string;
  installed: number;
  enabled: number;
}
