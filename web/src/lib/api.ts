import axios from 'axios'

const API_BASE_URL = import.meta.env.VITE_API_URL || '/api/v1'

export const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Types
export interface Flow {
  id: string
  name: string
  description: string
  status: 'idle' | 'running' | 'stopped' | 'error'
  nodes: Record<string, Node>
  connections: Connection[]
  config: Record<string, any>
  createdAt?: string
  updatedAt?: string
}

export interface Node {
  id: string
  type: string
  name: string
  category: 'input' | 'output' | 'processing' | 'function'
  config: Record<string, any>
  inputs: string[]
  outputs: string[]
  status: 'idle' | 'running' | 'error'
}

export interface Connection {
  id: string
  source_id: string
  target_id: string
}

export interface NodeType {
  type: string
  name: string
  category: string
  description: string
  icon: string
  color: string
  properties: PropertySchema[]
  inputs: (string | PortSchema)[]
  outputs: (string | PortSchema)[]
}

export interface PropertySchema {
  name: string
  label: string
  type: string
  default: any
  required: boolean
  description: string
  options?: string[]
}

export interface PortSchema {
  name: string
  label: string
  type: string
  description: string
}

// API functions
export const flowsApi = {
  list: () => api.get<{ flows: Flow[]; count: number }>('/flows'),
  get: (id: string) => api.get<Flow>(`/flows/${id}`),
  create: (data: { name: string; description: string }) =>
    api.post<Flow>('/flows', data),
  update: (id: string, data: Partial<Flow>) =>
    api.put<Flow>(`/flows/${id}`, data),
  delete: (id: string) => api.delete(`/flows/${id}`),
  start: (id: string) => api.post(`/flows/${id}/start`),
  stop: (id: string) => api.post(`/flows/${id}/stop`),
}

export const nodesApi = {
  list: (flowId: string) =>
    api.get<{ nodes: Record<string, Node>; count: number }>(
      `/flows/${flowId}/nodes`
    ),
  add: (
    flowId: string,
    data: { type: string; name: string; config?: Record<string, any> }
  ) => api.post<Node>(`/flows/${flowId}/nodes`, data),
  get: (flowId: string, nodeId: string) =>
    api.get<Node>(`/flows/${flowId}/nodes/${nodeId}`),
  update: (flowId: string, nodeId: string, config: Record<string, any>) =>
    api.put<Node>(`/flows/${flowId}/nodes/${nodeId}`, { config }),
  delete: (flowId: string, nodeId: string) =>
    api.delete(`/flows/${flowId}/nodes/${nodeId}`),
}

export const connectionsApi = {
  list: (flowId: string) =>
    api.get<{ connections: Connection[]; count: number }>(
      `/flows/${flowId}/connections`
    ),
  create: (flowId: string, data: { source_id: string; target_id: string }) =>
    api.post<Connection>(`/flows/${flowId}/connections`, data),
  delete: (flowId: string, connId: string) =>
    api.delete(`/flows/${flowId}/connections/${connId}`),
}

export const nodeTypesApi = {
  list: () => api.get<{ node_types: NodeType[]; count: number }>('/node-types'),
  get: (type: string) => api.get<NodeType>(`/node-types/${type}`),
}

export const healthApi = {
  check: () =>
    api.get<{ status: string; version: string; websocket_clients: number }>(
      '/health'
    ),
}

// Module/Plugin Types
export interface Module {
  name: string
  version: string
  description: string
  author: string
  category: 'core' | 'network' | 'gpio' | 'database' | 'messaging' | 'ai' | 'industrial' | 'advanced' | 'ui'
  status: 'not_loaded' | 'loading' | 'loaded' | 'unloading' | 'error'
  loaded_at?: string
  unloaded_at?: string
  error?: string
  required_memory_mb: number
  required_disk_mb: number
  dependencies: string[]
  nodes: NodeDefinition[]
  config: Record<string, any>
  compatible: boolean
  compatible_reason?: string
}

export interface NodeDefinition {
  type: string
  name: string
  category: string
  description: string
  icon: string
  color: string
  inputs: number
  outputs: number
  config: Record<string, any>
}

