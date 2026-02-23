/**
 * Simple Widget Editors
 *
 * Editors for simple dashboard widgets: Text, Button, Slider, Switch, TextInput, Dropdown, DatePicker, Notification, Template
 */

import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { ColorPicker } from '@/components/Common/ColorPicker'
import { IconPicker } from '@/components/Common/IconPicker'

// ==================== TEXT WIDGET EDITOR ====================

interface TextEditorProps {
  config: {
    text?: string
    fontSize?: number
    fontWeight?: 'normal' | 'bold'
    textAlign?: 'left' | 'center' | 'right'
    color?: string
    backgroundColor?: string
    format?: 'plain' | 'markdown' | 'html'
  }
  onChange: (config: any) => void
}

export function TextEditor({ config: rawConfig, value, onChange }: TextEditorProps & { value?: any }) {
  const config = rawConfig || value || {}
  const text = config.text || ''
  const fontSize = config.fontSize || 14
  const fontWeight = config.fontWeight || 'normal'
  const textAlign = config.textAlign || 'left'
  const color = config.color || '#000000'
  const backgroundColor = config.backgroundColor || '#ffffff'
  const format = config.format || 'plain'

  return (
    <div className="space-y-4">
      <div className="space-y-2">
        <Label htmlFor="text" className="text-sm font-semibold">
          Text Content
        </Label>
        <Textarea
          id="text"
          value={text}
          onChange={(e) => onChange({ ...config, text: e.target.value })}
          placeholder="Enter text or {{msg.payload}} for dynamic values"
          rows={4}
        />
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label className="text-sm font-semibold">Text Color</Label>
          <ColorPicker value={color} onChange={(c) => onChange({ ...config, color: c })} />
        </div>
        <div className="space-y-2">
          <Label className="text-sm font-semibold">Background Color</Label>
          <ColorPicker
            value={backgroundColor}
            onChange={(c) => onChange({ ...config, backgroundColor: c })}
          />
        </div>
      </div>

      <div className="grid grid-cols-3 gap-4">
        <div className="space-y-2">
          <Label htmlFor="fontSize" className="text-sm font-semibold">
            Font Size (px)
          </Label>
          <Input
            id="fontSize"
            type="number"
            value={fontSize}
            onChange={(e) => onChange({ ...config, fontSize: Number(e.target.value) })}
            min={8}
            max={72}
            className="h-11"
          />
        </div>

        <div className="space-y-2">
          <Label className="text-sm font-semibold">Font Weight</Label>
          <Select
            value={fontWeight}
            onValueChange={(value: 'normal' | 'bold') =>
              onChange({ ...config, fontWeight: value })
            }
          >
            <SelectTrigger className="h-11">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="normal">Normal</SelectItem>
              <SelectItem value="bold">Bold</SelectItem>
            </SelectContent>
          </Select>
        </div>

        <div className="space-y-2">
          <Label className="text-sm font-semibold">Text Align</Label>
          <Select
            value={textAlign}
            onValueChange={(value: 'left' | 'center' | 'right') =>
              onChange({ ...config, textAlign: value })
            }
          >
            <SelectTrigger className="h-11">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="left">Left</SelectItem>
              <SelectItem value="center">Center</SelectItem>
              <SelectItem value="right">Right</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>

      <div className="space-y-2">
        <Label className="text-sm font-semibold">Format</Label>
        <Select
          value={format}
          onValueChange={(value: 'plain' | 'markdown' | 'html') =>
            onChange({ ...config, format: value })
          }
        >
          <SelectTrigger className="h-11">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="plain">Plain Text</SelectItem>
            <SelectItem value="markdown">Markdown</SelectItem>
            <SelectItem value="html">HTML</SelectItem>
          </SelectContent>
        </Select>
      </div>
    </div>
  )
}

// ==================== BUTTON WIDGET EDITOR ====================

interface ButtonEditorProps {
  config: {
    buttonLabel?: string
    icon?: string
    bgColor?: string
    fgColor?: string
    size?: 'sm' | 'md' | 'lg'
    fullWidth?: boolean
    disabled?: boolean
    payload?: any
    outputTopic?: string
  }
  onChange: (config: any) => void
}

