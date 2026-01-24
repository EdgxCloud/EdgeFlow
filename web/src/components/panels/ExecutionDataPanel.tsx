import { useState } from 'react'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  Copy,
  Download,
  ChevronDown,
  ChevronRight,
  X,
  Table as TableIcon,
  Code2,
} from 'lucide-react'
import { cn } from '@/lib/utils'

interface ExecutionData {
  nodeId: string
  nodeName: string
  executions: {
    id: string
    timestamp: number
    input: any
    output: any
    executionTime: number
    status: 'success' | 'error'
    error?: string
  }[]
}

interface ExecutionDataPanelProps {
  nodeId?: string
  nodeName?: string
  onClose?: () => void
  className?: string
}

// Mock execution data - in real implementation, this would come from the backend
const getMockExecutionData = (nodeId: string, nodeName: string): ExecutionData => {
  return {
    nodeId,
    nodeName,
    executions: [
      {
        id: '1',
        timestamp: Date.now() - 10000,
        input: {
          payload: { temperature: 25.5, humidity: 60, sensor: 'DHT22' },
          topic: 'home/sensor/data',
          timestamp: Date.now() - 10000,
        },
        output: {
          payload: {
            temperature: 25.5,
            humidity: 60,
            sensor: 'DHT22',
            status: 'normal',
            processed: true,
          },
          topic: 'home/sensor/processed',
          timestamp: Date.now() - 10000,
        },
        executionTime: 15,
        status: 'success',
      },
      {
        id: '2',
        timestamp: Date.now() - 40000,
        input: {
          payload: { temperature: 22.3, humidity: 55, sensor: 'DHT22' },
          topic: 'home/sensor/data',
          timestamp: Date.now() - 40000,
        },
        output: {
          payload: {
            temperature: 22.3,
            humidity: 55,
            sensor: 'DHT22',
            status: 'normal',
            processed: true,
          },
          topic: 'home/sensor/processed',
          timestamp: Date.now() - 40000,
        },
        executionTime: 12,
        status: 'success',
      },
    ],
  }
}

