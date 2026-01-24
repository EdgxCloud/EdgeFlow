import { useState, useEffect } from 'react'
import { Input } from '@/components/ui/input'
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from '@/components/ui/accordion'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Badge } from '@/components/ui/badge'
import {
  Search, Clock, Loader2, Download, Upload, Code, Cpu, CircuitBoard,
  Thermometer, Gauge, Radio, Network, Database, HardDrive, MessageSquare,
  Brain, Factory, LayoutDashboard, Settings, Box, ChevronDown
} from 'lucide-react'
import { cn } from '@/lib/utils'
import { nodesApi, NodeType, NodeTypeCategory } from '@/services/nodes'
import { toast } from 'sonner'

// Icon mapping for categories
const CATEGORY_ICONS: Record<string, React.ComponentType<any>> = {
  input: Download,
  output: Upload,
  function: Code,
  processing: Cpu,
  gpio: CircuitBoard,
  sensors: Thermometer,
  actuators: Gauge,
  communication: Radio,
  network: Network,
  database: Database,
  storage: HardDrive,
  messaging: MessageSquare,
  ai: Brain,
  industrial: Factory,
  dashboard: LayoutDashboard,
  advanced: Settings,
}

// Default category config for fallback
const defaultCategoryConfig: Record<string, { color: string; bgColor: string }> = {
  input: { color: 'text-emerald-600', bgColor: 'bg-emerald-100 dark:bg-emerald-900/30' },
  output: { color: 'text-red-600', bgColor: 'bg-red-100 dark:bg-red-900/30' },
  function: { color: 'text-amber-600', bgColor: 'bg-amber-100 dark:bg-amber-900/30' },
  processing: { color: 'text-violet-600', bgColor: 'bg-violet-100 dark:bg-violet-900/30' },
  gpio: { color: 'text-green-600', bgColor: 'bg-green-100 dark:bg-green-900/30' },
  sensors: { color: 'text-green-500', bgColor: 'bg-green-100 dark:bg-green-900/30' },
  actuators: { color: 'text-pink-600', bgColor: 'bg-pink-100 dark:bg-pink-900/30' },
  communication: { color: 'text-sky-600', bgColor: 'bg-sky-100 dark:bg-sky-900/30' },
  network: { color: 'text-cyan-600', bgColor: 'bg-cyan-100 dark:bg-cyan-900/30' },
  database: { color: 'text-blue-600', bgColor: 'bg-blue-100 dark:bg-blue-900/30' },
  storage: { color: 'text-indigo-600', bgColor: 'bg-indigo-100 dark:bg-indigo-900/30' },
  messaging: { color: 'text-teal-600', bgColor: 'bg-teal-100 dark:bg-teal-900/30' },
  ai: { color: 'text-purple-600', bgColor: 'bg-purple-100 dark:bg-purple-900/30' },
  industrial: { color: 'text-orange-600', bgColor: 'bg-orange-100 dark:bg-orange-900/30' },
  dashboard: { color: 'text-cyan-700', bgColor: 'bg-cyan-100 dark:bg-cyan-900/30' },
  advanced: { color: 'text-slate-600', bgColor: 'bg-slate-100 dark:bg-slate-900/30' },
}

interface CategoryData {
  id: string
  name: string
  description?: string
  icon?: string
  color?: string
  order?: number
  count?: number
}

interface NodePaletteProps {
  className?: string
}