export interface ModuleStats {
  total_plugins: number
  loaded_plugins: number
  enabled_plugins: number
  total_nodes: number
  load_order: string[]
  memory_available_mb: number
}

export interface ResourceStats {
  timestamp: string
  memory: {
    total_mb: number
    used_mb: number
    available_mb: number
    percent: string
  }
  disk: {
    total_mb: number
    used_mb: number
    available_mb: number
    percent: string
  }
  cpu: {
    cores: number
    goroutines: number
  }
  limits: {
    memory_limit_mb: number
    memory_hard_limit_mb: number
    low_memory_threshold_mb: number
  }
  modules: {
    enabled: string[]
  }
}

// ============================================
// Module Marketplace Types
// ============================================

export interface MarketplaceSearchResult {
  name: string
  version?: string
  description: string
  keywords?: string[]
  author?: string
  url?: string
  repository?: string
  score?: number
  downloads?: number
  rating?: number
  stars?: number
  forks?: number
  language?: string
  topics?: string[]
  updated?: string
  types?: string[]
  source: 'npm' | 'node-red' | 'github'
  avatar?: string
  owner?: string
}

export interface MarketplaceSearchResponse {
  results: MarketplaceSearchResult[]
  total: number
  query: string
}

export interface ModuleInstallRequest {
  url?: string
  npm?: string
  github?: string
  path?: string
}

export interface ModuleInstallResponse {
  message: string
  module: {
    name: string
    version: string
    description: string
    format: string
    nodes: NodeDefinition[]
  }
  status: string
  validation: {
    valid: boolean
    errors?: string[]
    warnings?: string[]
  }
}

// ============================================
// Module APIs
// ============================================

export const modulesApi = {
  // List & get modules
  list: () => api.get<{ modules: Module[]; count: number }>('/modules'),
  get: (name: string) => api.get<Module>(`/modules/${name}`),

  // Module lifecycle
  load: (name: string) => api.post(`/modules/${name}/load`),
  unload: (name: string) => api.post(`/modules/${name}/unload`),
  enable: (name: string) => api.post(`/modules/${name}/enable`),
  disable: (name: string) => api.post(`/modules/${name}/disable`),
  reload: (name: string) => api.post(`/modules/${name}/reload`),
  stats: () => api.get<ModuleStats>('/modules/stats'),

  // Search marketplace
  searchNpm: (query: string) =>
    api.get<MarketplaceSearchResponse>(`/modules/search/npm?q=${encodeURIComponent(query)}`),
  searchNodeRed: (query: string) =>
    api.get<MarketplaceSearchResponse>(`/modules/search/nodered?q=${encodeURIComponent(query)}`),
  searchGitHub: (query: string) =>
    api.get<MarketplaceSearchResponse>(`/modules/search/github?q=${encodeURIComponent(query)}`),

  // Install modules
  install: (request: ModuleInstallRequest) =>
    api.post<ModuleInstallResponse>('/modules/install', request),

  // Upload module file
  upload: (file: File) => {
    const formData = new FormData()
    formData.append('module', file)
    return api.post<ModuleInstallResponse>('/modules/upload', formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })
  },

  // Uninstall module
  uninstall: (name: string) =>
    api.delete<{ message: string }>(`/modules/${name}`),
}

export interface ExecutionRecord {
  id: string
  flow_id: string
  flow_name: string
  status: 'running' | 'completed' | 'failed'
  start_time: string
  end_time?: string
  duration?: number
  node_count: number
  completed_nodes: number
  error_nodes: number
  error?: string
  node_events?: NodeExecutionEvent[]
}

export interface NodeExecutionEvent {
  node_id: string
  node_name: string
  node_type: string
  status: string
  execution_time: number
  timestamp: number
  error?: string
}

export const executionsApi = {
  list: (status?: string) => {
    const params = status ? `?status=${status}` : ''
    return api.get<{ executions: ExecutionRecord[]; count: number }>(`/executions${params}`)
  },
  get: (id: string) => api.get<ExecutionRecord>(`/executions/${id}`),
}

export const resourcesApi = {
  stats: () => api.get<ResourceStats>('/resources/stats'),
  report: () => api.get<ResourceStats>('/resources/report'),
}
