/**
 * Completion Step
 *
 * Summary and installation completion
 */

import { useState } from 'react'
import {
  Check,
  Cpu,
  Wifi,
  Globe,
  MessageSquare,
  Zap,
  Download,
  Loader2,
  CheckCircle2,
  Copy,
  Terminal,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Progress } from '@/components/ui/progress'
import { cn } from '@/lib/utils'
import type { SetupConfig } from '../types'
import { SUPPORTED_BOARDS as BOARDS } from '../types'

interface CompletionStepProps {
  config: SetupConfig
  onInstall: () => Promise<void>
  isInstalling: boolean
  installProgress: number
  installComplete: boolean
  onGoToEditor?: () => void
}

interface SummaryItemProps {
  icon: React.ComponentType<{ className?: string }>
  iconColor: string
  label: string
  value: string
  subValue?: string
}

function SummaryItem({
  icon: Icon,
  iconColor,
  label,
  value,
  subValue,
}: SummaryItemProps) {
  return (
    <div className="flex items-center gap-3 p-3 bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700">
      <div
        className={cn(
          'w-10 h-10 rounded-lg flex items-center justify-center',
          iconColor.includes('bg-') ? iconColor : `bg-${iconColor}-100 dark:bg-${iconColor}-900/30`
        )}
      >
        <Icon
          className={cn(
            'w-5 h-5',
            iconColor.includes('text-') ? iconColor : `text-${iconColor}-500`
          )}
        />
      </div>
      <div className="flex-1 min-w-0">
        <p className="text-xs text-muted-foreground">{label}</p>
        <p className="font-medium text-gray-900 dark:text-white truncate">
          {value}
        </p>
        {subValue && (
          <p className="text-xs text-muted-foreground truncate">{subValue}</p>
        )}
      </div>
      <Check className="w-5 h-5 text-green-500" />
    </div>
  )
}

