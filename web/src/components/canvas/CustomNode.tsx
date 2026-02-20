import { memo } from 'react'
import { Handle, Position, NodeProps, type Node } from '@xyflow/react'
import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'
import {
  Activity,
  AlertCircle,
  Bug,
  Code,
  Database,
  GitBranch,
  HardDrive,
  Layers,
  Mail,
  MessageSquare,
  Network,
  Play,
  Power,
  Send,
  Server,
  Terminal,
  Wifi,
  Zap,
} from 'lucide-react'

// Top-level logging to verify file is loaded
console.log('🚀 CustomNode.tsx FILE LOADED')

// Icon mapping for different node types
const getNodeIcon = (nodeType: string) => {
  const icons: Record<string, any> = {
    inject: Play,
    debug: Bug,
    function: Code,
    if: GitBranch,
    delay: Activity,
    exec: Terminal,
    python: Code,
    'gpio-in': Power,
    'gpio-out': Power,
    pwm: Zap,
    servo: Zap,
    i2c: HardDrive,
    spi: HardDrive,
    serial: Server,
    'http-request': Send,
    'http-in': Network,
    'http-response': Network,
    'mqtt-in': MessageSquare,
    'mqtt-out': MessageSquare,
    websocket: Wifi,
    telegram: Send,
    email: Mail,
    mysql: Database,
    postgresql: Database,
    mongodb: Database,
    redis: Layers,
  }
  return icons[nodeType] || Activity
}

// Category colors
const categoryColors: Record<string, string> = {
  core: 'border-blue-500 bg-blue-500/10',
  input: 'border-blue-500 bg-blue-500/10',
  output: 'border-green-500 bg-green-500/10',
  function: 'border-purple-500 bg-purple-500/10',
  hardware: 'border-orange-500 bg-orange-500/10',
  network: 'border-cyan-500 bg-cyan-500/10',
  storage: 'border-indigo-500 bg-indigo-500/10',
}

const categoryTextColors: Record<string, string> = {
  core: 'text-blue-600 dark:text-blue-400',
  input: 'text-blue-600 dark:text-blue-400',
  output: 'text-green-600 dark:text-green-400',
  function: 'text-purple-600 dark:text-purple-400',
  hardware: 'text-orange-600 dark:text-orange-400',
  network: 'text-cyan-600 dark:text-cyan-400',
  storage: 'text-indigo-600 dark:text-indigo-400',
}

interface CustomNodeData extends Record<string, unknown> {
  category?: string
  nodeType?: string
  label?: string
  subtitle?: string
  status?: string
  error?: string
  info?: string
  hasInputs?: boolean
  hasOutputs?: boolean
}

type CustomNodeType = Node<CustomNodeData>

function CustomNode({ data, selected }: NodeProps<CustomNodeType>) {
  const category = (data?.category as string) || 'core'
  const nodeType = (data?.nodeType as string) || 'unknown'
  const Icon = getNodeIcon(nodeType)
  const hasInputs = data?.hasInputs !== false
  const hasOutputs = data?.hasOutputs !== false

  const handleDoubleClick = (e: React.MouseEvent) => {
    console.log('✅ CustomNode onDoubleClick triggered for:', data?.label)
    console.log('Event type:', e.type, 'Button:', e.button, 'Detail:', e.detail)
    // Don't stop propagation - let it bubble to ReactFlow
  }

  const handleClick = (e: React.MouseEvent) => {
    console.log('👆 CustomNode onClick triggered for:', data?.label, 'Detail:', e.detail)
  }

  console.log('CustomNode rendering for:', data?.label)

  return (
    <div
      className="relative pointer-events-auto"
      onDoubleClick={handleDoubleClick}
      onClick={handleClick}
      style={{ pointerEvents: 'auto' }}
    >
      {/* Input Handle */}
      {hasInputs && (
        <Handle
          type="target"
          position={Position.Left}
          className={cn(
            'w-3 h-3 border-2 border-background',
            categoryColors[category]?.replace('bg-', 'bg-opacity-100 bg-')
          )}
        />
      )}

      {/* Node Card */}
      <Card
        className={cn(
          'min-w-[180px] max-w-[250px] transition-all shadow-md hover:shadow-lg cursor-pointer',
          categoryColors[category],
          selected && 'ring-2 ring-primary ring-offset-2 ring-offset-background'
        )}
        title="Double-click to configure"
        onDoubleClick={handleDoubleClick}
        onClick={handleClick}
        style={{ pointerEvents: 'auto' }}
      >
        <CardContent className="p-3">
          <div className="flex items-start gap-2 relative">
            {/* DEBUG: Test button */}
            <button
              onClick={(e) => {
                e.stopPropagation()
                alert('TEST: Button clicked! Node: ' + data?.label)
                console.log('🔴 TEST BUTTON CLICKED')
              }}
              style={{
                position: 'absolute',
                top: 0,
                right: 0,
                zIndex: 1000,
                background: 'red',
                color: 'white',
                padding: '4px 8px',
                fontSize: '10px',
                border: 'none',
                cursor: 'pointer',
              }}
            >
              TEST
            </button>
            {/* Icon */}
            <div
              className={cn(
                'flex-shrink-0 w-8 h-8 rounded-md flex items-center justify-center',
                categoryColors[category]
              )}
            >
              <Icon className={cn('h-4 w-4', categoryTextColors[category])} />
            </div>

            {/* Content */}
            <div className="flex-1 min-w-0">
              <div className="flex items-start justify-between gap-2">
                <div className="flex-1 min-w-0">
                  <h4 className="text-sm font-semibold truncate">
                    {data?.label || 'Node'}
                  </h4>
                  {data?.subtitle && (
                    <p className="text-xs text-muted-foreground truncate mt-0.5">
                      {data.subtitle}
                    </p>
                  )}
                </div>

                {/* Status Badge */}
                {data?.status && (
                  <Badge
                    variant={
                      data.status === 'running'
                        ? 'success'
                        : data.status === 'error'
                        ? 'destructive'
                        : 'secondary'
                    }
                    className="text-[10px] px-1.5 py-0"
                  >
                    {data.status}
                  </Badge>
                )}
              </div>

              {/* Error Message */}
              {data?.error && (
                <div className="flex items-center gap-1 mt-1 text-destructive">
                  <AlertCircle className="h-3 w-3" />
                  <span className="text-xs truncate">{data.error}</span>
                </div>
              )}

              {/* Quick Info */}
              {data?.info && (
                <p className="text-xs text-muted-foreground mt-1 line-clamp-2">
                  {data.info}
                </p>
              )}
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Output Handle */}
      {hasOutputs && (
        <Handle
          type="source"
          position={Position.Right}
          className={cn(
            'w-3 h-3 border-2 border-background',
            categoryColors[category]?.replace('bg-', 'bg-opacity-100 bg-')
          )}
        />
      )}
    </div>
  )
}

// Temporarily remove memo to test if it's blocking double-click events
export default CustomNode
