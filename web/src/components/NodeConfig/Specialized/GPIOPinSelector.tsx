/**
 * GPIO Pin Selector
 *
 * Visual Raspberry Pi 5 GPIO pinout selector with pin details
 */

import { useState } from 'react'
import { Label } from '@/components/ui/label'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { cn } from '@/lib/utils'

export type PinMode = 'input' | 'output' | 'pwm' | 'i2c' | 'spi' | 'uart'
export type PullMode = 'off' | 'up' | 'down'

interface GPIOPin {
  physical: number
  bcm?: number
  name: string
  type: 'gpio' | 'power' | 'ground' | 'i2c' | 'spi' | 'uart' | 'reserved'
  voltage?: '3.3V' | '5V'
  pwmChannel?: number
}

interface GPIOPinSelectorProps {
  value: {
    pin?: number // BCM pin number
    mode?: PinMode
    pullMode?: PullMode
    initialValue?: boolean | number
    debounceMs?: number
  }
  onChange: (value: any) => void
  disabled?: boolean
}

// Raspberry Pi 5 GPIO pinout (40-pin header)
const GPIO_PINS: GPIOPin[] = [
  // Left side (odd pins)
  { physical: 1, name: '3.3V', type: 'power', voltage: '3.3V' },
  { physical: 3, bcm: 2, name: 'GPIO2 (SDA)', type: 'i2c' },
  { physical: 5, bcm: 3, name: 'GPIO3 (SCL)', type: 'i2c' },
  { physical: 7, bcm: 4, name: 'GPIO4 (GPCLK0)', type: 'gpio' },
  { physical: 9, name: 'Ground', type: 'ground' },
  { physical: 11, bcm: 17, name: 'GPIO17', type: 'gpio' },
  { physical: 13, bcm: 27, name: 'GPIO27', type: 'gpio' },
  { physical: 15, bcm: 22, name: 'GPIO22', type: 'gpio' },
  { physical: 17, name: '3.3V', type: 'power', voltage: '3.3V' },
  { physical: 19, bcm: 10, name: 'GPIO10 (MOSI)', type: 'spi' },
  { physical: 21, bcm: 9, name: 'GPIO9 (MISO)', type: 'spi' },
  { physical: 23, bcm: 11, name: 'GPIO11 (SCLK)', type: 'spi' },
  { physical: 25, name: 'Ground', type: 'ground' },
  { physical: 27, bcm: 0, name: 'ID_SD', type: 'reserved' },
  { physical: 29, bcm: 5, name: 'GPIO5', type: 'gpio' },
  { physical: 31, bcm: 6, name: 'GPIO6', type: 'gpio' },
  { physical: 33, bcm: 13, name: 'GPIO13 (PWM1)', type: 'gpio', pwmChannel: 1 },
  { physical: 35, bcm: 19, name: 'GPIO19 (PWM1)', type: 'gpio', pwmChannel: 1 },
  { physical: 37, bcm: 26, name: 'GPIO26', type: 'gpio' },
  { physical: 39, name: 'Ground', type: 'ground' },

  // Right side (even pins)
  { physical: 2, name: '5V', type: 'power', voltage: '5V' },
  { physical: 4, name: '5V', type: 'power', voltage: '5V' },
  { physical: 6, name: 'Ground', type: 'ground' },
  { physical: 8, bcm: 14, name: 'GPIO14 (TXD)', type: 'uart' },
  { physical: 10, bcm: 15, name: 'GPIO15 (RXD)', type: 'uart' },
  { physical: 12, bcm: 18, name: 'GPIO18 (PWM0)', type: 'gpio', pwmChannel: 0 },
  { physical: 14, name: 'Ground', type: 'ground' },
  { physical: 16, bcm: 23, name: 'GPIO23', type: 'gpio' },
  { physical: 18, bcm: 24, name: 'GPIO24', type: 'gpio' },
  { physical: 20, name: 'Ground', type: 'ground' },
  { physical: 22, bcm: 25, name: 'GPIO25', type: 'gpio' },
  { physical: 24, bcm: 8, name: 'GPIO8 (CE0)', type: 'spi' },
  { physical: 26, bcm: 7, name: 'GPIO7 (CE1)', type: 'spi' },
  { physical: 28, bcm: 1, name: 'ID_SC', type: 'reserved' },
  { physical: 30, name: 'Ground', type: 'ground' },
  { physical: 32, bcm: 12, name: 'GPIO12 (PWM0)', type: 'gpio', pwmChannel: 0 },
  { physical: 34, name: 'Ground', type: 'ground' },
  { physical: 36, bcm: 16, name: 'GPIO16', type: 'gpio' },
  { physical: 38, bcm: 20, name: 'GPIO20', type: 'gpio' },
  { physical: 40, bcm: 21, name: 'GPIO21', type: 'gpio' },
]

