/**
 * Subflows Page
 *
 * Main page for managing reusable subflows
 */

import { useState, useEffect } from 'react'
import { Plus, Download, Upload, Copy, Trash2, Edit, FolderOpen } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { toast } from 'sonner'
import { SubflowEditor } from '@/components/Subflow/SubflowEditor'

interface SubflowDefinition {
  id: string
  name: string
  description?: string
  category?: string
  icon?: string
  color?: string
  version?: string
  author?: string
  updatedAt: string
  nodes: any[]
  in: any[]
  out: any[]
}

export function Subflows() {
  const [subflows, setSubflows] = useState<SubflowDefinition[]>([])
  const [filteredSubflows, setFilteredSubflows] = useState<SubflowDefinition[]>([])
  const [searchQuery, setSearchQuery] = useState('')
  const [selectedCategory, setSelectedCategory] = useState<string>('all')
  const [categories, setCategories] = useState<Record<string, number>>({})
  const [isLoading, setIsLoading] = useState(true)
  const [showNewDialog, setShowNewDialog] = useState(false)
  const [showImportDialog, setShowImportDialog] = useState(false)
  const [editingSubflow, setEditingSubflow] = useState<SubflowDefinition | null>(null)
  const [newSubflowName, setNewSubflowName] = useState('')
  const [newSubflowCategory, setNewSubflowCategory] = useState('general')
  useEffect(() => {
    loadSubflows()
    loadCategories()
  }, [])

  useEffect(() => {
    filterSubflows()
  }, [subflows, searchQuery, selectedCategory])

  const loadSubflows = async () => {
    try {
      setIsLoading(true)
      const response = await fetch('/api/subflows')
      const data = await response.json()
      setSubflows(data.subflows || [])
    } catch (error) {
      toast.error('Failed to load subflows')
    } finally {
      setIsLoading(false)
    }
  }

  const loadCategories = async () => {
    try {
      const response = await fetch('/api/subflows/library/categories')
      const data = await response.json()
      setCategories(data.categories || {})
    } catch (error) {
      console.error('Failed to load categories:', error)
    }
  }

  const filterSubflows = () => {
    let filtered = subflows

    if (searchQuery) {
      const query = searchQuery.toLowerCase()
      filtered = filtered.filter(
        (sf) =>
          sf.name.toLowerCase().includes(query) ||
          sf.description?.toLowerCase().includes(query) ||
          sf.category?.toLowerCase().includes(query)
      )
    }

    if (selectedCategory !== 'all') {
      filtered = filtered.filter((sf) => sf.category === selectedCategory)
    }

    setFilteredSubflows(filtered)
  }

  const handleCreateNew = async () => {
    if (!newSubflowName.trim()) {
      toast.error('Please enter a subflow name')
      return
    }

    const newSubflow: SubflowDefinition = {
      id: `subflow-${Date.now()}`,
      name: newSubflowName,
      category: newSubflowCategory,
      updatedAt: new Date().toISOString(),
      nodes: [],
      in: [{ type: 'input', index: 0 }],
      out: [{ type: 'output', index: 0 }],
    }

    try {
      const response = await fetch('/api/subflows', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(newSubflow),
      })

      if (response.ok) {
        toast.success('Subflow created successfully')
        setShowNewDialog(false)
        setNewSubflowName('')
        loadSubflows()
        loadCategories()
      }
    } catch (error) {
      toast.error('Failed to create subflow')
    }
  }

  const handleDelete = async (id: string) => {
    if (!confirm('Are you sure you want to delete this subflow?')) {
      return
    }

    try {
      const response = await fetch(`/api/subflows/${id}`, {
        method: 'DELETE',
      })

      if (response.ok) {
        toast.success('Subflow deleted successfully')
        loadSubflows()
        loadCategories()
      }
    } catch (error) {
      toast.error('Failed to delete subflow')
    }
  }

  const handleClone = async (id: string, name: string) => {
    try {
      const response = await fetch(`/api/subflows/${id}/clone`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          newId: `subflow-${Date.now()}`,
          newName: `${name} (Copy)`,
        }),
      })

      if (response.ok) {
        toast.success('Subflow cloned successfully')
        loadSubflows()
      }
    } catch (error) {
      toast.error('Failed to clone subflow')
    }
  }

  const handleExport = async (id: string, name: string) => {
    try {
      const response = await fetch(`/api/subflows/${id}/export`)
      const blob = await response.blob()
      const url = window.URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `${name}.json`
      a.click()
      window.URL.revokeObjectURL(url)
    } catch (error) {
      toast.error('Failed to export subflow')
    }
  }

  const handleImport = async (file: File) => {
    const formData = new FormData()
    formData.append('file', file)

    try {
      const response = await fetch('/api/subflows/import', {
        method: 'POST',
        body: formData,
      })

      if (response.ok) {
        toast.success('Subflow imported successfully')
        setShowImportDialog(false)
        loadSubflows()
        loadCategories()
      }
    } catch (error) {
      toast.error('Failed to import subflow')
    }
  }

  if (editingSubflow) {
    return (
      <SubflowEditor
        subflow={editingSubflow}
        onClose={() => {
          setEditingSubflow(null)
          loadSubflows()
        }}
      />
    )
  }

  return (
    <div className="h-full flex flex-col">
      {/* Header */}
      <div className="border-b bg-background px-6 py-4">
        <div className="flex items-center justify-between mb-4">
          <div>
            <h1 className="text-2xl font-bold">Subflows</h1>
            <p className="text-sm text-muted-foreground">
              Create and manage reusable flow components
            </p>
          </div>

          <div className="flex gap-2">
            <Button variant="outline" size="sm" onClick={() => setShowImportDialog(true)}>
              <Upload className="w-4 h-4 mr-2" />
              Import
            </Button>
            <Button size="sm" onClick={() => setShowNewDialog(true)}>
              <Plus className="w-4 h-4 mr-2" />
              New Subflow
            </Button>
          </div>
        </div>

        {/* Search and Filter */}
        <div className="flex gap-4">
          <div className="flex-1">
            <Input
              placeholder="Search subflows..."
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
            />
          </div>

          <div className="flex gap-2">
            <Button
              variant={selectedCategory === 'all' ? 'default' : 'outline'}
              size="sm"
              onClick={() => setSelectedCategory('all')}
            >
              All
            </Button>
            {Object.entries(categories).map(([cat, count]) => (
              <Button
                key={cat}
                variant={selectedCategory === cat ? 'default' : 'outline'}
                size="sm"
                onClick={() => setSelectedCategory(cat)}
              >
                {cat} ({count})
              </Button>
            ))}
          </div>
        </div>
      </div>

      {/* Subflow Grid */}
      <div className="flex-1 overflow-auto p-6">
        {isLoading ? (
          <div className="text-center py-12">
            <p className="text-muted-foreground">Loading subflows...</p>
          </div>
        ) : filteredSubflows.length === 0 ? (
          <div className="text-center py-12">
            <FolderOpen className="w-12 h-12 mx-auto mb-4 text-muted-foreground" />
            <p className="text-lg font-semibold mb-2">No subflows found</p>
            <p className="text-sm text-muted-foreground mb-4">
              Create a new subflow to get started
            </p>
            <Button onClick={() => setShowNewDialog(true)}>
              <Plus className="w-4 h-4 mr-2" />
              Create Subflow
            </Button>
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
            {filteredSubflows.map((subflow) => (
              <Card key={subflow.id} className="hover:shadow-lg transition-shadow">
                <CardHeader>
                  <div className="flex items-start justify-between">
                    <div className="flex-1">
                      <CardTitle className="text-lg">{subflow.name}</CardTitle>
                      {subflow.description && (
                        <CardDescription className="mt-1">{subflow.description}</CardDescription>
                      )}
                    </div>
                    {subflow.icon && (
                      <div
                        className="w-8 h-8 rounded flex items-center justify-center text-white"
                        style={{ backgroundColor: subflow.color || '#666' }}
                      >
                        {subflow.icon}
                      </div>
                    )}
                  </div>
                </CardHeader>

                <CardContent>
                  <div className="flex gap-2 flex-wrap mb-3">
                    {subflow.category && <Badge variant="secondary">{subflow.category}</Badge>}
                    {subflow.version && <Badge variant="outline">v{subflow.version}</Badge>}
                  </div>

                  <div className="text-xs text-muted-foreground space-y-1">
                    <div>Nodes: {subflow.nodes?.length || 0}</div>
                    <div>Inputs: {subflow.in?.length || 0}</div>
                    <div>Outputs: {subflow.out?.length || 0}</div>
                    {subflow.author && <div>By: {subflow.author}</div>}
                  </div>
                </CardContent>

                <CardFooter className="flex gap-2">
                  <Button
                    variant="default"
                    size="sm"
                    className="flex-1"
                    onClick={() => setEditingSubflow(subflow)}
                  >
                    <Edit className="w-4 h-4 mr-1" />
                    Edit
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => handleClone(subflow.id, subflow.name)}
                  >
                    <Copy className="w-4 h-4" />
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => handleExport(subflow.id, subflow.name)}
                  >
                    <Download className="w-4 h-4" />
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    onClick={() => handleDelete(subflow.id)}
                  >
                    <Trash2 className="w-4 h-4" />
                  </Button>
                </CardFooter>
              </Card>
            ))}
          </div>
        )}
      </div>

      {/* New Subflow Dialog */}
      <Dialog open={showNewDialog} onOpenChange={setShowNewDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create New Subflow</DialogTitle>
            <DialogDescription>
              Create a reusable subflow that can be used across multiple flows
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            <div>
              <Label htmlFor="name">Name</Label>
              <Input
                id="name"
                value={newSubflowName}
                onChange={(e) => setNewSubflowName(e.target.value)}
                placeholder="My Subflow"
              />
            </div>

            <div>
              <Label htmlFor="category">Category</Label>
              <Input
                id="category"
                value={newSubflowCategory}
                onChange={(e) => setNewSubflowCategory(e.target.value)}
                placeholder="general"
              />
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setShowNewDialog(false)}>
              Cancel
            </Button>
            <Button onClick={handleCreateNew}>Create</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Import Dialog */}
      <Dialog open={showImportDialog} onOpenChange={setShowImportDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Import Subflow</DialogTitle>
            <DialogDescription>Import a subflow from a JSON file</DialogDescription>
          </DialogHeader>

          <div>
            <Input
              type="file"
              accept=".json"
              onChange={(e) => {
                const file = e.target.files?.[0]
                if (file) {
                  handleImport(file)
                }
              }}
            />
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setShowImportDialog(false)}>
              Cancel
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
