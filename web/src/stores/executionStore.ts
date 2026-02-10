import { create } from 'zustand'
import { executionsApi } from '../lib/api'
import type { ExecutionRecord } from '../lib/api'

interface ExecutionState {
  executions: ExecutionRecord[]
  isLoading: boolean
  error: string | null

  fetchExecutions: (status?: string) => Promise<void>
  getExecutionById: (id: string) => ExecutionRecord | undefined
  clearExecutions: () => void
}

export const useExecutionStore = create<ExecutionState>((set, get) => ({
  executions: [],
  isLoading: false,
  error: null,

  fetchExecutions: async (status?: string) => {
    set({ isLoading: true, error: null })
    try {
      const response = await executionsApi.list(status)
      set({ executions: response.data.executions || [], isLoading: false })
    } catch (error) {
      set({
        error: error instanceof Error ? error.message : 'Failed to fetch executions',
        isLoading: false,
      })
    }
  },

  getExecutionById: (id: string) => {
    return get().executions.find((exec) => exec.id === id)
  },

  clearExecutions: () => {
    set({ executions: [], error: null })
  },
}))
