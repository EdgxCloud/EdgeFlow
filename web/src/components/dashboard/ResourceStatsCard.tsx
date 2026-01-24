/**
 * Resource Stats Card Component
 * Displays system resource usage (CPU, Memory, Disk)
 */

import { useEffect, useState } from 'react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Progress } from '@/components/ui/progress'
import { Cpu, HardDrive, MemoryStick, Activity, Loader2 } from 'lucide-react'
import { resourcesApi, ResourceStats } from '@/services/resources'
import { toast } from 'sonner'

interface ResourceStatsCardProps {
  autoRefresh?: boolean
  refreshInterval?: number
}

export default function ResourceStatsCard({
  autoRefresh = true,
  refreshInterval = 5000,
}: ResourceStatsCardProps) {
  const [stats, setStats] = useState<ResourceStats | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const fetchStats = async () => {
      try {
        const data = await resourcesApi.getStats()
        setStats(data)
        setError(null)
      } catch (err) {
        const errorMessage = err instanceof Error ? err.message : 'Failed to fetch stats'
        setError(errorMessage)
        console.error('❌ Failed to fetch resource stats:', err)
        if (isLoading) {
          // Only show toast on initial load failure
          toast.error('Failed to load resource stats', {
            description: errorMessage,
          })
        }
      } finally {
        setIsLoading(false)
      }
    }

    // Initial fetch
    fetchStats()

    // Auto-refresh
    if (autoRefresh) {
      const interval = setInterval(fetchStats, refreshInterval)
      return () => clearInterval(interval)
    }
  }, [autoRefresh, refreshInterval, isLoading])

  if (isLoading) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>System Resources</CardTitle>
          <CardDescription>Real-time resource monitoring</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="flex items-center justify-center py-8">
            <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
          </div>
        </CardContent>
      </Card>
    )
  }

  if (error || !stats) {
    return (
      <Card>
        <CardHeader>
          <CardTitle>System Resources</CardTitle>
          <CardDescription>Real-time resource monitoring</CardDescription>
        </CardHeader>
        <CardContent>
          <div className="text-center py-8 text-muted-foreground">
            <Activity className="h-8 w-8 mx-auto mb-2" />
            <p className="text-sm">{error || 'No data available'}</p>
          </div>
        </CardContent>
      </Card>
    )
  }

  const getUsageColor = (percentage: number): string => {
    if (percentage >= 90) return 'bg-red-500'
    if (percentage >= 75) return 'bg-orange-500'
    if (percentage >= 50) return 'bg-yellow-500'
    return 'bg-green-500'
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center justify-between">
          <span>System Resources</span>
          {stats.uptime_seconds && (
            <span className="text-sm font-normal text-muted-foreground">
              Uptime: {resourcesApi.formatUptime(stats.uptime_seconds)}
            </span>
          )}
        </CardTitle>
        <CardDescription>Real-time resource monitoring</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="space-y-6">
          {/* CPU Usage */}
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <Cpu className="h-4 w-4 text-blue-600" />
                <span className="text-sm font-medium">CPU Usage</span>
              </div>
              <span className="text-sm text-muted-foreground">
                {stats.cpu.usage_percent.toFixed(1)}%
                {stats.cpu.cores && ` (${stats.cpu.cores} cores)`}
              </span>
            </div>
            <Progress
              value={stats.cpu.usage_percent}
              className="h-2"
              indicatorClassName={getUsageColor(stats.cpu.usage_percent)}
            />
          </div>

          {/* Memory Usage */}
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <MemoryStick className="h-4 w-4 text-purple-600" />
                <span className="text-sm font-medium">Memory Usage</span>
              </div>
              <span className="text-sm text-muted-foreground">
                {resourcesApi.formatBytes(stats.memory.used_bytes)} /{' '}
                {resourcesApi.formatBytes(stats.memory.total_bytes)}
              </span>
            </div>
            <Progress
              value={stats.memory.usage_percent}
              className="h-2"
              indicatorClassName={getUsageColor(stats.memory.usage_percent)}
            />
            <p className="text-xs text-muted-foreground">
              {stats.memory.usage_percent.toFixed(1)}% used
            </p>
          </div>

          {/* Disk Usage */}
          <div className="space-y-2">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                <HardDrive className="h-4 w-4 text-orange-600" />
                <span className="text-sm font-medium">Disk Usage</span>
              </div>
              <span className="text-sm text-muted-foreground">
                {resourcesApi.formatBytes(stats.disk.used_bytes)} /{' '}
                {resourcesApi.formatBytes(stats.disk.total_bytes)}
              </span>
            </div>
            <Progress
              value={stats.disk.usage_percent}
              className="h-2"
              indicatorClassName={getUsageColor(stats.disk.usage_percent)}
            />
            <p className="text-xs text-muted-foreground">
              {stats.disk.usage_percent.toFixed(1)}% used
              {stats.disk.path && ` (${stats.disk.path})`}
            </p>
          </div>

          {/* Network Stats (if available) */}
          {stats.network && (
            <div className="pt-4 border-t">
              <h4 className="text-sm font-medium mb-2">Network</h4>
              <div className="grid grid-cols-2 gap-4 text-sm">
                <div>
                  <p className="text-muted-foreground">Sent</p>
                  <p className="font-medium">
                    {resourcesApi.formatBytes(stats.network.bytes_sent)}
                  </p>
                </div>
                <div>
                  <p className="text-muted-foreground">Received</p>
                  <p className="font-medium">
                    {resourcesApi.formatBytes(stats.network.bytes_recv)}
                  </p>
                </div>
              </div>
            </div>
          )}

          {/* Goroutines (Backend specific) */}
          {stats.goroutines && (
            <div className="pt-2 text-xs text-muted-foreground">
              Active goroutines: {stats.goroutines}
            </div>
          )}
        </div>
      </CardContent>
    </Card>
  )
}
