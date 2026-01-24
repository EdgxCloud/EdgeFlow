/**
 * MQTT Topic Builder
 *
 * Comprehensive MQTT configuration editor based on Node-RED and n8n patterns
 * Supports broker configuration, topic management, and advanced MQTT settings
 */

import { useState } from 'react'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'
import { Switch } from '@/components/ui/switch'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from '@/components/ui/accordion'
import { Plus, Trash2, ChevronRight, Server, Shield, Settings2, Wifi, Eye, EyeOff } from 'lucide-react'
import { cn } from '@/lib/utils'

// MQTT Output format options (Node-RED style)
type OutputFormat = 'auto' | 'string' | 'buffer' | 'json'

// MQTT Protocol versions
type MQTTVersion = '3.1.1' | '5.0'

interface MQTTBrokerConfig {
  server?: string
  port?: number
  clientId?: string
  username?: string
  password?: string
  useTLS?: boolean
  protocolVersion?: MQTTVersion
  keepAlive?: number
  cleanSession?: boolean
  reconnectPeriod?: number
}

interface MQTTTopicBuilderProps {
  value: {
    // Broker settings
    broker?: MQTTBrokerConfig
    // Topic settings
    topic?: string
    qos?: 0 | 1 | 2
    retain?: boolean
    wildcardMode?: boolean
    // Output settings (for subscribe mode)
    outputFormat?: OutputFormat
    // Dynamic subscription support
    dynamicSubscription?: boolean
  }
  onChange: (value: any) => void
  mode?: 'publish' | 'subscribe'
  disabled?: boolean
}

interface TopicLevel {
  id: string
  value: string
  isWildcard: boolean
  wildcardType?: '+' | '#'
}

const COMMON_TOPICS = [
  { category: 'Home Automation', topics: ['home/+/temperature', 'home/+/humidity', 'home/#'] },
  { category: 'IoT Devices', topics: ['devices/+/status', 'devices/+/telemetry', 'devices/#'] },
  { category: 'Sensors', topics: ['sensors/+/data', 'sensors/+/config', 'sensors/#'] },
  { category: 'Control', topics: ['control/+/command', 'control/+/response', 'control/#'] },
]

// Common MQTT broker presets
const BROKER_PRESETS = [
  { name: 'Local Mosquitto', server: 'localhost', port: 1883, useTLS: false },
  { name: 'EMQX Cloud', server: 'broker.emqx.io', port: 1883, useTLS: false },
  { name: 'HiveMQ Cloud', server: 'broker.hivemq.com', port: 1883, useTLS: false },
  { name: 'CloudMQTT', server: 'm2m.eclipse.org', port: 1883, useTLS: false },
]

