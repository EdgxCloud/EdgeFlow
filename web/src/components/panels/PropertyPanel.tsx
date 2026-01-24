import { useState } from 'react'
import { X } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { ScrollArea } from '@/components/ui/scroll-area'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { Separator } from '@/components/ui/separator'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { cn } from '@/lib/utils'

interface PropertyPanelProps {
  selectedNode?: any
  onClose: () => void
  onUpdate: (nodeId: string, data: any) => void
  className?: string
}

export default function PropertyPanel({
  selectedNode,
  onClose,
  onUpdate,
  className,
}: PropertyPanelProps) {
  const [formData, setFormData] = useState(selectedNode?.data || {})

  if (!selectedNode) {
    return (
      <div className={cn('bg-card border-l border-border p-6', className)}>
        <div className="flex flex-col items-center justify-center h-full text-center text-muted-foreground">
          <p className="text-sm">No node selected</p>
          <p className="text-xs mt-1">Select a node to view properties</p>
        </div>
      </div>
    )
  }

  const handleUpdate = () => {
    onUpdate(selectedNode.id, formData)
  }

  const handleFieldChange = (field: string, value: any) => {
    setFormData((prev: any) => ({ ...prev, [field]: value }))
  }

  return (
    <div className={cn('flex flex-col bg-card border-l border-border', className)}>
      {/* Header */}
      <div className="flex items-center justify-between p-4 border-b border-border">
        <div className="flex-1">
          <h3 className="font-semibold">{selectedNode.data?.label || 'Node Properties'}</h3>
          <p className="text-xs text-muted-foreground mt-0.5">
            {selectedNode.data?.nodeType || 'Unknown'}
          </p>
        </div>
        <Button variant="ghost" size="icon" onClick={onClose}>
          <X className="h-4 w-4" />
        </Button>
      </div>

      {/* Content */}
      <ScrollArea className="flex-1">
        <Tabs defaultValue="properties" className="w-full">
          <TabsList className="w-full justify-start rounded-none border-b border-border bg-transparent p-0">
            <TabsTrigger
              value="properties"
              className="rounded-none border-b-2 border-transparent data-[state=active]:border-primary"
            >
              Properties
            </TabsTrigger>
            <TabsTrigger
              value="appearance"
              className="rounded-none border-b-2 border-transparent data-[state=active]:border-primary"
            >
              Appearance
            </TabsTrigger>
          </TabsList>

          <TabsContent value="properties" className="p-4 space-y-4">
            {/* Common Properties */}
            <div className="space-y-2">
              <Label htmlFor="name">Name</Label>
              <Input
                id="name"
                value={formData.label || ''}
                onChange={(e) => handleFieldChange('label', e.target.value)}
                placeholder="Node name"
              />
            </div>

            <Separator />

            {/* Node-specific properties based on type */}
            {renderNodeSpecificFields(selectedNode.data?.nodeType, formData, handleFieldChange)}

            {/* Description */}
            <div className="space-y-2">
              <Label htmlFor="description">Description</Label>
              <Textarea
                id="description"
                value={formData.info || ''}
                onChange={(e) => handleFieldChange('info', e.target.value)}
                placeholder="Node description"
                rows={3}
              />
            </div>
          </TabsContent>

          <TabsContent value="appearance" className="p-4 space-y-4">
            <div className="space-y-2">
              <Label htmlFor="subtitle">Subtitle</Label>
              <Input
                id="subtitle"
                value={formData.subtitle || ''}
                onChange={(e) => handleFieldChange('subtitle', e.target.value)}
                placeholder="Displayed below node name"
              />
            </div>

            <div className="flex items-center justify-between">
              <Label htmlFor="has-inputs">Show Input Port</Label>
              <Switch
                id="has-inputs"
                checked={formData.hasInputs !== false}
                onCheckedChange={(checked) => handleFieldChange('hasInputs', checked)}
              />
            </div>

            <div className="flex items-center justify-between">
              <Label htmlFor="has-outputs">Show Output Port</Label>
              <Switch
                id="has-outputs"
                checked={formData.hasOutputs !== false}
                onCheckedChange={(checked) => handleFieldChange('hasOutputs', checked)}
              />
            </div>
          </TabsContent>
        </Tabs>
      </ScrollArea>

      {/* Footer */}
      <div className="flex items-center justify-end gap-2 p-4 border-t border-border">
        <Button variant="outline" size="sm" onClick={onClose}>
          Cancel
        </Button>
        <Button size="sm" onClick={handleUpdate}>
          Apply
        </Button>
      </div>
    </div>
  )
}

