/**
 * Switch Rule Builder
 *
 * Visual rule builder for Switch node with multiple conditions and outputs
 */

import { useState } from 'react'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Plus, Trash2, GripVertical, Copy } from 'lucide-react'
import { cn } from '@/lib/utils'

export type RuleOperator =
  | 'eq'
  | 'neq'
  | 'lt'
  | 'lte'
  | 'gt'
  | 'gte'
  | 'contains'
  | 'notContains'
  | 'matches'
  | 'isEmpty'
  | 'isNotEmpty'
  | 'isTrue'
  | 'isFalse'
  | 'isNull'
  | 'isNotNull'
  | 'between'

export type RuleProperty = 'msg.payload' | 'msg.topic' | 'msg' | 'flow' | 'global' | 'custom'

export interface SwitchRule {
  id: string
  property: string
  propertyType: RuleProperty
  operator: RuleOperator
  value: any
  value2?: any // for 'between' operator
  outputIndex: number
}

interface SwitchRuleBuilderProps {
  value: {
    rules?: SwitchRule[]
    checkAll?: boolean
    outputCount?: number
  }
  onChange: (value: any) => void
  disabled?: boolean
}

const OPERATORS: { value: RuleOperator; label: string; requiresValue: boolean; requiresValue2?: boolean }[] = [
  { value: 'eq', label: '== (equals)', requiresValue: true },
  { value: 'neq', label: '!= (not equals)', requiresValue: true },
  { value: 'lt', label: '< (less than)', requiresValue: true },
  { value: 'lte', label: '<= (less than or equal)', requiresValue: true },
  { value: 'gt', label: '> (greater than)', requiresValue: true },
  { value: 'gte', label: '>= (greater than or equal)', requiresValue: true },
  { value: 'contains', label: 'contains', requiresValue: true },
  { value: 'notContains', label: 'does not contain', requiresValue: true },
  { value: 'matches', label: 'matches regex', requiresValue: true },
  { value: 'between', label: 'is between', requiresValue: true, requiresValue2: true },
  { value: 'isEmpty', label: 'is empty', requiresValue: false },
  { value: 'isNotEmpty', label: 'is not empty', requiresValue: false },
  { value: 'isTrue', label: 'is true', requiresValue: false },
  { value: 'isFalse', label: 'is false', requiresValue: false },
  { value: 'isNull', label: 'is null', requiresValue: false },
  { value: 'isNotNull', label: 'is not null', requiresValue: false },
]

const PROPERTY_TYPES: { value: RuleProperty; label: string; example: string }[] = [
  { value: 'msg.payload', label: 'msg.payload', example: 'msg.payload' },
  { value: 'msg.topic', label: 'msg.topic', example: 'msg.topic' },
  { value: 'msg', label: 'msg property', example: 'msg.property' },
  { value: 'flow', label: 'flow context', example: 'flow.variable' },
  { value: 'global', label: 'global context', example: 'global.variable' },
  { value: 'custom', label: 'JSONata expression', example: '$length(payload)' },
]

