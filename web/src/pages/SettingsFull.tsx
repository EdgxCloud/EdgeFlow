import { useState, useEffect, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  Save,
  Info,
  Cpu,
  Settings2,
  Wifi,
  WifiOff,
  Globe,
  Network,
  RefreshCw,
  Signal,
  SignalLow,
  SignalMedium,
  SignalHigh,
  Lock,
  Unlock,
  Check,
  ChevronDown,
  ChevronUp,
  Eye,
  EyeOff,
  Server,
  Activity,
} from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'
import { Switch } from '@/components/ui/switch'
import { SetupWizard } from '@/components/SetupWizard'
import {
  COMMON_DNS_SERVERS,
  COMMON_SUBNET_MASKS,
  WIFI_COUNTRIES,
  type WifiSecurityType,
} from '@/components/SetupWizard/types'
import { cn } from '@/lib/utils'
import { networkApi, type NetworkInfo, type NetworkInterface } from '@/services/network'

interface WifiNetwork {
  ssid: string
  bssid: string
  signal: number
  frequency: number
  channel: number
  security: WifiSecurityType
  connected: boolean
}

interface NetworkSettings {
  hostname: string
  primaryInterface: 'ethernet' | 'wifi'
  ethernet: {
    enabled: boolean
    ipMethod: 'dhcp' | 'static'
    ipAddress: string
    subnetMask: string
    gateway: string
    dns1: string
    dns2: string
  }
  wifi: {
    enabled: boolean
    ssid: string
    password: string
    ipMethod: 'dhcp' | 'static'
    ipAddress: string
    subnetMask: string
    gateway: string
    dns1: string
    dns2: string
    country: string
  }
}

// Demo WiFi networks for simulation
const DEMO_WIFI_NETWORKS: WifiNetwork[] = [
  { ssid: 'Home_Network_5G', bssid: 'AA:BB:CC:DD:EE:01', signal: -45, frequency: 5180, channel: 36, security: 'wpa2', connected: false },
  { ssid: 'Home_Network_2.4G', bssid: 'AA:BB:CC:DD:EE:02', signal: -55, frequency: 2437, channel: 6, security: 'wpa2', connected: true },
  { ssid: 'Guest_WiFi', bssid: 'AA:BB:CC:DD:EE:03', signal: -65, frequency: 2412, channel: 1, security: 'wpa', connected: false },
  { ssid: 'Office_Network', bssid: 'AA:BB:CC:DD:EE:04', signal: -70, frequency: 5240, channel: 48, security: 'wpa2-enterprise', connected: false },
  { ssid: 'IoT_Network', bssid: 'AA:BB:CC:DD:EE:05', signal: -60, frequency: 2462, channel: 11, security: 'wpa2', connected: false },
]

