import React, { useState, useEffect, useCallback } from 'react'
import {
  PlayCircle,
  CheckCircle,
  XCircle,
  Clock,
  RefreshCw,
  ChevronDown,
  ChevronRight,
  AlertTriangle,
} from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import { executionsApi } from '@/lib/api'
import type { ExecutionRecord } from '@/lib/api'
import wsClient from '@/services/websocket'
import type { WSMessage } from '@/services/websocket'

export default function ExecutionsFull() {
  const [executions, setExecutions] = useState<ExecutionRecord[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [filter, setFilter] = useState<string>('all')
  const [expandedId, setExpandedId] = useState<string | null>(null)

  // Fetch executions from API
  const fetchExecutions = useCallback(async () => {
    try {
      setIsLoading(true)
      setError(null)
      const statusParam = filter !== 'all' ? filter : undefined
      const response = await executionsApi.list(statusParam)
      setExecutions(response.data.executions || [])
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch executions')
      setExecutions([])
    } finally {
      setIsLoading(false)
    }
  }, [filter])

  // Initial load + re-fetch on filter change
  useEffect(() => {
    fetchExecutions()
  }, [fetchExecutions])

  // Subscribe to WebSocket for real-time updates on flow status
  useEffect(() => {
    if (!wsClient.isConnected()) {
      wsClient.connect()
    }

    const handleFlowStatus = (msg: WSMessage) => {
      const data = msg.data as Record<string, unknown>
      const action = data.action as string
      // Re-fetch when flows start/stop/complete
      if (action === 'started' || action === 'stopped') {
        // Small delay to allow backend to update the record
        setTimeout(() => fetchExecutions(), 300)
      }
    }

    const unsub = wsClient.on('flow_status', handleFlowStatus)
    return () => { unsub() }
  }, [fetchExecutions])

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'running':
        return <PlayCircle className="w-4 h-4 text-blue-500 animate-pulse" />
      case 'completed':
        return <CheckCircle className="w-4 h-4 text-green-500" />
      case 'failed':
        return <XCircle className="w-4 h-4 text-red-500" />
      default:
        return <Clock className="w-4 h-4 text-gray-500" />
    }
  }

  const getStatusBadgeClass = (status: string) => {
    switch (status) {
      case 'running':
        return 'bg-blue-500/10 text-blue-500 border-blue-500/20'
      case 'completed':
        return 'bg-green-500/10 text-green-500 border-green-500/20'
      case 'failed':
        return 'bg-red-500/10 text-red-500 border-red-500/20'
      default:
        return 'bg-gray-500/10 text-gray-500 border-gray-500/20'
    }
  }

  const getProgressColor = (status: string) => {
    switch (status) {
      case 'completed': return 'bg-green-500'
      case 'failed': return 'bg-red-500'
      case 'running': return 'bg-blue-500'
      default: return 'bg-gray-500'
    }
  }

  const formatTime = (ts: string) => {
    try {
      return new Date(ts).toLocaleString('en-US', { hour12: false })
    } catch {
      return ts
    }
  }

  const formatDuration = (ms?: number) => {
    if (!ms) return '-'
    if (ms < 1000) return `${ms}ms`
    const seconds = Math.floor(ms / 1000)
    if (seconds < 60) return `${seconds}s`
    const minutes = Math.floor(seconds / 60)
    const remainingSeconds = seconds % 60
    return `${minutes}m ${remainingSeconds}s`
  }

  const filteredExecutions = filter === 'all'
    ? executions
    : executions.filter(e => e.status === filter)

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900 dark:text-white">
            Executions
          </h1>
          <p className="text-gray-600 dark:text-gray-400 mt-1">
            Flow execution history and status
          </p>
        </div>
        <Button
          variant="outline"
          size="sm"
          onClick={fetchExecutions}
          disabled={isLoading}
          className="gap-2"
        >
          <RefreshCw className={cn('w-4 h-4', isLoading && 'animate-spin')} />
          Refresh
        </Button>
      </div>

      {/* Filters */}
      <div className="flex items-center space-x-2">
        {['all', 'running', 'completed', 'failed'].map((status) => (
          <button
            key={status}
            onClick={() => setFilter(status)}
            className={cn(
              'px-4 py-2 rounded-lg text-sm font-medium transition-colors',
              filter === status
                ? 'bg-blue-600 text-white'
                : 'bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-600'
            )}
          >
            {status === 'all' ? 'All' : status.charAt(0).toUpperCase() + status.slice(1)}
          </button>
        ))}

        {executions.length > 0 && (
          <span className="ml-auto text-sm text-muted-foreground">
            {filteredExecutions.length} execution{filteredExecutions.length !== 1 ? 's' : ''}
          </span>
        )}
      </div>

      {/* Error state */}
      {error && (
        <div className="flex items-center gap-3 p-4 bg-red-500/10 border border-red-500/20 rounded-lg text-red-500">
          <AlertTriangle className="w-5 h-5 shrink-0" />
          <span className="text-sm">{error}</span>
          <Button variant="ghost" size="sm" onClick={fetchExecutions} className="ml-auto">
            Retry
          </Button>
        </div>
      )}

      {/* Executions table */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
        <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
          <thead className="bg-gray-50 dark:bg-gray-900">
            <tr>
              <th className="w-8 px-4 py-3" />
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                Flow Name
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                Status
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                Start Time
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                Duration
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                Nodes
              </th>
            </tr>
          </thead>
          <tbody className="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
            {isLoading && filteredExecutions.length === 0 ? (
              <tr>
                <td colSpan={6} className="px-6 py-12 text-center">
                  <RefreshCw className="w-6 h-6 text-muted-foreground animate-spin mx-auto mb-2" />
                  <p className="text-sm text-muted-foreground">Loading executions...</p>
                </td>
              </tr>
            ) : filteredExecutions.length === 0 ? (
              <tr>
                <td colSpan={6} className="px-6 py-12 text-center">
                  <Clock className="w-10 h-10 text-gray-400 mx-auto mb-3" />
                  <p className="text-sm text-gray-600 dark:text-gray-400 font-medium">
                    No executions found
                  </p>
                  <p className="text-xs text-muted-foreground mt-1">
                    Start a workflow to see execution history here
                  </p>
                </td>
              </tr>
            ) : (
              filteredExecutions.map((execution) => {
                const isExpanded = expandedId === execution.id
                const progress = execution.node_count > 0
                  ? (execution.completed_nodes / execution.node_count) * 100
                  : 0

                return (
                  <React.Fragment key={execution.id}>
                    <tr
                      className={cn(
                        'hover:bg-gray-50 dark:hover:bg-gray-700/50 cursor-pointer transition-colors',
                        isExpanded && 'bg-gray-50 dark:bg-gray-700/30'
                      )}
                      onClick={() => setExpandedId(isExpanded ? null : execution.id)}
                    >
                      {/* Expand */}
                      <td className="px-4 py-4">
                        {isExpanded
                          ? <ChevronDown className="w-4 h-4 text-muted-foreground" />
                          : <ChevronRight className="w-4 h-4 text-muted-foreground" />
                        }
                      </td>

                      {/* Flow name */}
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="flex items-center gap-2">
                          {getStatusIcon(execution.status)}
                          <span className="text-sm font-medium text-gray-900 dark:text-white">
                            {execution.flow_name}
                          </span>
                        </div>
                      </td>

                      {/* Status */}
                      <td className="px-6 py-4 whitespace-nowrap">
                        <span className={cn(
                          'inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium border',
                          getStatusBadgeClass(execution.status)
                        )}>
                          {execution.status.charAt(0).toUpperCase() + execution.status.slice(1)}
                        </span>
                      </td>

                      {/* Start Time */}
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-600 dark:text-gray-400">
                        {formatTime(execution.start_time)}
                      </td>

                      {/* Duration */}
                      <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-600 dark:text-gray-400">
                        {execution.status === 'running'
                          ? <span className="text-blue-500 animate-pulse">Running...</span>
                          : formatDuration(execution.duration)
                        }
                      </td>

                      {/* Nodes progress */}
                      <td className="px-6 py-4 whitespace-nowrap">
                        <div className="flex items-center gap-3">
                          <div className="w-24 bg-gray-200 dark:bg-gray-700 rounded-full h-2">
                            <div
                              className={cn('h-2 rounded-full transition-all', getProgressColor(execution.status))}
                              style={{ width: `${progress}%` }}
                            />
                          </div>
                          <span className="text-xs text-gray-600 dark:text-gray-400 whitespace-nowrap">
                            {execution.completed_nodes}/{execution.node_count}
                            {execution.error_nodes > 0 && (
                              <span className="text-red-500 ml-1">({execution.error_nodes} err)</span>
                            )}
                          </span>
                        </div>
                      </td>
                    </tr>

                    {/* Expanded details */}
                    {isExpanded && (
                      <tr>
                        <td colSpan={6} className="px-6 py-4 bg-gray-50/50 dark:bg-gray-900/30">
                          <div className="space-y-3">
                            {/* Meta info */}
                            <div className="grid grid-cols-4 gap-4 text-xs">
                              <div>
                                <span className="text-muted-foreground">Execution ID</span>
                                <p className="font-mono mt-0.5 text-gray-900 dark:text-gray-200 truncate">{execution.id}</p>
                              </div>
                              <div>
                                <span className="text-muted-foreground">Flow ID</span>
                                <p className="font-mono mt-0.5 text-gray-900 dark:text-gray-200 truncate">{execution.flow_id}</p>
                              </div>
                              <div>
                                <span className="text-muted-foreground">End Time</span>
                                <p className="mt-0.5 text-gray-900 dark:text-gray-200">
                                  {execution.end_time ? formatTime(execution.end_time) : '-'}
                                </p>
                              </div>
                              <div>
                                <span className="text-muted-foreground">Total Duration</span>
                                <p className="mt-0.5 text-gray-900 dark:text-gray-200">
                                  {formatDuration(execution.duration)}
                                </p>
                              </div>
                            </div>

                            {/* Error message */}
                            {execution.error && (
                              <div className="flex items-start gap-2 p-3 bg-red-500/10 border border-red-500/20 rounded-lg">
                                <XCircle className="w-4 h-4 text-red-500 shrink-0 mt-0.5" />
                                <p className="text-xs text-red-400 font-mono">{execution.error}</p>
                              </div>
                            )}

                            {/* Node events */}
                            {execution.node_events && execution.node_events.length > 0 && (
                              <div>
                                <h4 className="text-xs font-semibold uppercase text-muted-foreground mb-2">
                                  Node Executions ({execution.node_events.length})
                                </h4>
                                <div className="border border-border rounded-lg overflow-hidden">
                                  <table className="w-full text-xs">
                                    <thead>
                                      <tr className="bg-muted/50">
                                        <th className="px-3 py-1.5 text-left font-medium text-muted-foreground">Node</th>
                                        <th className="px-3 py-1.5 text-left font-medium text-muted-foreground">Type</th>
                                        <th className="px-3 py-1.5 text-left font-medium text-muted-foreground">Status</th>
                                        <th className="px-3 py-1.5 text-left font-medium text-muted-foreground">Time</th>
                                        <th className="px-3 py-1.5 text-left font-medium text-muted-foreground">Error</th>
                                      </tr>
                                    </thead>
                                    <tbody>
                                      {execution.node_events.map((event, idx) => (
                                        <tr key={idx} className="border-t border-border hover:bg-muted/30">
                                          <td className="px-3 py-1.5 font-medium">{event.node_name}</td>
                                          <td className="px-3 py-1.5">
                                            <Badge variant="secondary" className="text-[9px] h-4">{event.node_type}</Badge>
                                          </td>
                                          <td className="px-3 py-1.5">
                                            {event.status === 'success'
                                              ? <CheckCircle className="w-3.5 h-3.5 text-green-500 inline" />
                                              : <XCircle className="w-3.5 h-3.5 text-red-500 inline" />
                                            }
                                          </td>
                                          <td className="px-3 py-1.5 text-muted-foreground">{event.execution_time}ms</td>
                                          <td className="px-3 py-1.5 text-red-400 truncate max-w-[200px]">{event.error || '-'}</td>
                                        </tr>
                                      ))}
                                    </tbody>
                                  </table>
                                </div>
                              </div>
                            )}

                            {(!execution.node_events || execution.node_events.length === 0) && (
                              <p className="text-xs text-muted-foreground italic">
                                No node execution events recorded
                              </p>
                            )}
                          </div>
                        </td>
                      </tr>
                    )}
                  </React.Fragment>
                )
              })
            )}
          </tbody>
        </table>
      </div>
    </div>
  )
}
