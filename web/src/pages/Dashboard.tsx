import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { Workflow, PlayCircle, AlertCircle, Activity, Plus, TrendingUp, ArrowRight } from 'lucide-react'
import { useFlowStore } from '../stores/flowStore'
import { healthApi } from '../lib/api'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'
import ResourceStatsCard from '@/components/dashboard/ResourceStatsCard'

interface Stats {
  totalFlows: number
  runningFlows: number
  totalExecutions: number
  errors: number
}

export default function Dashboard() {
  const { flows, fetchFlows } = useFlowStore()
  const [health, setHealth] = useState<any>(null)
  const [stats, setStats] = useState<Stats>({
    totalFlows: 0,
    runningFlows: 0,
    totalExecutions: 0,
    errors: 0,
  })

  useEffect(() => {
    fetchFlows()
    checkHealth()
  }, [])

  useEffect(() => {
    // Calculate stats from flows
    setStats({
      totalFlows: flows.length,
      runningFlows: flows.filter((f) => f.status === 'running').length,
      totalExecutions: 0, // TODO: Get from executions API
      errors: flows.filter((f) => f.status === 'error').length,
    })
  }, [flows])

  const checkHealth = async () => {
    try {
      const response = await healthApi.check()
      setHealth(response.data)
    } catch (error) {
      console.error('Health check failed:', error)
    }
  }

  const statCards = [
    {
      name: 'Total Workflows',
      value: stats.totalFlows,
      icon: Workflow,
      color: 'text-blue-600',
      bgColor: 'bg-blue-100 dark:bg-blue-900/20',
      change: '+2 today',
      trend: 'up' as const,
    },
    {
      name: 'Running',
      value: stats.runningFlows,
      icon: Activity,
      color: 'text-green-600',
      bgColor: 'bg-green-100 dark:bg-green-900/20',
      change: 'Active',
      trend: 'neutral' as const,
    },
    {
      name: 'Total Executions',
      value: stats.totalExecutions,
      icon: PlayCircle,
      color: 'text-purple-600',
      bgColor: 'bg-purple-100 dark:bg-purple-900/20',
      change: '+15 today',
      trend: 'up' as const,
    },
    {
      name: 'Errors',
      value: stats.errors,
      icon: AlertCircle,
      color: 'text-red-600',
      bgColor: 'bg-red-100 dark:bg-red-900/20',
      change: stats.errors > 0 ? 'Needs attention' : 'No errors',
      trend: stats.errors > 0 ? ('down' as const) : ('neutral' as const),
    },
  ]

  return (
    <div className="space-y-6">
      {/* Page header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">
            Dashboard
          </h1>
          <p className="text-muted-foreground mt-1">
            Welcome to EdgeFlow - Monitor your workflows and system status
          </p>
        </div>
        <Button asChild>
          <Link to="/editor">
            <Plus className="mr-2 h-4 w-4" />
            New Workflow
          </Link>
        </Button>
      </div>

      {/* Stats cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
        {statCards.map((stat) => (
          <Card key={stat.name}>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-sm font-medium">
                {stat.name}
              </CardTitle>
              <div className={`${stat.bgColor} p-2 rounded-lg`}>
                <stat.icon className={`h-4 w-4 ${stat.color}`} />
              </div>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stat.value}</div>
              <p className="text-xs text-muted-foreground mt-1 flex items-center gap-1">
                {stat.trend === 'up' && <TrendingUp className="h-3 w-3 text-green-600" />}
                {stat.change}
              </p>
            </CardContent>
          </Card>
        ))}
      </div>

      {/* System status */}
      {health && (
        <Card>
          <CardHeader>
            <CardTitle>System Status</CardTitle>
            <CardDescription>Current system health and metrics</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              <div className="flex items-center gap-3">
                <div className="h-2 w-2 rounded-full bg-green-500"></div>
                <div>
                  <p className="text-sm text-muted-foreground">Server Status</p>
                  <p className="font-medium">{health.status}</p>
                </div>
              </div>
              <div className="flex items-center gap-3">
                <div className="h-2 w-2 rounded-full bg-blue-500"></div>
                <div>
                  <p className="text-sm text-muted-foreground">Version</p>
                  <p className="font-medium">{health.version}</p>
                </div>
              </div>
              <div className="flex items-center gap-3">
                <div className="h-2 w-2 rounded-full bg-purple-500"></div>
                <div>
                  <p className="text-sm text-muted-foreground">Online Users</p>
                  <p className="font-medium">{health.websocket_clients}</p>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      )}

      {/* Resource Stats */}
      <ResourceStatsCard autoRefresh={true} refreshInterval={5000} />

      {/* Recent flows */}
      <Card>
        <CardHeader className="flex flex-row items-center justify-between space-y-0">
          <div>
            <CardTitle>Recent Workflows</CardTitle>
            <CardDescription>Your latest workflow activities</CardDescription>
          </div>
          <Button variant="ghost" size="sm" asChild>
            <Link to="/workflows">
              View all
              <ArrowRight className="ml-2 h-4 w-4" />
            </Link>
          </Button>
        </CardHeader>
        <CardContent>
          <div className="space-y-4">
            {flows.slice(0, 5).map((flow, index) => (
              <div key={flow.id}>
                <div className="flex items-center justify-between">
                  <div className="flex-1">
                    <h3 className="font-medium">{flow.name}</h3>
                    <p className="text-sm text-muted-foreground mt-0.5">
                      {flow.description}
                    </p>
                  </div>
                  <div className="flex items-center gap-3">
                    <Badge
                      variant={
                        flow.status === 'running'
                          ? 'success'
                          : flow.status === 'error'
                          ? 'destructive'
                          : 'secondary'
                      }
                    >
                      {flow.status === 'running'
                        ? 'Running'
                        : flow.status === 'error'
                        ? 'Error'
                        : 'Stopped'}
                    </Badge>
                    <Button variant="ghost" size="sm" asChild>
                      <Link to={`/editor/${flow.id}`}>Edit</Link>
                    </Button>
                  </div>
                </div>
                {index < flows.slice(0, 5).length - 1 && <Separator className="mt-4" />}
              </div>
            ))}
            {flows.length === 0 && (
              <div className="py-12 text-center">
                <Workflow className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
                <p className="text-muted-foreground mb-4">
                  No workflows created yet
                </p>
                <Button asChild>
                  <Link to="/editor">
                    <Plus className="mr-2 h-4 w-4" />
                    Create your first workflow
                  </Link>
                </Button>
              </div>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