export function ButtonEditor({ config: rawConfig, value, onChange }: ButtonEditorProps & { value?: any }) {
  const config = rawConfig || value || {}
  const buttonLabel = config.buttonLabel || 'Click Me'
  const icon = config.icon || ''
  const bgColor = config.bgColor || '#3b82f6'
  const fgColor = config.fgColor || '#ffffff'
  const size = config.size || 'md'
  const fullWidth = config.fullWidth ?? false
  const disabled = config.disabled ?? false
  const payload = config.payload ?? { clicked: true }
  const outputTopic = config.outputTopic || ''

  return (
    <div className="space-y-4">
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label htmlFor="buttonLabel" className="text-sm font-semibold">
            Button Label
          </Label>
          <Input
            id="buttonLabel"
            value={buttonLabel}
            onChange={(e) => onChange({ ...config, buttonLabel: e.target.value })}
            className="h-11"
          />
        </div>

        <div className="space-y-2">
          <Label className="text-sm font-semibold">Icon</Label>
          <IconPicker value={icon} onChange={(i) => onChange({ ...config, icon: i })} />
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label className="text-sm font-semibold">Background Color</Label>
          <ColorPicker value={bgColor} onChange={(c) => onChange({ ...config, bgColor: c })} />
        </div>
        <div className="space-y-2">
          <Label className="text-sm font-semibold">Text Color</Label>
          <ColorPicker value={fgColor} onChange={(c) => onChange({ ...config, fgColor: c })} />
        </div>
      </div>

      <div className="space-y-2">
        <Label className="text-sm font-semibold">Size</Label>
        <Select
          value={size}
          onValueChange={(value: 'sm' | 'md' | 'lg') => onChange({ ...config, size: value })}
        >
          <SelectTrigger className="h-11">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="sm">Small</SelectItem>
            <SelectItem value="md">Medium</SelectItem>
            <SelectItem value="lg">Large</SelectItem>
          </SelectContent>
        </Select>
      </div>

      <div className="space-y-2">
        <Label htmlFor="outputTopic" className="text-sm font-semibold">
          Output Topic
        </Label>
        <Input
          id="outputTopic"
          value={outputTopic}
          onChange={(e) => onChange({ ...config, outputTopic: e.target.value })}
          placeholder="e.g., button/clicked"
          className="h-11"
        />
      </div>

      <div className="flex gap-4">
        <div className="flex items-center space-x-2">
          <Switch
            id="fullWidth"
            checked={fullWidth}
            onCheckedChange={(checked) => onChange({ ...config, fullWidth: checked })}
          />
          <Label htmlFor="fullWidth" className="text-sm font-normal cursor-pointer">
            Full Width
          </Label>
        </div>
        <div className="flex items-center space-x-2">
          <Switch
            id="disabled"
            checked={disabled}
            onCheckedChange={(checked) => onChange({ ...config, disabled: checked })}
          />
          <Label htmlFor="disabled" className="text-sm font-normal cursor-pointer">
            Disabled
          </Label>
        </div>
      </div>
    </div>
  )
}

// ==================== SLIDER WIDGET EDITOR ====================

interface SliderEditorProps {
  config: {
    min?: number
    max?: number
    step?: number
    defaultValue?: number
    showValue?: boolean
    showLabels?: boolean
    color?: string
    outputTopic?: string
    units?: string
  }
  onChange: (config: any) => void
}

