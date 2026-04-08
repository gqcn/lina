export type PluginType = 'runtime' | 'source' | string;

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
  releaseVersion: string;
  installed: number;
  installedAt: string;
  enabled: number;
  lifecycleState: string;
  nodeState: string;
  resourceCount: number;
  migrationState: string;
  statusKey: string;
  updatedAt: string;
}

export interface PluginRuntimeState {
  id: string;
  installed: number;
  enabled: number;
  statusKey: string;
}
