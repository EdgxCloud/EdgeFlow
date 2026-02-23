/**
 * Function Transform Builder
 *
 * Form-based editor for function node transformation rules.
 * Users add set/delete rules instead of writing code.
 */

import { useState } from 'react'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Plus, Trash2, GripVertical, Copy, Code, ArrowRight, Wand2 } from 'lucide-react'
import { cn } from '@/lib/utils'

export type FunctionAction = 'set' | 'delete'
export type FunctionValueType = 'string' | 'number' | 'boolean' | 'json' | 'msg' | 'expression'

export interface FunctionRule {
  id: string
  action: FunctionAction
  property: string
  valueType?: FunctionValueType
  value?: any
}

interface FunctionTransformBuilderProps {
  value: {
    rules?: FunctionRule[]
    code?: string // legacy DSL code support
  }
  onChange: (value: any) => void
  disabled?: boolean
}

const VALUE_TYPES: { value: FunctionValueType; label: string; placeholder: string }[] = [
  { value: 'string', label: 'String', placeholder: 'Enter text value...' },
  { value: 'number', label: 'Number', placeholder: 'e.g. 42' },
  { value: 'boolean', label: 'Boolean', placeholder: '' },
  { value: 'json', label: 'JSON', placeholder: '{"key": "value"}' },
  { value: 'msg', label: 'Reference', placeholder: 'e.g. temperature (copies from payload)' },
  { value: 'expression', label: 'Expression', placeholder: 'e.g. msg.payload.temp * 1.8 + 32' },
]

let ruleCounter = 0
function nextRuleId(): string {
  return `rule-${Date.now()}-${++ruleCounter}`
}

function createDefaultRule(action: FunctionAction = 'set'): FunctionRule {
  return {
    id: nextRuleId(),
    action,
    property: '',
    valueType: 'string',
    value: '',
  }
}

/**
 * Try to convert legacy DSL code into typed rules.
 * Handles: set key = value, delete key, msg.payload.key = value
 */
function parseLegacyCode(code: string): FunctionRule[] {
  const rules: FunctionRule[] = []
  const lines = code.split('\n')

  for (const rawLine of lines) {
    const line = rawLine.trim()
    if (!line || line.startsWith('//') || line.startsWith('return')) continue

    // set key = value
    if (line.startsWith('set ')) {
      const rest = line.slice(4)
      const eqIdx = rest.indexOf('=')
      if (eqIdx === -1) continue
      const key = rest.slice(0, eqIdx).trim()
      let val = rest.slice(eqIdx + 1).trim().replace(/;$/, '')

      rules.push({
        id: nextRuleId(),
        action: 'set',
        property: key,
        ...inferValueType(val),
      })
      continue
    }

    // msg.payload.key = value
    if (line.startsWith('msg.payload.') && line.includes('=')) {
      const rest = line.slice('msg.payload.'.length)
      const eqIdx = rest.indexOf('=')
      if (eqIdx === -1) continue
      const key = rest.slice(0, eqIdx).trim()
      let val = rest.slice(eqIdx + 1).trim().replace(/;$/, '')

      rules.push({
        id: nextRuleId(),
        action: 'set',
        property: key,
        ...inferValueType(val),
      })
      continue
    }

    // delete key
    if (line.startsWith('delete ')) {
      const key = line.slice(7).trim().replace(/;$/, '').replace(/^msg\.payload\./, '')
      rules.push({
        id: nextRuleId(),
        action: 'delete',
        property: key,
      })
    }
  }

  return rules
}

