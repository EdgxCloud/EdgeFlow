import { useState, useEffect } from 'react'
import {
  Search, ChevronDown, ChevronRight, Zap, Download, Upload, Cpu, Database, Network,
  MessageSquare, Bot, Settings, Gauge, Thermometer, Server, Code, Bug, GitBranch,
  Clock, Play, Scissors, Link2, AlertTriangle, CheckCircle, Activity, FileText,
  Terminal, Mail, Send, Sliders, Split, Timer, Workflow, FileCode, Globe,
  Wifi, Radio, Key, HardDrive, Cloud, Hash, Wind, Droplets, Eye, Lightbulb,
  ToggleLeft, Power, Gauge as GaugeIcon, Rotate3D, Brain, Sparkles, Binary,
  FileInput, Webhook, CircuitBoard
} from 'lucide-react'
import { nodeTypesApi, NodeType } from '../../lib/api'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'

// Icon mapping for each node type
const NODE_ICONS: Record<string, any> = {
  // Input nodes
  'inject': Zap,
  'mqtt-in': Download,

  // Output nodes
  'debug': Bug,
  'mqtt-out': Upload,

  // GPIO nodes (Raspberry Pi)
  'gpio-in': CircuitBoard,
  'gpio-out': CircuitBoard,

  // Function nodes
  'function': Code,
  'change': Settings,
  'range': Sliders,
  'template': FileCode,

  // Logic nodes
  'switch': GitBranch,
  'if': GitBranch,
  'delay': Clock,

  // Sensors
  'dht': Thermometer,
  'ds18b20': Thermometer,
  'bmp280': Gauge,

  // Actuators
  'pwm': Activity,
  'servo': Rotate3D,
  'relay': Power,

  // Network nodes
  'http-request': Globe,
  'http-webhook': Webhook,
  'webhook': Webhook,
  'http-in': Webhook,
  'websocket': Network,
  'tcp': Wifi,
  'udp': Radio,

  // Database nodes
  'mysql': Database,
  'postgresql': Database,
  'mongodb': Database,
  'redis': Hash,

  // Messaging nodes
  'email': Mail,
  'telegram': Send,
  'slack': MessageSquare,
  'discord': MessageSquare,

  // AI nodes
  'openai': Brain,
  'anthropic': Sparkles,
  'ollama': Bot,

  // Advanced nodes
  'exec': Terminal,
  'file': FileText,
  'split': Scissors,
  'join': Link2,
  'catch': AlertTriangle,
  'complete': CheckCircle,
  'status': Activity,
  'set': Settings,
}

// Better categorization based on Node-RED best practices
const CATEGORY_CONFIG = {
  input: {
    label: 'Input',
    icon: Download,
    color: '#10b981',
    description: 'Nodes that receive data from external sources'
  },
  output: {
    label: 'Output',
    icon: Upload,
    color: '#ef4444',
    description: 'Nodes that send data to external destinations'
  },
  function: {
    label: 'Function',
    icon: Code,
    color: '#f59e0b',
    description: 'Nodes for logic, transformation, and processing'
  },
  logic: {
    label: 'Logic',
    icon: Cpu,
    color: '#8b5cf6',
    description: 'Conditional routing and flow control'
  },
  gpio: {
    label: 'GPIO',
    icon: CircuitBoard,
    color: '#16a34a',
    description: 'Raspberry Pi GPIO pins input/output'
  },
  sensors: {
    label: 'Sensors',
    icon: Thermometer,
    color: '#22c55e',
    description: 'Physical sensors and hardware inputs'
  },
  actuators: {
    label: 'Actuators',
    icon: Gauge,
    color: '#ec4899',
    description: 'Physical hardware outputs and controls'
  },
  network: {
    label: 'Network',
    icon: Network,
    color: '#06b6d4',
    description: 'Network protocols and communication'
  },
  database: {
    label: 'Database',
    icon: Database,
    color: '#3b82f6',
    description: 'Database storage and queries'
  },
  messaging: {
    label: 'Messaging',
    icon: MessageSquare,
    color: '#14b8a6',
    description: 'Messaging platforms and notifications'
  },
  ai: {
    label: 'AI & ML',
    icon: Bot,
    color: '#a855f7',
    description: 'Artificial intelligence and machine learning'
  },
  advanced: {
    label: 'Advanced',
    icon: Settings,
    color: '#64748b',
    description: 'Advanced nodes and utilities'
  },
}

