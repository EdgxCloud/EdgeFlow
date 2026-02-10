/**
 * Services Index
 * Exports all API services
 */

export { api, healthCheck, WS_URL } from './api'
export { flowsApi } from './flows'
export { nodesApi } from './nodes'
export { resourcesApi } from './resources'
export { networkApi } from './network'
export { wsClient, WebSocketClient } from './websocket'

export type {
  Flow,
  FlowListItem,
  CreateFlowRequest,
  UpdateFlowRequest,
  DeployRequest,
  DeployResponse,
} from './flows'

export type {
  NodeType,
  NodeTypeCategory,
  NodeTypeInput,
  NodeTypeOutput,
  NodeTypesResponse,
} from './nodes'

export type {
  ResourceStats,
  CPUStats,
  MemoryStats,
  DiskStats,
  NetworkStats,
} from './resources'

export type {
  NetworkInfo,
  NetworkInterface,
} from './network'

export type {
  WSMessage,
  WSMessageType,
  FlowStatusMessage,
  NodeStatusMessage,
  ExecutionMessage,
  LogMessage,
  NotificationMessage,
  WSClientOptions,
} from './websocket'