function inferValueType(val: string): { valueType: FunctionValueType; value: any } {
  // Boolean
  if (val === 'true') return { valueType: 'boolean', value: true }
  if (val === 'false') return { valueType: 'boolean', value: false }

  // Quoted string
  if ((val.startsWith('"') && val.endsWith('"')) || (val.startsWith("'") && val.endsWith("'"))) {
    return { valueType: 'string', value: val.slice(1, -1) }
  }

  // Reference
  if (val.startsWith('msg.payload.') && !val.includes(' ')) {
    return { valueType: 'msg', value: val.slice('msg.payload.'.length) }
  }

  // Expression (contains operators or references)
  if (val.includes(' + ') || val.includes(' - ') || val.includes(' * ') || val.includes(' / ') || val.includes('msg.payload.')) {
    return { valueType: 'expression', value: val }
  }

  // Number
  const num = Number(val)
  if (!isNaN(num) && val !== '') {
    return { valueType: 'number', value: num }
  }

  // JSON object/array
  if ((val.startsWith('{') && val.endsWith('}')) || (val.startsWith('[') && val.endsWith(']'))) {
    try {
      return { valueType: 'json', value: val }
    } catch {
      // fall through
    }
  }

  // Default to string
  return { valueType: 'string', value: val }
}

export function FunctionTransformBuilder({
  value,
  onChange,
  disabled = false,
}: FunctionTransformBuilderProps) {
  const safeValue = value || {}
  const hasLegacyCode = !safeValue.rules && typeof safeValue.code === 'string' && safeValue.code.trim() !== ''
  const [showLegacy, setShowLegacy] = useState(hasLegacyCode)

  const rules: FunctionRule[] = (safeValue.rules || []).map((r: any) => ({
    ...r,
    id: r.id || nextRuleId(),
  }))

  // --- Drag state ---
  const [dragIdx, setDragIdx] = useState<number | null>(null)

  const updateRules = (newRules: FunctionRule[]) => {
    onChange({ ...safeValue, rules: newRules, code: undefined })
  }

  const addRule = (action: FunctionAction = 'set') => {
    updateRules([...rules, createDefaultRule(action)])
  }

  const removeRule = (id: string) => {
    updateRules(rules.filter((r) => r.id !== id))
  }

  const duplicateRule = (id: string) => {
    const idx = rules.findIndex((r) => r.id === id)
    if (idx === -1) return
    const clone = { ...rules[idx], id: nextRuleId() }
    const newRules = [...rules]
    newRules.splice(idx + 1, 0, clone)
    updateRules(newRules)
  }

  const updateRule = (id: string, updates: Partial<FunctionRule>) => {
    updateRules(
      rules.map((r) => {
        if (r.id !== id) return r
        const updated = { ...r, ...updates }
        // Reset value when action or value type changes
        if (updates.action === 'delete') {
          delete updated.valueType
          delete updated.value
        }
        if (updates.valueType !== undefined && updates.valueType !== r.valueType) {
          switch (updates.valueType) {
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
            case 'msg':
              updated.value = ''
              break
            case 'expression':
              updated.value = ''
              break
          }
        }
        return updated
      })
    )
  }

  // --- Drag handlers ---
  const handleDragStart = (idx: number) => setDragIdx(idx)
  const handleDragOver = (e: React.DragEvent, idx: number) => {
    e.preventDefault()
    if (dragIdx === null || dragIdx === idx) return
    const newRules = [...rules]
    const [moved] = newRules.splice(dragIdx, 1)
    newRules.splice(idx, 0, moved)
    updateRules(newRules)
    setDragIdx(idx)
  }
  const handleDragEnd = () => setDragIdx(null)

  // --- Convert legacy code ---
  const convertLegacy = () => {
    if (safeValue.code) {
      const converted = parseLegacyCode(safeValue.code)
      onChange({ rules: converted })
      setShowLegacy(false)
    }
  }

  // --- Render value input based on type ---
  const renderValueInput = (rule: FunctionRule) => {
    const vt = rule.valueType || 'string'
    const vtInfo = VALUE_TYPES.find((v) => v.value === vt)

    switch (vt) {
      case 'boolean':
        return (
          <div className="flex items-center gap-2 h-9">
            <Switch
              checked={Boolean(rule.value)}
              onCheckedChange={(checked) => updateRule(rule.id, { value: checked })}
              disabled={disabled}
            />
            <span className="text-xs text-muted-foreground">
              {rule.value ? 'true' : 'false'}
            </span>
          </div>
        )
      case 'number':
        return (
          <Input
            type="number"
            value={rule.value ?? 0}
            onChange={(e) =>
              updateRule(rule.id, { value: e.target.value ? Number(e.target.value) : 0 })
            }
            disabled={disabled}
            placeholder={vtInfo?.placeholder}
            className="h-9"
          />
        )
      case 'json':
        return (
          <Textarea
            value={typeof rule.value === 'string' ? rule.value : JSON.stringify(rule.value, null, 2)}
            onChange={(e) => updateRule(rule.id, { value: e.target.value })}
            disabled={disabled}
            placeholder={vtInfo?.placeholder}
            className="font-mono text-xs min-h-[60px]"
            rows={2}
          />
        )
      case 'expression':
        return (
          <Input
            type="text"
            value={rule.value ?? ''}
            onChange={(e) => updateRule(rule.id, { value: e.target.value })}
            disabled={disabled}
            placeholder={vtInfo?.placeholder}
            className="h-9 font-mono text-xs"
          />
        )
      case 'msg':
        return (
          <div className="flex items-center gap-1">
            <span className="text-xs text-muted-foreground whitespace-nowrap">payload.</span>
            <Input
              type="text"
              value={rule.value ?? ''}
              onChange={(e) => updateRule(rule.id, { value: e.target.value })}
              disabled={disabled}
              placeholder="key name"
              className="h-9 text-xs"
            />
          </div>
        )
      default: // string
        return (
          <Input
            type="text"
            value={rule.value ?? ''}
            onChange={(e) => updateRule(rule.id, { value: e.target.value })}
            disabled={disabled}
            placeholder={vtInfo?.placeholder}
            className="h-9"
          />
        )
    }
  }

  // --- Legacy code banner ---
  if (showLegacy && hasLegacyCode) {
    return (
      <div className="space-y-3">
        <div className="rounded-lg border border-yellow-500/30 bg-yellow-500/5 p-3 space-y-2">
          <div className="flex items-center gap-2">
            <Code className="w-4 h-4 text-yellow-500" />
            <span className="text-sm font-medium text-yellow-500">Legacy Code Detected</span>
          </div>
          <p className="text-xs text-muted-foreground">
            This function node uses legacy code format. Convert to rules for easier editing.
          </p>
          <pre className="text-xs bg-muted/50 rounded-md p-2 overflow-x-auto font-mono border max-h-32">
            {safeValue.code}
          </pre>
          <div className="flex gap-2">
            <Button
              variant="default"
              size="sm"
              onClick={convertLegacy}
              className="text-xs gap-1"
            >
              <Wand2 className="w-3 h-3" />
              Convert to Rules
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => {
                onChange({ rules: [] })
                setShowLegacy(false)
              }}
              className="text-xs"
            >
              Start Fresh
            </Button>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-3">
      {/* Rules List */}
      <div className="space-y-2">
        {rules.length === 0 && (
          <div className="text-center py-6 text-muted-foreground border border-dashed rounded-lg">
            <Wand2 className="w-5 h-5 mx-auto mb-2 opacity-50" />
            <p className="text-xs">No transformation rules</p>
            <p className="text-xs opacity-70">
              Add rules to modify message payload, or leave empty for pass-through.
            </p>
          </div>
        )}

        {rules.map((rule, idx) => (
          <div
            key={rule.id}
            draggable={!disabled}
            onDragStart={() => handleDragStart(idx)}
            onDragOver={(e) => handleDragOver(e, idx)}
            onDragEnd={handleDragEnd}
            className={cn(
              'rounded-lg border bg-card p-3 space-y-2 transition-colors',
              dragIdx === idx && 'border-primary/50 bg-primary/5'
            )}
          >
            {/* Rule Header */}
            <div className="flex items-center gap-1.5">
              <GripVertical className="w-3.5 h-3.5 text-muted-foreground cursor-grab flex-shrink-0" />
              <span className="text-xs text-muted-foreground font-medium">
                #{idx + 1}
              </span>

              {/* Action */}
              <Select
                value={rule.action}
                onValueChange={(v) => updateRule(rule.id, { action: v as FunctionAction })}
                disabled={disabled}
              >
                <SelectTrigger className="h-7 w-[90px] text-xs">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="set">Set</SelectItem>
                  <SelectItem value="delete">Delete</SelectItem>
                </SelectContent>
              </Select>

              {/* Arrow */}
              <ArrowRight className="w-3 h-3 text-muted-foreground flex-shrink-0" />

              {/* Property name */}
              <div className="flex items-center gap-1 flex-1 min-w-0">
                <span className="text-xs text-muted-foreground whitespace-nowrap">payload.</span>
                <Input
                  type="text"
                  value={rule.property}
                  onChange={(e) => updateRule(rule.id, { property: e.target.value })}
                  disabled={disabled}
                  placeholder="property name"
                  className="h-7 text-xs"
                />
              </div>

              {/* Actions */}
              <Button
                variant="ghost"
                size="sm"
                onClick={() => duplicateRule(rule.id)}
                disabled={disabled}
                className="h-7 w-7 p-0 text-muted-foreground hover:text-foreground"
                title="Duplicate rule"
              >
                <Copy className="w-3 h-3" />
              </Button>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => removeRule(rule.id)}
                disabled={disabled}
                className="h-7 w-7 p-0 text-muted-foreground hover:text-red-500"
                title="Remove rule"
              >
                <Trash2 className="w-3 h-3" />
              </Button>
            </div>

            {/* Value section (only for "set" action) */}
            {rule.action === 'set' && (
              <div className="flex items-start gap-1.5 ml-5">
                {/* Value type selector */}
                <Select
                  value={rule.valueType || 'string'}
                  onValueChange={(v) =>
                    updateRule(rule.id, { valueType: v as FunctionValueType })
                  }
                  disabled={disabled}
                >
                  <SelectTrigger className="h-9 w-[110px] text-xs flex-shrink-0">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {VALUE_TYPES.map((vt) => (
                      <SelectItem key={vt.value} value={vt.value}>
                        {vt.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>

                {/* Value input */}
                <div className="flex-1 min-w-0">{renderValueInput(rule)}</div>
              </div>
            )}
          </div>
        ))}
      </div>

      {/* Add Rule Buttons */}
      <div className="flex gap-2">
        <Button
          variant="outline"
          size="sm"
          onClick={() => addRule('set')}
          disabled={disabled}
          className="flex-1 h-8 text-xs gap-1"
        >
          <Plus className="w-3 h-3" />
          Set Property
        </Button>
        <Button
          variant="outline"
          size="sm"
          onClick={() => addRule('delete')}
          disabled={disabled}
          className="flex-1 h-8 text-xs gap-1"
        >
          <Trash2 className="w-3 h-3" />
          Delete Property
        </Button>
      </div>

      {/* Help Section */}
      <div className="rounded-lg border bg-muted/30 p-3 space-y-2">
        <Label className="text-xs font-medium text-muted-foreground">How it works</Label>
        <div className="space-y-1.5 text-xs text-muted-foreground">
          <p>
            <span className="font-medium text-foreground">Set</span> — Add or update a property in
            the message payload
          </p>
          <p>
            <span className="font-medium text-foreground">Delete</span> — Remove a property from
            the payload
          </p>
          <p className="pt-1 border-t border-border/50">
            <span className="font-medium text-foreground">Value types:</span>{' '}
            String, Number, Boolean, JSON (object/array), Reference (copy another payload key),
            Expression (arithmetic like <code className="bg-muted px-1 rounded">temp * 1.8 + 32</code>)
          </p>
          <p>
            Leave rules empty for <span className="font-medium text-foreground">pass-through</span>{' '}
            (message flows unchanged).
          </p>
        </div>
      </div>
    </div>
  )
}
