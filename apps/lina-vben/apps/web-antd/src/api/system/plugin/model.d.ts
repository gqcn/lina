export type PluginType = 'package' | 'source' | 'wasm' | string;

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
  runtime: string;
  entry: string;
  description: string;
  installed: number;
  enabled: number;
  statusKey: string;
  updatedAt: string;
}

export interface PluginRuntimeState {
  id: string;
  installed: number;
  enabled: number;
  statusKey: string;
}
