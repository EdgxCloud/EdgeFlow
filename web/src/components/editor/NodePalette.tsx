import { useState, useEffect } from 'react'
import {
  Search, ChevronDown, ChevronRight, Zap, Download, Upload, Cpu, Database, Network,
  MessageSquare, Bot, Settings, Gauge, Thermometer, Server, Code, Bug, GitBranch,
  Clock, Scissors, Link2, AlertTriangle, CheckCircle, Activity, FileText,
  Terminal, Mail, Send, Sliders, Globe,
  Wifi, Radio, HardDrive, Cloud, Hash, Eye,
  Power, Rotate3D, Brain, Sparkles,
  Webhook, CircuitBoard, Factory, LayoutDashboard, Bluetooth,
  CreditCard, Smartphone, Camera, Volume2, BarChart3, Type, MousePointer,
  Keyboard, List, CalendarDays, Bell, Palette, Table, PenLine, Filter,
  MessageCircle, Lock, Box, Save, Braces, Timer, FileCode
} from 'lucide-react'
import { nodeTypesApi, NodeType } from '../../lib/api'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'

// Icon mapping for each node type
const NODE_ICONS: Record<string, any> = {
  // Input nodes
  'inject': Zap,
  'schedule': Clock,
  'mqtt-in': Download,
  'http-webhook': Webhook,
  'http-in': Webhook,
  'watch': Eye,
  'file-in': FileText,
  'serial-in': Terminal,
  'websocket-in': Radio,

  // Output nodes
  'debug': Bug,
  'mqtt-out': Upload,
  'http-response': Send,
  'file-out': Save,
  'serial-out': Terminal,
  'websocket-out': Send,

  // Function nodes
  'function': Code,
  'change': PenLine,
  'range': Sliders,
  'template': FileCode,
  'switch': GitBranch,
  'if': GitBranch,
  'delay': Clock,
  'split': Scissors,
  'join': Link2,
  'catch': AlertTriangle,
  'complete': CheckCircle,
  'status': Activity,
  'set': Settings,
  'link-in': Link2,
  'link-out': Link2,
  'trigger': Timer,
  'rbe': Filter,
  'filter': Filter,
  'comment': MessageSquare,

  // Processing nodes
  'exec': Terminal,
  'python': Code,
  'http-request': Globe,
  'websocket-client': Network,
  'tcp-client': Wifi,
  'udp': Radio,
  'json-parser': Braces,
  'xml-parser': Code,
  'csv-parser': Table,
  'yaml-parser': FileText,
  'html': Code,

  // GPIO nodes
  'gpio-in': CircuitBoard,
  'gpio-out': CircuitBoard,
  'pwm': Activity,
  'i2c': CircuitBoard,
  'spi': CircuitBoard,
  'serial': Terminal,
  'interrupt': Zap,
  'one-wire': CircuitBoard,

  // Sensors
  'dht': Thermometer,
  'ds18b20': Thermometer,
  'bmp280': Gauge,
  'bme280': Gauge,
  'bme680': Gauge,
  'sht3x': Thermometer,
  'aht20': Thermometer,
  'bh1750': Eye,
  'tsl2561': Eye,
  'veml7700': Eye,
  'ccs811': Activity,
  'sgp30': Activity,
  'hcsr04': Radio,
  'vl53l0x': Radio,
  'vl53l1x': Radio,
  'pir': Eye,
  'rcwl0516': Radio,
  'gps': Globe,
  'gps_neom8n': Globe,
  'compass_bn880': Globe,
  'mcp3008': Sliders,
  'ads1015': Sliders,
  'pcf8591': Sliders,
  'voltage-monitor': Gauge,
  'current-monitor': Gauge,
  'max31855': Thermometer,
  'max31865': Thermometer,

  // Actuators
  'relay': Power,
  'servo': Rotate3D,
  'motor_l298n': Rotate3D,
  'buzzer': Volume2,
  'ws2812': Sparkles,
  'lcd_i2c': Type,
  'oled_ssd1306': Type,

  // Communication
  'modbus': Cpu,
  'lora_sx1276': Radio,
  'nrf24l01': Radio,
  'rf433': Radio,
  'rfid_rc522': CreditCard,
  'nfc_pn532': Smartphone,
  'can_mcp2515': Cpu,
  'pi-camera': Camera,
  'audio': Volume2,

  // RTC
  'rtc_ds3231': Clock,
  'rtc_ds1307': Clock,
  'rtc_pcf8523': Clock,

  // Database nodes
  'mysql': Database,
  'postgresql': Database,
  'sqlite': Database,
  'mongodb': Database,
  'redis': Box,
  'influxdb': Database,

  // Storage nodes
  'google-drive': Cloud,
  'aws-s3': Cloud,
  'sftp': Lock,
  'dropbox': Cloud,
  'onedrive': Cloud,
  'ftp': Server,

  // Messaging nodes
  'email': Mail,
  'telegram': Send,
  'slack': MessageSquare,
  'discord': MessageCircle,

  // AI nodes
  'openai': Brain,
  'anthropic': Sparkles,
  'ollama': Bot,

  // Industrial nodes
  'modbus-tcp': Factory,
  'modbus-rtu': Factory,
  'opcua': Server,
  'bacnet': Factory,
  'profinet': Network,
  'can-bus': Cpu,

  // Wireless nodes
  'ble': Bluetooth,
  'zigbee': Radio,
  'zwave': Radio,
  'lora': Radio,
  'nrf24': Radio,
  'rfid': CreditCard,
  'nfc': Smartphone,
  'ir': Radio,

  // Dashboard nodes
  'dashboard-chart': BarChart3,
  'dashboard-gauge': Gauge,
  'dashboard-text': Type,
  'dashboard-button': MousePointer,
  'dashboard-slider': Sliders,
  'dashboard-switch': Power,
  'dashboard-text-input': Keyboard,
  'dashboard-dropdown': List,
  'dashboard-form': FileText,
  'dashboard-date-picker': CalendarDays,
  'dashboard-notification': Bell,
  'dashboard-template': Code,
  'dashboard-color-picker': Palette,
  'dashboard-table': Table,
}

