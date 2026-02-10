/**
 * Logs Panel
 * Real-time log streaming via WebSocket + backend log API
 */

import { useState, useEffect, useRef, useCallback } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Trash2, Search, ArrowDown, Pause, Play } from 'lucide-react'
import { cn } from '@/lib/utils'
import wsClient from '@/services/websocket'
import type { WSMessage } from '@/services/websocket'

interface LogEntry {
  id: number
  timestamp: string
  level: 'info' | 'debug' | 'warn' | 'error' | 'success'
  message: string
  source?: string
}

let logCounter = 0

const LEVEL_COLORS: Record<string, string> = {
  info: 'text-gray-400',
  debug: 'text-blue-400',
  warn: 'text-yellow-400',
  error: 'text-red-400',
  success: 'text-green-400',
}

const LEVEL_BADGES: Record<string, string> = {
  info: 'text-gray-500 bg-gray-500/10',
  debug: 'text-blue-500 bg-blue-500/10',
  warn: 'text-yellow-500 bg-yellow-500/10',
  error: 'text-red-500 bg-red-500/10',
  success: 'text-green-500 bg-green-500/10',
}

export function LogsPanel() {
  const [logs, setLogs] = useState<LogEntry[]>([
    {
      id: logCounter++,
      timestamp: new Date().toISOString(),
      level: 'info',
      message: 'Log panel initialized. Listening for backend logs...',
      source: 'frontend',
    },
  ])
  const [filter, setFilter] = useState('')
  const [isPaused, setIsPaused] = useState(false)
  const [autoScroll, setAutoScroll] = useState(true)
  const scrollRef = useRef<HTMLDivElement>(null)
  const pausedLogsRef = useRef<LogEntry[]>([])

  // Auto-scroll
  useEffect(() => {
    if (autoScroll && scrollRef.current && !isPaused) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight
    }
  }, [logs, autoScroll, isPaused])

  // Subscribe to WebSocket log messages
  useEffect(() => {
    if (!wsClient.isConnected()) {
      wsClient.connect()
    }

    const unsubscribe = wsClient.on('log', (msg: WSMessage) => {
      const data = msg.data as Record<string, unknown>
      const entry: LogEntry = {
        id: logCounter++,
        timestamp: (data.timestamp as string) || new Date().toISOString(),
        level: (data.level as LogEntry['level']) || 'info',
        message: (data.message as string) || JSON.stringify(data),
        source: (data.source as string) || 'backend',
      }

      if (isPaused) {
        pausedLogsRef.current.push(entry)
      } else {
        setLogs(prev => [...prev.slice(-500), entry]) // Keep last 500 entries
      }
    })

    return () => unsubscribe()
  }, [isPaused])

  const handleResume = useCallback(() => {
    setIsPaused(false)
    if (pausedLogsRef.current.length > 0) {
      setLogs(prev => [...prev, ...pausedLogsRef.current].slice(-500))
      pausedLogsRef.current = []
    }
  }, [])

  const clearLogs = () => {
    setLogs([])
    pausedLogsRef.current = []
  }

  const filteredLogs = filter
    ? logs.filter(log =>
        log.message.toLowerCase().includes(filter.toLowerCase()) ||
        log.level.includes(filter.toLowerCase()) ||
        log.source?.toLowerCase().includes(filter.toLowerCase())
      )
    : logs

  const formatTime = (ts: string) => {
    try {
      return new Date(ts).toLocaleTimeString('en-US', { hour12: false })
    } catch {
      return ts
    }
  }

  return (
    <div className="h-full flex flex-col">
      {/* Toolbar */}
      <div className="flex items-center gap-2 px-3 py-1.5 border-b border-gray-200 dark:border-gray-800">
        <div className="relative flex-1 max-w-xs">
          <Search className="absolute left-2 top-1/2 -translate-y-1/2 w-3 h-3 text-muted-foreground" />
          <Input
            value={filter}
            onChange={(e) => setFilter(e.target.value)}
            placeholder="Filter logs..."
            className="h-7 pl-7 text-xs"
          />
        </div>

        <Button
          variant="ghost"
          size="sm"
          className="h-7 text-xs gap-1"
          onClick={() => isPaused ? handleResume() : setIsPaused(true)}
        >
          {isPaused ? (
            <>
              <Play className="w-3 h-3" />
              Resume
              {pausedLogsRef.current.length > 0 && (
                <span className="ml-1 text-yellow-500">({pausedLogsRef.current.length})</span>
              )}
            </>
          ) : (
            <>
              <Pause className="w-3 h-3" />
              Pause
            </>
          )}
        </Button>

        <Button
          variant="ghost"
          size="sm"
          className="h-7 text-xs gap-1"
          onClick={() => setAutoScroll(!autoScroll)}
        >
          <ArrowDown className={cn('w-3 h-3', autoScroll && 'text-blue-500')} />
        </Button>

        <Button
          variant="ghost"
          size="sm"
          className="h-7 text-xs gap-1"
          onClick={clearLogs}
        >
          <Trash2 className="w-3 h-3" />
        </Button>

        <span className="text-xs text-muted-foreground">{filteredLogs.length} entries</span>
      </div>

      {/* Log entries */}
      <div ref={scrollRef} className="flex-1 overflow-auto font-mono text-xs">
        {filteredLogs.length === 0 ? (
          <div className="flex items-center justify-center h-full text-muted-foreground">
            No logs yet. Deploy and run a flow to see logs.
          </div>
        ) : (
          filteredLogs.map((log) => (
            <div
              key={log.id}
              className={cn(
                'px-3 py-0.5 flex items-start gap-2 hover:bg-muted/30',
                log.level === 'error' && 'bg-red-500/5'
              )}
            >
              <span className="text-muted-foreground shrink-0 w-16">
                {formatTime(log.timestamp)}
              </span>
              <span className={cn('shrink-0 w-14 uppercase text-[10px] font-medium px-1 rounded', LEVEL_BADGES[log.level])}>
                {log.level}
              </span>
              {log.source && (
                <span className="shrink-0 text-muted-foreground/60 w-16 truncate">
                  {log.source}
                </span>
              )}
              <span className={cn('flex-1', LEVEL_COLORS[log.level])}>
                {log.message}
              </span>
            </div>
          ))
        )}
      </div>
    </div>
  )
}