export function MQTTTopicBuilder({
  value,
  onChange,
  mode = 'subscribe',
  disabled = false,
}: MQTTTopicBuilderProps) {
  // Handle undefined or null value
  const safeValue = value || {}

  // Broker configuration
  const broker = safeValue.broker || {}
  const server = broker.server || ''
  const port = broker.port ?? 1883
  const clientId = broker.clientId || ''
  const username = broker.username || ''
  const password = broker.password || ''
  const useTLS = broker.useTLS ?? false
  const protocolVersion = broker.protocolVersion || '3.1.1'
  const keepAlive = broker.keepAlive ?? 60
  const cleanSession = broker.cleanSession ?? true
  const reconnectPeriod = broker.reconnectPeriod ?? 5000

  // Topic configuration
  const topic = safeValue.topic || ''
  const qos = safeValue.qos ?? 0
  const retain = safeValue.retain ?? false
  const wildcardMode = safeValue.wildcardMode ?? (mode === 'subscribe')

  // Subscribe-specific settings
  const outputFormat = safeValue.outputFormat || 'auto'
  const dynamicSubscription = safeValue.dynamicSubscription ?? false

  // UI state
  const [showPassword, setShowPassword] = useState(false)
  const [activeAccordion, setActiveAccordion] = useState<string[]>(['topic'])

  // Update broker config
  const updateBroker = (field: keyof MQTTBrokerConfig, fieldValue: any) => {
    onChange({
      ...safeValue,
      broker: {
        ...broker,
        [field]: fieldValue,
      },
    })
  }

  // Apply broker preset
  const applyBrokerPreset = (preset: typeof BROKER_PRESETS[0]) => {
    onChange({
      ...safeValue,
      broker: {
        ...broker,
        server: preset.server,
        port: preset.port,
        useTLS: preset.useTLS,
      },
    })
  }

  // Parse topic into levels
  const parseTopic = (topicStr: string): TopicLevel[] => {
    if (!topicStr) return []
    return topicStr.split('/').map((level, index) => ({
      id: `${index}-${level}`,
      value: level === '+' || level === '#' ? '' : level,
      isWildcard: level === '+' || level === '#',
      wildcardType: level === '+' || level === '#' ? (level as '+' | '#') : undefined,
    }))
  }

  const [levels, setLevels] = useState<TopicLevel[]>(() => parseTopic(topic))

  const buildTopic = (levelsList: TopicLevel[]): string => {
    return levelsList
      .map((level) => {
        if (level.isWildcard) {
          return level.wildcardType || '+'
        }
        return level.value || '+'
      })
      .join('/')
  }

  const updateTopic = (newLevels: TopicLevel[]) => {
    setLevels(newLevels)
    const newTopic = buildTopic(newLevels)
    onChange({ ...safeValue, topic: newTopic })
  }

  const addLevel = () => {
    const newLevels = [
      ...levels,
      {
        id: `${Date.now()}`,
        value: '',
        isWildcard: false,
      },
    ]
    updateTopic(newLevels)
  }

  const removeLevel = (index: number) => {
    const newLevels = levels.filter((_, i) => i !== index)
    updateTopic(newLevels)
  }

  const updateLevel = (index: number, updates: Partial<TopicLevel>) => {
    const newLevels = levels.map((level, i) =>
      i === index ? { ...level, ...updates } : level
    )
    updateTopic(newLevels)
  }

  const setPresetTopic = (topicStr: string) => {
    const newLevels = parseTopic(topicStr)
    setLevels(newLevels)
    onChange({ ...safeValue, topic: topicStr })
  }

  const handleDirectInput = (topicStr: string) => {
    const newLevels = parseTopic(topicStr)
    setLevels(newLevels)
    onChange({ ...safeValue, topic: topicStr })
  }

  return (
    <div className="space-y-4">
      <Accordion
        type="multiple"
        value={activeAccordion}
        onValueChange={setActiveAccordion}
        className="space-y-2"
      >
        {/* Broker Configuration Section */}
        <AccordionItem value="broker" className="border rounded-lg px-4">
          <AccordionTrigger className="hover:no-underline">
            <div className="flex items-center gap-2">
              <Server className="w-4 h-4 text-blue-500" />
              <span className="font-semibold">Broker Configuration</span>
              {server && (
                <span className="text-xs text-muted-foreground ml-2">
                  ({server}:{port})
                </span>
              )}
            </div>
          </AccordionTrigger>
          <AccordionContent className="space-y-4 pt-4">
            {/* Broker Presets */}
            <div className="space-y-2">
              <Label className="text-xs text-muted-foreground">Quick Connect</Label>
              <div className="flex flex-wrap gap-2">
                {BROKER_PRESETS.map((preset) => (
                  <Button
                    key={preset.name}
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={() => applyBrokerPreset(preset)}
                    className="h-7 text-xs"
                    disabled={disabled}
                  >
                    {preset.name}
                  </Button>
                ))}
              </div>
            </div>

            {/* Server and Port */}
            <div className="grid grid-cols-3 gap-3">
              <div className="col-span-2 space-y-2">
                <Label htmlFor="server" className="text-xs">Server / Host</Label>
                <Input
                  id="server"
                  value={server}
                  onChange={(e) => updateBroker('server', e.target.value)}
                  placeholder="mqtt.example.com"
                  className="h-9"
                  disabled={disabled}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="port" className="text-xs">Port</Label>
                <Input
                  id="port"
                  type="number"
                  value={port}
                  onChange={(e) => updateBroker('port', Number(e.target.value))}
                  placeholder="1883"
                  className="h-9"
                  disabled={disabled}
                />
              </div>
            </div>

            {/* Client ID */}
            <div className="space-y-2">
              <Label htmlFor="clientId" className="text-xs">Client ID (optional)</Label>
              <Input
                id="clientId"
                value={clientId}
                onChange={(e) => updateBroker('clientId', e.target.value)}
                placeholder="Leave empty for auto-generated ID"
                className="h-9"
                disabled={disabled}
              />
              <p className="text-xs text-muted-foreground">
                Unique identifier for this client. Auto-generated if left blank.
              </p>
            </div>

            {/* Authentication */}
            <div className="space-y-3">
              <Label className="text-xs font-semibold flex items-center gap-2">
                <Shield className="w-3 h-3" />
                Authentication
              </Label>
              <div className="grid grid-cols-2 gap-3">
                <div className="space-y-2">
                  <Label htmlFor="username" className="text-xs">Username</Label>
                  <Input
                    id="username"
                    value={username}
                    onChange={(e) => updateBroker('username', e.target.value)}
                    placeholder="Username"
                    className="h-9"
                    disabled={disabled}
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="password" className="text-xs">Password</Label>
                  <div className="relative">
                    <Input
                      id="password"
                      type={showPassword ? 'text' : 'password'}
                      value={password}
                      onChange={(e) => updateBroker('password', e.target.value)}
                      placeholder="Password"
                      className="h-9 pr-9"
                      disabled={disabled}
                    />
                    <button
                      type="button"
                      onClick={() => setShowPassword(!showPassword)}
                      className="absolute right-2 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                      disabled={disabled}
                    >
                      {showPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                    </button>
                  </div>
                </div>
              </div>
            </div>

            {/* TLS/SSL */}
            <div className="flex items-center justify-between py-2 px-3 bg-muted/50 rounded-lg">
              <div className="flex items-center gap-2">
                <Shield className="w-4 h-4 text-green-500" />
                <div>
                  <Label htmlFor="useTLS" className="text-sm font-medium cursor-pointer">
                    Enable TLS/SSL
                  </Label>
                  <p className="text-xs text-muted-foreground">
                    Encrypt connection (port 8883 recommended)
                  </p>
                </div>
              </div>
              <Switch
                id="useTLS"
                checked={useTLS}
                onCheckedChange={(checked) => {
                  updateBroker('useTLS', checked)
                  // Auto-adjust port when TLS is toggled
                  if (checked && port === 1883) {
                    updateBroker('port', 8883)
                  } else if (!checked && port === 8883) {
                    updateBroker('port', 1883)
                  }
                }}
                disabled={disabled}
              />
            </div>

            {/* Protocol Version */}
            <div className="space-y-2">
              <Label htmlFor="protocolVersion" className="text-xs">MQTT Protocol Version</Label>
              <Select
                value={protocolVersion}
                onValueChange={(value) => updateBroker('protocolVersion', value as MQTTVersion)}
                disabled={disabled}
              >
                <SelectTrigger className="h-9" id="protocolVersion">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="3.1.1">MQTT 3.1.1 (Default)</SelectItem>
                  <SelectItem value="5.0">MQTT 5.0</SelectItem>
                </SelectContent>
              </Select>
            </div>

            {/* Advanced Connection Settings */}
            <div className="grid grid-cols-2 gap-3">
              <div className="space-y-2">
                <Label htmlFor="keepAlive" className="text-xs">Keep Alive (seconds)</Label>
                <Input
                  id="keepAlive"
                  type="number"
                  value={keepAlive}
                  onChange={(e) => updateBroker('keepAlive', Number(e.target.value))}
                  min={0}
                  max={65535}
                  className="h-9"
                  disabled={disabled}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="reconnectPeriod" className="text-xs">Reconnect (ms)</Label>
                <Input
                  id="reconnectPeriod"
                  type="number"
                  value={reconnectPeriod}
                  onChange={(e) => updateBroker('reconnectPeriod', Number(e.target.value))}
                  min={0}
                  step={1000}
                  className="h-9"
                  disabled={disabled}
                />
              </div>
            </div>

            {/* Clean Session */}
            <div className="flex items-center justify-between py-2 px-3 bg-muted/50 rounded-lg">
              <div>
                <Label htmlFor="cleanSession" className="text-sm font-medium cursor-pointer">
                  Clean Session
                </Label>
                <p className="text-xs text-muted-foreground">
                  Start fresh session on each connect (no stored subscriptions)
                </p>
              </div>
              <Switch
                id="cleanSession"
                checked={cleanSession}
                onCheckedChange={(checked) => updateBroker('cleanSession', checked)}
                disabled={disabled}
              />
            </div>
          </AccordionContent>
        </AccordionItem>

        {/* Topic Configuration Section */}
        <AccordionItem value="topic" className="border rounded-lg px-4">
          <AccordionTrigger className="hover:no-underline">
            <div className="flex items-center gap-2">
              <Wifi className="w-4 h-4 text-green-500" />
              <span className="font-semibold">Topic Configuration</span>
              {topic && (
                <code className="text-xs bg-muted px-2 py-0.5 rounded ml-2">
                  {topic}
                </code>
              )}
            </div>
          </AccordionTrigger>
          <AccordionContent className="space-y-4 pt-4">
            {/* Direct Topic Input */}
            <div className="space-y-2">
              <Label htmlFor="topicDirect" className="text-sm font-semibold">
                MQTT Topic
              </Label>
              <Input
                id="topicDirect"
                value={topic}
                onChange={(e) => handleDirectInput(e.target.value)}
                placeholder="e.g., home/livingroom/temperature"
                className="h-11 font-mono"
                disabled={disabled}
              />
              <p className="text-xs text-muted-foreground">
                {mode === 'subscribe'
                  ? 'Use + for single-level wildcards, # for multi-level wildcards'
                  : 'Topic to publish to (no wildcards)'}
              </p>
            </div>

      {/* Topic Builder */}
      <div className="space-y-3">
        <div className="flex items-center justify-between">
          <Label className="text-sm font-semibold">Topic Builder</Label>
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={addLevel}
            className="h-8"
            disabled={disabled}
          >
            <Plus className="w-4 h-4 mr-1" />
            Add Level
          </Button>
        </div>

        {/* Topic Levels */}
        {levels.length === 0 ? (
          <div className="text-center py-8 border-2 border-dashed rounded-lg">
            <p className="text-sm text-muted-foreground">
              Add levels to build your MQTT topic hierarchy
            </p>
          </div>
        ) : (
          <div className="space-y-2">
            {levels.map((level, index) => (
              <div key={level.id} className="flex items-center gap-2">
                {/* Level indicator */}
                {index > 0 && (
                  <ChevronRight className="w-4 h-4 text-muted-foreground flex-shrink-0" />
                )}

                {/* Level Input */}
                <div className="flex-1 flex items-center gap-2">
                  <Input
                    value={level.value}
                    onChange={(e) => updateLevel(index, { value: e.target.value })}
                    placeholder={`Level ${index + 1}`}
                    className={cn(
                      'h-9 font-mono',
                      level.isWildcard && 'bg-muted'
                    )}
                    disabled={disabled || level.isWildcard}
                  />

                  {/* Wildcard toggle (only for subscribe mode) */}
                  {mode === 'subscribe' && wildcardMode && (
                    <Select
                      value={
                        level.isWildcard
                          ? level.wildcardType || '+'
                          : 'none'
                      }
                      onValueChange={(value) => {
                        if (value === 'none') {
                          updateLevel(index, {
                            isWildcard: false,
                            wildcardType: undefined,
                          })
                        } else {
                          updateLevel(index, {
                            isWildcard: true,
                            wildcardType: value as '+' | '#',
                          })
                        }
                      }}
                      disabled={disabled}
                    >
                      <SelectTrigger className="w-24 h-9">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="none">Text</SelectItem>
                        <SelectItem value="+">+ (Single)</SelectItem>
                        <SelectItem value="#"># (Multi)</SelectItem>
                      </SelectContent>
                    </Select>
                  )}
                </div>

                {/* Remove button */}
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  onClick={() => removeLevel(index)}
                  className="h-9 w-9 p-0 flex-shrink-0"
                  disabled={disabled}
                >
                  <Trash2 className="w-4 h-4" />
                </Button>
              </div>
            ))}
          </div>
        )}

        {/* Preview */}
        <div className="p-3 bg-muted rounded-lg border">
          <Label className="text-xs text-muted-foreground">Preview:</Label>
          <p className="font-mono text-sm font-semibold mt-1">
            {topic || <span className="text-muted-foreground italic">empty topic</span>}
          </p>
          {topic && mode === 'subscribe' && (
            <div className="mt-2 text-xs text-muted-foreground space-y-1">
              {topic.includes('+') && (
                <p>
                  <code className="bg-background px-1 rounded">+</code> matches one level (e.g.
                  home/<strong>bedroom</strong>/temp)
                </p>
              )}
              {topic.includes('#') && (
                <p>
                  <code className="bg-background px-1 rounded">#</code> matches all remaining
                  levels
                </p>
              )}
            </div>
          )}
        </div>
      </div>

      {/* Common Topics */}
      {mode === 'subscribe' && (
        <div className="space-y-3">
          <Label className="text-sm font-semibold">Common Topic Patterns</Label>
          <div className="space-y-2">
            {COMMON_TOPICS.map((category) => (
              <div key={category.category} className="space-y-1">
                <p className="text-xs font-semibold text-muted-foreground">
                  {category.category}
                </p>
                <div className="flex flex-wrap gap-2">
                  {category.topics.map((t) => (
                    <Button
                      key={t}
                      type="button"
                      variant="outline"
                      size="sm"
                      onClick={() => setPresetTopic(t)}
                      className="h-7 text-xs font-mono"
                      disabled={disabled}
                    >
                      {t}
                    </Button>
                  ))}
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

            {/* QoS */}
            <div className="space-y-2">
              <Label htmlFor="qos" className="text-xs">
                Quality of Service (QoS)
              </Label>
              <Select
                value={qos.toString()}
                onValueChange={(value) => onChange({ ...safeValue, qos: Number(value) as 0 | 1 | 2 })}
                disabled={disabled}
              >
                <SelectTrigger className="h-11" id="qos">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="0">
                    <div>
                      <div className="font-medium">QoS 0 - At most once</div>
                      <div className="text-xs text-muted-foreground">Fire and forget</div>
                    </div>
                  </SelectItem>
                  <SelectItem value="1">
                    <div>
                      <div className="font-medium">QoS 1 - At least once</div>
                      <div className="text-xs text-muted-foreground">Acknowledged delivery</div>
                    </div>
                  </SelectItem>
                  <SelectItem value="2">
                    <div>
                      <div className="font-medium">QoS 2 - Exactly once</div>
                      <div className="text-xs text-muted-foreground">Assured delivery</div>
                    </div>
                  </SelectItem>
                </SelectContent>
              </Select>
            </div>

            {/* Retain (only for publish) */}
            {mode === 'publish' && (
              <div className="flex items-center justify-between py-2 px-3 bg-muted/50 rounded-lg">
                <div>
                  <Label htmlFor="retain" className="text-sm font-medium cursor-pointer">
                    Retain Message
                  </Label>
                  <p className="text-xs text-muted-foreground">
                    Broker stores last message for new subscribers
                  </p>
                </div>
                <Switch
                  id="retain"
                  checked={retain}
                  onCheckedChange={(checked) => onChange({ ...safeValue, retain: checked })}
                  disabled={disabled}
                />
              </div>
            )}

            {/* Output Format (only for subscribe) */}
            {mode === 'subscribe' && (
              <div className="space-y-2">
                <Label htmlFor="outputFormat" className="text-xs">
                  Output Format
                </Label>
                <Select
                  value={outputFormat}
                  onValueChange={(value) => onChange({ ...safeValue, outputFormat: value as OutputFormat })}
                  disabled={disabled}
                >
                  <SelectTrigger className="h-11" id="outputFormat">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="auto">
                      <div>
                        <div className="font-medium">Auto-detect</div>
                        <div className="text-xs text-muted-foreground">String or Buffer based on content</div>
                      </div>
                    </SelectItem>
                    <SelectItem value="string">
                      <div>
                        <div className="font-medium">String</div>
                        <div className="text-xs text-muted-foreground">Always output as string</div>
                      </div>
                    </SelectItem>
                    <SelectItem value="buffer">
                      <div>
                        <div className="font-medium">Buffer</div>
                        <div className="text-xs text-muted-foreground">Raw binary buffer</div>
                      </div>
                    </SelectItem>
                    <SelectItem value="json">
                      <div>
                        <div className="font-medium">JSON Object</div>
                        <div className="text-xs text-muted-foreground">Automatically parse JSON</div>
                      </div>
                    </SelectItem>
                  </SelectContent>
                </Select>
              </div>
            )}

            {/* Dynamic Subscription (only for subscribe) */}
            {mode === 'subscribe' && (
              <div className="flex items-center justify-between py-2 px-3 bg-muted/50 rounded-lg">
                <div>
                  <Label htmlFor="dynamicSubscription" className="text-sm font-medium cursor-pointer">
                    Dynamic Subscription
                  </Label>
                  <p className="text-xs text-muted-foreground">
                    Allow runtime topic subscription changes via input messages
                  </p>
                </div>
                <Switch
                  id="dynamicSubscription"
                  checked={dynamicSubscription}
                  onCheckedChange={(checked) => onChange({ ...safeValue, dynamicSubscription: checked })}
                  disabled={disabled}
                />
              </div>
            )}
          </AccordionContent>
        </AccordionItem>

        {/* Advanced Settings Section */}
        <AccordionItem value="advanced" className="border rounded-lg px-4">
          <AccordionTrigger className="hover:no-underline">
            <div className="flex items-center gap-2">
              <Settings2 className="w-4 h-4 text-orange-500" />
              <span className="font-semibold">Advanced Settings</span>
            </div>
          </AccordionTrigger>
          <AccordionContent className="space-y-4 pt-4">
            {/* Wildcard Help */}
            {mode === 'subscribe' && (
              <div className="p-4 bg-blue-50 dark:bg-blue-950/20 border border-blue-200 dark:border-blue-900 rounded-lg">
                <p className="text-sm font-semibold text-blue-900 dark:text-blue-100 mb-2">
                  MQTT Wildcard Guide
                </p>
                <div className="space-y-2 text-xs text-blue-700 dark:text-blue-300">
                  <div>
                    <p className="font-semibold">Single-level wildcard (+)</p>
                    <p>
                      <code className="bg-blue-100 dark:bg-blue-900 px-1 rounded">home/+/temperature</code>{' '}
                      matches home/<strong>bedroom</strong>/temperature, home/<strong>kitchen</strong>
                      /temperature
                    </p>
                  </div>
                  <div>
                    <p className="font-semibold">Multi-level wildcard (#)</p>
                    <p>
                      <code className="bg-blue-100 dark:bg-blue-900 px-1 rounded">home/#</code> matches
                      home/bedroom/temperature, home/kitchen/humidity/sensor1, etc.
                    </p>
                  </div>
                  <div>
                    <p className="font-semibold">Rules:</p>
                    <ul className="list-disc list-inside ml-2">
                      <li># must be the last character in the topic</li>
                      <li>+ and # cannot be mixed with text in the same level</li>
                    </ul>
                  </div>
                </div>
              </div>
            )}

            {/* QoS Explanation */}
            <div className="p-4 bg-amber-50 dark:bg-amber-950/20 border border-amber-200 dark:border-amber-900 rounded-lg">
              <p className="text-sm font-semibold text-amber-900 dark:text-amber-100 mb-2">
                QoS Levels Explained
              </p>
              <div className="space-y-2 text-xs text-amber-700 dark:text-amber-300">
                <div>
                  <p className="font-semibold">QoS 0 - At most once</p>
                  <p>Best effort delivery, no acknowledgment. Use for high-frequency sensor data.</p>
                </div>
                <div>
                  <p className="font-semibold">QoS 1 - At least once</p>
                  <p>Guaranteed delivery with possible duplicates. Good for status updates.</p>
                </div>
                <div>
                  <p className="font-semibold">QoS 2 - Exactly once</p>
                  <p>Four-step handshake ensures exactly one delivery. Use for critical messages.</p>
                </div>
              </div>
            </div>

            {/* Connection Info */}
            <div className="p-4 bg-green-50 dark:bg-green-950/20 border border-green-200 dark:border-green-900 rounded-lg">
              <p className="text-sm font-semibold text-green-900 dark:text-green-100 mb-2">
                Connection Best Practices
              </p>
              <ul className="list-disc list-inside text-xs text-green-700 dark:text-green-300 space-y-1">
                <li>Always use TLS (port 8883) for production environments</li>
                <li>Use unique Client IDs to avoid connection conflicts</li>
                <li>Set Keep Alive to match your network timeout settings</li>
                <li>Enable Clean Session for stateless applications</li>
                <li>Disable Clean Session if you need offline message queuing</li>
              </ul>
            </div>
          </AccordionContent>
        </AccordionItem>
      </Accordion>
    </div>
  )
}
