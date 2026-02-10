/**
 * Network Info API Service
 * Handles fetching real network interface data
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

export const networkApi = {
  getInfo: async (): Promise<NetworkInfo> => {
    return api.get<NetworkInfo>('/system/network')
  },
}

export default networkApi
