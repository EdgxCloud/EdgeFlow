import { useCallback, useState, useEffect, forwardRef, useImperativeHandle } from 'react'
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
  useReactFlow,
  XYPosition,
} from '@xyflow/react'
import '@xyflow/react/dist/style.css'
import CustomNode from './CustomNode'
import { NodeConfigDialog } from '../NodeConfig/NodeConfigDialog'
import { useFlowStore } from '../../stores/flowStore'
import { toObjectPosition, toArrayPosition } from '@/utils/position'
import { importFlow } from '@/utils/flowImportExport'
import { useCopyPaste, useKeyboardShortcuts } from '@/hooks/useCopyPaste'
import { useUndoRedo } from '@/hooks/useUndoRedo'
import { toast } from 'sonner'
import CanvasContextMenu, { ContextMenuState } from './CanvasContextMenu'

const nodeTypes: NodeTypes = {
  custom: CustomNode,
}

export interface FlowCanvasRef {
  handleCopy: () => void
  handlePaste: () => void
  handleCut: () => void
  handleDelete: () => void
  handleDuplicate: () => void
  handleUndo: () => void
  handleRedo: () => void
  handleSelectAll: () => void
  getFlowData: () => { nodes: any[]; connections: any[] }
  loadFlowData: (nodes: any[], connections: any[]) => void
  canUndo: boolean
  canRedo: boolean
  canPaste: boolean
  hasSelection: boolean
}

interface FlowCanvasProps {
  flowId?: string
  onNodeSelect?: (nodeId: string | null, nodeName?: string) => void
}

