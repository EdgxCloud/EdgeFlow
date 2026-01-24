/**
 * Node Configuration Dialog
 *
 * Main dialog component for configuring nodes with dynamic property forms
 */

import { useState, useEffect } from 'react'
import { X, Settings, Info, AlertCircle, Loader2, Code2, Cpu, Database, MessageSquare, Zap, Globe, HardDrive, Bot } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { PropertyField } from './PropertyField'
import { useNodeConfig } from '@/hooks/useNodeConfig'
import { shouldUseNodeEditor, renderNodeEditor } from './NodeConfigFactory'
import { cn } from '@/lib/utils'
import type { NodeConfig } from '@/types/node'

interface NodeConfigDialogProps {
  node: NodeConfig | null
  flowId?: string
  onClose: () => void
  onSave: (nodeId: string, config: Record<string, any>) => void
}

export function NodeConfigDialog({ node, flowId, onClose, onSave }: NodeConfigDialogProps) {
  const [activeTab, setActiveTab] = useState('configuration')

  const {
    nodeInfo,
    config,
    errors,
    isLoading,
    isSaving,
    isDirty,
    updateField,
    save,
    reset,
  } = useNodeConfig({
    flowId,
    nodeId: node?.id,
    nodeType: node?.type || '',
    initialConfig: node?.config || {},
  })

  // Reset tab when dialog opens
  useEffect(() => {
    if (node) {
      setActiveTab('configuration')
    }
  }, [node])

  if (!node) return null

  const handleSave = async () => {
    const success = await save()
    if (success) {
      onSave(node.id, config)
      onClose()
    }
  }

  const handleCancel = () => {
    if (isDirty) {
      const confirmed = window.confirm('You have unsaved changes. Are you sure you want to close?')
      if (!confirmed) return
    }
    reset()
    onClose()
  }

  const nodeLabel = config.label || node.name || nodeInfo?.name || 'Node'
  const nodeColor = nodeInfo?.color || '#6b7280'

  // Get icon component based on category
  const getCategoryIcon = (category: string) => {
    const icons: Record<string, any> = {
      input: Zap,
      output: HardDrive,
      function: Code2,
      processing: Cpu,
      database: Database,
      messaging: MessageSquare,
      network: Globe,
      ai: Bot,
    }
    return icons[category] || Settings
  }

  const CategoryIcon = getCategoryIcon(nodeInfo?.category || '')

  // Group properties by group if they have one
  const groupedProperties = (nodeInfo?.properties || []).reduce(
    (acc, prop) => {
      const group = prop.group || 'General'
      if (!acc[group]) acc[group] = []
      acc[group].push(prop)
      return acc
    },
    {} as Record<string, any[]>
  )

  return (
    <div
      className="fixed inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center z-[9999] p-4"
      onClick={handleCancel}
    >
      <div
        className="bg-white dark:bg-gray-900 rounded-2xl shadow-2xl w-full max-w-3xl max-h-[85vh] overflow-hidden flex flex-col border border-gray-200 dark:border-gray-800"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header - Clean Minimal Design */}
        <div className="flex items-start justify-between p-6 border-b border-gray-100 dark:border-gray-800">
          <div className="flex items-center gap-4">
            {/* Colored accent bar + Icon */}
            <div className="relative">
              <div
                className="w-11 h-11 rounded-lg flex items-center justify-center"
                style={{ backgroundColor: `${nodeColor}12` }}
              >
                <CategoryIcon
                  className="w-5 h-5"
                  style={{ color: nodeColor }}
                />
              </div>
              {/* Colored accent line */}
              <div
                className="absolute -left-6 top-0 bottom-0 w-1 rounded-r-full"
                style={{ backgroundColor: nodeColor }}
              />
            </div>

            <div className="space-y-1">
              {/* Title */}
              <h2 className="text-lg font-semibold text-gray-900 dark:text-white leading-none">
                {nodeLabel}
              </h2>
              {/* Subtitle with category */}
              <p className="text-sm text-gray-500 dark:text-gray-400 flex items-center gap-2">
                <span
                  className="inline-block w-2 h-2 rounded-full"
                  style={{ backgroundColor: nodeColor }}
                />
                {nodeInfo?.category || 'node'} · {node.type}
                {isDirty && (
                  <span className="text-amber-500 dark:text-amber-400">· Unsaved</span>
                )}
              </p>
            </div>
          </div>

          {/* Close button */}
          <button
            onClick={handleCancel}
            className="p-2 -m-2 text-gray-400 hover:text-gray-600 dark:hover:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 rounded-lg transition-colors"
            aria-label="Close dialog"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Loading State */}
        {isLoading && (
          <div className="flex-1 flex items-center justify-center p-12">
            <div className="text-center">
              <Loader2 className="w-8 h-8 animate-spin mx-auto mb-3 text-primary" />
              <p className="text-sm text-muted-foreground">Loading configuration...</p>
            </div>
          </div>
        )}

        {/* Content */}
        {!isLoading && (
          <>
            <Tabs
              value={activeTab}
              onValueChange={setActiveTab}
              className="flex-1 flex flex-col overflow-hidden"
            >
              <TabsList className="w-full justify-start rounded-none border-b bg-transparent px-6 pt-4">
                <TabsTrigger
                  value="configuration"
                  className="data-[state=active]:bg-primary data-[state=active]:text-primary-foreground"
                >
                  <Settings className="w-4 h-4 mr-2" />
                  Configuration
                </TabsTrigger>
                <TabsTrigger
                  value="info"
                  className="data-[state=active]:bg-primary data-[state=active]:text-primary-foreground"
                >
                  <Info className="w-4 h-4 mr-2" />
                  Info
                </TabsTrigger>
              </TabsList>

              {/* Configuration Tab */}
              <div className="flex-1 overflow-y-auto">
                <TabsContent value="configuration" className="p-6 space-y-6 mt-0">
                  {/* Validation Errors */}
                  {errors.length > 0 && (
                    <Alert variant="destructive">
                      <AlertCircle className="h-4 w-4" />
                      <AlertDescription>
                        <p className="font-semibold mb-2">Please fix the following errors:</p>
                        <ul className="list-disc list-inside space-y-1">
                          {errors.map((error, index) => (
                            <li key={index} className="text-sm">
                              {error.message}
                            </li>
                          ))}
                        </ul>
                      </AlertDescription>
                    </Alert>
                  )}

                  {/* Specialized Node Editor or Property Fields */}
                  {shouldUseNodeEditor(node.type) ? (
                    // Use specialized editor for entire node config
                    <div>
                      {renderNodeEditor(
                        node.type,
                        config,
                        (newConfig) => {
                          // Update all fields from specialized editor
                          Object.keys(newConfig).forEach((key) => {
                            updateField(key, newConfig[key])
                          })
                        },
                        isSaving
                      )}
                    </div>
                  ) : (
                    // Use generic property fields
                    groupedProperties &&
                    Object.entries(groupedProperties).map(([group, properties]) => (
                      <div key={group} className="space-y-4">
                        {Object.keys(groupedProperties).length > 1 && (
                          <h3 className="text-sm font-semibold text-muted-foreground uppercase tracking-wide">
                            {group}
                          </h3>
                        )}
                        {properties.map((prop) => (
                          <PropertyField
                            key={prop.name}
                            schema={prop}
                            value={config[prop.name]}
                            onChange={(value) => updateField(prop.name, value)}
                            error={errors.find((e) => e.field === prop.name)?.message}
                            disabled={isSaving}
                          />
                        ))}
                      </div>
                    ))
                  )}

                  {/* No Properties - Only show if NOT using specialized editor */}
                  {(!nodeInfo?.properties || nodeInfo.properties.length === 0) &&
                    !shouldUseNodeEditor(node.type) && (
                    <div className="text-center py-12">
                      <Settings className="w-12 h-12 mx-auto mb-3 text-muted-foreground/50" />
                      <p className="text-sm text-muted-foreground">
                        This node has no configurable properties
                      </p>
                    </div>
                  )}
                </TabsContent>

                {/* Info Tab */}
                <TabsContent value="info" className="p-6 space-y-6 mt-0">
                  {/* Description */}
                  {nodeInfo?.description && (
                    <div className="space-y-2">
                      <h4 className="text-sm font-semibold">Description</h4>
                      <p className="text-sm text-muted-foreground">{nodeInfo.description}</p>
                    </div>
                  )}

                  {/* Node Information Card */}
                  <div className="p-4 bg-muted/50 rounded-lg border border-border">
                    <h4 className="text-sm font-semibold mb-3">Node Information</h4>
                    <dl className="space-y-2 text-sm">
                      <div className="flex justify-between">
                        <dt className="text-muted-foreground">Type:</dt>
                        <dd className="font-medium">{node.type}</dd>
                      </div>
                      <div className="flex justify-between">
                        <dt className="text-muted-foreground">Category:</dt>
                        <dd className="font-medium capitalize">
                          {nodeInfo?.category || 'Unknown'}
                        </dd>
                      </div>
                      <div className="flex justify-between">
                        <dt className="text-muted-foreground">ID:</dt>
                        <dd className="font-mono text-xs">{node.id}</dd>
                      </div>
                      {nodeInfo?.version && (
                        <div className="flex justify-between">
                          <dt className="text-muted-foreground">Version:</dt>
                          <dd className="font-medium">{nodeInfo.version}</dd>
                        </div>
                      )}
                      {nodeInfo?.author && (
                        <div className="flex justify-between">
                          <dt className="text-muted-foreground">Author:</dt>
                          <dd className="font-medium">{nodeInfo.author}</dd>
                        </div>
                      )}
                    </dl>
                  </div>

                  {/* Inputs/Outputs */}
                  {(nodeInfo?.inputs || nodeInfo?.outputs) && (
                    <div className="grid grid-cols-2 gap-4">
                      {nodeInfo.inputs && nodeInfo.inputs.length > 0 && (
                        <div className="space-y-2">
                          <h4 className="text-sm font-semibold">Inputs</h4>
                          <div className="space-y-1">
                            {nodeInfo.inputs.map((input, index) => (
                              <div
                                key={index}
                                className="text-xs p-2 bg-blue-50 dark:bg-blue-950/20 rounded border border-blue-200 dark:border-blue-900"
                              >
                                <div className="font-medium text-blue-900 dark:text-blue-100">
                                  {input.label}
                                </div>
                                {input.type && (
                                  <div className="text-blue-700 dark:text-blue-300 mt-0.5">
                                    Type: {input.type}
                                  </div>
                                )}
                              </div>
                            ))}
                          </div>
                        </div>
                      )}

                      {nodeInfo.outputs && nodeInfo.outputs.length > 0 && (
                        <div className="space-y-2">
                          <h4 className="text-sm font-semibold">Outputs</h4>
                          <div className="space-y-1">
                            {nodeInfo.outputs.map((output, index) => (
                              <div
                                key={index}
                                className="text-xs p-2 bg-green-50 dark:bg-green-950/20 rounded border border-green-200 dark:border-green-900"
                              >
                                <div className="font-medium text-green-900 dark:text-green-100">
                                  {output.label}
                                </div>
                                {output.type && (
                                  <div className="text-green-700 dark:text-green-300 mt-0.5">
                                    Type: {output.type}
                                  </div>
                                )}
                              </div>
                            ))}
                          </div>
                        </div>
                      )}
                    </div>
                  )}
                </TabsContent>
              </div>
            </Tabs>

            {/* Footer */}
            <div className="flex items-center justify-between p-6 border-t border-border bg-muted/30">
              <p className="text-xs text-muted-foreground">
                {isDirty ? 'You have unsaved changes' : 'Double-click nodes to edit settings'}
              </p>
              <div className="flex gap-3">
                <Button variant="outline" onClick={handleCancel} disabled={isSaving}>
                  Cancel
                </Button>
                <Button onClick={handleSave} disabled={isSaving || !isDirty}>
                  {isSaving ? (
                    <>
                      <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                      Saving...
                    </>
                  ) : (
                    'Save Changes'
                  )}
                </Button>
              </div>
            </div>
          </>
        )}
      </div>
    </div>
  )
}
