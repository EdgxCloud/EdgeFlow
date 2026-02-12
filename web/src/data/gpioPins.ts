/**
 * Raspberry Pi 5 GPIO Pin Data
 * Shared between GPIOPinSelector (configuration) and GPIOPanel (monitoring)
 */

export interface GPIOPin {
  physical: number
  bcm?: number
  name: string
  type: 'gpio' | 'power' | 'ground' | 'i2c' | 'spi' | 'uart' | 'reserved'
  voltage?: '3.3V' | '5V'
  pwmChannel?: number
}

// Raspberry Pi 5 GPIO pinout (40-pin header)
export const GPIO_PINS: GPIOPin[] = [
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

export const PIN_TYPE_COLORS: Record<GPIOPin['type'], string> = {
  gpio: 'bg-green-500 hover:bg-green-600',
  power: 'bg-red-500 hover:bg-red-600 cursor-not-allowed',
  ground: 'bg-black hover:bg-gray-800 cursor-not-allowed',
  i2c: 'bg-blue-500 hover:bg-blue-600',
  spi: 'bg-purple-500 hover:bg-purple-600',
  uart: 'bg-yellow-500 hover:bg-yellow-600',
  reserved: 'bg-gray-400 hover:bg-gray-500 cursor-not-allowed',
}
