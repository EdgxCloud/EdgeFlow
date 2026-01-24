/**
 * Module Marketplace Dialog
 * Provides search, upload, and direct URL installation for modules
 */

import React, { useState, useEffect, useCallback } from 'react'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Progress } from '@/components/ui/progress'
import {
  Search,
  Upload,
  Link,
  Package,
  Star,
  Download,
  AlertCircle,
  Loader2,
  ExternalLink,
} from 'lucide-react'
import { modulesApi, MarketplaceSearchResult } from '@/lib/api'
import { useDebounce } from '@/hooks/useDebounce'

interface ModuleMarketplaceDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  onModuleInstalled?: () => void
}

type SearchSource = 'npm' | 'node-red' | 'github'

export default function ModuleMarketplaceDialog({
  open,
  onOpenChange,
  onModuleInstalled,
}: ModuleMarketplaceDialogProps) {
  // Tab state
  const [activeTab, setActiveTab] = useState('search')

  // Search state
  const [searchQuery, setSearchQuery] = useState('')
  const [searchSource, setSearchSource] = useState<SearchSource>('npm')
  const [searchResults, setSearchResults] = useState<MarketplaceSearchResult[]>([])
  const [isSearching, setIsSearching] = useState(false)

  // Install state
  const [isInstalling, setIsInstalling] = useState(false)
  const [installProgress, setInstallProgress] = useState(0)
  const [installingPackage, setInstallingPackage] = useState<string | null>(null)

  // URL tab state
  const [urlInput, setUrlInput] = useState('')
  const [urlType, setUrlType] = useState<'url' | 'npm' | 'github'>('npm')

  // Upload tab state
  const [selectedFile, setSelectedFile] = useState<File | null>(null)
  const [isUploading, setIsUploading] = useState(false)
  const [uploadProgress, setUploadProgress] = useState(0)

  // Error state
  const [error, setError] = useState<string | null>(null)
  const [successMessage, setSuccessMessage] = useState<string | null>(null)

  // Debounced search query
  const debouncedQuery = useDebounce(searchQuery, 500)

  // Perform search when query changes
  useEffect(() => {
    if (debouncedQuery.length >= 2) {
      performSearch()
    } else {
      setSearchResults([])
    }
  }, [debouncedQuery, searchSource])

  const performSearch = useCallback(async () => {
    setIsSearching(true)
    setError(null)

    try {
      let response
      switch (searchSource) {
        case 'npm':
          response = await modulesApi.searchNpm(debouncedQuery)
          break
        case 'node-red':
          response = await modulesApi.searchNodeRed(debouncedQuery)
          break
        case 'github':
          response = await modulesApi.searchGitHub(debouncedQuery)
          break
      }
      setSearchResults(response.data.results || [])
    } catch (err: any) {
      setError(err.response?.data?.error || 'Search failed')
      setSearchResults([])
    } finally {
      setIsSearching(false)
    }
  }, [debouncedQuery, searchSource])

  const handleInstallFromSearch = async (result: MarketplaceSearchResult) => {
    setIsInstalling(true)
    setInstallingPackage(result.name)
    setInstallProgress(0)
    setError(null)
    setSuccessMessage(null)

    try {
      // Simulate progress
      const progressInterval = setInterval(() => {
        setInstallProgress((prev) => Math.min(prev + 10, 90))
      }, 300)

      let installRequest
      switch (result.source) {
        case 'npm':
        case 'node-red':
          installRequest = { npm: result.name }
          break
        case 'github':
          installRequest = { github: result.name }
          break
      }

      const response = await modulesApi.install(installRequest)

      clearInterval(progressInterval)
      setInstallProgress(100)

      setSuccessMessage(`${response.data.module?.name || result.name} installed successfully!`)
      onModuleInstalled?.()
    } catch (err: any) {
      setError(err.response?.data?.error || 'Installation failed')
    } finally {
      setIsInstalling(false)
      setInstallingPackage(null)
      setTimeout(() => setInstallProgress(0), 1000)
    }
  }

  const handleInstallFromURL = async () => {
    if (!urlInput.trim()) {
      setError('Please enter a URL, npm package, or GitHub repository')
      return
    }

    setIsInstalling(true)
    setInstallProgress(0)
    setError(null)
    setSuccessMessage(null)

    try {
      const progressInterval = setInterval(() => {
        setInstallProgress((prev) => Math.min(prev + 10, 90))
      }, 300)

      let installRequest
      switch (urlType) {
        case 'url':
          installRequest = { url: urlInput }
          break
        case 'npm':
          installRequest = { npm: urlInput }
          break
        case 'github':
          installRequest = { github: urlInput }
          break
      }

      const response = await modulesApi.install(installRequest)

      clearInterval(progressInterval)
      setInstallProgress(100)

      setSuccessMessage(`${response.data.module?.name || urlInput} installed successfully!`)
      setUrlInput('')
      onModuleInstalled?.()
    } catch (err: any) {
      setError(err.response?.data?.error || 'Installation failed')
    } finally {
      setIsInstalling(false)
      setTimeout(() => setInstallProgress(0), 1000)
    }
  }

  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0]
    if (file) {
      // Validate file type
      const validTypes = ['.zip', '.tgz', '.tar.gz']
      const isValid = validTypes.some((ext) => file.name.toLowerCase().endsWith(ext))

      if (!isValid) {
        setError('Invalid file type. Please upload a .zip or .tgz file')
        return
      }

      setSelectedFile(file)
      setError(null)
    }
  }

  const handleUpload = async () => {
    if (!selectedFile) {
      setError('Please select a file to upload')
      return
    }

    setIsUploading(true)
    setUploadProgress(0)
    setError(null)
    setSuccessMessage(null)

    try {
      const progressInterval = setInterval(() => {
        setUploadProgress((prev) => Math.min(prev + 15, 90))
      }, 200)

      const response = await modulesApi.upload(selectedFile)

      clearInterval(progressInterval)
      setUploadProgress(100)

      setSuccessMessage(`${response.data.module?.name || selectedFile.name} uploaded and installed!`)
      setSelectedFile(null)
      onModuleInstalled?.()
    } catch (err: any) {
      setError(err.response?.data?.error || 'Upload failed')
    } finally {
      setIsUploading(false)
      setTimeout(() => setUploadProgress(0), 1000)
    }
  }

  const getSourceIcon = (source: SearchSource) => {
    switch (source) {
      case 'npm':
        return <Package className="w-4 h-4 text-red-500" />
      case 'node-red':
        return <Package className="w-4 h-4 text-red-600" />
      case 'github':
        return <svg className="w-4 h-4" viewBox="0 0 24 24" fill="currentColor"><path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/></svg>
    }
  }

  // Reset state when dialog closes
  useEffect(() => {
    if (!open) {
      setError(null)
      setSuccessMessage(null)
      setInstallProgress(0)
      setUploadProgress(0)
    }
  }, [open])

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[700px] max-h-[85vh]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Package className="h-5 w-5" />
            Module Marketplace
          </DialogTitle>
          <DialogDescription>
            Search and install modules from npm, Node-RED catalog, or GitHub
          </DialogDescription>
        </DialogHeader>

        {/* Success Message */}
        {successMessage && (
          <Alert className="bg-green-50 border-green-200 dark:bg-green-900/20">
            <AlertDescription className="text-green-800 dark:text-green-200">
              {successMessage}
            </AlertDescription>
          </Alert>
        )}

        {/* Error Message */}
        {error && (
          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>{error}</AlertDescription>
          </Alert>
        )}

        <Tabs value={activeTab} onValueChange={setActiveTab}>
          <TabsList className="grid w-full grid-cols-3">
            <TabsTrigger value="search" className="flex items-center gap-2">
              <Search className="w-4 h-4" />
              Search
            </TabsTrigger>
            <TabsTrigger value="upload" className="flex items-center gap-2">
              <Upload className="w-4 h-4" />
              Upload
            </TabsTrigger>
            <TabsTrigger value="url" className="flex items-center gap-2">
              <Link className="w-4 h-4" />
              URL / NPM
            </TabsTrigger>
          </TabsList>

          {/* Search Tab */}
          <TabsContent value="search" className="space-y-4">
            {/* Search Input */}
            <div className="flex gap-2">
              <div className="flex-1 relative">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-muted-foreground" />
                <Input
                  placeholder="Search for modules..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="pl-10"
                />
              </div>
            </div>

            {/* Source Selection */}
            <div className="flex gap-2">
              {(['npm', 'node-red', 'github'] as SearchSource[]).map((source) => (
                <Button
                  key={source}
                  variant={searchSource === source ? 'default' : 'outline'}
                  size="sm"
                  onClick={() => setSearchSource(source)}
                  className="flex items-center gap-2"
                >
                  {getSourceIcon(source)}
                  {source === 'node-red' ? 'Node-RED' : source === 'npm' ? 'NPM' : 'GitHub'}
                </Button>
              ))}
            </div>

            {/* Search Results */}
            <ScrollArea className="h-[320px] pr-4">
              {isSearching ? (
                <div className="flex items-center justify-center py-8">
                  <Loader2 className="w-6 h-6 animate-spin" />
                  <span className="ml-2">Searching...</span>
                </div>
              ) : searchResults.length > 0 ? (
                <div className="space-y-2">
                  {searchResults.map((result, index) => (
                    <SearchResultCard
                      key={`${result.name}-${index}`}
                      result={result}
                      onInstall={() => handleInstallFromSearch(result)}
                      isInstalling={isInstalling && installingPackage === result.name}
                    />
                  ))}
                </div>
              ) : searchQuery.length >= 2 ? (
                <div className="text-center py-8 text-muted-foreground">
                  No modules found for "{searchQuery}"
                </div>
              ) : (
                <div className="text-center py-8 text-muted-foreground">
                  <Package className="w-12 h-12 mx-auto mb-4 opacity-50" />
                  <p>Enter a search term to find modules</p>
                  <p className="text-sm mt-2">
                    Search npm, Node-RED catalog, or GitHub repositories
                  </p>
                </div>
              )}
            </ScrollArea>

            {installProgress > 0 && (
              <Progress value={installProgress} className="mt-4" />
            )}
          </TabsContent>

          {/* Upload Tab */}
          <TabsContent value="upload" className="space-y-4">
            <div className="border-2 border-dashed rounded-lg p-8 text-center">
              <Upload className="w-12 h-12 mx-auto mb-4 text-muted-foreground" />
              <p className="text-sm text-muted-foreground mb-4">
                Upload a module package (.zip or .tgz)
              </p>
              <input
                type="file"
                accept=".zip,.tgz,.tar.gz"
                onChange={handleFileSelect}
                className="hidden"
                id="module-file"
              />
              <Label htmlFor="module-file" className="cursor-pointer">
                <Button variant="outline" asChild>
                  <span>Choose File</span>
                </Button>
              </Label>

              {selectedFile && (
                <div className="mt-4 p-3 bg-muted rounded-lg">
                  <div className="flex items-center justify-between">
                    <span className="text-sm font-medium">{selectedFile.name}</span>
                    <span className="text-xs text-muted-foreground">
                      {(selectedFile.size / 1024).toFixed(1)} KB
                    </span>
                  </div>
                </div>
              )}

              {uploadProgress > 0 && (
                <div className="mt-4">
                  <Progress value={uploadProgress} />
                </div>
              )}
            </div>

            <Button
              onClick={handleUpload}
              disabled={!selectedFile || isUploading}
              className="w-full"
            >
              {isUploading ? (
                <>
                  <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                  Uploading...
                </>
              ) : (
                <>
                  <Upload className="mr-2 h-4 w-4" />
                  Upload & Install
                </>
              )}
            </Button>
          </TabsContent>

          {/* URL Tab */}
          <TabsContent value="url" className="space-y-4">
            <div className="space-y-4">
              {/* URL Type Selection */}
              <div className="flex gap-2">
                {(['npm', 'github', 'url'] as const).map((type) => (
                  <Button
                    key={type}
                    variant={urlType === type ? 'default' : 'outline'}
                    size="sm"
                    onClick={() => setUrlType(type)}
                  >
                    {type === 'url' && <Link className="w-4 h-4 mr-2" />}
                    {type === 'npm' && <Package className="w-4 h-4 mr-2" />}
                    {type === 'github' && (
                      <svg className="w-4 h-4 mr-2" viewBox="0 0 24 24" fill="currentColor"><path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z"/></svg>
                    )}
                    {type.toUpperCase()}
                  </Button>
                ))}
              </div>

              {/* Input */}
              <div className="space-y-2">
                <Label>
                  {urlType === 'url' && 'Direct URL'}
                  {urlType === 'npm' && 'NPM Package Name'}
                  {urlType === 'github' && 'GitHub Repository'}
                </Label>
                <Input
                  placeholder={
                    urlType === 'url'
                      ? 'https://example.com/module.zip'
                      : urlType === 'npm'
                      ? 'node-red-contrib-telegram'
                      : 'owner/repository'
                  }
                  value={urlInput}
                  onChange={(e) => setUrlInput(e.target.value)}
                />
                <p className="text-xs text-muted-foreground">
                  {urlType === 'url' && 'Enter a direct URL to a .zip or .tgz file'}
                  {urlType === 'npm' && 'Enter an npm package name (e.g., node-red-contrib-telegram)'}
                  {urlType === 'github' && 'Enter owner/repo (e.g., node-red/node-red-dashboard)'}
                </p>
              </div>

              {installProgress > 0 && (
                <Progress value={installProgress} />
              )}

              <Button
                onClick={handleInstallFromURL}
                disabled={!urlInput.trim() || isInstalling}
                className="w-full"
              >
                {isInstalling ? (
                  <>
                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                    Installing...
                  </>
                ) : (
                  <>
                    <Download className="mr-2 h-4 w-4" />
                    Install Module
                  </>
                )}
              </Button>
            </div>
          </TabsContent>
        </Tabs>
      </DialogContent>
    </Dialog>
  )
}