const PIN_TYPE_COLORS = {
  gpio: 'bg-green-500 hover:bg-green-600',
  power: 'bg-red-500 hover:bg-red-600 cursor-not-allowed',
  ground: 'bg-black hover:bg-gray-800 cursor-not-allowed',
  i2c: 'bg-blue-500 hover:bg-blue-600',
  spi: 'bg-purple-500 hover:bg-purple-600',
  uart: 'bg-yellow-500 hover:bg-yellow-600',
  reserved: 'bg-gray-400 hover:bg-gray-500 cursor-not-allowed',
}

export function GPIOPinSelector({ value, onChange, disabled = false }: GPIOPinSelectorProps) {
  // Handle undefined or null value
  const safeValue = value || {}
  const selectedBCM = safeValue.pin
  const mode = safeValue.mode || 'input'
  const pullMode = safeValue.pullMode || 'off'
  const initialValue = safeValue.initialValue ?? false
  const debounceMs = safeValue.debounceMs ?? 0

  const [hoveredPin, setHoveredPin] = useState<number | null>(null)

  const selectedPin = GPIO_PINS.find((p) => p.bcm === selectedBCM)

  const handlePinClick = (pin: GPIOPin) => {
    if (disabled) return
    if (pin.type === 'power' || pin.type === 'ground' || pin.type === 'reserved') return

    onChange({ ...safeValue, pin: pin.bcm })
  }

  const isSelectable = (pin: GPIOPin) => {
    return pin.type !== 'power' && pin.type !== 'ground' && pin.type !== 'reserved'
  }

  const canUsePWM = selectedPin?.pwmChannel !== undefined

  return (
    <div className="space-y-6">
      {/* Pin Selection Info */}
      <div className="p-4 bg-muted rounded-lg border">
        <div className="flex items-center justify-between mb-2">
          <Label className="text-sm font-semibold">Selected Pin</Label>
          {selectedPin && (
            <span className="text-xs font-mono px-2 py-1 bg-background rounded border">
              BCM {selectedPin.bcm} (Physical {selectedPin.physical})
            </span>
          )}
        </div>
        {selectedPin ? (
          <div className="text-sm">
            <p className="font-medium">{selectedPin.name}</p>
            <p className="text-xs text-muted-foreground">
              Type: {selectedPin.type.toUpperCase()}
              {selectedPin.pwmChannel !== undefined && ` • PWM Channel ${selectedPin.pwmChannel}`}
            </p>
          </div>
        ) : (
          <p className="text-sm text-muted-foreground">Click a pin on the diagram below</p>
        )}
      </div>

      {/* GPIO Pinout Diagram */}
      <div className="space-y-2">
        <Label className="text-sm font-semibold">Raspberry Pi 5 GPIO Pinout</Label>
        <div className="relative bg-card border-2 border-border rounded-lg p-4">
          {/* Header */}
          <div className="text-center mb-4">
            <p className="text-sm font-semibold">40-Pin GPIO Header</p>
            <p className="text-xs text-muted-foreground">Click a GPIO pin to select</p>
          </div>

          {/* Pin Grid */}
          <div className="grid grid-cols-2 gap-1 max-w-md mx-auto">
            {/* Render pins in pairs (odd on left, even on right) */}
            {Array.from({ length: 20 }, (_, i) => {
              const leftPin = GPIO_PINS.find((p) => p.physical === i * 2 + 1)
              const rightPin = GPIO_PINS.find((p) => p.physical === i * 2 + 2)

              return (
                <div key={i} className="contents">
                  {/* Left Pin (Odd) */}
                  {leftPin && (
                    <div
                      className={cn(
                        'relative group',
                        !isSelectable(leftPin) && 'opacity-60'
                      )}
                      onMouseEnter={() => setHoveredPin(leftPin.physical)}
                      onMouseLeave={() => setHoveredPin(null)}
                      onClick={() => handlePinClick(leftPin)}
                    >
                      <div
                        className={cn(
                          'flex items-center justify-between p-2 rounded-l-md transition-all',
                          PIN_TYPE_COLORS[leftPin.type],
                          selectedBCM === leftPin.bcm && 'ring-2 ring-white ring-offset-2',
                          isSelectable(leftPin) && !disabled && 'cursor-pointer',
                          disabled && 'opacity-50 cursor-not-allowed'
                        )}
                      >
                        <span className="text-xs font-mono text-white font-bold">
                          {leftPin.physical}
                        </span>
                        <span className="text-xs text-white truncate ml-2">
                          {leftPin.bcm !== undefined ? `GPIO${leftPin.bcm}` : leftPin.name}
                        </span>
                      </div>

                      {/* Tooltip */}
                      {hoveredPin === leftPin.physical && (
                        <div className="absolute left-0 top-full mt-1 z-10 bg-popover text-popover-foreground border rounded-md shadow-lg p-2 text-xs whitespace-nowrap">
                          <p className="font-semibold">{leftPin.name}</p>
                          <p className="text-muted-foreground">
                            Physical: {leftPin.physical}
                            {leftPin.bcm !== undefined && ` • BCM: ${leftPin.bcm}`}
                          </p>
                        </div>
                      )}
                    </div>
                  )}

                  {/* Right Pin (Even) */}
                  {rightPin && (
                    <div
                      className={cn(
                        'relative group',
                        !isSelectable(rightPin) && 'opacity-60'
                      )}
                      onMouseEnter={() => setHoveredPin(rightPin.physical)}
                      onMouseLeave={() => setHoveredPin(null)}
                      onClick={() => handlePinClick(rightPin)}
                    >
                      <div
                        className={cn(
                          'flex items-center justify-between p-2 rounded-r-md transition-all',
                          PIN_TYPE_COLORS[rightPin.type],
                          selectedBCM === rightPin.bcm && 'ring-2 ring-white ring-offset-2',
                          isSelectable(rightPin) && !disabled && 'cursor-pointer',
                          disabled && 'opacity-50 cursor-not-allowed'
                        )}
                      >
                        <span className="text-xs text-white truncate mr-2">
                          {rightPin.bcm !== undefined ? `GPIO${rightPin.bcm}` : rightPin.name}
                        </span>
                        <span className="text-xs font-mono text-white font-bold">
                          {rightPin.physical}
                        </span>
                      </div>

                      {/* Tooltip */}
                      {hoveredPin === rightPin.physical && (
                        <div className="absolute right-0 top-full mt-1 z-10 bg-popover text-popover-foreground border rounded-md shadow-lg p-2 text-xs whitespace-nowrap">
                          <p className="font-semibold">{rightPin.name}</p>
                          <p className="text-muted-foreground">
                            Physical: {rightPin.physical}
                            {rightPin.bcm !== undefined && ` • BCM: ${rightPin.bcm}`}
                          </p>
                        </div>
                      )}
                    </div>
                  )}
                </div>
              )
            })}
          </div>

          {/* Legend */}
          <div className="mt-4 pt-4 border-t grid grid-cols-2 gap-2 text-xs">
            <div className="flex items-center gap-2">
              <div className="w-4 h-4 rounded bg-green-500"></div>
              <span>GPIO</span>
            </div>
            <div className="flex items-center gap-2">
              <div className="w-4 h-4 rounded bg-blue-500"></div>
              <span>I2C</span>
            </div>
            <div className="flex items-center gap-2">
              <div className="w-4 h-4 rounded bg-purple-500"></div>
              <span>SPI</span>
            </div>
            <div className="flex items-center gap-2">
              <div className="w-4 h-4 rounded bg-yellow-500"></div>
              <span>UART</span>
            </div>
            <div className="flex items-center gap-2">
              <div className="w-4 h-4 rounded bg-red-500"></div>
              <span>Power</span>
            </div>
            <div className="flex items-center gap-2">
              <div className="w-4 h-4 rounded bg-black"></div>
              <span>Ground</span>
            </div>
          </div>
        </div>
      </div>

      {/* Pin Configuration */}
      {selectedPin && (
        <div className="space-y-4">
          <Label className="text-sm font-semibold">Pin Configuration</Label>

          {/* Mode Selection */}
          <div className="space-y-2">
            <Label htmlFor="mode" className="text-xs">
              Pin Mode
            </Label>
            <Select
              value={mode}
              onValueChange={(value: PinMode) => onChange({ ...safeValue, mode: value })}
              disabled={disabled}
            >
              <SelectTrigger className="h-11" id="mode">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="input">Input (Read)</SelectItem>
                <SelectItem value="output">Output (Write)</SelectItem>
                <SelectItem value="pwm" disabled={!canUsePWM}>
                  PWM {!canUsePWM && '(Not available)'}
                </SelectItem>
                <SelectItem value="i2c" disabled={selectedPin.type !== 'i2c'}>
                  I2C {selectedPin.type !== 'i2c' && '(Not available)'}
                </SelectItem>
                <SelectItem value="spi" disabled={selectedPin.type !== 'spi'}>
                  SPI {selectedPin.type !== 'spi' && '(Not available)'}
                </SelectItem>
                <SelectItem value="uart" disabled={selectedPin.type !== 'uart'}>
                  UART {selectedPin.type !== 'uart' && '(Not available)'}
                </SelectItem>
              </SelectContent>
            </Select>
          </div>

          {/* Pull Mode (for input) */}
          {mode === 'input' && (
            <div className="space-y-2">
              <Label htmlFor="pullMode" className="text-xs">
                Pull Resistor
              </Label>
              <Select
                value={pullMode}
                onValueChange={(value: PullMode) => onChange({ ...safeValue, pullMode: value })}
                disabled={disabled}
              >
                <SelectTrigger className="h-11" id="pullMode">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="off">None (Floating)</SelectItem>
                  <SelectItem value="up">Pull Up (3.3V)</SelectItem>
                  <SelectItem value="down">Pull Down (GND)</SelectItem>
                </SelectContent>
              </Select>
            </div>
          )}

          {/* Debounce (for input) */}
          {mode === 'input' && (
            <div className="space-y-2">
              <Label htmlFor="debounce" className="text-xs">
                Debounce (ms)
              </Label>
              <Input
                id="debounce"
                type="number"
                value={debounceMs}
                onChange={(e) =>
                  onChange({ ...safeValue, debounceMs: Number(e.target.value) })
                }
                min={0}
                max={1000}
                className="h-11"
                disabled={disabled}
              />
              <p className="text-xs text-muted-foreground">
                Delay to ignore rapid input changes (useful for buttons)
              </p>
            </div>
          )}

          {/* Initial Value (for output) */}
          {mode === 'output' && (
            <div className="space-y-2">
              <div className="flex items-center space-x-2">
                <Switch
                  id="initialValue"
                  checked={initialValue as boolean}
                  onCheckedChange={(checked) =>
                    onChange({ ...safeValue, initialValue: checked })
                  }
                  disabled={disabled}
                />
                <Label htmlFor="initialValue" className="text-xs cursor-pointer">
                  Initial State: {initialValue ? 'HIGH' : 'LOW'}
                </Label>
              </div>
            </div>
          )}

          {/* PWM Frequency (for PWM mode) */}
          {mode === 'pwm' && (
            <div className="space-y-2">
              <Label htmlFor="pwmFreq" className="text-xs">
                PWM Frequency (Hz)
              </Label>
              <Input
                id="pwmFreq"
                type="number"
                value={(value.initialValue as number) || 1000}
                onChange={(e) =>
                  onChange({ ...safeValue, initialValue: Number(e.target.value) })
                }
                min={1}
                max={100000}
                className="h-11"
                disabled={disabled}
              />
            </div>
          )}
        </div>
      )}
    </div>
  )
}
