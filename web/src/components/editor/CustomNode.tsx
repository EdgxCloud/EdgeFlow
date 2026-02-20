import { memo } from 'react'
import { Handle, Position, NodeProps, type Node } from '@xyflow/react'
import {
  Zap,
  Bug,
  Code,
  GitBranch,
  Clock,
  Play,
  Scissors,
  Link2,
  Settings,
  AlertTriangle,
  CheckCircle,
  Activity,
  FileText,
  Terminal,
  Database,
  Network,
  MessageSquare,
  Bot,
  Thermometer,
  Gauge,
  Download,
  Upload
} from 'lucide-react'

interface CustomNodeData extends Record<string, unknown> {
  label: string
  nodeType: string
  config?: any
  status?: 'idle' | 'running' | 'error' | 'success'
}

type CustomNodeType = Node<CustomNodeData>

// Node configuration with icons and colors matching categories
const NODE_CONFIG: Record<string, { icon: any; color: string; category: string }> = {
  // Input nodes
  inject: { icon: Zap, color: '#10b981', category: 'input' },
  'mqtt-in': { icon: Download, color: '#10b981', category: 'input' },

  // Output nodes
  debug: { icon: Bug, color: '#ef4444', category: 'output' },
  'mqtt-out': { icon: Upload, color: '#ef4444', category: 'output' },

  // Function nodes
  function: { icon: Code, color: '#f59e0b', category: 'function' },
  change: { icon: Settings, color: '#f59e0b', category: 'function' },
  range: { icon: Gauge, color: '#f59e0b', category: 'function' },
  template: { icon: FileText, color: '#06b6d4', category: 'function' },

  // Logic nodes
  if: { icon: GitBranch, color: '#8b5cf6', category: 'logic' },
  switch: { icon: GitBranch, color: '#8b5cf6', category: 'logic' },
  delay: { icon: Clock, color: '#8b5cf6', category: 'logic' },

  // Advanced nodes
  exec: { icon: Terminal, color: '#64748b', category: 'advanced' },
  split: { icon: Scissors, color: '#14b8a6', category: 'advanced' },
  join: { icon: Link2, color: '#06b6d4', category: 'advanced' },
  catch: { icon: AlertTriangle, color: '#dc2626', category: 'advanced' },
  complete: { icon: CheckCircle, color: '#3b82f6', category: 'advanced' },
  status: { icon: Activity, color: '#22c55e', category: 'advanced' },
  set: { icon: Settings, color: '#f59e0b', category: 'advanced' },

  // Sensors
  dht: { icon: Thermometer, color: '#22c55e', category: 'sensors' },
  ds18b20: { icon: Thermometer, color: '#22c55e', category: 'sensors' },
  bmp280: { icon: Thermometer, color: '#22c55e', category: 'sensors' },

  // Network
  'http-request': { icon: Network, color: '#06b6d4', category: 'network' },
  websocket: { icon: Network, color: '#06b6d4', category: 'network' },
  tcp: { icon: Network, color: '#06b6d4', category: 'network' },
  udp: { icon: Network, color: '#06b6d4', category: 'network' },

  // Database
  mysql: { icon: Database, color: '#3b82f6', category: 'database' },
  postgresql: { icon: Database, color: '#3b82f6', category: 'database' },
  mongodb: { icon: Database, color: '#3b82f6', category: 'database' },
  redis: { icon: Database, color: '#3b82f6', category: 'database' },

  // Messaging
  email: { icon: MessageSquare, color: '#14b8a6', category: 'messaging' },
  telegram: { icon: MessageSquare, color: '#14b8a6', category: 'messaging' },
  slack: { icon: MessageSquare, color: '#14b8a6', category: 'messaging' },
  discord: { icon: MessageSquare, color: '#14b8a6', category: 'messaging' },

  // AI
  openai: { icon: Bot, color: '#a855f7', category: 'ai' },
  anthropic: { icon: Bot, color: '#a855f7', category: 'ai' },
  ollama: { icon: Bot, color: '#a855f7', category: 'ai' },
}