export default function ExecutionDataPanel({
  nodeId,
  nodeName,
  onClose,
  className,
}: ExecutionDataPanelProps) {
  const [expandedItems, setExpandedItems] = useState<Set<string>>(new Set(['input', 'output']))
  const [selectedExecution, setSelectedExecution] = useState(0)
  const [viewMode, setViewMode] = useState<'json' | 'table'>('json')

  if (!nodeId) {
    return (
      <div className={cn('flex flex-col bg-card border-l border-border', className)}>
        <div className="flex items-center justify-between p-4 border-b border-border">
          <h3 className="font-semibold text-sm">Node Execution Data</h3>
        </div>
        <div className="flex-1 flex items-center justify-center p-8 text-center text-muted-foreground">
          <div>
            <p className="text-sm">No node selected</p>
            <p className="text-xs mt-1">Click on a node to see its execution data</p>
          </div>
        </div>
      </div>
    )
  }

  const executionData = getMockExecutionData(nodeId, nodeName || 'Unknown')
  const currentExecution = executionData.executions[selectedExecution]

  const toggleExpand = (key: string) => {
    setExpandedItems((prev) => {
      const next = new Set(prev)
      if (next.has(key)) {
        next.delete(key)
      } else {
        next.add(key)
      }
      return next
    })
  }

  const handleCopy = (data: any) => {
    navigator.clipboard.writeText(JSON.stringify(data, null, 2))
  }

  const handleDownload = () => {
    const dataStr = JSON.stringify(currentExecution, null, 2)
    const dataBlob = new Blob([dataStr], { type: 'application/json' })
    const url = URL.createObjectURL(dataBlob)
    const link = document.createElement('a')
    link.href = url
    link.download = `execution-${nodeId}-${currentExecution.id}.json`
    link.click()
    URL.revokeObjectURL(url)
  }

  const renderJsonView = (data: any, parentKey: string) => {
    const isExpanded = expandedItems.has(parentKey)

    return (
      <div className="space-y-1">
        <div className="flex items-center gap-2">
          <Button
            variant="ghost"
            size="sm"
            className="h-5 w-5 p-0"
            onClick={() => toggleExpand(parentKey)}
          >
            {isExpanded ? (
              <ChevronDown className="h-3 w-3" />
            ) : (
              <ChevronRight className="h-3 w-3" />
            )}
          </Button>
          <span className="text-xs font-semibold capitalize">{parentKey}</span>
          <Badge variant="secondary" className="text-[10px] ml-auto">
            {typeof data === 'object'
              ? `${Object.keys(data).length} ${Object.keys(data).length === 1 ? 'key' : 'keys'}`
              : typeof data}
          </Badge>
          <Button
            variant="ghost"
            size="sm"
            className="h-5 w-5 p-0"
            onClick={() => handleCopy(data)}
          >
            <Copy className="h-3 w-3" />
          </Button>
        </div>

        {isExpanded && (
          <div className="ml-4 mt-2">
            <pre className="text-xs font-mono bg-muted p-3 rounded-lg overflow-x-auto">
              {JSON.stringify(data, null, 2)}
            </pre>
          </div>
        )}
      </div>
    )
  }

  return (
    <div className={cn('flex flex-col bg-card border-l border-border', className)}>
      {/* Header */}
      <div className="flex items-center justify-between p-3 border-b border-border">
        <div className="flex-1 min-w-0">
          <h3 className="font-semibold text-sm truncate">{executionData.nodeName}</h3>
          <p className="text-xs text-muted-foreground truncate">Execution Data</p>
        </div>

        <div className="flex items-center gap-1">
          <Button
            variant="ghost"
            size="sm"
            className="h-7 w-7 p-0"
            onClick={() => setViewMode(viewMode === 'json' ? 'table' : 'json')}
          >
            {viewMode === 'json' ? (
              <TableIcon className="h-3 w-3" />
            ) : (
              <Code2 className="h-3 w-3" />
            )}
          </Button>
          <Button
            variant="ghost"
            size="sm"
            className="h-7 w-7 p-0"
            onClick={handleDownload}
          >
            <Download className="h-3 w-3" />
          </Button>
          {onClose && (
            <Button variant="ghost" size="sm" className="h-7 w-7 p-0" onClick={onClose}>
              <X className="h-3 w-3" />
            </Button>
          )}
        </div>
      </div>

      {/* Execution selector */}
      {executionData.executions.length > 1 && (
        <div className="p-2 border-b border-border">
          <div className="flex gap-1 overflow-x-auto">
            {executionData.executions.map((exec, idx) => (
              <Button
                key={exec.id}
                variant={selectedExecution === idx ? 'default' : 'outline'}
                size="sm"
                className="h-7 text-xs"
                onClick={() => setSelectedExecution(idx)}
              >
                #{executionData.executions.length - idx}
                <Badge
                  variant={exec.status === 'success' ? 'default' : 'destructive'}
                  className="ml-2 h-4 text-[10px]"
                >
                  {exec.status}
                </Badge>
              </Button>
            ))}
          </div>
        </div>
      )}

      {/* Main content */}
      <ScrollArea className="flex-1">
        {!currentExecution ? (
          <div className="p-8 text-center text-muted-foreground">
            <p className="text-sm">No execution data available</p>
            <p className="text-xs mt-1">Run the workflow to see execution results</p>
          </div>
        ) : (
          <div className="p-4 space-y-4">
            {/* Execution metadata */}
            <div className="flex items-center justify-between text-xs">
              <div className="flex items-center gap-2">
                <Badge variant={currentExecution.status === 'success' ? 'default' : 'destructive'}>
                  {currentExecution.status}
                </Badge>
                <span className="text-muted-foreground">
                  {new Date(currentExecution.timestamp).toLocaleString()}
                </span>
              </div>
              <span className="text-muted-foreground">
                {currentExecution.executionTime}ms
              </span>
            </div>

            {/* Tabs for Input/Output */}
            <Tabs defaultValue="input" className="w-full">
              <TabsList className="grid w-full grid-cols-2">
                <TabsTrigger value="input" className="text-xs">
                  Input
                  {currentExecution.input && (
                    <Badge variant="secondary" className="ml-2 h-4 text-[10px]">
                      {typeof currentExecution.input === 'object'
                        ? Object.keys(currentExecution.input).length
                        : 1}
                    </Badge>
                  )}
                </TabsTrigger>
                <TabsTrigger value="output" className="text-xs">
                  Output
                  {currentExecution.output && (
                    <Badge variant="secondary" className="ml-2 h-4 text-[10px]">
                      {typeof currentExecution.output === 'object'
                        ? Object.keys(currentExecution.output).length
                        : 1}
                    </Badge>
                  )}
                </TabsTrigger>
              </TabsList>

              <TabsContent value="input" className="space-y-3 mt-4">
                {renderJsonView(currentExecution.input, 'input')}
              </TabsContent>

              <TabsContent value="output" className="space-y-3 mt-4">
                {currentExecution.error ? (
                  <div className="p-3 bg-destructive/10 border border-destructive rounded-lg">
                    <p className="text-sm font-semibold text-destructive mb-2">Error</p>
                    <pre className="text-xs font-mono">{currentExecution.error}</pre>
                  </div>
                ) : (
                  renderJsonView(currentExecution.output, 'output')
                )}
              </TabsContent>
            </Tabs>
          </div>
        )}
      </ScrollArea>
    </div>
  )
}
