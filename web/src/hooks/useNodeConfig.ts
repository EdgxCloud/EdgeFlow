/**
 * Node Configuration Hook
 *
 * Custom hook for managing node configuration state
 */

import { useState, useEffect, useCallback } from 'react'
import { getNodeType, validateNodeConfig } from '@/api/nodes'
import type { NodeInfo, NodeConfig, ValidationError } from '@/types/node'
import { toast } from 'sonner'

interface UseNodeConfigOptions {
  flowId?: string
  nodeId?: string
  nodeType: string
  initialConfig?: Record<string, any>
}

interface UseNodeConfigReturn {
  nodeInfo: NodeInfo | null
  config: Record<string, any>
  errors: ValidationError[]
  isLoading: boolean
  isSaving: boolean
  isDirty: boolean
  updateField: (field: string, value: any) => void
  updateConfig: (newConfig: Record<string, any>) => void
  validate: () => Promise<boolean>
  save: () => Promise<boolean>
  reset: () => void
}

export function useNodeConfig(options: UseNodeConfigOptions): UseNodeConfigReturn {
  const { flowId, nodeId, nodeType, initialConfig = {} } = options

  // Debug: log what config the dialog receives
  if (nodeType === 'inject') {
    console.log('[useNodeConfig] Init inject with config:', JSON.stringify(initialConfig))
  }

  const [nodeInfo, setNodeInfo] = useState<NodeInfo | null>(null)
  const [config, setConfig] = useState<Record<string, any>>(initialConfig)
  const [initialState, setInitialState] = useState<Record<string, any>>(initialConfig)
  const [errors, setErrors] = useState<ValidationError[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [isSaving, setIsSaving] = useState(false)
  const [isDirty, setIsDirty] = useState(false)

  // Fetch node type information
  useEffect(() => {
    async function fetchNodeInfo() {
      if (!nodeType) return

      setIsLoading(true)
      try {
        const info = await getNodeType(nodeType)
        setNodeInfo(info)

        // Set default values from schema for fields not already in config
        const defaultConfig: Record<string, any> = {}
        if (info.properties) {
          info.properties.forEach((prop) => {
            if (prop.default !== undefined && config[prop.name] === undefined) {
              defaultConfig[prop.name] = prop.default
            }
          })
        }

        if (Object.keys(defaultConfig).length > 0) {
          if (nodeType === 'inject') {
            console.log('[useNodeConfig] Applying defaults to inject:', JSON.stringify(defaultConfig))
          }
          setConfig((prev) => ({ ...defaultConfig, ...prev }))
          setInitialState((prev) => ({ ...defaultConfig, ...prev }))
        }
      } catch (error) {
        console.error('Failed to fetch node info:', error)
        toast.error('Failed to load node configuration')
      } finally {
        setIsLoading(false)
      }
    }

    fetchNodeInfo()
  }, [nodeType])

  // Track if config has changed
  useEffect(() => {
    const hasChanged = JSON.stringify(config) !== JSON.stringify(initialState)
    setIsDirty(hasChanged)
  }, [config, initialState])

  // Update a single field
  const updateField = useCallback((field: string, value: any) => {
    setConfig((prev) => ({ ...prev, [field]: value }))
    // Clear error for this field if it exists
    setErrors((prev) => prev.filter((err) => err.field !== field))
  }, [])

  // Update entire config
  const updateConfig = useCallback((newConfig: Record<string, any>) => {
    setConfig(newConfig)
    setErrors([])
  }, [])

  // Validate configuration
  const validate = useCallback(async (): Promise<boolean> => {
    // If no nodeInfo or no properties, skip validation (return valid)
    if (!nodeInfo || !nodeInfo.properties) return true

    const validationErrors: ValidationError[] = []

    // Client-side validation
    nodeInfo.properties.forEach((prop) => {
      const value = config[prop.name]

      // Required field validation
      if (prop.required && (value === undefined || value === null || value === '')) {
        validationErrors.push({
          field: prop.name,
          message: `${prop.label} is required`,
        })
      }

      // Type validation
      if (value !== undefined && value !== null) {
        switch (prop.type) {
          case 'number':
            if (isNaN(Number(value))) {
              validationErrors.push({
                field: prop.name,
                message: `${prop.label} must be a number`,
              })
            } else {
              const numValue = Number(value)
              if (prop.min !== undefined && numValue < prop.min) {
                validationErrors.push({
                  field: prop.name,
                  message: `${prop.label} must be at least ${prop.min}`,
                })
              }
              if (prop.max !== undefined && numValue > prop.max) {
                validationErrors.push({
                  field: prop.name,
                  message: `${prop.label} must be at most ${prop.max}`,
                })
              }
            }
            break

          case 'string':
            if (typeof value !== 'string') {
              validationErrors.push({
                field: prop.name,
                message: `${prop.label} must be a string`,
              })
            } else if (prop.validation) {
              const regex = new RegExp(prop.validation)
              if (!regex.test(value)) {
                validationErrors.push({
                  field: prop.name,
                  message: `${prop.label} format is invalid`,
                })
              }
            }
            break

          case 'boolean':
            if (typeof value !== 'boolean') {
              validationErrors.push({
                field: prop.name,
                message: `${prop.label} must be true or false`,
              })
            }
            break

          case 'array':
            if (!Array.isArray(value)) {
              validationErrors.push({
                field: prop.name,
                message: `${prop.label} must be an array`,
              })
            }
            break

          case 'object':
          case 'json':
            if (typeof value !== 'object' || Array.isArray(value)) {
              validationErrors.push({
                field: prop.name,
                message: `${prop.label} must be an object`,
              })
            }
            break
        }
      }
    })

    setErrors(validationErrors)

    if (validationErrors.length > 0) {
      toast.error(`Validation failed: ${validationErrors.length} error(s)`)
      return false
    }

    // Server-side validation (if available)
    try {
      const result = await validateNodeConfig(nodeType, config)
      if (!result.valid) {
        const serverErrors =
          result.errors?.map((msg) => ({
            field: '',
            message: msg,
          })) || []
        setErrors(serverErrors)
        toast.error('Server validation failed')
        return false
      }
    } catch (error) {
      // Server validation is optional
      console.warn('Server validation not available:', error)
    }

    return true
  }, [nodeInfo, config, nodeType])

  // Save configuration
  const save = useCallback(async (): Promise<boolean> => {
    // Validate first
    const isValid = await validate()
    if (!isValid) return false

    setIsSaving(true)
    try {
      // Config is saved to storage via the parent's onSave callback
      // which triggers handleNodeSettingsSave → updateFlow (PUT /flows/:id)
      // No need for a separate updateNodeConfig call here

      // Update local state
      setInitialState(config)
      setIsDirty(false)
      toast.success('Configuration saved successfully')
      return true
    } catch (error) {
      console.error('Failed to save config:', error)
      toast.error('Failed to save configuration')
      return false
    } finally {
      setIsSaving(false)
    }
  }, [validate, config])

  // Reset to initial state
  const reset = useCallback(() => {
    setConfig(initialState)
    setErrors([])
    setIsDirty(false)
  }, [initialState])

  return {
    nodeInfo,
    config,
    errors,
    isLoading,
    isSaving,
    isDirty,
    updateField,
    updateConfig,
    validate,
    save,
    reset,
  }
}