export default function SettingsFull() {
  const navigate = useNavigate()
  const [showSetupWizard, setShowSetupWizard] = useState(false)
  const [showPassword, setShowPassword] = useState(false)
  const [scanning, setScanning] = useState(false)
  const [wifiNetworks, setWifiNetworks] = useState<WifiNetwork[]>([])
  const [showNetworkAdvanced, setShowNetworkAdvanced] = useState(false)
  const [showDnsPresets, setShowDnsPresets] = useState(false)
  const [networkInfo, setNetworkInfo] = useState<NetworkInfo | null>(null)
  const [networkLoading, setNetworkLoading] = useState(false)

  const [settings, setSettings] = useState({
    serverHost: 'localhost',
    serverPort: '8080',
    logLevel: 'info',
    maxExecutions: '100',
    enableWebSocket: true,
    darkMode: true,
  })

  const [networkSettings, setNetworkSettings] = useState<NetworkSettings>({
    hostname: 'edgeflow-device',
    primaryInterface: 'ethernet',
    ethernet: {
      enabled: true,
      ipMethod: 'dhcp',
      ipAddress: '',
      subnetMask: '255.255.255.0',
      gateway: '',
      dns1: '',
      dns2: '',
    },
    wifi: {
      enabled: false,
      ssid: '',
      password: '',
      ipMethod: 'dhcp',
      ipAddress: '',
      subnetMask: '255.255.255.0',
      gateway: '',
      dns1: '',
      dns2: '',
      country: 'US',
    },
  })

  // Scan for WiFi networks
  const handleScanNetworks = useCallback(() => {
    setScanning(true)
    setTimeout(() => {
      const networks = DEMO_WIFI_NETWORKS.map((n) => ({
        ...n,
        signal: n.signal + Math.floor(Math.random() * 10) - 5,
      })).sort((a, b) => b.signal - a.signal)
      setWifiNetworks(networks)
      setScanning(false)
    }, 2000)
  }, [])

  // Fetch real network info when diagnostics panel opens
  useEffect(() => {
    if (showNetworkAdvanced) {
      setNetworkLoading(true)
      networkApi.getInfo()
        .then(data => setNetworkInfo(data))
        .catch(err => console.error('Failed to fetch network info:', err))
        .finally(() => setNetworkLoading(false))
    }
  }, [showNetworkAdvanced])

  // Auto-scan on WiFi enable
  useEffect(() => {
    if (networkSettings.wifi.enabled && wifiNetworks.length === 0) {
      handleScanNetworks()
    }
  }, [networkSettings.wifi.enabled, wifiNetworks.length, handleScanNetworks])

  const handleSave = () => {
    console.log('Saving settings:', settings)
    console.log('Saving network settings:', networkSettings)
  }

  const updateNetworkSettings = (updates: Partial<NetworkSettings>) => {
    setNetworkSettings((prev) => ({ ...prev, ...updates }))
  }

  const updateEthernet = (updates: Partial<NetworkSettings['ethernet']>) => {
    setNetworkSettings((prev) => ({
      ...prev,
      ethernet: { ...prev.ethernet, ...updates },
    }))
  }

  const updateWifi = (updates: Partial<NetworkSettings['wifi']>) => {
    setNetworkSettings((prev) => ({
      ...prev,
      wifi: { ...prev.wifi, ...updates },
    }))
  }

  const selectWifiNetwork = (network: WifiNetwork) => {
    updateWifi({ ssid: network.ssid })
  }

  const applyDnsPreset = (preset: (typeof COMMON_DNS_SERVERS)[0], target: 'ethernet' | 'wifi') => {
    if (target === 'wifi') {
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

  const currentInterface = networkSettings.primaryInterface === 'wifi'
    ? networkSettings.wifi
    : networkSettings.ethernet

  return (
    <div className="space-y-6 max-w-4xl">
      {/* Header */}
      <div>
        <h1 className="text-3xl font-bold text-gray-900 dark:text-white">
          Settings
        </h1>
        <p className="text-gray-600 dark:text-gray-400 mt-1">
          System settings and EdgeFlow configuration
        </p>
      </div>

      {/* Device Setup */}
      <div className="bg-gradient-to-r from-blue-500 to-purple-600 rounded-lg shadow p-6 text-white">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <div className="w-12 h-12 bg-white/20 rounded-xl flex items-center justify-center">
              <Cpu className="w-6 h-6" />
            </div>
            <div>
              <h2 className="text-lg font-semibold">Device Setup Wizard</h2>
              <p className="text-white/80 text-sm">
                Configure board, network, MQTT, and GPIO settings
              </p>
            </div>
          </div>
          <Button
            onClick={() => setShowSetupWizard(true)}
            className="bg-white text-blue-600 hover:bg-white/90"
          >
            <Settings2 className="w-4 h-4 mr-2" />
            Open Setup Wizard
          </Button>
        </div>
      </div>

      {/* Setup Wizard Modal */}
      <SetupWizard
        isOpen={showSetupWizard}
        onClose={() => setShowSetupWizard(false)}
        onComplete={(config) => {
          console.log('Setup complete:', config)
          setShowSetupWizard(false)
        }}
        onGoToEditor={() => {
          setShowSetupWizard(false)
          navigate('/editor')
        }}
      />

      {/* Network Configuration */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <div className="flex items-center gap-2 mb-6">
          <Network className="w-5 h-5 text-blue-500" />
          <h2 className="text-lg font-semibold text-gray-900 dark:text-white">
            Network Configuration
          </h2>
        </div>

        <div className="space-y-6">
          {/* Hostname */}
          <div className="space-y-2">
            <Label htmlFor="hostname">Device Hostname</Label>
            <div className="relative">
              <Server className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-gray-400" />
              <Input
                id="hostname"
                placeholder="edgeflow-device"
                value={networkSettings.hostname}
                onChange={(e) => updateNetworkSettings({ hostname: e.target.value })}
                className="pl-10"
              />
            </div>
            <p className="text-xs text-muted-foreground">
              Access via {networkSettings.hostname}.local
            </p>
          </div>

          {/* Connection Type */}
          <div className="space-y-3">
            <Label className="text-sm font-semibold">Primary Connection</Label>
            <div className="grid grid-cols-2 gap-3">
              <button
                onClick={() => updateNetworkSettings({ primaryInterface: 'ethernet' })}
                className={cn(
                  'p-4 rounded-xl border-2 transition-all flex items-center gap-3',
                  networkSettings.primaryInterface === 'ethernet'
                    ? 'border-blue-500 bg-blue-50 dark:bg-blue-950/30'
                    : 'border-gray-200 dark:border-gray-700 hover:border-gray-300'
                )}
              >
                <div className={cn(
                  'w-10 h-10 rounded-lg flex items-center justify-center',
                  networkSettings.primaryInterface === 'ethernet'
                    ? 'bg-blue-100 dark:bg-blue-900/50'
                    : 'bg-gray-100 dark:bg-gray-700'
                )}>
                  <Globe className="w-5 h-5 text-blue-500" />
                </div>
                <div className="text-left">
                  <div className="font-medium text-gray-900 dark:text-white">Ethernet</div>
                  <div className="text-xs text-muted-foreground">Wired (eth0)</div>
                </div>
                {networkSettings.primaryInterface === 'ethernet' && (
                  <Check className="w-5 h-5 text-blue-500 ml-auto" />
                )}
              </button>

              <button
                onClick={() => {
                  updateNetworkSettings({ primaryInterface: 'wifi' })
                  updateWifi({ enabled: true })
                }}
                className={cn(
                  'p-4 rounded-xl border-2 transition-all flex items-center gap-3',
                  networkSettings.primaryInterface === 'wifi'
                    ? 'border-green-500 bg-green-50 dark:bg-green-950/30'
                    : 'border-gray-200 dark:border-gray-700 hover:border-gray-300'
                )}
              >
                <div className={cn(
                  'w-10 h-10 rounded-lg flex items-center justify-center',
                  networkSettings.primaryInterface === 'wifi'
                    ? 'bg-green-100 dark:bg-green-900/50'
                    : 'bg-gray-100 dark:bg-gray-700'
                )}>
                  <Wifi className="w-5 h-5 text-green-500" />
                </div>
                <div className="text-left">
                  <div className="font-medium text-gray-900 dark:text-white">WiFi</div>
                  <div className="text-xs text-muted-foreground">Wireless (wlan0)</div>
                </div>
                {networkSettings.primaryInterface === 'wifi' && (
                  <Check className="w-5 h-5 text-green-500 ml-auto" />
                )}
              </button>
            </div>
          </div>

          {/* WiFi Configuration */}
          {networkSettings.primaryInterface === 'wifi' && (
            <div className="space-y-4 p-4 bg-gray-50 dark:bg-gray-900/50 rounded-xl border border-gray-200 dark:border-gray-700">
              <div className="flex items-center justify-between">
                <Label className="text-sm font-semibold flex items-center gap-2">
                  <Wifi className="w-4 h-4 text-green-500" />
                  WiFi Networks
                </Label>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={handleScanNetworks}
                  disabled={scanning}
                >
                  <RefreshCw className={cn('w-4 h-4 mr-2', scanning && 'animate-spin')} />
                  {scanning ? 'Scanning...' : 'Scan'}
                </Button>
              </div>

              {/* Networks List */}
              {wifiNetworks.length > 0 ? (
                <div className="space-y-2 max-h-48 overflow-y-auto">
                  {wifiNetworks.map((network) => {
                    const SignalIcon = getSignalIcon(network.signal)
                    const signalQuality = getSignalQuality(network.signal)
                    const isSelected = networkSettings.wifi.ssid === network.ssid

                    return (
                      <button
                        key={network.bssid}
                        onClick={() => selectWifiNetwork(network)}
                        className={cn(
                          'w-full p-3 rounded-lg border transition-all flex items-center gap-3 text-left',
                          isSelected
                            ? 'border-green-500 bg-green-50 dark:bg-green-950/30'
                            : 'border-gray-200 dark:border-gray-600 hover:border-gray-300 hover:bg-white dark:hover:bg-gray-800'
                        )}
                      >
                        <SignalIcon className={cn('w-5 h-5', signalQuality.color)} />
                        <div className="flex-1 min-w-0">
                          <div className="flex items-center gap-2">
                            <span className="font-medium text-gray-900 dark:text-white truncate">
                              {network.ssid}
                            </span>
                            {network.connected && (
                              <span className="text-xs bg-green-100 dark:bg-green-900 text-green-700 dark:text-green-300 px-2 py-0.5 rounded">
                                Connected
                              </span>
                            )}
                            {network.security !== 'open' ? (
                              <Lock className="w-3 h-3 text-muted-foreground" />
                            ) : (
                              <Unlock className="w-3 h-3 text-amber-500" />
                            )}
                          </div>
                          <div className="flex items-center gap-2 text-xs text-muted-foreground">
                            <span>{getSecurityLabel(network.security)}</span>
                            <span>•</span>
                            <span>{network.frequency > 5000 ? '5GHz' : '2.4GHz'}</span>
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
                        {isSelected && <Check className="w-5 h-5 text-green-500 flex-shrink-0" />}
                      </button>
                    )
                  })}
                </div>
              ) : scanning ? (
                <div className="text-center py-6">
                  <RefreshCw className="w-6 h-6 mx-auto text-muted-foreground animate-spin mb-2" />
                  <p className="text-sm text-muted-foreground">Scanning for networks...</p>
                </div>
              ) : (
                <div className="text-center py-6">
                  <WifiOff className="w-6 h-6 mx-auto text-muted-foreground mb-2" />
                  <p className="text-sm text-muted-foreground">Click Scan to find networks</p>
                </div>
              )}

              {/* WiFi Credentials */}
              <div className="pt-3 border-t border-gray-200 dark:border-gray-600 space-y-3">
                <div className="grid grid-cols-2 gap-3">
                  <div className="space-y-2">
                    <Label htmlFor="wifi-ssid">Network Name (SSID)</Label>
                    <Input
                      id="wifi-ssid"
                      placeholder="Enter WiFi name"
                      value={networkSettings.wifi.ssid}
                      onChange={(e) => updateWifi({ ssid: e.target.value })}
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="wifi-password">Password</Label>
                    <div className="relative">
                      <Input
                        id="wifi-password"
                        type={showPassword ? 'text' : 'password'}
                        placeholder="Enter password"
                        value={networkSettings.wifi.password}
                        onChange={(e) => updateWifi({ password: e.target.value })}
                        className="pr-10"
                      />
                      <button
                        type="button"
                        onClick={() => setShowPassword(!showPassword)}
                        className="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600"
                      >
                        {showPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                      </button>
                    </div>
                  </div>
                </div>

                <div className="space-y-2">
                  <Label htmlFor="wifi-country">WiFi Country/Region</Label>
                  <select
                    id="wifi-country"
                    value={networkSettings.wifi.country}
                    onChange={(e) => updateWifi({ country: e.target.value })}
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                  >
                    {WIFI_COUNTRIES.map((country) => (
                      <option key={country.code} value={country.code}>
                        {country.name} ({country.code})
                      </option>
                    ))}
                  </select>
                </div>
              </div>
            </div>
          )}

          {/* IP Configuration */}
          <div className="space-y-3">
            <Label className="text-sm font-semibold">IP Configuration</Label>
            <div className="grid grid-cols-2 gap-3">
              <button
                onClick={() => {
                  if (networkSettings.primaryInterface === 'wifi') {
                    updateWifi({ ipMethod: 'dhcp' })
                  } else {
                    updateEthernet({ ipMethod: 'dhcp' })
                  }
                }}
                className={cn(
                  'p-3 rounded-lg border-2 transition-all text-left',
                  currentInterface.ipMethod === 'dhcp'
                    ? 'border-blue-500 bg-blue-50 dark:bg-blue-950/30'
                    : 'border-gray-200 dark:border-gray-700 hover:border-gray-300'
                )}
              >
                <div className="font-medium text-gray-900 dark:text-white">DHCP (Automatic)</div>
                <div className="text-xs text-muted-foreground">Get IP from router</div>
              </button>

              <button
                onClick={() => {
                  if (networkSettings.primaryInterface === 'wifi') {
                    updateWifi({ ipMethod: 'static' })
                  } else {
                    updateEthernet({ ipMethod: 'static' })
                  }
                }}
                className={cn(
                  'p-3 rounded-lg border-2 transition-all text-left',
                  currentInterface.ipMethod === 'static'
                    ? 'border-blue-500 bg-blue-50 dark:bg-blue-950/30'
                    : 'border-gray-200 dark:border-gray-700 hover:border-gray-300'
                )}
              >
                <div className="font-medium text-gray-900 dark:text-white">Static IP</div>
                <div className="text-xs text-muted-foreground">Fixed IP address</div>
              </button>
            </div>
          </div>

          {/* Static IP Settings */}
          {currentInterface.ipMethod === 'static' && (
            <div className="space-y-4 p-4 bg-gray-50 dark:bg-gray-900/50 rounded-xl border border-gray-200 dark:border-gray-700">
              <Label className="text-sm font-semibold flex items-center gap-2">
                <Activity className="w-4 h-4" />
                Static IP Configuration
              </Label>

              <div className="grid grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="ip-address">IP Address</Label>
                  <Input
                    id="ip-address"
                    placeholder="192.168.1.100"
                    value={currentInterface.ipAddress}
                    onChange={(e) => {
                      if (networkSettings.primaryInterface === 'wifi') {
                        updateWifi({ ipAddress: e.target.value })
                      } else {
                        updateEthernet({ ipAddress: e.target.value })
                      }
                    }}
                  />
                </div>

                <div className="space-y-2">
                  <Label htmlFor="subnet">Subnet Mask</Label>
                  <select
                    id="subnet"
                    value={currentInterface.subnetMask}
                    onChange={(e) => {
                      if (networkSettings.primaryInterface === 'wifi') {
                        updateWifi({ subnetMask: e.target.value })
                      } else {
                        updateEthernet({ subnetMask: e.target.value })
                      }
                    }}
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                  >
                    {COMMON_SUBNET_MASKS.map((subnet) => (
                      <option key={subnet.mask} value={subnet.mask}>
                        {subnet.mask} ({subnet.cidr})
                      </option>
                    ))}
                  </select>
                </div>

                <div className="space-y-2">
                  <Label htmlFor="gateway">Gateway</Label>
                  <Input
                    id="gateway"
                    placeholder="192.168.1.1"
                    value={currentInterface.gateway}
                    onChange={(e) => {
                      if (networkSettings.primaryInterface === 'wifi') {
                        updateWifi({ gateway: e.target.value })
                      } else {
                        updateEthernet({ gateway: e.target.value })
                      }
                    }}
                  />
                </div>

                <div className="space-y-2">
                  <div className="flex items-center justify-between">
                    <Label htmlFor="dns1">Primary DNS</Label>
                    <button
                      onClick={() => setShowDnsPresets(!showDnsPresets)}
                      className="text-xs text-blue-500 hover:underline"
                    >
                      Presets
                    </button>
                  </div>
                  <Input
                    id="dns1"
                    placeholder="8.8.8.8"
                    value={currentInterface.dns1}
                    onChange={(e) => {
                      if (networkSettings.primaryInterface === 'wifi') {
                        updateWifi({ dns1: e.target.value })
                      } else {
                        updateEthernet({ dns1: e.target.value })
                      }
                    }}
                  />
                </div>

                {showDnsPresets && (
                  <div className="col-span-2 p-3 bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-600">
                    <p className="text-xs font-semibold text-muted-foreground mb-2">
                      Quick DNS Presets
                    </p>
                    <div className="flex flex-wrap gap-2">
                      {COMMON_DNS_SERVERS.map((preset) => (
                        <Button
                          key={preset.name}
                          variant="outline"
                          size="sm"
                          onClick={() => applyDnsPreset(preset, networkSettings.primaryInterface)}
                        >
                          {preset.name}
                        </Button>
                      ))}
                    </div>
                  </div>
                )}

                <div className="space-y-2">
                  <Label htmlFor="dns2">Secondary DNS</Label>
                  <Input
                    id="dns2"
                    placeholder="8.8.4.4"
                    value={currentInterface.dns2}
                    onChange={(e) => {
                      if (networkSettings.primaryInterface === 'wifi') {
                        updateWifi({ dns2: e.target.value })
                      } else {
                        updateEthernet({ dns2: e.target.value })
                      }
                    }}
                  />
                </div>
              </div>

              {/* nmcli Command Preview */}
              <div className="pt-3 border-t border-gray-200 dark:border-gray-600">
                <p className="text-xs font-medium text-muted-foreground mb-2">
                  nmcli command:
                </p>
                <pre className="text-xs bg-gray-900 text-gray-100 p-3 rounded-lg overflow-x-auto">
{`sudo nmcli c mod "${networkSettings.primaryInterface === 'wifi' ? networkSettings.wifi.ssid || 'WiFi' : 'Wired'}" \\
  ipv4.addresses ${currentInterface.ipAddress || '192.168.1.100'}/${currentInterface.subnetMask === '255.255.255.0' ? '24' : '16'} \\
  ipv4.gateway ${currentInterface.gateway || '192.168.1.1'} \\
  ipv4.dns "${currentInterface.dns1 || '8.8.8.8'}${currentInterface.dns2 ? ',' + currentInterface.dns2 : ''}" \\
  ipv4.method manual`}
                </pre>
              </div>
            </div>
          )}

          {/* Network Status */}
          <button
            onClick={() => setShowNetworkAdvanced(!showNetworkAdvanced)}
            className="w-full flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-900/50 rounded-lg hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
          >
            <span className="text-sm font-medium text-gray-700 dark:text-gray-300">
              Network Status & Diagnostics
            </span>
            {showNetworkAdvanced ? <ChevronUp className="w-4 h-4" /> : <ChevronDown className="w-4 h-4" />}
          </button>

          {showNetworkAdvanced && (
            <div className="p-4 bg-gray-50 dark:bg-gray-900/50 rounded-xl border border-gray-200 dark:border-gray-700 space-y-3">
              {networkLoading ? (
                <div className="flex items-center justify-center py-4 gap-2">
                  <RefreshCw className="w-4 h-4 animate-spin text-muted-foreground" />
                  <span className="text-sm text-muted-foreground">Loading network info...</span>
                </div>
              ) : networkInfo ? (
                <>
                  <div className="grid grid-cols-2 gap-4 text-sm">
                    <div>
                      <span className="text-muted-foreground">Hostname:</span>
                      <span className="ml-2 font-mono text-gray-900 dark:text-white">
                        {networkInfo.hostname}
                      </span>
                    </div>
                    <div>
                      <span className="text-muted-foreground">Active Interface:</span>
                      <span className="ml-2 font-mono text-gray-900 dark:text-white">
                        {networkInfo.interfaces.find(i => i.status === 'up' && i.ipv4.length > 0 && i.name !== 'lo')?.name || 'None'}
                      </span>
                    </div>
                    <div>
                      <span className="text-muted-foreground">Status:</span>
                      {(() => {
                        const active = networkInfo.interfaces.find(i => i.status === 'up' && i.ipv4.length > 0 && i.name !== 'lo')
                        return (
                          <span className={`ml-2 font-medium ${active ? 'text-green-500' : 'text-red-500'}`}>
                            {active ? 'Connected' : 'Disconnected'}
                          </span>
                        )
                      })()}
                    </div>
                    <div>
                      <span className="text-muted-foreground">IP Address:</span>
                      <span className="ml-2 font-mono text-gray-900 dark:text-white">
                        {networkInfo.interfaces.find(i => i.status === 'up' && i.ipv4.length > 0 && i.name !== 'lo')?.ipv4?.[0] || 'N/A'}
                      </span>
                    </div>
                    <div>
                      <span className="text-muted-foreground">MAC:</span>
                      <span className="ml-2 font-mono text-gray-900 dark:text-white">
                        {networkInfo.interfaces.find(i => i.status === 'up' && i.ipv4.length > 0 && i.name !== 'lo')?.mac || 'N/A'}
                      </span>
                    </div>
                    <div>
                      <span className="text-muted-foreground">IP Method:</span>
                      <span className="ml-2 font-mono text-gray-900 dark:text-white uppercase">
                        {currentInterface.ipMethod}
                      </span>
                    </div>
                  </div>

                  {networkInfo.interfaces.filter(i => i.name !== 'lo').length > 1 && (
                    <div className="pt-3 border-t border-gray-200 dark:border-gray-600">
                      <p className="text-xs text-muted-foreground mb-2">All Interfaces:</p>
                      <div className="space-y-1">
                        {networkInfo.interfaces
                          .filter(i => i.name !== 'lo')
                          .map(iface => (
                            <div key={iface.name} className="flex items-center gap-3 text-xs font-mono">
                              <span className={`w-2 h-2 rounded-full ${iface.status === 'up' ? 'bg-green-500' : 'bg-gray-400'}`} />
                              <span className="w-16">{iface.name}</span>
                              <span className="text-muted-foreground">{iface.mac || 'no mac'}</span>
                              <span className="text-gray-900 dark:text-white">{iface.ipv4?.[0] || 'no ip'}</span>
                            </div>
                          ))}
                      </div>
                    </div>
                  )}

                  <div className="pt-3 border-t border-gray-200 dark:border-gray-600">
                    <p className="text-xs text-muted-foreground mb-2">Diagnostic Commands:</p>
                    <div className="flex flex-wrap gap-2">
                      <code className="text-xs bg-gray-800 text-gray-100 px-2 py-1 rounded">
                        nmcli device status
                      </code>
                      <code className="text-xs bg-gray-800 text-gray-100 px-2 py-1 rounded">
                        ip addr show
                      </code>
                      <code className="text-xs bg-gray-800 text-gray-100 px-2 py-1 rounded">
                        nmcli c show
                      </code>
                    </div>
                  </div>
                </>
              ) : (
                <div className="grid grid-cols-2 gap-4 text-sm">
                  <div>
                    <span className="text-muted-foreground">Interface:</span>
                    <span className="ml-2 font-mono text-gray-900 dark:text-white">
                      {networkSettings.primaryInterface === 'wifi' ? 'wlan0' : 'eth0'}
                    </span>
                  </div>
                  <div>
                    <span className="text-muted-foreground">Status:</span>
                    <span className="ml-2 text-yellow-500 font-medium">Unknown</span>
                  </div>
                </div>
              )}
            </div>
          )}
        </div>
      </div>

      {/* Server Settings */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
          Server Settings
        </h2>
        <div className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <Input
              label="Server Address"
              value={settings.serverHost}
              onChange={(e) =>
                setSettings({ ...settings, serverHost: e.target.value })
              }
              placeholder="localhost"
            />
            <Input
              label="Port"
              type="number"
              value={settings.serverPort}
              onChange={(e) =>
                setSettings({ ...settings, serverPort: e.target.value })
              }
              placeholder="8080"
            />
          </div>
        </div>
      </div>

      {/* Engine Settings */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
          Engine Settings
        </h2>
        <div className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-1">
                Log Level
              </label>
              <select
                value={settings.logLevel}
                onChange={(e) =>
                  setSettings({ ...settings, logLevel: e.target.value })
                }
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="debug">Debug</option>
                <option value="info">Info</option>
                <option value="warn">Warning</option>
                <option value="error">Error</option>
              </select>
            </div>
            <Input
              label="Max Concurrent Executions"
              type="number"
              value={settings.maxExecutions}
              onChange={(e) =>
                setSettings({ ...settings, maxExecutions: e.target.value })
              }
              placeholder="100"
            />
          </div>

          <div className="flex items-center space-x-3">
            <input
              type="checkbox"
              id="enableWebSocket"
              checked={settings.enableWebSocket}
              onChange={(e) =>
                setSettings({ ...settings, enableWebSocket: e.target.checked })
              }
              className="w-4 h-4 text-blue-600 bg-gray-100 border-gray-300 rounded focus:ring-blue-500 dark:focus:ring-blue-600 dark:ring-offset-gray-800 focus:ring-2 dark:bg-gray-700 dark:border-gray-600"
            />
            <label
              htmlFor="enableWebSocket"
              className="text-sm font-medium text-gray-700 dark:text-gray-300"
            >
              Enable WebSocket for real-time updates
            </label>
          </div>
        </div>
      </div>

      {/* UI Settings */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
          UI Settings
        </h2>
        <div className="space-y-4">
          <div className="flex items-center space-x-3">
            <input
              type="checkbox"
              id="darkMode"
              checked={settings.darkMode}
              onChange={(e) =>
                setSettings({ ...settings, darkMode: e.target.checked })
              }
              className="w-4 h-4 text-blue-600 bg-gray-100 border-gray-300 rounded focus:ring-blue-500 dark:focus:ring-blue-600 dark:ring-offset-gray-800 focus:ring-2 dark:bg-gray-700 dark:border-gray-600"
            />
            <label
              htmlFor="darkMode"
              className="text-sm font-medium text-gray-700 dark:text-gray-300"
            >
              Dark Mode
            </label>
          </div>
        </div>
      </div>

      {/* About */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <div className="flex items-start space-x-3">
          <Info className="w-5 h-5 text-blue-500 mt-0.5" />
          <div>
            <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-2">
              About EdgeFlow
            </h2>
            <div className="text-sm text-gray-600 dark:text-gray-400 space-y-1">
              <p>
                <strong>Version:</strong> 0.1.0
              </p>
              <p>
                <strong>Environment:</strong> Development
              </p>
              <p>
                <strong>Backend:</strong> Go 1.25
              </p>
              <p>
                <strong>Frontend:</strong> React 18 + TypeScript
              </p>
              <p className="mt-4">
                EdgeFlow - Lightweight automation platform for Edge & IoT
              </p>
              <p className="text-xs text-gray-500 dark:text-gray-500 mt-2">
                © 2026 EdgeFlow Team. All rights reserved.
              </p>
            </div>
          </div>
        </div>
      </div>

      {/* Actions */}
      <div className="flex items-center justify-end space-x-3">
        <Button variant="secondary">Cancel</Button>
        <Button variant="primary" icon={<Save className="w-4 h-4" />} onClick={handleSave}>
          Save Settings
        </Button>
      </div>
    </div>
  )
}
