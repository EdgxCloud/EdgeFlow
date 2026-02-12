import { useState, useEffect, useRef, useCallback } from 'react'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Trash2,
  Search,
  Copy,
  ChevronDown,
  ChevronRight,
  CheckCircle2,
  XCircle,
  Clock,
  ArrowDown,
  Pause,
  Play,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import wsClient from '@/services/websocket'
import type { WSMessage } from '@/services/websocket'

interface DebugMessage {
  id: string
  timestamp: number
  nodeId: string
  nodeName: string
  nodeType: string
  topic: string
  input: unknown
  output: unknown
  level: 'info' | 'warn' | 'error' | 'debug' | 'success'
  executionTime?: number
  status?: 'success' | 'error'
  errorMsg?: string
}

let debugCounter = 0

interface DebugPanelProps {
  className?: string
}

export default function DebugPanel({ className }: DebugPanelProps) {
  const [messages, setMessages] = useState<DebugMessage[]>([])
  const [filterLevel, setFilterLevel] = useState<string>('all')
  const [filterNode, setFilterNode] = useState<string>('all')
  const [searchQuery, setSearchQuery] = useState('')
  const [expandedMessages, setExpandedMessages] = useState<Set<string>>(new Set())
  const [isPaused, setIsPaused] = useState(false)
  const [autoScroll, setAutoScroll] = useState(true)
  const pausedRef = useRef<DebugMessage[]>([])
  const isPausedRef = useRef(false)
  const scrollRef = useRef<HTMLDivElement>(null)

  // Keep ref in sync with state
  useEffect(() => { isPausedRef.current = isPaused }, [isPaused])

  const addMessage = useCallback((entry: DebugMessage) => {
    if (isPausedRef.current) {
      pausedRef.current.push(entry)
    } else {
      setMessages(prev => [...prev.slice(-300), entry])
    }
  }, [])

  // Auto-scroll to bottom
  useEffect(() => {
    if (autoScroll && scrollRef.current && !isPaused) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight
    }
  }, [messages, autoScroll, isPaused])

  // Subscribe to WebSocket execution and node_status messages
  useEffect(() => {
    if (!wsClient.isConnected()) {
      wsClient.connect()
    }

    // Real execution data from node processing
    const handleExecution = (msg: WSMessage) => {
      const data = msg.data as Record<string, unknown>
      const status = (data.status as string) || 'success'
      const entry: DebugMessage = {
        id: `debug-${debugCounter++}`,
        timestamp: (data.timestamp as number) || Date.now(),
        nodeId: (data.node_id as string) || 'unknown',
        nodeName: (data.node_name as string) || (data.node_id as string) || 'Node',
        nodeType: (data.node_type as string) || '',
        topic: '',
        input: data.input || null,
        output: data.output || null,
        level: status === 'error' ? 'error' : 'success',
        executionTime: (data.execution_time as number) || 0,
        status: status as 'success' | 'error',
        errorMsg: (data.error as string) || undefined,
      }
      addMessage(entry)
    }

    // Node status change events
    const handleNodeStatus = (msg: WSMessage) => {
      const data = msg.data as Record<string, unknown>
      const level = data.status === 'error' ? 'error'
        : data.status === 'warning' ? 'warn'
        : 'debug'
      const entry: DebugMessage = {
        id: `debug-${debugCounter++}`,
        timestamp: Date.now(),
        nodeId: (data.node_id as string) || 'unknown',
        nodeName: (data.node_id as string) || 'Node',
        nodeType: '',
        topic: '',
        input: null,
        output: { status: data.status, message: data.message, text: data.text },
        level: level as DebugMessage['level'],
      }
      addMessage(entry)
    }

    // Stop receiving debug messages when flow is stopped
    const handleFlowStatus = (msg: WSMessage) => {
      const data = msg.data as Record<string, unknown>
      if (data.action === 'stopped') {
        setMessages([])
        pausedRef.current = []
      }
    }

    const unsub1 = wsClient.on('execution', handleExecution)
    const unsub2 = wsClient.on('node_status', handleNodeStatus)
    const unsub3 = wsClient.on('flow_status', handleFlowStatus)

    return () => { unsub1(); unsub2(); unsub3() }
  }, [addMessage])

  const handleResume = useCallback(() => {
    setIsPaused(false)
    if (pausedRef.current.length > 0) {
      setMessages(prev => [...prev, ...pausedRef.current].slice(-300))
      pausedRef.current = []
    }
  }, [])

  const handleClear = () => {
    setMessages([])
    pausedRef.current = []
  }

  const handleCopy = (message: DebugMessage) => {
    const text = JSON.stringify(
      {
        timestamp: new Date(message.timestamp).toISOString(),
        node: message.nodeName,
        nodeType: message.nodeType,
        status: message.status,
        executionTime: message.executionTime,
        input: message.input,
        output: message.output,
        error: message.errorMsg,
      },
      null,
      2
    )
    navigator.clipboard.writeText(text)
  }

  const toggleExpand = (id: string) => {
    setExpandedMessages((prev) => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }

  // Get unique node names for filter
  const nodeNames = Array.from(new Set(messages.map(m => m.nodeName))).sort()

  // Filter messages
  const filteredMessages = messages.filter((msg) => {
    if (filterLevel !== 'all' && msg.level !== filterLevel) return false
    if (filterNode !== 'all' && msg.nodeName !== filterNode) return false
    if (searchQuery && !JSON.stringify(msg).toLowerCase().includes(searchQuery.toLowerCase()))
      return false
    return true
  })

  const getLevelColor = (level: string) => {
    switch (level) {
      case 'error': return 'text-red-500 bg-red-500/10'
      case 'warn': return 'text-yellow-500 bg-yellow-500/10'
      case 'success': return 'text-green-500 bg-green-500/10'
      case 'debug': return 'text-blue-500 bg-blue-500/10'
      default: return 'text-gray-500 bg-gray-500/10'
    }
  }

  const formatTime = (ts: number) => {
    try {
      return new Date(ts).toLocaleTimeString('en-US', { hour12: false })
    } catch {
      return String(ts)
    }
  }

  return (
    <div className={cn('flex flex-col h-full', className)}>
      {/* Toolbar */}
      <div className="flex items-center gap-2 px-3 py-1.5 border-b border-gray-200 dark:border-gray-800">
        {/* Search */}
        <div className="relative flex-1 max-w-xs">
          <Search className="absolute left-2 top-1/2 -translate-y-1/2 w-3 h-3 text-muted-foreground" />
          <Input
            type="search"
            placeholder="Search debug..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="h-7 pl-7 text-xs"
          />
        </div>

        {/* Level filter */}
        <Select value={filterLevel} onValueChange={setFilterLevel}>
          <SelectTrigger className="w-24 h-7 text-xs">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All</SelectItem>
            <SelectItem value="success">Success</SelectItem>
            <SelectItem value="error">Error</SelectItem>
            <SelectItem value="warn">Warning</SelectItem>
            <SelectItem value="debug">Debug</SelectItem>
            <SelectItem value="info">Info</SelectItem>
          </SelectContent>
        </Select>

        {/* Node filter */}
        {nodeNames.length > 1 && (
          <Select value={filterNode} onValueChange={setFilterNode}>
            <SelectTrigger className="w-28 h-7 text-xs">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All Nodes</SelectItem>
              {nodeNames.map(name => (
                <SelectItem key={name} value={name}>{name}</SelectItem>
              ))}
            </SelectContent>
          </Select>
        )}

        {/* Pause/Resume */}
        <Button
          variant="ghost" size="sm" className="h-7 text-xs gap-1"
          onClick={() => isPaused ? handleResume() : setIsPaused(true)}
        >
          {isPaused ? (
            <>
              <Play className="w-3 h-3" />
              {pausedRef.current.length > 0 && (
                <span className="text-yellow-500">({pausedRef.current.length})</span>
              )}
            </>
          ) : (
            <Pause className="w-3 h-3" />
          )}
        </Button>

        {/* Auto-scroll */}
        <Button
          variant="ghost" size="sm" className="h-7 w-7 p-0"
          onClick={() => setAutoScroll(!autoScroll)}
        >
          <ArrowDown className={cn('w-3 h-3', autoScroll && 'text-blue-500')} />
        </Button>

        {/* Clear */}
        <Button variant="ghost" size="sm" className="h-7 w-7 p-0" onClick={handleClear}>
          <Trash2 className="w-3 h-3" />
        </Button>

        <span className="text-xs text-muted-foreground">{filteredMessages.length}</span>
      </div>

      {/* Messages */}
      <div ref={scrollRef} className="flex-1 overflow-auto font-mono text-xs">
        {filteredMessages.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-full text-muted-foreground gap-2">
            <p className="text-sm">No debug messages</p>
            <p className="text-xs">Execution events will appear here when flows run</p>
          </div>
        ) : (
          filteredMessages.map((message) => {
            const isExpanded = expandedMessages.has(message.id)

            return (
              <div
                key={message.id}
                className={cn(
                  'px-3 py-1 hover:bg-muted/30 border-b border-border/30',
                  message.level === 'error' && 'bg-red-500/5'
                )}
              >
                {/* Collapsed row */}
                <div className="flex items-center gap-2">
                  <Button
                    variant="ghost" size="sm" className="h-5 w-5 p-0 shrink-0"
                    onClick={() => toggleExpand(message.id)}
                  >
                    {isExpanded
                      ? <ChevronDown className="h-3 w-3" />
                      : <ChevronRight className="h-3 w-3" />
                    }
                  </Button>

                  {/* Status icon */}
                  {message.status === 'success' ? (
                    <CheckCircle2 className="h-3 w-3 text-green-500 shrink-0" />
                  ) : message.status === 'error' ? (
                    <XCircle className="h-3 w-3 text-red-500 shrink-0" />
                  ) : (
                    <span className={cn(
                      'shrink-0 w-12 uppercase text-[10px] font-medium px-1 rounded text-center',
                      getLevelColor(message.level)
                    )}>
                      {message.level}
                    </span>
                  )}

                  {/* Timestamp */}
                  <span className="text-muted-foreground shrink-0 w-16">
                    {formatTime(message.timestamp)}
                  </span>

                  {/* Node name */}
                  <span className="font-medium text-blue-400 shrink-0 max-w-[100px] truncate">
                    {message.nodeName}
                  </span>

                  {/* Node type */}
                  {message.nodeType && (
                    <Badge variant="secondary" className="text-[9px] h-3.5 shrink-0">
                      {message.nodeType}
                    </Badge>
                  )}

                  {/* Execution time */}
                  {message.executionTime !== undefined && (
                    <span className="text-muted-foreground shrink-0 flex items-center gap-0.5">
                      <Clock className="h-2.5 w-2.5" />
                      {message.executionTime}ms
                    </span>
                  )}

                  {/* Error preview */}
                  {message.errorMsg && !isExpanded && (
                    <span className="text-red-400 truncate flex-1 ml-1">
                      {message.errorMsg}
                    </span>
                  )}

                  {/* Output preview */}
                  {!message.errorMsg && !isExpanded && message.output && (
                    <span className="text-muted-foreground truncate flex-1 ml-1">
                      {typeof message.output === 'object' ? JSON.stringify(message.output) : String(message.output)}
                    </span>
                  )}

                  {/* Copy */}
                  <Button
                    variant="ghost" size="sm" className="h-5 w-5 p-0 shrink-0 ml-auto"
                    onClick={() => handleCopy(message)}
                  >
                    <Copy className="h-2.5 w-2.5" />
                  </Button>
                </div>

                {/* Expanded view */}
                {isExpanded && (
                  <div className="ml-7 mt-2 mb-1 space-y-2">
                    {message.errorMsg && (
                      <div className="p-2 bg-red-500/10 border border-red-500/20 rounded text-red-400">
                        <span className="font-semibold text-[10px] uppercase">Error: </span>
                        {message.errorMsg}
                      </div>
                    )}
                    {message.input && (
                      <div>
                        <span className="text-[10px] font-semibold uppercase text-muted-foreground">Input</span>
                        <pre className="mt-1 text-xs bg-black/20 dark:bg-black/40 p-2 rounded overflow-x-auto text-gray-300">
                          {JSON.stringify(message.input, null, 2)}
                        </pre>
                      </div>
                    )}
                    {message.output && (
                      <div>
                        <span className="text-[10px] font-semibold uppercase text-muted-foreground">Output</span>
                        <pre className="mt-1 text-xs bg-black/20 dark:bg-black/40 p-2 rounded overflow-x-auto text-gray-300">
                          {JSON.stringify(message.output, null, 2)}
                        </pre>
                      </div>
                    )}
                  </div>
                )}
              </div>
            )
          })
        )}
      </div>
    </div>
  )
}
