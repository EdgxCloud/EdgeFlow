/**
 * Monitoring Panel
 * Real-time resource stats from backend API
 */

import { useState, useEffect, useRef } from 'react'
import { api } from '@/services/api'
import { Activity, Cpu, HardDrive, MemoryStick } from 'lucide-react'

interface ResourceStats {
  cpu: { usage_percent: number; cores: number }
  memory: { total_bytes: number; used_bytes: number; free_bytes: number; usage_percent: number }
  disk: { total_bytes: number; used_bytes: number; free_bytes: number; usage_percent: number }
  goroutines: number
}

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

export function MonitoringPanel() {
  const [stats, setStats] = useState<ResourceStats | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [lastUpdate, setLastUpdate] = useState<Date | null>(null)
  const intervalRef = useRef<NodeJS.Timeout | null>(null)

  const fetchStats = async () => {
    try {
      const response = await api.get<ResourceStats>('/api/v1/resources/stats')
      setStats(response.data)
      setLastUpdate(new Date())
      setError(null)
    } catch {
      setError('Failed to fetch stats')
    }
  }

  useEffect(() => {
    fetchStats()
    intervalRef.current = setInterval(fetchStats, 3000)
    return () => {
      if (intervalRef.current) clearInterval(intervalRef.current)
    }
  }, [])

  if (error && !stats) {
    return (
      <div className="h-full flex items-center justify-center text-muted-foreground text-sm">
        <div className="text-center">
          <Activity className="w-8 h-8 mx-auto mb-2 opacity-50" />
          <p>{error}</p>
          <p className="text-xs mt-1">Backend may not be running</p>
        </div>
      </div>
    )
  }

  const cpuPercent = stats?.cpu?.usage_percent ?? 0
  const memPercent = stats?.memory?.usage_percent ?? 0
  const diskPercent = stats?.disk?.usage_percent ?? 0
  const memUsed = stats?.memory?.used_bytes ?? 0
  const memTotal = stats?.memory?.total_bytes ?? 0
  const diskUsed = stats?.disk?.used_bytes ?? 0
  const diskTotal = stats?.disk?.total_bytes ?? 0
  const goroutines = stats?.goroutines ?? 0
  const cpuCores = stats?.cpu?.cores ?? 0

  return (
    <div className="h-full p-4 overflow-auto">
      <div className="grid grid-cols-4 gap-4">
        {/* CPU */}
        <div className="p-3 bg-blue-50 dark:bg-blue-950/30 rounded-lg border border-blue-200 dark:border-blue-900">
          <div className="flex items-center gap-1.5 mb-2">
            <Cpu className="w-3.5 h-3.5 text-blue-500" />
            <span className="text-xs text-blue-600 dark:text-blue-400">CPU</span>
          </div>
          <div className="text-2xl font-bold text-blue-700 dark:text-blue-300">
            {cpuPercent.toFixed(1)}%
          </div>
          <div className="text-xs text-blue-500/70 mt-1">{cpuCores} cores</div>
          <div className="mt-2 h-1.5 bg-blue-200 dark:bg-blue-900 rounded-full overflow-hidden">
            <div
              className="h-full bg-blue-500 rounded-full transition-all duration-500"
              style={{ width: `${Math.min(100, cpuPercent)}%` }}
            />
          </div>
        </div>

        {/* Memory */}
        <div className="p-3 bg-green-50 dark:bg-green-950/30 rounded-lg border border-green-200 dark:border-green-900">
          <div className="flex items-center gap-1.5 mb-2">
            <MemoryStick className="w-3.5 h-3.5 text-green-500" />
            <span className="text-xs text-green-600 dark:text-green-400">Memory</span>
          </div>
          <div className="text-2xl font-bold text-green-700 dark:text-green-300">
            {memPercent.toFixed(1)}%
          </div>
          <div className="text-xs text-green-500/70 mt-1">
            {formatBytes(memUsed)} / {formatBytes(memTotal)}
          </div>
          <div className="mt-2 h-1.5 bg-green-200 dark:bg-green-900 rounded-full overflow-hidden">
            <div
              className="h-full bg-green-500 rounded-full transition-all duration-500"
              style={{ width: `${Math.min(100, memPercent)}%` }}
            />
          </div>
        </div>

        {/* Disk */}
        <div className="p-3 bg-purple-50 dark:bg-purple-950/30 rounded-lg border border-purple-200 dark:border-purple-900">
          <div className="flex items-center gap-1.5 mb-2">
            <HardDrive className="w-3.5 h-3.5 text-purple-500" />
            <span className="text-xs text-purple-600 dark:text-purple-400">Disk</span>
          </div>
          <div className="text-2xl font-bold text-purple-700 dark:text-purple-300">
            {diskPercent.toFixed(1)}%
          </div>
          <div className="text-xs text-purple-500/70 mt-1">
            {formatBytes(diskUsed)} / {formatBytes(diskTotal)}
          </div>
          <div className="mt-2 h-1.5 bg-purple-200 dark:bg-purple-900 rounded-full overflow-hidden">
            <div
              className="h-full bg-purple-500 rounded-full transition-all duration-500"
              style={{ width: `${Math.min(100, diskPercent)}%` }}
            />
          </div>
        </div>

        {/* Goroutines */}
        <div className="p-3 bg-orange-50 dark:bg-orange-950/30 rounded-lg border border-orange-200 dark:border-orange-900">
          <div className="flex items-center gap-1.5 mb-2">
            <Activity className="w-3.5 h-3.5 text-orange-500" />
            <span className="text-xs text-orange-600 dark:text-orange-400">Goroutines</span>
          </div>
          <div className="text-2xl font-bold text-orange-700 dark:text-orange-300">
            {goroutines}
          </div>
          <div className="text-xs text-orange-500/70 mt-1">Active threads</div>
          {lastUpdate && (
            <div className="text-xs text-orange-500/50 mt-2">
              Updated {lastUpdate.toLocaleTimeString()}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
