/**
 * WebSocket Service
 * Handles real-time updates from backend
 */

import { WS_URL } from './api'

// WebSocket Message Types
export type WSMessageType =
  | 'flow_status'
  | 'node_status'
  | 'module_status'
  | 'execution'
  | 'log'
  | 'notification'
  | 'gpio_state'

export interface WSMessage {
  type: WSMessageType
  data: unknown
  timestamp?: string
}

export interface FlowStatusMessage {
  flow_id: string
  status: 'running' | 'stopped' | 'error'
  message?: string
}

export interface NodeStatusMessage {
  flow_id: string
  node_id: string
  status: 'ok' | 'error' | 'warning'
  message?: string
  fill?: string
  shape?: string
  text?: string
}

export interface ExecutionMessage {
  flow_id: string
  node_id: string
  node_name: string
  node_type: string
  input: Record<string, unknown> | null
  output: Record<string, unknown> | null
  status: 'success' | 'error'
  error?: string
  execution_time: number
  timestamp: number
}

export interface LogMessage {
  level: 'debug' | 'info' | 'warn' | 'error'
  message: string
  timestamp: string
  source?: string
}

export interface NotificationMessage {
  title: string
  message: string
  level: 'info' | 'success' | 'warning' | 'error'
  timestamp: string
}

// WebSocket Client Options
export interface WSClientOptions {
  reconnect?: boolean
  reconnectInterval?: number
  maxReconnectAttempts?: number
  debug?: boolean
}

// WebSocket Event Handlers
export type WSEventHandler = (message: WSMessage) => void
export type WSErrorHandler = (error: Event) => void
export type WSCloseHandler = (event: CloseEvent) => void
export type WSOpenHandler = () => void

/**
 * WebSocket Client Class
 */
export class WebSocketClient {
  private ws: WebSocket | null = null
  private url: string
  private options: Required<WSClientOptions>
  private reconnectAttempts = 0
  private reconnectTimer: NodeJS.Timeout | null = null
  private handlers: Map<WSMessageType | 'all', Set<WSEventHandler>> = new Map()
  private errorHandlers: Set<WSErrorHandler> = new Set()
  private closeHandlers: Set<WSCloseHandler> = new Set()
  private openHandlers: Set<WSOpenHandler> = new Set()
  private isManualClose = false

  constructor(url: string = WS_URL, options: WSClientOptions = {}) {
    this.url = url
    this.options = {
      reconnect: options.reconnect ?? true,
      reconnectInterval: options.reconnectInterval ?? 3000,
      maxReconnectAttempts: options.maxReconnectAttempts ?? 10,
      debug: options.debug ?? false,
    }
  }

  /**
   * Connect to WebSocket server
   */
  connect(): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.log('Already connected')
      return
    }

    this.isManualClose = false
    this.log('Connecting to', this.url)

    try {
      this.ws = new WebSocket(this.url)

      this.ws.onopen = () => {
        this.log('Connected')
        this.reconnectAttempts = 0
        this.openHandlers.forEach(handler => handler())
      }

      this.ws.onmessage = (event) => {
        try {
          const message = JSON.parse(event.data) as WSMessage
          this.log('Received message:', message.type)

          // Call type-specific handlers
          const typeHandlers = this.handlers.get(message.type)
          typeHandlers?.forEach(handler => handler(message))

          // Call global handlers
          const allHandlers = this.handlers.get('all')
          allHandlers?.forEach(handler => handler(message))
        } catch (error) {
          console.error('Failed to parse WebSocket message:', error)
        }
      }

      this.ws.onerror = (error) => {
        this.log('Error:', error)
        this.errorHandlers.forEach(handler => handler(error))
      }

      this.ws.onclose = (event) => {
        this.log('Disconnected:', event.code, event.reason)
        this.closeHandlers.forEach(handler => handler(event))

        if (!this.isManualClose && this.options.reconnect) {
          this.scheduleReconnect()
        }
      }
    } catch (error) {
      console.error('Failed to create WebSocket:', error)
    }
  }

  /**
   * Disconnect from WebSocket server
   */
  disconnect(): void {
    this.isManualClose = true
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
    this.log('Disconnected manually')
  }

  /**
   * Send message to server
   */
  send(message: WSMessage): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message))
      this.log('Sent message:', message.type)
    } else {
      console.error('WebSocket not connected')
    }
  }

  /**
   * Subscribe to message type
   */
  on(type: WSMessageType | 'all', handler: WSEventHandler): () => void {
    if (!this.handlers.has(type)) {
      this.handlers.set(type, new Set())
    }
    this.handlers.get(type)!.add(handler)

    // Return unsubscribe function
    return () => this.off(type, handler)
  }

  /**
   * Unsubscribe from message type
   */
  off(type: WSMessageType | 'all', handler: WSEventHandler): void {
    const handlers = this.handlers.get(type)
    if (handlers) {
      handlers.delete(handler)
    }
  }

  /**
   * Subscribe to error events
   */
  onError(handler: WSErrorHandler): () => void {
    this.errorHandlers.add(handler)
    return () => this.errorHandlers.delete(handler)
  }

  /**
   * Subscribe to close events
   */
  onClose(handler: WSCloseHandler): () => void {
    this.closeHandlers.add(handler)
    return () => this.closeHandlers.delete(handler)
  }

  /**
   * Subscribe to open events
   */
  onOpen(handler: WSOpenHandler): () => void {
    this.openHandlers.add(handler)
    return () => this.openHandlers.delete(handler)
  }

  /**
   * Get connection status
   */
  isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN
  }

  /**
   * Schedule reconnect attempt
   */
  private scheduleReconnect(): void {
    if (this.reconnectAttempts >= this.options.maxReconnectAttempts) {
      console.error('Max reconnect attempts reached')
      return
    }

    this.reconnectAttempts++
    this.log(`Reconnecting in ${this.options.reconnectInterval}ms (attempt ${this.reconnectAttempts})`)

    this.reconnectTimer = setTimeout(() => {
      this.connect()
    }, this.options.reconnectInterval)
  }

  /**
   * Debug logging
   */
  private log(...args: unknown[]): void {
    if (this.options.debug) {
      console.log('[WebSocket]', ...args)
    }
  }
}

// Create singleton instance
export const wsClient = new WebSocketClient()

export default wsClient
