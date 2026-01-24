/**
 * Gauge Widget Editor
 *
 * Specialized editor for dashboard gauge widget with sector color configuration
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

export type GaugeType = 'full' | 'semi' | 'arc' | 'donut'

export interface GaugeSector {
  from: number
  to: number
  color: string
  label?: string
}

interface GaugeEditorProps {
  config: {
    gaugeType?: GaugeType
    min?: number
    max?: number
    units?: string
    showValue?: boolean
    showMinMax?: boolean
    showSectors?: boolean
    sectors?: GaugeSector[]
    needleColor?: string
    backgroundColor?: string
  }
  onChange: (config: any) => void
}

const GAUGE_TYPES: { value: GaugeType; label: string; description: string }[] = [
  { value: 'full', label: 'Full Circle', description: '360° circular gauge' },
  { value: 'semi', label: 'Semi Circle', description: '180° half-circle gauge' },
  { value: 'arc', label: 'Arc', description: 'Compact arc gauge' },
  { value: 'donut', label: 'Donut', description: 'Ring-style gauge' },
]

const DEFAULT_SECTOR_COLORS = [
  '#22c55e', // green (safe zone)
  '#f59e0b', // amber (warning zone)
  '#ef4444', // red (danger zone)
]

export function GaugeEditor({ config, onChange }: GaugeEditorProps) {
  const gaugeType = config.gaugeType || 'semi'
  const min = config.min ?? 0
  const max = config.max ?? 100
  const units = config.units || ''
  const showValue = config.showValue ?? true
  const showMinMax = config.showMinMax ?? true
  const showSectors = config.showSectors ?? true
  const sectors = config.sectors || []
  const needleColor = config.needleColor || '#3b82f6'
  const backgroundColor = config.backgroundColor || '#e5e7eb'

  const [draggedIndex, setDraggedIndex] = useState<number | null>(null)

  const handleGaugeTypeChange = (type: GaugeType) => {
    onChange({ ...config, gaugeType: type })
  }

  const handleAddSector = () => {
    // Calculate new sector range
    let from = min
    let to = max

    if (sectors.length > 0) {
      const lastSector = sectors[sectors.length - 1]
      from = lastSector.to
      to = max
    } else {
      // Create default 3-sector configuration
      const range = max - min
      const sectorSize = range / 3
      from = min + sectorSize * sectors.length
      to = Math.min(from + sectorSize, max)
    }

    const newSector: GaugeSector = {
      from,
      to,
      color: DEFAULT_SECTOR_COLORS[sectors.length % DEFAULT_SECTOR_COLORS.length],
      label: `Sector ${sectors.length + 1}`,
    }
    onChange({ ...config, sectors: [...sectors, newSector] })
  }

  const handleRemoveSector = (index: number) => {
    const newSectors = sectors.filter((_, i) => i !== index)
    onChange({ ...config, sectors: newSectors })
  }

  const handleUpdateSector = (index: number, updates: Partial<GaugeSector>) => {
    const newSectors = sectors.map((s, i) => (i === index ? { ...s, ...updates } : s))
    onChange({ ...config, sectors: newSectors })
  }

  const handleAutoDistribute = () => {
    if (sectors.length === 0) return

    const range = max - min
    const sectorSize = range / sectors.length

    const distributedSectors = sectors.map((sector, index) => ({
      ...sector,
      from: min + sectorSize * index,
      to: min + sectorSize * (index + 1),
    }))

    onChange({ ...config, sectors: distributedSectors })
  }

  const handleDragStart = (index: number) => {
    setDraggedIndex(index)
  }

  const handleDragOver = (e: React.DragEvent, index: number) => {
    e.preventDefault()
    if (draggedIndex === null || draggedIndex === index) return

    const newSectors = [...sectors]
    const draggedItem = newSectors[draggedIndex]
    newSectors.splice(draggedIndex, 1)
    newSectors.splice(index, 0, draggedItem)

    setDraggedIndex(index)
    onChange({ ...config, sectors: newSectors })
  }

  const handleDragEnd = () => {
    setDraggedIndex(null)
  }

  // Validate sector ranges
  const getSectorError = (sector: GaugeSector, index: number): string | null => {
    if (sector.from >= sector.to) {
      return 'From must be less than To'
    }
    if (sector.from < min || sector.to > max) {
      return `Range must be within ${min}-${max}`
    }
    // Check for overlaps with other sectors
    for (let i = 0; i < sectors.length; i++) {
      if (i === index) continue
      const other = sectors[i]
      if (
        (sector.from >= other.from && sector.from < other.to) ||
        (sector.to > other.from && sector.to <= other.to) ||
        (sector.from <= other.from && sector.to >= other.to)
      ) {
        return `Overlaps with Sector ${i + 1}`
      }
    }
    return null
  }

  return (
    <div className="space-y-6">
      {/* Gauge Type */}
      <div className="space-y-2">
        <Label className="text-sm font-semibold">Gauge Type</Label>
        <Select value={gaugeType} onValueChange={handleGaugeTypeChange}>
          <SelectTrigger className="h-11">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {GAUGE_TYPES.map((type) => (
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

      {/* Min/Max Values */}
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label htmlFor="min" className="text-sm font-semibold">
            Minimum Value
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
            Maximum Value
          </Label>
          <Input
            id="max"
            type="number"
            value={max}
            onChange={(e) => onChange({ ...config, max: Number(e.target.value) })}
            className="h-11"
          />
        </div>
      </div>

      {/* Units */}
      <div className="space-y-2">
        <Label htmlFor="units" className="text-sm font-semibold">
          Units
        </Label>
        <Input
          id="units"
          value={units}
          onChange={(e) => onChange({ ...config, units: e.target.value })}
          placeholder="e.g., °C, %, RPM"
          className="h-11"
        />
      </div>

      {/* Display Options */}
      <div className="space-y-3">
        <Label className="text-sm font-semibold">Display Options</Label>
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
            id="showMinMax"
            checked={showMinMax}
            onCheckedChange={(checked) => onChange({ ...config, showMinMax: checked })}
          />
          <Label htmlFor="showMinMax" className="text-sm font-normal cursor-pointer">
            Show Min/Max Labels
          </Label>
        </div>
        <div className="flex items-center space-x-2">
          <Switch
            id="showSectors"
            checked={showSectors}
            onCheckedChange={(checked) => onChange({ ...config, showSectors: checked })}
          />
          <Label htmlFor="showSectors" className="text-sm font-normal cursor-pointer">
            Show Color Sectors
          </Label>
        </div>
      </div>

      {/* Color Configuration */}
      <div className="grid grid-cols-2 gap-4">
        <div className="space-y-2">
          <Label className="text-sm font-semibold">Needle Color</Label>
          <ColorPicker
            value={needleColor}
            onChange={(color) => onChange({ ...config, needleColor: color })}
          />
        </div>
        <div className="space-y-2">
          <Label className="text-sm font-semibold">Background Color</Label>
          <ColorPicker
            value={backgroundColor}
            onChange={(color) => onChange({ ...config, backgroundColor: color })}
          />
        </div>
      </div>

      {/* Sector Configuration */}
      {showSectors && (
        <div className="space-y-3">
          <div className="flex items-center justify-between">
            <Label className="text-sm font-semibold">Color Sectors</Label>
            <div className="flex gap-2">
              {sectors.length > 1 && (
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={handleAutoDistribute}
                  className="h-8"
                >
                  Auto Distribute
                </Button>
              )}
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={handleAddSector}
                className="h-8"
              >
                <Plus className="w-4 h-4 mr-1" />
                Add Sector
              </Button>
            </div>
          </div>

          {sectors.length === 0 && (
            <div className="text-center py-8 border-2 border-dashed rounded-lg">
              <p className="text-sm text-muted-foreground">
                No sectors defined. Add sectors to create colored zones on the gauge.
              </p>
            </div>
          )}

          <div className="space-y-2">
            {sectors.map((sector, index) => {
              const error = getSectorError(sector, index)
              return (
                <div
                  key={index}
                  draggable
                  onDragStart={() => handleDragStart(index)}
                  onDragOver={(e) => handleDragOver(e, index)}
                  onDragEnd={handleDragEnd}
                  className={cn(
                    'flex items-start gap-3 p-3 border rounded-lg bg-card cursor-move hover:border-primary transition-colors',
                    draggedIndex === index && 'opacity-50',
                    error && 'border-red-500'
                  )}
                >
                  <GripVertical className="w-4 h-4 text-muted-foreground flex-shrink-0 mt-2" />

                  <div className="flex-1 space-y-3">
                    {/* Sector Label */}
                    <Input
                      value={sector.label || ''}
                      onChange={(e) => handleUpdateSector(index, { label: e.target.value })}
                      placeholder="Sector label (optional)"
                      className="h-9"
                    />

                    {/* Range Inputs */}
                    <div className="grid grid-cols-2 gap-2">
                      <div>
                        <Label className="text-xs text-muted-foreground">From</Label>
                        <Input
                          type="number"
                          value={sector.from}
                          onChange={(e) =>
                            handleUpdateSector(index, { from: Number(e.target.value) })
                          }
                          className="h-9"
                          min={min}
                          max={max}
                        />
                      </div>
                      <div>
                        <Label className="text-xs text-muted-foreground">To</Label>
                        <Input
                          type="number"
                          value={sector.to}
                          onChange={(e) =>
                            handleUpdateSector(index, { to: Number(e.target.value) })
                          }
                          className="h-9"
                          min={min}
                          max={max}
                        />
                      </div>
                    </div>

                    {/* Error Display */}
                    {error && <p className="text-xs text-red-500">{error}</p>}
                  </div>

                  {/* Sector Color */}
                  <div className="w-32">
                    <ColorPicker
                      value={sector.color}
                      onChange={(color) => handleUpdateSector(index, { color })}
                    />
                  </div>

                  {/* Remove Button */}
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    onClick={() => handleRemoveSector(index)}
                    className="h-9 w-9 p-0 text-red-600 hover:text-red-700 hover:bg-red-50"
                  >
                    <Trash2 className="w-4 h-4" />
                  </Button>
                </div>
              )
            })}
          </div>

          {sectors.length > 0 && (
            <p className="text-xs text-muted-foreground">
              Drag sectors to reorder. Use "Auto Distribute" to evenly divide the range. Sectors
              should not overlap.
            </p>
          )}
        </div>
      )}

      {/* Preview Note */}
      <div className="p-4 bg-muted/50 rounded-lg border">
        <p className="text-sm font-semibold mb-1">Preview</p>
        <p className="text-xs text-muted-foreground">
          The gauge will display values sent to this widget. Send data with{' '}
          <code className="bg-background px-1 rounded">
            {'{'}value: 75{'}'}
          </code>{' '}
          format. The needle will point to the value, and sectors will be colored according to the
          configuration.
        </p>
      </div>
    </div>
  )
}
