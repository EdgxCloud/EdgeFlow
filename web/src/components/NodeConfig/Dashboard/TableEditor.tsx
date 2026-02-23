/**
 * Table Widget Editor
 *
 * Specialized editor for dashboard table widget with column management
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
import { Plus, Trash2, GripVertical } from 'lucide-react'
import { cn } from '@/lib/utils'

export type ColumnType = 'string' | 'number' | 'boolean' | 'date' | 'progress' | 'badge' | 'link'
export type ColumnAlign = 'left' | 'center' | 'right'

export interface TableColumn {
  id: string
  header: string
  key: string
  type: ColumnType
  align?: ColumnAlign
  width?: number // px or %
  sortable?: boolean
  filterable?: boolean
  formatter?: string // optional custom formatter function
  visible?: boolean
}

interface TableEditorProps {
  config: {
    maxRows?: number
    pagination?: boolean
    pageSize?: number
    striped?: boolean
    bordered?: boolean
    hoverable?: boolean
    compact?: boolean
    sortable?: boolean
    filterable?: boolean
    exportable?: boolean
    columns?: TableColumn[]
  }
  onChange: (config: any) => void
}

const COLUMN_TYPES: { value: ColumnType; label: string; description: string }[] = [
  { value: 'string', label: 'Text', description: 'Plain text display' },
  { value: 'number', label: 'Number', description: 'Numeric values with formatting' },
  { value: 'boolean', label: 'Boolean', description: 'Yes/No or True/False' },
  { value: 'date', label: 'Date/Time', description: 'Formatted date and time' },
  { value: 'progress', label: 'Progress Bar', description: 'Visual progress indicator' },
  { value: 'badge', label: 'Badge', description: 'Colored badge/tag' },
  { value: 'link', label: 'Link', description: 'Clickable hyperlink' },
]

export function TableEditor({ config: rawConfig, value, onChange }: TableEditorProps & { value?: any }) {
  const config = rawConfig || value || {}
  const maxRows = config.maxRows || 100
  const pagination = config.pagination ?? true
  const pageSize = config.pageSize || 10
  const striped = config.striped ?? true
  const bordered = config.bordered ?? true
  const hoverable = config.hoverable ?? true
  const compact = config.compact ?? false
  const sortable = config.sortable ?? true
  const filterable = config.filterable ?? true
  const exportable = config.exportable ?? false
  const columns = config.columns || []

  const [draggedIndex, setDraggedIndex] = useState<number | null>(null)

  const generateColumnId = (): string => {
    return `col_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
  }

  const handleAddColumn = () => {
    const newColumn: TableColumn = {
      id: generateColumnId(),
      header: `Column ${columns.length + 1}`,
      key: `col${columns.length + 1}`,
      type: 'string',
      align: 'left',
      sortable: true,
      filterable: true,
      visible: true,
    }
    onChange({ ...config, columns: [...columns, newColumn] })
  }

  const handleRemoveColumn = (id: string) => {
    const newColumns = columns.filter((c) => c.id !== id)
    onChange({ ...config, columns: newColumns })
  }

  const handleUpdateColumn = (id: string, updates: Partial<TableColumn>) => {
    const newColumns = columns.map((c) => (c.id === id ? { ...c, ...updates } : c))
    onChange({ ...config, columns: newColumns })
  }

  const handleDragStart = (index: number) => {
    setDraggedIndex(index)
  }

  const handleDragOver = (e: React.DragEvent, index: number) => {
    e.preventDefault()
    if (draggedIndex === null || draggedIndex === index) return

    const newColumns = [...columns]
    const draggedItem = newColumns[draggedIndex]
    newColumns.splice(draggedIndex, 1)
    newColumns.splice(index, 0, draggedItem)

    setDraggedIndex(index)
    onChange({ ...config, columns: newColumns })
  }

  const handleDragEnd = () => {
    setDraggedIndex(null)
  }

  return (
    <div className="space-y-6">
      {/* Table Settings */}
      <div className="space-y-4">
        <h3 className="text-sm font-semibold">Table Settings</h3>

        <div className="grid grid-cols-2 gap-4">
          <div className="space-y-2">
            <Label htmlFor="maxRows" className="text-sm font-semibold">
              Max Rows to Store
            </Label>
            <Input
              id="maxRows"
              type="number"
              min={1}
              max={10000}
              value={maxRows}
              onChange={(e) => onChange({ ...config, maxRows: Number(e.target.value) })}
              className="h-11"
            />
            <p className="text-xs text-muted-foreground">
              Maximum number of rows to keep in memory
            </p>
          </div>

          <div className="space-y-2">
            <Label htmlFor="pageSize" className="text-sm font-semibold">
              Page Size
            </Label>
            <Input
              id="pageSize"
              type="number"
              min={5}
              max={100}
              value={pageSize}
              onChange={(e) => onChange({ ...config, pageSize: Number(e.target.value) })}
              className="h-11"
              disabled={!pagination}
            />
            <p className="text-xs text-muted-foreground">Rows per page</p>
          </div>
        </div>

        <div className="grid grid-cols-2 gap-3">
          <div className="flex items-center space-x-2">
            <Switch
              id="pagination"
              checked={pagination}
              onCheckedChange={(checked) => onChange({ ...config, pagination: checked })}
            />
            <Label htmlFor="pagination" className="text-sm font-normal cursor-pointer">
              Enable Pagination
            </Label>
          </div>

          <div className="flex items-center space-x-2">
            <Switch
              id="sortable"
              checked={sortable}
              onCheckedChange={(checked) => onChange({ ...config, sortable: checked })}
            />
            <Label htmlFor="sortable" className="text-sm font-normal cursor-pointer">
              Sortable Columns
            </Label>
          </div>

          <div className="flex items-center space-x-2">
            <Switch
              id="filterable"
              checked={filterable}
              onCheckedChange={(checked) => onChange({ ...config, filterable: checked })}
            />
            <Label htmlFor="filterable" className="text-sm font-normal cursor-pointer">
              Filterable Columns
            </Label>
          </div>

          <div className="flex items-center space-x-2">
            <Switch
              id="exportable"
              checked={exportable}
              onCheckedChange={(checked) => onChange({ ...config, exportable: checked })}
            />
            <Label htmlFor="exportable" className="text-sm font-normal cursor-pointer">
              Export to CSV
            </Label>
          </div>

          <div className="flex items-center space-x-2">
            <Switch
              id="striped"
              checked={striped}
              onCheckedChange={(checked) => onChange({ ...config, striped: checked })}
            />
            <Label htmlFor="striped" className="text-sm font-normal cursor-pointer">
              Striped Rows
            </Label>
          </div>

          <div className="flex items-center space-x-2">
            <Switch
              id="bordered"
              checked={bordered}
              onCheckedChange={(checked) => onChange({ ...config, bordered: checked })}
            />
            <Label htmlFor="bordered" className="text-sm font-normal cursor-pointer">
              Bordered
            </Label>
          </div>

          <div className="flex items-center space-x-2">
            <Switch
              id="hoverable"
              checked={hoverable}
              onCheckedChange={(checked) => onChange({ ...config, hoverable: checked })}
            />
            <Label htmlFor="hoverable" className="text-sm font-normal cursor-pointer">
              Hover Effect
            </Label>
          </div>

          <div className="flex items-center space-x-2">
            <Switch
              id="compact"
              checked={compact}
              onCheckedChange={(checked) => onChange({ ...config, compact: checked })}
            />
            <Label htmlFor="compact" className="text-sm font-normal cursor-pointer">
              Compact Mode
            </Label>
          </div>
        </div>
      </div>

      {/* Column Configuration */}
      <div className="space-y-3">
        <div className="flex items-center justify-between">
          <Label className="text-sm font-semibold">Columns</Label>
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={handleAddColumn}
            className="h-8"
          >
            <Plus className="w-4 h-4 mr-1" />
            Add Column
          </Button>
        </div>

        {columns.length === 0 && (
          <div className="text-center py-8 border-2 border-dashed rounded-lg">
            <p className="text-sm text-muted-foreground">
              No columns defined. Add columns to configure the table structure.
            </p>
          </div>
        )}

        <div className="space-y-2">
          {columns.map((column, index) => (
            <div
              key={column.id}
              draggable
              onDragStart={() => handleDragStart(index)}
              onDragOver={(e) => handleDragOver(e, index)}
              onDragEnd={handleDragEnd}
              className={cn(
                'flex items-start gap-3 p-3 border rounded-lg bg-card cursor-move hover:border-primary transition-colors',
                draggedIndex === index && 'opacity-50'
              )}
            >
              <GripVertical className="w-4 h-4 text-muted-foreground flex-shrink-0 mt-2" />

              <div className="flex-1 space-y-3">
                {/* Header and Key */}
                <div className="grid grid-cols-2 gap-3">
                  <div className="space-y-1">
                    <Label className="text-xs">Column Header</Label>
                    <Input
                      value={column.header}
                      onChange={(e) => handleUpdateColumn(column.id, { header: e.target.value })}
                      placeholder="Display name"
                      className="h-9"
                    />
                  </div>
                  <div className="space-y-1">
                    <Label className="text-xs">Data Key</Label>
                    <Input
                      value={column.key}
                      onChange={(e) => handleUpdateColumn(column.id, { key: e.target.value })}
                      placeholder="field_name"
                      className="h-9"
                    />
                  </div>
                </div>

                {/* Type and Alignment */}
                <div className="grid grid-cols-3 gap-3">
                  <div className="space-y-1">
                    <Label className="text-xs">Type</Label>
                    <Select
                      value={column.type}
                      onValueChange={(value: ColumnType) =>
                        handleUpdateColumn(column.id, { type: value })
                      }
                    >
                      <SelectTrigger className="h-9">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        {COLUMN_TYPES.map((type) => (
                          <SelectItem key={type.value} value={type.value}>
                            {type.label}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </div>

                  <div className="space-y-1">
                    <Label className="text-xs">Alignment</Label>
                    <Select
                      value={column.align || 'left'}
                      onValueChange={(value: ColumnAlign) =>
                        handleUpdateColumn(column.id, { align: value })
                      }
                    >
                      <SelectTrigger className="h-9">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="left">Left</SelectItem>
                        <SelectItem value="center">Center</SelectItem>
                        <SelectItem value="right">Right</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>

                  <div className="space-y-1">
                    <Label className="text-xs">Width (px)</Label>
                    <Input
                      type="number"
                      value={column.width || ''}
                      onChange={(e) =>
                        handleUpdateColumn(column.id, {
                          width: e.target.value ? Number(e.target.value) : undefined,
                        })
                      }
                      placeholder="Auto"
                      className="h-9"
                      min={50}
                      max={1000}
                    />
                  </div>
                </div>

                {/* Options */}
                <div className="flex gap-4">
                  <div className="flex items-center space-x-2">
                    <Switch
                      checked={column.sortable ?? true}
                      onCheckedChange={(checked) =>
                        handleUpdateColumn(column.id, { sortable: checked })
                      }
                    />
                    <Label className="text-xs cursor-pointer">Sortable</Label>
                  </div>
                  <div className="flex items-center space-x-2">
                    <Switch
                      checked={column.filterable ?? true}
                      onCheckedChange={(checked) =>
                        handleUpdateColumn(column.id, { filterable: checked })
                      }
                    />
                    <Label className="text-xs cursor-pointer">Filterable</Label>
                  </div>
                  <div className="flex items-center space-x-2">
                    <Switch
                      checked={column.visible ?? true}
                      onCheckedChange={(checked) =>
                        handleUpdateColumn(column.id, { visible: checked })
                      }
                    />
                    <Label className="text-xs cursor-pointer">Visible</Label>
                  </div>
                </div>
              </div>

              {/* Remove Button */}
              <Button
                type="button"
                variant="ghost"
                size="sm"
                onClick={() => handleRemoveColumn(column.id)}
                className="h-9 w-9 p-0 text-red-600 hover:text-red-700 hover:bg-red-50"
              >
                <Trash2 className="w-4 h-4" />
              </Button>
            </div>
          ))}
        </div>

        {columns.length > 0 && (
          <p className="text-xs text-muted-foreground">
            Drag columns to reorder. Column order determines display order in the table.
          </p>
        )}
      </div>

      {/* Preview Note */}
      <div className="p-4 bg-muted/50 rounded-lg border">
        <p className="text-sm font-semibold mb-1">Data Format</p>
        <p className="text-xs text-muted-foreground mb-2">
          Send table data as an array of objects where each key matches a column's "Data Key":
        </p>
        <pre className="text-xs bg-background p-3 rounded border overflow-x-auto">
          {JSON.stringify(
            {
              payload: [
                {
                  [columns[0]?.key || 'id']: 1,
                  [columns[1]?.key || 'name']: 'Example',
                  [columns[2]?.key || 'value']: 42,
                },
              ],
            },
            null,
            2
          )}
        </pre>
        <p className="text-xs text-muted-foreground mt-2">
          New data will be appended to the table. Send{' '}
          <code className="bg-background px-1 rounded">{'{'}clear: true{'}'}</code> to clear the
          table first.
        </p>
      </div>
    </div>
  )
}
