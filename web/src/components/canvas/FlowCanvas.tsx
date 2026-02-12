import { useCallback, useRef, useState, useEffect } from 'react'
import {
  ReactFlow,
  Background,
  Controls,
  MiniMap,
  Node,
  Edge,
  Connection,
  addEdge,
  useNodesState,
  useEdgesState,
  NodeTypes,
  ConnectionMode,
  Panel,
  useReactFlow,
  SelectionMode,
} from '@xyflow/react'
import '@xyflow/react/dist/style.css'
import CustomNode from './CustomNode'
import { NodeConfigDialog } from '../NodeConfig/NodeConfigDialog'
import { Button } from '@/components/ui/button'
import { ZoomIn, ZoomOut, Maximize2, Save, Undo, Redo, Copy, Scissors, Rocket } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useCopyPaste, useKeyboardShortcuts } from '@/hooks/useCopyPaste'
import { useUndoRedo } from '@/hooks/useUndoRedo'
import { flowsApi } from '@/services/flows'
import { useFlowStatus, useNodeStatus } from '@/hooks/useWebSocket'
import { toast } from 'sonner'
import DeployDialog from '../dialogs/DeployDialog'
import { toObjectPosition, toArrayPosition, isArrayPosition } from '@/utils/position'

console.log('🚀 FlowCanvas.tsx FILE LOADED')
console.log('🚀 CustomNode imported:', CustomNode)

const nodeTypes: NodeTypes = {
  custom: CustomNode,
}

console.log('🚀 nodeTypes registered:', nodeTypes)

// Edge validation rules
const isValidConnection = (connection: Connection): boolean => {
  // Prevent self-connections
  if (connection.source === connection.target) {
    return false
  }

  // You can add more validation logic here
  // e.g., type checking, max connections, etc.

  return true
}

interface FlowCanvasProps {
  flowId?: string
  flowName?: string
  className?: string
}

