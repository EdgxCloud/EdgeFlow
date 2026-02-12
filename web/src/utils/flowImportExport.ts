import { Flow, FlowNode } from '@/types/flow';
import { toArrayPosition, isArrayPosition, isObjectPosition, AnyPosition } from './position';

// Legacy edge type for backward compatibility
interface FlowEdge {
  id: string;
  source: string;
  target: string;
  sourceHandle?: string;
  targetHandle?: string;
}

export interface ExportedFlow {
  version: string;
  flows: Flow[];
  metadata?: {
    exportedAt: string;
    exportedBy: string;
    edgeflowVersion: string;
  };
}

export function exportFlow(flow: Flow): string {
  const exported: ExportedFlow = {
    version: '1.0',
    flows: [flow],
    metadata: {
      exportedAt: new Date().toISOString(),
      exportedBy: 'EdgeFlow',
      edgeflowVersion: '1.0.0',
    },
  };

  return JSON.stringify(exported, null, 2);
}

export function exportFlows(flows: Flow[]): string {
  const exported: ExportedFlow = {
    version: '1.0',
    flows,
    metadata: {
      exportedAt: new Date().toISOString(),
      exportedBy: 'EdgeFlow',
      edgeflowVersion: '1.0.0',
    },
  };

  return JSON.stringify(exported, null, 2);
}

export function importFlow(jsonString: string): Flow[] {
  try {
    const data = JSON.parse(jsonString);

    if (!data.version) {
      throw new Error('Invalid flow format: missing version');
    }

    if (!Array.isArray(data.flows)) {
      throw new Error('Invalid flow format: flows must be an array');
    }

    const flows = data.flows.map((flow: any) => validateFlow(flow));

    return flows;
  } catch (error) {
    if (error instanceof SyntaxError) {
      throw new Error('Invalid JSON format');
    }
    throw error;
  }
}

function validateFlow(flow: any): Flow {
  if (!flow.id || typeof flow.id !== 'string') {
    throw new Error('Invalid flow: missing or invalid id');
  }

  if (!flow.name || typeof flow.name !== 'string') {
    throw new Error('Invalid flow: missing or invalid name');
  }

  if (!Array.isArray(flow.nodes)) {
    throw new Error('Invalid flow: nodes must be an array');
  }

  // Accept both "edges" and "connections" field names
  const rawEdges = Array.isArray(flow.edges) ? flow.edges : Array.isArray(flow.connections) ? flow.connections : [];

  const validatedNodes = flow.nodes.map((node: any) => validateNode(node));
  const validatedConnections = rawEdges.map((edge: any) => validateEdge(edge));

  return {
    id: flow.id,
    name: flow.name,
    description: flow.description || '',
    nodes: validatedNodes,
    connections: validatedConnections,
    enabled: flow.enabled !== false,
    created: flow.created || new Date().toISOString(),
    modified: flow.modified || new Date().toISOString(),
  };
}

function validateNode(node: any): FlowNode {
  if (!node.id || typeof node.id !== 'string') {
    throw new Error('Invalid node: missing or invalid id');
  }

  if (!node.type || typeof node.type !== 'string') {
    throw new Error('Invalid node: missing or invalid type');
  }

  // Convert position to n8n-style [x, y] array format
  // Handles both legacy {x, y} object and new [x, y] array formats
  let position: [number, number] = [0, 0];
  if (node.position) {
    if (isArrayPosition(node.position)) {
      position = node.position;
    } else if (isObjectPosition(node.position)) {
      position = toArrayPosition(node.position);
    }
  }

  return {
    id: node.id,
    type: node.type,
    name: node.name || node.type,
    position,
    config: node.config || node.data || {},
  };
}

function validateEdge(edge: any): FlowEdge {
  if (!edge.id || typeof edge.id !== 'string') {
    throw new Error('Invalid edge: missing or invalid id');
  }

  if (!edge.source || typeof edge.source !== 'string') {
    throw new Error('Invalid edge: missing or invalid source');
  }

  if (!edge.target || typeof edge.target !== 'string') {
    throw new Error('Invalid edge: missing or invalid target');
  }

  return {
    id: edge.id,
    source: edge.source,
    target: edge.target,
    sourceHandle: edge.sourceHandle,
    targetHandle: edge.targetHandle,
  };
}

export function mergeFlows(existingFlows: Flow[], importedFlows: Flow[]): Flow[] {
  const merged = [...existingFlows];
  const existingIds = new Set(existingFlows.map((f) => f.id));

  for (const imported of importedFlows) {
    if (existingIds.has(imported.id)) {
      const suffix = Date.now().toString().slice(-6);
      imported.id = `${imported.id}-${suffix}`;
      imported.name = `${imported.name} (imported)`;
    }
    merged.push(imported);
  }

  return merged;
}

export async function copyFlowToClipboard(flow: Flow): Promise<void> {
  const json = exportFlow(flow);
  await navigator.clipboard.writeText(json);
}

export function downloadFlow(flow: Flow) {
  const json = exportFlow(flow);
  const blob = new Blob([json], { type: 'application/json' });
  const url = URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.download = `${flow.name.replace(/\s+/g, '_')}.json`;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  URL.revokeObjectURL(url);
}

export function downloadFlows(flows: Flow[]) {
  const json = exportFlows(flows);
  const blob = new Blob([json], { type: 'application/json' });
  const url = URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.download = `flows_${new Date().toISOString().split('T')[0]}.json`;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  URL.revokeObjectURL(url);
}
