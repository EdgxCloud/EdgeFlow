import { useState, useEffect, useRef, useCallback } from 'react'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  Copy,
  Download,
  ChevronDown,
  ChevronRight,
  X,
  Table as TableIcon,
  Code2,
  Clock,
  Zap,
  AlertTriangle,
  CheckCircle2,
  XCircle,
  Trash2,
  Pin,
  PinOff,
  ArrowDown,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import wsClient from '@/services/websocket'
import type { WSMessage } from '@/services/websocket'

interface ExecutionEntry {
  id: string
  nodeId: string
  nodeName: string
  nodeType: string
  timestamp: number
  input: Record<string, unknown> | null
  output: Record<string, unknown> | null
  executionTime: number
  status: 'success' | 'error'
  error?: string
}

interface ExecutionDataPanelProps {
  nodeId?: string
  nodeName?: string
  onClose?: () => void
  className?: string
}

let execIdCounter = 0

export default function ExecutionDataPanel({
  nodeId,
  nodeName,
  onClose,
  className,
}: ExecutionDataPanelProps) {
  // All execution entries (across all nodes)
  const [allExecutions, setAllExecutions] = useState<ExecutionEntry[]>([])
  const [selectedExecution, setSelectedExecution] = useState<number>(0)
  const [expandedSections, setExpandedSections] = useState<Set<string>>(new Set(['input', 'output']))
  const [viewMode, setViewMode] = useState<'json' | 'table'>('json')
  const [autoScroll, setAutoScroll] = useState(true)
  const scrollRef = useRef<HTMLDivElement>(null)
  const [copiedId, setCopiedId] = useState<string | null>(null)

  // Subscribe to execution WebSocket messages
  useEffect(() => {
    if (!wsClient.isConnected()) {
      wsClient.connect()
    }

    const handleExecution = (msg: WSMessage) => {
      const d = msg.data as Record<string, unknown>
      const entry: ExecutionEntry = {
        id: `exec-${execIdCounter++}`,
        nodeId: (d.node_id as string) || '',
        nodeName: (d.node_name as string) || (d.node_id as string) || 'Node',
        nodeType: (d.node_type as string) || '',
        timestamp: (d.timestamp as number) || Date.now(),
        input: (d.input as Record<string, unknown>) || null,
        output: (d.output as Record<string, unknown>) || null,
        executionTime: (d.execution_time as number) || 0,
        status: (d.status as 'success' | 'error') || 'success',
        error: (d.error as string) || undefined,
      }
      setAllExecutions(prev => [...prev.slice(-200), entry])
    }

    const handleFlowStatus = (msg: WSMessage) => {
      const data = msg.data as Record<string, unknown>
      if (data.action === 'stopped') {
        // Flow stopped — clear all execution data
        setAllExecutions([])
        setSelectedExecution(0)
      }
    }

    const unsub = wsClient.on('execution', handleExecution)
    const unsub2 = wsClient.on('flow_status', handleFlowStatus)
    return () => { unsub(); unsub2() }
  }, [])

  // Filter executions for the selected node
  const nodeExecutions = nodeId
    ? allExecutions.filter(e => e.nodeId === nodeId)
    : allExecutions

  // Auto-select latest execution when new data arrives
  useEffect(() => {
    if (autoScroll && nodeExecutions.length > 0) {
      setSelectedExecution(nodeExecutions.length - 1)
    }
  }, [nodeExecutions.length, autoScroll])

  const currentExecution = nodeExecutions[selectedExecution] || null

  // Stats
  const totalExecs = nodeExecutions.length
  const successCount = nodeExecutions.filter(e => e.status === 'success').length
  const errorCount = nodeExecutions.filter(e => e.status === 'error').length
  const avgTime = totalExecs > 0
    ? Math.round(nodeExecutions.reduce((sum, e) => sum + e.executionTime, 0) / totalExecs)
    : 0

  const toggleSection = (key: string) => {
    setExpandedSections(prev => {
      const next = new Set(prev)
      if (next.has(key)) next.delete(key)
      else next.add(key)
      return next
    })
  }

  const handleCopy = useCallback((data: unknown, id?: string) => {
    navigator.clipboard.writeText(JSON.stringify(data, null, 2))
    if (id) {
      setCopiedId(id)
      setTimeout(() => setCopiedId(null), 1500)
    }
  }, [])

  const handleDownload = useCallback(() => {
    if (!currentExecution) return
    const dataStr = JSON.stringify(currentExecution, null, 2)
    const dataBlob = new Blob([dataStr], { type: 'application/json' })
    const url = URL.createObjectURL(dataBlob)
    const link = document.createElement('a')
    link.href = url
    link.download = `exec-${currentExecution.nodeName}-${currentExecution.id}.json`
    link.click()
    URL.revokeObjectURL(url)
  }, [currentExecution])

  const handleDownloadAll = useCallback(() => {
    const dataStr = JSON.stringify(nodeExecutions, null, 2)
    const dataBlob = new Blob([dataStr], { type: 'application/json' })
    const url = URL.createObjectURL(dataBlob)
    const link = document.createElement('a')
    link.href = url
    link.download = `executions-${nodeId || 'all'}.json`
    link.click()
    URL.revokeObjectURL(url)
  }, [nodeExecutions, nodeId])

  const clearExecutions = useCallback(() => {
    if (nodeId) {
      setAllExecutions(prev => prev.filter(e => e.nodeId !== nodeId))
    } else {
      setAllExecutions([])
    }
    setSelectedExecution(0)
  }, [nodeId])

  const formatTimeFull = (ts: number) => {
    try {
      return new Date(ts).toLocaleString('en-US', { hour12: false })
    } catch {
      return String(ts)
    }
  }

  // Render table view for object data
  const renderTableView = (data: Record<string, unknown> | null) => {
    if (!data) return <span className="text-xs text-muted-foreground">null</span>

    return (
      <div className="border border-border rounded-lg overflow-hidden">
        <table className="w-full text-xs">
          <thead>
            <tr className="bg-muted/50">
              <th className="text-left px-3 py-1.5 font-medium text-muted-foreground border-b border-border">Key</th>
              <th className="text-left px-3 py-1.5 font-medium text-muted-foreground border-b border-border">Type</th>
              <th className="text-left px-3 py-1.5 font-medium text-muted-foreground border-b border-border">Value</th>
            </tr>
          </thead>
          <tbody>
            {Object.entries(data).map(([key, val]) => (
              <tr key={key} className="border-b border-border last:border-0 hover:bg-muted/30">
                <td className="px-3 py-1.5 font-mono font-medium text-blue-400">{key}</td>
                <td className="px-3 py-1.5">
                  <Badge variant="secondary" className="text-[9px] h-4">
                    {typeof val === 'object' ? (val === null ? 'null' : Array.isArray(val) ? 'array' : 'object') : typeof val}
                  </Badge>
                </td>
                <td className="px-3 py-1.5 font-mono text-muted-foreground max-w-[200px] truncate">
                  {typeof val === 'object' ? JSON.stringify(val) : String(val)}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    )
  }

  const renderDataSection = (label: string, data: Record<string, unknown> | null, sectionKey: string) => {
    const isExpanded = expandedSections.has(sectionKey)
    const keyCount = data ? Object.keys(data).length : 0

    return (
      <div className="border border-border rounded-lg overflow-hidden">
        {/* Section header */}
        <button
          className="w-full flex items-center gap-2 px-3 py-2 bg-muted/30 hover:bg-muted/50 transition-colors"
          onClick={() => toggleSection(sectionKey)}
        >
          {isExpanded
            ? <ChevronDown className="h-3.5 w-3.5 text-muted-foreground" />
            : <ChevronRight className="h-3.5 w-3.5 text-muted-foreground" />
          }
          <span className="text-xs font-semibold uppercase tracking-wide">{label}</span>
          <Badge variant="secondary" className="text-[10px] h-4 ml-auto">
            {keyCount} {keyCount === 1 ? 'key' : 'keys'}
          </Badge>
          <Button
            variant="ghost"
            size="sm"
            className="h-5 w-5 p-0 ml-1"
            onClick={(e) => { e.stopPropagation(); handleCopy(data, `${sectionKey}-copy`) }}
          >
            <Copy className={cn('h-3 w-3', copiedId === `${sectionKey}-copy` && 'text-green-500')} />
          </Button>
        </button>

        {/* Section content */}
        {isExpanded && (
          <div className="p-3">
            {!data ? (
              <span className="text-xs text-muted-foreground italic">No data</span>
            ) : viewMode === 'table' ? (
              renderTableView(data)
            ) : (
              <pre className="text-xs font-mono bg-black/20 dark:bg-black/40 p-3 rounded-lg overflow-x-auto text-gray-300 leading-relaxed">
                {JSON.stringify(data, null, 2)}
              </pre>
            )}
          </div>
        )}
      </div>
    )
  }

  // Empty state
  if (!nodeId) {
    return (
      <div className={cn('flex flex-col bg-card border-l border-border', className)}>
        <div className="flex items-center justify-between p-3 border-b border-border">
          <h3 className="font-semibold text-sm">Execution Data</h3>
          {onClose && (
            <Button variant="ghost" size="sm" className="h-7 w-7 p-0" onClick={onClose}>
              <X className="h-3.5 w-3.5" />
            </Button>
          )}
        </div>
        <div className="flex-1 flex flex-col items-center justify-center p-6 text-center text-muted-foreground gap-3">
          <Zap className="h-10 w-10 opacity-20" />
          <div>
            <p className="text-sm font-medium">No node selected</p>
            <p className="text-xs mt-1">Click on a node to inspect its execution data</p>
          </div>
          {allExecutions.length > 0 && (
            <Badge variant="secondary" className="mt-2 text-xs">
              {allExecutions.length} total executions captured
            </Badge>
          )}
        </div>
      </div>
    )
  }

  return (
    <div className={cn('flex flex-col bg-card border-l border-border', className)}>
      {/* Header */}
      <div className="flex items-center gap-2 p-3 border-b border-border">
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2">
            <h3 className="font-semibold text-sm truncate">{nodeName || nodeId}</h3>
            {currentExecution?.nodeType && (
              <Badge variant="outline" className="text-[10px] h-4 shrink-0">
                {currentExecution.nodeType}
              </Badge>
            )}
          </div>
          <p className="text-[11px] text-muted-foreground mt-0.5">
            {totalExecs} execution{totalExecs !== 1 ? 's' : ''} captured
          </p>
        </div>

        <div className="flex items-center gap-0.5">
          <Button
            variant="ghost" size="sm" className="h-7 w-7 p-0"
            onClick={() => setViewMode(viewMode === 'json' ? 'table' : 'json')}
            title={viewMode === 'json' ? 'Switch to table view' : 'Switch to JSON view'}
          >
            {viewMode === 'json' ? <TableIcon className="h-3.5 w-3.5" /> : <Code2 className="h-3.5 w-3.5" />}
          </Button>
          <Button
            variant="ghost" size="sm" className="h-7 w-7 p-0"
            onClick={() => setAutoScroll(!autoScroll)}
            title={autoScroll ? 'Unpin from latest' : 'Pin to latest'}
          >
            {autoScroll
              ? <Pin className="h-3.5 w-3.5 text-blue-500" />
              : <PinOff className="h-3.5 w-3.5" />
            }
          </Button>
          <Button variant="ghost" size="sm" className="h-7 w-7 p-0" onClick={handleDownloadAll} title="Download all">
            <Download className="h-3.5 w-3.5" />
          </Button>
          <Button variant="ghost" size="sm" className="h-7 w-7 p-0" onClick={clearExecutions} title="Clear">
            <Trash2 className="h-3.5 w-3.5" />
          </Button>
          {onClose && (
            <Button variant="ghost" size="sm" className="h-7 w-7 p-0" onClick={onClose}>
              <X className="h-3.5 w-3.5" />
            </Button>
          )}
        </div>
      </div>

      {/* Stats bar */}
      {totalExecs > 0 && (
        <div className="flex items-center gap-3 px-3 py-2 border-b border-border bg-muted/20 text-[11px]">
          <div className="flex items-center gap-1 text-green-500">
            <CheckCircle2 className="h-3 w-3" />
            <span>{successCount}</span>
          </div>
          <div className="flex items-center gap-1 text-red-500">
            <XCircle className="h-3 w-3" />
            <span>{errorCount}</span>
          </div>
          <div className="w-px h-3 bg-border" />
          <div className="flex items-center gap-1 text-muted-foreground">
            <Clock className="h-3 w-3" />
            <span>avg {avgTime}ms</span>
          </div>
          {currentExecution && (
            <>
              <div className="w-px h-3 bg-border" />
              <div className="flex items-center gap-1 text-muted-foreground">
                <Zap className="h-3 w-3" />
                <span>{currentExecution.executionTime}ms</span>
              </div>
            </>
          )}
        </div>
      )}

      {/* Execution timeline */}
      {nodeExecutions.length > 1 && (
        <div className="border-b border-border">
          <ScrollArea className="w-full">
            <div className="flex gap-1 p-2 overflow-x-auto">
              {nodeExecutions.map((exec, idx) => (
                <button
                  key={exec.id}
                  className={cn(
                    'shrink-0 flex items-center gap-1.5 px-2.5 py-1 rounded-md text-[11px] font-medium transition-colors border',
                    selectedExecution === idx
                      ? 'bg-primary text-primary-foreground border-primary'
                      : 'bg-muted/30 hover:bg-muted/60 border-transparent'
                  )}
                  onClick={() => { setSelectedExecution(idx); setAutoScroll(false) }}
                >
                  <span className={cn(
                    'w-1.5 h-1.5 rounded-full',
                    exec.status === 'success' ? 'bg-green-500' : 'bg-red-500'
                  )} />
                  #{idx + 1}
                  <span className="text-[10px] opacity-60">{exec.executionTime}ms</span>
                </button>
              ))}
            </div>
          </ScrollArea>
        </div>
      )}

      {/* Main content */}
      <ScrollArea className="flex-1" ref={scrollRef}>
        {!currentExecution ? (
          <div className="flex flex-col items-center justify-center p-8 text-center text-muted-foreground gap-3">
            <ArrowDown className="h-8 w-8 opacity-20 animate-bounce" />
            <div>
              <p className="text-sm font-medium">Waiting for execution data</p>
              <p className="text-xs mt-1">Run the workflow to see real-time results</p>
            </div>
          </div>
        ) : (
          <div className="p-3 space-y-3">
            {/* Execution metadata card */}
            <div className={cn(
              'rounded-lg border p-3 space-y-2',
              currentExecution.status === 'error'
                ? 'border-red-500/30 bg-red-500/5'
                : 'border-green-500/30 bg-green-500/5'
            )}>
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  {currentExecution.status === 'success'
                    ? <CheckCircle2 className="h-4 w-4 text-green-500" />
                    : <XCircle className="h-4 w-4 text-red-500" />
                  }
                  <span className={cn(
                    'text-sm font-semibold',
                    currentExecution.status === 'success' ? 'text-green-500' : 'text-red-500'
                  )}>
                    {currentExecution.status === 'success' ? 'Success' : 'Error'}
                  </span>
                </div>
                <div className="flex items-center gap-2 text-xs text-muted-foreground">
                  <Clock className="h-3 w-3" />
                  {currentExecution.executionTime}ms
                </div>
              </div>
              <div className="text-[11px] text-muted-foreground">
                {formatTimeFull(currentExecution.timestamp)}
              </div>

              {currentExecution.error && (
                <div className="flex items-start gap-2 mt-2 p-2 bg-red-500/10 rounded-md">
                  <AlertTriangle className="h-3.5 w-3.5 text-red-500 shrink-0 mt-0.5" />
                  <pre className="text-xs font-mono text-red-400 whitespace-pre-wrap break-all">
                    {currentExecution.error}
                  </pre>
                </div>
              )}
            </div>

            {/* Input section */}
            {renderDataSection('Input', currentExecution.input, 'input')}

            {/* Output section */}
            {renderDataSection('Output', currentExecution.output, 'output')}

            {/* Raw JSON (collapsed by default) */}
            <div className="border border-border rounded-lg overflow-hidden">
              <button
                className="w-full flex items-center gap-2 px-3 py-2 bg-muted/30 hover:bg-muted/50 transition-colors"
                onClick={() => toggleSection('raw')}
              >
                {expandedSections.has('raw')
                  ? <ChevronDown className="h-3.5 w-3.5 text-muted-foreground" />
                  : <ChevronRight className="h-3.5 w-3.5 text-muted-foreground" />
                }
                <span className="text-xs font-semibold uppercase tracking-wide">Full Execution Data</span>
                <Button
                  variant="ghost" size="sm" className="h-5 w-5 p-0 ml-auto"
                  onClick={(e) => { e.stopPropagation(); handleDownload() }}
                >
                  <Download className="h-3 w-3" />
                </Button>
              </button>
              {expandedSections.has('raw') && (
                <div className="p-3">
                  <pre className="text-xs font-mono bg-black/20 dark:bg-black/40 p-3 rounded-lg overflow-x-auto text-gray-300 leading-relaxed max-h-64 overflow-y-auto">
                    {JSON.stringify(currentExecution, null, 2)}
                  </pre>
                </div>
              )}
            </div>
          </div>
        )}
      </ScrollArea>
    </div>
  )
}