// Improved fallback nodes with better categorization
const fallbackNodes: NodeType[] = [
  // Input nodes
  { type: 'inject', name: 'Inject', description: 'Manually trigger flows or inject timestamps at intervals', category: 'input', color: '#10b981', icon: '', inputs: [], outputs: ['output'], properties: [] },
  { type: 'mqtt-in', name: 'MQTT In', description: 'Subscribe to MQTT broker topics and receive messages', category: 'input', color: '#10b981', icon: '', inputs: [], outputs: ['output'], properties: [] },

  // Output nodes
  { type: 'debug', name: 'Debug', description: 'Display messages in the debug sidebar for troubleshooting', category: 'output', color: '#ef4444', icon: '', inputs: ['input'], outputs: [], properties: [] },
  { type: 'mqtt-out', name: 'MQTT Out', description: 'Publish messages to MQTT broker topics', category: 'output', color: '#ef4444', icon: '', inputs: ['input'], outputs: [], properties: [] },

  // GPIO nodes (Raspberry Pi)
  { type: 'gpio-in', name: 'GPIO In', description: 'Read digital values from Raspberry Pi GPIO pins', category: 'gpio', color: '#16a34a', icon: '', inputs: [], outputs: ['output'], properties: [] },
  { type: 'gpio-out', name: 'GPIO Out', description: 'Write digital or PWM values to Raspberry Pi GPIO pins', category: 'gpio', color: '#16a34a', icon: '', inputs: ['input'], outputs: [], properties: [] },

  // Function nodes
  { type: 'function', name: 'Function', description: 'Write custom JavaScript code to process messages', category: 'function', color: '#f59e0b', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'change', name: 'Change', description: 'Set, change, move or delete message properties', category: 'function', color: '#f59e0b', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'range', name: 'Range', description: 'Map and scale numeric values between ranges', category: 'function', color: '#f59e0b', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'template', name: 'Template', description: 'Generate text using Mustache template syntax', category: 'function', color: '#f59e0b', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },

  // Logic nodes
  { type: 'switch', name: 'Switch', description: 'Route messages based on property values and rules', category: 'logic', color: '#8b5cf6', icon: '', inputs: ['input'], outputs: ['output1', 'output2', 'output3'], properties: [] },
  { type: 'if', name: 'If', description: 'Simple conditional branching with if-then-else logic', category: 'logic', color: '#8b5cf6', icon: '', inputs: ['input'], outputs: ['true', 'false'], properties: [] },
  { type: 'delay', name: 'Delay', description: 'Delay messages or limit message rate', category: 'logic', color: '#8b5cf6', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },

  // Sensors
  { type: 'dht', name: 'DHT Sensor', description: 'Read temperature and humidity from DHT11/DHT22 sensors', category: 'sensors', color: '#22c55e', icon: '', inputs: [], outputs: ['output'], properties: [] },
  { type: 'ds18b20', name: 'DS18B20', description: 'Read temperature from DS18B20 digital sensors', category: 'sensors', color: '#22c55e', icon: '', inputs: [], outputs: ['output'], properties: [] },
  { type: 'bmp280', name: 'BMP280', description: 'Read temperature and pressure from BMP280 sensor', category: 'sensors', color: '#22c55e', icon: '', inputs: [], outputs: ['output'], properties: [] },

  // Actuators
  { type: 'pwm', name: 'PWM', description: 'Generate pulse width modulation signals', category: 'actuators', color: '#ec4899', icon: '', inputs: ['input'], outputs: [], properties: [] },
  { type: 'servo', name: 'Servo', description: 'Control servo motor position and angle', category: 'actuators', color: '#ec4899', icon: '', inputs: ['input'], outputs: [], properties: [] },
  { type: 'relay', name: 'Relay', description: 'Control relay switches for high-power devices', category: 'actuators', color: '#ec4899', icon: '', inputs: ['input'], outputs: [], properties: [] },

  // Network nodes
  { type: 'http-request', name: 'HTTP Request', description: 'Make HTTP/HTTPS requests to APIs and web services', category: 'network', color: '#06b6d4', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'http-webhook', name: 'HTTP Webhook', description: 'Create HTTP endpoints to receive incoming webhook requests', category: 'input', color: '#10b981', icon: '', inputs: [], outputs: ['output'], properties: [] },
  { type: 'websocket', name: 'WebSocket', description: 'Establish bidirectional WebSocket connections', category: 'network', color: '#06b6d4', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'tcp', name: 'TCP', description: 'Create TCP client or server connections', category: 'network', color: '#06b6d4', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'udp', name: 'UDP', description: 'Send and receive UDP network packets', category: 'network', color: '#06b6d4', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },

  // Database nodes
  { type: 'mysql', name: 'MySQL', description: 'Execute queries on MySQL/MariaDB databases', category: 'database', color: '#3b82f6', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'postgresql', name: 'PostgreSQL', description: 'Execute queries on PostgreSQL databases', category: 'database', color: '#3b82f6', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'mongodb', name: 'MongoDB', description: 'Perform operations on MongoDB collections', category: 'database', color: '#3b82f6', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'redis', name: 'Redis', description: 'Store and retrieve data from Redis cache', category: 'database', color: '#3b82f6', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },

  // Messaging nodes
  { type: 'email', name: 'Email', description: 'Send emails via SMTP server', category: 'messaging', color: '#14b8a6', icon: '', inputs: ['input'], outputs: [], properties: [] },
  { type: 'telegram', name: 'Telegram', description: 'Send and receive Telegram bot messages', category: 'messaging', color: '#14b8a6', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'slack', name: 'Slack', description: 'Send messages to Slack channels and workspaces', category: 'messaging', color: '#14b8a6', icon: '', inputs: ['input'], outputs: [], properties: [] },
  { type: 'discord', name: 'Discord', description: 'Send messages to Discord channels via webhooks', category: 'messaging', color: '#14b8a6', icon: '', inputs: ['input'], outputs: [], properties: [] },

  // AI nodes
  { type: 'openai', name: 'OpenAI', description: 'Access OpenAI GPT models for text generation', category: 'ai', color: '#a855f7', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'anthropic', name: 'Anthropic', description: 'Use Anthropic Claude for conversations and analysis', category: 'ai', color: '#a855f7', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'ollama', name: 'Ollama', description: 'Run local LLM models with Ollama', category: 'ai', color: '#a855f7', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },

  // Advanced nodes
  { type: 'exec', name: 'Exec', description: 'Execute system shell commands and scripts', category: 'advanced', color: '#64748b', icon: '', inputs: ['input'], outputs: ['stdout', 'stderr'], properties: [] },
  { type: 'file', name: 'File', description: 'Read from and write to filesystem files', category: 'advanced', color: '#64748b', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'split', name: 'Split', description: 'Split messages into multiple parts', category: 'advanced', color: '#64748b', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'join', name: 'Join', description: 'Join multiple messages into one', category: 'advanced', color: '#64748b', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
]

export default function NodePalette() {
  const [nodeTypes, setNodeTypes] = useState<NodeType[]>(fallbackNodes)
  const [search, setSearch] = useState('')
  const [loading, setLoading] = useState(true)
  const [collapsedCategories, setCollapsedCategories] = useState<Set<string>>(new Set())

  useEffect(() => {
    loadNodeTypes()
  }, [])

  const loadNodeTypes = async () => {
    try {
      const response = await nodeTypesApi.list()
      const apiNodes = response.data.node_types || []

      // Merge API nodes with fallback nodes
      // Fallback nodes provide a baseline, API nodes can override or add more
      const apiNodeTypes = new Set(apiNodes.map((n: NodeType) => n.type))
      const mergedNodes = [
        ...fallbackNodes.filter(n => !apiNodeTypes.has(n.type)), // Keep fallback nodes not in API
        ...apiNodes // Add all API nodes
      ]

      setNodeTypes(mergedNodes)
    } catch (error) {
      console.error('Failed to load node types:', error)
      setNodeTypes(fallbackNodes)
    } finally {
      setLoading(false)
    }
  }

  const onDragStart = (event: React.DragEvent, nodeType: string) => {
    event.dataTransfer.setData('application/reactflow', nodeType)
    event.dataTransfer.effectAllowed = 'move'
  }

  const toggleCategory = (category: string) => {
    const newCollapsed = new Set(collapsedCategories)
    if (newCollapsed.has(category)) {
      newCollapsed.delete(category)
    } else {
      newCollapsed.add(category)
    }
    setCollapsedCategories(newCollapsed)
  }

  const filteredNodes = nodeTypes.filter((node) =>
    node.name.toLowerCase().includes(search.toLowerCase()) ||
    node.description.toLowerCase().includes(search.toLowerCase())
  )

  const groupedNodes = filteredNodes.reduce((acc, node) => {
    const category = node.category || 'advanced'
    if (!acc[category]) {
      acc[category] = []
    }
    acc[category].push(node)
    return acc
  }, {} as Record<string, NodeType[]>)

  // Sort nodes within each category alphabetically
  Object.keys(groupedNodes).forEach((category) => {
    groupedNodes[category].sort((a, b) => a.name.localeCompare(b.name))
  })

  // Sort categories by predefined order
  const categoryOrder = ['input', 'output', 'gpio', 'function', 'logic', 'sensors', 'actuators', 'network', 'database', 'messaging', 'ai', 'advanced']
  const sortedCategories = Object.keys(groupedNodes).sort((a, b) => {
    const indexA = categoryOrder.indexOf(a)
    const indexB = categoryOrder.indexOf(b)
    if (indexA === -1) return 1
    if (indexB === -1) return -1
    return indexA - indexB
  })

  return (
    <TooltipProvider delayDuration={300}>
      <div className="w-64 bg-white dark:bg-gray-800 border-l border-gray-200 dark:border-gray-700 flex flex-col h-full">
        {/* Header */}
        <div className="p-4 border-b border-gray-200 dark:border-gray-700">
          <h3 className="font-semibold text-gray-900 dark:text-white mb-3">
            Node Palette
          </h3>
          <div className="relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-400" aria-hidden="true" />
            <input
              type="text"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Search nodes..."
              aria-label="Search nodes"
              className="w-full pl-10 pr-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white text-sm focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
          </div>
        </div>

        {/* Node list */}
        <div className="flex-1 overflow-y-auto p-3 space-y-2" role="list" aria-label="Available nodes">
          {loading ? (
            <div className="flex items-center justify-center h-32">
              <div className="spinner" aria-label="Loading nodes"></div>
            </div>
          ) : (
            sortedCategories.map((category) => {
              const config = CATEGORY_CONFIG[category as keyof typeof CATEGORY_CONFIG] || {
                label: category,
                icon: Settings,
                color: '#64748b',
                description: category
              }
              const Icon = config.icon
              const isCollapsed = collapsedCategories.has(category)
              const nodes = groupedNodes[category]

              return (
                <div key={category} className="space-y-1">
                  {/* Category header - collapsible */}
                  <button
                    onClick={() => toggleCategory(category)}
                    className="w-full flex items-center gap-2 px-2 py-1.5 rounded hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors group"
                    aria-expanded={!isCollapsed}
                    aria-controls={`category-${category}`}
                  >
                    {isCollapsed ? (
                      <ChevronRight className="w-3.5 h-3.5 text-gray-500" aria-hidden="true" />
                    ) : (
                      <ChevronDown className="w-3.5 h-3.5 text-gray-500" aria-hidden="true" />
                    )}
                    <Icon className="w-4 h-4" style={{ color: config.color }} aria-hidden="true" />
                    <span className="text-xs font-semibold text-gray-700 dark:text-gray-300 uppercase tracking-wider">
                      {config.label}
                    </span>
                    <span className="ml-auto text-xs text-gray-400 dark:text-gray-500">
                      {nodes.length}
                    </span>
                  </button>

                  {/* Category nodes */}
                  {!isCollapsed && (
                    <div id={`category-${category}`} className="space-y-1 pl-2" role="group">
                      {nodes.map((node) => (
                        <Tooltip key={node.type}>
                          <TooltipTrigger asChild>
                            <div
                              draggable
                              onDragStart={(e) => onDragStart(e, node.type)}
                              className="p-2 rounded-md border border-gray-200 dark:border-gray-600 hover:border-blue-500 dark:hover:border-blue-500 hover:shadow-sm cursor-move transition-all group"
                              style={{ backgroundColor: node.color + '08' }}
                              role="listitem"
                              tabIndex={0}
                              aria-label={`${node.name}: ${node.description}`}
                              onKeyDown={(e) => {
                                if (e.key === 'Enter' || e.key === ' ') {
                                  e.preventDefault()
                                  // Could trigger drag or provide alternative interaction
                                }
                              }}
                            >
                              <div className="flex items-center gap-2.5">
                                {(() => {
                                  const NodeIcon = NODE_ICONS[node.type] || Activity
                                  return (
                                    <div
                                      className="w-7 h-7 rounded flex-shrink-0 flex items-center justify-center text-white"
                                      style={{ backgroundColor: node.color }}
                                      aria-hidden="true"
                                    >
                                      <NodeIcon className="w-4 h-4" strokeWidth={2.5} />
                                    </div>
                                  )
                                })()}
                                <span className="flex-1 min-w-0 font-medium text-xs text-gray-900 dark:text-white truncate">
                                  {node.name}
                                </span>
                              </div>
                            </div>
                          </TooltipTrigger>
                          <TooltipContent side="left" className="max-w-xs">
                            <div className="space-y-1">
                              <p className="font-semibold">{node.name}</p>
                              <p className="text-xs opacity-90">{node.description}</p>
                              {node.inputs && node.inputs.length > 0 && (
                                <p className="text-xs opacity-75">Inputs: {node.inputs.length}</p>
                              )}
                              {node.outputs && node.outputs.length > 0 && (
                                <p className="text-xs opacity-75">Outputs: {node.outputs.length}</p>
                              )}
                            </div>
                          </TooltipContent>
                        </Tooltip>
                      ))}
                    </div>
                  )}
                </div>
              )
            })
          )}

          {!loading && filteredNodes.length === 0 && (
            <div className="text-center py-8 text-gray-500 dark:text-gray-400 text-sm">
              No nodes found matching "{search}"
            </div>
          )}
        </div>
      </div>
    </TooltipProvider>
  )
}
