/**
 * GPIO Permissions Setup Step
 *
 * Configure GPIO and hardware interface permissions
 */

import { Zap, Cpu, Radio, CircuitBoard, Thermometer, Activity, AlertTriangle } from 'lucide-react'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { cn } from '@/lib/utils'
import type { GPIOConfig, BoardType } from '../types'

interface GPIOPermissionsStepProps {
  config: GPIOConfig
  boardType: BoardType | null
  onChange: (config: GPIOConfig) => void
}

interface PermissionItem {
  key: keyof GPIOConfig
  label: string
  description: string
  icon: React.ComponentType<{ className?: string }>
  iconColor: string
  bgColor: string
  borderColor: string
  recommendation?: string
}

const PERMISSION_ITEMS: PermissionItem[] = [
  {
    key: 'enableGPIO',
    label: 'GPIO Access',
    description: 'Digital input/output pins for buttons, LEDs, relays',
    icon: Zap,
    iconColor: 'text-yellow-500',
    bgColor: 'bg-yellow-50 dark:bg-yellow-950/30',
    borderColor: 'border-yellow-200 dark:border-yellow-900',
    recommendation: 'Recommended for most IoT projects',
  },
  {
    key: 'enableI2C',
    label: 'I2C Bus',
    description: 'Two-wire interface for sensors (BME280, MPU6050, displays)',
    icon: Cpu,
    iconColor: 'text-blue-500',
    bgColor: 'bg-blue-50 dark:bg-blue-950/30',
    borderColor: 'border-blue-200 dark:border-blue-900',
    recommendation: 'Enable for most sensors',
  },
  {
    key: 'enableSPI',
    label: 'SPI Bus',
    description: 'High-speed interface for displays, SD cards, LoRa modules',
    icon: Radio,
    iconColor: 'text-purple-500',
    bgColor: 'bg-purple-50 dark:bg-purple-950/30',
    borderColor: 'border-purple-200 dark:border-purple-900',
  },
  {
    key: 'enableUART',
    label: 'UART/Serial',
    description: 'Serial communication for GPS, modems, debug consoles',
    icon: CircuitBoard,
    iconColor: 'text-green-500',
    bgColor: 'bg-green-50 dark:bg-green-950/30',
    borderColor: 'border-green-200 dark:border-green-900',
  },
  {
    key: 'enable1Wire',
    label: '1-Wire Protocol',
    description: 'Single-wire protocol for DS18B20 temperature sensors',
    icon: Thermometer,
    iconColor: 'text-red-500',
    bgColor: 'bg-red-50 dark:bg-red-950/30',
    borderColor: 'border-red-200 dark:border-red-900',
  },
  {
    key: 'enablePWM',
    label: 'PWM Output',
    description: 'Pulse-width modulation for motors, dimmers, servos',
    icon: Activity,
    iconColor: 'text-cyan-500',
    bgColor: 'bg-cyan-50 dark:bg-cyan-950/30',
    borderColor: 'border-cyan-200 dark:border-cyan-900',
    recommendation: 'Enable for motor/LED control',
  },
]

