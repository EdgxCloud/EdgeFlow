import { useParams, useNavigate } from 'react-router-dom'
import { useEffect, useState, useRef, useCallback } from 'react'
import { useFlowStore } from '../stores/flowStore'
import FlowCanvas, { FlowCanvasRef } from '../components/editor/FlowCanvas'
import NodePalette from '../components/editor/NodePalette'
import ExecutionDataPanel from '../components/panels/ExecutionDataPanel'
import DebugPanel from '../components/panels/DebugPanel'
import { TerminalPanel } from '../components/panels/TerminalPanel'
import { MonitoringPanel } from '../components/panels/MonitoringPanel'
import { GPIOPanel } from '../components/panels/GPIOPanel'
import { LogsPanel, pushLog } from '../components/panels/LogsPanel'
import {
  Play,
  Square,
  Save,
  Upload,
  Download,
  Settings,
  Bug,
  Maximize2,
  Minimize2,
  ChevronLeft,
  ChevronRight,
  Zap,
  AlertCircle,
  CheckCircle2,
  Terminal,
  Activity,
  FileText,
  ChevronUp,
  ChevronDown,
  X,
  Database,
  Undo,
  Redo,
  Copy,
  Scissors,
  ClipboardPaste,
  CopyPlus,
  Trash2,
  GripHorizontal,
  FileJson,
  ClipboardCopy,
  CircuitBoard,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  DropdownMenuSeparator,
} from '@/components/ui/dropdown-menu'
import { cn } from '@/lib/utils'
import { ReactFlowProvider } from '@xyflow/react'
import { toast } from 'sonner'
import { downloadFlow, importFlow, copyFlowToClipboard } from '@/utils/flowImportExport'
import { Flow as ExportFlow } from '@/types/flow'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Textarea } from '@/components/ui/textarea'