export function SliderEditor({ config: rawConfig, value, onChange }: SliderEditorProps & { value?: any }) {
  const config = rawConfig || value || {}
  const min = config.min ?? 0
  const max = config.max ?? 100
  const step = config.step ?? 1
  const defaultValue = config.defaultValue ?? 50
  const showValue = config.showValue ?? true
  const showLabels = config.showLabels ?? true
  const color = config.color || '#3b82f6'
  const outputTopic = config.outputTopic || ''
  const units = config.units || ''

  return (
    <div className="space-y-4">
      <div className="grid grid-cols-3 gap-4">
        <div className="space-y-2">
          <Label htmlFor="min" className="text-sm font-semibold">
            Minimum
          </Label>
          <Input
            id="min"
            type="number"
            value={min}
            onChange={(e) => onChange({ ...config, min: Number(e.target.value) })}
            className="h-11"
          />
        </div>
        <div className="space-y-2">
          <Label htmlFor="max" className="text-sm font-semibold">
            Maximum
          </Label>
          <Input
            id="max"
            type="number"
            value={max}
            onChange={(e) => onChange({ ...config, max: Number(e.target.value) })}
            className="h-11"
          />
        </div>
        <div className="space-y-2">
          <Label htmlFor="step" className="text-sm font-semibold">
            Step
          </Label>
          <Input
            id="step"
            type="number"
            value={step}
            onChange={(e) => onChange({ ...config, step: Number(e.target.value) })}
            className="h-11"
            min={0.01}
          />
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label htmlFor="defaultValue" className="text-sm font-semibold">
            Default Value
          </Label>
          <Input
            id="defaultValue"
            type="number"
            value={defaultValue}
            onChange={(e) => onChange({ ...config, defaultValue: Number(e.target.value) })}
            className="h-11"
            min={min}
            max={max}
          />
        </div>
        <div className="space-y-2">
          <Label htmlFor="units" className="text-sm font-semibold">
            Units
          </Label>
          <Input
            id="units"
            value={units}
            onChange={(e) => onChange({ ...config, units: e.target.value })}
            placeholder="e.g., %, °C"
            className="h-11"
          />
        </div>
      </div>

      <div className="space-y-2">
        <Label className="text-sm font-semibold">Slider Color</Label>
        <ColorPicker value={color} onChange={(c) => onChange({ ...config, color: c })} />
      </div>

      <div className="space-y-2">
        <Label htmlFor="outputTopic" className="text-sm font-semibold">
          Output Topic
        </Label>
        <Input
          id="outputTopic"
          value={outputTopic}
          onChange={(e) => onChange({ ...config, outputTopic: e.target.value })}
          placeholder="e.g., slider/value"
          className="h-11"
        />
      </div>

      <div className="flex gap-4">
        <div className="flex items-center space-x-2">
          <Switch
            id="showValue"
            checked={showValue}
            onCheckedChange={(checked) => onChange({ ...config, showValue: checked })}
          />
          <Label htmlFor="showValue" className="text-sm font-normal cursor-pointer">
            Show Current Value
          </Label>
        </div>
        <div className="flex items-center space-x-2">
          <Switch
            id="showLabels"
            checked={showLabels}
            onCheckedChange={(checked) => onChange({ ...config, showLabels: checked })}
          />
          <Label htmlFor="showLabels" className="text-sm font-normal cursor-pointer">
            Show Min/Max Labels
          </Label>
        </div>
      </div>
    </div>
  )
}

// ==================== SWITCH WIDGET EDITOR ====================

interface SwitchEditorProps {
  config: {
    onLabel?: string
    offLabel?: string
    defaultValue?: boolean
    outputTopic?: string
    onPayload?: any
    offPayload?: any
    color?: string
  }
  onChange: (config: any) => void
}

export function SwitchEditor({ config: rawConfig, value, onChange }: SwitchEditorProps & { value?: any }) {
  const config = rawConfig || value || {}
  const onLabel = config.onLabel || 'On'
  const offLabel = config.offLabel || 'Off'
  const defaultValue = config.defaultValue ?? false
  const outputTopic = config.outputTopic || ''
  const onPayload = config.onPayload ?? true
  const offPayload = config.offPayload ?? false
  const color = config.color || '#3b82f6'

  return (
    <div className="space-y-4">
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label htmlFor="onLabel" className="text-sm font-semibold">
            On Label
          </Label>
          <Input
            id="onLabel"
            value={onLabel}
            onChange={(e) => onChange({ ...config, onLabel: e.target.value })}
            className="h-11"
          />
        </div>
        <div className="space-y-2">
          <Label htmlFor="offLabel" className="text-sm font-semibold">
            Off Label
          </Label>
          <Input
            id="offLabel"
            value={offLabel}
            onChange={(e) => onChange({ ...config, offLabel: e.target.value })}
            className="h-11"
          />
        </div>
      </div>

      <div className="space-y-2">
        <Label className="text-sm font-semibold">Switch Color</Label>
        <ColorPicker value={color} onChange={(c) => onChange({ ...config, color: c })} />
      </div>

      <div className="space-y-2">
        <Label htmlFor="outputTopic" className="text-sm font-semibold">
          Output Topic
        </Label>
        <Input
          id="outputTopic"
          value={outputTopic}
          onChange={(e) => onChange({ ...config, outputTopic: e.target.value })}
          placeholder="e.g., switch/state"
          className="h-11"
        />
      </div>

      <div className="flex items-center space-x-2">
        <Switch
          id="defaultValue"
          checked={defaultValue}
          onCheckedChange={(checked) => onChange({ ...config, defaultValue: checked })}
        />
        <Label htmlFor="defaultValue" className="text-sm font-normal cursor-pointer">
          Default State: {defaultValue ? 'On' : 'Off'}
        </Label>
      </div>
    </div>
  )
}

