/**
 * Change Transformation Builder
 *
 * Visual builder for message transformation rules (set, change, delete, move)
 */

import { useState } from 'react'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Plus, Trash2, GripVertical, Copy } from 'lucide-react'
import { cn } from '@/lib/utils'

export type ChangeAction = 'set' | 'change' | 'delete' | 'move'
export type ValueType = 'string' | 'number' | 'boolean' | 'json' | 'timestamp' | 'env' | 'msg' | 'flow' | 'global' | 'jsonata'

export interface ChangeRule {
  id: string
  action: ChangeAction
  property: string
  propertyType: 'msg' | 'flow' | 'global'
  valueType?: ValueType
  value?: any
  to?: string
  toType?: 'msg' | 'flow' | 'global'
}

interface ChangeTransformBuilderProps {
  value: {
    rules?: ChangeRule[]
  }
  onChange: (value: any) => void
  disabled?: boolean
}

const VALUE_TYPES: { value: ValueType; label: string; example: string }[] = [
  { value: 'string', label: 'String', example: 'Hello World' },
  { value: 'number', label: 'Number', example: '42' },
  { value: 'boolean', label: 'Boolean', example: 'true' },
  { value: 'json', label: 'JSON', example: '{"key": "value"}' },
  { value: 'timestamp', label: 'Timestamp', example: 'Current timestamp' },
  { value: 'env', label: 'Environment Variable', example: 'HOME' },
  { value: 'msg', label: 'Message Property', example: 'msg.payload' },
  { value: 'flow', label: 'Flow Context', example: 'flow.variable' },
  { value: 'global', label: 'Global Context', example: 'global.variable' },
  { value: 'jsonata', label: 'JSONata Expression', example: '$uppercase(payload)' },
]

