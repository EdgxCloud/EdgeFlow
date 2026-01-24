import { create } from 'zustand';
import { api } from '../lib/api';

export interface NodeType {
  type: string;
  name: string;
  category: string;
  description: string;
  icon?: string;
  color?: string;
  inputs: number;
  outputs: number;
  config?: Record<string, any>;
  properties?: NodeProperty[];
}

export interface NodeProperty {
  name: string;
  type: 'string' | 'number' | 'boolean' | 'select' | 'textarea' | 'code';
  label: string;
  defaultValue?: any;
  options?: Array<{ value: string; label: string }>;
  required?: boolean;
  description?: string;
}

interface NodeState {
  nodeTypes: NodeType[];
  isLoading: boolean;
  error: string | null;

  fetchNodeTypes: () => Promise<void>;
  getNodeType: (type: string) => NodeType | undefined;
  getNodesByCategory: (category: string) => NodeType[];
  searchNodes: (query: string) => NodeType[];
}

export const useNodeStore = create<NodeState>((set, get) => ({
  nodeTypes: [],
  isLoading: false,
  error: null,

  fetchNodeTypes: async () => {
    set({ isLoading: true, error: null });
    try {
      const response = await api.get('/nodes/types');
      set({ nodeTypes: response.data, isLoading: false });
    } catch (error) {
      set({
        error: error instanceof Error ? error.message : 'Failed to fetch node types',
        isLoading: false,
      });
    }
  },

  getNodeType: (type: string) => {
    return get().nodeTypes.find((node) => node.type === type);
  },

  getNodesByCategory: (category: string) => {
    return get().nodeTypes.filter((node) => node.category === category);
  },

  searchNodes: (query: string) => {
    const lowerQuery = query.toLowerCase();
    return get().nodeTypes.filter(
      (node) =>
        node.name.toLowerCase().includes(lowerQuery) ||
        node.description.toLowerCase().includes(lowerQuery) ||
        node.category.toLowerCase().includes(lowerQuery)
    );
  },
}));
