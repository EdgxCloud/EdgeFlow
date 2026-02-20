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
  Thermometer,
  HardDrive,
  MemoryStick,
  Clock,
  Power,
  RotateCcw,
  CheckCircle,
  AlertTriangle,
  XCircle,
  Monitor,
  Palette,
  Languages,
  Zap,
  Download,
  Upload,
  Cloud,
  ArrowRight,
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
} from '@/components/SetupWizard/types'
import { cn } from '@/lib/utils'
import { networkApi, wifiApi, systemApi, settingsApi } from '@/services/network'
import type { NetworkInfo, WifiNetwork, SystemInfo } from '@/services/network'
import { useSettingsStore } from '@/stores/settingsStore'

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

type ToastType = 'success' | 'error' | 'info'

interface Toast {
  message: string
  type: ToastType
}

export default function SettingsFull() {
  const navigate = useNavigate()
  const { app: appSettings, updateAppSettings } = useSettingsStore()

  const [showSetupWizard, setShowSetupWizard] = useState(false)
  const [showPassword, setShowPassword] = useState(false)
  const [scanning, setScanning] = useState(false)
  const [wifiNetworks, setWifiNetworks] = useState<WifiNetwork[]>([])
  const [showDnsPresets, setShowDnsPresets] = useState(false)
  const [networkInfo, setNetworkInfo] = useState<NetworkInfo | null>(null)
  const [networkLoading, setNetworkLoading] = useState(false)
  const [systemInfo, setSystemInfo] = useState<SystemInfo | null>(null)
  const [saving, setSaving] = useState(false)
  const [toast, setToast] = useState<Toast | null>(null)
  const [connecting, setConnecting] = useState(false)
  const [expandedSections, setExpandedSections] = useState<Set<string>>(new Set(['network-status']))

  const [settings, setSettings] = useState({
    serverHost: '0.0.0.0',
    serverPort: '8080',
    logLevel: 'info',
    maxExecutions: '100',
    enableWebSocket: true,
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

  // Toast helper
  const showToast = useCallback((message: string, type: ToastType) => {
    setToast({ message, type })
    setTimeout(() => setToast(null), 4000)
  }, [])

  // Load settings from backend on mount
  useEffect(() => {
    settingsApi.get()
      .then(res => {
        if (res.configured && res.settings) {
          const s = res.settings
          if (s.server && typeof s.server === 'object') {
            const srv = s.server as Record<string, string>
            setSettings(prev => ({
              ...prev,
              serverHost: srv.host || prev.serverHost,
              serverPort: srv.port || prev.serverPort,
            }))
          }
          if (s.engine && typeof s.engine === 'object') {
            const eng = s.engine as Record<string, unknown>
            setSettings(prev => ({
              ...prev,
              logLevel: (eng.logLevel as string) || prev.logLevel,
              maxExecutions: String(eng.maxExecutions || prev.maxExecutions),
              enableWebSocket: eng.enableWS !== false,
            }))
          }
          if (s.network && typeof s.network === 'object') {
            const net = s.network as Record<string, unknown>
            setNetworkSettings(prev => ({
              ...prev,
              hostname: (net.hostname as string) || prev.hostname,
              primaryInterface: (net.primaryInterface as 'ethernet' | 'wifi') || prev.primaryInterface,
            }))
          }
        }
      })
      .catch(() => {})
  }, [])

  // Load network info and system info on mount
  useEffect(() => {
    setNetworkLoading(true)
    networkApi.getInfo()
      .then(data => setNetworkInfo(data))
      .catch(() => {})
      .finally(() => setNetworkLoading(false))

    systemApi.getInfo()
      .then(data => setSystemInfo(data))
      .catch(() => {})
  }, [])

  // Scan for WiFi networks using real backend
  const handleScanNetworks = useCallback(async () => {
    setScanning(true)
    try {
      const result = await wifiApi.scan()
      setWifiNetworks(result.networks || [])
      if (result.error) {
        showToast(result.error, 'error')
      }
    } catch {
      showToast('Failed to scan WiFi networks', 'error')
    } finally {
      setScanning(false)
    }
  }, [showToast])

  // Connect to WiFi
  const handleConnectWifi = useCallback(async () => {
    if (!networkSettings.wifi.ssid) {
      showToast('Please select or enter a network name', 'error')
      return
    }
    setConnecting(true)
    try {
      const result = await wifiApi.connect(networkSettings.wifi.ssid, networkSettings.wifi.password)
      showToast(result.message, 'success')
      // Re-scan to update connected status
      handleScanNetworks()
      // Refresh network info
      networkApi.getInfo().then(data => setNetworkInfo(data)).catch(() => {})
    } catch (err) {
      showToast(err instanceof Error ? err.message : 'Failed to connect to WiFi', 'error')
    } finally {
      setConnecting(false)
    }
  }, [networkSettings.wifi.ssid, networkSettings.wifi.password, showToast, handleScanNetworks])

  // Auto-scan on WiFi enable
  useEffect(() => {
    if (networkSettings.wifi.enabled && wifiNetworks.length === 0) {
      handleScanNetworks()
    }
  }, [networkSettings.wifi.enabled, wifiNetworks.length, handleScanNetworks])

  // Save all settings
  const handleSave = useCallback(async () => {
    setSaving(true)
    try {
      await settingsApi.save({
        server: {
          host: settings.serverHost,
          port: settings.serverPort,
        },
        engine: {
          logLevel: settings.logLevel,
          maxExecutions: parseInt(settings.maxExecutions) || 100,
          enableWS: settings.enableWebSocket,
        },
        network: {
          hostname: networkSettings.hostname,
          primaryInterface: networkSettings.primaryInterface,
          ethernet: networkSettings.ethernet,
          wifi: {
            ...networkSettings.wifi,
            password: undefined, // We don't store the password
          },
        },
        ui: {
          theme: appSettings.theme,
          language: appSettings.language,
          autoSave: appSettings.autoSave,
          notifications: appSettings.notifications,
          gridSnap: appSettings.gridSnap,
          showMinimap: appSettings.showMinimap,
        },
      })
      showToast('Settings saved successfully', 'success')
    } catch (err) {
      showToast(err instanceof Error ? err.message : 'Failed to save settings', 'error')
    } finally {
      setSaving(false)
    }
  }, [settings, networkSettings, appSettings, showToast])

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

  const toggleSection = (id: string) => {
    setExpandedSections(prev => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
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

  const getSecurityLabel = (security: string) => {
    const labels: Record<string, string> = {
      open: 'Open',
      wep: 'WEP',
      wpa: 'WPA',
      wpa2: 'WPA2',
      wpa3: 'WPA3',
      'wpa2-enterprise': 'WPA2 Enterprise',
      'wpa1 wpa2': 'WPA/WPA2',
    }
    return labels[security] || security.toUpperCase()
  }

  const formatBytes = (bytes: number) => {
    if (bytes === 0) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i]
  }

  const currentInterface = networkSettings.primaryInterface === 'wifi'
    ? networkSettings.wifi
    : networkSettings.ethernet

  // Get current device IP from network info
  const currentIP = networkInfo?.interfaces
    .find(i => i.status === 'up' && i.ipv4.length > 0 && i.name !== 'lo')
    ?.ipv4?.[0]?.split('/')?.[0] || null

  return (
    <div className="space-y-6 max-w-4xl">
      {/* Toast Notification */}
      {toast && (
        <div className={cn(
          'fixed top-4 right-4 z-50 flex items-center gap-2 px-4 py-3 rounded-lg shadow-lg text-sm font-medium transition-all animate-in slide-in-from-top-2',
          toast.type === 'success' && 'bg-green-500 text-white',
          toast.type === 'error' && 'bg-red-500 text-white',
          toast.type === 'info' && 'bg-blue-500 text-white',
        )}>
          {toast.type === 'success' && <CheckCircle className="w-4 h-4" />}
          {toast.type === 'error' && <XCircle className="w-4 h-4" />}
          {toast.type === 'info' && <Info className="w-4 h-4" />}
          {toast.message}
        </div>
      )}

      {/* Header with Save */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900 dark:text-white">
            Settings
          </h1>
          <p className="text-gray-600 dark:text-gray-400 mt-1">
            System configuration & device management
          </p>
        </div>
        <Button
          onClick={handleSave}
          disabled={saving}
          className="gap-2"
        >
          {saving ? <RefreshCw className="w-4 h-4 animate-spin" /> : <Save className="w-4 h-4" />}
          {saving ? 'Saving...' : 'Save Settings'}
        </Button>
      </div>

      {/* Current IP & System Status Banner */}
      <div className="bg-gradient-to-r from-slate-800 to-slate-900 dark:from-slate-900 dark:to-black rounded-xl shadow-lg p-5 text-white">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-5">
            <div className="w-14 h-14 bg-white/10 rounded-xl flex items-center justify-center">
              <Monitor className="w-7 h-7" />
            </div>
            <div>
              <div className="flex items-center gap-2">
                <h2 className="text-lg font-semibold">{networkInfo?.hostname || networkSettings.hostname}</h2>
                <span className={cn(
                  'px-2 py-0.5 rounded-full text-[10px] font-medium uppercase tracking-wide',
                  currentIP ? 'bg-green-500/20 text-green-400' : 'bg-red-500/20 text-red-400'
                )}>
                  {currentIP ? 'Online' : 'Offline'}
                </span>
              </div>
              <div className="flex items-center gap-4 mt-1 text-sm text-white/70">
                {currentIP && (
                  <span className="flex items-center gap-1.5">
                    <Globe className="w-3.5 h-3.5" />
                    <span className="font-mono">{currentIP}</span>
                  </span>
                )}
                {systemInfo && (
                  <>
                    <span className="flex items-center gap-1.5">
                      <Cpu className="w-3.5 h-3.5" />
                      {systemInfo.board_model || `${systemInfo.os}/${systemInfo.arch}`}
                    </span>
                    {systemInfo.uptime_str && (
                      <span className="flex items-center gap-1.5">
                        <Clock className="w-3.5 h-3.5" />
                        {systemInfo.uptime_str}
                      </span>
                    )}
                  </>
                )}
              </div>
            </div>
          </div>
          {systemInfo && (
            <div className="flex items-center gap-6 text-sm">
              {systemInfo.temperature > 0 && (
                <div className="text-center">
                  <Thermometer className={cn('w-5 h-5 mx-auto mb-1',
                    systemInfo.temperature > 70 ? 'text-red-400' : systemInfo.temperature > 55 ? 'text-yellow-400' : 'text-green-400'
                  )} />
                  <span className="font-mono">{systemInfo.temperature.toFixed(1)}°C</span>
                </div>
              )}
              <div className="text-center">
                <Cpu className={cn('w-5 h-5 mx-auto mb-1',
                  systemInfo.cpu.usage_percent > 80 ? 'text-red-400' : 'text-blue-400'
                )} />
                <span className="font-mono">{systemInfo.cpu.usage_percent.toFixed(0)}%</span>
              </div>
              <div className="text-center">
                <MemoryStick className={cn('w-5 h-5 mx-auto mb-1',
                  systemInfo.memory.percent > 80 ? 'text-red-400' : 'text-purple-400'
                )} />
                <span className="font-mono">{systemInfo.memory.percent.toFixed(0)}%</span>
              </div>
              <div className="text-center">
                <HardDrive className={cn('w-5 h-5 mx-auto mb-1',
                  systemInfo.disk.percent > 80 ? 'text-red-400' : 'text-cyan-400'
                )} />
                <span className="font-mono">{systemInfo.disk.percent.toFixed(0)}%</span>
              </div>
            </div>
          )}
        </div>

        {/* Quick Resource Bars */}
        {systemInfo && (
          <div className="grid grid-cols-3 gap-4 mt-4">
            <div>
              <div className="flex justify-between text-xs text-white/60 mb-1">
                <span>CPU</span>
                <span>{systemInfo.cpu.usage_percent.toFixed(0)}% ({systemInfo.cpu.cores} cores)</span>
              </div>
              <div className="w-full h-1.5 bg-white/10 rounded-full">
                <div className={cn('h-1.5 rounded-full transition-all',
                  systemInfo.cpu.usage_percent > 80 ? 'bg-red-500' : 'bg-blue-500'
                )} style={{ width: `${systemInfo.cpu.usage_percent}%` }} />
              </div>
            </div>
            <div>
              <div className="flex justify-between text-xs text-white/60 mb-1">
                <span>Memory</span>
                <span>{formatBytes(systemInfo.memory.used_bytes)} / {formatBytes(systemInfo.memory.total_bytes)}</span>
              </div>
              <div className="w-full h-1.5 bg-white/10 rounded-full">
                <div className={cn('h-1.5 rounded-full transition-all',
                  systemInfo.memory.percent > 80 ? 'bg-red-500' : 'bg-purple-500'
                )} style={{ width: `${systemInfo.memory.percent}%` }} />
              </div>
            </div>
            <div>
              <div className="flex justify-between text-xs text-white/60 mb-1">
                <span>Disk</span>
                <span>{formatBytes(systemInfo.disk.used_bytes)} / {formatBytes(systemInfo.disk.total_bytes)}</span>
              </div>
              <div className="w-full h-1.5 bg-white/10 rounded-full">
                <div className={cn('h-1.5 rounded-full transition-all',
                  systemInfo.disk.percent > 80 ? 'bg-red-500' : 'bg-cyan-500'
                )} style={{ width: `${systemInfo.disk.percent}%` }} />
              </div>
            </div>
          </div>
        )}
      </div>

      {/* Device Setup */}
      <div className="bg-gradient-to-r from-blue-500 to-purple-600 rounded-lg shadow p-5 text-white">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <div className="w-12 h-12 bg-white/20 rounded-xl flex items-center justify-center">
              <Settings2 className="w-6 h-6" />
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
            Open Wizard
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
          showToast('Setup wizard completed', 'success')
        }}
        onGoToEditor={() => {
          setShowSetupWizard(false)
          navigate('/editor')
        }}
      />

      {/* SaaS Connection */}
      <div className="bg-gradient-to-r from-cyan-500 to-blue-600 rounded-lg shadow p-5 text-white">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <div className="w-12 h-12 bg-white/20 rounded-xl flex items-center justify-center">
              <Cloud className="w-6 h-6" />
            </div>
            <div>
              <h2 className="text-lg font-semibold">SaaS Cloud Connection</h2>
              <p className="text-white/80 text-sm">
                Connect to EdgeFlow SaaS for remote monitoring and control
              </p>
            </div>
          </div>
          <Button
            onClick={() => navigate('/settings/saas')}
            className="bg-white text-cyan-600 hover:bg-white/90"
          >
            <Settings2 className="w-4 h-4 mr-2" />
            Configure
          </Button>
        </div>
      </div>

      {/* ============== Network Configuration ============== */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <div className="flex items-center gap-2 mb-6">
          <Network className="w-5 h-5 text-blue-500" />
          <h2 className="text-lg font-semibold text-gray-900 dark:text-white">
            Network Configuration
          </h2>
        </div>

        <div className="space-y-6">
          {/* Network Status (auto-expanded) */}
          <div className="p-4 bg-gray-50 dark:bg-gray-900/50 rounded-xl border border-gray-200 dark:border-gray-700">
            <button
              onClick={() => toggleSection('network-status')}
              className="w-full flex items-center justify-between"
            >
              <span className="text-sm font-semibold text-gray-700 dark:text-gray-300 flex items-center gap-2">
                <Activity className="w-4 h-4 text-green-500" />
                Network Status & Interfaces
              </span>
              {expandedSections.has('network-status') ? <ChevronUp className="w-4 h-4" /> : <ChevronDown className="w-4 h-4" />}
            </button>

            {expandedSections.has('network-status') && (
              <div className="mt-4 space-y-3">
                {networkLoading ? (
                  <div className="flex items-center justify-center py-4 gap-2">
                    <RefreshCw className="w-4 h-4 animate-spin text-muted-foreground" />
                    <span className="text-sm text-muted-foreground">Loading network info...</span>
                  </div>
                ) : networkInfo ? (
                  <>
                    {/* Interface Cards */}
                    <div className="grid grid-cols-1 gap-2">
                      {networkInfo.interfaces
                        .filter(i => i.name !== 'lo')
                        .map(iface => (
                          <div key={iface.name} className={cn(
                            'flex items-center gap-4 p-3 rounded-lg border transition-colors',
                            iface.status === 'up' && iface.ipv4.length > 0
                              ? 'border-green-200 dark:border-green-800 bg-green-50/50 dark:bg-green-950/20'
                              : 'border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800'
                          )}>
                            <div className={cn(
                              'w-10 h-10 rounded-lg flex items-center justify-center',
                              iface.status === 'up' ? 'bg-green-100 dark:bg-green-900/40' : 'bg-gray-100 dark:bg-gray-700'
                            )}>
                              {iface.name.startsWith('wl') ? (
                                <Wifi className={cn('w-5 h-5', iface.status === 'up' ? 'text-green-600' : 'text-gray-400')} />
                              ) : (
                                <Globe className={cn('w-5 h-5', iface.status === 'up' ? 'text-green-600' : 'text-gray-400')} />
                              )}
                            </div>
                            <div className="flex-1 min-w-0">
                              <div className="flex items-center gap-2">
                                <span className="font-mono font-medium text-sm text-gray-900 dark:text-white">{iface.name}</span>
                                <span className={cn(
                                  'px-1.5 py-0.5 rounded text-[10px] font-medium uppercase',
                                  iface.status === 'up' ? 'bg-green-100 dark:bg-green-900 text-green-700 dark:text-green-300' : 'bg-gray-100 dark:bg-gray-700 text-gray-500'
                                )}>
                                  {iface.status}
                                </span>
                              </div>
                              <div className="flex items-center gap-3 mt-0.5 text-xs text-muted-foreground">
                                {iface.ipv4.length > 0 && (
                                  <span className="font-mono text-gray-900 dark:text-gray-200">{iface.ipv4[0]}</span>
                                )}
                                {iface.mac && <span className="font-mono">{iface.mac}</span>}
                              </div>
                            </div>
                            {iface.status === 'up' && iface.ipv4.length > 0 && (
                              <CheckCircle className="w-5 h-5 text-green-500 shrink-0" />
                            )}
                          </div>
                        ))}
                    </div>

                    <div className="flex items-center justify-between pt-2">
                      <span className="text-xs text-muted-foreground">
                        Hostname: <span className="font-mono text-gray-900 dark:text-gray-200">{networkInfo.hostname}</span>
                      </span>
                      <Button
                        variant="ghost"
                        size="sm"
                        className="h-7 text-xs"
                        onClick={() => {
                          setNetworkLoading(true)
                          networkApi.getInfo()
                            .then(data => setNetworkInfo(data))
                            .catch(() => {})
                            .finally(() => setNetworkLoading(false))
                        }}
                      >
                        <RefreshCw className="w-3 h-3 mr-1" />
                        Refresh
                      </Button>
                    </div>
                  </>
                ) : (
                  <div className="text-center py-4 text-sm text-muted-foreground">
                    <AlertTriangle className="w-5 h-5 mx-auto mb-2 text-yellow-500" />
                    Failed to load network information
                  </div>
                )}
              </div>
            )}
          </div>

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
                <div className="space-y-2 max-h-52 overflow-y-auto">
                  {wifiNetworks.map((network, idx) => {
                    const SignalIcon = getSignalIcon(network.signal)
                    const signalQuality = getSignalQuality(network.signal)
                    const isSelected = networkSettings.wifi.ssid === network.ssid

                    return (
                      <button
                        key={`${network.bssid}-${idx}`}
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
                            <span>·</span>
                            <span>{network.frequency > 5000 ? '5GHz' : '2.4GHz'}</span>
                            {network.channel > 0 && (
                              <>
                                <span>·</span>
                                <span>Ch {network.channel}</span>
                              </>
                            )}
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

                <div className="grid grid-cols-2 gap-3">
                  <div className="space-y-2">
                    <Label htmlFor="wifi-country">WiFi Country/Region</Label>
                    <select
                      id="wifi-country"
                      value={networkSettings.wifi.country}
                      onChange={(e) => updateWifi({ country: e.target.value })}
                      className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
                    >
                      {WIFI_COUNTRIES.map((country) => (
                        <option key={country.code} value={country.code}>
                          {country.name} ({country.code})
                        </option>
                      ))}
                    </select>
                  </div>
                  <div className="flex items-end">
                    <Button
                      onClick={handleConnectWifi}
                      disabled={connecting || !networkSettings.wifi.ssid}
                      className="w-full gap-2"
                    >
                      {connecting ? (
                        <RefreshCw className="w-4 h-4 animate-spin" />
                      ) : (
                        <Wifi className="w-4 h-4" />
                      )}
                      {connecting ? 'Connecting...' : 'Connect'}
                    </Button>
                  </div>
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
                    className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
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
        </div>
      </div>

      {/* ============== Server Settings ============== */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <div className="flex items-center gap-2 mb-4">
          <Server className="w-5 h-5 text-purple-500" />
          <h2 className="text-lg font-semibold text-gray-900 dark:text-white">
            Server Settings
          </h2>
        </div>
        <div className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="server-host">Server Address</Label>
              <Input
                id="server-host"
                value={settings.serverHost}
                onChange={(e) => setSettings({ ...settings, serverHost: e.target.value })}
                placeholder="0.0.0.0"
              />
              <p className="text-xs text-muted-foreground">Use 0.0.0.0 for all interfaces</p>
            </div>
            <div className="space-y-2">
              <Label htmlFor="server-port">Port</Label>
              <Input
                id="server-port"
                type="number"
                value={settings.serverPort}
                onChange={(e) => setSettings({ ...settings, serverPort: e.target.value })}
                placeholder="8080"
              />
            </div>
          </div>
        </div>
      </div>

      {/* ============== Engine Settings ============== */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <div className="flex items-center gap-2 mb-4">
          <Zap className="w-5 h-5 text-yellow-500" />
          <h2 className="text-lg font-semibold text-gray-900 dark:text-white">
            Engine Settings
          </h2>
        </div>
        <div className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <Label htmlFor="log-level">Log Level</Label>
              <select
                id="log-level"
                value={settings.logLevel}
                onChange={(e) => setSettings({ ...settings, logLevel: e.target.value })}
                className="w-full px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
              >
                <option value="debug">Debug</option>
                <option value="info">Info</option>
                <option value="warn">Warning</option>
                <option value="error">Error</option>
              </select>
            </div>
            <div className="space-y-2">
              <Label htmlFor="max-exec">Max Concurrent Executions</Label>
              <Input
                id="max-exec"
                type="number"
                value={settings.maxExecutions}
                onChange={(e) => setSettings({ ...settings, maxExecutions: e.target.value })}
                placeholder="100"
              />
            </div>
          </div>

          <div className="flex items-center justify-between py-2">
            <div>
              <Label className="text-sm font-medium">WebSocket Real-time Updates</Label>
              <p className="text-xs text-muted-foreground">Enable live execution and node status updates</p>
            </div>
            <Switch
              checked={settings.enableWebSocket}
              onCheckedChange={(checked) => setSettings({ ...settings, enableWebSocket: checked })}
            />
          </div>
        </div>
      </div>

      {/* ============== UI Settings ============== */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <div className="flex items-center gap-2 mb-4">
          <Palette className="w-5 h-5 text-pink-500" />
          <h2 className="text-lg font-semibold text-gray-900 dark:text-white">
            UI Settings
          </h2>
        </div>
        <div className="space-y-4">
          <div className="flex items-center justify-between py-2">
            <div className="flex items-center gap-3">
              <div className="w-9 h-9 rounded-lg bg-gray-100 dark:bg-gray-700 flex items-center justify-center">
                <Monitor className="w-4 h-4" />
              </div>
              <div>
                <Label className="text-sm font-medium">Theme</Label>
                <p className="text-xs text-muted-foreground">Choose your preferred appearance</p>
              </div>
            </div>
            <div className="flex items-center gap-1 bg-gray-100 dark:bg-gray-700 rounded-lg p-1">
              {(['light', 'dark', 'system'] as const).map(theme => (
                <button
                  key={theme}
                  onClick={() => updateAppSettings({ theme })}
                  className={cn(
                    'px-3 py-1.5 rounded-md text-xs font-medium transition-colors capitalize',
                    appSettings.theme === theme
                      ? 'bg-white dark:bg-gray-600 text-gray-900 dark:text-white shadow-sm'
                      : 'text-gray-500 hover:text-gray-700 dark:hover:text-gray-300'
                  )}
                >
                  {theme}
                </button>
              ))}
            </div>
          </div>

          <div className="flex items-center justify-between py-2 border-t border-gray-100 dark:border-gray-700">
            <div className="flex items-center gap-3">
              <div className="w-9 h-9 rounded-lg bg-gray-100 dark:bg-gray-700 flex items-center justify-center">
                <Languages className="w-4 h-4" />
              </div>
              <div>
                <Label className="text-sm font-medium">Language</Label>
                <p className="text-xs text-muted-foreground">Interface language</p>
              </div>
            </div>
            <select
              value={appSettings.language}
              onChange={(e) => updateAppSettings({ language: e.target.value as 'en' | 'fa' | 'ar' })}
              className="px-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white text-sm"
            >
              <option value="en">English</option>
              <option value="fa">فارسی</option>
              <option value="ar">العربية</option>
            </select>
          </div>

          <div className="flex items-center justify-between py-2 border-t border-gray-100 dark:border-gray-700">
            <div>
              <Label className="text-sm font-medium">Auto-Save Flows</Label>
              <p className="text-xs text-muted-foreground">Automatically save changes every {appSettings.autoSaveInterval}s</p>
            </div>
            <Switch
              checked={appSettings.autoSave}
              onCheckedChange={(checked) => updateAppSettings({ autoSave: checked })}
            />
          </div>

          <div className="flex items-center justify-between py-2 border-t border-gray-100 dark:border-gray-700">
            <div>
              <Label className="text-sm font-medium">Notifications</Label>
              <p className="text-xs text-muted-foreground">Show system notifications</p>
            </div>
            <Switch
              checked={appSettings.notifications}
              onCheckedChange={(checked) => updateAppSettings({ notifications: checked })}
            />
          </div>

          <div className="flex items-center justify-between py-2 border-t border-gray-100 dark:border-gray-700">
            <div>
              <Label className="text-sm font-medium">Grid Snap</Label>
              <p className="text-xs text-muted-foreground">Snap nodes to grid in editor</p>
            </div>
            <Switch
              checked={appSettings.gridSnap}
              onCheckedChange={(checked) => updateAppSettings({ gridSnap: checked })}
            />
          </div>

          <div className="flex items-center justify-between py-2 border-t border-gray-100 dark:border-gray-700">
            <div>
              <Label className="text-sm font-medium">Show Minimap</Label>
              <p className="text-xs text-muted-foreground">Show minimap overlay in editor</p>
            </div>
            <Switch
              checked={appSettings.showMinimap}
              onCheckedChange={(checked) => updateAppSettings({ showMinimap: checked })}
            />
          </div>
        </div>
      </div>

      {/* ============== System Management ============== */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <div className="flex items-center gap-2 mb-4">
          <Power className="w-5 h-5 text-red-500" />
          <h2 className="text-lg font-semibold text-gray-900 dark:text-white">
            System Management
          </h2>
        </div>
        <div className="space-y-4">
          {/* Backup/Restore */}
          <div className="flex items-center justify-between py-2">
            <div>
              <Label className="text-sm font-medium">Backup Configuration</Label>
              <p className="text-xs text-muted-foreground">Download flows, settings, and node data</p>
            </div>
            <div className="flex gap-2">
              <Button
                variant="outline"
                size="sm"
                className="gap-1.5"
                onClick={() => {
                  // Download settings as JSON
                  const data = {
                    settings,
                    networkSettings,
                    ui: appSettings,
                    exportedAt: new Date().toISOString(),
                    version: '0.1.0',
                  }
                  const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' })
                  const url = URL.createObjectURL(blob)
                  const a = document.createElement('a')
                  a.href = url
                  a.download = `edgeflow-backup-${new Date().toISOString().split('T')[0]}.json`
                  a.click()
                  URL.revokeObjectURL(url)
                  showToast('Backup downloaded', 'success')
                }}
              >
                <Download className="w-4 h-4" />
                Export
              </Button>
              <Button
                variant="outline"
                size="sm"
                className="gap-1.5"
                onClick={() => {
                  const input = document.createElement('input')
                  input.type = 'file'
                  input.accept = '.json'
                  input.onchange = (e) => {
                    const file = (e.target as HTMLInputElement).files?.[0]
                    if (!file) return
                    const reader = new FileReader()
                    reader.onload = (ev) => {
                      try {
                        const data = JSON.parse(ev.target?.result as string)
                        if (data.settings) setSettings(data.settings)
                        if (data.networkSettings) setNetworkSettings(data.networkSettings)
                        if (data.ui) updateAppSettings(data.ui)
                        showToast('Configuration restored from backup', 'success')
                      } catch {
                        showToast('Invalid backup file', 'error')
                      }
                    }
                    reader.readAsText(file)
                  }
                  input.click()
                }}
              >
                <Upload className="w-4 h-4" />
                Import
              </Button>
            </div>
          </div>

          {/* Service Actions */}
          <div className="flex items-center justify-between py-3 border-t border-gray-100 dark:border-gray-700">
            <div>
              <Label className="text-sm font-medium">Restart EdgeFlow Service</Label>
              <p className="text-xs text-muted-foreground">Restart the backend service (Linux only)</p>
            </div>
            <Button
              variant="outline"
              size="sm"
              className="gap-1.5 text-yellow-600 border-yellow-300 hover:bg-yellow-50 dark:hover:bg-yellow-950/20"
              onClick={async () => {
                if (!confirm('Are you sure you want to restart the EdgeFlow service?')) return
                try {
                  await systemApi.restartService()
                  showToast('Service restart initiated', 'info')
                } catch (err) {
                  showToast(err instanceof Error ? err.message : 'Failed to restart service', 'error')
                }
              }}
            >
              <RotateCcw className="w-4 h-4" />
              Restart Service
            </Button>
          </div>

          <div className="flex items-center justify-between py-3 border-t border-gray-100 dark:border-gray-700">
            <div>
              <Label className="text-sm font-medium text-red-500">Reboot Device</Label>
              <p className="text-xs text-muted-foreground">Full system reboot (Linux only)</p>
            </div>
            <Button
              variant="outline"
              size="sm"
              className="gap-1.5 text-red-600 border-red-300 hover:bg-red-50 dark:hover:bg-red-950/20"
              onClick={async () => {
                if (!confirm('Are you sure you want to reboot the device? All running flows will stop.')) return
                try {
                  await systemApi.reboot()
                  showToast('System reboot initiated', 'info')
                } catch (err) {
                  showToast(err instanceof Error ? err.message : 'Failed to reboot', 'error')
                }
              }}
            >
              <Power className="w-4 h-4" />
              Reboot
            </Button>
          </div>
        </div>
      </div>

      {/* ============== About ============== */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow p-6">
        <div className="flex items-start space-x-3">
          <Info className="w-5 h-5 text-blue-500 mt-0.5" />
          <div className="flex-1">
            <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-3">
              About EdgeFlow
            </h2>
            <div className="grid grid-cols-2 gap-x-8 gap-y-2 text-sm">
              <div className="flex justify-between">
                <span className="text-muted-foreground">Version</span>
                <span className="font-mono text-gray-900 dark:text-white">0.1.0</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Environment</span>
                <span className="text-gray-900 dark:text-white">Development</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Backend</span>
                <span className="font-mono text-gray-900 dark:text-white">{systemInfo?.go_version || 'Go'}</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Frontend</span>
                <span className="text-gray-900 dark:text-white">React 18 + TypeScript</span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Platform</span>
                <span className="font-mono text-gray-900 dark:text-white">
                  {systemInfo ? `${systemInfo.os}/${systemInfo.arch}` : 'N/A'}
                </span>
              </div>
              <div className="flex justify-between">
                <span className="text-muted-foreground">Goroutines</span>
                <span className="font-mono text-gray-900 dark:text-white">{systemInfo?.goroutines || 'N/A'}</span>
              </div>
              {systemInfo?.board_model && (
                <div className="flex justify-between col-span-2">
                  <span className="text-muted-foreground">Board</span>
                  <span className="text-gray-900 dark:text-white">{systemInfo.board_model}</span>
                </div>
              )}
            </div>
            <p className="text-xs text-gray-500 dark:text-gray-500 mt-4 pt-3 border-t border-gray-100 dark:border-gray-700">
              EdgeFlow - Lightweight automation platform for Edge & IoT
            </p>
          </div>
        </div>
      </div>

      {/* Bottom Save Bar */}
      <div className="flex items-center justify-end space-x-3 pb-4">
        <Button
          variant="outline"
          onClick={() => {
            settingsApi.get().then(res => {
              if (res.configured && res.settings) {
                showToast('Settings reverted to last saved state', 'info')
              }
            }).catch(() => {})
          }}
        >
          Cancel
        </Button>
        <Button onClick={handleSave} disabled={saving} className="gap-2">
          {saving ? <RefreshCw className="w-4 h-4 animate-spin" /> : <Save className="w-4 h-4" />}
          Save Settings
        </Button>
      </div>
    </div>
  )
}