// Search Result Card Component
interface SearchResultCardProps {
  result: MarketplaceSearchResult
  onInstall: () => void
  isInstalling: boolean
}

function SearchResultCard({ result, onInstall, isInstalling }: SearchResultCardProps) {
  return (
    <div className="p-3 rounded-lg border hover:bg-accent/50 transition-colors">
      <div className="flex items-start justify-between gap-4">
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 flex-wrap">
            <h4 className="font-medium truncate">{result.name}</h4>
            {result.version && (
              <Badge variant="secondary" className="text-xs">
                v{result.version}
              </Badge>
            )}
          </div>
          <p className="text-sm text-muted-foreground line-clamp-2 mt-1">
            {result.description || 'No description available'}
          </p>
          <div className="flex items-center gap-3 mt-2 text-xs text-muted-foreground flex-wrap">
            {result.stars !== undefined && (
              <span className="flex items-center gap-1">
                <Star className="w-3 h-3" />
                {result.stars.toLocaleString()}
              </span>
            )}
            {result.downloads !== undefined && (
              <span className="flex items-center gap-1">
                <Download className="w-3 h-3" />
                {result.downloads.toLocaleString()}
              </span>
            )}
            {result.author && (
              <span>by {result.author}</span>
            )}
            {result.owner && !result.author && (
              <span>by {result.owner}</span>
            )}
            <Badge variant="outline" className="text-xs">
              {result.source}
            </Badge>
          </div>
        </div>
        <div className="flex items-center gap-2 flex-shrink-0">
          {result.url && (
            <Button
              variant="ghost"
              size="icon"
              onClick={() => window.open(result.url, '_blank')}
              title="Open in new tab"
            >
              <ExternalLink className="w-4 h-4" />
            </Button>
          )}
          <Button
            size="sm"
            onClick={onInstall}
            disabled={isInstalling}
          >
            {isInstalling ? (
              <Loader2 className="w-4 h-4 animate-spin" />
            ) : (
              <>
                <Download className="w-4 h-4 mr-2" />
                Install
              </>
            )}
          </Button>
        </div>
      </div>
    </div>
  )
}