// Category configuration matching backend
const CATEGORY_CONFIG: Record<string, { label: string; icon: any; color: string; description: string }> = {
  input: {
    label: 'Input',
    icon: Download,
    color: '#10b981',
    description: 'Nodes that receive data from external sources'
  },
  output: {
    label: 'Output',
    icon: Upload,
    color: '#ef4444',
    description: 'Nodes that send data to external destinations'
  },
  function: {
    label: 'Function',
    icon: Code,
    color: '#f59e0b',
    description: 'Logic, transformation, and data processing'
  },
  processing: {
    label: 'Processing',
    icon: Cpu,
    color: '#8b5cf6',
    description: 'Data processing and transformation nodes'
  },
  gpio: {
    label: 'GPIO',
    icon: CircuitBoard,
    color: '#16a34a',
    description: 'Raspberry Pi GPIO pins and basic I/O'
  },
  sensors: {
    label: 'Sensors',
    icon: Thermometer,
    color: '#22c55e',
    description: 'Temperature, humidity, light, and other sensors'
  },
  actuators: {
    label: 'Actuators',
    icon: Gauge,
    color: '#ec4899',
    description: 'Motors, relays, LEDs, and output devices'
  },
  communication: {
    label: 'Communication',
    icon: Radio,
    color: '#0ea5e9',
    description: 'LoRa, NRF24, RF433, and wireless protocols'
  },
  network: {
    label: 'Network',
    icon: Network,
    color: '#06b6d4',
    description: 'HTTP, MQTT, WebSocket, TCP/UDP protocols'
  },
  database: {
    label: 'Database',
    icon: Database,
    color: '#3b82f6',
    description: 'MySQL, PostgreSQL, MongoDB, Redis, InfluxDB'
  },
  storage: {
    label: 'Storage',
    icon: HardDrive,
    color: '#6366f1',
    description: 'File storage, S3, Google Drive, FTP'
  },
  messaging: {
    label: 'Messaging',
    icon: MessageSquare,
    color: '#14b8a6',
    description: 'Telegram, Email, Slack, Discord notifications'
  },
  ai: {
    label: 'AI & ML',
    icon: Brain,
    color: '#a855f7',
    description: 'OpenAI, Anthropic, Ollama LLM integration'
  },
  industrial: {
    label: 'Industrial',
    icon: Factory,
    color: '#f97316',
    description: 'Modbus RTU/TCP, OPC-UA, BACnet protocols'
  },
  wireless: {
    label: 'Wireless',
    icon: Bluetooth,
    color: '#0082fc',
    description: 'BLE, Zigbee, Z-Wave, LoRa, RFID, NFC'
  },
  dashboard: {
    label: 'Dashboard',
    icon: LayoutDashboard,
    color: '#0891b2',
    description: 'UI widgets, charts, gauges, buttons'
  },
  advanced: {
    label: 'Advanced',
    icon: Settings,
    color: '#64748b',
    description: 'System commands, file operations, utilities'
  },
}

