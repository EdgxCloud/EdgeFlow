/**
 * Node Config Factory
 *
 * Maps node types to specialized configuration editors
 */

import { PropertySchema } from '@/types/node'

// Dashboard Editors
import {
  ChartEditor,
  GaugeEditor,
  FormBuilderEditor,
  TableEditor,
  TextEditor,
  ButtonEditor,
  SliderEditor,
  SwitchEditor,
  TextInputEditor,
  DropdownEditor,
  NotificationEditor,
  TemplateEditor,
} from './Dashboard'

// Specialized Editors
import {
  GPIOPinSelector,
  MQTTTopicBuilder,
  SwitchRuleBuilder,
  HTTPRequestBuilder,
  HTTPWebhookBuilder,
  ChangeTransformBuilder,
  FunctionTransformBuilder,
} from './Specialized'

interface EditorProps {
  value: any
  config?: any
  onChange: (value: any) => void
  disabled?: boolean
  mode?: string
}

export type NodeEditorComponent = React.ComponentType<EditorProps>

/**
 * Node type to editor component mapping
 */
const NODE_EDITORS: Record<string, NodeEditorComponent> = {
  // Dashboard Widgets
  'dashboard-chart': ChartEditor as any,
  'dashboard-gauge': GaugeEditor as any,
  'dashboard-form': FormBuilderEditor as any,
  'dashboard-table': TableEditor as any,
  'dashboard-text': TextEditor as any,
  'dashboard-button': ButtonEditor as any,
  'dashboard-slider': SliderEditor as any,
  'dashboard-switch': SwitchEditor as any,
  'dashboard-text-input': TextInputEditor as any,
  'dashboard-dropdown': DropdownEditor as any,
  'dashboard-notification': NotificationEditor as any,
  'dashboard-template': TemplateEditor as any,

  // GPIO Nodes
  'gpio-in': GPIOPinSelector as any,
  'gpio-out': GPIOPinSelector as any,

  // MQTT Nodes
  'mqtt-in': MQTTTopicBuilder as any,
  'mqtt-out': MQTTTopicBuilder as any,

  // HTTP Nodes
  'http-request': HTTPRequestBuilder as any,
  'http-webhook': HTTPWebhookBuilder as any,
  'http-in': HTTPWebhookBuilder as any,
  'webhook': HTTPWebhookBuilder as any,

  // Logic Nodes
  switch: SwitchRuleBuilder as any,
  change: ChangeTransformBuilder as any,
  function: FunctionTransformBuilder as any,
}

/**
 * Property type to editor component mapping
 */
const PROPERTY_EDITORS: Record<string, NodeEditorComponent> = {
  'gpio-pin': GPIOPinSelector as any,
  'mqtt-topic': MQTTTopicBuilder as any,
  'switch-rules': SwitchRuleBuilder as any,
  'http-request': HTTPRequestBuilder as any,
  'change-rules': ChangeTransformBuilder as any,
  'function-rules': FunctionTransformBuilder as any,
}

/**
 * Get specialized editor for a node type
 */
export function getNodeEditor(nodeType: string): NodeEditorComponent | null {
  return NODE_EDITORS[nodeType] || null
}

/**
 * Get specialized editor for a property type
 */
export function getPropertyEditor(propertyType: string): NodeEditorComponent | null {
  return PROPERTY_EDITORS[propertyType] || null
}

/**
 * Check if a node type has a specialized editor
 */
export function hasNodeEditor(nodeType: string): boolean {
  return nodeType in NODE_EDITORS
}

/**
 * Check if a property type has a specialized editor
 */
export function hasPropertyEditor(propertyType: string): boolean {
  return propertyType in PROPERTY_EDITORS
}

/**
 * Get editor mode for MQTT nodes
 */
export function getMQTTMode(nodeType: string): 'publish' | 'subscribe' {
  if (nodeType === 'mqtt-in') return 'subscribe'
  if (nodeType === 'mqtt-out') return 'publish'
  return 'subscribe'
}

/**
 * Check if entire node config should use specialized editor
 * instead of generic PropertyField rendering
 */
export function shouldUseNodeEditor(nodeType: string): boolean {
  // Dashboard widgets and some nodes render their entire config
  return (
    nodeType.startsWith('dashboard-') ||
    nodeType.startsWith('mqtt-') ||
    nodeType.startsWith('gpio-') ||
    nodeType.startsWith('http-') ||
    ['switch', 'change', 'webhook', 'function'].includes(nodeType)
  )
}

/**
 * Check if a specific property should use a specialized editor
 */
export function shouldUsePropertyEditor(schema: PropertySchema): boolean {
  // Check if property type has a specialized editor
  return schema.type in PROPERTY_EDITORS
}

/**
 * Render specialized editor for a node type
 */
export function renderNodeEditor(
  nodeType: string,
  config: any,
  onChange: (config: any) => void,
  disabled?: boolean
): React.ReactNode {
  const Editor = getNodeEditor(nodeType)
  if (!Editor) return null

  // Special handling for MQTT nodes
  if (nodeType.startsWith('mqtt-')) {
    return <Editor value={config} onChange={onChange} disabled={disabled} mode={getMQTTMode(nodeType)} />
  }

  // Dashboard editors use 'config' prop, other editors use 'value'
  return <Editor value={config} config={config} onChange={onChange} disabled={disabled} />
}

/**
 * Render specialized editor for a property
 */
export function renderPropertyEditor(
  schema: PropertySchema,
  value: any,
  onChange: (value: any) => void,
  disabled?: boolean
): React.ReactNode {
  const Editor = getPropertyEditor(schema.type)
  if (!Editor) return null

  return <Editor value={value} onChange={onChange} disabled={disabled} />
}
