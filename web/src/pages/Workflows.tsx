import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { useFlowStore } from '@/stores/flowStore'
import { toast } from 'sonner'
import { downloadFlow } from '@/utils/flowImportExport'
import { Flow as ExportFlow } from '@/types/flow'
import { wsClient, WSMessage } from '@/services/websocket'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  DropdownMenuSeparator,
} from '@/components/ui/dropdown-menu'
import {
  Plus,
  Search,
  MoreVertical,
  Play,
  Square,
  Edit,
  Copy,
  Trash2,
  Download,
  FileJson,
} from 'lucide-react'

export default function Workflows() {
  const { flows, fetchFlows, startFlow, stopFlow, deleteFlow, createFlow, updateFlow, loading } = useFlowStore()
  const [searchQuery, setSearchQuery] = useState('')

  useEffect(() => {
    fetchFlows()

    // Subscribe to real-time flow status updates
    wsClient.connect()
    const unsub = wsClient.on('flow_status', (msg: WSMessage) => {
      const data = msg.data as { flow_id: string; action: string; status: string }
      if (data.flow_id) {
        const newStatus = data.action === 'started' ? 'running' : 'stopped'
        useFlowStore.setState((state) => ({
          flows: state.flows.map((f) =>
            f.id === data.flow_id ? { ...f, status: newStatus } : f
          ),
        }))
      }
    })

    return () => {
      unsub()
    }
  }, [fetchFlows])

  const filteredFlows = flows.filter(
    (flow) =>
      flow.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      flow.description?.toLowerCase().includes(searchQuery.toLowerCase())
  )

  const getStatusVariant = (status: string) => {
    switch (status) {
      case 'running':
        return 'success'
      case 'error':
        return 'destructive'
      default:
        return 'secondary'
    }
  }

  const handleDelete = async (id: string, name: string) => {
    if (confirm(`Are you sure you want to delete workflow "${name}"?`)) {
      await deleteFlow(id)
    }
  }

  const handleExport = (flow: any) => {
    // Convert nodes from map (API format) to array (export format)
    const nodesArray = flow.nodes
      ? Array.isArray(flow.nodes)
        ? flow.nodes
        : Object.values(flow.nodes).map((n: any) => ({
            id: n.id,
            type: n.type,
            name: n.name,
            config: n.config || {},
            position: n.config?.position || n.position,
          }))
      : []

    const exportFlow: ExportFlow = {
      id: flow.id,
      name: flow.name,
      description: flow.description || '',
      nodes: nodesArray,
      connections: flow.connections || [],
      status: flow.status,
    }
    downloadFlow(exportFlow)
    toast.success(`Workflow "${flow.name}" exported`)
  }

  const handleDuplicate = async (flow: any) => {
    try {
      const newFlow = await createFlow(`${flow.name} (copy)`, flow.description || '')
      if (!newFlow) {
        toast.error('Failed to duplicate workflow')
        return
      }
      await updateFlow(newFlow.id, {
        nodes: flow.nodes,
        connections: flow.connections || [],
      })
      await fetchFlows()
      toast.success(`Workflow duplicated as "${flow.name} (copy)"`)
    } catch (error) {
      toast.error('Failed to duplicate workflow')
    }
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Workflows</h1>
          <p className="text-muted-foreground mt-1">
            Manage and organize your automation workflows
          </p>
        </div>
        <Button asChild>
          <Link to="/editor">
            <Plus className="mr-2 h-4 w-4" />
            New Workflow
          </Link>
        </Button>
      </div>

      {/* Search & Filters */}
      <div className="flex items-center gap-4">
        <div className="relative flex-1 max-w-md">
          <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 h-4 w-4 text-muted-foreground" />
          <Input
            type="search"
            placeholder="Search workflows..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            className="pl-9"
          />
        </div>
        <Badge variant="secondary">
          {filteredFlows.length} workflow{filteredFlows.length !== 1 ? 's' : ''}
        </Badge>
      </div>

      {/* Workflows Grid */}
      {filteredFlows.length === 0 ? (
        <Card>
          <CardContent className="py-12 text-center">
            <FileJson className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
            <p className="text-muted-foreground mb-4">
              {searchQuery ? 'No workflows found' : 'No workflows created yet'}
            </p>
            <Button asChild>
              <Link to="/editor">
                <Plus className="mr-2 h-4 w-4" />
                Create your first workflow
              </Link>
            </Button>
          </CardContent>
        </Card>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {filteredFlows.map((flow) => (
            <Card key={flow.id} className="hover:shadow-lg transition-shadow">
              <CardHeader className="pb-3">
                <div className="flex items-start justify-between">
                  <div className="flex-1 min-w-0">
                    <CardTitle className="truncate">{flow.name}</CardTitle>
                    <CardDescription className="line-clamp-2 mt-1">
                      {flow.description || 'No description'}
                    </CardDescription>
                  </div>
                  <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                      <Button variant="ghost" size="icon" className="h-8 w-8">
                        <MoreVertical className="h-4 w-4" />
                      </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                      <DropdownMenuItem asChild>
                        <Link to={`/editor/${flow.id}`}>
                          <Edit className="h-4 w-4 mr-2" />
                          Edit
                        </Link>
                      </DropdownMenuItem>
                      <DropdownMenuItem onClick={() => handleDuplicate(flow)}>
                        <Copy className="h-4 w-4 mr-2" />
                        Duplicate
                      </DropdownMenuItem>
                      <DropdownMenuItem onClick={() => handleExport(flow)}>
                        <Download className="h-4 w-4 mr-2" />
                        Export
                      </DropdownMenuItem>
                      <DropdownMenuSeparator />
                      <DropdownMenuItem
                        className="text-destructive"
                        onClick={() => handleDelete(flow.id, flow.name)}
                      >
                        <Trash2 className="h-4 w-4 mr-2" />
                        Delete
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </div>
              </CardHeader>
              <CardContent>
                <div className="flex items-center justify-between mb-4">
                  <Badge variant={getStatusVariant(flow.status)}>
                    {flow.status === 'running'
                      ? 'Running'
                      : flow.status === 'error'
                      ? 'Error'
                      : 'Stopped'}
                  </Badge>
                  <div className="text-xs text-muted-foreground">
                    {Object.keys(flow.nodes || {}).length} nodes •{' '}
                    {(flow.connections || []).length} connections
                  </div>
                </div>
                <div className="flex items-center gap-1">
                  {flow.status === 'running' ? (
                    <Button variant="outline" size="sm" onClick={() => stopFlow(flow.id)}>
                      <Square className="h-4 w-4 mr-1" />
                      Stop
                    </Button>
                  ) : (
                    <Button variant="outline" size="sm" onClick={() => startFlow(flow.id)}>
                      <Play className="h-4 w-4 mr-1" />
                      Run
                    </Button>
                  )}
                  <Button variant="ghost" size="sm" asChild className="flex-1">
                    <Link to={`/editor/${flow.id}`}>
                      <Edit className="h-4 w-4 mr-1" />
                      Edit
                    </Link>
                  </Button>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  )
}