// Fallback nodes shown when API is unavailable
const fallbackNodes: NodeType[] = [
  // Input
  { type: 'inject', name: 'Inject', description: 'Manually trigger flows or inject timestamps at intervals', category: 'input', color: '#10b981', icon: '', inputs: [], outputs: ['output'], properties: [] },
  { type: 'schedule', name: 'Schedule', description: 'Trigger flows based on cron expressions', category: 'input', color: '#10b981', icon: '', inputs: [], outputs: ['output'], properties: [] },
  { type: 'mqtt-in', name: 'MQTT In', description: 'Subscribe to MQTT broker topics and receive messages', category: 'input', color: '#10b981', icon: '', inputs: [], outputs: ['output'], properties: [] },
  { type: 'http-webhook', name: 'HTTP Webhook', description: 'Receive HTTP requests via webhook endpoint', category: 'input', color: '#10b981', icon: '', inputs: [], outputs: ['output'], properties: [] },
  { type: 'http-in', name: 'HTTP In', description: 'Create an HTTP endpoint to receive requests', category: 'input', color: '#10b981', icon: '', inputs: [], outputs: ['output'], properties: [] },
  { type: 'watch', name: 'File Watch', description: 'Monitor file/directory changes', category: 'input', color: '#10b981', icon: '', inputs: [], outputs: ['output'], properties: [] },
  { type: 'file-in', name: 'File Read', description: 'Read file contents', category: 'input', color: '#10b981', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'serial-in', name: 'Serial In', description: 'Receive data from serial port', category: 'input', color: '#10b981', icon: '', inputs: [], outputs: ['output'], properties: [] },
  { type: 'websocket-in', name: 'WebSocket In', description: 'Accept WebSocket connections and receive messages', category: 'input', color: '#10b981', icon: '', inputs: [], outputs: ['output'], properties: [] },

  // Output
  { type: 'debug', name: 'Debug', description: 'Display messages in debug sidebar for troubleshooting', category: 'output', color: '#ef4444', icon: '', inputs: ['input'], outputs: [], properties: [] },
  { type: 'mqtt-out', name: 'MQTT Out', description: 'Publish messages to MQTT broker topics', category: 'output', color: '#ef4444', icon: '', inputs: ['input'], outputs: [], properties: [] },
  { type: 'http-response', name: 'HTTP Response', description: 'Send HTTP response', category: 'output', color: '#ef4444', icon: '', inputs: ['input'], outputs: [], properties: [] },
  { type: 'file-out', name: 'File Write', description: 'Write content to file', category: 'output', color: '#ef4444', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'serial-out', name: 'Serial Out', description: 'Send data to serial port', category: 'output', color: '#ef4444', icon: '', inputs: ['input'], outputs: [], properties: [] },
  { type: 'websocket-out', name: 'WebSocket Out', description: 'Send messages to WebSocket clients', category: 'output', color: '#ef4444', icon: '', inputs: ['input'], outputs: [], properties: [] },

  // Function
  { type: 'function', name: 'Function', description: 'Transform message data using rules (set, delete properties)', category: 'function', color: '#f59e0b', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'change', name: 'Change', description: 'Set, change, move or delete message properties', category: 'function', color: '#f59e0b', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'range', name: 'Range', description: 'Scale numeric values between ranges', category: 'function', color: '#f59e0b', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'template', name: 'Template', description: 'Render Mustache templates', category: 'function', color: '#f59e0b', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'switch', name: 'Switch', description: 'Route messages based on property values and rules', category: 'function', color: '#f59e0b', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'if', name: 'If', description: 'Route messages based on conditions', category: 'function', color: '#f59e0b', icon: '', inputs: ['input'], outputs: ['true', 'false'], properties: [] },
  { type: 'delay', name: 'Delay', description: 'Delay message processing', category: 'function', color: '#f59e0b', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'split', name: 'Split', description: 'Split messages into multiple parts', category: 'function', color: '#f59e0b', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'join', name: 'Join', description: 'Join multiple messages into one', category: 'function', color: '#f59e0b', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'catch', name: 'Catch', description: 'Catch errors from other nodes', category: 'function', color: '#f59e0b', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'status', name: 'Status', description: 'Monitor node status changes', category: 'function', color: '#f59e0b', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'complete', name: 'Complete', description: 'Trigger when node or flow completes', category: 'function', color: '#f59e0b', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'set', name: 'Set', description: 'Set, delete, move message properties', category: 'function', color: '#f59e0b', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'trigger', name: 'Trigger', description: 'Send message, then optionally send second after delay', category: 'function', color: '#f59e0b', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'rbe', name: 'RBE', description: 'Report by exception - only pass changed values', category: 'function', color: '#f59e0b', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'filter', name: 'Filter', description: 'Filter messages based on conditions', category: 'function', color: '#f59e0b', icon: '', inputs: ['input'], outputs: ['match', 'no_match'], properties: [] },
  { type: 'link-in', name: 'Link In', description: 'Receive messages from Link Out nodes', category: 'function', color: '#f59e0b', icon: '', inputs: [], outputs: ['output'], properties: [] },
  { type: 'link-out', name: 'Link Out', description: 'Send messages to Link In nodes', category: 'function', color: '#f59e0b', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'comment', name: 'Comment', description: 'Add documentation notes to flows', category: 'function', color: '#f59e0b', icon: '', inputs: [], outputs: [], properties: [] },

  // Processing
  { type: 'exec', name: 'Exec', description: 'Execute shell commands', category: 'function', color: '#f59e0b', icon: '', inputs: ['input'], outputs: ['stdout', 'stderr', 'return'], properties: [] },
  { type: 'python', name: 'Python', description: 'Execute Python code', category: 'function', color: '#f59e0b', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'http-request', name: 'HTTP Request', description: 'Send HTTP requests with all methods', category: 'processing', color: '#8b5cf6', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'websocket-client', name: 'WebSocket Client', description: 'Connect to WebSocket server', category: 'processing', color: '#8b5cf6', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'tcp-client', name: 'TCP Client', description: 'Connect to TCP server', category: 'processing', color: '#8b5cf6', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'udp', name: 'UDP', description: 'Send and receive UDP packets', category: 'processing', color: '#8b5cf6', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'json-parser', name: 'JSON Parser', description: 'Convert JSON to object and vice versa', category: 'processing', color: '#8b5cf6', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'xml-parser', name: 'XML Parser', description: 'Convert XML to JSON and vice versa', category: 'processing', color: '#8b5cf6', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'csv-parser', name: 'CSV Parser', description: 'Convert CSV to JSON and vice versa', category: 'processing', color: '#8b5cf6', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'yaml-parser', name: 'YAML Parser', description: 'Convert YAML to JSON and vice versa', category: 'processing', color: '#8b5cf6', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'html', name: 'HTML Parser', description: 'Parse HTML with CSS selectors', category: 'function', color: '#f59e0b', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },

  // GPIO
  { type: 'gpio-in', name: 'GPIO In', description: 'Read digital values from GPIO pins', category: 'gpio', color: '#16a34a', icon: '', inputs: [], outputs: ['output'], properties: [] },
  { type: 'gpio-out', name: 'GPIO Out', description: 'Write digital values to GPIO pins', category: 'gpio', color: '#16a34a', icon: '', inputs: ['input'], outputs: [], properties: [] },
  { type: 'pwm', name: 'PWM', description: 'Generate pulse width modulation signals', category: 'gpio', color: '#16a34a', icon: '', inputs: ['input'], outputs: [], properties: [] },
  { type: 'i2c', name: 'I2C', description: 'I2C bus communication', category: 'gpio', color: '#16a34a', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'spi', name: 'SPI', description: 'SPI bus communication', category: 'gpio', color: '#16a34a', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'serial', name: 'Serial', description: 'Serial port communication', category: 'gpio', color: '#16a34a', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'interrupt', name: 'GPIO Interrupt', description: 'GPIO interrupt detection', category: 'gpio', color: '#16a34a', icon: '', inputs: [], outputs: ['output'], properties: [] },
  { type: 'one-wire', name: '1-Wire', description: '1-Wire bus communication', category: 'gpio', color: '#16a34a', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },

  // Sensors
  { type: 'dht', name: 'DHT Sensor', description: 'Read temperature/humidity from DHT11/DHT22', category: 'sensors', color: '#22c55e', icon: '', inputs: [], outputs: ['output'], properties: [] },
  { type: 'ds18b20', name: 'DS18B20', description: 'Read temperature from DS18B20', category: 'sensors', color: '#22c55e', icon: '', inputs: [], outputs: ['output'], properties: [] },
  { type: 'bmp280', name: 'BMP280', description: 'Temperature and pressure sensor', category: 'sensors', color: '#22c55e', icon: '', inputs: [], outputs: ['output'], properties: [] },
  { type: 'bme280', name: 'BME280', description: 'Temperature, humidity and pressure sensor', category: 'sensors', color: '#22c55e', icon: '', inputs: [], outputs: ['output'], properties: [] },
  { type: 'bme680', name: 'BME680', description: 'Temperature, humidity, pressure and gas sensor', category: 'sensors', color: '#22c55e', icon: '', inputs: [], outputs: ['output'], properties: [] },
  { type: 'sht3x', name: 'SHT3x', description: 'High-accuracy temperature and humidity', category: 'sensors', color: '#22c55e', icon: '', inputs: [], outputs: ['output'], properties: [] },
  { type: 'aht20', name: 'AHT20', description: 'Temperature and humidity sensor', category: 'sensors', color: '#22c55e', icon: '', inputs: [], outputs: ['output'], properties: [] },
  { type: 'bh1750', name: 'BH1750', description: 'Ambient light sensor (lux)', category: 'sensors', color: '#22c55e', icon: '', inputs: [], outputs: ['output'], properties: [] },
  { type: 'tsl2561', name: 'TSL2561', description: 'Light sensor with IR channel', category: 'sensors', color: '#22c55e', icon: '', inputs: [], outputs: ['output'], properties: [] },
  { type: 'veml7700', name: 'VEML7700', description: 'High-accuracy ambient light sensor', category: 'sensors', color: '#22c55e', icon: '', inputs: [], outputs: ['output'], properties: [] },
  { type: 'ccs811', name: 'CCS811', description: 'CO2 and TVOC air quality sensor', category: 'sensors', color: '#22c55e', icon: '', inputs: [], outputs: ['output'], properties: [] },
  { type: 'sgp30', name: 'SGP30', description: 'Indoor air quality sensor', category: 'sensors', color: '#22c55e', icon: '', inputs: [], outputs: ['output'], properties: [] },
  { type: 'hcsr04', name: 'HC-SR04', description: 'Ultrasonic distance sensor', category: 'sensors', color: '#22c55e', icon: '', inputs: [], outputs: ['output'], properties: [] },
  { type: 'vl53l0x', name: 'VL53L0X', description: 'Laser distance sensor (ToF)', category: 'sensors', color: '#22c55e', icon: '', inputs: [], outputs: ['output'], properties: [] },
  { type: 'vl53l1x', name: 'VL53L1X', description: 'Long-range laser distance sensor', category: 'sensors', color: '#22c55e', icon: '', inputs: [], outputs: ['output'], properties: [] },
  { type: 'pir', name: 'PIR Motion', description: 'Passive infrared motion detector', category: 'sensors', color: '#22c55e', icon: '', inputs: [], outputs: ['output'], properties: [] },
  { type: 'rcwl0516', name: 'RCWL-0516', description: 'Microwave motion sensor', category: 'sensors', color: '#22c55e', icon: '', inputs: [], outputs: ['output'], properties: [] },
  { type: 'gps', name: 'GPS', description: 'GPS location and time data', category: 'sensors', color: '#22c55e', icon: '', inputs: [], outputs: ['output'], properties: [] },
  { type: 'gps_neom8n', name: 'GPS NEO-M8N', description: 'u-blox NEO-M8N GPS module', category: 'sensors', color: '#22c55e', icon: '', inputs: [], outputs: ['output'], properties: [] },
  { type: 'compass_bn880', name: 'BN-880 GPS+Compass', description: 'GPS and compass module', category: 'sensors', color: '#22c55e', icon: '', inputs: [], outputs: ['output'], properties: [] },
  { type: 'mcp3008', name: 'MCP3008', description: '10-bit ADC (8 channels)', category: 'sensors', color: '#22c55e', icon: '', inputs: [], outputs: ['output'], properties: [] },
  { type: 'ads1015', name: 'ADS1015', description: '12-bit ADC with PGA', category: 'sensors', color: '#22c55e', icon: '', inputs: [], outputs: ['output'], properties: [] },
  { type: 'pcf8591', name: 'PCF8591', description: '8-bit ADC/DAC converter', category: 'sensors', color: '#22c55e', icon: '', inputs: [], outputs: ['output'], properties: [] },
  { type: 'voltage-monitor', name: 'Voltage Monitor', description: 'Monitor voltage levels', category: 'sensors', color: '#22c55e', icon: '', inputs: [], outputs: ['output'], properties: [] },
  { type: 'current-monitor', name: 'Current Monitor', description: 'Monitor current draw', category: 'sensors', color: '#22c55e', icon: '', inputs: [], outputs: ['output'], properties: [] },
  { type: 'max31855', name: 'MAX31855', description: 'Thermocouple amplifier (K-type)', category: 'sensors', color: '#22c55e', icon: '', inputs: [], outputs: ['output'], properties: [] },
  { type: 'max31865', name: 'MAX31865', description: 'RTD-to-digital converter (PT100/PT1000)', category: 'sensors', color: '#22c55e', icon: '', inputs: [], outputs: ['output'], properties: [] },

  // Actuators
  { type: 'relay', name: 'Relay', description: 'Control relay switches', category: 'actuators', color: '#ec4899', icon: '', inputs: ['input'], outputs: [], properties: [] },
  { type: 'servo', name: 'Servo Motor', description: 'Control servo motor position', category: 'actuators', color: '#ec4899', icon: '', inputs: ['input'], outputs: [], properties: [] },
  { type: 'motor_l298n', name: 'Motor L298N', description: 'DC motor driver (H-bridge)', category: 'actuators', color: '#ec4899', icon: '', inputs: ['input'], outputs: [], properties: [] },
  { type: 'buzzer', name: 'Buzzer', description: 'Piezo buzzer with tone control', category: 'actuators', color: '#ec4899', icon: '', inputs: ['input'], outputs: [], properties: [] },
  { type: 'ws2812', name: 'WS2812 LED Strip', description: 'Addressable RGB LED strip control', category: 'actuators', color: '#ec4899', icon: '', inputs: ['input'], outputs: [], properties: [] },
  { type: 'lcd_i2c', name: 'LCD I2C', description: 'I2C LCD display (16x2, 20x4)', category: 'actuators', color: '#ec4899', icon: '', inputs: ['input'], outputs: [], properties: [] },
  { type: 'oled_ssd1306', name: 'OLED SSD1306', description: 'SSD1306 OLED display', category: 'actuators', color: '#ec4899', icon: '', inputs: ['input'], outputs: [], properties: [] },

  // Database
  { type: 'mysql', name: 'MySQL', description: 'Execute queries on MySQL', category: 'database', color: '#3b82f6', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'postgresql', name: 'PostgreSQL', description: 'Execute queries on PostgreSQL', category: 'database', color: '#3b82f6', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'sqlite', name: 'SQLite', description: 'Execute queries on SQLite', category: 'database', color: '#3b82f6', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'mongodb', name: 'MongoDB', description: 'MongoDB operations', category: 'database', color: '#3b82f6', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'redis', name: 'Redis', description: 'Redis cache operations', category: 'database', color: '#3b82f6', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'influxdb', name: 'InfluxDB', description: 'InfluxDB time-series operations', category: 'database', color: '#3b82f6', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },

  // Storage
  { type: 'google-drive', name: 'Google Drive', description: 'Upload/download files on Google Drive', category: 'storage', color: '#6366f1', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'aws-s3', name: 'AWS S3', description: 'Upload/download objects in AWS S3', category: 'storage', color: '#6366f1', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'sftp', name: 'SFTP', description: 'Upload/download files via SFTP', category: 'storage', color: '#6366f1', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'dropbox', name: 'Dropbox', description: 'Manage files on Dropbox', category: 'storage', color: '#6366f1', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'onedrive', name: 'OneDrive', description: 'Manage files on Microsoft OneDrive', category: 'storage', color: '#6366f1', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'ftp', name: 'FTP', description: 'Upload/download files via FTP', category: 'storage', color: '#6366f1', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },

  // Messaging
  { type: 'email', name: 'Email', description: 'Send email via SMTP', category: 'messaging', color: '#14b8a6', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'telegram', name: 'Telegram', description: 'Send/receive Telegram bot messages', category: 'messaging', color: '#14b8a6', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'slack', name: 'Slack', description: 'Send messages to Slack', category: 'messaging', color: '#14b8a6', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'discord', name: 'Discord', description: 'Send messages to Discord', category: 'messaging', color: '#14b8a6', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },

  // AI
  { type: 'openai', name: 'OpenAI', description: 'Text generation with OpenAI GPT', category: 'ai', color: '#a855f7', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'anthropic', name: 'Anthropic', description: 'Text generation with Claude', category: 'ai', color: '#a855f7', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'ollama', name: 'Ollama', description: 'Local LLM inference with Ollama', category: 'ai', color: '#a855f7', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },

  // Industrial
  { type: 'modbus-tcp', name: 'Modbus TCP', description: 'Modbus TCP client for PLCs', category: 'industrial', color: '#f97316', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'modbus-rtu', name: 'Modbus RTU', description: 'Modbus RTU over serial port', category: 'industrial', color: '#f97316', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'opcua', name: 'OPC-UA', description: 'OPC-UA client for industrial automation', category: 'industrial', color: '#f97316', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'bacnet', name: 'BACnet', description: 'BACnet/IP for building automation', category: 'industrial', color: '#f97316', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'profinet', name: 'PROFINET', description: 'PROFINET DCP discovery and I/O', category: 'industrial', color: '#f97316', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'can-bus', name: 'CAN Bus', description: 'CAN bus communication via SocketCAN', category: 'industrial', color: '#f97316', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },

  // Wireless
  { type: 'ble', name: 'Bluetooth LE', description: 'Bluetooth Low Energy communication', category: 'wireless', color: '#0082fc', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'zigbee', name: 'Zigbee', description: 'Zigbee via zigbee2mqtt or deCONZ', category: 'wireless', color: '#0082fc', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'zwave', name: 'Z-Wave', description: 'Z-Wave via zwave2mqtt', category: 'wireless', color: '#0082fc', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'lora', name: 'LoRa', description: 'LoRa long-range wireless', category: 'wireless', color: '#0082fc', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'nrf24', name: 'NRF24L01', description: 'NRF24L01 2.4GHz transceiver', category: 'wireless', color: '#0082fc', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'rfid', name: 'RFID RC522', description: 'RFID reader for MIFARE cards', category: 'wireless', color: '#0082fc', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'nfc', name: 'NFC PN532', description: 'NFC reader/writer', category: 'wireless', color: '#0082fc', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'ir', name: 'IR Transceiver', description: 'Infrared transmit/receive', category: 'wireless', color: '#0082fc', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },

  // Dashboard
  { type: 'dashboard-chart', name: 'Chart', description: 'Display data as charts', category: 'dashboard', color: '#0891b2', icon: '', inputs: ['input'], outputs: [], properties: [] },
  { type: 'dashboard-gauge', name: 'Gauge', description: 'Display numeric values as gauges', category: 'dashboard', color: '#0891b2', icon: '', inputs: ['input'], outputs: [], properties: [] },
  { type: 'dashboard-text', name: 'Text Display', description: 'Display text on dashboard', category: 'dashboard', color: '#0891b2', icon: '', inputs: ['input'], outputs: [], properties: [] },
  { type: 'dashboard-button', name: 'Button', description: 'Interactive button widget', category: 'dashboard', color: '#0891b2', icon: '', inputs: [], outputs: ['output'], properties: [] },
  { type: 'dashboard-slider', name: 'Slider', description: 'Interactive slider input', category: 'dashboard', color: '#0891b2', icon: '', inputs: [], outputs: ['output'], properties: [] },
  { type: 'dashboard-switch', name: 'Switch', description: 'Interactive toggle switch', category: 'dashboard', color: '#0891b2', icon: '', inputs: [], outputs: ['output'], properties: [] },
  { type: 'dashboard-text-input', name: 'Text Input', description: 'Text input field', category: 'dashboard', color: '#0891b2', icon: '', inputs: [], outputs: ['output'], properties: [] },
  { type: 'dashboard-dropdown', name: 'Dropdown', description: 'Dropdown select input', category: 'dashboard', color: '#0891b2', icon: '', inputs: [], outputs: ['output'], properties: [] },
  { type: 'dashboard-form', name: 'Form', description: 'Form builder with multiple fields', category: 'dashboard', color: '#0891b2', icon: '', inputs: [], outputs: ['output'], properties: [] },
  { type: 'dashboard-date-picker', name: 'Date Picker', description: 'Date and time picker', category: 'dashboard', color: '#0891b2', icon: '', inputs: [], outputs: ['output'], properties: [] },
  { type: 'dashboard-notification', name: 'Notification', description: 'Display toast notifications', category: 'dashboard', color: '#0891b2', icon: '', inputs: ['input'], outputs: [], properties: [] },
  { type: 'dashboard-template', name: 'Template', description: 'Custom HTML template content', category: 'dashboard', color: '#0891b2', icon: '', inputs: ['input'], outputs: [], properties: [] },
  { type: 'dashboard-color-picker', name: 'Color Picker', description: 'Color picker input', category: 'dashboard', color: '#0891b2', icon: '', inputs: ['input'], outputs: ['output'], properties: [] },
  { type: 'dashboard-table', name: 'Table', description: 'Data table with sorting and filtering', category: 'dashboard', color: '#0891b2', icon: '', inputs: ['input'], outputs: [], properties: [] },
]

