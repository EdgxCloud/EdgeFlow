/**
 * Color Picker Component
 *
 * Reusable color picker with preset colors and custom input
 */

import { useState } from 'react'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'

interface ColorPickerProps {
  value: string
  onChange: (color: string) => void
  label?: string
  allowAlpha?: boolean
  disabled?: boolean
}

const PRESET_COLORS = [
  '#ef4444', // red
  '#f97316', // orange
  '#f59e0b', // amber
  '#eab308', // yellow
  '#84cc16', // lime
  '#22c55e', // green
  '#10b981', // emerald
  '#14b8a6', // teal
  '#06b6d4', // cyan
  '#0ea5e9', // sky
  '#3b82f6', // blue
  '#6366f1', // indigo
  '#8b5cf6', // violet
  '#a855f7', // purple
  '#d946ef', // fuchsia
  '#ec4899', // pink
  '#f43f5e', // rose
  '#64748b', // slate
  '#6b7280', // gray
  '#000000', // black
  '#ffffff', // white
]

const RECENT_COLORS_KEY = 'colorpicker_recent'
const MAX_RECENT_COLORS = 8

export function ColorPicker({
  value,
  onChange,
  label,
  allowAlpha = false,
  disabled = false,
}: ColorPickerProps) {
  const [customColor, setCustomColor] = useState(value || '#000000')
  const [recentColors, setRecentColors] = useState<string[]>(() => {
    try {
      const stored = localStorage.getItem(RECENT_COLORS_KEY)
      return stored ? JSON.parse(stored) : []
    } catch {
      return []
    }
  })

  const handleColorSelect = (color: string) => {
    onChange(color)
    setCustomColor(color)
    addToRecent(color)
  }

  const addToRecent = (color: string) => {
    const updated = [color, ...recentColors.filter((c) => c !== color)].slice(
      0,
      MAX_RECENT_COLORS
    )
    setRecentColors(updated)
    try {
      localStorage.setItem(RECENT_COLORS_KEY, JSON.stringify(updated))
    } catch {
      // Ignore localStorage errors
    }
  }

  const handleCustomChange = (color: string) => {
    setCustomColor(color)
    // Validate hex color
    if (/^#([0-9A-F]{3}){1,2}$/i.test(color) || (allowAlpha && /^#([0-9A-F]{4}){1,2}$/i.test(color))) {
      onChange(color)
      addToRecent(color)
    }
  }

  return (
    <div className="space-y-2">
      {label && (
        <Label className="text-sm font-semibold">{label}</Label>
      )}

      <Popover>
        <PopoverTrigger asChild>
          <Button
            variant="outline"
            className="w-full h-11 justify-start gap-3"
            disabled={disabled}
          >
            <div
              className="w-6 h-6 rounded border-2 border-white shadow-sm"
              style={{ backgroundColor: value || '#000000' }}
            />
            <span className="font-mono text-sm">{value || '#000000'}</span>
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-80" align="start">
          <div className="space-y-4">
            {/* Preset Colors */}
            <div>
              <Label className="text-xs font-semibold mb-2 block">Preset Colors</Label>
              <div className="grid grid-cols-7 gap-2">
                {PRESET_COLORS.map((color) => (
                  <button
                    key={color}
                    onClick={() => handleColorSelect(color)}
                    className={cn(
                      'w-8 h-8 rounded border-2 transition-all hover:scale-110',
                      value === color ? 'border-primary ring-2 ring-primary' : 'border-gray-300'
                    )}
                    style={{ backgroundColor: color }}
                    title={color}
                  />
                ))}
              </div>
            </div>

            {/* Recent Colors */}
            {recentColors.length > 0 && (
              <div>
                <Label className="text-xs font-semibold mb-2 block">Recent Colors</Label>
                <div className="grid grid-cols-8 gap-2">
                  {recentColors.map((color, index) => (
                    <button
                      key={`${color}-${index}`}
                      onClick={() => handleColorSelect(color)}
                      className={cn(
                        'w-8 h-8 rounded border-2 transition-all hover:scale-110',
                        value === color ? 'border-primary ring-2 ring-primary' : 'border-gray-300'
                      )}
                      style={{ backgroundColor: color }}
                      title={color}
                    />
                  ))}
                </div>
              </div>
            )}

            {/* Custom Color Input */}
            <div className="space-y-2">
              <Label htmlFor="custom-color" className="text-xs font-semibold">
                Custom Color
              </Label>
              <div className="flex gap-2">
                <Input
                  id="custom-color"
                  type="text"
                  value={customColor}
                  onChange={(e) => handleCustomChange(e.target.value.toUpperCase())}
                  placeholder={allowAlpha ? '#RRGGBBAA' : '#RRGGBB'}
                  className="font-mono text-sm"
                  maxLength={allowAlpha ? 9 : 7}
                />
                <Input
                  type="color"
                  value={customColor.substring(0, 7)}
                  onChange={(e) => handleColorSelect(e.target.value.toUpperCase())}
                  className="w-16 h-11 p-1 cursor-pointer"
                />
              </div>
              <p className="text-xs text-muted-foreground">
                {allowAlpha
                  ? 'Enter hex color with optional alpha (#RRGGBB or #RRGGBBAA)'
                  : 'Enter hex color (#RRGGBB)'}
              </p>
            </div>
          </div>
        </PopoverContent>
      </Popover>
    </div>
  )
}