export function CompletionStep({
  config,
  onInstall,
  isInstalling,
  installProgress,
  installComplete,
  onGoToEditor,
}: CompletionStepProps) {
  const [copied, setCopied] = useState(false)

  const board = BOARDS.find((b) => b.id === config.board)
  const enabledInterfaces = [
    config.gpio.enableGPIO && 'GPIO',
    config.gpio.enableI2C && 'I2C',
    config.gpio.enableSPI && 'SPI',
    config.gpio.enableUART && 'UART',
    config.gpio.enable1Wire && '1-Wire',
    config.gpio.enablePWM && 'PWM',
  ].filter(Boolean)

  const installCommand = `curl -sSL https://edgeflow.io/install.sh | sudo bash -s -- \\
  --board=${config.board} \\
  --hostname=${config.network.hostname} \\
  ${config.network.useWifi ? `--wifi-ssid="${config.network.ssid}" \\` : ''}
  ${config.mqtt.enabled ? (config.mqtt.useBuiltIn ? '--mqtt-builtin' : `--mqtt-broker=${config.mqtt.externalBroker}:${config.mqtt.externalPort}`) : '--mqtt-disabled'} \\
  ${enabledInterfaces.length > 0 ? `--interfaces=${enabledInterfaces.join(',')}` : ''}`

  const handleCopyCommand = () => {
    navigator.clipboard.writeText(installCommand)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  if (installComplete) {
    return (
      <div className="space-y-8 text-center">
        <div className="w-20 h-20 mx-auto bg-green-100 dark:bg-green-900/30 rounded-full flex items-center justify-center">
          <CheckCircle2 className="w-10 h-10 text-green-500" />
        </div>

        <div className="space-y-2">
          <h2 className="text-2xl font-bold text-gray-900 dark:text-white">
            Setup Complete!
          </h2>
          <p className="text-muted-foreground max-w-md mx-auto">
            EdgeFlow has been configured successfully. Your device is ready for
            IoT workflows.
          </p>
        </div>

        <div className="space-y-4 max-w-md mx-auto">
          <div className="p-4 bg-green-50 dark:bg-green-950/30 rounded-xl border border-green-200 dark:border-green-900">
            <p className="text-sm text-green-700 dark:text-green-300">
              <strong>Next steps:</strong>
            </p>
            <ul className="mt-2 text-sm text-green-600 dark:text-green-400 space-y-1 text-left">
              <li>• Create your first flow in the Flow Editor</li>
              <li>• Add GPIO, MQTT, or HTTP nodes</li>
              <li>• Connect to sensors and devices</li>
              <li>• Deploy and monitor your flows</li>
            </ul>
          </div>

          <Button className="w-full" size="lg" onClick={onGoToEditor}>
            Go to Flow Editor
          </Button>
        </div>
      </div>
    )
  }

  if (isInstalling) {
    return (
      <div className="space-y-8 text-center">
        <div className="w-20 h-20 mx-auto bg-blue-100 dark:bg-blue-900/30 rounded-full flex items-center justify-center">
          <Loader2 className="w-10 h-10 text-blue-500 animate-spin" />
        </div>

        <div className="space-y-2">
          <h2 className="text-2xl font-bold text-gray-900 dark:text-white">
            Installing EdgeFlow
          </h2>
          <p className="text-muted-foreground">
            Please wait while we configure your device...
          </p>
        </div>

        <div className="max-w-md mx-auto space-y-3">
          <Progress value={installProgress} className="h-3" />
          <p className="text-sm text-muted-foreground">
            {installProgress < 20 && 'Initializing...'}
            {installProgress >= 20 && installProgress < 40 && 'Configuring network...'}
            {installProgress >= 40 && installProgress < 60 && 'Setting up MQTT broker...'}
            {installProgress >= 60 && installProgress < 80 && 'Configuring GPIO permissions...'}
            {installProgress >= 80 && installProgress < 100 && 'Finalizing setup...'}
            {installProgress === 100 && 'Complete!'}
          </p>
        </div>

        <div className="p-4 bg-gray-50 dark:bg-gray-800/50 rounded-xl max-w-md mx-auto">
          <p className="text-xs text-muted-foreground">
            Do not close this window or disconnect from power during
            installation.
          </p>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="text-center space-y-2">
        <h2 className="text-2xl font-bold text-gray-900 dark:text-white">
          Review & Install
        </h2>
        <p className="text-muted-foreground">
          Review your configuration and start the installation
        </p>
      </div>

      {/* Configuration Summary */}
      <div className="space-y-3">
        <h3 className="text-sm font-semibold text-muted-foreground uppercase tracking-wide">
          Configuration Summary
        </h3>

        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          <SummaryItem
            icon={Cpu}
            iconColor="bg-purple-100 dark:bg-purple-900/30 text-purple-500"
            label="Board"
            value={board?.name || 'Unknown'}
            subValue={board?.cpu}
          />

          <SummaryItem
            icon={config.network.useWifi ? Wifi : Globe}
            iconColor={
              config.network.useWifi
                ? 'bg-green-100 dark:bg-green-900/30 text-green-500'
                : 'bg-blue-100 dark:bg-blue-900/30 text-blue-500'
            }
            label="Network"
            value={config.network.useWifi ? 'WiFi' : 'Ethernet'}
            subValue={
              config.network.useWifi
                ? config.network.ssid || 'No SSID'
                : config.network.useStaticIP
                  ? config.network.ipAddress
                  : 'DHCP'
            }
          />

          <SummaryItem
            icon={MessageSquare}
            iconColor="bg-indigo-100 dark:bg-indigo-900/30 text-indigo-500"
            label="MQTT Broker"
            value={
              !config.mqtt.enabled
                ? 'Disabled'
                : config.mqtt.useBuiltIn
                  ? 'Built-in Mosquitto'
                  : 'External'
            }
            subValue={
              !config.mqtt.useBuiltIn && config.mqtt.enabled
                ? `${config.mqtt.externalBroker}:${config.mqtt.externalPort}`
                : config.mqtt.enabled
                  ? 'Port 1883'
                  : undefined
            }
          />

          <SummaryItem
            icon={Zap}
            iconColor="bg-amber-100 dark:bg-amber-900/30 text-amber-500"
            label="Hardware Interfaces"
            value={`${enabledInterfaces.length} enabled`}
            subValue={enabledInterfaces.join(', ') || 'None'}
          />
        </div>
      </div>

      {/* Install Command */}
      <div className="space-y-3">
        <h3 className="text-sm font-semibold text-muted-foreground uppercase tracking-wide flex items-center gap-2">
          <Terminal className="w-4 h-4" />
          Installation Command
        </h3>

        <div className="relative">
          <pre className="p-4 bg-gray-900 text-gray-100 rounded-xl text-xs overflow-x-auto font-mono">
            {installCommand}
          </pre>
          <Button
            variant="secondary"
            size="sm"
            className="absolute top-2 right-2"
            onClick={handleCopyCommand}
          >
            {copied ? (
              <>
                <Check className="w-4 h-4 mr-1" />
                Copied
              </>
            ) : (
              <>
                <Copy className="w-4 h-4 mr-1" />
                Copy
              </>
            )}
          </Button>
        </div>

        <p className="text-xs text-muted-foreground">
          You can also run this command manually on your device via SSH.
        </p>
      </div>

      {/* Install Button */}
      <div className="pt-4">
        <Button onClick={onInstall} className="w-full" size="lg">
          <Download className="w-5 h-5 mr-2" />
          Start Installation
        </Button>
        <p className="text-xs text-center text-muted-foreground mt-3">
          This will apply the configuration to your device. A reboot may be
          required.
        </p>
      </div>
    </div>
  )
}
