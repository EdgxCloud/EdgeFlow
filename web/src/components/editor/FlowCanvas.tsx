import { useCallback, useState, useEffect } from 'react'
import {
  ReactFlow,
  Background,
  Controls,
  MiniMap,
  addEdge,
  useNodesState,
  useEdgesState,
  Connection,
  Edge,
  Node,
  NodeTypes,
  SelectionMode,
  Panel,
} from '@xyflow/react'
import '@xyflow/react/dist/style.css'
import CustomNode from './CustomNode'
import { MousePointer2, BoxSelect } from 'lucide-react'
import { NodeConfigDialog } from '../NodeConfig/NodeConfigDialog'
import { useFlowStore } from '../../stores/flowStore'
import { toObjectPosition, toArrayPosition } from '@/utils/position'

const nodeTypes: NodeTypes = {
  custom: CustomNode,
}

interface FlowCanvasProps {
  flowId?: string
  onNodeSelect?: (nodeId: string | null, nodeName?: string) => void
}

export default function FlowCanvas({ flowId, onNodeSelect }: FlowCanvasProps) {
  const { currentFlow, updateFlow } = useFlowStore()
  const [nodes, setNodes, onNodesChange] = useNodesState([])
  const [edges, setEdges, onEdgesChange] = useEdgesState([])
  const [reactFlowInstance, setReactFlowInstance] = useState<any>(null)
  const [selectedNode, setSelectedNode] = useState<Node | null>(null)

  // Selection mode: 'pan' for normal drag-to-pan, 'select' for marquee selection
  const [isSelectionMode, setIsSelectionMode] = useState(false)

  // Load nodes from currentFlow when it changes
  // Track if we've loaded this specific flow to avoid re-loading on user edits
  const [loadedFlowId, setLoadedFlowId] = useState<string | null>(null)

  // Reset loadedFlowId when flowId prop changes (navigating to different flow)
  useEffect(() => {
    if (flowId !== loadedFlowId) {
      console.log('FlowId prop changed, resetting loadedFlowId')
      setLoadedFlowId(null)
    }
  }, [flowId])

  useEffect(() => {
    // Only load if we have a currentFlow that matches our flowId and we haven't loaded it yet
    // This prevents re-loading when user edits nodes (which updates currentFlow)

    // Handle nodes as both array and object formats
    const nodesData = currentFlow?.nodes
    const nodeArray = nodesData
      ? (Array.isArray(nodesData) ? nodesData : Object.values(nodesData))
      : []

    const flowHasNodes = nodeArray.length > 0
    const isNewFlow = currentFlow?.id !== loadedFlowId

    console.log('FlowCanvas useEffect:', {
      flowId,
      currentFlowId: currentFlow?.id,
      loadedFlowId,
      flowHasNodes,
      isNewFlow,
      nodesIsArray: Array.isArray(nodesData),
      nodeCount: nodeArray.length,
      connectionsCount: currentFlow?.connections?.length || 0
    })

    if (currentFlow && flowHasNodes && isNewFlow) {
      console.log('Loading flow into canvas:', currentFlow.id, 'with', nodeArray.length, 'nodes')

      const flowNodes: Node[] = nodeArray.map((node: any, index) => ({
        id: node.id,
        type: 'custom',
        // Convert n8n-style [x, y] array to ReactFlow {x, y} object format
        // Use stored position if available, otherwise calculate grid position
        position: node.position
          ? toObjectPosition(node.position)
          : { x: 100 + (index % 3) * 300, y: 100 + Math.floor(index / 3) * 150 },
        data: {
          label: node.name || node.type,
          nodeType: node.type,
          config: node.config || {},
        },
      }))
      setNodes(flowNodes)

      // Convert connections to edges
      const connections = currentFlow.connections || []
      console.log('Loading connections:', connections.length, connections)

      const flowEdges: Edge[] = connections.map((conn: any, index) => ({
        id: conn.id || `edge-${conn.source}-${conn.target}-${index}`,
        source: conn.source,
        target: conn.target,
        // Don't set sourceHandle/targetHandle - nodes use default handles without IDs
        type: 'default',
        animated: false,
        style: { stroke: '#3b82f6', strokeWidth: 2 },
      }))
      setEdges(flowEdges)

      console.log('Loaded', flowNodes.length, 'nodes and', flowEdges.length, 'edges')

      // Mark this flow as loaded
      setLoadedFlowId(currentFlow.id)
    }
  }, [currentFlow?.id, currentFlow?.nodes, currentFlow?.connections, flowId]) // Re-run when currentFlow changes

  const onConnect = useCallback(
    (params: Connection) => setEdges((eds) => addEdge(params, eds)),
    [setEdges]
  )

  const onDrop = useCallback(
    (event: React.DragEvent) => {
      event.preventDefault()

      if (!reactFlowInstance) return

      const type = event.dataTransfer.getData('application/reactflow')
      if (!type) return

      const position = reactFlowInstance.screenToFlowPosition({
        x: event.clientX,
        y: event.clientY,
      })

      const newNode: Node = {
        id: `node-${Date.now()}`,
        type: 'custom',
        position,
        data: {
          label: type,
          nodeType: type,
          config: {},
        },
      }

      setNodes((nds) => nds.concat(newNode))
    },
    [reactFlowInstance, setNodes]
  )

  const onDragOver = useCallback((event: React.DragEvent) => {
    event.preventDefault()
    event.dataTransfer.dropEffect = 'move'
  }, [])

  const handleSave = async () => {
    if (!currentFlow || !flowId) return

    // Convert ReactFlow nodes to EdgeFlow format with n8n-style [x, y] positions (as array)
    const flowNodes = nodes.map((node) => ({
      id: node.id,
      type: node.data.nodeType || 'unknown',
      name: node.data.label || node.data.nodeType,
      config: node.data.config || {},
      position: toArrayPosition(node.position), // Store as n8n-style [x, y] array
    }))

    // Convert ReactFlow edges to EdgeFlow connections
    const connections = edges.map((edge) => ({
      id: edge.id,
      source: edge.source,
      target: edge.target,
      sourceOutput: edge.sourceHandle ? parseInt(edge.sourceHandle) : 0,
    }))

    await updateFlow(flowId, {
      nodes: flowNodes,
      connections: connections,
    })
  }

  // Track if initial load sync has happened (to avoid overwriting connections on first render)
  const [initialSyncDone, setInitialSyncDone] = useState(false)

  // Mark initial sync as done after edges are loaded
  useEffect(() => {
    if (loadedFlowId && edges.length > 0 && !initialSyncDone) {
      setInitialSyncDone(true)
      console.log('Initial sync marked as done, edges loaded:', edges.length)
    }
  }, [loadedFlowId, edges.length, initialSyncDone])

  // Reset initialSyncDone when switching flows
  useEffect(() => {
    if (flowId !== loadedFlowId) {
      setInitialSyncDone(false)
    }
  }, [flowId, loadedFlowId])

  // Sync nodes/edges to currentFlow whenever they change
  // Only sync after initial load is complete AND edges have been loaded
  useEffect(() => {
    if (currentFlow && flowId && loadedFlowId && initialSyncDone && nodes.length > 0) {
      // Convert nodes to EdgeFlow format with n8n-style [x, y] positions (as array)
      const flowNodes = nodes.map((node) => ({
        id: node.id,
        type: node.data.nodeType || 'unknown',
        name: node.data.label || node.data.nodeType,
        config: node.data.config || {},
        position: toArrayPosition(node.position), // Store as n8n-style [x, y] array
      }))

      // Convert edges to connections
      const connections = edges.map((edge) => ({
        id: edge.id,
        source: edge.source,
        target: edge.target,
        sourceOutput: edge.sourceHandle ? parseInt(edge.sourceHandle) : 0,
      }))

      // Update currentFlow directly (mutations are OK with Zustand)
      currentFlow.nodes = flowNodes
      currentFlow.connections = connections

      console.log('Canvas synced to store:', flowNodes.length, 'nodes,', connections.length, 'connections')
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [nodes, edges, flowId, loadedFlowId, initialSyncDone]) // Don't include currentFlow to avoid infinite loop

  const onNodeClick = useCallback((_event: React.MouseEvent, node: Node) => {
    // Notify parent component of node selection for execution data display
    if (onNodeSelect) {
      onNodeSelect(node.id, node.data.label || node.data.nodeType)
    }
  }, [onNodeSelect])

  const onNodeDoubleClick = useCallback((_event: React.MouseEvent, node: Node) => {
    setSelectedNode(node)
  }, [])

  const handleNodeSettingsClose = useCallback(() => {
    setSelectedNode(null)
  }, [])

  const handleNodeSettingsSave = useCallback((nodeId: string, data: any) => {
    setNodes((nds) =>
      nds.map((node) => {
        if (node.id === nodeId) {
          return { ...node, data: { ...node.data, ...data } }
        }
        return node
      })
    )
    setSelectedNode(null)
  }, [setNodes])

  const defaultEdgeOptions = {
    type: 'default',
    animated: false,
    style: { stroke: '#3b82f6', strokeWidth: 2 },
  }

  // Handle keyboard shortcut for selection mode (hold Shift)
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Shift' && !e.repeat) {
        setIsSelectionMode(true)
      }
    }
    const handleKeyUp = (e: KeyboardEvent) => {
      if (e.key === 'Shift') {
        setIsSelectionMode(false)
      }
    }
    window.addEventListener('keydown', handleKeyDown)
    window.addEventListener('keyup', handleKeyUp)
    return () => {
      window.removeEventListener('keydown', handleKeyDown)
      window.removeEventListener('keyup', handleKeyUp)
    }
  }, [])

  return (
    <div className="h-full w-full">
      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onConnect={onConnect}
        onInit={setReactFlowInstance}
        onDrop={onDrop}
        onDragOver={onDragOver}
        onNodeClick={onNodeClick}
        onNodeDoubleClick={onNodeDoubleClick}
        nodeTypes={nodeTypes}
        defaultEdgeOptions={defaultEdgeOptions}
        fitView
        className="bg-gray-50 dark:bg-gray-900"
        // Enable marquee selection when in selection mode
        selectionOnDrag={isSelectionMode}
        selectionMode={SelectionMode.Partial}
        panOnDrag={!isSelectionMode}
        // Allow multi-selection with Ctrl/Cmd click
        multiSelectionKeyCode="Control"
        selectionKeyCode={null}
      >
        <Background />
        <Controls />
        <MiniMap
          className="bg-white dark:bg-gray-800"
          maskColor="rgba(0, 0, 0, 0.1)"
        />

        {/* Selection Mode Toggle Panel */}
        <Panel position="top-left" className="flex gap-1">
          <button
            onClick={() => setIsSelectionMode(false)}
            className={`p-2 rounded-lg border transition-all ${
              !isSelectionMode
                ? 'bg-blue-500 text-white border-blue-500 shadow-lg'
                : 'bg-white dark:bg-gray-800 text-gray-600 dark:text-gray-300 border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700'
            }`}
            title="Pan Mode (Default)"
          >
            <MousePointer2 className="w-4 h-4" />
          </button>
          <button
            onClick={() => setIsSelectionMode(true)}
            className={`p-2 rounded-lg border transition-all ${
              isSelectionMode
                ? 'bg-blue-500 text-white border-blue-500 shadow-lg'
                : 'bg-white dark:bg-gray-800 text-gray-600 dark:text-gray-300 border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700'
            }`}
            title="Selection Mode (Hold Shift)"
          >
            <BoxSelect className="w-4 h-4" />
          </button>
          {isSelectionMode && (
            <span className="ml-2 px-2 py-1 text-xs bg-blue-100 dark:bg-blue-900 text-blue-700 dark:text-blue-300 rounded-md flex items-center">
              Drag to select multiple nodes
            </span>
          )}
        </Panel>
      </ReactFlow>

      {/* Node Settings Panel */}
      <NodeConfigDialog
        node={selectedNode ? {
          id: selectedNode.id,
          type: selectedNode.data.nodeType,
          name: selectedNode.data.label,
          config: selectedNode.data.config || {}
        } : null}
        flowId={flowId}
        onClose={handleNodeSettingsClose}
        onSave={handleNodeSettingsSave}
      />
    </div>
  )
}
