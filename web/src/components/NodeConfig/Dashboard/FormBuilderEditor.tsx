/**
 * Form Builder Editor
 *
 * Specialized editor for dashboard form widget with drag-to-reorder fields
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
import { Textarea } from '@/components/ui/textarea'
import { Plus, Trash2, GripVertical, Copy } from 'lucide-react'
import { cn } from '@/lib/utils'

export type FormFieldType =
  | 'text'
  | 'number'
  | 'email'
  | 'password'
  | 'textarea'
  | 'select'
  | 'checkbox'
  | 'radio'
  | 'date'
  | 'time'
  | 'datetime'
  | 'file'
  | 'color'
  | 'range'

export interface FormField {
  id: string
  type: FormFieldType
  label: string
  name: string
  placeholder?: string
  defaultValue?: any
  required?: boolean
  disabled?: boolean
  options?: string[] // for select, radio
  min?: number // for number, range
  max?: number // for number, range
  step?: number // for number, range
  pattern?: string // for text validation
  helperText?: string
  rows?: number // for textarea
}

interface FormBuilderEditorProps {
  config: {
    submitButtonText?: string
    resetButtonText?: string
    showResetButton?: boolean
    formLayout?: 'vertical' | 'horizontal'
    fields?: FormField[]
    outputTopic?: string
    validateOnChange?: boolean
  }
  onChange: (config: any) => void
}

const FIELD_TYPES: { value: FormFieldType; label: string; hasOptions?: boolean }[] = [
  { value: 'text', label: 'Text Input' },
  { value: 'number', label: 'Number Input' },
  { value: 'email', label: 'Email Input' },
  { value: 'password', label: 'Password Input' },
  { value: 'textarea', label: 'Text Area' },
  { value: 'select', label: 'Dropdown Select', hasOptions: true },
  { value: 'checkbox', label: 'Checkbox' },
  { value: 'radio', label: 'Radio Buttons', hasOptions: true },
  { value: 'date', label: 'Date Picker' },
  { value: 'time', label: 'Time Picker' },
  { value: 'datetime', label: 'Date & Time' },
  { value: 'file', label: 'File Upload' },
  { value: 'color', label: 'Color Picker' },
  { value: 'range', label: 'Range Slider' },
]

export function FormBuilderEditor({ config: rawConfig, value, onChange }: FormBuilderEditorProps & { value?: any }) {
  const config = rawConfig || value || {}
  const submitButtonText = config.submitButtonText || 'Submit'
  const resetButtonText = config.resetButtonText || 'Reset'
  const showResetButton = config.showResetButton ?? true
  const formLayout = config.formLayout || 'vertical'
  const fields = config.fields || []
  const outputTopic = config.outputTopic || ''
  const validateOnChange = config.validateOnChange ?? false

  const [draggedIndex, setDraggedIndex] = useState<number | null>(null)
  const [expandedField, setExpandedField] = useState<string | null>(null)

  const generateFieldId = (): string => {
    return `field_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
  }

  const handleAddField = (type: FormFieldType = 'text') => {
    const newField: FormField = {
      id: generateFieldId(),
      type,
      label: `Field ${fields.length + 1}`,
      name: `field${fields.length + 1}`,
      placeholder: '',
      required: false,
      disabled: false,
    }

    // Set type-specific defaults
    if (type === 'select' || type === 'radio') {
      newField.options = ['Option 1', 'Option 2', 'Option 3']
    }
    if (type === 'textarea') {
      newField.rows = 3
    }
    if (type === 'number' || type === 'range') {
      newField.min = 0
      newField.max = 100
      newField.step = 1
    }

    onChange({ ...config, fields: [...fields, newField] })
    setExpandedField(newField.id)
  }

  const handleRemoveField = (id: string) => {
    const newFields = fields.filter((f) => f.id !== id)
    onChange({ ...config, fields: newFields })
    if (expandedField === id) {
      setExpandedField(null)
    }
  }

  const handleDuplicateField = (field: FormField) => {
    const duplicatedField: FormField = {
      ...field,
      id: generateFieldId(),
      name: `${field.name}_copy`,
      label: `${field.label} (Copy)`,
    }
    onChange({ ...config, fields: [...fields, duplicatedField] })
  }

  const handleUpdateField = (id: string, updates: Partial<FormField>) => {
    const newFields = fields.map((f) => (f.id === id ? { ...f, ...updates } : f))
    onChange({ ...config, fields: newFields })
  }

  const handleDragStart = (index: number) => {
    setDraggedIndex(index)
  }

  const handleDragOver = (e: React.DragEvent, index: number) => {
    e.preventDefault()
    if (draggedIndex === null || draggedIndex === index) return

    const newFields = [...fields]
    const draggedItem = newFields[draggedIndex]
    newFields.splice(draggedIndex, 1)
    newFields.splice(index, 0, draggedItem)

    setDraggedIndex(index)
    onChange({ ...config, fields: newFields })
  }

  const handleDragEnd = () => {
    setDraggedIndex(null)
  }

  const toggleFieldExpanded = (id: string) => {
    setExpandedField(expandedField === id ? null : id)
  }

  const getFieldTypeConfig = (type: FormFieldType) => {
    return FIELD_TYPES.find((ft) => ft.value === type)
  }

  return (
    <div className="space-y-6">
      {/* Form Settings */}
      <div className="space-y-4">
        <h3 className="text-sm font-semibold">Form Settings</h3>

        <div className="grid grid-cols-2 gap-4">
          <div className="space-y-2">
            <Label htmlFor="submitButtonText" className="text-sm font-semibold">
              Submit Button Text
            </Label>
            <Input
              id="submitButtonText"
              value={submitButtonText}
              onChange={(e) => onChange({ ...config, submitButtonText: e.target.value })}
              placeholder="Submit"
              className="h-11"
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="resetButtonText" className="text-sm font-semibold">
              Reset Button Text
            </Label>
            <Input
              id="resetButtonText"
              value={resetButtonText}
              onChange={(e) => onChange({ ...config, resetButtonText: e.target.value })}
              placeholder="Reset"
              className="h-11"
              disabled={!showResetButton}
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
            placeholder="e.g., form/submit"
            className="h-11"
          />
          <p className="text-xs text-muted-foreground">
            Topic to send form data when submitted
          </p>
        </div>

        <div className="space-y-2">
          <Label className="text-sm font-semibold">Layout</Label>
          <Select
            value={formLayout}
            onValueChange={(value: 'vertical' | 'horizontal') =>
              onChange({ ...config, formLayout: value })
            }
          >
            <SelectTrigger className="h-11">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="vertical">Vertical (Label above field)</SelectItem>
              <SelectItem value="horizontal">Horizontal (Label beside field)</SelectItem>
            </SelectContent>
          </Select>
        </div>

        <div className="flex items-center space-x-2">
          <Switch
            id="showResetButton"
            checked={showResetButton}
            onCheckedChange={(checked) => onChange({ ...config, showResetButton: checked })}
          />
          <Label htmlFor="showResetButton" className="text-sm font-normal cursor-pointer">
            Show Reset Button
          </Label>
        </div>

        <div className="flex items-center space-x-2">
          <Switch
            id="validateOnChange"
            checked={validateOnChange}
            onCheckedChange={(checked) => onChange({ ...config, validateOnChange: checked })}
          />
          <Label htmlFor="validateOnChange" className="text-sm font-normal cursor-pointer">
            Validate Fields on Change
          </Label>
        </div>
      </div>

      {/* Field Builder */}
      <div className="space-y-3">
        <div className="flex items-center justify-between">
          <Label className="text-sm font-semibold">Form Fields</Label>
          <Select value="" onValueChange={(value: FormFieldType) => handleAddField(value)}>
            <SelectTrigger className="w-48 h-8">
              <Plus className="w-4 h-4 mr-1" />
              <SelectValue placeholder="Add Field" />
            </SelectTrigger>
            <SelectContent>
              {FIELD_TYPES.map((ft) => (
                <SelectItem key={ft.value} value={ft.value}>
                  {ft.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        {fields.length === 0 && (
          <div className="text-center py-8 border-2 border-dashed rounded-lg">
            <p className="text-sm text-muted-foreground">
              No fields defined. Add fields using the dropdown above.
            </p>
          </div>
        )}

        <div className="space-y-2">
          {fields.map((field, index) => {
            const isExpanded = expandedField === field.id
            const fieldTypeConfig = getFieldTypeConfig(field.type)

            return (
              <div
                key={field.id}
                draggable
                onDragStart={() => handleDragStart(index)}
                onDragOver={(e) => handleDragOver(e, index)}
                onDragEnd={handleDragEnd}
                className={cn(
                  'border rounded-lg bg-card transition-colors',
                  draggedIndex === index && 'opacity-50'
                )}
              >
                {/* Field Header */}
                <div
                  className="flex items-center gap-3 p-3 cursor-pointer hover:bg-muted/50"
                  onClick={() => toggleFieldExpanded(field.id)}
                >
                  <GripVertical className="w-4 h-4 text-muted-foreground flex-shrink-0 cursor-move" />

                  <div className="flex-1">
                    <div className="font-medium text-sm">{field.label}</div>
                    <div className="text-xs text-muted-foreground">
                      {fieldTypeConfig?.label} • {field.name}
                      {field.required && ' • Required'}
                    </div>
                  </div>

                  <div className="flex gap-1">
                    <Button
                      type="button"
                      variant="ghost"
                      size="sm"
                      onClick={(e) => {
                        e.stopPropagation()
                        handleDuplicateField(field)
                      }}
                      className="h-8 w-8 p-0"
                    >
                      <Copy className="w-4 h-4" />
                    </Button>
                    <Button
                      type="button"
                      variant="ghost"
                      size="sm"
                      onClick={(e) => {
                        e.stopPropagation()
                        handleRemoveField(field.id)
                      }}
                      className="h-8 w-8 p-0 text-red-600 hover:text-red-700 hover:bg-red-50"
                    >
                      <Trash2 className="w-4 h-4" />
                    </Button>
                  </div>
                </div>

                {/* Field Configuration (Expanded) */}
                {isExpanded && (
                  <div className="p-4 pt-0 space-y-3 border-t">
                    <div className="grid grid-cols-2 gap-3">
                      {/* Label */}
                      <div className="space-y-1">
                        <Label className="text-xs">Label</Label>
                        <Input
                          value={field.label}
                          onChange={(e) => handleUpdateField(field.id, { label: e.target.value })}
                          className="h-9"
                        />
                      </div>

                      {/* Name */}
                      <div className="space-y-1">
                        <Label className="text-xs">Field Name</Label>
                        <Input
                          value={field.name}
                          onChange={(e) => handleUpdateField(field.id, { name: e.target.value })}
                          className="h-9"
                        />
                      </div>
                    </div>

                    {/* Placeholder */}
                    {!['checkbox', 'radio', 'file', 'color', 'range'].includes(field.type) && (
                      <div className="space-y-1">
                        <Label className="text-xs">Placeholder</Label>
                        <Input
                          value={field.placeholder || ''}
                          onChange={(e) =>
                            handleUpdateField(field.id, { placeholder: e.target.value })
                          }
                          className="h-9"
                        />
                      </div>
                    )}

                    {/* Helper Text */}
                    <div className="space-y-1">
                      <Label className="text-xs">Helper Text</Label>
                      <Input
                        value={field.helperText || ''}
                        onChange={(e) =>
                          handleUpdateField(field.id, { helperText: e.target.value })
                        }
                        className="h-9"
                        placeholder="Optional help text"
                      />
                    </div>

                    {/* Options (for select/radio) */}
                    {fieldTypeConfig?.hasOptions && (
                      <div className="space-y-1">
                        <Label className="text-xs">Options (one per line)</Label>
                        <Textarea
                          value={(field.options || []).join('\n')}
                          onChange={(e) =>
                            handleUpdateField(field.id, {
                              options: e.target.value.split('\n').filter((o) => o.trim()),
                            })
                          }
                          rows={4}
                          className="font-mono text-xs"
                        />
                      </div>
                    )}

                    {/* Number/Range Config */}
                    {['number', 'range'].includes(field.type) && (
                      <div className="grid grid-cols-3 gap-2">
                        <div className="space-y-1">
                          <Label className="text-xs">Min</Label>
                          <Input
                            type="number"
                            value={field.min ?? 0}
                            onChange={(e) =>
                              handleUpdateField(field.id, { min: Number(e.target.value) })
                            }
                            className="h-9"
                          />
                        </div>
                        <div className="space-y-1">
                          <Label className="text-xs">Max</Label>
                          <Input
                            type="number"
                            value={field.max ?? 100}
                            onChange={(e) =>
                              handleUpdateField(field.id, { max: Number(e.target.value) })
                            }
                            className="h-9"
                          />
                        </div>
                        <div className="space-y-1">
                          <Label className="text-xs">Step</Label>
                          <Input
                            type="number"
                            value={field.step ?? 1}
                            onChange={(e) =>
                              handleUpdateField(field.id, { step: Number(e.target.value) })
                            }
                            className="h-9"
                          />
                        </div>
                      </div>
                    )}

                    {/* Textarea Rows */}
                    {field.type === 'textarea' && (
                      <div className="space-y-1">
                        <Label className="text-xs">Rows</Label>
                        <Input
                          type="number"
                          value={field.rows ?? 3}
                          onChange={(e) =>
                            handleUpdateField(field.id, { rows: Number(e.target.value) })
                          }
                          className="h-9"
                          min={1}
                          max={20}
                        />
                      </div>
                    )}

                    {/* Pattern (for text) */}
                    {field.type === 'text' && (
                      <div className="space-y-1">
                        <Label className="text-xs">Validation Pattern (Regex)</Label>
                        <Input
                          value={field.pattern || ''}
                          onChange={(e) => handleUpdateField(field.id, { pattern: e.target.value })}
                          className="h-9"
                          placeholder="e.g., ^[A-Za-z]+$"
                        />
                      </div>
                    )}

                    {/* Flags */}
                    <div className="flex gap-4">
                      <div className="flex items-center space-x-2">
                        <Switch
                          checked={field.required ?? false}
                          onCheckedChange={(checked) =>
                            handleUpdateField(field.id, { required: checked })
                          }
                        />
                        <Label className="text-xs cursor-pointer">Required</Label>
                      </div>
                      <div className="flex items-center space-x-2">
                        <Switch
                          checked={field.disabled ?? false}
                          onCheckedChange={(checked) =>
                            handleUpdateField(field.id, { disabled: checked })
                          }
                        />
                        <Label className="text-xs cursor-pointer">Disabled</Label>
                      </div>
                    </div>
                  </div>
                )}
              </div>
            )
          })}
        </div>

        {fields.length > 0 && (
          <p className="text-xs text-muted-foreground">
            Drag fields to reorder. Click a field to expand configuration options.
          </p>
        )}
      </div>

      {/* Preview Note */}
      <div className="p-4 bg-muted/50 rounded-lg border">
        <p className="text-sm font-semibold mb-1">Preview</p>
        <p className="text-xs text-muted-foreground">
          When the form is submitted, it will send a message with{' '}
          <code className="bg-background px-1 rounded">payload</code> containing all field values
          as key-value pairs. The message will be sent to the output topic if specified.
        </p>
      </div>
    </div>
  )
}