export default function FlowCanvas({ flowId, flowName = 'Untitled Flow', className }: FlowCanvasProps) {
  const reactFlowWrapper = useRef<HTMLDivElement>(null)
  const { screenToFlowPosition, fitView, zoomIn, zoomOut, getNodes, getEdges } = useReactFlow()

  const [nodes, setNodes, onNodesChange] = useNodesState([])
  const [edges, setEdges, onEdgesChange] = useEdgesState([])
  const [selectedNode, setSelectedNode] = useState<Node | null>(null)
  const [currentFlowId, setCurrentFlowId] = useState<string | undefined>(flowId)
  const [isSaving, setIsSaving] = useState(false)
  const [isLoading, setIsLoading] = useState(false)
  const [isDeployDialogOpen, setIsDeployDialogOpen] = useState(false)

  // Copy/Paste functionality
  const { copy, paste, cut, canPaste } = useCopyPaste()

  // Undo/Redo functionality
  const [historyState, undoRedoActions] = useUndoRedo({ nodes, edges })

  // WebSocket real-time updates
  useFlowStatus(currentFlowId || '', (message) => {
    console.log('📡 Flow status update:', message)
    toast.info(`Flow ${message.status}`, {
      description: message.message,
    })
  })

  useNodeStatus(currentFlowId || '', (message) => {
    console.log('📡 Node status update:', message)
    // Update node visual status
    setNodes(nds => nds.map(node => {
      if (node.id === message.node_id) {
        return {
          ...node,
          data: {
            ...node.data,
            status: message.status,
            statusMessage: message.message,
          },
        }
      }
      return node
    }))
  })

  // Debug: Log when component mounts
  console.log('FlowCanvas mounted, nodes:', nodes.length, 'flowId:', currentFlowId)

  const onConnect = useCallback(
    (params: Connection) => {
      if (isValidConnection(params)) {
        setEdges((eds) => addEdge(params, eds))
      }
    },
    [setEdges]
  )

  const onDragOver = useCallback((event: React.DragEvent) => {
    event.preventDefault()
    event.dataTransfer.dropEffect = 'move'
  }, [])

  const onDrop = useCallback(
    (event: React.DragEvent) => {
      event.preventDefault()

      const type = event.dataTransfer.getData('application/reactflow')
      const nodeData = event.dataTransfer.getData('application/json')

      console.log('🎯 onDrop triggered, type:', type, 'nodeData:', nodeData)

      if (!type) {
        console.log('❌ No type found, aborting drop')
        return
      }

      const parsedData = nodeData ? JSON.parse(nodeData) : {}
      const position = screenToFlowPosition({
        x: event.clientX,
        y: event.clientY,
      })

      const newNode: Node = {
        id: `${type}-${Date.now()}`,
        type: 'custom',
        position,
        data: {
          label: parsedData.label || type,
          nodeType: type,
          category: parsedData.category || 'core',
          config: parsedData.config || {},
          ...parsedData,
        },
      }

      console.log('✅ Creating new node:', newNode)
      setNodes((nds) => {
        const updatedNodes = [...nds, newNode]
        console.log('✅ Nodes updated, total:', updatedNodes.length, 'nodes:', updatedNodes)
        return updatedNodes
      })
    },
    [screenToFlowPosition, setNodes]
  )

  const handleSave = useCallback(async () => {
    // Validate flow before saving
    const validation = flowsApi.validate(nodes, edges)
    if (!validation.valid) {
      toast.error('Flow validation failed', {
        description: validation.errors.join(', '),
      })
      return
    }

    // Convert node positions to n8n-style [x, y] array format for storage
    const nodesWithArrayPositions = nodes.map(node => ({
      ...node,
      position: toArrayPosition(node.position),
    }))

    setIsSaving(true)
    try {
      if (currentFlowId) {
        // Update existing flow
        await flowsApi.update(currentFlowId, {
          name: flowName,
          nodes: nodesWithArrayPositions as any,
          edges,
        })
        toast.success('Flow saved successfully')
        console.log('✅ Flow updated:', currentFlowId)
      } else {
        // Create new flow
        const newFlow = await flowsApi.create({
          name: flowName,
          description: 'Created from Flow Editor',
          nodes: nodesWithArrayPositions as any,
          edges,
        })
        setCurrentFlowId(newFlow.id)
        toast.success('Flow created successfully')
        console.log('✅ Flow created:', newFlow.id)
      }
    } catch (error) {
      console.error('❌ Failed to save flow:', error)
      toast.error('Failed to save flow', {
        description: error instanceof Error ? error.message : 'Unknown error',
      })
    } finally {
      setIsSaving(false)
    }
  }, [nodes, edges, currentFlowId, flowName])

  const handleFitView = useCallback(() => {
    fitView({ padding: 0.2, duration: 300 })
  }, [fitView])

  const onNodeDoubleClick = useCallback((_event: React.MouseEvent, node: Node) => {
    console.log('🎯 Node double-clicked:', node.id, node.data)
    alert(`Double-click detected on node: ${node.data?.label || node.id}`)
    setSelectedNode(node)
  }, [])

  const handleNodeSettingsClose = useCallback(() => {
    setSelectedNode(null)
  }, [])

  const handleNodeSettingsSave = useCallback((nodeId: string, config: any) => {
    setNodes((nds) => {
      const updated = nds.map((node) => {
        if (node.id === nodeId) {
          return { ...node, data: { ...node.data, config } }
        }
        return node
      })

      // Auto-save flow to backend so config persists
      if (currentFlowId) {
        const flowNodes = updated.map((node) => ({
          id: node.id,
          type: node.data.nodeType || 'unknown',
          name: node.data.label || node.data.nodeType,
          config: node.data.config || {},
          position: toArrayPosition(node.position),
        }))
        const connections = getEdges().map((edge) => ({
          id: edge.id,
          source: edge.source,
          target: edge.target,
          sourceOutput: edge.sourceHandle ? parseInt(edge.sourceHandle) : 0,
        }))
        flowsApi.update(currentFlowId, {
          nodes: flowNodes as any,
          edges: connections as any,
        })
      }

      return updated
    })
    // Take snapshot for undo/redo
    undoRedoActions.push({ nodes: getNodes(), edges: getEdges() })
  }, [setNodes, undoRedoActions, getNodes, getEdges, currentFlowId])

  // Copy selected nodes and edges
  const handleCopy = useCallback(() => {
    const selectedNodes = nodes.filter(node => node.selected)
    const selectedEdges = edges.filter(edge => edge.selected)

    if (selectedNodes.length > 0) {
      copy(selectedNodes, selectedEdges)
      console.log('📋 Copied', selectedNodes.length, 'nodes')
    }
  }, [nodes, edges, copy])

  // Paste copied nodes
  const handlePaste = useCallback(() => {
    const pasted = paste()

    if (pasted) {
      // Deselect all current nodes/edges
      setNodes(nds => nds.map(n => ({ ...n, selected: false })))
      setEdges(eds => eds.map(e => ({ ...e, selected: false })))

      // Add pasted nodes and edges
      setNodes(nds => [...nds, ...pasted.nodes])
      setEdges(eds => [...eds, ...pasted.edges])

      // Take snapshot for undo/redo
      undoRedoActions.push({ nodes: getNodes(), edges: getEdges() })

      console.log('📌 Pasted', pasted.nodes.length, 'nodes')
    }
  }, [paste, setNodes, setEdges, undoRedoActions, getNodes, getEdges])

  // Cut selected nodes and edges
  const handleCut = useCallback(() => {
    const selectedNodes = nodes.filter(node => node.selected)
    const selectedEdges = edges.filter(edge => edge.selected)

    if (selectedNodes.length > 0) {
      cut(selectedNodes, selectedEdges)

      // Remove selected nodes and edges
      setNodes(nds => nds.filter(n => !n.selected))
      setEdges(eds => eds.filter(e => !e.selected))

      // Take snapshot for undo/redo
      undoRedoActions.push({ nodes: getNodes(), edges: getEdges() })

      console.log('✂️ Cut', selectedNodes.length, 'nodes')
    }
  }, [nodes, edges, cut, setNodes, setEdges, undoRedoActions, getNodes, getEdges])

  // Delete selected nodes and edges
  const handleDelete = useCallback(() => {
    const hasSelection = nodes.some(n => n.selected) || edges.some(e => e.selected)

    if (hasSelection) {
      setNodes(nds => nds.filter(n => !n.selected))
      setEdges(eds => eds.filter(e => !e.selected))

      // Take snapshot for undo/redo
      undoRedoActions.push({ nodes: getNodes(), edges: getEdges() })

      console.log('🗑️ Deleted selected elements')
    }
  }, [nodes, edges, setNodes, setEdges, undoRedoActions, getNodes, getEdges])

  // Select all nodes and edges
  const handleSelectAll = useCallback(() => {
    setNodes(nds => nds.map(n => ({ ...n, selected: true })))
    setEdges(eds => eds.map(e => ({ ...e, selected: true })))
    console.log('✨ Selected all elements')
  }, [setNodes, setEdges])

  // Undo last action
  const handleUndo = useCallback(() => {
    if (undoRedoActions.canUndo) {
      undoRedoActions.undo()
      const prevState = historyState
      setNodes(prevState.nodes)
      setEdges(prevState.edges)
      console.log('⏪ Undo')
    }
  }, [undoRedoActions, historyState, setNodes, setEdges])

  // Redo last undone action
  const handleRedo = useCallback(() => {
    if (undoRedoActions.canRedo) {
      undoRedoActions.redo()
      const nextState = historyState
      setNodes(nextState.nodes)
      setEdges(nextState.edges)
      console.log('⏩ Redo')
    }
  }, [undoRedoActions, historyState, setNodes, setEdges])

  // Keyboard shortcuts
  const handleKeyDown = useKeyboardShortcuts(
    handleCopy,
    handlePaste,
    handleCut,
    handleDelete,
    handleUndo,
    handleRedo,
    handleSelectAll
  )

  // Attach keyboard shortcuts
  useEffect(() => {
    document.addEventListener('keydown', handleKeyDown)
    return () => {
      document.removeEventListener('keydown', handleKeyDown)
    }
  }, [handleKeyDown])

  // Load flow from backend on mount
  useEffect(() => {
    const loadFlow = async () => {
      if (currentFlowId) {
        setIsLoading(true)
        try {
          const flow = await flowsApi.get(currentFlowId) as any
          console.log('📥 Loading flow:', flow)

          // Convert nodes: backend returns map, ReactFlow needs array
          // Also convert n8n-style [x, y] array positions to ReactFlow {x, y} object format
          const nodeArray = flow.nodes
            ? (Array.isArray(flow.nodes)
                ? flow.nodes
                : Object.values(flow.nodes))
            : []

          const nodesWithObjectPositions = nodeArray.map((node: any) => ({
            id: node.id,
            type: 'custom',
            position: toObjectPosition(node.position as any),
            data: {
              label: node.name || node.type,
              nodeType: node.type,
              category: node.category || 'core',
              config: node.config || {},
            },
          }))

          // Convert connections to ReactFlow edges format
          // Backend returns 'connections', ReactFlow expects 'edges' with id field
          const connections = flow.connections || flow.edges || []
          const edgesWithIds = connections.map((conn: any, index: number) => ({
            id: conn.id || `e-${conn.source}-${conn.target}-${index}`,
            source: conn.source,
            target: conn.target,
            sourceHandle: conn.sourceOutput?.toString() || conn.sourceHandle,
            targetHandle: conn.targetInput?.toString() || conn.targetHandle,
            type: 'default',
            animated: false,
            style: { stroke: '#3b82f6', strokeWidth: 2 },
          }))

          setNodes(nodesWithObjectPositions)
          setEdges(edgesWithIds)
          toast.success('Flow loaded successfully')
        } catch (error) {
          console.error('❌ Failed to load flow:', error)
          toast.error('Failed to load flow', {
            description: error instanceof Error ? error.message : 'Unknown error',
          })
        } finally {
          setIsLoading(false)
        }
      }
    }

    loadFlow()
  }, [currentFlowId]) // eslint-disable-line react-hooks/exhaustive-deps

  // Take snapshot on node/edge changes for undo/redo
  useEffect(() => {
    if (nodes.length > 0 || edges.length > 0) {
      undoRedoActions.push({ nodes, edges })
    }
  }, [nodes, edges]) // eslint-disable-line react-hooks/exhaustive-deps

  return (
    <div ref={reactFlowWrapper} className={cn('w-full h-full', className)}>
      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={onNodesChange}
        onEdgesChange={onEdgesChange}
        onConnect={onConnect}
        onDragOver={onDragOver}
        onDrop={onDrop}
        onNodeDoubleClick={onNodeDoubleClick}
        onNodeClick={(event, node) => console.log('⚡ ReactFlow onNodeClick:', node.id)}
        nodeTypes={nodeTypes}
        connectionMode={ConnectionMode.Loose}
        isValidConnection={isValidConnection}
        fitView
        snapToGrid
        snapGrid={[15, 15]}
        nodesDraggable={true}
        nodesConnectable={true}
        elementsSelectable={true}
        selectNodesOnDrag={false}
        selectionOnDrag={true}
        selectionMode={SelectionMode.Partial}
        multiSelectionKeyCode="Shift"
        panOnDrag={[1, 2]}
        minZoom={0.5}
        maxZoom={2}
        nodeClickDistance={0}
        defaultEdgeOptions={{
          animated: true,
          style: { strokeWidth: 2 },
        }}
        className="bg-background"
      >
        <Background
          gap={15}
          size={1}
          className="bg-muted/30"
        />

        <Controls
          className="bg-card border border-border rounded-lg shadow-sm"
          showInteractive={false}
        />

        <MiniMap
          className="bg-card border border-border rounded-lg shadow-sm"
          nodeColor={(node) => {
            const category = (node.data?.category as string) || 'core'
            const colors: Record<string, string> = {
              core: '#3B82F6',
              input: '#3B82F6',
              output: '#10B981',
              function: '#8B5CF6',
              hardware: '#F59E0B',
              network: '#06B6D4',
              storage: '#6366F1',
            }
            return colors[category] || '#6B7280'
          }}
          maskColor="rgba(0, 0, 0, 0.1)"
        />

        {/* Custom Control Panel */}
        <Panel position="top-right" className="flex items-center gap-2">
          {/* Undo/Redo */}
          <Button
            variant="secondary"
            size="sm"
            onClick={handleUndo}
            disabled={!undoRedoActions.canUndo}
            title="Undo (Ctrl+Z)"
          >
            <Undo className="h-4 w-4" />
          </Button>
          <Button
            variant="secondary"
            size="sm"
            onClick={handleRedo}
            disabled={!undoRedoActions.canRedo}
            title="Redo (Ctrl+Y)"
          >
            <Redo className="h-4 w-4" />
          </Button>

          <div className="w-px h-6 bg-border" />

          {/* Copy/Cut/Paste */}
          <Button
            variant="secondary"
            size="sm"
            onClick={handleCopy}
            disabled={!nodes.some(n => n.selected)}
            title="Copy (Ctrl+C)"
          >
            <Copy className="h-4 w-4" />
          </Button>
          <Button
            variant="secondary"
            size="sm"
            onClick={handleCut}
            disabled={!nodes.some(n => n.selected)}
            title="Cut (Ctrl+X)"
          >
            <Scissors className="h-4 w-4" />
          </Button>
          <Button
            variant="secondary"
            size="sm"
            onClick={handlePaste}
            disabled={!canPaste}
            title="Paste (Ctrl+V)"
          >
            <Copy className="h-4 w-4" />
          </Button>

          <div className="w-px h-6 bg-border" />

          {/* Zoom controls */}
          <Button
            variant="secondary"
            size="sm"
            onClick={() => zoomIn({ duration: 300 })}
            title="Zoom In"
          >
            <ZoomIn className="h-4 w-4" />
          </Button>
          <Button
            variant="secondary"
            size="sm"
            onClick={() => zoomOut({ duration: 300 })}
            title="Zoom Out"
          >
            <ZoomOut className="h-4 w-4" />
          </Button>
          <Button
            variant="secondary"
            size="sm"
            onClick={handleFitView}
            title="Fit View"
          >
            <Maximize2 className="h-4 w-4" />
          </Button>

          <div className="w-px h-6 bg-border" />

          <Button
            variant="default"
            size="sm"
            onClick={handleSave}
            disabled={isSaving || isLoading}
            title="Save Flow"
          >
            <Save className="h-4 w-4 mr-2" />
            {isSaving ? 'Saving...' : 'Save'}
          </Button>

          <Button
            variant="default"
            size="sm"
            onClick={() => setIsDeployDialogOpen(true)}
            disabled={isLoading}
            title="Deploy Flows"
            className="bg-emerald-600 hover:bg-emerald-700"
          >
            <Rocket className="h-4 w-4 mr-2" />
            Deploy
          </Button>
        </Panel>

        {/* Empty State */}
        {nodes.length === 0 && (
          <Panel position="top-center" className="pointer-events-none">
            <div className="bg-card border border-border rounded-lg shadow-lg p-6 text-center">
              <p className="text-sm text-muted-foreground">
                Drag nodes from the palette to get started
              </p>
            </div>
          </Panel>
        )}
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

      {/* Deploy Dialog */}
      <DeployDialog
        open={isDeployDialogOpen}
        onOpenChange={setIsDeployDialogOpen}
        flowIds={currentFlowId ? [currentFlowId] : undefined}
      />
    </div>
  )
}