// ==================== TEXT INPUT WIDGET EDITOR ====================

interface TextInputEditorProps {
  config: {
    placeholder?: string
    defaultValue?: string
    inputType?: 'text' | 'email' | 'password' | 'number' | 'tel' | 'url'
    maxLength?: number
    required?: boolean
    outputTopic?: string
    sendOnChange?: boolean
  }
  onChange: (config: any) => void
}

export function TextInputEditor({ config: rawConfig, value, onChange }: TextInputEditorProps & { value?: any }) {
  const config = rawConfig || value || {}
  const placeholder = config.placeholder || ''
  const defaultValue = config.defaultValue || ''
  const inputType = config.inputType || 'text'
  const maxLength = config.maxLength
  const required = config.required ?? false
  const outputTopic = config.outputTopic || ''
  const sendOnChange = config.sendOnChange ?? false

  return (
    <div className="space-y-4">
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label htmlFor="placeholder" className="text-sm font-semibold">
            Placeholder
          </Label>
          <Input
            id="placeholder"
            value={placeholder}
            onChange={(e) => onChange({ ...config, placeholder: e.target.value })}
            className="h-11"
          />
        </div>
        <div className="space-y-2">
          <Label htmlFor="defaultValue" className="text-sm font-semibold">
            Default Value
          </Label>
          <Input
            id="defaultValue"
            value={defaultValue}
            onChange={(e) => onChange({ ...config, defaultValue: e.target.value })}
            className="h-11"
          />
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label className="text-sm font-semibold">Input Type</Label>
          <Select
            value={inputType}
            onValueChange={(value: typeof inputType) => onChange({ ...config, inputType: value })}
          >
            <SelectTrigger className="h-11">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="text">Text</SelectItem>
              <SelectItem value="email">Email</SelectItem>
              <SelectItem value="password">Password</SelectItem>
              <SelectItem value="number">Number</SelectItem>
              <SelectItem value="tel">Telephone</SelectItem>
              <SelectItem value="url">URL</SelectItem>
            </SelectContent>
          </Select>
        </div>
        <div className="space-y-2">
          <Label htmlFor="maxLength" className="text-sm font-semibold">
            Max Length
          </Label>
          <Input
            id="maxLength"
            type="number"
            value={maxLength || ''}
            onChange={(e) =>
              onChange({
                ...config,
                maxLength: e.target.value ? Number(e.target.value) : undefined,
              })
            }
            placeholder="Unlimited"
            className="h-11"
          />
        </div>
      </div>

      <div className="space-y-2">
        <Label htmlFor="outputTopic" className="text-sm font-semibold">
          Output Topic
        </Label>
        <Input
          id="outputTopic"
          value={outputTopic}
          onChange={(e) => onChange({ ...config, outputTopic: e.target.value })}
          placeholder="e.g., input/value"
          className="h-11"
        />
      </div>

      <div className="flex gap-4">
        <div className="flex items-center space-x-2">
          <Switch
            id="required"
            checked={required}
            onCheckedChange={(checked) => onChange({ ...config, required: checked })}
          />
          <Label htmlFor="required" className="text-sm font-normal cursor-pointer">
            Required
          </Label>
        </div>
        <div className="flex items-center space-x-2">
          <Switch
            id="sendOnChange"
            checked={sendOnChange}
            onCheckedChange={(checked) => onChange({ ...config, sendOnChange: checked })}
          />
          <Label htmlFor="sendOnChange" className="text-sm font-normal cursor-pointer">
            Send on Change (live)
          </Label>
        </div>
      </div>
    </div>
  )
}

