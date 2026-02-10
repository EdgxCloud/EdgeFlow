/**
 * Monitoring Panel
 * Real-time hardware monitoring dashboard for Raspberry Pi
 */

import { useState, useEffect, useRef, useCallback } from 'react'
import { api } from '@/services/api'
import {
  Activity,
  Cpu,
  HardDrive,
  MemoryStick,
  Thermometer,
  Clock,
  Wifi,
  ArrowDown,
  ArrowUp,
  Server,
  Gauge,
} from 'lucide-react'
import { cn } from '@/lib/utils'

interface SystemInfo {
  hostname: string
  os: string
  arch: string
  board_model: string
  uptime: number
  temperature: number
  load_avg: { '1min': number; '5min': number; '15min': number }
  swap: { total_bytes: number; used_bytes: number }
  network: { rx_bytes: number; tx_bytes: number }
}

interface ResourceStats {
  cpu: { usage_percent: number; cores: number }
  memory: { total_bytes: number; used_bytes: number; free_bytes: number; usage_percent: number }
  disk: { total_bytes: number; used_bytes: number; free_bytes: number; usage_percent: number }
  goroutines: number
  system?: SystemInfo
}

function formatBytes(bytes: number): string {
  if (!bytes || bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
}

function formatUptime(seconds: number): string {
  if (!seconds) return '--'
  const days = Math.floor(seconds / 86400)
  const hours = Math.floor((seconds % 86400) / 3600)
  const mins = Math.floor((seconds % 3600) / 60)
  if (days > 0) return `${days}d ${hours}h ${mins}m`
  if (hours > 0) return `${hours}h ${mins}m`
  return `${mins}m`
}

function getTempColor(temp: number): string {
  if (temp >= 80) return 'text-red-500'
  if (temp >= 70) return 'text-orange-500'
  if (temp >= 60) return 'text-yellow-500'
  return 'text-green-500'
}

function getTempBg(temp: number): string {
  if (temp >= 80) return 'bg-red-500'
  if (temp >= 70) return 'bg-orange-500'
  if (temp >= 60) return 'bg-yellow-500'
  return 'bg-green-500'
}

function getUsageColor(percent: number): string {
  if (percent >= 90) return 'bg-red-500'
  if (percent >= 75) return 'bg-orange-500'
  if (percent >= 50) return 'bg-yellow-500'
  return 'bg-emerald-500'
}

// Circular gauge component
function CircularGauge({ value, max = 100, size = 80, strokeWidth = 6, color, label, sublabel }: {
  value: number
  max?: number
  size?: number
  strokeWidth?: number
  color: string
  label: string
  sublabel?: string
}) {
  const radius = (size - strokeWidth) / 2
  const circumference = radius * 2 * Math.PI
  const percent = Math.min(100, (value / max) * 100)
  const offset = circumference - (percent / 100) * circumference

  return (
    <div className="flex flex-col items-center">
      <div className="relative" style={{ width: size, height: size }}>
        <svg width={size} height={size} className="-rotate-90">
          <circle
            cx={size / 2}
            cy={size / 2}
            r={radius}
            fill="none"
            stroke="currentColor"
            strokeWidth={strokeWidth}
            className="text-muted/20"
          />
          <circle
            cx={size / 2}
            cy={size / 2}
            r={radius}
            fill="none"
            stroke={color}
            strokeWidth={strokeWidth}
            strokeDasharray={circumference}
            strokeDashoffset={offset}
            strokeLinecap="round"
            className="transition-all duration-700 ease-out"
          />
        </svg>
        <div className="absolute inset-0 flex items-center justify-center">
          <span className="text-sm font-bold">{value.toFixed(1)}%</span>
        </div>
      </div>
      <span className="text-[11px] font-medium mt-1">{label}</span>
      {sublabel && <span className="text-[10px] text-muted-foreground">{sublabel}</span>}
    </div>
  )
}

export function MonitoringPanel() {
  const [stats, setStats] = useState<ResourceStats | null>(null)
  const [prevNet, setPrevNet] = useState<{ rx: number; tx: number; time: number } | null>(null)
  const [netSpeed, setNetSpeed] = useState<{ rx: number; tx: number }>({ rx: 0, tx: 0 })
  const [error, setError] = useState<string | null>(null)
  const intervalRef = useRef<NodeJS.Timeout | null>(null)

  const fetchStats = useCallback(async () => {
    try {
      const data = await api.get<ResourceStats>('/resources/stats')
      setStats(data)
      setError(null)

      // Calculate network speed
      const net = data.system?.network
      if (net) {
        setPrevNet(prev => {
          if (prev) {
            const elapsed = (Date.now() - prev.time) / 1000
            if (elapsed > 0) {
              setNetSpeed({
                rx: Math.max(0, (net.rx_bytes - prev.rx) / elapsed),
                tx: Math.max(0, (net.tx_bytes - prev.tx) / elapsed),
              })
            }
          }
          return { rx: net.rx_bytes, tx: net.tx_bytes, time: Date.now() }
        })
      }
    } catch {
      setError('Failed to fetch stats')
    }
  }, [])

  useEffect(() => {
    fetchStats()
    intervalRef.current = setInterval(fetchStats, 3000)
    return () => {
      if (intervalRef.current) clearInterval(intervalRef.current)
    }
  }, [fetchStats])

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

  const sys = stats?.system
  const cpuPercent = stats?.cpu?.usage_percent ?? 0
  const cpuCores = stats?.cpu?.cores ?? 0
  const memPercent = stats?.memory?.usage_percent ?? 0
  const memUsed = stats?.memory?.used_bytes ?? 0
  const memTotal = stats?.memory?.total_bytes ?? 0
  const diskPercent = stats?.disk?.usage_percent ?? 0
  const diskUsed = stats?.disk?.used_bytes ?? 0
  const diskTotal = stats?.disk?.total_bytes ?? 0
  const goroutines = stats?.goroutines ?? 0
  const temp = sys?.temperature ?? 0
  const uptime = sys?.uptime ?? 0
  const swapTotal = sys?.swap?.total_bytes ?? 0
  const swapUsed = sys?.swap?.used_bytes ?? 0

  return (
    <div className="h-full overflow-auto">
      <div className="p-3 space-y-3">
        {/* Row 1: Board Info Bar */}
        <div className="flex items-center justify-between px-3 py-2 bg-muted/30 rounded-lg border border-border">
          <div className="flex items-center gap-3">
            <Server className="w-4 h-4 text-primary" />
            <div>
              <span className="text-xs font-semibold">{sys?.board_model || 'EdgeFlow Device'}</span>
              <span className="text-[10px] text-muted-foreground ml-2">
                {sys?.hostname || '--'}
              </span>
            </div>
          </div>
          <div className="flex items-center gap-4 text-[11px] text-muted-foreground">
            <div className="flex items-center gap-1">
              <Clock className="w-3 h-3" />
              <span>Uptime: {formatUptime(uptime)}</span>
            </div>
            <div className="flex items-center gap-1">
              <span>{sys?.os || '--'}/{sys?.arch || '--'}</span>
            </div>
          </div>
        </div>

        {/* Row 2: Main gauges */}
        <div className="grid grid-cols-6 gap-3">
          {/* CPU */}
          <div className="col-span-1 p-3 rounded-lg border border-border bg-card flex flex-col items-center">
            <CircularGauge
              value={cpuPercent}
              color={cpuPercent >= 90 ? '#ef4444' : cpuPercent >= 70 ? '#f97316' : '#3b82f6'}
              label="CPU"
              sublabel={`${cpuCores} cores`}
            />
          </div>

          {/* Memory */}
          <div className="col-span-1 p-3 rounded-lg border border-border bg-card flex flex-col items-center">
            <CircularGauge
              value={memPercent}
              color={memPercent >= 90 ? '#ef4444' : memPercent >= 75 ? '#f97316' : '#22c55e'}
              label="Memory"
              sublabel={`${formatBytes(memUsed)} / ${formatBytes(memTotal)}`}
            />
          </div>

          {/* Disk */}
          <div className="col-span-1 p-3 rounded-lg border border-border bg-card flex flex-col items-center">
            <CircularGauge
              value={diskPercent}
              color={diskPercent >= 90 ? '#ef4444' : diskPercent >= 75 ? '#f97316' : '#8b5cf6'}
              label="Disk"
              sublabel={`${formatBytes(diskUsed)} / ${formatBytes(diskTotal)}`}
            />
          </div>

          {/* Temperature */}
          <div className="col-span-1 p-3 rounded-lg border border-border bg-card flex flex-col items-center justify-center">
            <Thermometer className={cn('w-5 h-5 mb-1', getTempColor(temp))} />
            <span className={cn('text-lg font-bold', getTempColor(temp))}>
              {temp > 0 ? `${temp.toFixed(1)}°` : '--'}
            </span>
            <span className="text-[11px] font-medium mt-0.5">Temp</span>
            {temp > 0 && (
              <div className="w-full mt-1.5 h-1 bg-muted rounded-full overflow-hidden">
                <div
                  className={cn('h-full rounded-full transition-all duration-500', getTempBg(temp))}
                  style={{ width: `${Math.min(100, (temp / 85) * 100)}%` }}
                />
              </div>
            )}
          </div>

          {/* Network */}
          <div className="col-span-1 p-3 rounded-lg border border-border bg-card">
            <div className="flex items-center justify-center gap-1 mb-1.5">
              <Wifi className="w-3.5 h-3.5 text-cyan-500" />
              <span className="text-[11px] font-medium">Network</span>
            </div>
            <div className="space-y-1.5">
              <div className="flex items-center gap-1 text-[11px]">
                <ArrowDown className="w-3 h-3 text-green-500" />
                <span className="text-muted-foreground">RX</span>
                <span className="ml-auto font-mono text-[10px]">{formatBytes(netSpeed.rx)}/s</span>
              </div>
              <div className="flex items-center gap-1 text-[11px]">
                <ArrowUp className="w-3 h-3 text-blue-500" />
                <span className="text-muted-foreground">TX</span>
                <span className="ml-auto font-mono text-[10px]">{formatBytes(netSpeed.tx)}/s</span>
              </div>
            </div>
            <div className="text-[9px] text-muted-foreground/60 mt-1.5 text-center">
              Total: {formatBytes(sys?.network?.rx_bytes ?? 0)} / {formatBytes(sys?.network?.tx_bytes ?? 0)}
            </div>
          </div>

          {/* System Info */}
          <div className="col-span-1 p-3 rounded-lg border border-border bg-card">
            <div className="flex items-center justify-center gap-1 mb-1.5">
              <Gauge className="w-3.5 h-3.5 text-orange-500" />
              <span className="text-[11px] font-medium">System</span>
            </div>
            <div className="space-y-1 text-[11px]">
              <div className="flex justify-between">
                <span className="text-muted-foreground">Goroutines</span>
                <span className="font-mono font-medium">{goroutines}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Load 1m</span>
                <span className="font-mono font-medium">{sys?.load_avg?.['1min']?.toFixed(2) ?? '--'}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Load 5m</span>
                <span className="font-mono font-medium">{sys?.load_avg?.['5min']?.toFixed(2) ?? '--'}</span>
              </div>
            </div>
          </div>
        </div>

        {/* Row 3: Detailed bars */}
        <div className="grid grid-cols-3 gap-3">
          {/* Memory Detail */}
          <div className="p-3 rounded-lg border border-border bg-card">
            <div className="flex items-center gap-1.5 mb-2">
              <MemoryStick className="w-3.5 h-3.5 text-emerald-500" />
              <span className="text-xs font-medium">Memory Details</span>
            </div>
            <div className="space-y-2">
              <div>
                <div className="flex justify-between text-[11px] mb-0.5">
                  <span className="text-muted-foreground">RAM Usage</span>
                  <span className="font-mono">{formatBytes(memUsed)} / {formatBytes(memTotal)}</span>
                </div>
                <div className="h-2 bg-muted rounded-full overflow-hidden">
                  <div
                    className={cn('h-full rounded-full transition-all duration-500', getUsageColor(memPercent))}
                    style={{ width: `${Math.min(100, memPercent)}%` }}
                  />
                </div>
              </div>
              {swapTotal > 0 && (
                <div>
                  <div className="flex justify-between text-[11px] mb-0.5">
                    <span className="text-muted-foreground">Swap</span>
                    <span className="font-mono">{formatBytes(swapUsed)} / {formatBytes(swapTotal)}</span>
                  </div>
                  <div className="h-2 bg-muted rounded-full overflow-hidden">
                    <div
                      className="h-full bg-amber-500 rounded-full transition-all duration-500"
                      style={{ width: `${swapTotal > 0 ? Math.min(100, (swapUsed / swapTotal) * 100) : 0}%` }}
                    />
                  </div>
                </div>
              )}
            </div>
          </div>

          {/* Disk Detail */}
          <div className="p-3 rounded-lg border border-border bg-card">
            <div className="flex items-center gap-1.5 mb-2">
              <HardDrive className="w-3.5 h-3.5 text-violet-500" />
              <span className="text-xs font-medium">Disk Details</span>
            </div>
            <div className="space-y-2">
              <div>
                <div className="flex justify-between text-[11px] mb-0.5">
                  <span className="text-muted-foreground">Storage</span>
                  <span className="font-mono">{formatBytes(diskUsed)} / {formatBytes(diskTotal)}</span>
                </div>
                <div className="h-2 bg-muted rounded-full overflow-hidden">
                  <div
                    className={cn('h-full rounded-full transition-all duration-500', getUsageColor(diskPercent))}
                    style={{ width: `${Math.min(100, diskPercent)}%` }}
                  />
                </div>
              </div>
              <div className="flex justify-between text-[11px]">
                <span className="text-muted-foreground">Free Space</span>
                <span className="font-mono">{formatBytes(stats?.disk?.free_bytes ?? 0)}</span>
              </div>
            </div>
          </div>

          {/* CPU Detail */}
          <div className="p-3 rounded-lg border border-border bg-card">
            <div className="flex items-center gap-1.5 mb-2">
              <Cpu className="w-3.5 h-3.5 text-blue-500" />
              <span className="text-xs font-medium">CPU Details</span>
            </div>
            <div className="space-y-2">
              <div>
                <div className="flex justify-between text-[11px] mb-0.5">
                  <span className="text-muted-foreground">Usage</span>
                  <span className="font-mono">{cpuPercent.toFixed(1)}%</span>
                </div>
                <div className="h-2 bg-muted rounded-full overflow-hidden">
                  <div
                    className={cn('h-full rounded-full transition-all duration-500', getUsageColor(cpuPercent))}
                    style={{ width: `${Math.min(100, cpuPercent)}%` }}
                  />
                </div>
              </div>
              <div className="flex justify-between text-[11px]">
                <span className="text-muted-foreground">Cores</span>
                <span className="font-mono">{cpuCores}</span>
              </div>
              <div className="flex justify-between text-[11px]">
                <span className="text-muted-foreground">Load (15m)</span>
                <span className="font-mono">{sys?.load_avg?.['15min']?.toFixed(2) ?? '--'}</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