export default function NodePalette({ className }: NodePaletteProps) {
  const [searchQuery, setSearchQuery] = useState('')
  const [recentNodes] = useState<string[]>(['inject', 'debug', 'function'])
  const [categories, setCategories] = useState<CategoryData[]>([])
  const [nodeTypes, setNodeTypes] = useState<NodeType[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [expandedCategory, setExpandedCategory] = useState<string | undefined>(undefined)

  // Load node types from backend on mount
  useEffect(() => {
    const loadNodeTypes = async () => {
      setIsLoading(true)
      try {
        const response = await nodesApi.getTypes()
        setCategories(response.categories)
        setNodeTypes(response.node_types)
        console.log('📦 Loaded', response.node_types.length, 'node types in', response.categories.length, 'categories')

        // All categories closed by default
        setExpandedCategory(undefined)
      } catch (error) {
        console.error('❌ Failed to load node types:', error)
        toast.error('Failed to load node types', {
          description: 'Using fallback node list',
        })
      } finally {
        setIsLoading(false)
      }
    }

    loadNodeTypes()
  }, [])

  const onDragStart = (event: React.DragEvent, nodeType: string, nodeData: any) => {
    event.dataTransfer.setData('application/reactflow', nodeType)
    event.dataTransfer.setData('application/json', JSON.stringify(nodeData))
    event.dataTransfer.effectAllowed = 'move'
  }

  // Group node types by category
  const nodeCategories = categories
    .map((category) => {
      const categoryNodes = nodeTypes
        .filter((node) => node.category === category.id)
        .sort((a, b) => a.name.localeCompare(b.name))

      const colorConfig = defaultCategoryConfig[category.id] || {
        color: 'text-gray-600',
        bgColor: 'bg-gray-100 dark:bg-gray-900/20',
      }

      const IconComponent = CATEGORY_ICONS[category.id] || Box

      return {
        id: category.id,
        name: category.name,
        description: category.description || '',
        icon: IconComponent,
        apiColor: category.color,
        ...colorConfig,
        count: category.count || categoryNodes.length,
        nodes: categoryNodes.map((node) => ({
          type: node.type || node.id,
          label: node.name,
          description: node.description || '',
          category: category.id,
          color: node.color,
        })),
      }
    })
    .filter(cat => cat.nodes.length > 0)

  // Filter nodes based on search
  const filteredCategories = searchQuery
    ? nodeCategories
        .map((category) => ({
          ...category,
          nodes: category.nodes.filter(
            (node) =>
              node.label.toLowerCase().includes(searchQuery.toLowerCase()) ||
              node.description.toLowerCase().includes(searchQuery.toLowerCase()) ||
              node.type.toLowerCase().includes(searchQuery.toLowerCase())
          ),
        }))
        .filter((category) => category.nodes.length > 0)
    : nodeCategories

  // Handle accordion value change - only one open at a time
  const handleAccordionChange = (value: string) => {
    setExpandedCategory(value)
  }

  return (
    <div className={cn('flex flex-col h-full bg-card border-r border-border', className)}>
      {/* Header */}
      <div className="p-4 border-b border-border">
        <h2 className="text-lg font-semibold mb-3">Node Palette</h2>
        <div className="relative">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            type="search"
            placeholder="Search nodes..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-9"
          />
        </div>
        {/* Category count */}
        {!isLoading && (
          <div className="mt-2 text-xs text-muted-foreground">
            {nodeTypes.length} nodes in {categories.length} categories
          </div>
        )}
      </div>

      {/* Recent Nodes */}
      {recentNodes.length > 0 && !searchQuery && !isLoading && (
        <div className="p-4 border-b border-border">
          <div className="flex items-center gap-2 mb-2">
            <Clock className="h-4 w-4 text-muted-foreground" />
            <h3 className="text-sm font-medium">Recently Used</h3>
          </div>
          <div className="flex flex-wrap gap-2">
            {recentNodes.map((nodeType) => {
              const node = nodeCategories
                .flatMap((cat) => cat.nodes)
                .find((n) => n.type === nodeType)
              if (!node) return null

              return (
                <Badge
                  key={nodeType}
                  variant="secondary"
                  className="cursor-move hover:bg-accent"
                  draggable
                  onDragStart={(e) =>
                    onDragStart(e, node.type, {
                      label: node.label,
                      category: node.category,
                    })
                  }
                >
                  {node.label}
                </Badge>
              )
            })}
          </div>
        </div>
      )}

      {/* Loading State */}
      {isLoading && (
        <div className="flex-1 flex items-center justify-center">
          <div className="text-center">
            <Loader2 className="h-8 w-8 animate-spin text-primary mx-auto mb-2" />
            <p className="text-sm text-muted-foreground">Loading node types...</p>
          </div>
        </div>
      )}

      {/* Node Categories with Accordion */}
      {!isLoading && (
        <ScrollArea className="flex-1">
          <div className="p-2">
            <Accordion
              type="single"
              collapsible
              value={expandedCategory}
              onValueChange={handleAccordionChange}
              className="space-y-1"
            >
              {filteredCategories.map((category) => {
                const IconComponent = category.icon
                return (
                  <AccordionItem
                    key={category.id}
                    value={category.id}
                    className="border rounded-lg overflow-hidden bg-background"
                  >
                    <AccordionTrigger className="hover:no-underline px-3 py-2.5 hover:bg-accent/50 [&[data-state=open]>div>svg.chevron]:rotate-180">
                      <div className="flex items-center gap-3 flex-1">
                        <div
                          className={cn(
                            'w-8 h-8 rounded-lg flex items-center justify-center',
                            category.bgColor
                          )}
                          style={category.apiColor ? { backgroundColor: category.apiColor + '20' } : undefined}
                        >
                          <IconComponent
                            className={cn('h-4 w-4', category.color)}
                            style={category.apiColor ? { color: category.apiColor } : undefined}
                          />
                        </div>
                        <div className="flex-1 text-left">
                          <div className="font-medium text-sm">{category.name}</div>
                          {category.description && (
                            <div className="text-xs text-muted-foreground line-clamp-1">
                              {category.description}
                            </div>
                          )}
                        </div>
                        <Badge variant="outline" className="ml-2 text-xs font-normal px-2">
                          {category.count}
                        </Badge>
                        <ChevronDown className="h-4 w-4 shrink-0 text-muted-foreground transition-transform duration-200 chevron" />
                      </div>
                    </AccordionTrigger>
                    <AccordionContent className="pb-0">
                      <div className="space-y-1 p-2 pt-1 border-t bg-muted/30">
                        {category.nodes.map((node) => (
                          <div
                            key={node.type}
                            draggable
                            onDragStart={(e) =>
                              onDragStart(e, node.type, {
                                label: node.label,
                                category: category.id,
                              })
                            }
                            className={cn(
                              'p-2.5 rounded-md cursor-move transition-all',
                              'hover:bg-accent border border-transparent hover:border-border',
                              'hover:shadow-sm group'
                            )}
                          >
                            <div className="flex items-start gap-2.5">
                              <div
                                className="w-6 h-6 rounded flex-shrink-0 flex items-center justify-center text-xs font-bold text-white shadow-sm"
                                style={{ backgroundColor: node.color || category.apiColor || '#64748b' }}
                              >
                                {node.label.charAt(0).toUpperCase()}
                              </div>
                              <div className="flex-1 min-w-0">
                                <div className="font-medium text-sm group-hover:text-primary">
                                  {node.label}
                                </div>
                                {node.description && (
                                  <div className="text-xs text-muted-foreground line-clamp-2 mt-0.5">
                                    {node.description}
                                  </div>
                                )}
                              </div>
                            </div>
                          </div>
                        ))}
                      </div>
                    </AccordionContent>
                  </AccordionItem>
                )
              })}
            </Accordion>

            {filteredCategories.length === 0 && (
              <div className="p-8 text-center text-muted-foreground">
                <Search className="h-8 w-8 mx-auto mb-2 opacity-50" />
                <p className="text-sm font-medium">No nodes found</p>
                <p className="text-xs mt-1">Try a different search term</p>
              </div>
            )}
          </div>
        </ScrollArea>
      )}
    </div>
  )
}
