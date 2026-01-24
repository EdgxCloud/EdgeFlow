/**
 * Cron Expression Builder Component
 *
 * Visual cron expression builder with tabs and preview
 */

import { useState, useEffect, useMemo } from 'react'
import { Label } from '@/components/ui/label'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Switch } from '@/components/ui/switch'
import { Clock, Calendar, AlertCircle } from 'lucide-react'
import { cn } from '@/lib/utils'

interface CronBuilderProps {
  value: string
  onChange: (cron: string) => void
  label?: string
  enableSeconds?: boolean
  showPreview?: boolean
  disabled?: boolean
}

const DAYS_OF_WEEK = [
  { label: 'Mon', value: '1' },
  { label: 'Tue', value: '2' },
  { label: 'Wed', value: '3' },
  { label: 'Thu', value: '4' },
  { label: 'Fri', value: '5' },
  { label: 'Sat', value: '6' },
  { label: 'Sun', value: '0' },
]

const MONTHS = [
  'January',
  'February',
  'March',
  'April',
  'May',
  'June',
  'July',
  'August',
  'September',
  'October',
  'November',
  'December',
]

export function CronBuilder({
  value,
  onChange,
  label,
  enableSeconds = false,
  showPreview = true,
  disabled = false,
}: CronBuilderProps) {
  const [activeTab, setActiveTab] = useState('minutes')
  const [customCron, setCustomCron] = useState(value || '* * * * *')

  // Parse existing cron value
  useEffect(() => {
    if (value) {
      setCustomCron(value)
    }
  }, [value])

  const buildCronExpression = (config: {
    minutes?: string
    hours?: string
    dayOfMonth?: string
    month?: string
    dayOfWeek?: string
  }) => {
    const { minutes = '*', hours = '*', dayOfMonth = '*', month = '*', dayOfWeek = '*' } = config

    return enableSeconds
      ? `0 ${minutes} ${hours} ${dayOfMonth} ${month} ${dayOfWeek}`
      : `${minutes} ${hours} ${dayOfMonth} ${month} ${dayOfWeek}`
  }

  const handlePresetChange = (preset: string) => {
    onChange(preset)
    setCustomCron(preset)
  }

  // Minutes Tab - Every N minutes
  const MinutesTab = () => {
    const [interval, setInterval] = useState(1)

    return (
      <div className="space-y-4 p-4">
        <div className="space-y-2">
          <Label>Run every</Label>
          <div className="flex items-center gap-2">
            <Input
              type="number"
              min={1}
              max={59}
              value={interval}
              onChange={(e) => {
                const val = Number(e.target.value)
                setInterval(val)
                handlePresetChange(buildCronExpression({ minutes: `*/${val}` }))
              }}
              className="w-24"
            />
            <span className="text-sm">minute(s)</span>
          </div>
        </div>
        <p className="text-xs text-muted-foreground">
          Expression: <code className="bg-muted px-2 py-1 rounded">{customCron}</code>
        </p>
      </div>
    )
  }

  // Hourly Tab - At minute X
  const HourlyTab = () => {
    const [minute, setMinute] = useState(0)

    return (
      <div className="space-y-4 p-4">
        <div className="space-y-2">
          <Label>At minute</Label>
          <Input
            type="number"
            min={0}
            max={59}
            value={minute}
            onChange={(e) => {
              const val = Number(e.target.value)
              setMinute(val)
              handlePresetChange(buildCronExpression({ minutes: val.toString() }))
            }}
            className="w-24"
          />
        </div>
        <p className="text-xs text-muted-foreground">
          Expression: <code className="bg-muted px-2 py-1 rounded">{customCron}</code>
        </p>
      </div>
    )
  }

  // Daily Tab - At specific time
  const DailyTab = () => {
    const [hour, setHour] = useState(9)
    const [minute, setMinute] = useState(0)

    return (
      <div className="space-y-4 p-4">
        <div className="space-y-2">
          <Label>At time</Label>
          <div className="flex items-center gap-2">
            <Input
              type="number"
              min={0}
              max={23}
              value={hour}
              onChange={(e) => {
                const val = Number(e.target.value)
                setHour(val)
                handlePresetChange(
                  buildCronExpression({
                    minutes: minute.toString(),
                    hours: val.toString(),
                  })
                )
              }}
              className="w-20"
              placeholder="HH"
            />
            <span>:</span>
            <Input
              type="number"
              min={0}
              max={59}
              value={minute}
              onChange={(e) => {
                const val = Number(e.target.value)
                setMinute(val)
                handlePresetChange(
                  buildCronExpression({
                    minutes: val.toString(),
                    hours: hour.toString(),
                  })
                )
              }}
              className="w-20"
              placeholder="MM"
            />
          </div>
        </div>
        <p className="text-xs text-muted-foreground">
          Expression: <code className="bg-muted px-2 py-1 rounded">{customCron}</code>
        </p>
      </div>
    )
  }

  // Weekly Tab - Specific days + time
  const WeeklyTab = () => {
    const [hour, setHour] = useState(9)
    const [minute, setMinute] = useState(0)
    const [selectedDays, setSelectedDays] = useState<string[]>(['1'])

    const toggleDay = (day: string) => {
      const newDays = selectedDays.includes(day)
        ? selectedDays.filter((d) => d !== day)
        : [...selectedDays, day].sort()

      setSelectedDays(newDays)
      handlePresetChange(
        buildCronExpression({
          minutes: minute.toString(),
          hours: hour.toString(),
          dayOfWeek: newDays.join(','),
        })
      )
    }

    return (
      <div className="space-y-4 p-4">
        <div className="space-y-2">
          <Label>Days of week</Label>
          <div className="flex gap-2">
            {DAYS_OF_WEEK.map((day) => (
              <Button
                key={day.value}
                type="button"
                variant={selectedDays.includes(day.value) ? 'default' : 'outline'}
                size="sm"
                onClick={() => toggleDay(day.value)}
                className="w-12"
              >
                {day.label}
              </Button>
            ))}
          </div>
        </div>
        <div className="space-y-2">
          <Label>At time</Label>
          <div className="flex items-center gap-2">
            <Input
              type="number"
              min={0}
              max={23}
              value={hour}
              onChange={(e) => {
                const val = Number(e.target.value)
                setHour(val)
                handlePresetChange(
                  buildCronExpression({
                    minutes: minute.toString(),
                    hours: val.toString(),
                    dayOfWeek: selectedDays.join(','),
                  })
                )
              }}
              className="w-20"
            />
            <span>:</span>
            <Input
              type="number"
              min={0}
              max={59}
              value={minute}
              onChange={(e) => {
                const val = Number(e.target.value)
                setMinute(val)
                handlePresetChange(
                  buildCronExpression({
                    minutes: val.toString(),
                    hours: hour.toString(),
                    dayOfWeek: selectedDays.join(','),
                  })
                )
              }}
              className="w-20"
            />
          </div>
        </div>
        <p className="text-xs text-muted-foreground">
          Expression: <code className="bg-muted px-2 py-1 rounded">{customCron}</code>
        </p>
      </div>
    )
  }

  // Custom Tab - Direct input
  const CustomTab = () => {
    const [localCron, setLocalCron] = useState(customCron)
    const [isValid, setIsValid] = useState(true)

    const validateCron = (expr: string) => {
      const parts = expr.trim().split(/\s+/)
      const expected = enableSeconds ? 6 : 5
      return parts.length === expected
    }

    const handleChange = (value: string) => {
      setLocalCron(value)
      const valid = validateCron(value)
      setIsValid(valid)
      if (valid) {
        onChange(value)
        setCustomCron(value)
      }
    }

    return (
      <div className="space-y-4 p-4">
        <div className="space-y-2">
          <Label>Cron Expression</Label>
          <Input
            type="text"
            value={localCron}
            onChange={(e) => handleChange(e.target.value)}
            placeholder={enableSeconds ? '0 * * * * *' : '* * * * *'}
            className={cn('font-mono', !isValid && 'border-red-500')}
          />
          {!isValid && (
            <p className="text-xs text-red-500 flex items-center gap-1">
              <AlertCircle className="w-3 h-3" />
              Invalid cron expression (expected {enableSeconds ? 6 : 5} parts)
            </p>
          )}
        </div>
        <div className="p-3 bg-muted rounded-lg">
          <p className="text-xs font-semibold mb-2">Format:</p>
          <code className="text-xs">
            {enableSeconds ? (
              <>
                second minute hour day month dayOfWeek
                <br />
                (0-59) (0-59) (0-23) (1-31) (1-12) (0-6)
              </>
            ) : (
              <>
                minute hour day month dayOfWeek
                <br />
                (0-59) (0-23) (1-31) (1-12) (0-6)
              </>
            )}
          </code>
        </div>
      </div>
    )
  }

  // Preset buttons
  const presets = [
    { label: 'Every minute', value: enableSeconds ? '0 * * * * *' : '* * * * *' },
    { label: 'Every hour', value: enableSeconds ? '0 0 * * * *' : '0 * * * *' },
    { label: 'Daily at 9 AM', value: enableSeconds ? '0 0 9 * * *' : '0 9 * * *' },
    { label: 'Weekly (Monday 9 AM)', value: enableSeconds ? '0 0 9 * * 1' : '0 9 * * 1' },
  ]

  return (
    <div className="space-y-2">
      {label && <Label className="text-sm font-semibold">{label}</Label>}

      <div className="border rounded-lg overflow-hidden">
        {/* Preset Buttons */}
        <div className="p-3 bg-muted/50 border-b flex flex-wrap gap-2">
          <span className="text-xs font-semibold text-muted-foreground mr-2 self-center">
            Quick:
          </span>
          {presets.map((preset) => (
            <Button
              key={preset.label}
              type="button"
              variant="outline"
              size="sm"
              onClick={() => handlePresetChange(preset.value)}
              disabled={disabled}
              className="h-7 text-xs"
            >
              {preset.label}
            </Button>
          ))}
        </div>

        {/* Tabs */}
        <Tabs value={activeTab} onValueChange={setActiveTab}>
          <TabsList className="w-full rounded-none border-b grid grid-cols-5">
            <TabsTrigger value="minutes">Minutes</TabsTrigger>
            <TabsTrigger value="hourly">Hourly</TabsTrigger>
            <TabsTrigger value="daily">Daily</TabsTrigger>
            <TabsTrigger value="weekly">Weekly</TabsTrigger>
            <TabsTrigger value="custom">Custom</TabsTrigger>
          </TabsList>

          <TabsContent value="minutes" className="mt-0">
            <MinutesTab />
          </TabsContent>
          <TabsContent value="hourly" className="mt-0">
            <HourlyTab />
          </TabsContent>
          <TabsContent value="daily" className="mt-0">
            <DailyTab />
          </TabsContent>
          <TabsContent value="weekly" className="mt-0">
            <WeeklyTab />
          </TabsContent>
          <TabsContent value="custom" className="mt-0">
            <CustomTab />
          </TabsContent>
        </Tabs>

        {/* Current Expression */}
        <div className="p-3 bg-muted/30 border-t flex items-center justify-between">
          <div className="flex items-center gap-2">
            <Clock className="w-4 h-4 text-muted-foreground" />
            <span className="text-sm font-medium">Current:</span>
            <code className="text-sm bg-background px-2 py-1 rounded border">
              {customCron}
            </code>
          </div>
        </div>
      </div>

      {showPreview && (
        <p className="text-xs text-muted-foreground">
          {enableSeconds ? '6-part cron (with seconds)' : '5-part cron (minute precision)'}
        </p>
      )}
    </div>
  )
}
