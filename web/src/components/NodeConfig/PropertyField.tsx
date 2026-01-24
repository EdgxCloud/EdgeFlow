/**
 * Property Field Component
 *
 * Generic property field renderer based on property schema
 */

import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import { CodeEditor } from '@/components/Common/CodeEditor'
import { JSONEditor } from '@/components/Common/JSONEditor'
import { ColorPicker } from '@/components/Common/ColorPicker'
import { IconPicker } from '@/components/Common/IconPicker'
import { CronBuilder } from '@/components/Common/CronBuilder'
import { MQTTTopicBuilder } from '@/components/NodeConfig/Specialized/MQTTTopicBuilder'
import type { PropertySchema } from '@/types/node'
import { cn } from '@/lib/utils'

interface PropertyFieldProps {
  schema: PropertySchema
  value: any
  onChange: (value: any) => void
  error?: string
  disabled?: boolean
}

export function PropertyField({
  schema,
  value,
  onChange,
  error,
  disabled = false,
}: PropertyFieldProps) {
  const { name, label, type, required, placeholder, min, max, step, options, description } = schema

  const renderInput = () => {
    switch (type) {
      case 'string':
        return (
          <Input
            id={name}
            type="text"
            value={value || ''}
            onChange={(e) => onChange(e.target.value)}
            placeholder={placeholder}
            disabled={disabled}
            className={cn('h-11', error && 'border-red-500')}
          />
        )

      case 'number':
        return (
          <Input
            id={name}
            type="number"
            value={value ?? ''}
            onChange={(e) => onChange(e.target.value ? Number(e.target.value) : undefined)}
            placeholder={placeholder}
            min={min}
            max={max}
            step={step}
            disabled={disabled}
            className={cn('h-11', error && 'border-red-500')}
          />
        )

      case 'boolean':
        return (
          <div className="flex items-center space-x-2">
            <Switch
              id={name}
              checked={value ?? false}
              onCheckedChange={onChange}
              disabled={disabled}
            />
            <Label htmlFor={name} className="text-sm font-normal cursor-pointer">
              {description || 'Enable'}
            </Label>
          </div>
        )

      case 'select':
        return (
          <Select value={value || ''} onValueChange={onChange} disabled={disabled}>
            <SelectTrigger className={cn('h-11', error && 'border-red-500')}>
              <SelectValue placeholder={placeholder || 'Select an option'} />
            </SelectTrigger>
            <SelectContent>
              {options?.map((option) => {
                const optionValue = typeof option === 'string' ? option : option.value
                const optionLabel = typeof option === 'string' ? option : option.label
                return (
                  <SelectItem key={optionValue} value={optionValue}>
                    {optionLabel}
                  </SelectItem>
                )
              })}
            </SelectContent>
          </Select>
        )

      case 'object':
      case 'json':
        return (
          <JSONEditor
            value={value || {}}
            onChange={onChange}
            height={150}
            showValidation={true}
            disabled={disabled}
          />
        )

      case 'array':
        return (
          <JSONEditor
            value={value || []}
            onChange={onChange}
            height={150}
            showValidation={true}
            disabled={disabled}
          />
        )

      case 'code':
        // Detect language from property name or use default
        const language = name.includes('python') || name.includes('py')
          ? 'python'
          : 'javascript'
        return (
          <CodeEditor
            value={value || ''}
            onChange={onChange}
            language={language}
            height={300}
            showLanguageSelector={true}
            disabled={disabled}
          />
        )

      case 'color':
        return <ColorPicker value={value || '#000000'} onChange={onChange} disabled={disabled} />

      case 'icon':
        return <IconPicker value={value || ''} onChange={onChange} disabled={disabled} />

      case 'cron':
        return (
          <CronBuilder
            value={value || '* * * * *'}
            onChange={onChange}
            enableSeconds={false}
            showPreview={true}
            disabled={disabled}
          />
        )

      case 'mqtt-config':
      case 'mqtt-topic':
        // Determine mode from property name or schema
        const mqttMode = name.includes('out') || name.includes('publish') ? 'publish' : 'subscribe'
        return (
          <MQTTTopicBuilder
            value={value || {}}
            onChange={onChange}
            mode={mqttMode}
            disabled={disabled}
          />
        )

      default:
        return (
          <Input
            id={name}
            type="text"
            value={value || ''}
            onChange={(e) => onChange(e.target.value)}
            placeholder={placeholder}
            disabled={disabled}
            className={cn('h-11', error && 'border-red-500')}
          />
        )
    }
  }

  // For boolean type, the label is inline with the switch
  if (type === 'boolean') {
    return (
      <div className="space-y-2">
        {renderInput()}
        {error && <p className="text-sm text-red-500">{error}</p>}
      </div>
    )
  }

  return (
    <div className="space-y-2">
      <Label htmlFor={name} className="text-sm font-semibold">
        {label}
        {required && <span className="text-red-500 ml-1">*</span>}
      </Label>
      {renderInput()}
      {description && !error && <p className="text-xs text-muted-foreground">{description}</p>}
      {error && <p className="text-sm text-red-500">{error}</p>}
    </div>
  )
}
