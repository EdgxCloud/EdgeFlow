import { create } from 'zustand';
import { api } from '../lib/api';

export interface Execution {
  id: string;
  flowId: string;
  flowName: string;
  status: 'running' | 'completed' | 'failed';
  startTime: string;
  endTime?: string;
  duration?: number;
  error?: string;
  logs: ExecutionLog[];
}

export interface ExecutionLog {
  timestamp: string;
  nodeId: string;
  nodeName: string;
  level: 'info' | 'warn' | 'error';
  message: string;
  data?: any;
}

interface ExecutionState {
  executions: Execution[];
  isLoading: boolean;
  error: string | null;

  fetchExecutions: () => Promise<void>;
  getExecutionById: (id: string) => Execution | undefined;
  filterExecutions: (filters: ExecutionFilters) => Execution[];
  clearExecutions: () => void;
}

export interface ExecutionFilters {
  status?: 'running' | 'completed' | 'failed';
  flowId?: string;
  startDate?: Date;
  endDate?: Date;
}

export const useExecutionStore = create<ExecutionState>((set, get) => ({
  executions: [],
  isLoading: false,
  error: null,

  fetchExecutions: async () => {
    set({ isLoading: true, error: null });
    try {
      const response = await api.get('/executions');
      set({ executions: response.data, isLoading: false });
    } catch (error) {
      set({
        error: error instanceof Error ? error.message : 'Failed to fetch executions',
        isLoading: false,
      });
    }
  },

  getExecutionById: (id: string) => {
    return get().executions.find((exec) => exec.id === id);
  },

  filterExecutions: (filters: ExecutionFilters) => {
    let filtered = get().executions;

    if (filters.status) {
      filtered = filtered.filter((exec) => exec.status === filters.status);
    }

    if (filters.flowId) {
      filtered = filtered.filter((exec) => exec.flowId === filters.flowId);
    }

    if (filters.startDate) {
      filtered = filtered.filter(
        (exec) => new Date(exec.startTime) >= filters.startDate!
      );
    }

    if (filters.endDate) {
      filtered = filtered.filter(
        (exec) => new Date(exec.startTime) <= filters.endDate!
      );
    }

    return filtered;
  },

  clearExecutions: () => {
    set({ executions: [], error: null });
  },
}));
