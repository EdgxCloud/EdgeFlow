/**
 * Network Configuration Step
 *
 * Comprehensive network setup with WiFi scanning, static IP, DHCP, DNS configuration
 * Based on Raspberry Pi NetworkManager (nmcli) configuration patterns
 */

import { useState, useEffect, useCallback } from 'react'
import {
  Wifi,
  WifiOff,
  Globe,
  Server,
  Eye,
  EyeOff,
  RefreshCw,
  Signal,
  SignalLow,
  SignalMedium,
  SignalHigh,
  Lock,
  Unlock,
  Check,
  AlertTriangle,
  ChevronDown,
  ChevronUp,
  Network,
  Radio,
} from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import type {
  NetworkConfig,
  BoardType,
  WifiNetwork,
  WifiSecurityType,
  IPMethod,
} from '../types'
import {
  boardHasWifi,
  getBoardInfo,
  WIFI_COUNTRIES,
  COMMON_DNS_SERVERS,
  COMMON_SUBNET_MASKS,
} from '../types'

interface NetworkConfigStepProps {
  config: NetworkConfig
  boardType: BoardType | null
  onChange: (config: NetworkConfig) => void
}

// Simulated WiFi networks for demo/development
const DEMO_WIFI_NETWORKS: WifiNetwork[] = [
  {
    ssid: 'Home_Network_5G',
    bssid: 'AA:BB:CC:DD:EE:01',
    signal: -45,
    frequency: 5180,
    channel: 36,
    security: 'wpa2',
    connected: false,
  },
  {
    ssid: 'Home_Network_2.4G',
    bssid: 'AA:BB:CC:DD:EE:02',
    signal: -55,
    frequency: 2437,
    channel: 6,
    security: 'wpa2',
    connected: false,
  },
  {
    ssid: 'Guest_WiFi',
    bssid: 'AA:BB:CC:DD:EE:03',
    signal: -65,
    frequency: 2412,
    channel: 1,
    security: 'wpa',
    connected: false,
  },
  {
    ssid: 'Office_Network',
    bssid: 'AA:BB:CC:DD:EE:04',
    signal: -70,
    frequency: 5240,
    channel: 48,
    security: 'wpa2-enterprise',
    connected: false,
  },
  {
    ssid: 'IoT_Network',
    bssid: 'AA:BB:CC:DD:EE:05',
    signal: -60,
    frequency: 2462,
    channel: 11,
    security: 'wpa2',
    connected: false,
  },
  {
    ssid: 'Public_Hotspot',
    bssid: 'AA:BB:CC:DD:EE:06',
    signal: -80,
    frequency: 2437,
    channel: 6,
    security: 'open',
    connected: false,
  },
]

