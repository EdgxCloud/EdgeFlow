/**
 * useWebSocket Hook
 * React hook for WebSocket connection and message handling
 */

import { useEffect, useCallback, useRef, useState } from 'react'
import {
  wsClient,
  WSMessage,
  WSMessageType,
  FlowStatusMessage,
  NodeStatusMessage,
  ExecutionMessage,
  LogMessage,
  NotificationMessage,
} from '@/services/websocket'

export interface UseWebSocketOptions {
  autoConnect?: boolean
  debug?: boolean
}

export interface UseWebSocketReturn {
  isConnected: boolean
  connect: () => void
  disconnect: () => void
  send: (message: WSMessage) => void
  subscribe: (type: WSMessageType | 'all', handler: (message: WSMessage) => void) => () => void
}

/**
 * Hook for WebSocket connection management
 */
export function useWebSocket(options: UseWebSocketOptions = {}): UseWebSocketReturn {
  const { autoConnect = true, debug = false } = options
  const [isConnected, setIsConnected] = useState(false)
  const unsubscribersRef = useRef<Array<() => void>>([])

  // Connect/disconnect functions
  const connect = useCallback(() => {
    if (debug) console.log('[useWebSocket] Connecting...')
    wsClient.connect()
  }, [debug])

  const disconnect = useCallback(() => {
    if (debug) console.log('[useWebSocket] Disconnecting...')
    wsClient.disconnect()
  }, [debug])

  const send = useCallback((message: WSMessage) => {
    wsClient.send(message)
  }, [])

  const subscribe = useCallback((
    type: WSMessageType | 'all',
    handler: (message: WSMessage) => void
  ): (() => void) => {
    return wsClient.on(type, handler)
  }, [])

  // Setup WebSocket event handlers
  useEffect(() => {
    const handleOpen = () => {
      if (debug) console.log('[useWebSocket] Connected')
      setIsConnected(true)
    }

    const handleClose = () => {
      if (debug) console.log('[useWebSocket] Disconnected')
      setIsConnected(false)
    }

    const handleError = (error: Event) => {
      if (debug) console.error('[useWebSocket] Error:', error)
    }

    const unsubOpen = wsClient.onOpen(handleOpen)
    const unsubClose = wsClient.onClose(handleClose)
    const unsubError = wsClient.onError(handleError)

    // Auto-connect if enabled
    if (autoConnect) {
      connect()
    }

    return () => {
      unsubOpen()
      unsubClose()
      unsubError()
      // Clean up all message subscriptions
      unsubscribersRef.current.forEach(unsub => unsub())
      if (autoConnect) {
        disconnect()
      }
    }
  }, [autoConnect, connect, disconnect, debug])

  return {
    isConnected,
    connect,
    disconnect,
    send,
    subscribe,
  }
}

/**
 * Hook for flow status updates
 */
export function useFlowStatus(
  flowId: string,
  onStatusChange?: (message: FlowStatusMessage) => void
) {
  const { subscribe } = useWebSocket({ autoConnect: true })

  useEffect(() => {
    const unsubscribe = subscribe('flow_status', (message) => {
      const data = message.data as FlowStatusMessage
      if (data.flow_id === flowId) {
        onStatusChange?.(data)
      }
    })

    return unsubscribe
  }, [flowId, onStatusChange, subscribe])
}

/**
 * Hook for node status updates
 */
export function useNodeStatus(
  flowId: string,
  onStatusChange?: (message: NodeStatusMessage) => void
) {
  const { subscribe } = useWebSocket({ autoConnect: true })

  useEffect(() => {
    const unsubscribe = subscribe('node_status', (message) => {
      const data = message.data as NodeStatusMessage
      if (data.flow_id === flowId) {
        onStatusChange?.(data)
      }
    })

    return unsubscribe
  }, [flowId, onStatusChange, subscribe])
}

/**
 * Hook for execution messages
 */
export function useExecution(
  flowId: string,
  onExecution?: (message: ExecutionMessage) => void
) {
  const { subscribe } = useWebSocket({ autoConnect: true })

  useEffect(() => {
    const unsubscribe = subscribe('execution', (message) => {
      const data = message.data as ExecutionMessage
      if (data.flow_id === flowId) {
        onExecution?.(data)
      }
    })

    return unsubscribe
  }, [flowId, onExecution, subscribe])
}

/**
 * Hook for log messages
 */
export function useLogs(onLog?: (message: LogMessage) => void) {
  const { subscribe } = useWebSocket({ autoConnect: true })

  useEffect(() => {
    const unsubscribe = subscribe('log', (message) => {
      const data = message.data as LogMessage
      onLog?.(data)
    })

    return unsubscribe
  }, [onLog, subscribe])
}

/**
 * Hook for notifications
 */
export function useNotifications(onNotification?: (message: NotificationMessage) => void) {
  const { subscribe } = useWebSocket({ autoConnect: true })

  useEffect(() => {
    const unsubscribe = subscribe('notification', (message) => {
      const data = message.data as NotificationMessage
      onNotification?.(data)
    })

    return unsubscribe
  }, [onNotification, subscribe])
}

export default useWebSocket