export function GPIOPermissionsStep({
  config,
  boardType,
  onChange,
}: GPIOPermissionsStepProps) {
  const updateConfig = (key: keyof GPIOConfig, value: boolean) => {
    onChange({ ...config, [key]: value })
  }

  const enableAll = () => {
    onChange({
      enableGPIO: true,
      enableI2C: true,
      enableSPI: true,
      enableUART: true,
      enable1Wire: true,
      enablePWM: true,
    })
  }

  const disableAll = () => {
    onChange({
      enableGPIO: false,
      enableI2C: false,
      enableSPI: false,
      enableUART: false,
      enable1Wire: false,
      enablePWM: false,
    })
  }

  const allEnabled = Object.values(config).every((v) => v)
  const anyEnabled = Object.values(config).some((v) => v)

  return (
    <div className="space-y-6">
      <div className="text-center space-y-2">
        <h2 className="text-2xl font-bold text-gray-900 dark:text-white">
          Hardware Permissions
        </h2>
        <p className="text-muted-foreground">
          Configure which hardware interfaces EdgeFlow can access
        </p>
      </div>

      {/* Quick Actions */}
      <div className="flex justify-center gap-3">
        <button
          onClick={enableAll}
          className={cn(
            'px-4 py-2 text-sm rounded-lg transition-colors',
            allEnabled
              ? 'bg-primary text-primary-foreground'
              : 'bg-gray-100 dark:bg-gray-800 hover:bg-gray-200 dark:hover:bg-gray-700'
          )}
        >
          Enable All
        </button>
        <button
          onClick={disableAll}
          className={cn(
            'px-4 py-2 text-sm rounded-lg transition-colors',
            !anyEnabled
              ? 'bg-primary text-primary-foreground'
              : 'bg-gray-100 dark:bg-gray-800 hover:bg-gray-200 dark:hover:bg-gray-700'
          )}
        >
          Disable All
        </button>
      </div>

      {/* Permission Grid */}
      <div className="space-y-3">
        {PERMISSION_ITEMS.map((item) => {
          const Icon = item.icon
          const isEnabled = config[item.key]

          return (
            <div
              key={item.key}
              className={cn(
                'p-4 rounded-xl border-2 transition-all',
                isEnabled
                  ? `${item.bgColor} ${item.borderColor}`
                  : 'bg-gray-50 dark:bg-gray-800/50 border-gray-200 dark:border-gray-700'
              )}
            >
              <div className="flex items-start justify-between gap-4">
                <div className="flex items-start gap-3">
                  <div
                    className={cn(
                      'w-10 h-10 rounded-lg flex items-center justify-center flex-shrink-0',
                      isEnabled
                        ? item.bgColor
                        : 'bg-gray-200 dark:bg-gray-700'
                    )}
                  >
                    <Icon
                      className={cn(
                        'w-5 h-5',
                        isEnabled ? item.iconColor : 'text-gray-400'
                      )}
                    />
                  </div>
                  <div>
                    <Label className="font-medium text-gray-900 dark:text-white">
                      {item.label}
                    </Label>
                    <p className="text-sm text-muted-foreground mt-1">
                      {item.description}
                    </p>
                    {item.recommendation && (
                      <p className="text-xs text-primary mt-1">
                        {item.recommendation}
                      </p>
                    )}
                  </div>
                </div>
                <Switch
                  checked={isEnabled}
                  onCheckedChange={(checked) => updateConfig(item.key, checked)}
                />
              </div>
            </div>
          )
        })}
      </div>

      {/* Board-specific warning */}
      {boardType === 'rpi-zero-w' && (
        <div className="flex items-start gap-3 p-4 bg-amber-50 dark:bg-amber-950/30 rounded-xl border border-amber-200 dark:border-amber-900">
          <AlertTriangle className="w-5 h-5 text-amber-500 flex-shrink-0 mt-0.5" />
          <div>
            <p className="text-sm font-medium text-amber-800 dark:text-amber-200">
              Raspberry Pi Zero W Notice
            </p>
            <p className="text-sm text-amber-700 dark:text-amber-300 mt-1">
              Due to limited resources, we recommend enabling only the
              interfaces you need. Disable SPI and UART if not required.
            </p>
          </div>
        </div>
      )}

      {/* Setup Commands Info */}
      <div className="p-4 bg-gray-50 dark:bg-gray-800/50 rounded-xl">
        <p className="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">
          System Configuration
        </p>
        <p className="text-xs text-muted-foreground">
          EdgeFlow will configure the following during installation:
        </p>
        <ul className="mt-2 text-xs text-muted-foreground space-y-1">
          {config.enableGPIO && (
            <li className="flex items-center gap-2">
              <span className="w-1.5 h-1.5 bg-yellow-500 rounded-full" />
              Add user to gpio group
            </li>
          )}
          {config.enableI2C && (
            <li className="flex items-center gap-2">
              <span className="w-1.5 h-1.5 bg-blue-500 rounded-full" />
              Enable I2C in /boot/config.txt
            </li>
          )}
          {config.enableSPI && (
            <li className="flex items-center gap-2">
              <span className="w-1.5 h-1.5 bg-purple-500 rounded-full" />
              Enable SPI in /boot/config.txt
            </li>
          )}
          {config.enableUART && (
            <li className="flex items-center gap-2">
              <span className="w-1.5 h-1.5 bg-green-500 rounded-full" />
              Enable UART in /boot/config.txt
            </li>
          )}
          {config.enable1Wire && (
            <li className="flex items-center gap-2">
              <span className="w-1.5 h-1.5 bg-red-500 rounded-full" />
              Enable 1-Wire on GPIO4
            </li>
          )}
          {config.enablePWM && (
            <li className="flex items-center gap-2">
              <span className="w-1.5 h-1.5 bg-cyan-500 rounded-full" />
              Enable hardware PWM overlay
            </li>
          )}
        </ul>
      </div>
    </div>
  )
}
