export interface Flow {
  id: string;
  name: string;
  description?: string;
  nodes: FlowNode[];
  connections: Connection[];
  status?: 'running' | 'stopped' | 'error';
  createdAt?: string;
  updatedAt?: string;
}

// n8n-style position format: [x, y] array
export type Position = [number, number];

export interface FlowNode {
  id: string;
  type: string;
  name: string;
  config: Record<string, any>;
  // n8n-style position as [x, y] array
  position?: Position;
}

// Utility type for ReactFlow position format
export interface ReactFlowPosition {
  x: number;
  y: number;
}

export interface Connection {
  id: string;
  source: string;
  sourceOutput?: string;
  target: string;
  targetInput?: string;
}

export interface SampleFlow {
  id: string;
  name: string;
  description: string;
  category: 'basic' | 'iot' | 'integration' | 'automation';
  flow: Flow;
  thumbnail?: string;
}
