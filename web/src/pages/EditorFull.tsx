import { useParams, useNavigate } from 'react-router-dom'
import { useEffect, useState, useRef } from 'react'
import { useFlowStore } from '../stores/flowStore'
import FlowCanvas, { FlowCanvasRef } from '../components/editor/FlowCanvas'
import NodePalette from '../components/editor/NodePalette'
import ExecutionDataPanel from '../components/panels/ExecutionDataPanel'
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
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  DropdownMenuSeparator,
} from '@/components/ui/dropdown-menu'
import { cn } from '@/lib/utils'

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

  useEffect(() => {
    if (id) {
      fetchFlow(id).catch((error) => {
        console.error('Failed to load flow:', error)
        // If flow not found, redirect to create new flow
        if (error.response?.status === 404) {
          console.log('Flow not found, redirecting to new flow editor')
          navigate('/editor', { replace: true })
        }
      })
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
      if (id) {
        // Update existing flow (we have an ID from URL)
        console.log('Updating existing flow:', id)
        console.log('Nodes to save:', currentFlow?.nodes)
        console.log('Connections to save:', currentFlow?.connections)
        await updateFlow(id, {
          name: flowName,
          nodes: currentFlow?.nodes || {},
          connections: currentFlow?.connections || [],
          config: currentFlow?.config || {}
        })
        console.log('Flow updated successfully')
      } else {
        // Create new flow (no ID in URL)
        console.log('Creating new flow:', flowName)
        const newFlow = await createFlow(flowName, 'Created from editor')
        if (newFlow) {
          console.log('Flow created successfully:', newFlow.id)
          // Navigate to the new flow's editor
          navigate(`/editor/${newFlow.id}`, { replace: true })
        }
      }
    } catch (error) {
      console.error('Failed to save flow:', error)
    } finally {
      setIsSaving(false)
    }
  }

  const handleDeploy = async () => {
    if (!id) {
      console.error('No flow to deploy')
      return
    }

    console.log('Deploying flow...')
    try {
      // First save the flow
      await handleSave()
      // Then deploy (start) it
      await startFlow(id)
      console.log('Flow deployed successfully')
    } catch (error) {
      console.error('Failed to deploy flow:', error)
    }
  }

  const handleRun = async () => {
    if (id) {
      await startFlow(id)
    }
  }

  const handleStop = async () => {
    if (id) {
      await stopFlow(id)
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
              <DropdownMenuItem>
                <Upload className="w-4 h-4 mr-2" />
                Deploy All Flows
              </DropdownMenuItem>
              <DropdownMenuItem>
                <Upload className="w-4 h-4 mr-2" />
                Deploy Modified Only
              </DropdownMenuItem>
              <DropdownMenuItem>
                <Upload className="w-4 h-4 mr-2" />
                Deploy This Flow
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
              <DropdownMenuItem>
                <Upload className="w-4 h-4 mr-2" />
                Import Flow
              </DropdownMenuItem>
              <DropdownMenuItem>
                <Download className="w-4 h-4 mr-2" />
                Export Flow
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem>
                <Download className="w-4 h-4 mr-2" />
                Export as JSON
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
            {/* Panel Header */}
            <div className="flex items-center justify-between px-4 py-2 border-b border-gray-200 dark:border-gray-800 bg-gray-50 dark:bg-gray-800/50">
              <Tabs value={activeBottomTab} onValueChange={setActiveBottomTab} className="flex-1">
                <TabsList className="bg-transparent p-0 h-auto">
                  <TabsTrigger value="debug" className="gap-2 data-[state=active]:bg-white dark:data-[state=active]:bg-gray-900">
                    <Bug className="w-4 h-4" />
                    Debug
                  </TabsTrigger>
                  <TabsTrigger value="monitoring" className="gap-2 data-[state=active]:bg-white dark:data-[state=active]:bg-gray-900">
                    <Activity className="w-4 h-4" />
                    Monitoring
                  </TabsTrigger>
                  <TabsTrigger value="terminal" className="gap-2 data-[state=active]:bg-white dark:data-[state=active]:bg-gray-900">
                    <Terminal className="w-4 h-4" />
                    Terminal
                  </TabsTrigger>
                  <TabsTrigger value="logs" className="gap-2 data-[state=active]:bg-white dark:data-[state=active]:bg-gray-900">
                    <FileText className="w-4 h-4" />
                    Logs
                  </TabsTrigger>
                </TabsList>
              </Tabs>

              <div className="flex items-center gap-2">
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
                  onClick={() => {
                    if (!isBottomPanelMinimized) {
                      const newHeight = Math.min(500, bottomPanelHeight + 50)
                      setBottomPanelHeight(newHeight)
                      setPreviousBottomPanelHeight(newHeight)
                    }
                  }}
                  disabled={isBottomPanelMinimized}
                  title="Increase panel height"
                >
                  <ChevronUp className="w-4 h-4" />
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

            {/* Panel Content */}
            {!isBottomPanelMinimized && (
            <Tabs value={activeBottomTab} className="flex-1 overflow-hidden">
              <TabsContent value="debug" className="h-full p-4 overflow-auto m-0">
                <div className="font-mono text-xs space-y-1">
                  <div className="text-gray-500 dark:text-gray-400">[12:34:56] Flow started</div>
                  <div className="text-blue-600 dark:text-blue-400">[12:34:57] inject-1: Injected message</div>
                  <div className="text-green-600 dark:text-green-400">[12:34:58] function-1: Processing temperature data</div>
                  <div className="text-yellow-600 dark:text-yellow-400">[12:34:59] debug-1: {"{"}"temperature": 25.4{"}"}</div>
                </div>
              </TabsContent>

              <TabsContent value="monitoring" className="h-full p-4 overflow-auto m-0">
                <div className="grid grid-cols-4 gap-4">
                  <div className="p-3 bg-blue-50 dark:bg-blue-950/30 rounded-lg border border-blue-200 dark:border-blue-900">
                    <div className="text-xs text-blue-600 dark:text-blue-400 mb-1">Messages/sec</div>
                    <div className="text-2xl font-bold text-blue-700 dark:text-blue-300">0</div>
                  </div>
                  <div className="p-3 bg-green-50 dark:bg-green-950/30 rounded-lg border border-green-200 dark:border-green-900">
                    <div className="text-xs text-green-600 dark:text-green-400 mb-1">Active Nodes</div>
                    <div className="text-2xl font-bold text-green-700 dark:text-green-300">0</div>
                  </div>
                  <div className="p-3 bg-purple-50 dark:bg-purple-950/30 rounded-lg border border-purple-200 dark:border-purple-900">
                    <div className="text-xs text-purple-600 dark:text-purple-400 mb-1">CPU Usage</div>
                    <div className="text-2xl font-bold text-purple-700 dark:text-purple-300">0%</div>
                  </div>
                  <div className="p-3 bg-orange-50 dark:bg-orange-950/30 rounded-lg border border-orange-200 dark:border-orange-900">
                    <div className="text-xs text-orange-600 dark:text-orange-400 mb-1">Memory</div>
                    <div className="text-2xl font-bold text-orange-700 dark:text-orange-300">0 MB</div>
                  </div>
                </div>
              </TabsContent>

              <TabsContent value="terminal" className="h-full p-0 overflow-auto m-0 bg-black text-green-400">
                <div className="font-mono text-sm p-4">
                  <div>EdgeFlow Terminal v1.0.0</div>
                  <div className="mt-2 flex items-center gap-2">
                    <span className="text-blue-400">$</span>
                    <span className="animate-pulse">_</span>
                  </div>
                </div>
              </TabsContent>

              <TabsContent value="logs" className="h-full p-4 overflow-auto m-0">
                <div className="font-mono text-xs space-y-1">
                  <div className="text-gray-500 dark:text-gray-400">[INFO] 2026-01-20 10:15:29 - Server started on port 8080</div>
                  <div className="text-gray-500 dark:text-gray-400">[INFO] 2026-01-20 10:15:30 - Database connection established</div>
                  <div className="text-blue-600 dark:text-blue-400">[DEBUG] 2026-01-20 10:15:31 - Loading workflow: sample-rpi5-demo</div>
                  <div className="text-green-600 dark:text-green-400">[SUCCESS] 2026-01-20 10:15:32 - Workflow loaded successfully</div>
                </div>
              </TabsContent>
            </Tabs>
            )}
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
    </div>
  )
}