// ==================== DROPDOWN WIDGET EDITOR ====================

interface DropdownEditorProps {
  config: {
    options?: string[]
    defaultValue?: string
    placeholder?: string
    outputTopic?: string
    allowClear?: boolean
  }
  onChange: (config: any) => void
}

export function DropdownEditor({ config: rawConfig, value, onChange }: DropdownEditorProps & { value?: any }) {
  const config = rawConfig || value || {}
  const options = config.options || []
  const defaultValue = config.defaultValue || ''
  const placeholder = config.placeholder || 'Select an option'
  const outputTopic = config.outputTopic || ''
  const allowClear = config.allowClear ?? true

  return (
    <div className="space-y-4">
      <div className="space-y-2">
        <Label className="text-sm font-semibold">Options (one per line)</Label>
        <Textarea
          value={options.join('\n')}
          onChange={(e) =>
            onChange({
              ...config,
              options: e.target.value.split('\n').filter((o) => o.trim()),
            })
          }
          rows={6}
          placeholder="Option 1&#10;Option 2&#10;Option 3"
        />
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label htmlFor="placeholder" className="text-sm font-semibold">
            Placeholder
          </Label>
          <Input
            id="placeholder"
            value={placeholder}
            onChange={(e) => onChange({ ...config, placeholder: e.target.value })}
            className="h-11"
          />
        </div>
        <div className="space-y-2">
          <Label htmlFor="defaultValue" className="text-sm font-semibold">
            Default Value
          </Label>
          <Input
            id="defaultValue"
            value={defaultValue}
            onChange={(e) => onChange({ ...config, defaultValue: e.target.value })}
            className="h-11"
          />
        </div>
      </div>

      <div className="space-y-2">
        <Label htmlFor="outputTopic" className="text-sm font-semibold">
          Output Topic
        </Label>
        <Input
          id="outputTopic"
          value={outputTopic}
          onChange={(e) => onChange({ ...config, outputTopic: e.target.value })}
          placeholder="e.g., dropdown/selected"
          className="h-11"
        />
      </div>

      <div className="flex items-center space-x-2">
        <Switch
          id="allowClear"
          checked={allowClear}
          onCheckedChange={(checked) => onChange({ ...config, allowClear: checked })}
        />
        <Label htmlFor="allowClear" className="text-sm font-normal cursor-pointer">
          Allow Clear Selection
        </Label>
      </div>
    </div>
  )
}

// ==================== NOTIFICATION WIDGET EDITOR ====================

interface NotificationEditorProps {
  config: {
    position?:
      | 'top-left'
      | 'top-center'
      | 'top-right'
      | 'bottom-left'
      | 'bottom-center'
      | 'bottom-right'
    duration?: number
    showCloseButton?: boolean
    sound?: boolean
    maxNotifications?: number
  }
  onChange: (config: any) => void
}