const FlowCanvas = forwardRef<FlowCanvasRef, FlowCanvasProps>(({ flowId, onNodeSelect }, ref) => {
  const { currentFlow, updateFlow } = useFlowStore()
  const { getNodes, getEdges, screenToFlowPosition, fitView } = useReactFlow()
  const [nodes, setNodes, onNodesChange] = useNodesState([])
  const [edges, setEdges, onEdgesChange] = useEdgesState([])
  const [reactFlowInstance, setReactFlowInstance] = useState<any>(null)
  const [selectedNode, setSelectedNode] = useState<Node | null>(null)

  // Context menu state
  const [contextMenu, setContextMenu] = useState<ContextMenuState>({
    visible: false,
    x: 0,
    y: 0,
    type: 'pane',
  })

  // Copy/Paste functionality
  const { copy, paste, cut, canPaste } = useCopyPaste()

  // Undo/Redo functionality
  const [historyState, undoRedoActions] = useUndoRedo({ nodes, edges })

  // Track if we've loaded this specific flow
  const [loadedFlowId, setLoadedFlowId] = useState<string | null>(null)

  // Reset loadedFlowId when flowId prop changes
  useEffect(() => {
    if (flowId !== loadedFlowId) {
      setLoadedFlowId(null)
    }
  }, [flowId])

  useEffect(() => {
    const nodesData = currentFlow?.nodes
    const nodeArray = nodesData
      ? (Array.isArray(nodesData) ? nodesData : Object.values(nodesData))
      : []

    const flowHasNodes = nodeArray.length > 0
    const isNewFlow = currentFlow?.id !== loadedFlowId

    if (currentFlow && flowHasNodes && isNewFlow) {
      const flowNodes: Node[] = nodeArray.map((node: any, index) => ({
        id: node.id,
        type: 'custom',
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

      const connections = currentFlow.connections || []
      const flowEdges: Edge[] = connections.map((conn: any, index) => ({
        id: conn.id || `edge-${conn.source}-${conn.target}-${index}`,
        source: conn.source,
        target: conn.target,
        type: 'default',
        animated: false,
        style: { stroke: '#3b82f6', strokeWidth: 2 },
      }))
      setEdges(flowEdges)

      setLoadedFlowId(currentFlow.id)
    }
  }, [currentFlow?.id, currentFlow?.nodes, currentFlow?.connections, flowId])

  const onConnect = useCallback(
    (params: Connection) => setEdges((eds) => addEdge(params, eds)),
    [setEdges]
  )

  const onDrop = useCallback(
    (event: React.DragEvent) => {
      event.preventDefault()

      const instance = reactFlowInstance
      if (!instance) return

      const type = event.dataTransfer.getData('application/reactflow')
      if (!type) return

      const position = instance.screenToFlowPosition({
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
      undoRedoActions.push({ nodes: [...getNodes(), newNode], edges: getEdges() })
    },
    [reactFlowInstance, setNodes, undoRedoActions, getNodes, getEdges]
  )

  const onDragOver = useCallback((event: React.DragEvent) => {
    event.preventDefault()
    event.dataTransfer.dropEffect = 'move'
  }, [])

  // ========== Copy/Paste/Cut/Delete/Duplicate/Undo/Redo handlers ==========

  const handleCopy = useCallback(() => {
    const selectedNodes = nodes.filter(node => node.selected)
    const selectedNodeIds = new Set(selectedNodes.map(n => n.id))
    const selectedEdges = edges.filter(edge =>
      selectedNodeIds.has(edge.source) && selectedNodeIds.has(edge.target)
    )

    if (selectedNodes.length > 0) {
      copy(selectedNodes, selectedEdges)
      toast.success(`Copied ${selectedNodes.length} node${selectedNodes.length > 1 ? 's' : ''}`)
    }
  }, [nodes, edges, copy])

  const handlePaste = useCallback(async (position?: XYPosition) => {
    // First try internal clipboard (copied nodes within the editor)
    const pasted = paste(position)
    if (pasted) {
      setNodes(nds => nds.map(n => ({ ...n, selected: false })))
      setEdges(eds => eds.map(e => ({ ...e, selected: false })))
      setNodes(nds => [...nds, ...pasted.nodes])
      setEdges(eds => [...eds, ...pasted.edges])
      undoRedoActions.push({ nodes: getNodes(), edges: getEdges() })
      toast.success(`Pasted ${pasted.nodes.length} node${pasted.nodes.length > 1 ? 's' : ''}`)
      return
    }

    // If internal clipboard empty, try system clipboard for workflow JSON
    try {
      const clipText = await navigator.clipboard.readText()
      if (!clipText || !clipText.trim().startsWith('{') && !clipText.trim().startsWith('[')) {
        return
      }
      const flows = importFlow(clipText)
      if (flows.length === 0) return

      const imported = flows[0]
      const flowNodes: Node[] = imported.nodes.map((node: any, index) => ({
        id: node.id,
        type: 'custom',
        position: node.position
          ? toObjectPosition(node.position)
          : { x: 100 + (index % 3) * 300, y: 100 + Math.floor(index / 3) * 150 },
        data: {
          label: node.name || node.type,
          nodeType: node.type,
          config: node.config || {},
        },
      }))
      const flowEdges: Edge[] = (imported.connections || []).map((conn: any, index) => ({
        id: conn.id || `edge-${conn.source}-${conn.target}-${index}`,
        source: conn.source,
        target: conn.target,
        type: 'default',
        animated: false,
        style: { stroke: '#3b82f6', strokeWidth: 2 },
      }))

      setNodes(flowNodes)
      setEdges(flowEdges)
      setTimeout(() => fitView({ padding: 0.2 }), 100)
      toast.success(`Pasted workflow: ${flowNodes.length} nodes, ${flowEdges.length} connections`)
    } catch {
      // Clipboard read failed (permissions) or invalid JSON - silently ignore
    }
  }, [paste, setNodes, setEdges, undoRedoActions, getNodes, getEdges, fitView])

  const handleCut = useCallback(() => {
    const selectedNodes = nodes.filter(node => node.selected)
    const selectedNodeIds = new Set(selectedNodes.map(n => n.id))
    const selectedEdges = edges.filter(edge =>
      selectedNodeIds.has(edge.source) && selectedNodeIds.has(edge.target)
    )

    if (selectedNodes.length > 0) {
      cut(selectedNodes, selectedEdges)
      setNodes(nds => nds.filter(n => !n.selected))
      setEdges(eds => eds.filter(e => {
        if (e.selected) return false
        if (selectedNodeIds.has(e.source) || selectedNodeIds.has(e.target)) return false
        return true
      }))
      undoRedoActions.push({ nodes: getNodes(), edges: getEdges() })
      toast.success(`Cut ${selectedNodes.length} node${selectedNodes.length > 1 ? 's' : ''}`)
    }
  }, [nodes, edges, cut, setNodes, setEdges, undoRedoActions, getNodes, getEdges])

  const handleDelete = useCallback(() => {
    const selectedNodes = nodes.filter(n => n.selected)
    const selectedNodeIds = new Set(selectedNodes.map(n => n.id))
    const hasSelection = selectedNodes.length > 0 || edges.some(e => e.selected)

    if (hasSelection) {
      setNodes(nds => nds.filter(n => !n.selected))
      setEdges(eds => eds.filter(e => {
        if (e.selected) return false
        if (selectedNodeIds.has(e.source) || selectedNodeIds.has(e.target)) return false
        return true
      }))
      undoRedoActions.push({ nodes: getNodes(), edges: getEdges() })
      if (selectedNodes.length > 0) {
        toast.success(`Deleted ${selectedNodes.length} node${selectedNodes.length > 1 ? 's' : ''}`)
      }
    }
  }, [nodes, edges, setNodes, setEdges, undoRedoActions, getNodes, getEdges])

  const handleSelectAll = useCallback(() => {
    setNodes(nds => nds.map(n => ({ ...n, selected: true })))
    setEdges(eds => eds.map(e => ({ ...e, selected: true })))
  }, [setNodes, setEdges])

  const handleDuplicate = useCallback(() => {
    const selectedNodes = nodes.filter(node => node.selected)
    const selectedNodeIds = new Set(selectedNodes.map(n => n.id))
    const selectedEdges = edges.filter(edge =>
      selectedNodeIds.has(edge.source) && selectedNodeIds.has(edge.target)
    )

    if (selectedNodes.length > 0) {
      copy(selectedNodes, selectedEdges)
      const pasted = paste()
      if (pasted) {
        setNodes(nds => nds.map(n => ({ ...n, selected: false })))
        setEdges(eds => eds.map(e => ({ ...e, selected: false })))
        setNodes(nds => [...nds, ...pasted.nodes])
        setEdges(eds => [...eds, ...pasted.edges])
        undoRedoActions.push({ nodes: getNodes(), edges: getEdges() })
        toast.success(`Duplicated ${pasted.nodes.length} node${pasted.nodes.length > 1 ? 's' : ''}`)
      }
    }
  }, [nodes, edges, copy, paste, setNodes, setEdges, undoRedoActions, getNodes, getEdges])

  const handleUndo = useCallback(() => {
    if (undoRedoActions.canUndo) {
      undoRedoActions.undo()
      const prevState = historyState as { nodes: Node[], edges: Edge[] }
      if (prevState.nodes) setNodes(prevState.nodes)
      if (prevState.edges) setEdges(prevState.edges)
    }
  }, [undoRedoActions, historyState, setNodes, setEdges])

  const handleRedo = useCallback(() => {
    if (undoRedoActions.canRedo) {
      undoRedoActions.redo()
      const nextState = historyState as { nodes: Node[], edges: Edge[] }
      if (nextState.nodes) setNodes(nextState.nodes)
      if (nextState.edges) setEdges(nextState.edges)
    }
  }, [undoRedoActions, historyState, setNodes, setEdges])

  const hasSelection = nodes.some(n => n.selected)

  // Get current flow data from canvas (for saving)
  const getFlowData = useCallback(() => {
    const flowNodes = nodes.map((node) => ({
      id: node.id,
      type: node.data.nodeType || 'unknown',
      name: node.data.label || node.data.nodeType,
      config: node.data.config || {},
      position: toArrayPosition(node.position),
    }))

    const connections = edges.map((edge) => ({
      id: edge.id,
      source: edge.source,
      target: edge.target,
      sourceOutput: edge.sourceHandle ? parseInt(edge.sourceHandle) : 0,
    }))

    return { nodes: flowNodes, connections }
  }, [nodes, edges])

  // Load flow data onto canvas (for paste/import)
  const loadFlowData = useCallback((importedNodes: any[], importedConnections: any[]) => {
    const flowNodes: Node[] = importedNodes.map((node: any, index) => ({
      id: node.id,
      type: 'custom',
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

    const flowEdges: Edge[] = (importedConnections || []).map((conn: any, index) => ({
      id: conn.id || `edge-${conn.source}-${conn.target}-${index}`,
      source: conn.source,
      target: conn.target,
      type: 'default',
      animated: false,
      style: { stroke: '#3b82f6', strokeWidth: 2 },
    }))
    setEdges(flowEdges)

    setTimeout(() => fitView({ padding: 0.2 }), 100)
  }, [setNodes, setEdges, fitView])

  // Expose handlers to parent via ref
  useImperativeHandle(ref, () => ({
    handleCopy,
    handlePaste: () => handlePaste(),
    handleCut,
    handleDelete,
    handleDuplicate,
    handleUndo,
    handleRedo,
    handleSelectAll,
    getFlowData,
    loadFlowData,
    canUndo: undoRedoActions.canUndo,
    canRedo: undoRedoActions.canRedo,
    canPaste,
    hasSelection,
  }), [handleCopy, handlePaste, handleCut, handleDelete, handleDuplicate, handleUndo, handleRedo, handleSelectAll, getFlowData, loadFlowData, undoRedoActions.canUndo, undoRedoActions.canRedo, canPaste, hasSelection])

  // Keyboard shortcuts
  const handleKeyDown = useKeyboardShortcuts(
    handleCopy,
    () => handlePaste(),
    handleCut,
    handleDelete,
    handleUndo,
    handleRedo,
    handleSelectAll,
    handleDuplicate
  )

  useEffect(() => {
    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [handleKeyDown])

  // ========== Save & Sync ==========

  const handleSave = async () => {
    if (!currentFlow || !flowId) return

    const flowNodes = nodes.map((node) => ({
      id: node.id,
      type: node.data.nodeType || 'unknown',
      name: node.data.label || node.data.nodeType,
      config: node.data.config || {},
      position: toArrayPosition(node.position),
    }))

    const connections = edges.map((edge) => ({
      id: edge.id,
      source: edge.source,
      target: edge.target,
      sourceOutput: edge.sourceHandle ? parseInt(edge.sourceHandle) : 0,
    }))

    await updateFlow(flowId, { nodes: flowNodes, connections })
  }

  // Track if initial load sync has happened
  const [initialSyncDone, setInitialSyncDone] = useState(false)

  useEffect(() => {
    if (loadedFlowId && (nodes.length > 0 || edges.length > 0) && !initialSyncDone) {
      setInitialSyncDone(true)
    }
  }, [loadedFlowId, nodes.length, edges.length, initialSyncDone])

  useEffect(() => {
    if (flowId !== loadedFlowId) {
      setInitialSyncDone(false)
    }
  }, [flowId, loadedFlowId])

  // Sync nodes/edges to currentFlow
  useEffect(() => {
    if (currentFlow && flowId && loadedFlowId && initialSyncDone && nodes.length > 0) {
      const flowNodes = nodes.map((node) => ({
        id: node.id,
        type: node.data.nodeType || 'unknown',
        name: node.data.label || node.data.nodeType,
        config: node.data.config || {},
        position: toArrayPosition(node.position),
      }))

      const connections = edges.map((edge) => ({
        id: edge.id,
        source: edge.source,
        target: edge.target,
        sourceOutput: edge.sourceHandle ? parseInt(edge.sourceHandle) : 0,
      }))

      currentFlow.nodes = flowNodes
      currentFlow.connections = connections
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [nodes, edges, flowId, loadedFlowId, initialSyncDone])

  // ========== Node events ==========

  const onNodeClick = useCallback((_event: React.MouseEvent, node: Node) => {
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
    undoRedoActions.push({ nodes: getNodes(), edges: getEdges() })
  }, [setNodes, undoRedoActions, getNodes, getEdges])

  // ========== Context menu ==========

  const onPaneContextMenu = useCallback((event: React.MouseEvent) => {
    event.preventDefault()
    setContextMenu({
      visible: true,
      x: event.clientX,
      y: event.clientY,
      type: 'pane',
      flowPosition: screenToFlowPosition({ x: event.clientX, y: event.clientY }),
    })
  }, [screenToFlowPosition])

  const onNodeContextMenu = useCallback((event: React.MouseEvent, node: Node) => {
    event.preventDefault()
    // Select the right-clicked node if not already selected
    if (!node.selected) {
      setNodes(nds => nds.map(n => ({ ...n, selected: n.id === node.id })))
    }
    setContextMenu({
      visible: true,
      x: event.clientX,
      y: event.clientY,
      type: hasSelection && nodes.filter(n => n.selected).length > 1 ? 'selection' : 'node',
      nodeId: node.id,
    })
  }, [hasSelection, nodes, setNodes])

  const closeContextMenu = useCallback(() => {
    setContextMenu(prev => ({ ...prev, visible: false }))
  }, [])

  const handleContextMenuAction = useCallback((action: string) => {
    switch (action) {
      case 'copy':
        handleCopy()
        break
      case 'paste':
        handlePaste(contextMenu.flowPosition)
        break
      case 'cut':
        handleCut()
        break
      case 'duplicate':
        handleDuplicate()
        break
      case 'delete':
        handleDelete()
        break
      case 'selectAll':
        handleSelectAll()
        break
      case 'fitView':
        fitView({ padding: 0.2, duration: 300 })
        break
      case 'configure':
        if (contextMenu.nodeId) {
          const node = nodes.find(n => n.id === contextMenu.nodeId)
          if (node) setSelectedNode(node)
        }
        break
    }
    closeContextMenu()
  }, [handleCopy, handlePaste, handleCut, handleDuplicate, handleDelete, handleSelectAll, fitView, contextMenu, nodes, closeContextMenu])

  const defaultEdgeOptions = {
    type: 'default',
    animated: false,
    style: { stroke: '#3b82f6', strokeWidth: 2 },
  }

  return (
    <div className="h-full w-full relative">
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
        onPaneContextMenu={onPaneContextMenu}
        onNodeContextMenu={onNodeContextMenu}
        onPaneClick={closeContextMenu}
        nodeTypes={nodeTypes}
        defaultEdgeOptions={defaultEdgeOptions}
        fitView
        snapToGrid
        snapGrid={[15, 15]}
        className="bg-gray-50 dark:bg-gray-900"
        selectionOnDrag
        selectionMode={SelectionMode.Partial}
        panOnDrag={[1, 2]}
        multiSelectionKeyCode="Control"
        selectionKeyCode={null}
        deleteKeyCode={null}
        minZoom={0.1}
        maxZoom={4}
      >
        <Background />
        <Controls />
        <MiniMap
          className="bg-white dark:bg-gray-800"
          maskColor="rgba(0, 0, 0, 0.1)"
        />

      </ReactFlow>

      {/* Context Menu */}
      <CanvasContextMenu
        state={contextMenu}
        canPaste={canPaste}
        onAction={handleContextMenuAction}
        onClose={closeContextMenu}
      />

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
})

FlowCanvas.displayName = 'FlowCanvas'

export default FlowCanvas
