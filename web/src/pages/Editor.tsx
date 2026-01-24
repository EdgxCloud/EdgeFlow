import { useParams } from 'react-router-dom'
import { useEffect, useState } from 'react'
import { ReactFlowProvider } from '@xyflow/react'
import { useFlowStore } from '../stores/flowStore'
import FlowCanvas from '../components/canvas/FlowCanvas'
import NodePalette from '../components/panels/NodePalette'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  Play,
  Square,
  Upload,
  Download,
  Settings,
  Bug,
  ChevronDown,
  ZoomIn,
  ZoomOut,
  Maximize2,
  ChevronUp,
  Terminal as TerminalIcon,
  Activity,
  X,
} from 'lucide-react'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  DropdownMenuSeparator,
} from '@/components/ui/dropdown-menu'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { ScrollArea } from '@/components/ui/scroll-area'
import DeployDialog from '../components/dialogs/DeployDialog'
import { toast } from 'sonner'
import { cn } from '@/lib/utils'

console.log('🚀 Editor.tsx FILE LOADED')

export default function Editor() {
  const { id } = useParams()
  const { currentFlow, fetchFlow, startFlow, stopFlow } = useFlowStore()
  const [isDeployDialogOpen, setIsDeployDialogOpen] = useState(false)
  const [zoom, setZoom] = useState(100)
  const [isDebugPanelOpen, setIsDebugPanelOpen] = useState(false)
  const [debugPanelHeight, setDebugPanelHeight] = useState(300)
  const [debugLogs, setDebugLogs] = useState<Array<{ time: string; message: string; type: string }>>([])
  const [terminalInput, setTerminalInput] = useState('')
  const [terminalOutput, setTerminalOutput] = useState<string[]>([
    'EdgeFlow Terminal v1.0',
    'Type "help" for available commands',
    '',
  ])

  console.log('🚀 Editor component rendering, id:', id)

  useEffect(() => {
    console.log('🚀 Editor useEffect, id:', id)
    if (id) {
      fetchFlow(id)
    }
  }, [id, fetchFlow])

  const handleRun = async () => {
    if (!id) {
      toast.error('No flow ID available')
      return
    }
    try {
      await startFlow(id)
      toast.success('Flow started successfully')
      addDebugLog('Flow started', 'success')
    } catch (error) {
      toast.error('Failed to start flow', {
        description: error instanceof Error ? error.message : 'Unknown error'
      })
      addDebugLog('Failed to start flow: ' + (error instanceof Error ? error.message : 'Unknown error'), 'error')
    }
  }

  const handleStop = async () => {
    if (!id) {
      toast.error('No flow ID available')
      return
    }
    try {
      await stopFlow(id)
      toast.success('Flow stopped successfully')
      addDebugLog('Flow stopped', 'info')
    } catch (error) {
      toast.error('Failed to stop flow', {
        description: error instanceof Error ? error.message : 'Unknown error'
      })
      addDebugLog('Failed to stop flow: ' + (error instanceof Error ? error.message : 'Unknown error'), 'error')
    }
  }

  const addDebugLog = (message: string, type: 'info' | 'success' | 'error' | 'warning' = 'info') => {
    const time = new Date().toLocaleTimeString()
    setDebugLogs(prev => [...prev, { time, message, type }])
  }

  const handleDeploy = (mode: 'full' | 'modified' | 'flow') => {
    setIsDeployDialogOpen(true)
  }

  const handleImport = () => {
    const input = document.createElement('input')
    input.type = 'file'
    input.accept = '.json'
    input.onchange = async (e) => {
      const file = (e.target as HTMLInputElement).files?.[0]
      if (file) {
        const reader = new FileReader()
        reader.onload = (event) => {
          try {
            const flowData = JSON.parse(event.target?.result as string)
            toast.success('Flow imported successfully')
            addDebugLog('Flow imported: ' + file.name, 'success')
            console.log('Imported flow:', flowData)
          } catch (error) {
            toast.error('Failed to import flow', {
              description: 'Invalid JSON format'
            })
            addDebugLog('Failed to import flow: Invalid JSON', 'error')
          }
        }
        reader.readAsText(file)
      }
    }
    input.click()
  }

  const handleExport = () => {
    if (!currentFlow) {
      toast.error('No flow to export')
      return
    }
    const dataStr = JSON.stringify(currentFlow, null, 2)
    const dataBlob = new Blob([dataStr], { type: 'application/json' })
    const url = URL.createObjectURL(dataBlob)
    const link = document.createElement('a')
    link.href = url
    link.download = `${currentFlow.name || 'flow'}.json`
    link.click()
    URL.revokeObjectURL(url)
    toast.success('Flow exported successfully')
    addDebugLog('Flow exported: ' + (currentFlow.name || 'flow') + '.json', 'success')
  }

  const handleZoomIn = () => {
    setZoom(prev => Math.min(prev + 10, 200))
  }

  const handleZoomOut = () => {
    setZoom(prev => Math.max(prev - 10, 50))
  }

  const handleFitView = () => {
    setZoom(100)
  }

  const toggleDebugPanel = () => {
    setIsDebugPanelOpen(!isDebugPanelOpen)
  }

  const handleTerminalCommand = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && terminalInput.trim()) {
      const command = terminalInput.trim()
      setTerminalOutput(prev => [...prev, `$ ${command}`, executeTerminalCommand(command)])
      setTerminalInput('')
    }
  }

  const executeTerminalCommand = (command: string): string => {
    const parts = command.split(' ')
    const cmd = parts[0].toLowerCase()

    switch (cmd) {
      case 'help':
        return 'Available commands:\n  help - Show this help\n  clear - Clear terminal\n  flow - Show current flow info\n  nodes - List flow nodes\n  status - Show flow status'
      case 'clear':
        setTimeout(() => setTerminalOutput([]), 0)
        return ''
      case 'flow':
        return currentFlow ? `Flow: ${currentFlow.name}\nID: ${currentFlow.id}\nStatus: ${currentFlow.status}` : 'No flow loaded'
      case 'nodes':
        return currentFlow ? `Nodes: ${Object.keys(currentFlow.nodes || {}).length}` : 'No flow loaded'
      case 'status':
        return currentFlow?.status || 'unknown'
      default:
        return `Command not found: ${cmd}. Type 'help' for available commands.`
    }
  }

  const isRunning = currentFlow?.status === 'running'

  return (
    <div className="flex flex-col h-[calc(100vh-4rem)] -m-6">
      {/* Editor Toolbar */}
      <div className="flex items-center justify-between px-4 py-2 border-b border-border bg-card">
        <div className="flex items-center gap-3">
          {/* Flow Name */}
          <Input
            placeholder="Untitled Workflow"
            defaultValue={currentFlow?.name || ''}
            className="w-64 h-9"
          />
          <Badge variant="secondary">
            {currentFlow?.status || 'Stopped'}
          </Badge>
        </div>

        <div className="flex items-center gap-2">
          {/* Deploy Menu */}
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="default" size="sm">
                <Upload className="h-4 w-4 mr-2" />
                Deploy
                <ChevronDown className="h-3 w-3 ml-1" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem onClick={() => handleDeploy('full')}>
                <Upload className="h-4 w-4 mr-2" />
                Deploy All
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => handleDeploy('modified')}>
                <Upload className="h-4 w-4 mr-2" />
                Deploy Modified Flows
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => handleDeploy('flow')}>
                <Upload className="h-4 w-4 mr-2" />
                Deploy This Flow
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>

          {/* Run/Stop */}
          {isRunning ? (
            <Button variant="outline" size="sm" onClick={handleStop}>
              <Square className="h-4 w-4 mr-2" />
              Stop
            </Button>
          ) : (
            <Button variant="outline" size="sm" onClick={handleRun}>
              <Play className="h-4 w-4 mr-2" />
              Run
            </Button>
          )}

          {/* Debug Toggle */}
          <Button
            variant={isDebugPanelOpen ? "default" : "ghost"}
            size="sm"
            title="Toggle Debug Panel"
            onClick={toggleDebugPanel}
          >
            <Bug className="h-4 w-4" />
          </Button>

          {/* Import/Export */}
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="sm">
                <Download className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuItem onClick={handleImport}>
                <Upload className="h-4 w-4 mr-2" />
                Import
              </DropdownMenuItem>
              <DropdownMenuItem onClick={handleExport}>
                <Download className="h-4 w-4 mr-2" />
                Export
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem onClick={handleExport}>
                <Download className="h-4 w-4 mr-2" />
                Export as JSON
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>

          {/* Settings */}
          <Button variant="ghost" size="sm" title="Flow Settings">
            <Settings className="h-4 w-4" />
          </Button>
        </div>
      </div>

      {/* Main Editor Area */}
      <div
        className="flex-1 flex overflow-hidden"
        style={{ height: isDebugPanelOpen ? `calc(100% - ${debugPanelHeight}px)` : '100%' }}
      >
        {/* Node Palette */}
        <NodePalette className="w-80" />

        {/* Canvas */}
        <div className="flex-1 relative">
          <ReactFlowProvider>
            <FlowCanvas flowId={id} flowName={currentFlow?.name} className="flex-1" />
          </ReactFlowProvider>

          {/* Zoom Controls */}
          <div className="absolute bottom-4 right-4 flex flex-col gap-1 bg-background border rounded-md shadow-lg">
            <Button
              variant="ghost"
              size="sm"
              onClick={handleZoomIn}
              className="h-9 w-9 p-0"
              title="Zoom In"
            >
              <ZoomIn className="h-4 w-4" />
            </Button>
            <div className="px-2 py-1 text-xs text-center border-t border-b">
              {zoom}%
            </div>
            <Button
              variant="ghost"
              size="sm"
              onClick={handleZoomOut}
              className="h-9 w-9 p-0"
              title="Zoom Out"
            >
              <ZoomOut className="h-4 w-4" />
            </Button>
            <Button
              variant="ghost"
              size="sm"
              onClick={handleFitView}
              className="h-9 w-9 p-0 border-t"
              title="Fit View"
            >
              <Maximize2 className="h-4 w-4" />
            </Button>
          </div>
        </div>
      </div>

      {/* Debug/Monitor/Terminal Panel */}
      {isDebugPanelOpen && (
        <div
          className="border-t bg-background"
          style={{ height: `${debugPanelHeight}px` }}
        >
          <div className="flex items-center justify-between px-4 py-2 border-b bg-muted/30">
            <div className="flex items-center gap-2">
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setDebugPanelHeight(prev => Math.max(prev + 50, 200))}
                className="h-6 w-6 p-0"
                title="Increase Height"
              >
                <ChevronUp className="h-3 w-3" />
              </Button>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => setDebugPanelHeight(prev => Math.max(prev - 50, 150))}
                className="h-6 w-6 p-0"
                title="Decrease Height"
              >
                <ChevronDown className="h-3 w-3" />
              </Button>
            </div>
            <Button
              variant="ghost"
              size="sm"
              onClick={toggleDebugPanel}
              className="h-6 w-6 p-0"
            >
              <X className="h-3 w-3" />
            </Button>
          </div>

          <Tabs defaultValue="debug" className="h-[calc(100%-40px)]">
            <TabsList className="px-4">
              <TabsTrigger value="debug" className="gap-2">
                <Bug className="h-3 w-3" />
                Debug
              </TabsTrigger>
              <TabsTrigger value="monitor" className="gap-2">
                <Activity className="h-3 w-3" />
                Monitor
              </TabsTrigger>
              <TabsTrigger value="terminal" className="gap-2">
                <TerminalIcon className="h-3 w-3" />
                Terminal
              </TabsTrigger>
            </TabsList>

            {/* Debug Tab */}
            <TabsContent value="debug" className="h-[calc(100%-40px)] m-0 p-0">
              <ScrollArea className="h-full">
                <div className="p-4 font-mono text-xs space-y-1">
                  {debugLogs.length === 0 ? (
                    <div className="text-muted-foreground text-center py-8">
                      No debug messages yet. Run your flow to see debug output.
                    </div>
                  ) : (
                    debugLogs.map((log, i) => (
                      <div key={i} className={cn(
                        "flex gap-2",
                        log.type === 'error' && 'text-red-500',
                        log.type === 'success' && 'text-green-500',
                        log.type === 'warning' && 'text-yellow-500',
                        log.type === 'info' && 'text-blue-500'
                      )}>
                        <span className="text-muted-foreground">[{log.time}]</span>
                        <span>{log.message}</span>
                      </div>
                    ))
                  )}
                </div>
              </ScrollArea>
            </TabsContent>

            {/* Monitor Tab */}
            <TabsContent value="monitor" className="h-[calc(100%-40px)] m-0 p-0">
              <ScrollArea className="h-full">
                <div className="p-4 space-y-2">
                  <div className="grid grid-cols-3 gap-4">
                    <div className="border rounded-lg p-3">
                      <div className="text-xs text-muted-foreground">Flow Status</div>
                      <div className="text-lg font-semibold">{currentFlow?.status || 'Unknown'}</div>
                    </div>
                    <div className="border rounded-lg p-3">
                      <div className="text-xs text-muted-foreground">Nodes</div>
                      <div className="text-lg font-semibold">{Object.keys(currentFlow?.nodes || {}).length}</div>
                    </div>
                    <div className="border rounded-lg p-3">
                      <div className="text-xs text-muted-foreground">Connections</div>
                      <div className="text-lg font-semibold">{(currentFlow?.connections || []).length}</div>
                    </div>
                  </div>
                  <div className="text-xs text-muted-foreground mt-4">
                    Real-time monitoring will show node execution status, message flow, and performance metrics here.
                  </div>
                </div>
              </ScrollArea>
            </TabsContent>

            {/* Terminal Tab */}
            <TabsContent value="terminal" className="h-[calc(100%-40px)] m-0 p-0">
              <div className="h-full flex flex-col bg-black text-green-400 font-mono text-sm">
                <ScrollArea className="flex-1 p-4">
                  {terminalOutput.map((line, i) => (
                    <div key={i} className="whitespace-pre-wrap">{line}</div>
                  ))}
                </ScrollArea>
                <div className="border-t border-green-900 px-4 py-2 flex items-center gap-2">
                  <span>$</span>
                  <input
                    type="text"
                    value={terminalInput}
                    onChange={(e) => setTerminalInput(e.target.value)}
                    onKeyDown={handleTerminalCommand}
                    className="flex-1 bg-transparent outline-none"
                    placeholder="Type command..."
                    autoFocus
                  />
                </div>
              </div>
            </TabsContent>
          </Tabs>
        </div>
      )}

      {/* Deploy Dialog */}
      <DeployDialog
        open={isDeployDialogOpen}
        onOpenChange={setIsDeployDialogOpen}
        flowIds={id ? [id] : undefined}
      />
    </div>
  )
}