export function NotificationEditor({ config: rawConfig, value, onChange }: NotificationEditorProps & { value?: any }) {
  const config = rawConfig || value || {}
  const position = config.position || 'top-right'
  const duration = config.duration ?? 5000
  const showCloseButton = config.showCloseButton ?? true
  const sound = config.sound ?? false
  const maxNotifications = config.maxNotifications || 3

  return (
    <div className="space-y-4">
      <div className="space-y-2">
        <Label className="text-sm font-semibold">Position</Label>
        <Select
          value={position}
          onValueChange={(value: typeof position) => onChange({ ...config, position: value })}
        >
          <SelectTrigger className="h-11">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="top-left">Top Left</SelectItem>
            <SelectItem value="top-center">Top Center</SelectItem>
            <SelectItem value="top-right">Top Right</SelectItem>
            <SelectItem value="bottom-left">Bottom Left</SelectItem>
            <SelectItem value="bottom-center">Bottom Center</SelectItem>
            <SelectItem value="bottom-right">Bottom Right</SelectItem>
          </SelectContent>
        </Select>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label htmlFor="duration" className="text-sm font-semibold">
            Duration (ms)
          </Label>
          <Input
            id="duration"
            type="number"
            value={duration}
            onChange={(e) => onChange({ ...config, duration: Number(e.target.value) })}
            className="h-11"
            min={1000}
            max={30000}
          />
        </div>
        <div className="space-y-2">
          <Label htmlFor="maxNotifications" className="text-sm font-semibold">
            Max Visible
          </Label>
          <Input
            id="maxNotifications"
            type="number"
            value={maxNotifications}
            onChange={(e) => onChange({ ...config, maxNotifications: Number(e.target.value) })}
            className="h-11"
            min={1}
            max={10}
          />
        </div>
      </div>

      <div className="flex gap-4">
        <div className="flex items-center space-x-2">
          <Switch
            id="showCloseButton"
            checked={showCloseButton}
            onCheckedChange={(checked) => onChange({ ...config, showCloseButton: checked })}
          />
          <Label htmlFor="showCloseButton" className="text-sm font-normal cursor-pointer">
            Show Close Button
          </Label>
        </div>
        <div className="flex items-center space-x-2">
          <Switch
            id="sound"
            checked={sound}
            onCheckedChange={(checked) => onChange({ ...config, sound: checked })}
          />
          <Label htmlFor="sound" className="text-sm font-normal cursor-pointer">
            Play Sound
          </Label>
        </div>
      </div>

      <div className="p-4 bg-muted/50 rounded-lg border">
        <p className="text-sm font-semibold mb-1">Message Format</p>
        <pre className="text-xs bg-background p-3 rounded border overflow-x-auto">
          {JSON.stringify(
            {
              type: 'info', // 'info' | 'success' | 'warning' | 'error'
              title: 'Notification Title',
              message: 'Notification message text',
            },
            null,
            2
          )}
        </pre>
      </div>
    </div>
  )
}

// ==================== TEMPLATE WIDGET EDITOR ====================

interface TemplateEditorProps {
  config: {
    template?: string
    templateEngine?: 'mustache' | 'handlebars'
    outputFormat?: 'text' | 'html'
  }
  onChange: (config: any) => void
}

export function TemplateEditor({ config: rawConfig, value, onChange }: TemplateEditorProps & { value?: any }) {
  const config = rawConfig || value || {}
  const template = config.template || ''
  const templateEngine = config.templateEngine || 'mustache'
  const outputFormat = config.outputFormat || 'text'

  return (
    <div className="space-y-4">
      <div className="space-y-2">
        <Label className="text-sm font-semibold">Template</Label>
        <Textarea
          value={template}
          onChange={(e) => onChange({ ...config, template: e.target.value })}
          rows={8}
          placeholder="Hello {{name}}!&#10;Your value is: {{msg.payload}}"
          className="font-mono text-sm"
        />
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label className="text-sm font-semibold">Template Engine</Label>
          <Select
            value={templateEngine}
            onValueChange={(value: 'mustache' | 'handlebars') =>
              onChange({ ...config, templateEngine: value })
            }
          >
            <SelectTrigger className="h-11">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="mustache">Mustache</SelectItem>
              <SelectItem value="handlebars">Handlebars</SelectItem>
            </SelectContent>
          </Select>
        </div>
        <div className="space-y-2">
          <Label className="text-sm font-semibold">Output Format</Label>
          <Select
            value={outputFormat}
            onValueChange={(value: 'text' | 'html') =>
              onChange({ ...config, outputFormat: value })
            }
          >
            <SelectTrigger className="h-11">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="text">Plain Text</SelectItem>
              <SelectItem value="html">HTML</SelectItem>
            </SelectContent>
          </Select>
        </div>
      </div>

      <div className="p-4 bg-muted/50 rounded-lg border">
        <p className="text-sm font-semibold mb-1">Template Variables</p>
        <p className="text-xs text-muted-foreground mb-2">
          Use <code className="bg-background px-1 rounded">{'{{variable}}'}</code> syntax to access
          message properties:
        </p>
        <ul className="text-xs text-muted-foreground space-y-1 ml-4 list-disc">
          <li>
            <code className="bg-background px-1 rounded">{'{{msg.payload}}'}</code> - Message payload
          </li>
          <li>
            <code className="bg-background px-1 rounded">{'{{msg.topic}}'}</code> - Message topic
          </li>
          <li>
            <code className="bg-background px-1 rounded">{'{{timestamp}}'}</code> - Current timestamp
          </li>
        </ul>
      </div>
    </div>
  )
}