export default function EditorFull() {
  const { id } = useParams()
  const navigate = useNavigate()
  const canvasRef = useRef<FlowCanvasRef>(null)
  const { currentFlow, fetchFlow, startFlow, stopFlow, updateFlow, createFlow } = useFlowStore()
  const [isRunning, setIsRunning] = useState(false)
  const [isPaletteOpen, setIsPaletteOpen] = useState(true)
  const [isFullscreen, setIsFullscreen] = useState(false)
  const [flowName, setFlowName] = useState('Untitled Workflow')
  const [isBottomPanelOpen, setIsBottomPanelOpen] = useState(false)
  const [bottomPanelHeight, setBottomPanelHeight] = useState(250)
  const [isBottomPanelMinimized, setIsBottomPanelMinimized] = useState(false)
  const [previousBottomPanelHeight, setPreviousBottomPanelHeight] = useState(250)
  const [activeBottomTab, setActiveBottomTab] = useState('debug')
  const [isSaving, setIsSaving] = useState(false)
  const [selectedNodeId, setSelectedNodeId] = useState<string | undefined>()
  const [selectedNodeName, setSelectedNodeName] = useState<string | undefined>()
  const [isDataPanelOpen, setIsDataPanelOpen] = useState(false)
  const [isPasteDialogOpen, setIsPasteDialogOpen] = useState(false)
  const [pasteJSON, setPasteJSON] = useState('')

  useEffect(() => {
    if (id) {
      fetchFlow(id).then(() => {
        pushLog('info', `Workflow loaded: ${id}`, 'editor')
      }).catch((error) => {
        console.error('Failed to load flow:', error)
        pushLog('error', `Failed to load workflow: ${error}`, 'editor')
        // If flow not found, redirect to create new flow
        if (error.response?.status === 404) {
          console.log('Flow not found, redirecting to new flow editor')
          navigate('/editor', { replace: true })
        }
      })
    } else {
      pushLog('info', 'New workflow editor opened', 'editor')
    }
  }, [id, fetchFlow, navigate])

  useEffect(() => {
    if (currentFlow) {
      setIsRunning(currentFlow.status === 'running')
      setFlowName(currentFlow.name || 'Untitled Workflow')
    }
  }, [currentFlow])

  const handleSave = async () => {
    setIsSaving(true)
    try {
      // Get current canvas data directly from FlowCanvas ref
      const flowData = canvasRef.current?.getFlowData()
      const nodesToSave = flowData?.nodes || []
      const connectionsToSave = flowData?.connections || []

      let flowId = id

      // If no flow ID yet, create a new flow first
      if (!flowId) {
        const newFlow = await createFlow(flowName, 'Created from editor')
        if (!newFlow) {
          toast.error('Failed to create workflow')
          return
        }
        flowId = newFlow.id
      }

      // Save nodes and connections to the flow
      await updateFlow(flowId, {
        name: flowName,
        nodes: nodesToSave,
        connections: connectionsToSave,
        config: currentFlow?.config || {}
      })

      toast.success(`Workflow saved (${nodesToSave.length} nodes, ${connectionsToSave.length} connections)`)
      pushLog('success', `Workflow saved: ${flowName} (${nodesToSave.length} nodes, ${connectionsToSave.length} connections)`, 'editor')

      // Navigate to the flow URL if it was newly created
      if (!id && flowId) {
        navigate(`/editor/${flowId}`, { replace: true })
        pushLog('info', `New workflow created: ${flowName}`, 'editor')
      }
    } catch (error) {
      console.error('Failed to save flow:', error)
      toast.error('Failed to save workflow')
      pushLog('error', `Failed to save workflow: ${error}`, 'editor')
    } finally {
      setIsSaving(false)
    }
  }

  const handleDeploy = async () => {
    pushLog('info', `Deploying workflow: ${flowName}...`, 'editor')
    try {
      // Save first (this also creates the flow if needed)
      await handleSave()
      const flowId = id || useFlowStore.getState().currentFlow?.id
      if (!flowId) {
        toast.error('No flow to deploy')
        return
      }
      await startFlow(flowId)
      setIsRunning(true)
      toast.success('Flow deployed and running')
      pushLog('success', `Workflow deployed and running: ${flowName}`, 'deploy')
    } catch (error) {
      console.error('Failed to deploy flow:', error)
      toast.error('Failed to deploy flow')
      pushLog('error', `Failed to deploy workflow: ${error}`, 'deploy')
    }
  }

  const handleRun = async () => {
    const flowId = id || currentFlow?.id
    if (flowId) {
      try {
        await startFlow(flowId)
        setIsRunning(true)
        toast.success('Flow started')
        pushLog('success', `Flow started: ${flowName}`, 'runtime')
      } catch (error) {
        toast.error('Failed to start flow')
        pushLog('error', `Failed to start flow: ${error}`, 'runtime')
      }
    } else {
      toast.error('Save the flow first before running')
      pushLog('warn', 'Cannot run: flow not saved yet', 'runtime')
    }
  }

  const handleStop = async () => {
    const flowId = id || currentFlow?.id
    if (flowId) {
      try {
        await stopFlow(flowId)
        setIsRunning(false)
        toast.success('Flow stopped')
        pushLog('info', `Flow stopped: ${flowName}`, 'runtime')
      } catch (error) {
        toast.error('Failed to stop flow')
        pushLog('error', `Failed to stop flow: ${error}`, 'runtime')
      }
    }
  }

  const togglePalette = () => {
    setIsPaletteOpen(!isPaletteOpen)
  }

  const toggleFullscreen = () => {
    if (!document.fullscreenElement) {
      document.documentElement.requestFullscreen().catch((err) => {
        console.error('Error attempting to enable fullscreen:', err)
      })
    } else {
      if (document.exitFullscreen) {
        document.exitFullscreen().catch((err) => {
          console.error('Error attempting to exit fullscreen:', err)
        })
      }
    }
  }

  const buildCurrentFlow = useCallback((): ExportFlow => {
    const flowData = canvasRef.current?.getFlowData()
    return {
      id: id || currentFlow?.id || 'unsaved',
      name: flowName,
      description: currentFlow?.description || '',
      nodes: flowData?.nodes || [],
      connections: flowData?.connections || [],
      status: isRunning ? 'running' : 'stopped',
    }
  }, [id, currentFlow, flowName, isRunning])

  const handleCopyWorkflowJSON = useCallback(async () => {
    try {
      const flow = buildCurrentFlow()
      await copyFlowToClipboard(flow)
      toast.success('Workflow JSON copied to clipboard')
      pushLog('info', `Workflow JSON copied to clipboard: ${flowName}`, 'editor')
    } catch (error) {
      toast.error('Failed to copy workflow JSON')
      pushLog('error', `Failed to copy workflow JSON: ${error}`, 'editor')
    }
  }, [buildCurrentFlow, flowName])

  const handleDownloadWorkflowJSON = useCallback(() => {
    try {
      const flow = buildCurrentFlow()
      downloadFlow(flow)
      toast.success('Workflow downloaded as JSON')
      pushLog('info', `Workflow downloaded: ${flowName}.json`, 'editor')
    } catch (error) {
      toast.error('Failed to download workflow')
      pushLog('error', `Failed to download workflow: ${error}`, 'editor')
    }
  }, [buildCurrentFlow, flowName])

  const handleImportWorkflowJSON = useCallback(() => {
    const input = document.createElement('input')
    input.type = 'file'
    input.accept = '.json'
    input.onchange = (e) => {
      const file = (e.target as HTMLInputElement).files?.[0]
      if (!file) return

      const reader = new FileReader()
      reader.onload = async (event) => {
        try {
          const json = event.target?.result as string
          const flows = importFlow(json)
          if (flows.length === 0) {
            toast.error('No valid flows found in file')
            return
          }

          const imported = flows[0]
          // Create a new flow from imported data
          const newFlow = await createFlow(imported.name, imported.description || '')
          if (!newFlow) {
            toast.error('Failed to create imported flow')
            return
          }

          // Update with imported nodes and connections
          await updateFlow(newFlow.id, {
            name: imported.name,
            nodes: imported.nodes,
            connections: imported.connections || [],
          })

          toast.success(`Flow "${imported.name}" imported successfully`)
          pushLog('success', `Flow imported: ${imported.name} (${imported.nodes.length} nodes)`, 'editor')
          navigate(`/editor/${newFlow.id}`)
        } catch (error: any) {
          toast.error(`Import failed: ${error.message}`)
          pushLog('error', `Flow import failed: ${error.message}`, 'editor')
        }
      }
      reader.readAsText(file)
    }
    input.click()
  }, [createFlow, updateFlow, navigate])

  const handlePasteWorkflowJSON = useCallback(() => {
    setIsPasteDialogOpen(true)
    setPasteJSON('')
    // Auto-read from clipboard if available
    navigator.clipboard.readText().then((text) => {
      if (text.trim().startsWith('{')) {
        setPasteJSON(text)
      }
    }).catch(() => {
      // Clipboard read may fail due to permissions - user can paste manually
    })
  }, [])

  const handlePasteDialogConfirm = useCallback(() => {
    try {
      const json = pasteJSON.trim()
      if (!json) {
        toast.error('Please paste workflow JSON')
        return
      }

      const flows = importFlow(json)
      if (flows.length === 0) {
        toast.error('No valid flows found in JSON')
        return
      }

      const imported = flows[0]
      const nodeCount = imported.nodes?.length || 0
      const connectionCount = imported.connections?.length || 0

      // Load directly onto current canvas
      canvasRef.current?.loadFlowData(imported.nodes || [], imported.connections || [])

      // Update flow name if it came from the pasted data
      if (imported.name && imported.name !== flowName) {
        setFlowName(imported.name)
      }

      setIsPasteDialogOpen(false)
      setPasteJSON('')
      toast.success(`Workflow loaded: ${nodeCount} nodes, ${connectionCount} connections`)
      pushLog('success', `Workflow pasted from JSON: ${imported.name} (${nodeCount} nodes)`, 'editor')
    } catch (error: any) {
      toast.error(`Invalid JSON: ${error.message}`)
      pushLog('error', `Failed to paste workflow JSON: ${error.message}`, 'editor')
    }
  }, [pasteJSON, flowName])

  const toggleMinimizeBottomPanel = () => {
    if (isBottomPanelMinimized) {
      // Restore to previous height
      setBottomPanelHeight(previousBottomPanelHeight)
      setIsBottomPanelMinimized(false)
    } else {
      // Minimize to just show header (40px)
      setPreviousBottomPanelHeight(bottomPanelHeight)
      setBottomPanelHeight(40)
      setIsBottomPanelMinimized(true)
    }
  }

  // Drag-to-resize for bottom panel
  const isDraggingRef = useRef(false)
  const dragStartYRef = useRef(0)
  const dragStartHeightRef = useRef(0)

  const handleDragStart = useCallback((e: React.MouseEvent) => {
    e.preventDefault()
    isDraggingRef.current = true
    dragStartYRef.current = e.clientY
    dragStartHeightRef.current = bottomPanelHeight
    document.body.style.cursor = 'row-resize'
    document.body.style.userSelect = 'none'

    const handleMouseMove = (e: MouseEvent) => {
      if (!isDraggingRef.current) return
      const delta = dragStartYRef.current - e.clientY
      const newHeight = Math.max(100, Math.min(window.innerHeight * 0.7, dragStartHeightRef.current + delta))
      setBottomPanelHeight(newHeight)
      setPreviousBottomPanelHeight(newHeight)
      if (isBottomPanelMinimized && newHeight > 60) {
        setIsBottomPanelMinimized(false)
      }
    }

    const handleMouseUp = () => {
      isDraggingRef.current = false
      document.body.style.cursor = ''
      document.body.style.userSelect = ''
      document.removeEventListener('mousemove', handleMouseMove)
      document.removeEventListener('mouseup', handleMouseUp)
    }

    document.addEventListener('mousemove', handleMouseMove)
    document.addEventListener('mouseup', handleMouseUp)
  }, [bottomPanelHeight, isBottomPanelMinimized])

  // Listen for fullscreen changes
  useEffect(() => {
    const handleFullscreenChange = () => {
      setIsFullscreen(!!document.fullscreenElement)
    }

    document.addEventListener('fullscreenchange', handleFullscreenChange)

    return () => {
      document.removeEventListener('fullscreenchange', handleFullscreenChange)
    }
  }, [])

  return (
    <div className="h-screen flex flex-col bg-gradient-to-br from-gray-50 to-gray-100 dark:from-gray-950 dark:to-gray-900">
      {/* Modern Top Bar */}
      <div className="h-16 bg-white dark:bg-gray-900 border-b border-gray-200 dark:border-gray-800 shadow-sm flex items-center justify-between px-6">
        {/* Left Section - Flow Info */}
        <div className="flex items-center gap-4">
          <button
            onClick={() => navigate('/workflows')}
            className="p-2 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg transition-colors"
            title="Back to Workflows"
          >
            <ChevronLeft className="w-5 h-5 text-gray-600 dark:text-gray-400" />
          </button>

          <div className="flex items-center gap-3">
            <div className="w-10 h-10 bg-gradient-to-br from-blue-500 to-blue-600 rounded-xl flex items-center justify-center shadow-lg shadow-blue-500/30">
              <Zap className="w-5 h-5 text-white" />
            </div>
            <div>
              <Input
                value={flowName}
                onChange={(e) => setFlowName(e.target.value)}
                className="text-lg font-semibold border-0 bg-transparent px-2 h-8 focus-visible:ring-0 focus-visible:bg-gray-50 dark:focus-visible:bg-gray-800 rounded"
              />
              <div className="flex items-center gap-2 px-2">
                <Badge
                  variant={isRunning ? 'default' : 'secondary'}
                  className={cn(
                    'text-xs px-2 py-0',
                    isRunning
                      ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
                      : 'bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-400'
                  )}
                >
                  {isRunning ? (
                    <>
                      <CheckCircle2 className="w-3 h-3 mr-1" />
                      Running
                    </>
                  ) : (
                    <>
                      <AlertCircle className="w-3 h-3 mr-1" />
                      Stopped
                    </>
                  )}
                </Badge>
                <span className="text-xs text-muted-foreground">
                  Last saved 2 minutes ago
                </span>
              </div>
            </div>
          </div>
        </div>

        {/* Center Section - Quick Actions */}
        <div className="flex items-center gap-2">
          {/* Edit Toolbar */}
          <div className="flex items-center gap-0.5 bg-gray-100 dark:bg-gray-800 rounded-lg p-0.5">
            <Button
              variant="ghost"
              size="icon"
              className="h-7 w-7"
              onClick={() => canvasRef.current?.handleUndo()}
              title="Undo (Ctrl+Z)"
            >
              <Undo className="w-3.5 h-3.5" />
            </Button>
            <Button
              variant="ghost"
              size="icon"
              className="h-7 w-7"
              onClick={() => canvasRef.current?.handleRedo()}
              title="Redo (Ctrl+Y)"
            >
              <Redo className="w-3.5 h-3.5" />
            </Button>
          </div>

          <div className="w-px h-6 bg-gray-300 dark:bg-gray-700" />

          <div className="flex items-center gap-0.5 bg-gray-100 dark:bg-gray-800 rounded-lg p-0.5">
            <Button
              variant="ghost"
              size="icon"
              className="h-7 w-7"
              onClick={() => canvasRef.current?.handleCopy()}
              title="Copy (Ctrl+C)"
            >
              <Copy className="w-3.5 h-3.5" />
            </Button>
            <Button
              variant="ghost"
              size="icon"
              className="h-7 w-7"
              onClick={() => canvasRef.current?.handleCut()}
              title="Cut (Ctrl+X)"
            >
              <Scissors className="w-3.5 h-3.5" />
            </Button>
            <Button
              variant="ghost"
              size="icon"
              className="h-7 w-7"
              onClick={() => canvasRef.current?.handlePaste()}
              title="Paste (Ctrl+V)"
            >
              <ClipboardPaste className="w-3.5 h-3.5" />
            </Button>
            <Button
              variant="ghost"
              size="icon"
              className="h-7 w-7"
              onClick={() => canvasRef.current?.handleDuplicate()}
              title="Duplicate (Ctrl+D)"
            >
              <CopyPlus className="w-3.5 h-3.5" />
            </Button>
            <Button
              variant="ghost"
              size="icon"
              className="h-7 w-7"
              onClick={() => canvasRef.current?.handleDelete()}
              title="Delete (Del)"
            >
              <Trash2 className="w-3.5 h-3.5" />
            </Button>
          </div>

          <div className="w-px h-6 bg-gray-300 dark:bg-gray-700" />

          {isRunning ? (
            <Button
              onClick={handleStop}
              variant="outline"
              size="sm"
              className="gap-2 border-red-200 text-red-600 hover:bg-red-50 dark:border-red-900 dark:text-red-400 dark:hover:bg-red-950"
            >
              <Square className="w-4 h-4" />
              Stop
            </Button>
          ) : (
            <Button
              onClick={handleRun}
              variant="outline"
              size="sm"
              className="gap-2 border-green-200 text-green-600 hover:bg-green-50 dark:border-green-900 dark:text-green-400 dark:hover:bg-green-950"
            >
              <Play className="w-4 h-4" />
              Run
            </Button>
          )}

          <Button
            onClick={handleSave}
            variant="outline"
            size="sm"
            className="gap-2"
            disabled={isSaving}
          >
            <Save className="w-4 h-4" />
            {isSaving ? 'Saving...' : 'Save'}
          </Button>

          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="default" size="sm" className="gap-2">
                <Upload className="w-4 h-4" />
                Deploy
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-56">
              <DropdownMenuItem onClick={handleDeploy}>
                <Upload className="w-4 h-4 mr-2" />
                Deploy This Flow
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem onClick={handleRun} disabled={isRunning}>
                <Play className="w-4 h-4 mr-2" />
                Run Flow
              </DropdownMenuItem>
              <DropdownMenuItem onClick={handleStop} disabled={!isRunning}>
                <Square className="w-4 h-4 mr-2" />
                Stop Flow
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>

        {/* Right Section - Tools */}
        <div className="flex items-center gap-2">
          <Button
            onClick={togglePalette}
            variant="ghost"
            size="sm"
            className="gap-2"
          >
            {isPaletteOpen ? (
              <ChevronLeft className="w-4 h-4" />
            ) : (
              <ChevronRight className="w-4 h-4" />
            )}
            {isPaletteOpen ? 'Hide' : 'Show'} Palette
          </Button>

          <Button
            onClick={() => setIsDataPanelOpen(!isDataPanelOpen)}
            variant={isDataPanelOpen ? "default" : "ghost"}
            size="sm"
            className="gap-2"
            title="Toggle Execution Data Panel"
          >
            <Database className="w-4 h-4" />
            Data
          </Button>

          <div className="w-px h-6 bg-gray-300 dark:bg-gray-700" />

          <Button variant="ghost" size="icon">
            <Bug className="w-4 h-4" />
          </Button>

          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="icon">
                <Download className="w-4 h-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem onClick={handleImportWorkflowJSON}>
                <Upload className="w-4 h-4 mr-2" />
                Import Flow
              </DropdownMenuItem>
              <DropdownMenuItem onClick={handleDownloadWorkflowJSON}>
                <Download className="w-4 h-4 mr-2" />
                Export Flow
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem onClick={handleCopyWorkflowJSON}>
                <ClipboardCopy className="w-4 h-4 mr-2" />
                Copy as JSON
              </DropdownMenuItem>
              <DropdownMenuItem onClick={handlePasteWorkflowJSON}>
                <ClipboardPaste className="w-4 h-4 mr-2" />
                Paste from JSON
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>

          <Button variant="ghost" size="icon">
            <Settings className="w-4 h-4" />
          </Button>

          <Button
            onClick={toggleFullscreen}
            variant="ghost"
            size="icon"
          >
            {isFullscreen ? (
              <Minimize2 className="w-4 h-4" />
            ) : (
              <Maximize2 className="w-4 h-4" />
            )}
          </Button>
        </div>
      </div>

      {/* Main Editor Area */}
      <div className="flex-1 flex flex-col overflow-hidden relative">
        {/* Top Section: Palette + Canvas */}
        <div className="flex-1 flex overflow-hidden relative" style={{ height: isBottomPanelOpen ? `calc(100% - ${bottomPanelHeight}px)` : '100%' }}>
          {/* Node Palette - Collapsible */}
          <div
            className={cn(
              'transition-all duration-300 ease-in-out border-r border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900',
              isPaletteOpen ? 'w-80' : 'w-0 overflow-hidden'
            )}
          >
            <NodePalette />
          </div>

          {/* Canvas Area */}
          <div className="flex-1 relative">
          {/* Canvas Background Pattern */}
          <div className="absolute inset-0 opacity-5 dark:opacity-10">
            <div
              className="h-full w-full"
              style={{
                backgroundImage: `radial-gradient(circle, currentColor 1px, transparent 1px)`,
                backgroundSize: '24px 24px',
              }}
            />
          </div>

          {/* Flow Canvas */}
          <div className="relative h-full">
            <ReactFlowProvider>
              <FlowCanvas
                ref={canvasRef}
                flowId={id}
                onNodeSelect={(nodeId, nodeName) => {
                  setSelectedNodeId(nodeId || undefined)
                  setSelectedNodeName(nodeName)
                  if (nodeId) {
                    setIsDataPanelOpen(true)
                  }
                }}
              />
            </ReactFlowProvider>
          </div>

          {/* Right Panel - Execution Data */}
          {isDataPanelOpen && (
            <div className="absolute top-0 right-0 bottom-0 w-96 shadow-2xl z-10">
              <ExecutionDataPanel
                nodeId={selectedNodeId}
                nodeName={selectedNodeName}
                onClose={() => setIsDataPanelOpen(false)}
                className="h-full"
              />
            </div>
          )}

          {/* Floating Stats Bar */}
          <div className="absolute bottom-6 left-6 right-6 flex items-center justify-between">
            <div className="bg-white/90 dark:bg-gray-900/90 backdrop-blur-lg border border-gray-200 dark:border-gray-800 rounded-xl shadow-xl px-4 py-2 flex items-center gap-6">
              <div className="flex items-center gap-2 text-sm">
                <div className="w-2 h-2 bg-blue-500 rounded-full" />
                <span className="text-gray-700 dark:text-gray-300">
                  0 Nodes
                </span>
              </div>
              <div className="w-px h-4 bg-gray-300 dark:bg-gray-700" />
              <div className="flex items-center gap-2 text-sm">
                <div className="w-2 h-2 bg-green-500 rounded-full" />
                <span className="text-gray-700 dark:text-gray-300">
                  0 Connections
                </span>
              </div>
              <div className="w-px h-4 bg-gray-300 dark:bg-gray-700" />
              <div className="flex items-center gap-2 text-sm">
                <Zap className="w-4 h-4 text-yellow-500" />
                <span className="text-gray-700 dark:text-gray-300">
                  Ready
                </span>
              </div>
            </div>

            {/* Quick Tip */}
            <div className="bg-blue-50/90 dark:bg-blue-950/90 backdrop-blur-lg border border-blue-200 dark:border-blue-900 rounded-xl shadow-xl px-4 py-2">
              <p className="text-xs text-blue-700 dark:text-blue-400">
                💡 Drag nodes from the palette to get started
              </p>
            </div>
          </div>
          </div>
        </div>

        {/* Bottom Panel - Debug, Terminal, Logs */}
        {isBottomPanelOpen && (
          <div
            className="border-t border-gray-200 dark:border-gray-800 bg-white dark:bg-gray-900 flex flex-col"
            style={{ height: `${bottomPanelHeight}px` }}
          >
            {/* Drag Handle for Resizing */}
            <div
              className="h-1.5 cursor-row-resize bg-transparent hover:bg-blue-400/40 transition-colors flex items-center justify-center group shrink-0"
              onMouseDown={handleDragStart}
            >
              <GripHorizontal className="w-4 h-3 text-gray-400 opacity-0 group-hover:opacity-100 transition-opacity" />
            </div>

            {/* Panel Header */}
            <div className="flex items-center justify-between px-4 py-1.5 border-b border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-800/50 shrink-0">
              <div className="flex items-center gap-1">
                {[
                  { value: 'debug', icon: Bug, label: 'Debug' },
                  { value: 'monitoring', icon: Activity, label: 'Monitoring' },
                  { value: 'gpio', icon: CircuitBoard, label: 'GPIO' },
                  { value: 'terminal', icon: Terminal, label: 'Terminal' },
                  { value: 'logs', icon: FileText, label: 'Logs' },
                ].map((tab) => (
                  <button
                    key={tab.value}
                    onClick={() => setActiveBottomTab(tab.value)}
                    className={cn(
                      'inline-flex items-center gap-1.5 px-3 py-1 text-sm font-medium rounded-md transition-colors',
                      activeBottomTab === tab.value
                        ? 'bg-white dark:bg-gray-900 text-foreground shadow-sm'
                        : 'text-muted-foreground hover:text-foreground hover:bg-gray-100 dark:hover:bg-gray-800'
                    )}
                  >
                    <tab.icon className="w-3.5 h-3.5" />
                    {tab.label}
                  </button>
                ))}
              </div>

              <div className="flex items-center gap-1">
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-6 w-6"
                  onClick={toggleMinimizeBottomPanel}
                  title={isBottomPanelMinimized ? "Expand panel" : "Minimize panel"}
                >
                  {isBottomPanelMinimized ? (
                    <ChevronUp className="w-4 h-4" />
                  ) : (
                    <ChevronDown className="w-4 h-4" />
                  )}
                </Button>
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-6 w-6"
                  onClick={() => setIsBottomPanelOpen(false)}
                  title="Close panel"
                >
                  <X className="w-4 h-4" />
                </Button>
              </div>
            </div>

            {/* Panel Content - all panels always rendered, CSS hides inactive/minimized */}
              <div className={cn('flex-1 overflow-hidden relative', isBottomPanelMinimized && 'hidden')}>
                <div className={cn('absolute inset-0', activeBottomTab === 'debug' ? 'block' : 'hidden')}>
                  <DebugPanel className="h-full" />
                </div>
                <div className={cn('absolute inset-0', activeBottomTab === 'monitoring' ? 'block' : 'hidden')}>
                  <MonitoringPanel />
                </div>
                <div className={cn('absolute inset-0', activeBottomTab === 'gpio' ? 'block' : 'hidden')}>
                  <GPIOPanel />
                </div>
                <div className={cn('absolute inset-0', activeBottomTab === 'terminal' ? 'block' : 'hidden')}>
                  <TerminalPanel />
                </div>
                <div className={cn('absolute inset-0', activeBottomTab === 'logs' ? 'block' : 'hidden')}>
                  <LogsPanel />
                </div>
              </div>
          </div>
        )}

        {/* Toggle Bottom Panel Button (when closed) */}
        {!isBottomPanelOpen && (
          <div className="border-t border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-800/50 flex items-center justify-center py-1">
            <Button
              variant="ghost"
              size="sm"
              className="gap-2 h-7 text-xs"
              onClick={() => setIsBottomPanelOpen(true)}
            >
              <ChevronUp className="w-3 h-3" />
              Show Debug Panel
            </Button>
          </div>
        )}
      </div>

      {/* Paste from JSON Dialog */}
      <Dialog open={isPasteDialogOpen} onOpenChange={setIsPasteDialogOpen}>
        <DialogContent className="sm:max-w-2xl">
          <DialogHeader>
            <DialogTitle>Paste Workflow JSON</DialogTitle>
            <DialogDescription>
              Paste a workflow JSON exported from EdgeFlow or n8n to load it onto the canvas.
            </DialogDescription>
          </DialogHeader>
          <Textarea
            value={pasteJSON}
            onChange={(e) => setPasteJSON(e.target.value)}
            placeholder='{"version": "1.0", "flows": [...]}'
            className="min-h-[300px] font-mono text-sm"
            autoFocus
          />
          <DialogFooter>
            <Button variant="outline" onClick={() => setIsPasteDialogOpen(false)}>
              Cancel
            </Button>
            <Button onClick={handlePasteDialogConfirm} disabled={!pasteJSON.trim()}>
              <ClipboardPaste className="w-4 h-4 mr-2" />
              Load Workflow
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
