/**
 * Setup Wizard Types
 *
 * Type definitions for the IoT board installation wizard
 */

export type BoardType =
  | 'rpi5'
  | 'rpi4'
  | 'rpi4-2gb'
  | 'rpi3b+'
  | 'rpi3b'
  | 'rpi-zero2w'
  | 'rpi-zero-w'
  | 'jetson-nano'
  | 'jetson-orin'
  | 'beaglebone'
  | 'orange-pi'
  | 'rock-pi'
  | 'custom'

export interface BoardInfo {
  id: BoardType
  name: string
  description: string
  gpioCount: number
  ram: string
  cpu: string
  supported: boolean
  hasWifi: boolean
  hasBluetooth: boolean
  wifiChip?: string
  ethernetSpeed?: string
  image?: string
}

/**
 * WiFi Security Types
 */
export type WifiSecurityType = 'open' | 'wep' | 'wpa' | 'wpa2' | 'wpa3' | 'wpa2-enterprise'

/**
 * Detected WiFi Network
 */
export interface WifiNetwork {
  ssid: string
  bssid: string
  signal: number // -100 to 0 dBm
  frequency: number // MHz (2.4GHz = 2400-2500, 5GHz = 5000-5900)
  channel: number
  security: WifiSecurityType
  connected: boolean
}

/**
 * IP Configuration Method
 */
export type IPMethod = 'dhcp' | 'static' | 'link-local'

/**
 * Network Interface Type
 */
export type NetworkInterfaceType = 'ethernet' | 'wifi' | 'usb-ethernet'

/**
 * Network Interface Configuration
 */
export interface NetworkInterfaceConfig {
  type: NetworkInterfaceType
  enabled: boolean
  ipMethod: IPMethod
  ipAddress?: string
  subnetMask?: string
  gateway?: string
  dns1?: string
  dns2?: string
  mtu?: number
}

/**
 * WiFi Configuration
 */
export interface WifiConfig {
  enabled: boolean
  ssid?: string
  password?: string
  security?: WifiSecurityType
  hidden?: boolean
  autoConnect: boolean
  band?: '2.4GHz' | '5GHz' | 'auto'
  country?: string
}

/**
 * Full Network Configuration
 */
export interface NetworkConfig {
  // Connection type
  primaryInterface: NetworkInterfaceType

  // Hostname
  hostname: string

  // Ethernet Configuration
  ethernet: NetworkInterfaceConfig

  // WiFi Configuration
  wifi: WifiConfig & NetworkInterfaceConfig

  // Legacy compatibility
  useWifi: boolean
  ssid?: string
  password?: string
  useStaticIP: boolean
  ipAddress?: string
  gateway?: string
  dns?: string
}

export interface MQTTConfig {
  enabled: boolean
  useBuiltIn: boolean
  externalBroker?: string
  externalPort?: number
  username?: string
  password?: string
  useTLS: boolean
}

export interface GPIOConfig {
  enableGPIO: boolean
  enableI2C: boolean
  enableSPI: boolean
  enableUART: boolean
  enable1Wire: boolean
  enablePWM: boolean
}

export interface SetupConfig {
  board: BoardType | null
  network: NetworkConfig
  mqtt: MQTTConfig
  gpio: GPIOConfig
}

export const DEFAULT_NETWORK_CONFIG: NetworkConfig = {
  primaryInterface: 'ethernet',
  hostname: 'edgeflow-device',

  ethernet: {
    type: 'ethernet',
    enabled: true,
    ipMethod: 'dhcp',
  },

  wifi: {
    type: 'wifi',
    enabled: false,
    ipMethod: 'dhcp',
    autoConnect: true,
    band: 'auto',
    country: 'US',
  },

  // Legacy compatibility
  useWifi: false,
  useStaticIP: false,
}

/**
 * WiFi Country Codes
 */