function CustomNode({ data, selected }: NodeProps<CustomNodeType>) {
  const config = NODE_CONFIG[data.nodeType] || {
    icon: Activity,
    color: '#64748b',
    category: 'unknown'
  }
  const Icon = config.icon

  const getStatusIndicator = () => {
    switch (data.status) {
      case 'running':
        return (
          <div className="absolute -top-1 -right-1 w-3 h-3">
            <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
            <span className="relative inline-flex rounded-full h-3 w-3 bg-green-500"></span>
          </div>
        )
      case 'error':
        return (
          <div className="absolute -top-1 -right-1 w-3 h-3 bg-red-500 rounded-full border-2 border-white dark:border-gray-800"></div>
        )
      case 'success':
        return (
          <div className="absolute -top-1 -right-1 w-3 h-3 bg-green-500 rounded-full border-2 border-white dark:border-gray-800"></div>
        )
      default:
        return null
    }
  }

  return (
    <div className="relative group">
      {/* Status indicator */}
      {getStatusIndicator()}

      {/* Main node container - n8n style */}
      <div
        className={`
          min-w-[180px] rounded-xl overflow-hidden
          bg-white dark:bg-gray-800
          border-2 transition-all duration-200
          ${
            selected
              ? 'border-blue-500 shadow-xl shadow-blue-500/20 scale-105'
              : 'border-gray-200 dark:border-gray-700 hover:border-gray-300 dark:hover:border-gray-600 shadow-lg hover:shadow-xl'
          }
          cursor-pointer
        `}
        style={{
          boxShadow: selected
            ? `0 20px 25px -5px ${config.color}20, 0 10px 10px -5px ${config.color}10`
            : undefined
        }}
        title="Double-click to configure"
      >
        {/* Icon circle - like n8n */}
        <div className="relative h-16 flex items-center justify-center bg-gradient-to-br from-gray-50 to-gray-100 dark:from-gray-900 dark:to-gray-800">
          <div
            className="w-12 h-12 rounded-full flex items-center justify-center shadow-lg transition-transform group-hover:scale-110"
            style={{
              backgroundColor: config.color,
              boxShadow: `0 4px 14px ${config.color}40`
            }}
          >
            <Icon className="w-6 h-6 text-white" strokeWidth={2.5} />
          </div>
        </div>

        {/* Node info */}
        <div className="px-4 py-3 border-t border-gray-100 dark:border-gray-700">
          <div className="text-sm font-semibold text-gray-900 dark:text-white text-center truncate mb-1">
            {data.label}
          </div>
          <div className="text-xs text-gray-500 dark:text-gray-400 text-center truncate">
            {data.nodeType}
          </div>
        </div>

        {/* Hover indicator */}
        <div
          className="absolute bottom-0 left-0 right-0 h-0.5 bg-gradient-to-r from-transparent via-blue-500 to-transparent opacity-0 group-hover:opacity-100 transition-opacity"
        ></div>
      </div>

      {/* Connection handles - styled like n8n */}
      <Handle
        type="target"
        position={Position.Left}
        className="!w-4 !h-4 !border-2 !border-white dark:!border-gray-800 transition-all hover:!scale-125"
        style={{
          background: config.color,
          left: -8
        }}
      />
      <Handle
        type="source"
        position={Position.Right}
        className="!w-4 !h-4 !border-2 !border-white dark:!border-gray-800 transition-all hover:!scale-125"
        style={{
          background: config.color,
          right: -8
        }}
      />

      {/* Subtle category badge on hover */}
      <div className="absolute -bottom-6 left-1/2 transform -translate-x-1/2 opacity-0 group-hover:opacity-100 transition-opacity pointer-events-none">
        <div className="px-2 py-0.5 rounded-full bg-gray-900 dark:bg-gray-100 text-white dark:text-gray-900 text-xs font-medium whitespace-nowrap">
          {config.category}
        </div>
      </div>
    </div>
  )
}

export default memo(CustomNode)
