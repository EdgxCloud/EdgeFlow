/**
 * GPIO Panel
 * Real-time GPIO pin monitoring with visual Raspberry Pi header diagram
 */

import { useState, useEffect, useCallback, useRef } from 'react'
import { api } from '@/services/api'
import { wsClient, WSMessage } from '@/services/websocket'
import { GPIO_PINS } from '@/data/gpioPins'
import type { GPIOPin } from '@/data/gpioPins'
import { cn } from '@/lib/utils'
import { CircuitBoard, AlertTriangle, Cpu } from 'lucide-react'

interface PinState {
  bcm_pin: number
  value: boolean
  mode: string
  edge_count: number
  last_change: string
}

interface GPIOMonitorState {
  pins: Record<string, PinState>
  board_name: string
  gpio_chip: string
  available: boolean
  timestamp: string
}

export function GPIOPanel() {
  const [gpioState, setGpioState] = useState<GPIOMonitorState | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [hoveredPin, setHoveredPin] = useState<number | null>(null)
  const mountedRef = useRef(true)

  // Fetch initial state from REST
  const fetchState = useCallback(async () => {
    try {
      const data = await api.get<GPIOMonitorState>('/gpio/state')
      if (mountedRef.current) {
        setGpioState(data)
        setIsLoading(false)
      }
    } catch {
      if (mountedRef.current) {
        setIsLoading(false)
      }
    }
  }, [])

  useEffect(() => {
    mountedRef.current = true
    fetchState()

    // Ensure WebSocket is connected for real-time updates
    wsClient.connect()

    // Subscribe to WebSocket gpio_state updates
    const unsub = wsClient.on('gpio_state', (msg: WSMessage) => {
      if (mountedRef.current) {
        setGpioState(msg.data as GPIOMonitorState)
      }
    })

    // Re-fetch on WebSocket reconnect
    const unsubOpen = wsClient.onOpen(() => {
      fetchState()
    })

    // Poll REST endpoint as fallback (every 2s) in case WebSocket misses updates
    const pollInterval = setInterval(() => {
      if (mountedRef.current) {
        fetchState()
      }
    }, 2000)

    return () => {
      mountedRef.current = false
      unsub()
      unsubOpen()
      clearInterval(pollInterval)
    }
  }, [fetchState])

  // Get active pin state by BCM number
  const getPinState = (bcm: number | undefined): PinState | null => {
    if (bcm === undefined || !gpioState?.pins) return null
    return gpioState.pins[String(bcm)] || null
  }

  const activePinCount = gpioState?.pins ? Object.keys(gpioState.pins).length : 0

  // Not available (non-Linux / no HAL)
  if (!isLoading && gpioState && !gpioState.available) {
    return (
      <div className="h-full flex items-center justify-center text-muted-foreground text-sm">
        <div className="text-center">
          <AlertTriangle className="w-8 h-8 mx-auto mb-2 opacity-50 text-yellow-500" />
          <p className="font-medium">GPIO monitoring requires Raspberry Pi</p>
          <p className="text-xs mt-1 text-muted-foreground">
            Connect to a Raspberry Pi to view live GPIO states
          </p>
        </div>
      </div>
    )
  }

  // Loading
  if (isLoading) {
    return (
      <div className="h-full flex items-center justify-center text-muted-foreground text-sm">
        <div className="text-center">
          <CircuitBoard className="w-8 h-8 mx-auto mb-2 opacity-50 animate-pulse" />
          <p>Loading GPIO state...</p>
        </div>
      </div>
    )
  }

  // Get pin color for monitoring view
  const getMonitorPinColor = (pin: GPIOPin): string => {
    if (pin.type === 'power') {
      return pin.voltage === '5V' ? 'bg-red-600' : 'bg-orange-500'
    }
    if (pin.type === 'ground') return 'bg-gray-800 dark:bg-gray-600'
    if (pin.type === 'reserved') return 'bg-gray-400'

    // Check if pin is active
    const state = getPinState(pin.bcm)
    if (!state) {
      // Inactive GPIO pin
      return 'bg-gray-500/40 dark:bg-gray-700/50'
    }

    // Active pin
    if (state.mode === 'pwm') {
      return 'bg-purple-500'
    }
    if (state.value) {
      return 'bg-green-500'
    }
    return 'bg-gray-500 dark:bg-gray-600'
  }

  // Get pin glow for active HIGH pins
  const getMonitorPinGlow = (pin: GPIOPin): string => {
    const state = getPinState(pin.bcm)
    if (!state) return ''
    if (state.value) {
      return 'shadow-[0_0_8px_2px_rgba(34,197,94,0.6)] ring-1 ring-green-400/50'
    }
    if (state.mode === 'pwm') {
      return 'shadow-[0_0_6px_1px_rgba(168,85,247,0.5)] ring-1 ring-purple-400/50'
    }
    return ''
  }

  const formatLastChange = (timestamp: string): string => {
    if (!timestamp) return '--'
    const diff = Date.now() - new Date(timestamp).getTime()
    if (diff < 1000) return 'just now'
    if (diff < 60000) return `${Math.floor(diff / 1000)}s ago`
    if (diff < 3600000) return `${Math.floor(diff / 60000)}m ago`
    return `${Math.floor(diff / 3600000)}h ago`
  }

  return (
    <div className="h-full overflow-auto">
      <div className="p-3 space-y-3">
        {/* Header Bar */}
        <div className="flex items-center justify-between px-3 py-2 bg-muted/30 rounded-lg border border-border">
          <div className="flex items-center gap-3">
            <CircuitBoard className="w-4 h-4 text-green-500" />
            <div>
              <span className="text-xs font-semibold">
                {gpioState?.board_name || 'Raspberry Pi'}
              </span>
              {gpioState?.gpio_chip && (
                <span className="text-[10px] text-muted-foreground ml-2">
                  {gpioState.gpio_chip}
                </span>
              )}
            </div>
          </div>
          <div className="flex items-center gap-4 text-[11px] text-muted-foreground">
            <div className="flex items-center gap-1">
              <Cpu className="w-3 h-3" />
              <span>Active Pins: {activePinCount}</span>
            </div>
            <div className={cn(
              'w-2 h-2 rounded-full',
              activePinCount > 0 ? 'bg-green-500 animate-pulse' : 'bg-gray-400'
            )} />
          </div>
        </div>

        {/* Main Content: Pin Diagram + Details */}
        <div className="flex gap-4">
          {/* Pin Header Diagram */}
          <div className="flex-1 bg-card border rounded-lg p-4">
            <div className="text-center mb-3">
              <p className="text-sm font-semibold">40-Pin GPIO Header</p>
              {activePinCount === 0 && (
                <p className="text-xs text-muted-foreground mt-1">
                  Start a flow with GPIO nodes to see live pin states
                </p>
              )}
            </div>

            {/* Pin Grid */}
            <div className="grid grid-cols-2 gap-1 max-w-md mx-auto">
              {Array.from({ length: 20 }, (_, i) => {
                const leftPin = GPIO_PINS.find((p) => p.physical === i * 2 + 1)
                const rightPin = GPIO_PINS.find((p) => p.physical === i * 2 + 2)

                return (
                  <div key={i} className="contents">
                    {/* Left Pin (Odd) */}
                    {leftPin && (
                      <div
                        className="relative"
                        onMouseEnter={() => setHoveredPin(leftPin.physical)}
                        onMouseLeave={() => setHoveredPin(null)}
                      >
                        <div
                          className={cn(
                            'flex items-center justify-between p-1.5 rounded-l-md transition-all duration-300',
                            getMonitorPinColor(leftPin),
                            getMonitorPinGlow(leftPin),
                            !getPinState(leftPin.bcm) && leftPin.type === 'gpio' && 'opacity-50'
                          )}
                        >
                          <span className="text-[10px] font-mono text-white font-bold w-5 text-center">
                            {leftPin.physical}
                          </span>
                          <span className="text-[10px] text-white truncate ml-1 flex-1 text-right">
                            {leftPin.bcm !== undefined ? `GPIO${leftPin.bcm}` : leftPin.name}
                          </span>
                          {/* State indicator */}
                          {getPinState(leftPin.bcm) && (
                            <span className={cn(
                              'ml-1 text-[9px] font-bold px-1 rounded',
                              getPinState(leftPin.bcm)?.value
                                ? 'bg-white/30 text-white'
                                : 'bg-black/20 text-white/70'
                            )}>
                              {getPinState(leftPin.bcm)?.mode === 'pwm' ? 'PWM' : getPinState(leftPin.bcm)?.value ? 'H' : 'L'}
                            </span>
                          )}
                        </div>

                        {/* Tooltip */}
                        {hoveredPin === leftPin.physical && (
                          <div className="absolute left-0 top-full mt-1 z-20 bg-popover text-popover-foreground border rounded-md shadow-lg p-2 text-xs whitespace-nowrap">
                            <p className="font-semibold">{leftPin.name}</p>
                            <p className="text-muted-foreground">
                              Physical: {leftPin.physical}
                              {leftPin.bcm !== undefined && ` | BCM: ${leftPin.bcm}`}
                            </p>
                            {getPinState(leftPin.bcm) && (
                              <div className="mt-1 pt-1 border-t border-border space-y-0.5">
                                <p>Mode: <span className="font-mono">{getPinState(leftPin.bcm)?.mode}</span></p>
                                <p>Value: <span className={cn('font-bold', getPinState(leftPin.bcm)?.value ? 'text-green-500' : 'text-gray-400')}>
                                  {getPinState(leftPin.bcm)?.value ? 'HIGH' : 'LOW'}
                                </span></p>
                                <p>Edges: <span className="font-mono">{getPinState(leftPin.bcm)?.edge_count}</span></p>
                                <p>Changed: {formatLastChange(getPinState(leftPin.bcm)?.last_change || '')}</p>
                              </div>
                            )}
                          </div>
                        )}
                      </div>
                    )}

                    {/* Right Pin (Even) */}
                    {rightPin && (
                      <div
                        className="relative"
                        onMouseEnter={() => setHoveredPin(rightPin.physical)}
                        onMouseLeave={() => setHoveredPin(null)}
                      >
                        <div
                          className={cn(
                            'flex items-center justify-between p-1.5 rounded-r-md transition-all duration-300',
                            getMonitorPinColor(rightPin),
                            getMonitorPinGlow(rightPin),
                            !getPinState(rightPin.bcm) && rightPin.type === 'gpio' && 'opacity-50'
                          )}
                        >
                          {/* State indicator */}
                          {getPinState(rightPin.bcm) && (
                            <span className={cn(
                              'mr-1 text-[9px] font-bold px-1 rounded',
                              getPinState(rightPin.bcm)?.value
                                ? 'bg-white/30 text-white'
                                : 'bg-black/20 text-white/70'
                            )}>
                              {getPinState(rightPin.bcm)?.mode === 'pwm' ? 'PWM' : getPinState(rightPin.bcm)?.value ? 'H' : 'L'}
                            </span>
                          )}
                          <span className="text-[10px] text-white truncate mr-1 flex-1">
                            {rightPin.bcm !== undefined ? `GPIO${rightPin.bcm}` : rightPin.name}
                          </span>
                          <span className="text-[10px] font-mono text-white font-bold w-5 text-center">
                            {rightPin.physical}
                          </span>
                        </div>

                        {/* Tooltip */}
                        {hoveredPin === rightPin.physical && (
                          <div className="absolute right-0 top-full mt-1 z-20 bg-popover text-popover-foreground border rounded-md shadow-lg p-2 text-xs whitespace-nowrap">
                            <p className="font-semibold">{rightPin.name}</p>
                            <p className="text-muted-foreground">
                              Physical: {rightPin.physical}
                              {rightPin.bcm !== undefined && ` | BCM: ${rightPin.bcm}`}
                            </p>
                            {getPinState(rightPin.bcm) && (
                              <div className="mt-1 pt-1 border-t border-border space-y-0.5">
                                <p>Mode: <span className="font-mono">{getPinState(rightPin.bcm)?.mode}</span></p>
                                <p>Value: <span className={cn('font-bold', getPinState(rightPin.bcm)?.value ? 'text-green-500' : 'text-gray-400')}>
                                  {getPinState(rightPin.bcm)?.value ? 'HIGH' : 'LOW'}
                                </span></p>
                                <p>Edges: <span className="font-mono">{getPinState(rightPin.bcm)?.edge_count}</span></p>
                                <p>Changed: {formatLastChange(getPinState(rightPin.bcm)?.last_change || '')}</p>
                              </div>
                            )}
                          </div>
                        )}
                      </div>
                    )}
                  </div>
                )
              })}
            </div>

            {/* Legend */}
            <div className="mt-3 pt-3 border-t flex flex-wrap gap-3 justify-center text-[10px]">
              <div className="flex items-center gap-1">
                <div className="w-3 h-3 rounded bg-green-500 shadow-[0_0_4px_1px_rgba(34,197,94,0.5)]"></div>
                <span>HIGH</span>
              </div>
              <div className="flex items-center gap-1">
                <div className="w-3 h-3 rounded bg-gray-500"></div>
                <span>LOW</span>
              </div>
              <div className="flex items-center gap-1">
                <div className="w-3 h-3 rounded bg-purple-500 shadow-[0_0_4px_1px_rgba(168,85,247,0.5)]"></div>
                <span>PWM</span>
              </div>
              <div className="flex items-center gap-1">
                <div className="w-3 h-3 rounded bg-gray-500/40"></div>
                <span>Unused</span>
              </div>
              <div className="flex items-center gap-1">
                <div className="w-3 h-3 rounded bg-red-600"></div>
                <span>Power</span>
              </div>
              <div className="flex items-center gap-1">
                <div className="w-3 h-3 rounded bg-gray-800"></div>
                <span>Ground</span>
              </div>
            </div>
          </div>

          {/* Active Pins Detail Table */}
          {activePinCount > 0 && (
            <div className="w-64 bg-card border rounded-lg p-3">
              <p className="text-xs font-semibold mb-2">Active Pins</p>
              <div className="space-y-1.5">
                {Object.values(gpioState?.pins || {}).map((pin) => {
                  const gpioPin = GPIO_PINS.find(p => p.bcm === pin.bcm_pin)
                  return (
                    <div
                      key={pin.bcm_pin}
                      className="flex items-center justify-between px-2 py-1.5 rounded bg-muted/30 border border-border"
                    >
                      <div className="flex items-center gap-2">
                        <div className={cn(
                          'w-2.5 h-2.5 rounded-full transition-all duration-300',
                          pin.value
                            ? 'bg-green-500 shadow-[0_0_6px_1px_rgba(34,197,94,0.6)]'
                            : pin.mode === 'pwm'
                              ? 'bg-purple-500 shadow-[0_0_4px_1px_rgba(168,85,247,0.5)]'
                              : 'bg-gray-400'
                        )} />
                        <div>
                          <span className="text-[11px] font-medium">
                            {gpioPin?.name || `GPIO${pin.bcm_pin}`}
                          </span>
                          <span className="text-[9px] text-muted-foreground ml-1">
                            BCM {pin.bcm_pin}
                          </span>
                        </div>
                      </div>
                      <div className="text-right">
                        <span className={cn(
                          'text-[10px] font-bold',
                          pin.value ? 'text-green-500' : 'text-muted-foreground'
                        )}>
                          {pin.mode === 'pwm' ? 'PWM' : pin.value ? 'HIGH' : 'LOW'}
                        </span>
                        <p className="text-[9px] text-muted-foreground">
                          {pin.edge_count} edges
                        </p>
                      </div>
                    </div>
                  )
                })}
              </div>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