export const WIFI_COUNTRIES = [
  { code: 'US', name: 'United States' },
  { code: 'GB', name: 'United Kingdom' },
  { code: 'DE', name: 'Germany' },
  { code: 'FR', name: 'France' },
  { code: 'JP', name: 'Japan' },
  { code: 'CN', name: 'China' },
  { code: 'AU', name: 'Australia' },
  { code: 'CA', name: 'Canada' },
  { code: 'IN', name: 'India' },
  { code: 'BR', name: 'Brazil' },
  { code: 'IT', name: 'Italy' },
  { code: 'ES', name: 'Spain' },
  { code: 'NL', name: 'Netherlands' },
  { code: 'SE', name: 'Sweden' },
  { code: 'KR', name: 'South Korea' },
  { code: 'SG', name: 'Singapore' },
  { code: 'IR', name: 'Iran' },
]

/**
 * Common DNS Servers
 */
export const COMMON_DNS_SERVERS = [
  { name: 'Google DNS', primary: '8.8.8.8', secondary: '8.8.4.4' },
  { name: 'Cloudflare DNS', primary: '1.1.1.1', secondary: '1.0.0.1' },
  { name: 'OpenDNS', primary: '208.67.222.222', secondary: '208.67.220.220' },
  { name: 'Quad9', primary: '9.9.9.9', secondary: '149.112.112.112' },
  { name: 'AdGuard DNS', primary: '94.140.14.14', secondary: '94.140.15.15' },
]

/**
 * Common Subnet Masks
 */
export const COMMON_SUBNET_MASKS = [
  { mask: '255.255.255.0', cidr: '/24', hosts: '254 hosts' },
  { mask: '255.255.255.128', cidr: '/25', hosts: '126 hosts' },
  { mask: '255.255.255.192', cidr: '/26', hosts: '62 hosts' },
  { mask: '255.255.0.0', cidr: '/16', hosts: '65,534 hosts' },
  { mask: '255.0.0.0', cidr: '/8', hosts: '16M hosts' },
]

export const DEFAULT_MQTT_CONFIG: MQTTConfig = {
  enabled: true,
  useBuiltIn: true,
  useTLS: false,
}

export const DEFAULT_GPIO_CONFIG: GPIOConfig = {
  enableGPIO: true,
  enableI2C: true,
  enableSPI: false,
  enableUART: false,
  enable1Wire: false,
  enablePWM: true,
}

export const DEFAULT_SETUP_CONFIG: SetupConfig = {
  board: null,
  network: DEFAULT_NETWORK_CONFIG,
  mqtt: DEFAULT_MQTT_CONFIG,
  gpio: DEFAULT_GPIO_CONFIG,
}