export function NetworkConfigStep({
  config,
  boardType,
  onChange,
}: NetworkConfigStepProps) {
  const [showPassword, setShowPassword] = useState(false)
  const [scanning, setScanning] = useState(false)
  const [wifiNetworks, setWifiNetworks] = useState<WifiNetwork[]>([])
  const [showAdvanced, setShowAdvanced] = useState(false)
  const [showDnsPresets, setShowDnsPresets] = useState(false)

  const boardInfo = getBoardInfo(boardType)
  const hasWifi = boardHasWifi(boardType)

  // Scan for WiFi networks
  const handleScanNetworks = useCallback(() => {
    setScanning(true)
    // Simulate network scan with random variations
    setTimeout(() => {
      const networks = DEMO_WIFI_NETWORKS.map((n) => ({
        ...n,
        signal: n.signal + Math.floor(Math.random() * 10) - 5,
      })).sort((a, b) => b.signal - a.signal)
      setWifiNetworks(networks)
      setScanning(false)
    }, 2000)
  }, [])

  // Auto-scan on mount if WiFi is enabled
  useEffect(() => {
    if (hasWifi && config.wifi.enabled && wifiNetworks.length === 0) {
      handleScanNetworks()
    }
  }, [hasWifi, config.wifi.enabled, wifiNetworks.length, handleScanNetworks])

  const updateConfig = (updates: Partial<NetworkConfig>) => {
    onChange({ ...config, ...updates })
  }

  const updateEthernet = (
    updates: Partial<NetworkConfig['ethernet']>
  ) => {
    onChange({
      ...config,
      ethernet: { ...config.ethernet, ...updates },
    })
  }

  const updateWifi = (updates: Partial<NetworkConfig['wifi']>) => {
    onChange({
      ...config,
      wifi: { ...config.wifi, ...updates },
      // Legacy compatibility
      ssid: updates.ssid ?? config.wifi.ssid,
      password: updates.password ?? config.wifi.password,
    })
  }

  const selectWifiNetwork = (network: WifiNetwork) => {
    updateWifi({
      ssid: network.ssid,
      security: network.security,
    })
  }

  const applyDnsPreset = (preset: (typeof COMMON_DNS_SERVERS)[0]) => {
    if (config.primaryInterface === 'wifi') {
      updateWifi({ dns1: preset.primary, dns2: preset.secondary })
    } else {
      updateEthernet({ dns1: preset.primary, dns2: preset.secondary })
    }
    setShowDnsPresets(false)
  }

  const getSignalIcon = (signal: number) => {
    if (signal >= -50) return SignalHigh
    if (signal >= -65) return SignalMedium
    if (signal >= -80) return SignalLow
    return Signal
  }

  const getSignalQuality = (signal: number) => {
    if (signal >= -50) return { text: 'Excellent', color: 'text-green-500' }
    if (signal >= -65) return { text: 'Good', color: 'text-blue-500' }
    if (signal >= -80) return { text: 'Fair', color: 'text-yellow-500' }
    return { text: 'Weak', color: 'text-red-500' }
  }

  const getSecurityLabel = (security: WifiSecurityType) => {
    const labels: Record<WifiSecurityType, string> = {
      open: 'Open',
      wep: 'WEP',
      wpa: 'WPA',
      wpa2: 'WPA2',
      wpa3: 'WPA3',
      'wpa2-enterprise': 'WPA2 Enterprise',
    }
    return labels[security]
  }

  const currentInterface =
    config.primaryInterface === 'wifi' ? config.wifi : config.ethernet

  return (
    <div className="space-y-6">
      <div className="text-center space-y-2">
        <h2 className="text-2xl font-bold text-gray-900 dark:text-white">
          Network Configuration
        </h2>
        <p className="text-muted-foreground">
          Configure how your device connects to the network
        </p>
      </div>

      {/* Board WiFi Info */}
      {boardInfo && (
        <div
          className={cn(
            'p-3 rounded-lg border flex items-center gap-3',
            hasWifi
              ? 'bg-green-50 dark:bg-green-950/30 border-green-200 dark:border-green-900'
              : 'bg-amber-50 dark:bg-amber-950/30 border-amber-200 dark:border-amber-900'
          )}
        >
          {hasWifi ? (
            <Wifi className="w-5 h-5 text-green-500" />
          ) : (
            <WifiOff className="w-5 h-5 text-amber-500" />
          )}
          <div className="flex-1">
            <p
              className={cn(
                'text-sm font-medium',
                hasWifi
                  ? 'text-green-800 dark:text-green-200'
                  : 'text-amber-800 dark:text-amber-200'
              )}
            >
              {hasWifi
                ? `WiFi Available: ${boardInfo.wifiChip || 'Built-in'}`
                : 'No built-in WiFi - USB adapter required'}
            </p>
            {boardInfo.ethernetSpeed && (
              <p className="text-xs text-muted-foreground">
                Ethernet: {boardInfo.ethernetSpeed}
              </p>
            )}
          </div>
        </div>
      )}

      {/* Connection Type Selection */}
      <div className="space-y-4">
        <Label className="text-sm font-semibold">Primary Connection</Label>
        <div className="grid grid-cols-2 gap-3">
          <button
            onClick={() => updateConfig({ primaryInterface: 'ethernet', useWifi: false })}
            className={cn(
              'p-4 rounded-xl border-2 transition-all flex items-center gap-3',
              config.primaryInterface === 'ethernet'
                ? 'border-primary bg-primary/5'
                : 'border-gray-200 dark:border-gray-700 hover:border-gray-300'
            )}
          >
            <div
              className={cn(
                'w-10 h-10 rounded-lg flex items-center justify-center',
                config.primaryInterface === 'ethernet'
                  ? 'bg-blue-100 dark:bg-blue-900/30'
                  : 'bg-gray-100 dark:bg-gray-800'
              )}
            >
              <Globe className="w-5 h-5 text-blue-500" />
            </div>
            <div className="text-left">
              <div className="font-medium">Ethernet</div>
              <div className="text-xs text-muted-foreground">
                Wired connection (eth0)
              </div>
            </div>
          </button>

          <button
            onClick={() => {
              if (hasWifi) {
                updateConfig({ primaryInterface: 'wifi', useWifi: true })
                updateWifi({ enabled: true })
              }
            }}
            disabled={!hasWifi}
            className={cn(
              'p-4 rounded-xl border-2 transition-all flex items-center gap-3',
              config.primaryInterface === 'wifi'
                ? 'border-primary bg-primary/5'
                : 'border-gray-200 dark:border-gray-700 hover:border-gray-300',
              !hasWifi && 'opacity-50 cursor-not-allowed'
            )}
          >
            <div
              className={cn(
                'w-10 h-10 rounded-lg flex items-center justify-center',
                config.primaryInterface === 'wifi'
                  ? 'bg-green-100 dark:bg-green-900/30'
                  : 'bg-gray-100 dark:bg-gray-800'
              )}
            >
              <Wifi className="w-5 h-5 text-green-500" />
            </div>
            <div className="text-left">
              <div className="font-medium">WiFi</div>
              <div className="text-xs text-muted-foreground">
                Wireless (wlan0)
              </div>
            </div>
          </button>
        </div>
      </div>

      {/* WiFi Network Selection */}
      {config.primaryInterface === 'wifi' && hasWifi && (
        <div className="space-y-4 p-4 bg-gray-50 dark:bg-gray-800/50 rounded-xl">
          <div className="flex items-center justify-between">
            <Label className="text-sm font-semibold flex items-center gap-2">
              <Radio className="w-4 h-4" />
              Available Networks
            </Label>
            <Button
              variant="outline"
              size="sm"
              onClick={handleScanNetworks}
              disabled={scanning}
            >
              <RefreshCw
                className={cn('w-4 h-4 mr-2', scanning && 'animate-spin')}
              />
              {scanning ? 'Scanning...' : 'Scan'}
            </Button>
          </div>

          {/* WiFi Networks List */}
          {wifiNetworks.length > 0 ? (
            <div className="space-y-2 max-h-64 overflow-y-auto">
              {wifiNetworks.map((network) => {
                const SignalIcon = getSignalIcon(network.signal)
                const signalQuality = getSignalQuality(network.signal)
                const isSelected = config.wifi.ssid === network.ssid

                return (
                  <button
                    key={network.bssid}
                    onClick={() => selectWifiNetwork(network)}
                    className={cn(
                      'w-full p-3 rounded-lg border transition-all flex items-center gap-3 text-left',
                      isSelected
                        ? 'border-primary bg-primary/5'
                        : 'border-gray-200 dark:border-gray-700 hover:border-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700'
                    )}
                  >
                    <SignalIcon className={cn('w-5 h-5', signalQuality.color)} />
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2">
                        <span className="font-medium truncate">
                          {network.ssid}
                        </span>
                        {network.security !== 'open' ? (
                          <Lock className="w-3 h-3 text-muted-foreground" />
                        ) : (
                          <Unlock className="w-3 h-3 text-amber-500" />
                        )}
                      </div>
                      <div className="flex items-center gap-2 text-xs text-muted-foreground">
                        <span>{getSecurityLabel(network.security)}</span>
                        <span>•</span>
                        <span>
                          {network.frequency > 5000 ? '5GHz' : '2.4GHz'}
                        </span>
                        <span>•</span>
                        <span>Ch {network.channel}</span>
                      </div>
                    </div>
                    <div className="text-right">
                      <div className={cn('text-xs font-medium', signalQuality.color)}>
                        {signalQuality.text}
                      </div>
                      <div className="text-xs text-muted-foreground">
                        {network.signal} dBm
                      </div>
                    </div>
                    {isSelected && (
                      <Check className="w-5 h-5 text-primary flex-shrink-0" />
                    )}
                  </button>
                )
              })}
            </div>
          ) : scanning ? (
            <div className="text-center py-8">
              <RefreshCw className="w-8 h-8 mx-auto text-muted-foreground animate-spin mb-2" />
              <p className="text-sm text-muted-foreground">
                Scanning for networks...
              </p>
            </div>
          ) : (
            <div className="text-center py-8">
              <WifiOff className="w-8 h-8 mx-auto text-muted-foreground mb-2" />
              <p className="text-sm text-muted-foreground">
                No networks found. Click Scan to search.
              </p>
            </div>
          )}

          {/* Manual SSID Entry */}
          <div className="pt-3 border-t border-gray-200 dark:border-gray-700">
            <div className="flex items-center gap-2 mb-3">
              <Switch
                checked={config.wifi.hidden || false}
                onCheckedChange={(checked) => updateWifi({ hidden: checked })}
              />
              <Label className="text-sm">Connect to hidden network</Label>
            </div>

            <div className="space-y-3">
              <div className="space-y-2">
                <Label htmlFor="ssid">Network Name (SSID)</Label>
                <Input
                  id="ssid"
                  placeholder="Enter WiFi network name"
                  value={config.wifi.ssid || ''}
                  onChange={(e) => updateWifi({ ssid: e.target.value })}
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="wifi-password">Password</Label>
                <div className="relative">
                  <Input
                    id="wifi-password"
                    type={showPassword ? 'text' : 'password'}
                    placeholder="Enter WiFi password"
                    value={config.wifi.password || ''}
                    onChange={(e) => updateWifi({ password: e.target.value })}
                    className="pr-10"
                  />
                  <button
                    type="button"
                    onClick={() => setShowPassword(!showPassword)}
                    className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600"
                  >
                    {showPassword ? (
                      <EyeOff className="w-4 h-4" />
                    ) : (
                      <Eye className="w-4 h-4" />
                    )}
                  </button>
                </div>
              </div>

              {/* WiFi Country */}
              <div className="space-y-2">
                <Label htmlFor="wifi-country">WiFi Country/Region</Label>
                <select
                  id="wifi-country"
                  value={config.wifi.country || 'US'}
                  onChange={(e) => updateWifi({ country: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary"
                >
                  {WIFI_COUNTRIES.map((country) => (
                    <option key={country.code} value={country.code}>
                      {country.name} ({country.code})
                    </option>
                  ))}
                </select>
                <p className="text-xs text-muted-foreground">
                  Required for correct 5GHz frequency bands
                </p>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Hostname */}
      <div className="space-y-2">
        <Label htmlFor="hostname">Device Hostname</Label>
        <div className="relative">
          <Server className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
          <Input
            id="hostname"
            placeholder="edgeflow-device"
            value={config.hostname}
            onChange={(e) => updateConfig({ hostname: e.target.value })}
            className="pl-10"
          />
        </div>
        <p className="text-xs text-muted-foreground">
          Access via {config.hostname}.local (mDNS/Avahi)
        </p>
      </div>

      {/* IP Configuration Method */}
      <div className="space-y-4">
        <Label className="text-sm font-semibold">IP Configuration</Label>
        <div className="grid grid-cols-2 gap-3">
          <button
            onClick={() => {
              if (config.primaryInterface === 'wifi') {
                updateWifi({ ipMethod: 'dhcp' })
              } else {
                updateEthernet({ ipMethod: 'dhcp' })
              }
              updateConfig({ useStaticIP: false })
            }}
            className={cn(
              'p-3 rounded-lg border-2 transition-all text-left',
              currentInterface.ipMethod === 'dhcp'
                ? 'border-primary bg-primary/5'
                : 'border-gray-200 dark:border-gray-700 hover:border-gray-300'
            )}
          >
            <div className="font-medium">DHCP (Automatic)</div>
            <div className="text-xs text-muted-foreground">
              Get IP from router automatically
            </div>
          </button>

          <button
            onClick={() => {
              if (config.primaryInterface === 'wifi') {
                updateWifi({ ipMethod: 'static' })
              } else {
                updateEthernet({ ipMethod: 'static' })
              }
              updateConfig({ useStaticIP: true })
            }}
            className={cn(
              'p-3 rounded-lg border-2 transition-all text-left',
              currentInterface.ipMethod === 'static'
                ? 'border-primary bg-primary/5'
                : 'border-gray-200 dark:border-gray-700 hover:border-gray-300'
            )}
          >
            <div className="font-medium">Static IP</div>
            <div className="text-xs text-muted-foreground">
              Configure fixed IP address
            </div>
          </button>
        </div>
      </div>

      {/* Static IP Configuration */}
      {currentInterface.ipMethod === 'static' && (
        <div className="space-y-4 p-4 bg-gray-50 dark:bg-gray-800/50 rounded-xl">
          <div className="flex items-center gap-2">
            <Network className="w-4 h-4 text-muted-foreground" />
            <Label className="text-sm font-semibold">Static IP Configuration</Label>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="ip-address">IP Address *</Label>
              <Input
                id="ip-address"
                placeholder="192.168.1.100"
                value={currentInterface.ipAddress || ''}
                onChange={(e) => {
                  const value = e.target.value
                  if (config.primaryInterface === 'wifi') {
                    updateWifi({ ipAddress: value })
                  } else {
                    updateEthernet({ ipAddress: value })
                  }
                  updateConfig({ ipAddress: value })
                }}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="subnet">Subnet Mask *</Label>
              <select
                id="subnet"
                value={currentInterface.subnetMask || '255.255.255.0'}
                onChange={(e) => {
                  const value = e.target.value
                  if (config.primaryInterface === 'wifi') {
                    updateWifi({ subnetMask: value })
                  } else {
                    updateEthernet({ subnetMask: value })
                  }
                }}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-primary"
              >
                {COMMON_SUBNET_MASKS.map((subnet) => (
                  <option key={subnet.mask} value={subnet.mask}>
                    {subnet.mask} ({subnet.cidr} - {subnet.hosts})
                  </option>
                ))}
              </select>
            </div>

            <div className="space-y-2">
              <Label htmlFor="gateway">Gateway *</Label>
              <Input
                id="gateway"
                placeholder="192.168.1.1"
                value={currentInterface.gateway || ''}
                onChange={(e) => {
                  const value = e.target.value
                  if (config.primaryInterface === 'wifi') {
                    updateWifi({ gateway: value })
                  } else {
                    updateEthernet({ gateway: value })
                  }
                  updateConfig({ gateway: value })
                }}
              />
            </div>

            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <Label htmlFor="dns1">Primary DNS *</Label>
                <button
                  onClick={() => setShowDnsPresets(!showDnsPresets)}
                  className="text-xs text-primary hover:underline"
                >
                  Use preset
                </button>
              </div>
              <Input
                id="dns1"
                placeholder="8.8.8.8"
                value={currentInterface.dns1 || ''}
                onChange={(e) => {
                  const value = e.target.value
                  if (config.primaryInterface === 'wifi') {
                    updateWifi({ dns1: value })
                  } else {
                    updateEthernet({ dns1: value })
                  }
                  updateConfig({ dns: value })
                }}
              />
            </div>

            {showDnsPresets && (
              <div className="md:col-span-2 p-3 bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700">
                <p className="text-xs font-semibold text-muted-foreground mb-2">
                  DNS Presets
                </p>
                <div className="flex flex-wrap gap-2">
                  {COMMON_DNS_SERVERS.map((preset) => (
                    <Button
                      key={preset.name}
                      variant="outline"
                      size="sm"
                      onClick={() => applyDnsPreset(preset)}
                    >
                      {preset.name}
                    </Button>
                  ))}
                </div>
              </div>
            )}

            <div className="space-y-2">
              <Label htmlFor="dns2">Secondary DNS (optional)</Label>
              <Input
                id="dns2"
                placeholder="8.8.4.4"
                value={currentInterface.dns2 || ''}
                onChange={(e) => {
                  const value = e.target.value
                  if (config.primaryInterface === 'wifi') {
                    updateWifi({ dns2: value })
                  } else {
                    updateEthernet({ dns2: value })
                  }
                }}
              />
            </div>
          </div>

          {/* nmcli command preview */}
          <div className="pt-3 border-t border-gray-200 dark:border-gray-700">
            <p className="text-xs font-medium text-muted-foreground mb-2">
              NetworkManager command (nmcli):
            </p>
            <pre className="text-xs bg-gray-900 text-gray-100 p-3 rounded-lg overflow-x-auto">
              {`sudo nmcli c mod "${config.primaryInterface === 'wifi' ? 'preconfigured' : 'Wired connection 1'}" \\
  ipv4.addresses ${currentInterface.ipAddress || '192.168.1.100'}/${
    currentInterface.subnetMask === '255.255.255.0'
      ? '24'
      : currentInterface.subnetMask === '255.255.0.0'
        ? '16'
        : '24'
  } \\
  ipv4.gateway ${currentInterface.gateway || '192.168.1.1'} \\
  ipv4.dns "${currentInterface.dns1 || '8.8.8.8'}${currentInterface.dns2 ? ',' + currentInterface.dns2 : ''}" \\
  ipv4.method manual`}
            </pre>
          </div>
        </div>
      )}

      {/* Advanced Settings */}
      <button
        onClick={() => setShowAdvanced(!showAdvanced)}
        className="w-full flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-800/50 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700/50 transition-colors"
      >
        <span className="text-sm font-medium">Advanced Settings</span>
        {showAdvanced ? (
          <ChevronUp className="w-4 h-4" />
        ) : (
          <ChevronDown className="w-4 h-4" />
        )}
      </button>

      {showAdvanced && (
        <div className="space-y-4 p-4 bg-gray-50 dark:bg-gray-800/50 rounded-xl">
          <div className="space-y-2">
            <Label htmlFor="mtu">MTU (Maximum Transmission Unit)</Label>
            <Input
              id="mtu"
              type="number"
              placeholder="1500"
              value={currentInterface.mtu || ''}
              onChange={(e) => {
                const value = parseInt(e.target.value) || undefined
                if (config.primaryInterface === 'wifi') {
                  updateWifi({ mtu: value })
                } else {
                  updateEthernet({ mtu: value })
                }
              }}
            />
            <p className="text-xs text-muted-foreground">
              Default is 1500. Lower values may help with VPN or tunneling
              issues.
            </p>
          </div>

          {config.primaryInterface === 'wifi' && (
            <div className="space-y-2">
              <Label>WiFi Band Preference</Label>
              <div className="flex gap-2">
                {(['auto', '2.4GHz', '5GHz'] as const).map((band) => (
                  <button
                    key={band}
                    onClick={() => updateWifi({ band })}
                    className={cn(
                      'px-3 py-2 text-sm rounded-lg border transition-colors',
                      config.wifi.band === band
                        ? 'border-primary bg-primary/5'
                        : 'border-gray-200 dark:border-gray-700 hover:border-gray-300'
                    )}
                  >
                    {band}
                  </button>
                ))}
              </div>
              <p className="text-xs text-muted-foreground">
                5GHz offers faster speeds but shorter range
              </p>
            </div>
          )}
        </div>
      )}

      {/* Tips */}
      <div className="p-4 bg-blue-50 dark:bg-blue-950/30 rounded-xl border border-blue-100 dark:border-blue-900">
        <p className="text-sm text-blue-700 dark:text-blue-300">
          <strong>Tip:</strong> For production IoT deployments, we recommend
          using Ethernet with a static IP or DHCP reservation on your router for
          reliable connectivity.
        </p>
      </div>
    </div>
  )
}
