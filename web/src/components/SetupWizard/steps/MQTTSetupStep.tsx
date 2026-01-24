/**
 * MQTT Setup Step
 *
 * Configure MQTT broker connection
 */

import { useState } from 'react'
import { MessageSquare, Server, Lock, Eye, EyeOff, Shield, ExternalLink } from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import type { MQTTConfig } from '../types'

interface MQTTSetupStepProps {
  config: MQTTConfig
  onChange: (config: MQTTConfig) => void
}

export function MQTTSetupStep({ config, onChange }: MQTTSetupStepProps) {
  const [showPassword, setShowPassword] = useState(false)
  const [testing, setTesting] = useState(false)

  const updateConfig = (updates: Partial<MQTTConfig>) => {
    onChange({ ...config, ...updates })
  }

  const handleTestConnection = () => {
    setTesting(true)
    // Simulate connection test
    setTimeout(() => {
      setTesting(false)
    }, 2000)
  }

  return (
    <div className="space-y-6">
      <div className="text-center space-y-2">
        <h2 className="text-2xl font-bold text-gray-900 dark:text-white">
          MQTT Broker Setup
        </h2>
        <p className="text-muted-foreground">
          Configure the MQTT broker for IoT device communication
        </p>
      </div>

      {/* Enable MQTT */}
      <div className="flex items-center justify-between p-4 bg-gray-50 dark:bg-gray-800/50 rounded-xl">
        <div className="flex items-center gap-3">
          <div className="w-10 h-10 bg-purple-500 rounded-lg flex items-center justify-center">
            <MessageSquare className="w-5 h-5 text-white" />
          </div>
          <div>
            <Label className="font-medium">Enable MQTT</Label>
            <p className="text-xs text-muted-foreground mt-1">
              Required for most IoT device communication
            </p>
          </div>
        </div>
        <Switch
          checked={config.enabled}
          onCheckedChange={(checked) => updateConfig({ enabled: checked })}
        />
      </div>

      {config.enabled && (
        <>
          {/* Broker Type */}
          <div className="space-y-4">
            <Label className="text-sm font-semibold">Broker Type</Label>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
              <button
                onClick={() => updateConfig({ useBuiltIn: true })}
                className={cn(
                  'p-4 rounded-xl border-2 transition-all text-left',
                  config.useBuiltIn
                    ? 'border-primary bg-primary/5'
                    : 'border-gray-200 dark:border-gray-700 hover:border-gray-300'
                )}
              >
                <div className="flex items-start gap-3">
                  <Server className="w-6 h-6 text-green-500 mt-0.5" />
                  <div>
                    <div className="font-medium">Built-in Broker</div>
                    <div className="text-xs text-muted-foreground mt-1">
                      EdgeFlow includes Mosquitto broker. Perfect for standalone
                      deployments.
                    </div>
                    <div className="mt-2 text-xs text-green-600 dark:text-green-400">
                      Recommended for beginners
                    </div>
                  </div>
                </div>
              </button>

              <button
                onClick={() => updateConfig({ useBuiltIn: false })}
                className={cn(
                  'p-4 rounded-xl border-2 transition-all text-left',
                  !config.useBuiltIn
                    ? 'border-primary bg-primary/5'
                    : 'border-gray-200 dark:border-gray-700 hover:border-gray-300'
                )}
              >
                <div className="flex items-start gap-3">
                  <ExternalLink className="w-6 h-6 text-blue-500 mt-0.5" />
                  <div>
                    <div className="font-medium">External Broker</div>
                    <div className="text-xs text-muted-foreground mt-1">
                      Connect to HiveMQ, EMQX, AWS IoT, or your existing broker.
                    </div>
                    <div className="mt-2 text-xs text-blue-600 dark:text-blue-400">
                      For advanced setups
                    </div>
                  </div>
                </div>
              </button>
            </div>
          </div>

          {/* Built-in Broker Settings */}
          {config.useBuiltIn && (
            <div className="space-y-4 p-4 bg-green-50 dark:bg-green-950/20 rounded-xl border border-green-100 dark:border-green-900">
              <div className="flex items-center gap-2">
                <Server className="w-5 h-5 text-green-600" />
                <Label className="text-sm font-semibold text-green-800 dark:text-green-200">
                  Built-in Mosquitto Broker
                </Label>
              </div>
              <p className="text-sm text-green-700 dark:text-green-300">
                The built-in broker will run on port <code className="bg-green-200 dark:bg-green-800 px-1 rounded">1883</code> (MQTT)
                and <code className="bg-green-200 dark:bg-green-800 px-1 rounded">9001</code> (WebSocket).
              </p>
              <div className="flex items-center justify-between">
                <div>
                  <Label className="font-medium text-green-800 dark:text-green-200">Enable TLS/SSL</Label>
                  <p className="text-xs text-green-600 dark:text-green-400 mt-1">
                    Secure connections on port 8883
                  </p>
                </div>
                <Switch
                  checked={config.useTLS}
                  onCheckedChange={(checked) => updateConfig({ useTLS: checked })}
                />
              </div>
            </div>
          )}

          {/* External Broker Settings */}
          {!config.useBuiltIn && (
            <div className="space-y-4 p-4 bg-gray-50 dark:bg-gray-800/50 rounded-xl">
              <Label className="text-sm font-semibold">External Broker Settings</Label>

              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <div className="space-y-2 md:col-span-2">
                  <Label htmlFor="broker-host">Broker Address</Label>
                  <Input
                    id="broker-host"
                    placeholder="mqtt.example.com"
                    value={config.externalBroker || ''}
                    onChange={(e) =>
                      updateConfig({ externalBroker: e.target.value })
                    }
                  />
                </div>

                <div className="space-y-2">
                  <Label htmlFor="broker-port">Port</Label>
                  <Input
                    id="broker-port"
                    type="number"
                    placeholder="1883"
                    value={config.externalPort || ''}
                    onChange={(e) =>
                      updateConfig({ externalPort: parseInt(e.target.value) || undefined })
                    }
                  />
                </div>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="space-y-2">
                  <Label htmlFor="mqtt-username">Username (optional)</Label>
                  <Input
                    id="mqtt-username"
                    placeholder="mqtt_user"
                    value={config.username || ''}
                    onChange={(e) => updateConfig({ username: e.target.value })}
                  />
                </div>

                <div className="space-y-2">
                  <Label htmlFor="mqtt-password">Password (optional)</Label>
                  <div className="relative">
                    <Input
                      id="mqtt-password"
                      type={showPassword ? 'text' : 'password'}
                      placeholder="********"
                      value={config.password || ''}
                      onChange={(e) =>
                        updateConfig({ password: e.target.value })
                      }
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
              </div>

              <div className="flex items-center justify-between pt-2">
                <div className="flex items-center gap-2">
                  <Shield className="w-4 h-4 text-muted-foreground" />
                  <div>
                    <Label className="font-medium">Use TLS/SSL</Label>
                    <p className="text-xs text-muted-foreground">
                      Encrypt connection to broker
                    </p>
                  </div>
                </div>
                <Switch
                  checked={config.useTLS}
                  onCheckedChange={(checked) => updateConfig({ useTLS: checked })}
                />
              </div>

              {/* Test Connection Button */}
              <div className="pt-2">
                <Button
                  variant="outline"
                  onClick={handleTestConnection}
                  disabled={testing || !config.externalBroker}
                  className="w-full"
                >
                  {testing ? (
                    <>
                      <div className="w-4 h-4 mr-2 border-2 border-current border-t-transparent rounded-full animate-spin" />
                      Testing Connection...
                    </>
                  ) : (
                    <>
                      <Lock className="w-4 h-4 mr-2" />
                      Test Connection
                    </>
                  )}
                </Button>
              </div>
            </div>
          )}

          {/* Popular Brokers */}
          {!config.useBuiltIn && (
            <div className="space-y-3">
              <Label className="text-sm font-semibold text-muted-foreground">
                Popular Cloud Brokers
              </Label>
              <div className="flex flex-wrap gap-2">
                {[
                  { name: 'HiveMQ Cloud', host: 'broker.hivemq.com', port: 1883 },
                  { name: 'EMQX Cloud', host: 'broker.emqx.io', port: 1883 },
                  { name: 'Mosquitto Test', host: 'test.mosquitto.org', port: 1883 },
                ].map((broker) => (
                  <Button
                    key={broker.name}
                    variant="outline"
                    size="sm"
                    onClick={() =>
                      updateConfig({
                        externalBroker: broker.host,
                        externalPort: broker.port,
                      })
                    }
                  >
                    {broker.name}
                  </Button>
                ))}
              </div>
            </div>
          )}
        </>
      )}

      {/* Info Box */}
      {!config.enabled && (
        <div className="p-4 bg-amber-50 dark:bg-amber-950/30 rounded-xl border border-amber-100 dark:border-amber-900">
          <p className="text-sm text-amber-700 dark:text-amber-300">
            <strong>Note:</strong> MQTT is recommended for IoT deployments. Many
            sensors and devices use MQTT for communication.
          </p>
        </div>
      )}
    </div>
  )
}
