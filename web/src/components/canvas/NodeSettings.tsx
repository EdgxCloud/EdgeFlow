import { useState } from 'react'
import { X, Settings, Info, Code, Zap } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { cn } from '@/lib/utils'

interface NodeSettingsProps {
  node: {
    id: string
    data: {
      label?: string
      nodeType?: string
      config?: Record<string, any>
      [key: string]: any
    }
  } | null
  onClose: () => void
  onSave: (nodeId: string, data: any) => void
}

// Node type configurations
const getNodeIcon = (nodeType: string) => {
  const icons: Record<string, any> = {
    inject: Zap,
    debug: Info,
    function: Code,
    default: Settings,
  }
  return icons[nodeType] || icons.default
}

const getNodeColor = (nodeType: string) => {
  const colors: Record<string, string> = {
    inject: 'from-blue-500 to-blue-600',
    debug: 'from-green-500 to-green-600',
    function: 'from-purple-500 to-purple-600',
    'http-request': 'from-cyan-500 to-cyan-600',
    'mqtt-in': 'from-orange-500 to-orange-600',
    'gpio-out': 'from-red-500 to-red-600',
    default: 'from-gray-500 to-gray-600',
  }
  return colors[nodeType] || colors.default
}

export default function NodeSettings({ node, onClose, onSave }: NodeSettingsProps) {
  const [activeTab, setActiveTab] = useState('general')

  if (!node) return null

  const nodeLabel = node.data?.label || 'Node'
  const nodeType = node.data?.nodeType || 'unknown'
  const NodeIcon = getNodeIcon(nodeType)
  const gradientColor = getNodeColor(nodeType)

  const handleSave = () => {
    onSave(node.id, node.data)
    onClose()
  }

  return (
    <div
      className="fixed inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center z-[9999] p-4"
      onClick={onClose}
    >
      <div
        className="bg-white dark:bg-gray-900 rounded-2xl shadow-2xl w-full max-w-3xl max-h-[85vh] overflow-hidden flex flex-col border border-gray-200 dark:border-gray-800"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header with gradient */}
        <div className={cn('bg-gradient-to-r p-6 text-white relative', gradientColor)}>
          <button
            onClick={onClose}
            className="absolute top-4 right-4 p-2 hover:bg-white/20 rounded-lg transition-colors"
          >
            <X className="w-5 h-5" />
          </button>

          <div className="flex items-center gap-4">
            <div className="w-14 h-14 bg-white/20 backdrop-blur-sm rounded-xl flex items-center justify-center">
              <NodeIcon className="w-7 h-7" />
            </div>
            <div>
              <h2 className="text-2xl font-bold">{nodeLabel}</h2>
              <p className="text-white/80 text-sm mt-1">
                {nodeType.charAt(0).toUpperCase() + nodeType.slice(1)} Node
              </p>
            </div>
          </div>
        </div>

        {/* Tabs Navigation */}
        <Tabs value={activeTab} onValueChange={setActiveTab} className="flex-1 flex flex-col">
          <TabsList className="w-full justify-start rounded-none border-b bg-transparent px-6 pt-4">
            <TabsTrigger
              value="general"
              className="data-[state=active]:bg-primary data-[state=active]:text-primary-foreground"
            >
              <Settings className="w-4 h-4 mr-2" />
              General
            </TabsTrigger>
            <TabsTrigger
              value="configuration"
              className="data-[state=active]:bg-primary data-[state=active]:text-primary-foreground"
            >
              <Code className="w-4 h-4 mr-2" />
              Configuration
            </TabsTrigger>
            <TabsTrigger
              value="info"
              className="data-[state=active]:bg-primary data-[state=active]:text-primary-foreground"
            >
              <Info className="w-4 h-4 mr-2" />
              Info
            </TabsTrigger>
          </TabsList>

          {/* Tab Content */}
          <div className="flex-1 overflow-y-auto">
            <TabsContent value="general" className="p-6 space-y-6 mt-0">
              {/* Node Name */}
              <div className="space-y-2">
                <Label htmlFor="node-name" className="text-sm font-semibold">
                  Node Name
                </Label>
                <Input
                  id="node-name"
                  defaultValue={nodeLabel}
                  placeholder="Enter node name"
                  className="h-11"
                />
              </div>

              {/* Node ID (Read-only) */}
              <div className="space-y-2">
                <Label className="text-sm font-semibold text-muted-foreground">
                  Node ID
                </Label>
                <Input
                  value={node.id}
                  readOnly
                  className="h-11 bg-muted/50"
                />
              </div>
            </TabsContent>

            <TabsContent value="configuration" className="p-6 space-y-6 mt-0">
              {/* Inject Node */}
              {nodeType === 'inject' && (
                <div className="space-y-6">
                  <div className="space-y-2">
                    <Label htmlFor="interval" className="text-sm font-semibold">
                      Interval (seconds)
                    </Label>
                    <Input
                      id="interval"
                      type="number"
                      defaultValue={node.data.config?.interval || 60}
                      placeholder="60"
                      className="h-11"
                    />
                    <p className="text-xs text-muted-foreground">
                      How often to trigger the inject node
                    </p>
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="payload" className="text-sm font-semibold">
                      Payload
                    </Label>
                    <textarea
                      id="payload"
                      defaultValue={node.data.config?.payload || '{}'}
                      className="w-full px-3 py-3 border border-input rounded-lg bg-background text-sm focus:ring-2 focus:ring-primary font-mono min-h-[120px]"
                      placeholder='{"message": "Hello"}'
                    />
                    <p className="text-xs text-muted-foreground">
                      JSON payload to send
                    </p>
                  </div>
                </div>
              )}

              {/* Debug Node */}
              {nodeType === 'debug' && (
                <div className="space-y-2">
                  <Label htmlFor="output" className="text-sm font-semibold">
                    Output Location
                  </Label>
                  <select
                    id="output"
                    defaultValue={node.data.config?.output || 'console'}
                    className="w-full h-11 px-3 border border-input rounded-lg bg-background text-sm focus:ring-2 focus:ring-primary"
                  >
                    <option value="console">Console</option>
                    <option value="sidebar">Debug Sidebar</option>
                    <option value="both">Console & Sidebar</option>
                  </select>
                  <p className="text-xs text-muted-foreground">
                    Where to display debug output
                  </p>
                </div>
              )}

              {/* Function Node */}
              {nodeType === 'function' && (
                <div className="space-y-2">
                  <Label htmlFor="function-code" className="text-sm font-semibold">
                    Function Code
                  </Label>
                  <textarea
                    id="function-code"
                    defaultValue={node.data.config?.code || 'return msg;'}
                    className="w-full px-3 py-3 border border-input rounded-lg bg-background text-sm font-mono focus:ring-2 focus:ring-primary min-h-[250px]"
                    placeholder="return msg;"
                  />
                  <p className="text-xs text-muted-foreground">
                    JavaScript code to execute on each message
                  </p>
                </div>
              )}

              {/* HTTP Request Node */}
              {nodeType === 'http-request' && (
                <div className="space-y-6">
                  <div className="space-y-2">
                    <Label htmlFor="method" className="text-sm font-semibold">
                      HTTP Method
                    </Label>
                    <select
                      id="method"
                      defaultValue={node.data.config?.method || 'GET'}
                      className="w-full h-11 px-3 border border-input rounded-lg bg-background text-sm focus:ring-2 focus:ring-primary"
                    >
                      <option value="GET">GET</option>
                      <option value="POST">POST</option>
                      <option value="PUT">PUT</option>
                      <option value="DELETE">DELETE</option>
                      <option value="PATCH">PATCH</option>
                    </select>
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="url" className="text-sm font-semibold">
                      URL
                    </Label>
                    <Input
                      id="url"
                      defaultValue={node.data.config?.url || ''}
                      placeholder="https://api.example.com/data"
                      className="h-11"
                    />
                  </div>
                </div>
              )}

              {/* MQTT In Node */}
              {nodeType === 'mqtt-in' && (
                <div className="space-y-6">
                  <div className="space-y-2">
                    <Label htmlFor="broker" className="text-sm font-semibold">
                      Broker URL
                    </Label>
                    <Input
                      id="broker"
                      defaultValue={node.data.config?.broker || ''}
                      placeholder="mqtt://localhost:1883"
                      className="h-11"
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="topic" className="text-sm font-semibold">
                      Topic
                    </Label>
                    <Input
                      id="topic"
                      defaultValue={node.data.config?.topic || ''}
                      placeholder="sensors/temperature"
                      className="h-11"
                    />
                  </div>
                </div>
              )}

              {/* GPIO Out Node */}
              {nodeType === 'gpio-out' && (
                <div className="space-y-6">
                  <div className="space-y-2">
                    <Label htmlFor="pin" className="text-sm font-semibold">
                      GPIO Pin Number
                    </Label>
                    <Input
                      id="pin"
                      type="number"
                      defaultValue={node.data.config?.pin || 17}
                      placeholder="17"
                      className="h-11"
                    />
                    <p className="text-xs text-muted-foreground">
                      BCM pin numbering
                    </p>
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="initial-state" className="text-sm font-semibold">
                      Initial State
                    </Label>
                    <select
                      id="initial-state"
                      defaultValue={node.data.config?.initialState || 'low'}
                      className="w-full h-11 px-3 border border-input rounded-lg bg-background text-sm focus:ring-2 focus:ring-primary"
                    >
                      <option value="low">Low (0V)</option>
                      <option value="high">High (3.3V)</option>
                    </select>
                  </div>
                </div>
              )}

              {/* Generic Configuration */}
              {!['inject', 'debug', 'function', 'http-request', 'mqtt-in', 'gpio-out'].includes(nodeType) && (
                <div className="space-y-2">
                  <Label htmlFor="config" className="text-sm font-semibold">
                    Configuration (JSON)
                  </Label>
                  <textarea
                    id="config"
                    defaultValue={JSON.stringify(node.data.config || {}, null, 2)}
                    className="w-full px-3 py-3 border border-input rounded-lg bg-background text-sm font-mono focus:ring-2 focus:ring-primary min-h-[200px]"
                    placeholder="{}"
                  />
                  <p className="text-xs text-muted-foreground">
                    Node configuration in JSON format
                  </p>
                </div>
              )}
            </TabsContent>

            <TabsContent value="info" className="p-6 space-y-6 mt-0">
              {/* Description */}
              <div className="space-y-2">
                <Label htmlFor="description" className="text-sm font-semibold">
                  Description
                </Label>
                <textarea
                  id="description"
                  defaultValue={node.data.config?.description || ''}
                  className="w-full px-3 py-3 border border-input rounded-lg bg-background text-sm focus:ring-2 focus:ring-primary min-h-[120px]"
                  placeholder="Add a description for this node..."
                />
                <p className="text-xs text-muted-foreground">
                  Optional notes about what this node does
                </p>
              </div>

              {/* Node Info Card */}
              <div className="p-4 bg-muted/50 rounded-lg border border-border">
                <h4 className="text-sm font-semibold mb-3">Node Information</h4>
                <dl className="space-y-2 text-sm">
                  <div className="flex justify-between">
                    <dt className="text-muted-foreground">Type:</dt>
                    <dd className="font-medium">{nodeType}</dd>
                  </div>
                  <div className="flex justify-between">
                    <dt className="text-muted-foreground">ID:</dt>
                    <dd className="font-mono text-xs">{node.id}</dd>
                  </div>
                  <div className="flex justify-between">
                    <dt className="text-muted-foreground">Created:</dt>
                    <dd className="font-medium">
                      {new Date(parseInt(node.id.split('-').pop() || '0')).toLocaleString()}
                    </dd>
                  </div>
                </dl>
              </div>
            </TabsContent>
          </div>
        </Tabs>

        {/* Footer */}
        <div className="flex items-center justify-between p-6 border-t border-border bg-muted/30">
          <p className="text-xs text-muted-foreground">
            Double-click nodes to edit their settings
          </p>
          <div className="flex gap-3">
            <Button variant="outline" onClick={onClose} className="min-w-[100px]">
              Cancel
            </Button>
            <Button onClick={handleSave} className="min-w-[100px]">
              Save Changes
            </Button>
          </div>
        </div>
      </div>
    </div>
  )
}
