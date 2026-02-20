/**
 * Icon Picker Component
 *
 * Icon selector with search and categories using lucide-react icons
 */

import { useState, useMemo } from 'react'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Search, Image as ImageIcon } from 'lucide-react'
import * as LucideIcons from 'lucide-react'
import { cn } from '@/lib/utils'

interface IconPickerProps {
  value: string
  onChange: (icon: string) => void
  label?: string
  disabled?: boolean
}

// Icon categories
const ICON_CATEGORIES = {
  all: 'All Icons',
  arrows: 'Arrows',
  ui: 'UI Elements',
  devices: 'Devices',
  files: 'Files',
  communication: 'Communication',
  weather: 'Weather',
  media: 'Media',
  misc: 'Miscellaneous',
}

const RECENT_ICONS_KEY = 'iconpicker_recent'
const MAX_RECENT_ICONS = 12

// Get all lucide icon names
const getAllIcons = (): string[] => {
  return Object.keys(LucideIcons).filter(
    (key) => key !== 'createLucideIcon' && key !== 'Icon' && typeof LucideIcons[key as keyof typeof LucideIcons] === 'function'
  )
}

// Simple categorization based on icon name patterns
const categorizeIcon = (iconName: string): string[] => {
  const name = iconName.toLowerCase()
  const categories: string[] = ['all']

  if (name.includes('arrow') || name.includes('chevron') || name.includes('move')) {
    categories.push('arrows')
  }
  if (name.includes('button') || name.includes('menu') || name.includes('panel') || name.includes('toggle')) {
    categories.push('ui')
  }
  if (name.includes('phone') || name.includes('mobile') || name.includes('computer') || name.includes('monitor') || name.includes('device')) {
    categories.push('devices')
  }
  if (name.includes('file') || name.includes('folder') || name.includes('document')) {
    categories.push('files')
  }
  if (name.includes('mail') || name.includes('message') || name.includes('chat') || name.includes('send')) {
    categories.push('communication')
  }
  if (name.includes('cloud') || name.includes('sun') || name.includes('rain') || name.includes('snow') || name.includes('wind')) {
    categories.push('weather')
  }
  if (name.includes('play') || name.includes('pause') || name.includes('music') || name.includes('video') || name.includes('image')) {
    categories.push('media')
  }

  if (categories.length === 1) {
    categories.push('misc')
  }

  return categories
}

export function IconPicker({ value, onChange, label, disabled = false }: IconPickerProps) {
  const [search, setSearch] = useState('')
  const [activeCategory, setActiveCategory] = useState('all')
  const [recentIcons, setRecentIcons] = useState<string[]>(() => {
    try {
      const stored = localStorage.getItem(RECENT_ICONS_KEY)
      return stored ? JSON.parse(stored) : []
    } catch {
      return []
    }
  })

  const allIcons = useMemo(() => getAllIcons(), [])

  const filteredIcons = useMemo(() => {
    let icons = allIcons

    // Filter by search
    if (search) {
      const searchLower = search.toLowerCase()
      icons = icons.filter((icon) => icon.toLowerCase().includes(searchLower))
    }

    // Filter by category
    if (activeCategory !== 'all') {
      icons = icons.filter((icon) => categorizeIcon(icon).includes(activeCategory))
    }

    return icons
  }, [allIcons, search, activeCategory])

  const handleIconSelect = (iconName: string) => {
    onChange(iconName)
    addToRecent(iconName)
  }

  const addToRecent = (iconName: string) => {
    const updated = [iconName, ...recentIcons.filter((i) => i !== iconName)].slice(
      0,
      MAX_RECENT_ICONS
    )
    setRecentIcons(updated)
    try {
      localStorage.setItem(RECENT_ICONS_KEY, JSON.stringify(updated))
    } catch {
      // Ignore localStorage errors
    }
  }

  const iconLookup = value ? LucideIcons[value as keyof typeof LucideIcons] : null
  const IconComponent: React.ComponentType<{ className?: string }> = (iconLookup && typeof iconLookup === 'object')
    ? iconLookup as unknown as React.ComponentType<{ className?: string }>
    : ImageIcon

  return (
    <div className="space-y-2">
      {label && <Label className="text-sm font-semibold">{label}</Label>}

      <Popover>
        <PopoverTrigger asChild>
          <Button
            variant="outline"
            className="w-full h-11 justify-start gap-3"
            disabled={disabled}
          >
            <IconComponent className="w-5 h-5" />
            <span className="font-mono text-sm">{value || 'Select icon'}</span>
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-96" align="start">
          <div className="space-y-4">
            {/* Search */}
            <div className="relative">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
              <Input
                type="text"
                placeholder="Search icons..."
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                className="pl-9"
              />
            </div>

            {/* Recent Icons */}
            {recentIcons.length > 0 && !search && (
              <div>
                <Label className="text-xs font-semibold mb-2 block">Recent Icons</Label>
                <div className="grid grid-cols-6 gap-2">
                  {recentIcons.map((iconName) => {
                    const IconComp = LucideIcons[iconName as keyof typeof LucideIcons] as React.ComponentType<{ className?: string }> | undefined
                    if (!IconComp || typeof IconComp !== 'object') return null
                    return (
                      <button
                        key={iconName}
                        onClick={() => handleIconSelect(iconName)}
                        className={cn(
                          'p-2 rounded border-2 transition-all hover:bg-accent',
                          value === iconName
                            ? 'border-primary bg-primary/10'
                            : 'border-transparent'
                        )}
                        title={iconName}
                      >
                        <IconComp className="w-5 h-5 mx-auto" />
                      </button>
                    )
                  })}
                </div>
              </div>
            )}

            {/* Categories and Icons */}
            <Tabs value={activeCategory} onValueChange={setActiveCategory}>
              <TabsList className="w-full grid grid-cols-3 h-auto">
                <TabsTrigger value="all" className="text-xs">All</TabsTrigger>
                <TabsTrigger value="ui" className="text-xs">UI</TabsTrigger>
                <TabsTrigger value="devices" className="text-xs">Devices</TabsTrigger>
              </TabsList>

              <ScrollArea className="h-[300px] mt-4">
                <TabsContent value={activeCategory} className="mt-0">
                  <div className="grid grid-cols-6 gap-2 pr-4">
                    {filteredIcons.map((iconName) => {
                      const IconComp = LucideIcons[iconName as keyof typeof LucideIcons] as React.ComponentType<{ className?: string }> | undefined
                      if (!IconComp || typeof IconComp !== 'object') return null
                      return (
                        <button
                          key={iconName}
                          onClick={() => handleIconSelect(iconName)}
                          className={cn(
                            'p-2 rounded border-2 transition-all hover:bg-accent',
                            value === iconName
                              ? 'border-primary bg-primary/10'
                              : 'border-transparent'
                          )}
                          title={iconName}
                        >
                          <IconComp className="w-5 h-5 mx-auto" />
                        </button>
                      )
                    })}
                  </div>

                  {filteredIcons.length === 0 && (
                    <div className="text-center py-8 text-sm text-muted-foreground">
                      No icons found
                    </div>
                  )}
                </TabsContent>
              </ScrollArea>
            </Tabs>

            <div className="pt-2 border-t text-xs text-muted-foreground">
              {filteredIcons.length} icon{filteredIcons.length !== 1 ? 's' : ''} available
            </div>
          </div>
        </PopoverContent>
      </Popover>
    </div>
  )
}
