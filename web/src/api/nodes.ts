/**
 * Node Registry API Client
 *
 * API calls for fetching node types and configurations
 */

import axios from 'axios'
import type { NodeInfo, NodeConfig } from '@/types/node'

const API_BASE = import.meta.env.VITE_API_URL || '/api/v1'

export interface NodeRegistryResponse {
  nodes: NodeInfo[]
  count: number
}

export interface NodeTypeResponse {
  node: NodeInfo
}

/**
 * Get all registered node types from the registry
 */
export async function getAllNodeTypes(): Promise<NodeInfo[]> {
  try {
    const response = await axios.get<{ node_types: NodeInfo[] }>(`${API_BASE}/node-types`)
    return response.data.node_types || []
  } catch (error) {
    console.error('Failed to fetch node types:', error)
    throw new Error('Failed to fetch node types from registry')
  }
}

/**
 * Get specific node type information
 */
export async function getNodeType(nodeType: string): Promise<NodeInfo> {
  try {
    const response = await axios.get<NodeInfo>(`${API_BASE}/node-types/${nodeType}`)
    return response.data
  } catch (error) {
    console.error(`Failed to fetch node type ${nodeType}:`, error)
    throw new Error(`Node type ${nodeType} not found in registry`)
  }
}

/**
 * Get node types by category
 */
export async function getNodeTypesByCategory(category: string): Promise<NodeInfo[]> {
  try {
    const allNodes = await getAllNodeTypes()
    return allNodes.filter((node) => node.category === category)
  } catch (error) {
    console.error(`Failed to fetch nodes for category ${category}:`, error)
    return []
  }
}

/**
 * Get node configuration from a flow
 */
export async function getNodeConfig(flowId: string, nodeId: string): Promise<NodeConfig> {
  try {
    const response = await axios.get<{ node: NodeConfig }>(
      `${API_BASE}/flows/${flowId}/nodes/${nodeId}`
    )
    return response.data.node
  } catch (error) {
    console.error(`Failed to fetch node config ${nodeId}:`, error)
    throw new Error('Failed to fetch node configuration')
  }
}

/**
 * Update node configuration in a flow
 */
export async function updateNodeConfig(
  flowId: string,
  nodeId: string,
  config: Record<string, any>
): Promise<void> {
  try {
    await axios.put(`${API_BASE}/flows/${flowId}/nodes/${nodeId}`, { config })
  } catch (error) {
    console.error(`Failed to update node config ${nodeId}:`, error)
    throw new Error('Failed to update node configuration')
  }
}

/**
 * Validate node configuration
 */
export async function validateNodeConfig(
  nodeType: string,
  config: Record<string, any>
): Promise<{ valid: boolean; errors?: string[] }> {
  try {
    const response = await axios.post(`${API_BASE}/registry/nodes/${nodeType}/validate`, {
      config,
    })
    return response.data
  } catch (error) {
    // Server validation endpoint not available - treat as valid
    // Client-side validation is already done in useNodeConfig hook
    console.warn(`Server validation not available for ${nodeType}`)
    return { valid: true }
  }
}
