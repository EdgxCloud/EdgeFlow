import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Alert, AlertDescription } from '@/components/ui/alert'
import {
  Cloud,
  Wifi,
  WifiOff,
  CheckCircle2,
  XCircle,
  ArrowLeft,
  RefreshCw,
  Key,
  Server,
  Shield,
} from 'lucide-react'
import { toast } from 'sonner'

interface SaaSConfig {
  enabled: boolean
  server_url: string
  device_id: string
  api_key: string
  provisioning_code: string
  enable_tls: boolean
  is_provisioned: boolean
  is_connected: boolean
}

interface SaaSStatus {
  connected: boolean
  device_id: string
  last_heartbeat: string
  uptime: string
}

export default function SaaSSettings() {
  const navigate = useNavigate()
  const [config, setConfig] = useState<SaaSConfig>({
    enabled: false,
    server_url: 'api.edgx.cloud',
    device_id: '',
    api_key: '',
    provisioning_code: '',
    enable_tls: true,
    is_provisioned: false,
    is_connected: false,
  })
  const [status, setStatus] = useState<SaaSStatus | null>(null)
  const [loading, setLoading] = useState(false)
  const [showApiKey, setShowApiKey] = useState(false)

  useEffect(() => {
    loadConfig()
    loadStatus()
    const interval = setInterval(loadStatus, 10000) // Refresh every 10s
    return () => clearInterval(interval)
  }, [])

  const loadConfig = async () => {
    try {
      const response = await fetch('/api/v1/saas/config')
      if (response.ok) {
        const data = await response.json()
        setConfig(data)
      }
    } catch (error) {
      console.error('Failed to load SaaS config:', error)
    }
  }

  const loadStatus = async () => {
    try {
      const response = await fetch('/api/v1/saas/status')
      if (response.ok) {
        const data = await response.json()
        setStatus(data)
      }
    } catch (error) {
      console.error('Failed to load SaaS status:', error)
    }
  }

  const handleSave = async () => {
    setLoading(true)
    try {
      const response = await fetch('/api/v1/saas/config', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(config),
      })

      if (!response.ok) {
        const error = await response.text()
        throw new Error(error)
      }

      toast.success('SaaS configuration saved')
      await loadConfig()
      await loadStatus()
    } catch (error) {
      toast.error(`Failed to save: ${error}`)
    } finally {
      setLoading(false)
    }
  }

  const handleProvision = async () => {
    if (!config.provisioning_code) {
      toast.error('Provisioning code is required')
      return
    }

    setLoading(true)
    try {
      const response = await fetch('/api/v1/saas/provision', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          provisioning_code: config.provisioning_code,
        }),
      })

      if (!response.ok) {
        const error = await response.text()
        throw new Error(error)
      }

      const result = await response.json()
      toast.success('Device provisioned successfully!')

      // Update config with credentials
      setConfig({
        ...config,
        device_id: result.device_id,
        api_key: result.api_key,
        provisioning_code: '',
        is_provisioned: true,
        enabled: true,
      })

      // Save and connect
      await handleSave()
    } catch (error) {
      toast.error(`Provisioning failed: ${error}`)
    } finally {
      setLoading(false)
    }
  }

  const handleConnect = async () => {
    setLoading(true)
    try {
      const response = await fetch('/api/v1/saas/connect', {
        method: 'POST',
      })

      if (!response.ok) {
        const error = await response.text()
        throw new Error(error)
      }

      toast.success('Connecting to SaaS...')
      setTimeout(loadStatus, 2000) // Refresh status after 2s
    } catch (error) {
      toast.error(`Connection failed: ${error}`)
    } finally {
      setLoading(false)
    }
  }

  const handleDisconnect = async () => {
    setLoading(true)
    try {
      const response = await fetch('/api/v1/saas/disconnect', {
        method: 'POST',
      })

      if (!response.ok) {
        const error = await response.text()
        throw new Error(error)
      }

      toast.success('Disconnected from SaaS')
      await loadStatus()
    } catch (error) {
      toast.error(`Disconnect failed: ${error}`)
    } finally {
      setLoading(false)
    }
  }

  const maskApiKey = (key: string) => {
    if (!key || key.length < 20) return key
    return key.substring(0, 12) + '•'.repeat(20) + key.substring(key.length - 4)
  }

  return (
    <div className="min-h-screen bg-background p-6">
      <div className="max-w-4xl mx-auto space-y-6">
        {/* Header */}
        <div className="flex items-center gap-4">
          <Button variant="ghost" size="sm" onClick={() => navigate('/settings')}>
            <ArrowLeft className="w-4 h-4 mr-2" />
            Back
          </Button>
          <div className="flex-1">
            <h1 className="text-2xl font-bold flex items-center gap-2">
              <Cloud className="w-6 h-6" />
              SaaS Connection
            </h1>
            <p className="text-sm text-muted-foreground">
              Connect to EdgeFlow SaaS platform for remote monitoring and control
            </p>
          </div>
        </div>

        {/* Status Card */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center justify-between">
              <span>Connection Status</span>
              <Button variant="ghost" size="sm" onClick={loadStatus}>
                <RefreshCw className="w-4 h-4" />
              </Button>
            </CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <div className="flex items-center gap-2">
                {status?.connected ? (
                  <>
                    <Wifi className="w-5 h-5 text-green-500" />
                    <span className="font-medium">Connected</span>
                  </>
                ) : (
                  <>
                    <WifiOff className="w-5 h-5 text-gray-400" />
                    <span className="font-medium text-muted-foreground">Disconnected</span>
                  </>
                )}
              </div>
              <Badge variant={status?.connected ? 'default' : 'secondary'}>
                {status?.connected ? 'Online' : 'Offline'}
              </Badge>
            </div>

            {status?.connected && (
              <>
                <div className="grid grid-cols-2 gap-4 pt-4 border-t">
                  <div>
                    <p className="text-xs text-muted-foreground">Device ID</p>
                    <p className="text-sm font-mono">{status.device_id.substring(0, 16)}...</p>
                  </div>
                  <div>
                    <p className="text-xs text-muted-foreground">Uptime</p>
                    <p className="text-sm font-medium">{status.uptime || 'N/A'}</p>
                  </div>
                  <div className="col-span-2">
                    <p className="text-xs text-muted-foreground">Last Heartbeat</p>
                    <p className="text-sm">{status.last_heartbeat || 'N/A'}</p>
                  </div>
                </div>
              </>
            )}
          </CardContent>
        </Card>

        {/* Provisioning Card (if not provisioned) */}
        {!config.is_provisioned && (
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Key className="w-5 h-5" />
                Device Provisioning
              </CardTitle>
              <CardDescription>
                Register this device with the EdgeFlow SaaS platform using a provisioning code
              </CardDescription>
            </CardHeader>
            <CardContent className="space-y-4">
              <Alert>
                <Shield className="w-4 h-4" />
                <AlertDescription>
                  <strong>New Device Setup:</strong>
                  <ol className="list-decimal list-inside mt-2 space-y-1 text-sm">
                    <li>Create a device in the SaaS admin panel</li>
                    <li>Copy the provisioning code (EDGE-XXXX-XXXX-XXXX)</li>
                    <li>Enter the code below and click Provision</li>
                    <li>Your device will receive credentials automatically</li>
                  </ol>
                </AlertDescription>
              </Alert>

              <div className="space-y-2">
                <label className="text-sm font-medium">Provisioning Code</label>
                <Input
                  placeholder="EDGE-XXXX-XXXX-XXXX"
                  value={config.provisioning_code}
                  onChange={(e) => setConfig({ ...config, provisioning_code: e.target.value })}
                  className="font-mono"
                />
                <p className="text-xs text-muted-foreground">
                  Provisioning codes expire after 24 hours
                </p>
              </div>

              <Button
                onClick={handleProvision}
                disabled={loading || !config.provisioning_code}
                className="w-full"
              >
                {loading ? (
                  <>
                    <RefreshCw className="w-4 h-4 mr-2 animate-spin" />
                    Provisioning...
                  </>
                ) : (
                  <>
                    <CheckCircle2 className="w-4 h-4 mr-2" />
                    Provision Device
                  </>
                )}
              </Button>
            </CardContent>
          </Card>
        )}

        {/* Configuration Card */}
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Server className="w-5 h-5" />
              Configuration
            </CardTitle>
            <CardDescription>
              Configure connection to EdgeFlow SaaS platform
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            {/* Enable/Disable */}
            <div className="flex items-center justify-between">
              <div>
                <p className="font-medium">Enable SaaS Connection</p>
                <p className="text-sm text-muted-foreground">
                  Connect this device to the cloud platform
                </p>
              </div>
              <Button
                variant={config.enabled ? 'default' : 'outline'}
                size="sm"
                onClick={() => setConfig({ ...config, enabled: !config.enabled })}
              >
                {config.enabled ? 'Enabled' : 'Disabled'}
              </Button>
            </div>

            {config.enabled && (
              <>
                {/* Server URL */}
                <div className="space-y-2">
                  <label className="text-sm font-medium">Server URL</label>
                  <Input
                    placeholder="api.edgx.cloud"
                    value={config.server_url}
                    onChange={(e) => setConfig({ ...config, server_url: e.target.value })}
                  />
                  <p className="text-xs text-muted-foreground">
                    Domain name or IP address (without http:// or /api/v1)
                  </p>
                </div>

                {/* TLS */}
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-sm font-medium">Use TLS (wss://)</p>
                    <p className="text-xs text-muted-foreground">
                      Secure WebSocket connection (recommended for production)
                    </p>
                  </div>
                  <Button
                    variant={config.enable_tls ? 'default' : 'outline'}
                    size="sm"
                    onClick={() => setConfig({ ...config, enable_tls: !config.enable_tls })}
                  >
                    {config.enable_tls ? 'Enabled' : 'Disabled'}
                  </Button>
                </div>

                {/* Device ID (read-only if provisioned) */}
                {config.is_provisioned && (
                  <div className="space-y-2">
                    <label className="text-sm font-medium">Device ID</label>
                    <Input
                      value={config.device_id}
                      readOnly
                      className="font-mono bg-muted"
                    />
                  </div>
                )}

                {/* API Key (masked) */}
                {config.is_provisioned && (
                  <div className="space-y-2">
                    <div className="flex items-center justify-between">
                      <label className="text-sm font-medium">API Key</label>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => setShowApiKey(!showApiKey)}
                      >
                        {showApiKey ? 'Hide' : 'Show'}
                      </Button>
                    </div>
                    <Input
                      value={showApiKey ? config.api_key : maskApiKey(config.api_key)}
                      readOnly
                      className="font-mono bg-muted"
                      type={showApiKey ? 'text' : 'password'}
                    />
                    <p className="text-xs text-muted-foreground">
                      Keep this key secure. Never share it publicly.
                    </p>
                  </div>
                )}
              </>
            )}

            {/* Action Buttons */}
            <div className="flex gap-2 pt-4 border-t">
              <Button onClick={handleSave} disabled={loading} className="flex-1">
                {loading ? (
                  <RefreshCw className="w-4 h-4 mr-2 animate-spin" />
                ) : (
                  <CheckCircle2 className="w-4 h-4 mr-2" />
                )}
                Save Configuration
              </Button>

              {config.is_provisioned && config.enabled && (
                <>
                  {!status?.connected ? (
                    <Button onClick={handleConnect} disabled={loading} variant="default">
                      <Wifi className="w-4 h-4 mr-2" />
                      Connect
                    </Button>
                  ) : (
                    <Button onClick={handleDisconnect} disabled={loading} variant="destructive">
                      <XCircle className="w-4 h-4 mr-2" />
                      Disconnect
                    </Button>
                  )}
                </>
              )}
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
