import React, { useState } from 'react'
import {
  Package,
  Power,
  PowerOff,
  RefreshCw,
  AlertCircle,
  CheckCircle,
  Clock,
  HardDrive,
  ChevronDown,
  ChevronUp,
  Users,
  Box
} from 'lucide-react'
import { Module } from '../../lib/api'
import { Button } from '@/components/ui/button'

interface ModuleCardProps {
  module: Module
  onLoad: (name: string) => void
  onUnload: (name: string) => void
  onEnable: (name: string) => void
  onDisable: (name: string) => void
  onReload: (name: string) => void
}

const ModuleCard: React.FC<ModuleCardProps> = ({
  module,
  onLoad,
  onUnload,
  onEnable,
  onDisable,
  onReload,
}) => {
  const [expanded, setExpanded] = useState(false)

  const getCategoryColor = (category: string) => {
    const colors: Record<string, string> = {
      core: 'bg-blue-100 text-blue-800 dark:bg-blue-900/30 dark:text-blue-400',
      network: 'bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400',
      gpio: 'bg-purple-100 text-purple-800 dark:bg-purple-900/30 dark:text-purple-400',
      database: 'bg-orange-100 text-orange-800 dark:bg-orange-900/30 dark:text-orange-400',
      messaging: 'bg-pink-100 text-pink-800 dark:bg-pink-900/30 dark:text-pink-400',
      ai: 'bg-indigo-100 text-indigo-800 dark:bg-indigo-900/30 dark:text-indigo-400',
      industrial: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400',
      advanced: 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400',
      ui: 'bg-teal-100 text-teal-800 dark:bg-teal-900/30 dark:text-teal-400',
    }
    return colors[category] || 'bg-gray-100 text-gray-800 dark:bg-gray-900/30 dark:text-gray-400'
  }

  const getCategoryText = (category: string) => {
    const texts: Record<string, string> = {
      core: 'Core',
      network: 'Network',
      gpio: 'Hardware',
      database: 'Database',
      messaging: 'Messaging',
      ai: 'AI',
      industrial: 'Industrial',
      advanced: 'Advanced',
      ui: 'UI',
    }
    return texts[category] || category
  }

  const getStatusBadge = (status: string) => {
    const config: Record<string, { icon: React.ElementType; color: string; text: string }> = {
      loaded: { icon: CheckCircle, color: 'text-green-600 bg-green-100 dark:bg-green-900/30', text: 'Loaded' },
      loading: { icon: Clock, color: 'text-yellow-600 bg-yellow-100 dark:bg-yellow-900/30', text: 'Loading' },
      not_loaded: { icon: Package, color: 'text-gray-600 bg-gray-100 dark:bg-gray-900/30', text: 'Not Loaded' },
      unloading: { icon: Clock, color: 'text-orange-600 bg-orange-100 dark:bg-orange-900/30', text: 'Unloading' },
      error: { icon: AlertCircle, color: 'text-red-600 bg-red-100 dark:bg-red-900/30', text: 'Error' },
    }

    const { icon: Icon, color, text } = config[status] || config.not_loaded

    return (
      <span className={`inline-flex items-center gap-1 px-2 py-1 rounded-full text-xs font-medium ${color}`}>
        <Icon className="w-3 h-3" />
        {text}
      </span>
    )
  }

  const isLoaded = module.status === 'loaded'
  const canLoad = module.compatible && module.status === 'not_loaded'
  const canUnload = isLoaded && module.category !== 'core' // Core modules cannot be unloaded
  const isCore = module.category === 'core'

  return (
    <div className="bg-white dark:bg-gray-800 rounded-lg shadow-md overflow-hidden border border-gray-200 dark:border-gray-700 hover:shadow-lg transition-shadow">
      {/* Header */}
      <div className="p-4 border-b border-gray-200 dark:border-gray-700">
        <div className="flex items-start justify-between mb-2">
          <div className="flex items-center gap-2">
            <Package className="w-5 h-5 text-blue-600 dark:text-blue-400" />
            <h3 className="text-lg font-bold text-gray-900 dark:text-white">
              {module.name}
            </h3>
          </div>
          {getStatusBadge(module.status)}
        </div>

        <p className="text-sm text-gray-600 dark:text-gray-400 mb-2">
          {module.description}
        </p>

        <div className="flex items-center gap-2 flex-wrap">
          <span className={`px-2 py-1 rounded text-xs font-medium ${getCategoryColor(module.category)}`}>
            {getCategoryText(module.category)}
          </span>
          <span className="text-xs text-gray-500 dark:text-gray-400">
            v{module.version}
          </span>
          {module.author && (
            <span className="text-xs text-gray-500 dark:text-gray-400 flex items-center gap-1">
              <Users className="w-3 h-3" />
              {module.author}
            </span>
          )}
        </div>
      </div>

      {/* Stats */}
      <div className="p-4 bg-gray-50 dark:bg-gray-900/50">
        <div className="grid grid-cols-3 gap-4 text-center">
          <div>
            <div className="text-xs text-gray-500 dark:text-gray-400 mb-1">
              Memory
            </div>
            <div className="text-sm font-bold text-gray-900 dark:text-white">
              {module.required_memory_mb}MB
            </div>
          </div>
          <div>
            <div className="text-xs text-gray-500 dark:text-gray-400 mb-1">
              Disk
            </div>
            <div className="text-sm font-bold text-gray-900 dark:text-white">
              {module.required_disk_mb}MB
            </div>
          </div>
          <div>
            <div className="text-xs text-gray-500 dark:text-gray-400 mb-1">
              Nodes
            </div>
            <div className="text-sm font-bold text-gray-900 dark:text-white flex items-center justify-center gap-1">
              <Box className="w-3 h-3" />
              {module.nodes.length}
            </div>
          </div>
        </div>
      </div>

      {/* Compatibility Warning */}
      {!module.compatible && (
        <div className="px-4 py-2 bg-red-50 dark:bg-red-900/20 border-t border-red-200 dark:border-red-800">
          <div className="flex items-start gap-2">
            <AlertCircle className="w-4 h-4 text-red-600 dark:text-red-400 flex-shrink-0 mt-0.5" />
            <div className="text-xs text-red-700 dark:text-red-300">
              <strong>Incompatible:</strong> {module.compatible_reason}
            </div>
          </div>
        </div>
      )}

      {/* Error Message */}
      {module.error && (
        <div className="px-4 py-2 bg-red-50 dark:bg-red-900/20 border-t border-red-200 dark:border-red-800">
          <div className="flex items-start gap-2">
            <AlertCircle className="w-4 h-4 text-red-600 dark:text-red-400 flex-shrink-0 mt-0.5" />
            <div className="text-xs text-red-700 dark:text-red-300">
              <strong>Error:</strong> {module.error}
            </div>
          </div>
        </div>
      )}

      {/* Actions */}
      <div className="p-4 border-t border-gray-200 dark:border-gray-700 space-y-2">
        <div className="flex items-center gap-2">
          {canLoad && (
            <Button
              onClick={() => onLoad(module.name)}
              variant="default"
              size="sm"
              className="flex-1"
            >
              <Power className="w-4 h-4 mr-2" />
              Load
            </Button>
          )}

          {isLoaded && !isCore && (
            <>
              <Button
                onClick={() => onUnload(module.name)}
                variant="secondary"
                size="sm"
                className="flex-1"
              >
                <PowerOff className="w-4 h-4 mr-2" />
                Unload
              </Button>

              <Button
                onClick={() => onReload(module.name)}
                variant="secondary"
                size="sm"
              >
                <RefreshCw className="w-4 h-4" />
              </Button>
            </>
          )}

          {isCore && isLoaded && (
            <div className="flex-1 text-center text-xs text-gray-500 dark:text-gray-400 py-2">
              Core module - Always active
            </div>
          )}
        </div>

        {/* Expand/Collapse */}
        <button
          onClick={() => setExpanded(!expanded)}
          className="w-full flex items-center justify-center gap-2 text-xs text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-white transition-colors"
        >
          {expanded ? (
            <>
              <ChevronUp className="w-4 h-4" />
              Less
            </>
          ) : (
            <>
              <ChevronDown className="w-4 h-4" />
              More
            </>
          )}
        </button>
      </div>

      {/* Expanded Content */}
      {expanded && (
        <div className="p-4 bg-gray-50 dark:bg-gray-900/50 border-t border-gray-200 dark:border-gray-700 space-y-3">
          {/* Dependencies */}
          {module.dependencies && module.dependencies.length > 0 && (
            <div>
              <div className="text-xs font-semibold text-gray-700 dark:text-gray-300 mb-1">
                Dependencies:
              </div>
              <div className="flex flex-wrap gap-1">
                {module.dependencies.map(dep => (
                  <span
                    key={dep}
                    className="px-2 py-1 bg-gray-200 dark:bg-gray-700 rounded text-xs text-gray-700 dark:text-gray-300"
                  >
                    {dep}
                  </span>
                ))}
              </div>
            </div>
          )}

          {/* Nodes */}
          {module.nodes && module.nodes.length > 0 && (
            <div>
              <div className="text-xs font-semibold text-gray-700 dark:text-gray-300 mb-1">
                Provided Nodes ({module.nodes.length}):
              </div>
              <div className="grid grid-cols-2 gap-1">
                {module.nodes.slice(0, 8).map(node => (
                  <div
                    key={node.type}
                    className="text-xs text-gray-600 dark:text-gray-400 flex items-center gap-1"
                  >
                    <div
                      className="w-2 h-2 rounded-full"
                      style={{ backgroundColor: node.color || '#gray' }}
                    />
                    {node.name}
                  </div>
                ))}
                {module.nodes.length > 8 && (
                  <div className="text-xs text-gray-500 dark:text-gray-400">
                    +{module.nodes.length - 8} more...
                  </div>
                )}
              </div>
            </div>
          )}

          {/* Timestamps */}
          {(module.loaded_at || module.unloaded_at) && (
            <div className="text-xs text-gray-500 dark:text-gray-400 space-y-1">
              {module.loaded_at && (
                <div className="flex items-center gap-2">
                  <Clock className="w-3 h-3" />
                  Loaded: {new Date(module.loaded_at).toLocaleString('en-US')}
                </div>
              )}
              {module.unloaded_at && (
                <div className="flex items-center gap-2">
                  <Clock className="w-3 h-3" />
                  Unloaded: {new Date(module.unloaded_at).toLocaleString('en-US')}
                </div>
              )}
            </div>
          )}
        </div>
      )}
    </div>
  )
}

export default ModuleCard
