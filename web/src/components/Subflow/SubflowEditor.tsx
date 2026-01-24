/**
 * Subflow Editor
 *
 * Visual editor for creating and editing subflows
 */

import { useState } from 'react'
import { ArrowLeft, Save, Plus, Settings } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Badge } from '@/components/ui/badge'
import { useToast } from '@/hooks/use-toast'

interface SubflowEditorProps {
  subflow: any
  onClose: () => void
}

export function SubflowEditor({ subflow: initialSubflow, onClose }: SubflowEditorProps) {
  const [subflow, setSubflow] = useState(initialSubflow)
  const [showPropertiesDialog, setShowPropertiesDialog] = useState(false)
  const [isSaving, setIsSaving] = useState(false)
  const { toast } = useToast()

  const handleSave = async () => {
    setIsSaving(true)
    try {
      const response = await fetch(`/api/subflows/${subflow.id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(subflow),
      })

      if (response.ok) {
        toast({
          title: 'Success',
          description: 'Subflow saved successfully',
        })
      } else {
        throw new Error('Failed to save')
      }
    } catch (error) {
      toast({
        title: 'Error',
        description: 'Failed to save subflow',
        variant: 'destructive',
      })
    } finally {
      setIsSaving(false)
    }
  }

  const addInputPort = () => {
    const newIndex = subflow.in?.length || 0
    setSubflow({
      ...subflow,
      in: [
        ...(subflow.in || []),
        {
          type: 'input',
          index: newIndex,
          name: `Input ${newIndex + 1}`,
        },
      ],
    })
  }

  const addOutputPort = () => {
    const newIndex = subflow.out?.length || 0
    setSubflow({
      ...subflow,
      out: [
        ...(subflow.out || []),
        {
          type: 'output',
          index: newIndex,
          name: `Output ${newIndex + 1}`,
        },
      ],
    })
  }

  const removePort = (type: 'in' | 'out', index: number) => {
    setSubflow({
      ...subflow,
      [type]: subflow[type].filter((_: any, i: number) => i !== index),
    })
  }

  const updatePort = (type: 'in' | 'out', index: number, updates: any) => {
    setSubflow({
      ...subflow,
      [type]: subflow[type].map((port: any, i: number) =>
        i === index ? { ...port, ...updates } : port
      ),
    })
  }

  return (
    <div className="h-full flex flex-col">
      {/* Header */}
      <div className="border-b bg-background px-6 py-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <Button variant="ghost" size="sm" onClick={onClose}>
              <ArrowLeft className="w-4 h-4 mr-2" />
              Back
            </Button>
            <div>
              <h1 className="text-2xl font-bold">{subflow.name}</h1>
              <p className="text-sm text-muted-foreground">{subflow.description || 'No description'}</p>
            </div>
          </div>

          <div className="flex gap-2">
            <Button variant="outline" size="sm" onClick={() => setShowPropertiesDialog(true)}>
              <Settings className="w-4 h-4 mr-2" />
              Properties
            </Button>
            <Button size="sm" onClick={handleSave} disabled={isSaving}>
              <Save className="w-4 h-4 mr-2" />
              {isSaving ? 'Saving...' : 'Save'}
            </Button>
          </div>
        </div>
      </div>

      {/* Main Content */}
      <div className="flex-1 overflow-auto p-6">
        <Tabs defaultValue="ports" className="w-full">
          <TabsList>
            <TabsTrigger value="ports">Input/Output Ports</TabsTrigger>
            <TabsTrigger value="nodes">Nodes</TabsTrigger>
            <TabsTrigger value="properties">Properties</TabsTrigger>
            <TabsTrigger value="info">Info</TabsTrigger>
          </TabsList>

          {/* Ports Tab */}
          <TabsContent value="ports" className="space-y-6">
            {/* Input Ports */}
            <Card>
              <CardHeader>
                <div className="flex items-center justify-between">
                  <div>
                    <CardTitle>Input Ports</CardTitle>
                    <CardDescription>Define input ports for this subflow</CardDescription>
                  </div>
                  <Button size="sm" onClick={addInputPort}>
                    <Plus className="w-4 h-4 mr-2" />
                    Add Input
                  </Button>
                </div>
              </CardHeader>
              <CardContent>
                {!subflow.in || subflow.in.length === 0 ? (
                  <p className="text-sm text-muted-foreground text-center py-8">
                    No input ports defined. Click "Add Input" to create one.
                  </p>
                ) : (
                  <div className="space-y-3">
                    {subflow.in.map((port: any, index: number) => (
                      <div key={index} className="border rounded-lg p-4 space-y-3">
                        <div className="flex items-center justify-between">
                          <Badge>Input {port.index}</Badge>
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => removePort('in', index)}
                          >
                            Remove
                          </Button>
                        </div>

                        <div className="grid grid-cols-2 gap-3">
                          <div>
                            <Label className="text-xs">Name</Label>
                            <Input
                              value={port.name || ''}
                              onChange={(e) =>
                                updatePort('in', index, { name: e.target.value })
                              }
                              placeholder={`Input ${index + 1}`}
                            />
                          </div>
                          <div>
                            <Label className="text-xs">Label</Label>
                            <Input
                              value={port.label || ''}
                              onChange={(e) =>
                                updatePort('in', index, { label: e.target.value })
                              }
                              placeholder="Optional label"
                            />
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </CardContent>
            </Card>

            {/* Output Ports */}
            <Card>
              <CardHeader>
                <div className="flex items-center justify-between">
                  <div>
                    <CardTitle>Output Ports</CardTitle>
                    <CardDescription>Define output ports for this subflow</CardDescription>
                  </div>
                  <Button size="sm" onClick={addOutputPort}>
                    <Plus className="w-4 h-4 mr-2" />
                    Add Output
                  </Button>
                </div>
              </CardHeader>
              <CardContent>
                {!subflow.out || subflow.out.length === 0 ? (
                  <p className="text-sm text-muted-foreground text-center py-8">
                    No output ports defined. Click "Add Output" to create one.
                  </p>
                ) : (
                  <div className="space-y-3">
                    {subflow.out.map((port: any, index: number) => (
                      <div key={index} className="border rounded-lg p-4 space-y-3">
                        <div className="flex items-center justify-between">
                          <Badge variant="secondary">Output {port.index}</Badge>
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => removePort('out', index)}
                          >
                            Remove
                          </Button>
                        </div>

                        <div className="grid grid-cols-2 gap-3">
                          <div>
                            <Label className="text-xs">Name</Label>
                            <Input
                              value={port.name || ''}
                              onChange={(e) =>
                                updatePort('out', index, { name: e.target.value })
                              }
                              placeholder={`Output ${index + 1}`}
                            />
                          </div>
                          <div>
                            <Label className="text-xs">Label</Label>
                            <Input
                              value={port.label || ''}
                              onChange={(e) =>
                                updatePort('out', index, { label: e.target.value })
                              }
                              placeholder="Optional label"
                            />
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </CardContent>
            </Card>
          </TabsContent>

          {/* Nodes Tab */}
          <TabsContent value="nodes">
            <Card>
              <CardHeader>
                <CardTitle>Subflow Nodes</CardTitle>
                <CardDescription>
                  Nodes within this subflow (visual editor coming soon)
                </CardDescription>
              </CardHeader>
              <CardContent>
                <div className="text-center py-12 text-muted-foreground">
                  <p className="mb-2">Visual flow editor integration coming soon</p>
                  <p className="text-sm">
                    For now, use the main flow editor to create flows, then convert them to
                    subflows
                  </p>
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          {/* Properties Tab */}
          <TabsContent value="properties">
            <Card>
              <CardHeader>
                <CardTitle>Configurable Properties</CardTitle>
                <CardDescription>
                  Define properties that can be configured when using this subflow
                </CardDescription>
              </CardHeader>
              <CardContent>
                <p className="text-sm text-muted-foreground text-center py-8">
                  Property configuration coming soon
                </p>
              </CardContent>
            </Card>
          </TabsContent>

          {/* Info Tab */}
          <TabsContent value="info">
            <Card>
              <CardHeader>
                <CardTitle>Documentation</CardTitle>
                <CardDescription>Information about this subflow</CardDescription>
              </CardHeader>
              <CardContent>
                <div className="space-y-4">
                  <div>
                    <Label>Description</Label>
                    <Textarea
                      value={subflow.description || ''}
                      onChange={(e) =>
                        setSubflow({ ...subflow, description: e.target.value })
                      }
                      placeholder="Describe what this subflow does..."
                      rows={4}
                    />
                  </div>

                  <div className="grid grid-cols-2 gap-4">
                    <div>
                      <Label>Version</Label>
                      <Input
                        value={subflow.version || ''}
                        onChange={(e) => setSubflow({ ...subflow, version: e.target.value })}
                        placeholder="1.0.0"
                      />
                    </div>
                    <div>
                      <Label>Author</Label>
                      <Input
                        value={subflow.author || ''}
                        onChange={(e) => setSubflow({ ...subflow, author: e.target.value })}
                        placeholder="Your name"
                      />
                    </div>
                  </div>
                </div>
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>
      </div>

      {/* Properties Dialog */}
      <Dialog open={showPropertiesDialog} onOpenChange={setShowPropertiesDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Subflow Properties</DialogTitle>
            <DialogDescription>Configure subflow metadata</DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            <div>
              <Label>Name</Label>
              <Input
                value={subflow.name}
                onChange={(e) => setSubflow({ ...subflow, name: e.target.value })}
              />
            </div>

            <div>
              <Label>Category</Label>
              <Input
                value={subflow.category || ''}
                onChange={(e) => setSubflow({ ...subflow, category: e.target.value })}
                placeholder="general"
              />
            </div>

            <div>
              <Label>Color</Label>
              <Input
                type="color"
                value={subflow.color || '#666666'}
                onChange={(e) => setSubflow({ ...subflow, color: e.target.value })}
              />
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setShowPropertiesDialog(false)}>
              Close
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
