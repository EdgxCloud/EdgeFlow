import React, { useState, useEffect } from 'react'
import {
  Package,
  Power,
  PowerOff,
  RefreshCw,
  AlertCircle,
  CheckCircle,
  Clock,
  HardDrive,
  Cpu,
  Download,
  Upload,
  Info,
  Plus
} from 'lucide-react'
import { modulesApi, resourcesApi, Module, ModuleStats, ResourceStats } from '../lib/api'
import ResourceBar from '../components/modules/ResourceBar'
import ModuleCard from '../components/modules/ModuleCard'
import ModuleMarketplaceDialog from '../components/modules/ModuleMarketplaceDialog'
import { Button } from '@/components/ui/button'

const ModuleManager: React.FC = () => {
  const [modules, setModules] = useState<Module[]>([])
  const [stats, setStats] = useState<ModuleStats | null>(null)
  const [resources, setResources] = useState<ResourceStats | null>(null)
  const [loading, setLoading] = useState(true)
  const [selectedCategory, setSelectedCategory] = useState<string>('all')
  const [searchQuery, setSearchQuery] = useState('')
  const [marketplaceOpen, setMarketplaceOpen] = useState(false)

  useEffect(() => {
    loadData()
    const interval = setInterval(loadData, 5000) // Refresh every 5 seconds
    return () => clearInterval(interval)
  }, [])

  const loadData = async () => {
    try {
      const [modulesRes, statsRes, resourcesRes] = await Promise.all([
        modulesApi.list(),
        modulesApi.stats(),
        resourcesApi.report(),
      ])

      setModules(modulesRes.data.modules || [])
      setStats(statsRes.data)
      setResources(resourcesRes.data)
    } catch (error) {
      console.error('Failed to load module data:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleLoadModule = async (name: string) => {
    try {
      await modulesApi.load(name)
      await loadData()
    } catch (error: any) {
      alert(`Error loading module: ${error.response?.data?.error || error.message}`)
    }
  }

  const handleUnloadModule = async (name: string) => {
    if (!confirm(`Are you sure you want to unload ${name}?`)) {
      return
    }

    try {
      await modulesApi.unload(name)
      await loadData()
    } catch (error: any) {
      alert(`Error unloading module: ${error.response?.data?.error || error.message}`)
    }
  }

  const handleEnableModule = async (name: string) => {
    try {
      await modulesApi.enable(name)
      await loadData()
    } catch (error: any) {
      alert(`Error enabling module: ${error.response?.data?.error || error.message}`)
    }
  }

  const handleDisableModule = async (name: string) => {
    if (!confirm(`Are you sure you want to disable ${name}?`)) {
      return
    }

    try {
      await modulesApi.disable(name)
      await loadData()
    } catch (error: any) {
      alert(`Error disabling module: ${error.response?.data?.error || error.message}`)
    }
  }

  const handleReloadModule = async (name: string) => {
    try {
      await modulesApi.reload(name)
      await loadData()
    } catch (error: any) {
      alert(`Error reloading module: ${error.response?.data?.error || error.message}`)
    }
  }

  const categories = [
    { id: 'all', name: 'All Modules', icon: Package },
    { id: 'core', name: 'Core', icon: Cpu },
    { id: 'network', name: 'Network', icon: Download },
    { id: 'gpio', name: 'Hardware', icon: HardDrive },
    { id: 'database', name: 'Database', icon: Upload },
    { id: 'messaging', name: 'Messaging', icon: Info },
    { id: 'ai', name: 'AI', icon: AlertCircle },
    { id: 'industrial', name: 'Industrial', icon: CheckCircle },
  ]

  const filteredModules = modules.filter(mod => {
    const matchesCategory = selectedCategory === 'all' || mod.category === selectedCategory
    const matchesSearch = mod.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
                         mod.description.toLowerCase().includes(searchQuery.toLowerCase())
    return matchesCategory && matchesSearch
  })

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'loaded': return 'text-green-600 bg-green-100'
      case 'loading': return 'text-yellow-600 bg-yellow-100'
      case 'not_loaded': return 'text-gray-600 bg-gray-100'
      case 'error': return 'text-red-600 bg-red-100'
      default: return 'text-gray-600 bg-gray-100'
    }
  }

  const getStatusText = (status: string) => {
    switch (status) {
      case 'loaded': return 'Loaded'
      case 'loading': return 'Loading'
      case 'not_loaded': return 'Not Loaded'
      case 'unloading': return 'Unloading'
      case 'error': return 'Error'
      default: return status
    }
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-screen">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
      </div>
    )
  }

  return (
    <div className="p-6 space-y-6 max-w-7xl mx-auto">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900 dark:text-white flex items-center gap-2">
            <Package className="w-8 h-8" />
            Module Manager
          </h1>
          <p className="text-gray-600 dark:text-gray-400 mt-1">
            Manage plugins and system resources
          </p>
        </div>
        <div className="flex items-center gap-2">
          <Button onClick={() => setMarketplaceOpen(true)} variant="default">
            <Plus className="w-4 h-4 mr-2" />
            Add Module
          </Button>
          <Button onClick={loadData} variant="secondary">
            <RefreshCw className="w-4 h-4 mr-2" />
            Refresh
          </Button>
        </div>
      </div>

      {/* Resource Stats */}
      {resources && (
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div className="bg-white dark:bg-gray-800 p-4 rounded-lg shadow">
            <div className="flex items-center justify-between mb-2">
              <span className="text-sm font-medium text-gray-600 dark:text-gray-400">
                Memory (RAM)
              </span>
              <span className="text-xs text-gray-500">
                {resources.memory.used_mb}MB / {resources.memory.total_mb}MB
              </span>
            </div>
            <ResourceBar
              used={resources.memory.used_mb}
              total={resources.memory.total_mb}
              type="memory"
            />
            <div className="mt-2 text-xs text-gray-500">
              Available: {resources.memory.available_mb}MB ({resources.memory.percent})
            </div>
          </div>

          <div className="bg-white dark:bg-gray-800 p-4 rounded-lg shadow">
            <div className="flex items-center justify-between mb-2">
              <span className="text-sm font-medium text-gray-600 dark:text-gray-400">
                Disk
              </span>
              <span className="text-xs text-gray-500">
                {resources.disk.used_mb}MB / {resources.disk.total_mb}MB
              </span>
            </div>
            <ResourceBar
              used={resources.disk.used_mb}
              total={resources.disk.total_mb}
              type="disk"
            />
            <div className="mt-2 text-xs text-gray-500">
              Available: {resources.disk.available_mb}MB ({resources.disk.percent})
            </div>
          </div>

          <div className="bg-white dark:bg-gray-800 p-4 rounded-lg shadow">
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <span className="text-sm font-medium text-gray-600 dark:text-gray-400">
                  <Cpu className="w-4 h-4 inline mr-2" />
                  CPU
                </span>
                <span className="text-sm font-bold text-gray-900 dark:text-white">
                  {resources.cpu.cores} cores
                </span>
              </div>
              <div className="flex items-center justify-between">
                <span className="text-sm text-gray-600 dark:text-gray-400">
                  Goroutines
                </span>
                <span className="text-sm font-bold text-gray-900 dark:text-white">
                  {resources.cpu.goroutines}
                </span>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Module Stats */}
      {stats && (
        <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
          <div className="bg-blue-50 dark:bg-blue-900/20 p-4 rounded-lg">
            <div className="text-2xl font-bold text-blue-600 dark:text-blue-400">
              {stats.total_plugins}
            </div>
            <div className="text-sm text-gray-600 dark:text-gray-400">
              Total Modules
            </div>
          </div>

          <div className="bg-green-50 dark:bg-green-900/20 p-4 rounded-lg">
            <div className="text-2xl font-bold text-green-600 dark:text-green-400">
              {stats.loaded_plugins}
            </div>
            <div className="text-sm text-gray-600 dark:text-gray-400">
              Loaded
            </div>
          </div>

          <div className="bg-purple-50 dark:bg-purple-900/20 p-4 rounded-lg">
            <div className="text-2xl font-bold text-purple-600 dark:text-purple-400">
              {stats.enabled_plugins}
            </div>
            <div className="text-sm text-gray-600 dark:text-gray-400">
              Enabled
            </div>
          </div>

          <div className="bg-orange-50 dark:bg-orange-900/20 p-4 rounded-lg">
            <div className="text-2xl font-bold text-orange-600 dark:text-orange-400">
              {stats.total_nodes}
            </div>
            <div className="text-sm text-gray-600 dark:text-gray-400">
              Total Nodes
            </div>
          </div>
        </div>
      )}

      {/* Category Filters */}
      <div className="flex items-center gap-2 overflow-x-auto pb-2">
        {categories.map(cat => {
          const Icon = cat.icon
          return (
            <button
              key={cat.id}
              onClick={() => setSelectedCategory(cat.id)}
              className={`flex items-center gap-2 px-4 py-2 rounded-lg whitespace-nowrap transition-colors ${
                selectedCategory === cat.id
                  ? 'bg-blue-600 text-white'
                  : 'bg-white dark:bg-gray-800 text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-700'
              }`}
            >
              <Icon className="w-4 h-4" />
              {cat.name}
            </button>
          )
        })}
      </div>

      {/* Search */}
      <div className="flex items-center gap-4">
        <input
          type="text"
          placeholder="Search modules..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="flex-1 px-4 py-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-white"
        />
      </div>

      {/* Module Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {filteredModules.map(module => (
          <ModuleCard
            key={module.name}
            module={module}
            onLoad={handleLoadModule}
            onUnload={handleUnloadModule}
            onEnable={handleEnableModule}
            onDisable={handleDisableModule}
            onReload={handleReloadModule}
          />
        ))}
      </div>

      {filteredModules.length === 0 && (
        <div className="text-center py-12 text-gray-500 dark:text-gray-400">
          <Package className="w-16 h-16 mx-auto mb-4 opacity-50" />
          <p>No modules found</p>
        </div>
      )}

      {/* Module Marketplace Dialog */}
      <ModuleMarketplaceDialog
        open={marketplaceOpen}
        onOpenChange={setMarketplaceOpen}
        onModuleInstalled={loadData}
      />
    </div>
  )
}

export default ModuleManager
