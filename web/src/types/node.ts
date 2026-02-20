/**
 * Node Configuration Types
 *
 * TypeScript interfaces for EdgeFlow node system
 */

export type PropertyType =
  | 'string'
  | 'number'
  | 'boolean'
  | 'select'
  | 'array'
  | 'object'
  | 'color'
  | 'icon'
  | 'json'
  | 'code'
  | 'cron'
  | 'gpio-pin'
  | 'mqtt-topic'
  | 'mqtt-config' // Full MQTT configuration including broker, topic, QoS, etc.
  | 'payload'     // Key-value payload builder with presets
  | 'password'    // Masked password input
  | 'any'         // Generic type, renders as text input

export interface PropertySchema {
  name: string
  label: string
  type: PropertyType
  required?: boolean
  default?: any
  options?: string[] | { label: string; value: string }[]
  placeholder?: string
  min?: number
  max?: number
  step?: number
  validation?: string // regex pattern
  description?: string
  group?: string
}

export interface PortSchema {
  name: string
  label: string
  type: string
  description?: string
}

export type NodeCategory =
  | 'input'
  | 'output'
  | 'function'
  | 'processing'
  | 'gpio'
  | 'sensors'
  | 'actuators'
  | 'communication'
  | 'network'
  | 'database'
  | 'storage'
  | 'messaging'
  | 'ai'
  | 'industrial'
  | 'wireless'
  | 'dashboard'
  | 'parser'
  | 'advanced'

export interface NodeInfo {
  type: string
  name: string
  category: NodeCategory
  description: string
  icon: string
  color: string
  properties: PropertySchema[]
  inputs?: PortSchema[]
  outputs?: PortSchema[]
  version?: string
  author?: string
}

// n8n-style position format: [x, y] array
export type Position = [number, number]

export interface NodeConfig {
  id: string
  type: string
  name?: string
  config: Record<string, any>
  // n8n-style position as [x, y] array
  position?: Position
  [key: string]: any
}

// Dashboard Widget Types
export type WidgetType =
  | 'chart'
  | 'gauge'
  | 'text'
  | 'button'
  | 'slider'
  | 'switch'
  | 'text-input'
  | 'dropdown'
  | 'form'
  | 'table'
  | 'date-picker'
  | 'notification'
  | 'template'

export type ChartType = 'line' | 'bar' | 'pie' | 'histogram' | 'scatter'

export interface ChartDataPoint {
  x?: number | string
  y?: number
  label?: string
  value?: number
  timestamp?: number
}

export interface ChartSeries {
  name: string
  color: string
  type?: ChartType
  data?: ChartDataPoint[]
}

export interface GaugeSector {
  from: number
  to: number
  color: string
  label?: string
}

export interface FormField {
  name: string
  label: string
  type: 'text' | 'number' | 'email' | 'password' | 'checkbox' | 'radio' | 'select' | 'textarea'
  required?: boolean
  placeholder?: string
  default?: any
  options?: string[] | { label: string; value: string }[]
  validation?: string
}

export interface TableColumn {
  field: string
  header: string
  width?: string
  sortable?: boolean
  filterable?: boolean
  format?: string
}

// GPIO Pin Types
export type PinType = 'power' | 'ground' | 'gpio' | 'special'

export interface Pin {
  physical: number
  bcm?: number
  wiringPi?: number
  name: string
  function: string
  type: PinType
  color: string
  available?: boolean
  inUse?: boolean
}

// MQTT Configuration Types (based on Node-RED and n8n patterns)

// MQTT Output format options
export type MQTTOutputFormat = 'auto' | 'string' | 'buffer' | 'json'

// MQTT Protocol versions
export type MQTTProtocolVersion = '3.1.1' | '5.0'

// MQTT Broker configuration (shared between mqtt-in and mqtt-out nodes)
export interface MQTTBrokerConfig {
  server?: string
  port?: number
  clientId?: string
  username?: string
  password?: string
  useTLS?: boolean
  protocolVersion?: MQTTProtocolVersion
  keepAlive?: number // in seconds (0-65535)
  cleanSession?: boolean
  reconnectPeriod?: number // in milliseconds
}

// Full MQTT node configuration
export interface MQTTConfig {
  // Broker settings
  broker?: MQTTBrokerConfig
  // Legacy flat broker config (for backwards compatibility)
  server?: string
  port?: number
  clientId?: string
  username?: string
  password?: string
  // Topic settings
  topic: string
  qos?: 0 | 1 | 2
  retain?: boolean
  wildcardMode?: boolean
  // Subscribe-specific settings
  outputFormat?: MQTTOutputFormat
  dynamicSubscription?: boolean
}

// MQTT Subscribe node configuration
export interface MQTTInConfig extends MQTTConfig {
  outputFormat?: MQTTOutputFormat
  dynamicSubscription?: boolean
}

// MQTT Publish node configuration
export interface MQTTOutConfig extends MQTTConfig {
  retain?: boolean
}

// Switch Rule Types
export type SwitchOperator =
  | '=='
  | '!='
  | '<'
  | '<='
  | '>'
  | '>='
  | 'is-true'
  | 'is-false'
  | 'is-null'
  | 'is-defined'
  | 'contains'
  | 'starts-with'
  | 'ends-with'
  | 'matches-regex'
  | 'is-empty'
  | 'has-length'
  | 'includes'

export interface SwitchRule {
  property: string
  operator: SwitchOperator
  value?: any
  output: number
}

// Change Transformation Types
export type TransformAction = 'set' | 'change' | 'delete' | 'move'
export type ValueType = 'str' | 'num' | 'bool' | 'json' | 'date' | 'expr' | 'env' | 'flow'

export interface Transformation {
  action: TransformAction
  property: string
  valueType?: ValueType
  value?: any
  from?: string
  to?: string
  regex?: boolean
}

// Validation Types
export interface ValidationRule {
  type: 'required' | 'min' | 'max' | 'pattern' | 'custom'
  message: string
  params?: any
  validator?: (value: any, formData: any) => boolean | Promise<boolean>
}

export interface ValidationError {
  field: string
  message: string
}

export interface ValidationResult {
  valid: boolean
  errors: ValidationError[]
}
