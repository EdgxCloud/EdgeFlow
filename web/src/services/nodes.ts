/**
 * Node Types API Service
 * Handles fetching available node types from backend
 */

import { api } from './api'

// Node Type Definitions
export interface NodeTypeCategory {
  id: string
  name: string
  description?: string
  icon?: string
}

export interface NodeTypeInput {
  name: string
  type: string
  required?: boolean
  default?: unknown
  description?: string
  options?: string[]
}

export interface NodeTypeOutput {
  name: string
  type: string
  description?: string
}

export interface NodeType {
  id: string
  name: string
  category: string
  description?: string
  icon?: string
  color?: string
  inputs?: NodeTypeInput[]
  outputs?: NodeTypeOutput[]
  properties?: Record<string, unknown>
  version?: string
  module?: string
}

export interface NodeTypesResponse {
  categories: NodeTypeCategory[]
  node_types: NodeType[]
}

// Node Types API Methods
export const nodesApi = {
  /**
   * Get all available node types
   */
  getTypes: async (): Promise<NodeTypesResponse> => {
    return api.get<NodeTypesResponse>('/node-types')
  },

  /**
   * Get node types by category
   */
  getTypesByCategory: async (category: string): Promise<NodeType[]> => {
    const response = await nodesApi.getTypes()
    return response.node_types.filter(nt => nt.category === category)
  },

  /**
   * Get all categories
   */
  getCategories: async (): Promise<NodeTypeCategory[]> => {
    const response = await nodesApi.getTypes()
    return response.categories
  },

  /**
   * Search node types by name or description
   */
  search: async (query: string): Promise<NodeType[]> => {
    const response = await nodesApi.getTypes()
    const lowerQuery = query.toLowerCase()
    return response.node_types.filter(
      nt =>
        nt.name.toLowerCase().includes(lowerQuery) ||
        nt.description?.toLowerCase().includes(lowerQuery)
    )
  },
}

export default nodesApi