export function SwitchRuleBuilder({ value, onChange, disabled = false }: SwitchRuleBuilderProps) {
  // Handle undefined or null value
  const safeValue = value || { rules: [], checkAll: false, outputCount: 2 }
  const rules = safeValue.rules || []
  const checkAll = safeValue.checkAll ?? false
  const outputCount = safeValue.outputCount || Math.max(2, rules.length)

  const [draggedIndex, setDraggedIndex] = useState<number | null>(null)

  const generateRuleId = (): string => {
    return `rule_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
  }

  const addRule = () => {
    const newRule: SwitchRule = {
      id: generateRuleId(),
      property: 'payload',
      propertyType: 'msg.payload',
      operator: 'eq',
      value: '',
      outputIndex: rules.length % outputCount,
    }
    onChange({ ...safeValue, rules: [...rules, newRule] })
  }

  const removeRule = (id: string) => {
    onChange({ ...safeValue, rules: rules.filter((r) => r.id !== id) })
  }

  const duplicateRule = (rule: SwitchRule) => {
    const newRule: SwitchRule = {
      ...rule,
      id: generateRuleId(),
    }
    onChange({ ...safeValue, rules: [...rules, newRule] })
  }

  const updateRule = (id: string, updates: Partial<SwitchRule>) => {
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

  const getOperatorConfig = (operator: RuleOperator) => {
    return OPERATORS.find((op) => op.value === operator)
  }

  return (
    <div className="space-y-6">
      {/* Settings */}
      <div className="space-y-4">
        <div className="grid grid-cols-2 gap-4">
          <div className="space-y-2">
            <Label htmlFor="outputCount" className="text-sm font-semibold">
              Number of Outputs
            </Label>
            <Input
              id="outputCount"
              type="number"
              min={1}
              max={16}
              value={outputCount}
              onChange={(e) => onChange({ ...safeValue, outputCount: Number(e.target.value) })}
              className="h-11"
              disabled={disabled}
            />
          </div>

          <div className="flex items-center space-x-2 mt-8">
            <input
              type="checkbox"
              id="checkAll"
              checked={checkAll}
              onChange={(e) => onChange({ ...safeValue, checkAll: e.target.checked })}
              disabled={disabled}
              className="rounded"
            />
            <Label htmlFor="checkAll" className="text-sm font-normal cursor-pointer">
              Check all rules (otherwise stop at first match)
            </Label>
          </div>
        </div>
      </div>

      {/* Rules */}
      <div className="space-y-3">
        <div className="flex items-center justify-between">
          <Label className="text-sm font-semibold">Rules</Label>
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
              No rules defined. Add rules to route messages to different outputs.
            </p>
          </div>
        )}

        <div className="space-y-2">
          {rules.map((rule, index) => {
            const operatorConfig = getOperatorConfig(rule.operator)
            const propertyTypeConfig = PROPERTY_TYPES.find((pt) => pt.value === rule.propertyType)

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
                    → Output {rule.outputIndex + 1}
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

                {/* Property Type and Name */}
                <div className="grid grid-cols-2 gap-2">
                  <div className="space-y-1">
                    <Label className="text-xs">Property Type</Label>
                    <Select
                      value={rule.propertyType}
                      onValueChange={(value: RuleProperty) => {
                        const example = PROPERTY_TYPES.find((pt) => pt.value === value)?.example || ''
                        updateRule(rule.id, {
                          propertyType: value,
                          property: value === 'msg.payload' ? 'payload' : value === 'msg.topic' ? 'topic' : example,
                        })
                      }}
                      disabled={disabled}
                    >
                      <SelectTrigger className="h-9">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        {PROPERTY_TYPES.map((pt) => (
                          <SelectItem key={pt.value} value={pt.value}>
                            {pt.label}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>

                  <div className="space-y-1">
                    <Label className="text-xs">Property Name</Label>
                    <Input
                      value={rule.property}
                      onChange={(e) => updateRule(rule.id, { property: e.target.value })}
                      placeholder={propertyTypeConfig?.example}
                      className="h-9 font-mono text-xs"
                      disabled={disabled}
                    />
                  </div>
                </div>

                {/* Operator */}
                <div className="space-y-1">
                  <Label className="text-xs">Operator</Label>
                  <Select
                    value={rule.operator}
                    onValueChange={(value: RuleOperator) => updateRule(rule.id, { operator: value })}
                    disabled={disabled}
                  >
                    <SelectTrigger className="h-9">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {OPERATORS.map((op) => (
                        <SelectItem key={op.value} value={op.value}>
                          {op.label}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>

                {/* Value(s) */}
                {operatorConfig?.requiresValue && (
                  <div className={cn('grid gap-2', operatorConfig.requiresValue2 ? 'grid-cols-2' : 'grid-cols-1')}>
                    <div className="space-y-1">
                      <Label className="text-xs">
                        {operatorConfig.requiresValue2 ? 'From Value' : 'Value'}
                      </Label>
                      <Input
                        value={rule.value}
                        onChange={(e) => updateRule(rule.id, { value: e.target.value })}
                        placeholder="Enter value"
                        className="h-9 font-mono text-xs"
                        disabled={disabled}
                      />
                    </div>

                    {operatorConfig.requiresValue2 && (
                      <div className="space-y-1">
                        <Label className="text-xs">To Value</Label>
                        <Input
                          value={rule.value2 || ''}
                          onChange={(e) => updateRule(rule.id, { value2: e.target.value })}
                          placeholder="Enter value"
                          className="h-9 font-mono text-xs"
                          disabled={disabled}
                        />
                      </div>
                    )}
                  </div>
                )}

                {/* Output Selection */}
                <div className="space-y-1">
                  <Label className="text-xs">Send to Output</Label>
                  <Select
                    value={rule.outputIndex.toString()}
                    onValueChange={(value) => updateRule(rule.id, { outputIndex: Number(value) })}
                    disabled={disabled}
                  >
                    <SelectTrigger className="h-9">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {Array.from({ length: outputCount }, (_, i) => (
                        <SelectItem key={i} value={i.toString()}>
                          Output {i + 1}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
              </div>
            )
          })}
        </div>

        {rules.length > 0 && (
          <p className="text-xs text-muted-foreground">
            Drag rules to reorder. {checkAll ? 'All matching rules will be executed.' : 'Messages will be sent to the first matching rule output.'}
          </p>
        )}
      </div>

      {/* Help Text */}
      <div className="p-4 bg-muted/50 rounded-lg border">
        <p className="text-sm font-semibold mb-2">How it works</p>
        <div className="space-y-2 text-xs text-muted-foreground">
          <p>
            1. Each rule checks a property against a condition
          </p>
          <p>
            2. If the condition matches, the message is sent to the specified output
          </p>
          <p>
            3. {checkAll
              ? 'All rules are checked, and messages may be sent to multiple outputs'
              : 'Processing stops at the first matching rule'}
          </p>
          <p>
            4. Messages that don't match any rule are dropped (unless you add a catch-all rule)
          </p>
        </div>
      </div>
    </div>
  )
}
