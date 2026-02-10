/**
 * Network & System Info API Service
 * Handles fetching real network interface data, WiFi scanning, and system info
 */

import { api } from './api'

export interface NetworkInterface {
  name: string
  mac: string
  mtu: number
  status: 'up' | 'down'
  ipv4: string[]
  ipv6: string[]
}

export interface NetworkInfo {
  hostname: string
  interfaces: NetworkInterface[]
  timestamp: string
}

export interface WifiNetwork {
  ssid: string
  bssid: string
  signal: number
  frequency: number
  channel: number
  security: string
  connected: boolean
}

export interface WifiScanResult {
  networks: WifiNetwork[]
  count: number
  platform: string
  error?: string
  timestamp: string
}

export interface SystemInfo {
  hostname: string
  os: string
  arch: string
  board_model: string
  uptime: number
  uptime_str: string
  temperature: number
  cpu: {
    cores: number
    usage_percent: number
  }
  memory: {
    total_bytes: number
    used_bytes: number
    free_bytes: number
    percent: number
  }
  disk: {
    total_bytes: number
    used_bytes: number
    free_bytes: number
    percent: number
  }
  network: {
    rx_bytes: number
    tx_bytes: number
  }
  load_avg: {
    '1min': number
    '5min': number
    '15min': number
  }
  go_version: string
  goroutines: number
  platform: string
  timestamp: string
}

export interface SettingsResponse {
  configured: boolean
  settings: Record<string, unknown>
}

export const networkApi = {
  getInfo: async (): Promise<NetworkInfo> => {
    return api.get<NetworkInfo>('/system/network')
  },
}

export const wifiApi = {
  scan: async (): Promise<WifiScanResult> => {
    return api.get<WifiScanResult>('/system/wifi/scan')
  },
  connect: async (ssid: string, password: string): Promise<{ message: string }> => {
    return api.post<{ message: string }>('/system/wifi/connect', { ssid, password })
  },
}

export const systemApi = {
  getInfo: async (): Promise<SystemInfo> => {
    return api.get<SystemInfo>('/system/info')
  },
  reboot: async (): Promise<{ message: string }> => {
    return api.post<{ message: string }>('/system/reboot')
  },
  restartService: async (): Promise<{ message: string }> => {
    return api.post<{ message: string }>('/system/restart-service')
  },
}

export const settingsApi = {
  get: async (): Promise<SettingsResponse> => {
    return api.get<SettingsResponse>('/settings')
  },
  save: async (settings: Record<string, unknown>): Promise<{ message: string; settings: Record<string, unknown> }> => {
    return api.put<{ message: string; settings: Record<string, unknown> }>('/settings', settings)
  },
}

export default networkApi