export function ChangeTransformBuilder({ value, onChange, disabled = false }: ChangeTransformBuilderProps) {
  // Handle undefined or null value
  const safeValue = value || { rules: [] }
  const rules = safeValue.rules || []
  const [draggedIndex, setDraggedIndex] = useState<number | null>(null)

  const generateRuleId = (): string => {
    return `rule_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
  }

  const addRule = () => {
    const newRule: ChangeRule = {
      id: generateRuleId(),
      action: 'set',
      property: 'payload',
      propertyType: 'msg',
      valueType: 'string',
      value: '',
    }
    onChange({ ...safeValue, rules: [...rules, newRule] })
  }

  const removeRule = (id: string) => {
    onChange({ ...safeValue, rules: rules.filter((r) => r.id !== id) })
  }

  const duplicateRule = (rule: ChangeRule) => {
    const newRule: ChangeRule = {
      ...rule,
      id: generateRuleId(),
    }
    onChange({ ...safeValue, rules: [...rules, newRule] })
  }

  const updateRule = (id: string, updates: Partial<ChangeRule>) => {
    const newRules = rules.map((r) => (r.id === id ? { ...r, ...updates } : r))
    onChange({ ...safeValue, rules: newRules })
  }

  const handleDragStart = (index: number) => {
    setDraggedIndex(index)
  }

  const handleDragOver = (e: React.DragEvent, index: number) => {
    e.preventDefault()
    if (draggedIndex === null || draggedIndex === index) return

    const newRules = [...rules]
    const draggedItem = newRules[draggedIndex]
    newRules.splice(draggedIndex, 1)
    newRules.splice(index, 0, draggedItem)

    setDraggedIndex(index)
    onChange({ ...safeValue, rules: newRules })
  }

  const handleDragEnd = () => {
    setDraggedIndex(null)
  }

  const getValueTypeConfig = (valueType: ValueType) => {
    return VALUE_TYPES.find((vt) => vt.value === valueType)
  }

  const needsValue = (action: ChangeAction) => {
    return action === 'set' || action === 'change'
  }

  const needsTo = (action: ChangeAction) => {
    return action === 'move'
  }

  return (
    <div className="space-y-6">
      {/* Rules */}
      <div className="space-y-3">
        <div className="flex items-center justify-between">
          <Label className="text-sm font-semibold">Transformation Rules</Label>
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={addRule}
            className="h-8"
            disabled={disabled}
          >
            <Plus className="w-4 h-4 mr-1" />
            Add Rule
          </Button>
        </div>

        {rules.length === 0 && (
          <div className="text-center py-8 border-2 border-dashed rounded-lg">
            <p className="text-sm text-muted-foreground">
              No transformation rules defined. Add rules to modify messages.
            </p>
          </div>
        )}

        <div className="space-y-2">
          {rules.map((rule, index) => {
            const valueTypeConfig = getValueTypeConfig(rule.valueType || 'string')

            return (
              <div
                key={rule.id}
                draggable
                onDragStart={() => handleDragStart(index)}
                onDragOver={(e) => handleDragOver(e, index)}
                onDragEnd={handleDragEnd}
                className={cn(
                  'border rounded-lg bg-card p-3 space-y-3 transition-colors',
                  draggedIndex === index && 'opacity-50'
                )}
              >
                {/* Rule Header */}
                <div className="flex items-center gap-2">
                  <GripVertical className="w-4 h-4 text-muted-foreground flex-shrink-0 cursor-move" />
                  <span className="text-xs font-semibold text-muted-foreground">
                    Rule {index + 1}
                  </span>
                  <span className="text-xs text-muted-foreground flex-1">
                    {rule.action.toUpperCase()} {rule.propertyType}.{rule.property}
                  </span>

                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    onClick={() => duplicateRule(rule)}
                    className="h-7 w-7 p-0"
                    disabled={disabled}
                  >
                    <Copy className="w-3 h-3" />
                  </Button>
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    onClick={() => removeRule(rule.id)}
                    className="h-7 w-7 p-0 text-red-600 hover:text-red-700 hover:bg-red-50"
                    disabled={disabled}
                  >
                    <Trash2 className="w-3 h-3" />
                  </Button>
                </div>

                {/* Action */}
                <div className="space-y-1">
                  <Label className="text-xs">Action</Label>
                  <Select
                    value={rule.action}
                    onValueChange={(value: ChangeAction) => updateRule(rule.id, { action: value })}
                    disabled={disabled}
                  >
                    <SelectTrigger className="h-9">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="set">Set - Set property value</SelectItem>
                      <SelectItem value="change">Change - Replace value</SelectItem>
                      <SelectItem value="delete">Delete - Remove property</SelectItem>
                      <SelectItem value="move">Move - Move/rename property</SelectItem>
                    </SelectContent>
                  </Select>
                </div>

                {/* Property */}
                <div className="grid grid-cols-2 gap-2">
                  <div className="space-y-1">
                    <Label className="text-xs">Property Type</Label>
                    <Select
                      value={rule.propertyType}
                      onValueChange={(value: 'msg' | 'flow' | 'global') =>
                        updateRule(rule.id, { propertyType: value })
                      }
                      disabled={disabled}
                    >
                      <SelectTrigger className="h-9">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="msg">msg</SelectItem>
                        <SelectItem value="flow">flow</SelectItem>
                        <SelectItem value="global">global</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>

                  <div className="space-y-1">
                    <Label className="text-xs">Property Name</Label>
                    <Input
                      value={rule.property}
                      onChange={(e) => updateRule(rule.id, { property: e.target.value })}
                      placeholder="payload"
                      className="h-9 font-mono text-xs"
                      disabled={disabled}
                    />
                  </div>
                </div>

                {/* Value (for set/change) */}
                {needsValue(rule.action) && (
                  <>
                    <div className="space-y-1">
                      <Label className="text-xs">Value Type</Label>
                      <Select
                        value={rule.valueType || 'string'}
                        onValueChange={(value: ValueType) => updateRule(rule.id, { valueType: value })}
                        disabled={disabled}
                      >
                        <SelectTrigger className="h-9">
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
                    </div>

                    {rule.valueType !== 'timestamp' && (
                      <div className="space-y-1">
                        <Label className="text-xs">
                          Value
                          <span className="text-muted-foreground ml-2">
                            (e.g., {valueTypeConfig?.example})
                          </span>
                        </Label>
                        {rule.valueType === 'json' || rule.valueType === 'jsonata' ? (
                          <Textarea
                            value={rule.value || ''}
                            onChange={(e) => updateRule(rule.id, { value: e.target.value })}
                            placeholder={valueTypeConfig?.example}
                            rows={3}
                            className="font-mono text-xs"
                            disabled={disabled}
                          />
                        ) : (
                          <Input
                            value={rule.value || ''}
                            onChange={(e) => updateRule(rule.id, { value: e.target.value })}
                            placeholder={valueTypeConfig?.example}
                            className="h-9 font-mono text-xs"
                            disabled={disabled}
                          />
                        )}
                      </div>
                    )}
                  </>
                )}

                {/* To (for move) */}
                {needsTo(rule.action) && (
                  <div className="grid grid-cols-2 gap-2">
                    <div className="space-y-1">
                      <Label className="text-xs">Move To Type</Label>
                      <Select
                        value={rule.toType || 'msg'}
                        onValueChange={(value: 'msg' | 'flow' | 'global') =>
                          updateRule(rule.id, { toType: value })
                        }
                        disabled={disabled}
                      >
                        <SelectTrigger className="h-9">
                          <SelectValue />
                        </SelectTrigger>
                        <SelectContent>
                          <SelectItem value="msg">msg</SelectItem>
                          <SelectItem value="flow">flow</SelectItem>
                          <SelectItem value="global">global</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>

                    <div className="space-y-1">
                      <Label className="text-xs">Move To Property</Label>
                      <Input
                        value={rule.to || ''}
                        onChange={(e) => updateRule(rule.id, { to: e.target.value })}
                        placeholder="newProperty"
                        className="h-9 font-mono text-xs"
                        disabled={disabled}
                      />
                    </div>
                  </div>
                )}
              </div>
            )
          })}
        </div>

        {rules.length > 0 && (
          <p className="text-xs text-muted-foreground">
            Drag rules to reorder. Rules are executed in order from top to bottom.
          </p>
        )}
      </div>

      {/* Help Text */}
      <div className="p-4 bg-muted/50 rounded-lg border">
        <p className="text-sm font-semibold mb-2">Action Types</p>
        <div className="space-y-2 text-xs text-muted-foreground">
          <div>
            <span className="font-semibold">Set:</span> Set a property to a value (creates if doesn't exist)
          </div>
          <div>
            <span className="font-semibold">Change:</span> Replace existing value (does nothing if property doesn't exist)
          </div>
          <div>
            <span className="font-semibold">Delete:</span> Remove a property completely
          </div>
          <div>
            <span className="font-semibold">Move:</span> Move/rename a property to a new location
          </div>
        </div>
      </div>
    </div>
  )
}
