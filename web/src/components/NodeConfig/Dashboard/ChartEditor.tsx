/**
 * Chart Widget Editor
 *
 * Specialized editor for dashboard chart widget with series management
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
import { ColorPicker } from '@/components/Common/ColorPicker'
import { Plus, Trash2, GripVertical } from 'lucide-react'
import { cn } from '@/lib/utils'

export type ChartType = 'line' | 'bar' | 'pie' | 'histogram' | 'scatter'

export interface ChartSeries {
  name: string
  color: string
  type?: ChartType
  visible?: boolean
}

interface ChartEditorProps {
  config: {
    chartType?: ChartType
    xAxisLabel?: string
    yAxisLabel?: string
    maxDataSize?: number
    legend?: boolean
    series?: ChartSeries[]
  }
  onChange: (config: any) => void
}

const CHART_TYPES: { value: ChartType; label: string; description: string }[] = [
  { value: 'line', label: 'Line', description: 'Line chart for time series data' },
  { value: 'bar', label: 'Bar', description: 'Bar chart for categorical data' },
  { value: 'pie', label: 'Pie', description: 'Pie chart for proportional data' },
  { value: 'histogram', label: 'Histogram', description: 'Distribution of numerical data' },
  { value: 'scatter', label: 'Scatter', description: 'XY scatter plot' },
]

const DEFAULT_COLORS = [
  '#3b82f6', // blue
  '#ef4444', // red
  '#10b981', // green
  '#f59e0b', // amber
  '#8b5cf6', // violet
  '#ec4899', // pink
  '#14b8a6', // teal
  '#f97316', // orange
]

export function ChartEditor({ config: rawConfig, value, onChange }: ChartEditorProps & { value?: any }) {
  const config = rawConfig || value || {}
  const chartType = config.chartType || 'line'
  const xAxisLabel = config.xAxisLabel || ''
  const yAxisLabel = config.yAxisLabel || ''
  const maxDataSize = config.maxDataSize || 100
  const legend = config.legend ?? true
  const series = config.series || []

  const [draggedIndex, setDraggedIndex] = useState<number | null>(null)

  const handleChartTypeChange = (type: ChartType) => {
    onChange({ ...config, chartType: type })
  }

  const handleAddSeries = () => {
    const newSeries: ChartSeries = {
      name: `Series ${series.length + 1}`,
      color: DEFAULT_COLORS[series.length % DEFAULT_COLORS.length],
      type: chartType,
      visible: true,
    }
    onChange({ ...config, series: [...series, newSeries] })
  }

  const handleRemoveSeries = (index: number) => {
    const newSeries = series.filter((_, i) => i !== index)
    onChange({ ...config, series: newSeries })
  }

  const handleUpdateSeries = (index: number, updates: Partial<ChartSeries>) => {
    const newSeries = series.map((s, i) => (i === index ? { ...s, ...updates } : s))
    onChange({ ...config, series: newSeries })
  }

  const handleDragStart = (index: number) => {
    setDraggedIndex(index)
  }

  const handleDragOver = (e: React.DragEvent, index: number) => {
    e.preventDefault()
    if (draggedIndex === null || draggedIndex === index) return

    const newSeries = [...series]
    const draggedItem = newSeries[draggedIndex]
    newSeries.splice(draggedIndex, 1)
    newSeries.splice(index, 0, draggedItem)

    setDraggedIndex(index)
    onChange({ ...config, series: newSeries })
  }

  const handleDragEnd = () => {
    setDraggedIndex(null)
  }

  const supportsMultipleSeries = ['line', 'bar', 'scatter'].includes(chartType)

  return (
    <div className="space-y-6">
      {/* Chart Type */}
      <div className="space-y-2">
        <Label className="text-sm font-semibold">Chart Type</Label>
        <Select value={chartType} onValueChange={handleChartTypeChange}>
          <SelectTrigger className="h-11">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {CHART_TYPES.map((type) => (
              <SelectItem key={type.value} value={type.value}>
                <div>
                  <div className="font-medium">{type.label}</div>
                  <div className="text-xs text-muted-foreground">{type.description}</div>
                </div>
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      {/* Axis Labels */}
      {!['pie'].includes(chartType) && (
        <div className="grid grid-cols-2 gap-4">
          <div className="space-y-2">
            <Label htmlFor="xAxisLabel" className="text-sm font-semibold">
              X-Axis Label
            </Label>
            <Input
              id="xAxisLabel"
              value={xAxisLabel}
              onChange={(e) => onChange({ ...config, xAxisLabel: e.target.value })}
              placeholder="e.g., Time"
              className="h-11"
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="yAxisLabel" className="text-sm font-semibold">
              Y-Axis Label
            </Label>
            <Input
              id="yAxisLabel"
              value={yAxisLabel}
              onChange={(e) => onChange({ ...config, yAxisLabel: e.target.value })}
              placeholder="e.g., Temperature (°C)"
              className="h-11"
            />
          </div>
        </div>
      )}

      {/* Max Data Points */}
      <div className="space-y-2">
        <Label htmlFor="maxDataSize" className="text-sm font-semibold">
          Max Data Points
        </Label>
        <Input
          id="maxDataSize"
          type="number"
          min={10}
          max={1000}
          value={maxDataSize}
          onChange={(e) => onChange({ ...config, maxDataSize: Number(e.target.value) })}
          className="h-11"
        />
        <p className="text-xs text-muted-foreground">
          Maximum number of data points to display (older points will be removed)
        </p>
      </div>

      {/* Show Legend */}
      <div className="flex items-center space-x-2">
        <Switch
          id="legend"
          checked={legend}
          onCheckedChange={(checked) => onChange({ ...config, legend: checked })}
        />
        <Label htmlFor="legend" className="text-sm font-normal cursor-pointer">
          Show Legend
        </Label>
      </div>

      {/* Series Configuration */}
      {supportsMultipleSeries && (
        <div className="space-y-3">
          <div className="flex items-center justify-between">
            <Label className="text-sm font-semibold">Data Series</Label>
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={handleAddSeries}
              className="h-8"
            >
              <Plus className="w-4 h-4 mr-1" />
              Add Series
            </Button>
          </div>

          {series.length === 0 && (
            <div className="text-center py-8 border-2 border-dashed rounded-lg">
              <p className="text-sm text-muted-foreground">
                No series defined. Add a series to display multiple data sets.
              </p>
            </div>
          )}

          <div className="space-y-2">
            {series.map((s, index) => (
              <div
                key={index}
                draggable
                onDragStart={() => handleDragStart(index)}
                onDragOver={(e) => handleDragOver(e, index)}
                onDragEnd={handleDragEnd}
                className={cn(
                  'flex items-center gap-3 p-3 border rounded-lg bg-card cursor-move hover:border-primary transition-colors',
                  draggedIndex === index && 'opacity-50'
                )}
              >
                <GripVertical className="w-4 h-4 text-muted-foreground flex-shrink-0" />

                {/* Series Name */}
                <Input
                  value={s.name}
                  onChange={(e) => handleUpdateSeries(index, { name: e.target.value })}
                  placeholder="Series name"
                  className="h-9 flex-1"
                />

                {/* Series Color */}
                <div className="w-32">
                  <ColorPicker
                    value={s.color}
                    onChange={(color) => handleUpdateSeries(index, { color })}
                  />
                </div>

                {/* Visibility Toggle */}
                <Switch
                  checked={s.visible ?? true}
                  onCheckedChange={(visible) => handleUpdateSeries(index, { visible })}
                />

                {/* Remove Button */}
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  onClick={() => handleRemoveSeries(index)}
                  className="h-9 w-9 p-0 text-red-600 hover:text-red-700 hover:bg-red-50"
                >
                  <Trash2 className="w-4 h-4" />
                </Button>
              </div>
            ))}
          </div>

          {series.length > 0 && (
            <p className="text-xs text-muted-foreground">
              Drag series to reorder. Toggle visibility or remove series as needed.
            </p>
          )}
        </div>
      )}

      {/* Preview Note */}
      <div className="p-4 bg-muted/50 rounded-lg border">
        <p className="text-sm font-semibold mb-1">Preview</p>
        <p className="text-xs text-muted-foreground">
          The chart will display data sent to this widget via messages. Send data with{' '}
          <code className="bg-background px-1 rounded">
            {'{'}value: 42{'}'}
          </code>{' '}
          for single values or{' '}
          <code className="bg-background px-1 rounded">
            {'{'}series: "Series 1", value: 42{'}'}
          </code>{' '}
          for multiple series.
        </p>
      </div>
    </div>
  )
}
