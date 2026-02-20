import { useCallback, useState, useEffect, useRef, forwardRef, useImperativeHandle } from 'react'
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
  custom: CustomNode as any,
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
  markSaving: () => void
  canUndo: boolean
  canRedo: boolean
  canPaste: boolean
  hasSelection: boolean
}

interface FlowCanvasProps {
  flowId?: string
  isRunning?: boolean
  onNodeSelect?: (nodeId: string | null, nodeName?: string) => void
}

const FlowCanvas = forwardRef<FlowCanvasRef, FlowCanvasProps>(({ flowId, isRunning = false, onNodeSelect }, ref) => {
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

  // Track which data source we last loaded from to avoid redundant reloads.
  // We skip reload only if the exact same nodes object reference was already loaded.
  // This lets fetchFlow data through (new reference) while blocking updateFlow echoes.
  const loadedNodesRef = useRef<unknown>(null)
  // Also track if we're the ones who triggered the update (via save)
  const savingRef = useRef(false)

  useEffect(() => {
    if (!currentFlow?.nodes) return

    const nodesData = currentFlow.nodes
    const nodeArray = Array.isArray(nodesData) ? nodesData : Object.values(nodesData)
    if (nodeArray.length === 0) return

    // Skip if this is the response from our own save
    if (savingRef.current) {
      savingRef.current = false
      loadedNodesRef.current = nodesData
      return
    }

    // Skip if we already loaded this exact object (same reference)
    if (nodesData === loadedNodesRef.current) return

    console.log('[FlowCanvas] Loading nodes from store, flow:', currentFlow.id)
    nodeArray.forEach((n: any) => {
      if (n.type === 'inject') {
        console.log('[FlowCanvas] Inject node config:', n.id, JSON.stringify(n.config))
      }
    })

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

    loadedNodesRef.current = nodesData
    console.log('[FlowCanvas] Loaded', flowNodes.length, 'nodes onto canvas')
  }, [currentFlow?.id, currentFlow?.nodes, currentFlow?.connections, flowId, setNodes, setEdges])

  const onConnect = useCallback(
    (params: Connection) => {
      if (isRunning) {
        toast.warning('Stop the workflow before making changes')
        return
      }
      setEdges((eds) => addEdge(params, eds))
    },
    [setEdges, isRunning]
  )

  const onDrop = useCallback(
    (event: React.DragEvent) => {
      event.preventDefault()

      if (isRunning) {
        toast.warning('Stop the workflow before adding nodes')
        return
      }

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
    [reactFlowInstance, setNodes, undoRedoActions, getNodes, getEdges, isRunning]
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
      type: (node.data.nodeType || 'unknown') as string,
      name: (node.data.label || node.data.nodeType) as string,
      config: (node.data.config || {}) as Record<string, any>,
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
  // Signal that we're about to save — prevents the store update from reloading the canvas
  const markSaving = useCallback(() => { savingRef.current = true }, [])

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
    markSaving,
    canUndo: undoRedoActions.canUndo,
    canRedo: undoRedoActions.canRedo,
    canPaste,
    hasSelection,
  }), [handleCopy, handlePaste, handleCut, handleDelete, handleDuplicate, handleUndo, handleRedo, handleSelectAll, getFlowData, loadFlowData, markSaving, undoRedoActions.canUndo, undoRedoActions.canRedo, canPaste, hasSelection])

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
      type: (node.data.nodeType || 'unknown') as string,
      name: (node.data.label || node.data.nodeType) as string,
      config: (node.data.config || {}) as Record<string, any>,
      position: toArrayPosition(node.position),
    }))

    const connections = edges.map((edge) => ({
      id: edge.id,
      source: edge.source,
      target: edge.target,
      sourceOutput: edge.sourceHandle ? parseInt(edge.sourceHandle) : 0,
    }))

    await updateFlow(flowId, { nodes: flowNodes as any, connections: connections as any })
  }

  // Note: No sync from canvas → currentFlow needed.
  // handleSave/getFlowData reads directly from React Flow's nodes state.
  // handleNodeSettingsSave persists to backend immediately.

  // ========== Node events ==========

  const onNodeClick = useCallback((_event: React.MouseEvent, node: Node) => {
    if (onNodeSelect) {
      onNodeSelect(node.id, (node.data.label || node.data.nodeType) as string)
    }
  }, [onNodeSelect])

  const onNodeDoubleClick = useCallback((_event: React.MouseEvent, node: Node) => {
    // Allow opening node settings even while running (user can edit config, stop, and restart)
    setSelectedNode(node)
  }, [])

  const handleNodeSettingsClose = useCallback(() => {
    setSelectedNode(null)
  }, [])

  const handleNodeSettingsSave = useCallback((nodeId: string, config: any) => {
    console.log('[FlowCanvas] handleNodeSettingsSave called:', nodeId, JSON.stringify(config))

    // Update React Flow node state with new config
    const currentNodes = getNodes()
    const updatedNodes = currentNodes.map((node) => {
      if (node.id === nodeId) {
        return { ...node, data: { ...node.data, config } }
      }
      return node
    })
    setNodes(updatedNodes)
    setSelectedNode(null)
    undoRedoActions.push({ nodes: currentNodes, edges: getEdges() })

    // Persist to backend immediately (outside setNodes to avoid stale closure)
    if (flowId) {
      const flowNodes = updatedNodes.map((node) => {
        const d = node.data as Record<string, any>
        return {
          id: node.id,
          type: (d.nodeType || 'unknown') as string,
          name: (d.label || d.nodeType) as string,
          config: (d.config || {}) as Record<string, any>,
          position: toArrayPosition(node.position),
        }
      })

      // Debug: log exactly what we're sending to backend
      flowNodes.forEach((n) => {
        if (n.type === 'inject') {
          console.log('[FlowCanvas] Saving inject config to backend:', JSON.stringify(n.config))
        }
      })

      const connections = getEdges().map((edge) => ({
        id: edge.id,
        source: edge.source,
        target: edge.target,
        sourceOutput: edge.sourceHandle ? parseInt(edge.sourceHandle) : 0,
      }))
      savingRef.current = true
      updateFlow(flowId, { nodes: flowNodes as any, connections: connections as any })
    }
  }, [setNodes, undoRedoActions, getNodes, getEdges, flowId, updateFlow])

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
        onEdgesChange={isRunning ? undefined : onEdgesChange}
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
        nodesDraggable={!isRunning}
        nodesConnectable={!isRunning}
        elementsSelectable={!isRunning}
        fitView
        snapToGrid
        snapGrid={[15, 15]}
        className="bg-gray-50 dark:bg-gray-900"
        selectionOnDrag={!isRunning}
        selectionMode={SelectionMode.Partial}
        panOnDrag={[1, 2]}
        multiSelectionKeyCode="Control"
        selectionKeyCode={null}
        deleteKeyCode={isRunning ? null : 'Delete'}
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
          type: selectedNode.data.nodeType as string,
          name: selectedNode.data.label as string,
          config: (selectedNode.data.config || {}) as Record<string, any>
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
