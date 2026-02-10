/**
 * Payload Builder
 *
 * User-friendly payload editor with key-value pair builder,
 * type selectors, quick presets, and raw JSON mode.
 */

import { useState, useEffect, useCallback } from 'react'
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
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { JSONEditor } from '@/components/Common/JSONEditor'
import { Plus, Trash2, List, Code, Zap } from 'lucide-react'
import { cn } from '@/lib/utils'

type ValueType = 'string' | 'number' | 'boolean' | 'json'

interface PayloadEntry {
  id: string
  key: string
  valueType: ValueType
  value: any
}

interface PayloadBuilderProps {
  value: Record<string, any>
  onChange: (value: Record<string, any>) => void
  disabled?: boolean
}

interface Preset {
  label: string
  payload: Record<string, any>
}

const PRESETS: Preset[] = [
  { label: 'On', payload: { value: true } },
  { label: 'Off', payload: { value: false } },
  { label: 'State On', payload: { state: 'on' } },
  { label: 'State Off', payload: { state: 'off' } },
  { label: 'Timestamp', payload: { timestamp: '{{now}}' } },
  { label: 'Numeric', payload: { value: 0 } },
  { label: 'Command', payload: { command: 'start' } },
]

let entryCounter = 0
function nextId(): string {
  return `entry-${++entryCounter}`
}

function inferType(val: any): ValueType {
  if (typeof val === 'boolean') return 'boolean'
  if (typeof val === 'number') return 'number'
  if (typeof val === 'string') return 'string'
  return 'json'
}

function objectToEntries(obj: Record<string, any>): PayloadEntry[] {
  if (!obj || typeof obj !== 'object' || Array.isArray(obj)) return []
  return Object.entries(obj).map(([key, val]) => ({
    id: nextId(),
    key,
    valueType: inferType(val),
    value: inferType(val) === 'json' ? JSON.stringify(val, null, 2) : val,
  }))
}

function entriesToObject(entries: PayloadEntry[]): Record<string, any> {
  const result: Record<string, any> = {}
  for (const entry of entries) {
    if (!entry.key.trim()) continue
    switch (entry.valueType) {
      case 'string':
        result[entry.key] = String(entry.value ?? '')
        break
      case 'number':
        result[entry.key] = Number(entry.value ?? 0)
        break
      case 'boolean':
        result[entry.key] = Boolean(entry.value)
        break
      case 'json':
        try {
          result[entry.key] = JSON.parse(entry.value)
        } catch {
          result[entry.key] = entry.value
        }
        break
    }
  }
  return result
}

