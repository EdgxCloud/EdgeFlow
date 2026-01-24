/**
 * Resource Monitoring API Service
 * Handles system resource stats
 */

import { api } from './api'

// Resource Stats Types
export interface CPUStats {
  usage_percent: number
  cores: number
  model?: string
}

export interface MemoryStats {
  total_bytes: number
  used_bytes: number
  free_bytes: number
  usage_percent: number
}

export interface DiskStats {
  total_bytes: number
  used_bytes: number
  free_bytes: number
  usage_percent: number
  path?: string
}

export interface NetworkStats {
  bytes_sent: number
  bytes_recv: number
  packets_sent: number
  packets_recv: number
  errors_in: number
  errors_out: number
}

export interface ResourceStats {
  timestamp: string
  cpu: CPUStats
  memory: MemoryStats
  disk: DiskStats
  network?: NetworkStats
  uptime_seconds?: number
  goroutines?: number
}

// Resources API Methods
export const resourcesApi = {
  /**
   * Get current resource stats
   */
  getStats: async (): Promise<ResourceStats> => {
    return api.get<ResourceStats>('/resources/stats')
  },

  /**
   * Format bytes to human-readable string
   */
  formatBytes: (bytes: number): string => {
    if (bytes === 0) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return `${(bytes / Math.pow(k, i)).toFixed(2)} ${sizes[i]}`
  },

  /**
   * Format uptime to human-readable string
   */
  formatUptime: (seconds: number): string => {
    const days = Math.floor(seconds / 86400)
    const hours = Math.floor((seconds % 86400) / 3600)
    const minutes = Math.floor((seconds % 3600) / 60)

    const parts = []
    if (days > 0) parts.push(`${days}d`)
    if (hours > 0) parts.push(`${hours}h`)
    if (minutes > 0) parts.push(`${minutes}m`)

    return parts.length > 0 ? parts.join(' ') : '< 1m'
  },
}

export default resourcesApi
