import { useState, useEffect } from 'react'
import { ScrollArea } from '@/components/ui/scroll-area'
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
  Filter,
  Search,
  Copy,
  ChevronDown,
  ChevronRight,
} from 'lucide-react'
import { cn } from '@/lib/utils'
import wsClient from '@/services/websocket'
import type { WSMessage } from '@/services/websocket'

interface DebugMessage {
  id: string
  timestamp: number
  nodeId: string
  nodeName: string
  topic: string
  payload: any
  level: 'info' | 'warn' | 'error' | 'debug'
}

let debugCounter = 0

interface DebugPanelProps {
  className?: string
}

export default function DebugPanel({ className }: DebugPanelProps) {
  const [messages, setMessages] = useState<DebugMessage[]>([])
  const [filterLevel, setFilterLevel] = useState<string>('all')
  const [searchQuery, setSearchQuery] = useState('')
  const [expandedMessages, setExpandedMessages] = useState<Set<string>>(new Set())

  // Subscribe to WebSocket execution and node_status messages
  useEffect(() => {
    if (!wsClient.isConnected()) {
      wsClient.connect()
    }

    const handleExecution = (msg: WSMessage) => {
      const data = msg.data as Record<string, unknown>
      const entry: DebugMessage = {
        id: `debug-${debugCounter++}`,
        timestamp: Date.now(),
        nodeId: (data.node_id as string) || 'unknown',
        nodeName: (data.node_name as string) || (data.node_id as string) || 'Node',
        topic: (data.topic as string) || '',
        payload: data.msg || data.payload || data,
        level: 'info',
      }
      setMessages(prev => [...prev.slice(-200), entry])
    }

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
        topic: '',
        payload: { status: data.status, message: data.message, text: data.text },
        level: level as DebugMessage['level'],
      }
      setMessages(prev => [...prev.slice(-200), entry])
    }

    const unsub1 = wsClient.on('execution', handleExecution)
    const unsub2 = wsClient.on('node_status', handleNodeStatus)

    return () => { unsub1(); unsub2() }
  }, [])

  const handleClear = () => {
    setMessages([])
  }

  const handleCopy = (message: DebugMessage) => {
    const text = JSON.stringify(
      {
        timestamp: new Date(message.timestamp).toISOString(),
        node: message.nodeName,
        topic: message.topic,
        payload: message.payload,
      },
      null,
      2
    )
    navigator.clipboard.writeText(text)
  }

  const toggleExpand = (id: string) => {
    setExpandedMessages((prev) => {
      const next = new Set(prev)
      if (next.has(id)) {
        next.delete(id)
      } else {
        next.add(id)
      }
      return next
    })
  }

  // Filter messages
  const filteredMessages = messages.filter((msg) => {
    if (filterLevel !== 'all' && msg.level !== filterLevel) return false
    if (searchQuery && !JSON.stringify(msg).toLowerCase().includes(searchQuery.toLowerCase()))
      return false
    return true
  })

  const getLevelColor = (level: string) => {
    switch (level) {
      case 'error':
        return 'destructive'
      case 'warn':
        return 'warning'
      case 'debug':
        return 'secondary'
      default:
        return 'default'
    }
  }

  return (
    <div className={cn('flex flex-col bg-card border-t border-border', className)}>
      {/* Header */}
      <div className="flex items-center justify-between p-3 border-b border-border">
        <div className="flex items-center gap-2 flex-1">
          <h3 className="font-semibold text-sm">Debug Console</h3>
          <Badge variant="secondary" className="text-xs">
            {filteredMessages.length}
          </Badge>
        </div>

        <div className="flex items-center gap-2">
          {/* Search */}
          <div className="relative w-48">
            <Search className="absolute left-2 top-1/2 transform -translate-y-1/2 h-3 w-3 text-muted-foreground" />
            <Input
              type="search"
              placeholder="Search..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="h-8 pl-7 text-xs"
            />
          </div>

          {/* Filter */}
          <Select value={filterLevel} onValueChange={setFilterLevel}>
            <SelectTrigger className="w-28 h-8 text-xs">
              <Filter className="h-3 w-3 mr-1" />
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="all">All</SelectItem>
              <SelectItem value="info">Info</SelectItem>
              <SelectItem value="debug">Debug</SelectItem>
              <SelectItem value="warn">Warning</SelectItem>
              <SelectItem value="error">Error</SelectItem>
            </SelectContent>
          </Select>

          {/* Clear */}
          <Button variant="ghost" size="sm" onClick={handleClear} className="h-8">
            <Trash2 className="h-3 w-3" />
          </Button>
        </div>
      </div>

      {/* Messages */}
      <ScrollArea className="flex-1">
        {filteredMessages.length === 0 ? (
          <div className="p-8 text-center text-muted-foreground">
            <p className="text-sm">No messages</p>
            <p className="text-xs mt-1">Debug output will appear here</p>
          </div>
        ) : (
          <div className="divide-y divide-border">
            {filteredMessages.map((message) => {
              const isExpanded = expandedMessages.has(message.id)

              return (
                <div
                  key={message.id}
                  className="p-3 hover:bg-accent/50 transition-colors"
                >
                  <div className="flex items-start gap-2">
                    {/* Expand Button */}
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-5 w-5 p-0"
                      onClick={() => toggleExpand(message.id)}
                    >
                      {isExpanded ? (
                        <ChevronDown className="h-3 w-3" />
                      ) : (
                        <ChevronRight className="h-3 w-3" />
                      )}
                    </Button>

                    {/* Content */}
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2 mb-1">
                        <Badge variant={getLevelColor(message.level)} className="text-[10px]">
                          {message.level}
                        </Badge>
                        <span className="text-xs text-muted-foreground">
                          {new Date(message.timestamp).toLocaleTimeString()}
                        </span>
                        <span className="text-xs font-medium">{message.nodeName}</span>
                        {message.topic && (
                          <span className="text-xs text-muted-foreground">
                            [{message.topic}]
                          </span>
                        )}
                      </div>

                      {/* Collapsed View */}
                      {!isExpanded && (
                        <div className="text-xs font-mono text-muted-foreground truncate">
                          {typeof message.payload === 'object'
                            ? JSON.stringify(message.payload)
                            : String(message.payload)}
                        </div>
                      )}

                      {/* Expanded View */}
                      {isExpanded && (
                        <div className="mt-2">
                          <pre className="text-xs font-mono bg-muted p-2 rounded overflow-x-auto">
                            {JSON.stringify(message.payload, null, 2)}
                          </pre>
                        </div>
                      )}
                    </div>

                    {/* Copy Button */}
                    <Button
                      variant="ghost"
                      size="sm"
                      className="h-6 w-6 p-0"
                      onClick={() => handleCopy(message)}
                    >
                      <Copy className="h-3 w-3" />
                    </Button>
                  </div>
                </div>
              )
            })}
          </div>
        )}
      </ScrollArea>
    </div>
  )
}