export function PayloadBuilder({ value, onChange, disabled = false }: PayloadBuilderProps) {
  const [entries, setEntries] = useState<PayloadEntry[]>(() => objectToEntries(value))
  const [activeTab, setActiveTab] = useState<string>('builder')

  // Sync entries → parent onChange
  const emitChange = useCallback(
    (newEntries: PayloadEntry[]) => {
      setEntries(newEntries)
      onChange(entriesToObject(newEntries))
    },
    [onChange]
  )

  // Update entry field
  const updateEntry = useCallback(
    (id: string, field: keyof PayloadEntry, val: any) => {
      const newEntries = entries.map((e) => {
        if (e.id !== id) return e
        const updated = { ...e, [field]: val }
        // Reset value when type changes
        if (field === 'valueType') {
          switch (val as ValueType) {
            case 'string':
              updated.value = ''
              break
            case 'number':
              updated.value = 0
              break
            case 'boolean':
              updated.value = false
              break
            case 'json':
              updated.value = '{}'
              break
          }
        }
        return updated
      })
      emitChange(newEntries)
    },
    [entries, emitChange]
  )

  // Add new entry
  const addEntry = useCallback(() => {
    const newEntries = [
      ...entries,
      { id: nextId(), key: '', valueType: 'string' as ValueType, value: '' },
    ]
    emitChange(newEntries)
  }, [entries, emitChange])

  // Remove entry
  const removeEntry = useCallback(
    (id: string) => {
      emitChange(entries.filter((e) => e.id !== id))
    },
    [entries, emitChange]
  )

  // Apply preset
  const applyPreset = useCallback(
    (preset: Preset) => {
      const newEntries = objectToEntries(preset.payload)
      emitChange(newEntries)
    },
    [emitChange]
  )

  // Handle JSON mode changes
  const handleJsonChange = useCallback(
    (val: any) => {
      if (val && typeof val === 'object' && !Array.isArray(val)) {
        setEntries(objectToEntries(val))
        onChange(val)
      }
    },
    [onChange]
  )

  // Switch tab - sync data between modes
  const handleTabChange = useCallback(
    (tab: string) => {
      if (tab === 'builder' && activeTab === 'json') {
        // Re-hydrate entries from current value
        setEntries(objectToEntries(value))
      }
      setActiveTab(tab)
    },
    [activeTab, value]
  )

  // Render value input based on type
  const renderValueInput = (entry: PayloadEntry) => {
    switch (entry.valueType) {
      case 'boolean':
        return (
          <div className="flex items-center h-9">
            <Switch
              checked={Boolean(entry.value)}
              onCheckedChange={(checked) => updateEntry(entry.id, 'value', checked)}
              disabled={disabled}
            />
            <span className="ml-2 text-xs text-muted-foreground">
              {entry.value ? 'true' : 'false'}
            </span>
          </div>
        )
      case 'number':
        return (
          <Input
            type="number"
            value={entry.value ?? 0}
            onChange={(e) => updateEntry(entry.id, 'value', e.target.value ? Number(e.target.value) : 0)}
            disabled={disabled}
            className="h-9"
          />
        )
      case 'json':
        return (
          <Input
            type="text"
            value={typeof entry.value === 'string' ? entry.value : JSON.stringify(entry.value)}
            onChange={(e) => updateEntry(entry.id, 'value', e.target.value)}
            disabled={disabled}
            placeholder='{"key": "value"}'
            className="h-9 font-mono text-xs"
          />
        )
      default: // string
        return (
          <Input
            type="text"
            value={entry.value ?? ''}
            onChange={(e) => updateEntry(entry.id, 'value', e.target.value)}
            disabled={disabled}
            placeholder="Enter value..."
            className="h-9"
          />
        )
    }
  }

  const currentObject = entriesToObject(entries)
  const hasEntries = entries.length > 0
  const hasDuplicateKeys = new Set(entries.map((e) => e.key).filter(Boolean)).size <
    entries.filter((e) => e.key.trim()).length

  return (
    <div className="space-y-3">
      <Tabs value={activeTab} onValueChange={handleTabChange}>
        <TabsList className="h-8 w-full">
          <TabsTrigger value="builder" className="text-xs gap-1 flex-1">
            <List className="w-3 h-3" />
            Builder
          </TabsTrigger>
          <TabsTrigger value="json" className="text-xs gap-1 flex-1">
            <Code className="w-3 h-3" />
            JSON
          </TabsTrigger>
        </TabsList>

        <TabsContent value="builder" className="mt-3 space-y-3">
          {/* Quick Presets */}
          <div className="space-y-1.5">
            <Label className="text-xs text-muted-foreground flex items-center gap-1">
              <Zap className="w-3 h-3" />
              Quick Presets
            </Label>
            <div className="flex flex-wrap gap-1.5">
              {PRESETS.map((preset) => (
                <Button
                  key={preset.label}
                  variant="outline"
                  size="sm"
                  onClick={() => applyPreset(preset)}
                  disabled={disabled}
                  className="h-7 text-xs px-2.5"
                >
                  {preset.label}
                </Button>
              ))}
            </div>
          </div>

          {/* Key-Value Entries */}
          <div className="space-y-2">
            {entries.map((entry) => (
              <div key={entry.id} className="flex items-start gap-1.5">
                {/* Key */}
                <div className="w-[28%]">
                  <Input
                    type="text"
                    value={entry.key}
                    onChange={(e) => updateEntry(entry.id, 'key', e.target.value)}
                    placeholder="key"
                    disabled={disabled}
                    className={cn(
                      'h-9 text-xs',
                      hasDuplicateKeys &&
                        entries.filter((e) => e.key === entry.key && e.key.trim()).length > 1 &&
                        'border-yellow-500'
                    )}
                  />
                </div>

                {/* Type Selector */}
                <div className="w-[22%]">
                  <Select
                    value={entry.valueType}
                    onValueChange={(v) => updateEntry(entry.id, 'valueType', v)}
                    disabled={disabled}
                  >
                    <SelectTrigger className="h-9 text-xs">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="string">string</SelectItem>
                      <SelectItem value="number">number</SelectItem>
                      <SelectItem value="boolean">boolean</SelectItem>
                      <SelectItem value="json">json</SelectItem>
                    </SelectContent>
                  </Select>
                </div>

                {/* Value Input */}
                <div className="flex-1">{renderValueInput(entry)}</div>

                {/* Delete Button */}
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => removeEntry(entry.id)}
                  disabled={disabled}
                  className="h-9 w-9 p-0 text-muted-foreground hover:text-red-500"
                >
                  <Trash2 className="w-3.5 h-3.5" />
                </Button>
              </div>
            ))}

            {/* Add Key Button */}
            <Button
              variant="outline"
              size="sm"
              onClick={addEntry}
              disabled={disabled}
              className="w-full h-8 text-xs gap-1"
            >
              <Plus className="w-3 h-3" />
              Add Key
            </Button>
          </div>

          {/* Live Preview */}
          {hasEntries && (
            <div className="space-y-1">
              <Label className="text-xs text-muted-foreground">Preview</Label>
              <pre className="text-xs bg-muted/50 rounded-md p-2 overflow-x-auto font-mono border">
                {JSON.stringify(currentObject, null, 2)}
              </pre>
            </div>
          )}

          {/* Duplicate key warning */}
          {hasDuplicateKeys && (
            <p className="text-xs text-yellow-500">Warning: Duplicate keys detected</p>
          )}
        </TabsContent>

        <TabsContent value="json" className="mt-3">
          <JSONEditor
            value={value || {}}
            onChange={handleJsonChange}
            height={150}
            showValidation={true}
            disabled={disabled}
          />
        </TabsContent>
      </Tabs>
    </div>
  )
}