export default function NodePalette() {
  const [nodeTypes, setNodeTypes] = useState<NodeType[]>(fallbackNodes)
  const [search, setSearch] = useState('')
  const [loading, setLoading] = useState(true)
  const [collapsedCategories, setCollapsedCategories] = useState<Set<string>>(new Set())

  useEffect(() => {
    loadNodeTypes()
  }, [])

  const loadNodeTypes = async () => {
    try {
      const response = await nodeTypesApi.list()
      const apiNodes = response.data.node_types || []

      // Merge API nodes with fallback nodes
      // Fallback nodes provide a baseline, API nodes can override or add more
      const apiNodeTypes = new Set(apiNodes.map((n: NodeType) => n.type))
      const mergedNodes = [
        ...fallbackNodes.filter(n => !apiNodeTypes.has(n.type)), // Keep fallback nodes not in API
        ...apiNodes // Add all API nodes
      ]

      setNodeTypes(mergedNodes)
    } catch (error) {
      console.error('Failed to load node types:', error)
      setNodeTypes(fallbackNodes)
    } finally {
      setLoading(false)
    }
  }

  const onDragStart = (event: React.DragEvent, nodeType: string) => {
    event.dataTransfer.setData('application/reactflow', nodeType)
    event.dataTransfer.effectAllowed = 'move'
  }

  const toggleCategory = (category: string) => {
    const newCollapsed = new Set(collapsedCategories)
    if (newCollapsed.has(category)) {
      newCollapsed.delete(category)
    } else {
      newCollapsed.add(category)
    }
    setCollapsedCategories(newCollapsed)
  }

  const filteredNodes = nodeTypes.filter((node) =>
    node.name.toLowerCase().includes(search.toLowerCase()) ||
    node.description.toLowerCase().includes(search.toLowerCase())
  )

  const groupedNodes = filteredNodes.reduce((acc, node) => {
    const category = node.category || 'advanced'
    if (!acc[category]) {
      acc[category] = []
    }
    acc[category].push(node)
    return acc
  }, {} as Record<string, NodeType[]>)

  // Sort nodes within each category alphabetically
  Object.keys(groupedNodes).forEach((category) => {
    groupedNodes[category].sort((a, b) => a.name.localeCompare(b.name))
  })

  // Sort categories by predefined order (matches backend)
  const categoryOrder = ['input', 'output', 'function', 'processing', 'gpio', 'sensors', 'actuators', 'communication', 'network', 'database', 'storage', 'messaging', 'ai', 'industrial', 'wireless', 'dashboard', 'advanced']
  const sortedCategories = Object.keys(groupedNodes).sort((a, b) => {
    const indexA = categoryOrder.indexOf(a)
    const indexB = categoryOrder.indexOf(b)
    if (indexA === -1) return 1
    if (indexB === -1) return -1
    return indexA - indexB
  })

  return (
    <TooltipProvider delayDuration={300}>
      <div className="w-64 bg-white dark:bg-gray-800 border-l border-gray-200 dark:border-gray-700 flex flex-col h-full">
        {/* Header */}
        <div className="p-4 border-b border-gray-200 dark:border-gray-700">
          <h3 className="font-semibold text-gray-900 dark:text-white mb-3">
            Node Palette
          </h3>
          <div className="relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 w-4 h-4 text-gray-400" aria-hidden="true" />
            <input
              type="text"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Search nodes..."
              aria-label="Search nodes"
              className="w-full pl-10 pr-3 py-2 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-700 text-gray-900 dark:text-white text-sm focus:ring-2 focus:ring-blue-500 focus:border-transparent"
            />
          </div>
        </div>

        {/* Node list */}
        <div className="flex-1 overflow-y-auto p-3 space-y-2" role="list" aria-label="Available nodes">
          {loading ? (
            <div className="flex items-center justify-center h-32">
              <div className="spinner" aria-label="Loading nodes"></div>
            </div>
          ) : (
            sortedCategories.map((category) => {
              const config = CATEGORY_CONFIG[category] || {
                label: category,
                icon: Settings,
                color: '#64748b',
                description: category
              }
              const Icon = config.icon
              const isCollapsed = collapsedCategories.has(category)
              const nodes = groupedNodes[category]

              return (
                <div key={category} className="space-y-1">
                  {/* Category header - collapsible */}
                  <button
                    onClick={() => toggleCategory(category)}
                    className="w-full flex items-center gap-2 px-2 py-1.5 rounded hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors group"
                    aria-expanded={!isCollapsed}
                    aria-controls={`category-${category}`}
                  >
                    {isCollapsed ? (
                      <ChevronRight className="w-3.5 h-3.5 text-gray-500" aria-hidden="true" />
                    ) : (
                      <ChevronDown className="w-3.5 h-3.5 text-gray-500" aria-hidden="true" />
                    )}
                    <Icon className="w-4 h-4" style={{ color: config.color }} aria-hidden="true" />
                    <span className="text-xs font-semibold text-gray-700 dark:text-gray-300 uppercase tracking-wider">
                      {config.label}
                    </span>
                    <span className="ml-auto text-xs text-gray-400 dark:text-gray-500">
                      {nodes.length}
                    </span>
                  </button>

                  {/* Category nodes */}
                  {!isCollapsed && (
                    <div id={`category-${category}`} className="space-y-1 pl-2" role="group">
                      {nodes.map((node) => (
                        <Tooltip key={node.type}>
                          <TooltipTrigger asChild>
                            <div
                              draggable
                              onDragStart={(e) => onDragStart(e, node.type)}
                              className="p-2 rounded-md border border-gray-200 dark:border-gray-600 hover:border-blue-500 dark:hover:border-blue-500 hover:shadow-sm cursor-move transition-all group"
                              style={{ backgroundColor: node.color + '08' }}
                              role="listitem"
                              tabIndex={0}
                              aria-label={`${node.name}: ${node.description}`}
                              onKeyDown={(e) => {
                                if (e.key === 'Enter' || e.key === ' ') {
                                  e.preventDefault()
                                }
                              }}
                            >
                              <div className="flex items-center gap-2.5">
                                {(() => {
                                  const NodeIcon = NODE_ICONS[node.type] || Activity
                                  return (
                                    <div
                                      className="w-7 h-7 rounded flex-shrink-0 flex items-center justify-center text-white"
                                      style={{ backgroundColor: node.color }}
                                      aria-hidden="true"
                                    >
                                      <NodeIcon className="w-4 h-4" strokeWidth={2.5} />
                                    </div>
                                  )
                                })()}
                                <span className="flex-1 min-w-0 font-medium text-xs text-gray-900 dark:text-white truncate">
                                  {node.name}
                                </span>
                              </div>
                            </div>
                          </TooltipTrigger>
                          <TooltipContent side="left" className="max-w-xs">
                            <div className="space-y-1">
                              <p className="font-semibold">{node.name}</p>
                              <p className="text-xs opacity-90">{node.description}</p>
                              {node.inputs && node.inputs.length > 0 && (
                                <p className="text-xs opacity-75">Inputs: {node.inputs.length}</p>
                              )}
                              {node.outputs && node.outputs.length > 0 && (
                                <p className="text-xs opacity-75">Outputs: {node.outputs.length}</p>
                              )}
                            </div>
                          </TooltipContent>
                        </Tooltip>
                      ))}
                    </div>
                  )}
                </div>
              )
            })
          )}

          {!loading && filteredNodes.length === 0 && (
            <div className="text-center py-8 text-gray-500 dark:text-gray-400 text-sm">
              No nodes found matching "{search}"
            </div>
          )}
        </div>
      </div>
    </TooltipProvider>
  )
}