// Render node-specific configuration fields
function renderNodeSpecificFields(
  nodeType: string,
  formData: any,
  onChange: (field: string, value: any) => void
) {
  switch (nodeType) {
    case 'inject':
      return (
        <>
          <div className="space-y-2">
            <Label>Repeat</Label>
            <Select
              value={formData.repeat || 'none'}
              onValueChange={(value) => onChange('repeat', value)}
            >
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="none">None</SelectItem>
                <SelectItem value="interval">Interval</SelectItem>
                <SelectItem value="cron">Cron</SelectItem>
              </SelectContent>
            </Select>
          </div>
          {formData.repeat === 'interval' && (
            <div className="space-y-2">
              <Label>Interval (seconds)</Label>
              <Input
                type="number"
                value={formData.interval || 1}
                onChange={(e) => onChange('interval', parseInt(e.target.value))}
              />
            </div>
          )}
        </>
      )

    case 'gpio-out':
      return (
        <>
          <div className="space-y-2">
            <Label>Pin Number</Label>
            <Input
              type="number"
              value={formData.pin || ''}
              onChange={(e) => onChange('pin', parseInt(e.target.value))}
              placeholder="e.g., 17"
            />
          </div>
          <div className="space-y-2">
            <Label>Initial State</Label>
            <Select
              value={formData.initialState || 'low'}
              onValueChange={(value) => onChange('initialState', value)}
            >
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="low">Low (0)</SelectItem>
                <SelectItem value="high">High (1)</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </>
      )

    case 'http-request':
      return (
        <>
          <div className="space-y-2">
            <Label>Method</Label>
            <Select
              value={formData.method || 'GET'}
              onValueChange={(value) => onChange('method', value)}
            >
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="GET">GET</SelectItem>
                <SelectItem value="POST">POST</SelectItem>
                <SelectItem value="PUT">PUT</SelectItem>
                <SelectItem value="DELETE">DELETE</SelectItem>
                <SelectItem value="PATCH">PATCH</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-2">
            <Label>URL</Label>
            <Input
              value={formData.url || ''}
              onChange={(e) => onChange('url', e.target.value)}
              placeholder="https://api.example.com"
            />
          </div>
        </>
      )

    case 'mqtt-in':
    case 'mqtt-out':
      return (
        <>
          <div className="space-y-2">
            <Label>Broker</Label>
            <Input
              value={formData.broker || ''}
              onChange={(e) => onChange('broker', e.target.value)}
              placeholder="mqtt://localhost:1883"
            />
          </div>
          <div className="space-y-2">
            <Label>Topic</Label>
            <Input
              value={formData.topic || ''}
              onChange={(e) => onChange('topic', e.target.value)}
              placeholder="sensors/temperature"
            />
          </div>
        </>
      )

    case 'delay':
      return (
        <div className="space-y-2">
          <Label>Delay (ms)</Label>
          <Input
            type="number"
            value={formData.delay || 1000}
            onChange={(e) => onChange('delay', parseInt(e.target.value))}
          />
        </div>
      )

    case 'function':
      return (
        <div className="space-y-2">
          <Label>JavaScript Code</Label>
          <Textarea
            value={formData.code || 'return msg;'}
            onChange={(e) => onChange('code', e.target.value)}
            placeholder="// JavaScript code here\nreturn msg;"
            rows={12}
            className="font-mono text-sm"
          />
          <p className="text-xs text-muted-foreground">
            Available: msg, context, flow, global, node
          </p>
        </div>
      )

    case 'python':
      return (
        <div className="space-y-2">
          <Label>Python Code</Label>
          <Textarea
            value={formData.code || 'return msg'}
            onChange={(e) => onChange('code', e.target.value)}
            placeholder="# Python code here\nreturn msg"
            rows={12}
            className="font-mono text-sm"
          />
          <p className="text-xs text-muted-foreground">
            Available: msg, context, flow, global, node
          </p>
        </div>
      )

    case 'exec':
      return (
        <>
          <div className="space-y-2">
            <Label>Command</Label>
            <Input
              value={formData.command || ''}
              onChange={(e) => onChange('command', e.target.value)}
              placeholder="ls -la"
            />
          </div>
          <div className="flex items-center justify-between">
            <Label htmlFor="append-payload">Append msg.payload</Label>
            <Switch
              id="append-payload"
              checked={formData.appendPayload !== false}
              onCheckedChange={(checked) => onChange('appendPayload', checked)}
            />
          </div>
          <div className="space-y-2">
            <Label>Timeout (seconds)</Label>
            <Input
              type="number"
              value={formData.timeout || 10}
              onChange={(e) => onChange('timeout', parseInt(e.target.value))}
            />
          </div>
        </>
      )

    case 'template':
      return (
        <>
          <div className="space-y-2">
            <Label>Template Syntax</Label>
            <Select
              value={formData.syntax || 'mustache'}
              onValueChange={(value) => onChange('syntax', value)}
            >
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="mustache">Mustache</SelectItem>
                <SelectItem value="plain">Plain Text</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-2">
            <Label>Template</Label>
            <Textarea
              value={formData.template || ''}
              onChange={(e) => onChange('template', e.target.value)}
              placeholder="Hello {{name}}!"
              rows={8}
              className="font-mono text-sm"
            />
          </div>
        </>
      )

    default:
      return (
        <p className="text-sm text-muted-foreground">
          No specific configuration available for this node type.
        </p>
      )
  }
}
