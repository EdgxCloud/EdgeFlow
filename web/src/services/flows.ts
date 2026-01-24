/**
 * Flow Management API Service
 * Handles all flow-related API calls
 */

import { api } from './api'
import { Node, Edge } from '@xyflow/react'

// Flow Types
export interface Flow {
  id: string
  name: string
  description?: string
  nodes: Node[]
  edges: Edge[]
  status?: 'running' | 'stopped' | 'error'
  created_at?: string
  updated_at?: string
  metadata?: Record<string, unknown>
}

export interface FlowListItem {
  id: string
  name: string
  description?: string
  status?: 'running' | 'stopped' | 'error'
  created_at?: string
  updated_at?: string
}

export interface CreateFlowRequest {
  name: string
  description?: string
  nodes?: Node[]
  edges?: Edge[]
  metadata?: Record<string, unknown>
}

export interface UpdateFlowRequest {
  name?: string
  description?: string
  nodes?: Node[]
  edges?: Edge[]
  metadata?: Record<string, unknown>
}

export interface DeployRequest {
  mode: 'full' | 'modified' | 'flow'
  flow_ids?: string[]
}

export interface DeployResponse {
  success: boolean
  message: string
  deployed_flows?: string[]
  errors?: Record<string, string>
}

// Flow API Methods
export const flowsApi = {
  /**
   * List all flows
   */
  list: async (): Promise<FlowListItem[]> => {
    return api.get<FlowListItem[]>('/flows')
  },

  /**
   * Get a specific flow by ID
   */
  get: async (id: string): Promise<Flow> => {
    return api.get<Flow>(`/flows/${id}`)
  },

  /**
   * Create a new flow
   */
  create: async (data: CreateFlowRequest): Promise<Flow> => {
    return api.post<Flow>('/flows', data)
  },

  /**
   * Update an existing flow
   */
  update: async (id: string, data: UpdateFlowRequest): Promise<Flow> => {
    return api.put<Flow>(`/flows/${id}`, data)
  },

  /**
   * Delete a flow
   */
  delete: async (id: string): Promise<void> => {
    return api.delete<void>(`/flows/${id}`)
  },

  /**
   * Start a flow
   */
  start: async (id: string): Promise<void> => {
    return api.post<void>(`/flows/${id}/start`)
  },

  /**
   * Stop a flow
   */
  stop: async (id: string): Promise<void> => {
    return api.post<void>(`/flows/${id}/stop`)
  },

  /**
   * Deploy flows (full, modified, or specific flows)
   */
  deploy: async (data: DeployRequest): Promise<DeployResponse> => {
    return api.post<DeployResponse>('/flows/deploy', data)
  },

  /**
   * Export flow to JSON
   */
  export: (flow: Flow): string => {
    return JSON.stringify(flow, null, 2)
  },

  /**
   * Import flow from JSON
   */
  import: (json: string): Flow => {
    try {
      const flow = JSON.parse(json) as Flow
      // Validate required fields
      if (!flow.name || !Array.isArray(flow.nodes)) {
        throw new Error('Invalid flow format')
      }
      return flow
    } catch (error) {
      throw new Error(`Failed to import flow: ${error instanceof Error ? error.message : 'Unknown error'}`)
    }
  },

  /**
   * Validate flow before save/deploy
   */
  validate: (nodes: Node[], edges: Edge[]): { valid: boolean; errors: string[] } => {
    const errors: string[] = []

    // Check for nodes
    if (nodes.length === 0) {
      errors.push('Flow must have at least one node')
    }

    // Check for duplicate node IDs
    const nodeIds = new Set<string>()
    for (const node of nodes) {
      if (nodeIds.has(node.id)) {
        errors.push(`Duplicate node ID: ${node.id}`)
      }
      nodeIds.add(node.id)
    }

    // Check edge references
    for (const edge of edges) {
      if (!nodeIds.has(edge.source)) {
        errors.push(`Edge ${edge.id} references missing source node: ${edge.source}`)
      }
      if (!nodeIds.has(edge.target)) {
        errors.push(`Edge ${edge.id} references missing target node: ${edge.target}`)
      }
    }

    return {
      valid: errors.length === 0,
      errors,
    }
  },
}

export default flowsApi