export const SUPPORTED_BOARDS: BoardInfo[] = [
  {
    id: 'rpi5',
    name: 'Raspberry Pi 5',
    description: 'Latest generation with 2.4GHz quad-core CPU',
    gpioCount: 40,
    ram: '4GB/8GB',
    cpu: 'Cortex-A76 Quad-core',
    supported: true,
    hasWifi: true,
    hasBluetooth: true,
    wifiChip: 'Broadcom BCM43455 (802.11ac)',
    ethernetSpeed: 'Gigabit',
  },
  {
    id: 'rpi4',
    name: 'Raspberry Pi 4 Model B',
    description: 'Powerful board with USB 3.0 and dual HDMI',
    gpioCount: 40,
    ram: '4GB/8GB',
    cpu: 'Cortex-A72 Quad-core',
    supported: true,
    hasWifi: true,
    hasBluetooth: true,
    wifiChip: 'Broadcom BCM43455 (802.11ac)',
    ethernetSpeed: 'Gigabit',
  },
  {
    id: 'rpi4-2gb',
    name: 'Raspberry Pi 4 (2GB)',
    description: 'Budget version with 2GB RAM',
    gpioCount: 40,
    ram: '2GB',
    cpu: 'Cortex-A72 Quad-core',
    supported: true,
    hasWifi: true,
    hasBluetooth: true,
    wifiChip: 'Broadcom BCM43455 (802.11ac)',
    ethernetSpeed: 'Gigabit',
  },
  {
    id: 'rpi3b+',
    name: 'Raspberry Pi 3 Model B+',
    description: 'Popular choice with built-in WiFi',
    gpioCount: 40,
    ram: '1GB',
    cpu: 'Cortex-A53 Quad-core',
    supported: true,
    hasWifi: true,
    hasBluetooth: true,
    wifiChip: 'Broadcom BCM43455 (802.11ac)',
    ethernetSpeed: 'Gigabit (300Mbps)',
  },
  {
    id: 'rpi3b',
    name: 'Raspberry Pi 3 Model B',
    description: 'Original Pi 3 with 64-bit CPU',
    gpioCount: 40,
    ram: '1GB',
    cpu: 'Cortex-A53 Quad-core',
    supported: true,
    hasWifi: true,
    hasBluetooth: true,
    wifiChip: 'Broadcom BCM43438 (802.11n)',
    ethernetSpeed: '100Mbps',
  },
  {
    id: 'rpi-zero2w',
    name: 'Raspberry Pi Zero 2 W',
    description: 'Compact board with quad-core CPU',
    gpioCount: 40,
    ram: '512MB',
    cpu: 'Cortex-A53 Quad-core',
    supported: true,
    hasWifi: true,
    hasBluetooth: true,
    wifiChip: 'Broadcom BCM43436 (802.11n)',
    ethernetSpeed: 'None (USB adapter required)',
  },
  {
    id: 'rpi-zero-w',
    name: 'Raspberry Pi Zero W',
    description: 'Ultra-compact with WiFi (limited performance)',
    gpioCount: 40,
    ram: '512MB',
    cpu: 'ARM11 Single-core',
    supported: true,
    hasWifi: true,
    hasBluetooth: true,
    wifiChip: 'Broadcom BCM43438 (802.11n)',
    ethernetSpeed: 'None (USB adapter required)',
  },
  {
    id: 'jetson-nano',
    name: 'NVIDIA Jetson Nano',
    description: 'AI-capable edge device with GPU',
    gpioCount: 40,
    ram: '4GB',
    cpu: 'Cortex-A57 Quad-core + GPU',
    supported: true,
    hasWifi: false,
    hasBluetooth: false,
    ethernetSpeed: 'Gigabit',
  },
  {
    id: 'jetson-orin',
    name: 'NVIDIA Jetson Orin Nano',
    description: 'Advanced AI edge computing platform',
    gpioCount: 40,
    ram: '8GB',
    cpu: 'Cortex-A78AE + GPU',
    supported: true,
    hasWifi: true,
    hasBluetooth: true,
    wifiChip: 'Optional M.2 WiFi module',
    ethernetSpeed: 'Gigabit',
  },
  {
    id: 'beaglebone',
    name: 'BeagleBone Black',
    description: 'Industrial-grade with PRU processors',
    gpioCount: 65,
    ram: '512MB',
    cpu: 'Cortex-A8 Single-core',
    supported: true,
    hasWifi: false,
    hasBluetooth: false,
    ethernetSpeed: '100Mbps',
  },
  {
    id: 'orange-pi',
    name: 'Orange Pi 5',
    description: 'High-performance alternative',
    gpioCount: 40,
    ram: '4GB/8GB/16GB',
    cpu: 'Rockchip RK3588S',
    supported: true,
    hasWifi: true,
    hasBluetooth: true,
    wifiChip: 'Optional WiFi6 module',
    ethernetSpeed: '2.5 Gigabit',
  },
  {
    id: 'rock-pi',
    name: 'Rock Pi 4',
    description: 'Powerful board with PCIe support',
    gpioCount: 40,
    ram: '4GB',
    cpu: 'Rockchip RK3399',
    supported: true,
    hasWifi: true,
    hasBluetooth: true,
    wifiChip: 'AP6256 (802.11ac)',
    ethernetSpeed: 'Gigabit',
  },
  {
    id: 'custom',
    name: 'Custom/Other Board',
    description: 'Generic Linux-based board with GPIO support',
    gpioCount: 0,
    ram: 'Varies',
    cpu: 'Varies',
    supported: true,
    hasWifi: true,
    hasBluetooth: false,
  },
]

/**
 * Check if a board has built-in WiFi
 */
export function boardHasWifi(boardType: BoardType | null): boolean {
  if (!boardType) return false
  const board = SUPPORTED_BOARDS.find((b) => b.id === boardType)
  return board?.hasWifi ?? false
}

/**
 * Get board info by type
 */
export function getBoardInfo(boardType: BoardType | null): BoardInfo | null {
  if (!boardType) return null
  return SUPPORTED_BOARDS.find((b) => b.id === boardType) || null
}
