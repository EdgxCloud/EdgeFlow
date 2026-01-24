/**
 * HTTP Request Builder
 *
 * Visual builder for HTTP requests with headers, body, and authentication
 */

import { useState } from 'react'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Plus, Trash2, Eye, EyeOff } from 'lucide-react'
import { JSONEditor } from '@/components/Common/JSONEditor'
import { cn } from '@/lib/utils'

export type HTTPMethod = 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE' | 'HEAD' | 'OPTIONS'
export type AuthType = 'none' | 'basic' | 'bearer' | 'api-key' | 'digest'
export type BodyType = 'none' | 'json' | 'form' | 'raw' | 'binary'

interface HTTPHeader {
  id: string
  key: string
  value: string
  enabled: boolean
}

interface HTTPRequestBuilderProps {
  value: {
    method?: HTTPMethod
    url?: string
    headers?: HTTPHeader[]
    authType?: AuthType
    authConfig?: Record<string, any>
    bodyType?: BodyType
    body?: any
    timeout?: number
    followRedirects?: boolean
    validateSSL?: boolean
  }
  onChange: (value: any) => void
  disabled?: boolean
}

const COMMON_HEADERS = [
  'Content-Type',
  'Accept',
  'Authorization',
  'User-Agent',
  'Cache-Control',
  'Cookie',
  'Referer',
  'Accept-Language',
  'Accept-Encoding',
]

export function HTTPRequestBuilder({ value, onChange, disabled = false }: HTTPRequestBuilderProps) {
  // Handle undefined or null value
  const safeValue = value || {}
  const method = safeValue.method || 'GET'
  const url = safeValue.url || ''
  const headers = safeValue.headers || []
  const authType = safeValue.authType || 'none'
  const authConfig = safeValue.authConfig || {}
  const bodyType = safeValue.bodyType || 'none'
  const body = safeValue.body || ''
  const timeout = safeValue.timeout ?? 30000
  const followRedirects = safeValue.followRedirects ?? true
  const validateSSL = safeValue.validateSSL ?? true

  const [showPassword, setShowPassword] = useState(false)

  const generateHeaderId = (): string => {
    return `header_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`
  }

  const addHeader = () => {
    const newHeader: HTTPHeader = {
      id: generateHeaderId(),
      key: '',
      value: '',
      enabled: true,
    }
    onChange({ ...safeValue, headers: [...headers, newHeader] })
  }

  const removeHeader = (id: string) => {
    onChange({ ...safeValue, headers: headers.filter((h) => h.id !== id) })
  }

  const updateHeader = (id: string, updates: Partial<HTTPHeader>) => {
    const newHeaders = headers.map((h) => (h.id === id ? { ...h, ...updates } : h))
    onChange({ ...safeValue, headers: newHeaders })
  }

  const canHaveBody = ['POST', 'PUT', 'PATCH'].includes(method)

  return (
    <div className="space-y-6">
      {/* Request Configuration */}
      <div className="space-y-4">
        <div className="grid grid-cols-4 gap-4">
          <div className="col-span-1 space-y-2">
            <Label htmlFor="method" className="text-sm font-semibold">
              Method
            </Label>
            <Select
              value={method}
              onValueChange={(value: HTTPMethod) => onChange({ ...safeValue, method: value })}
              disabled={disabled}
            >
              <SelectTrigger className="h-11" id="method">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="GET">GET</SelectItem>
                <SelectItem value="POST">POST</SelectItem>
                <SelectItem value="PUT">PUT</SelectItem>
                <SelectItem value="PATCH">PATCH</SelectItem>
                <SelectItem value="DELETE">DELETE</SelectItem>
                <SelectItem value="HEAD">HEAD</SelectItem>
                <SelectItem value="OPTIONS">OPTIONS</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="col-span-3 space-y-2">
            <Label htmlFor="url" className="text-sm font-semibold">
              URL
            </Label>
            <Input
              id="url"
              type="url"
              value={url}
              onChange={(e) => onChange({ ...safeValue, url: e.target.value })}
              placeholder="https://api.example.com/endpoint"
              className="h-11 font-mono"
              disabled={disabled}
            />
          </div>
        </div>
      </div>

      {/* Tabs for Headers, Auth, Body, Settings */}
      <Tabs defaultValue="headers" className="w-full">
        <TabsList className="grid w-full grid-cols-4">
          <TabsTrigger value="headers">Headers ({headers.filter((h) => h.enabled).length})</TabsTrigger>
          <TabsTrigger value="auth">Auth</TabsTrigger>
          <TabsTrigger value="body">Body</TabsTrigger>
          <TabsTrigger value="settings">Settings</TabsTrigger>
        </TabsList>

        {/* Headers Tab */}
        <TabsContent value="headers" className="space-y-3 mt-4">
          <div className="flex items-center justify-between">
            <Label className="text-sm font-semibold">Headers</Label>
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={addHeader}
              className="h-8"
              disabled={disabled}
            >
              <Plus className="w-4 h-4 mr-1" />
              Add Header
            </Button>
          </div>

          {headers.length === 0 && (
            <div className="text-center py-8 border-2 border-dashed rounded-lg">
              <p className="text-sm text-muted-foreground">
                No headers defined. Add custom headers if needed.
              </p>
            </div>
          )}

          <div className="space-y-2">
            {headers.map((header) => (
              <div key={header.id} className="flex items-center gap-2">
                <input
                  type="checkbox"
                  checked={header.enabled}
                  onChange={(e) => updateHeader(header.id, { enabled: e.target.checked })}
                  disabled={disabled}
                  className="rounded flex-shrink-0"
                />

                <Input
                  value={header.key}
                  onChange={(e) => updateHeader(header.id, { key: e.target.value })}
                  placeholder="Header name"
                  list="common-headers"
                  className="h-9 font-mono text-xs flex-1"
                  disabled={disabled || !header.enabled}
                />

                <Input
                  value={header.value}
                  onChange={(e) => updateHeader(header.id, { value: e.target.value })}
                  placeholder="Value"
                  className="h-9 font-mono text-xs flex-1"
                  disabled={disabled || !header.enabled}
                />

                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  onClick={() => removeHeader(header.id)}
                  className="h-9 w-9 p-0 flex-shrink-0"
                  disabled={disabled}
                >
                  <Trash2 className="w-4 h-4" />
                </Button>
              </div>
            ))}
          </div>

          <datalist id="common-headers">
            {COMMON_HEADERS.map((h) => (
              <option key={h} value={h} />
            ))}
          </datalist>
        </TabsContent>

        {/* Auth Tab */}
        <TabsContent value="auth" className="space-y-4 mt-4">
          <div className="space-y-2">
            <Label htmlFor="authType" className="text-sm font-semibold">
              Authentication Type
            </Label>
            <Select
              value={authType}
              onValueChange={(value: AuthType) =>
                onChange({ ...safeValue, authType: value, authConfig: {} })
              }
              disabled={disabled}
            >
              <SelectTrigger className="h-11" id="authType">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="none">No Authentication</SelectItem>
                <SelectItem value="basic">Basic Auth</SelectItem>
                <SelectItem value="bearer">Bearer Token</SelectItem>
                <SelectItem value="api-key">API Key</SelectItem>
                <SelectItem value="digest">Digest Auth</SelectItem>
              </SelectContent>
            </Select>
          </div>

          {/* Basic Auth */}
          {authType === 'basic' && (
            <div className="space-y-3">
              <div className="space-y-2">
                <Label htmlFor="username" className="text-xs">
                  Username
                </Label>
                <Input
                  id="username"
                  value={authConfig.username || ''}
                  onChange={(e) =>
                    onChange({ ...safeValue, authConfig: { ...authConfig, username: e.target.value } })
                  }
                  className="h-11"
                  disabled={disabled}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="password" className="text-xs">
                  Password
                </Label>
                <div className="relative">
                  <Input
                    id="password"
                    type={showPassword ? 'text' : 'password'}
                    value={authConfig.password || ''}
                    onChange={(e) =>
                      onChange({ ...safeValue, authConfig: { ...authConfig, password: e.target.value } })
                    }
                    className="h-11 pr-10"
                    disabled={disabled}
                  />
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    onClick={() => setShowPassword(!showPassword)}
                    className="absolute right-0 top-0 h-11 px-3"
                  >
                    {showPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                  </Button>
                </div>
              </div>
            </div>
          )}

          {/* Bearer Token */}
          {authType === 'bearer' && (
            <div className="space-y-2">
              <Label htmlFor="token" className="text-xs">
                Bearer Token
              </Label>
              <Input
                id="token"
                type={showPassword ? 'text' : 'password'}
                value={authConfig.token || ''}
                onChange={(e) =>
                  onChange({ ...safeValue, authConfig: { ...authConfig, token: e.target.value } })
                }
                className="h-11 font-mono"
                disabled={disabled}
              />
            </div>
          )}

          {/* API Key */}
          {authType === 'api-key' && (
            <div className="space-y-3">
              <div className="space-y-2">
                <Label htmlFor="keyName" className="text-xs">
                  Key Name
                </Label>
                <Input
                  id="keyName"
                  value={authConfig.keyName || 'X-API-Key'}
                  onChange={(e) =>
                    onChange({ ...safeValue, authConfig: { ...authConfig, keyName: e.target.value } })
                  }
                  className="h-11"
                  disabled={disabled}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="keyValue" className="text-xs">
                  Key Value
                </Label>
                <Input
                  id="keyValue"
                  type={showPassword ? 'text' : 'password'}
                  value={authConfig.keyValue || ''}
                  onChange={(e) =>
                    onChange({ ...safeValue, authConfig: { ...authConfig, keyValue: e.target.value } })
                  }
                  className="h-11 font-mono"
                  disabled={disabled}
                />
              </div>
              <div className="space-y-2">
                <Label htmlFor="addTo" className="text-xs">
                  Add To
                </Label>
                <Select
                  value={authConfig.addTo || 'header'}
                  onValueChange={(val) =>
                    onChange({ ...safeValue, authConfig: { ...authConfig, addTo: val } })
                  }
                  disabled={disabled}
                >
                  <SelectTrigger className="h-11" id="addTo">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="header">Header</SelectItem>
                    <SelectItem value="query">Query Parameter</SelectItem>
                  </SelectContent>
                </Select>
              </div>
            </div>
          )}
        </TabsContent>

        {/* Body Tab */}
        <TabsContent value="body" className="space-y-4 mt-4">
          {!canHaveBody && (
            <div className="p-4 bg-muted rounded-lg border">
              <p className="text-sm text-muted-foreground">
                {method} requests typically don't have a body. Use POST, PUT, or PATCH for request bodies.
              </p>
            </div>
          )}

          {canHaveBody && (
            <>
              <div className="space-y-2">
                <Label htmlFor="bodyType" className="text-sm font-semibold">
                  Body Type
                </Label>
                <Select
                  value={bodyType}
                  onValueChange={(value: BodyType) => onChange({ ...safeValue, bodyType: value })}
                  disabled={disabled}
                >
                  <SelectTrigger className="h-11" id="bodyType">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="none">None</SelectItem>
                    <SelectItem value="json">JSON</SelectItem>
                    <SelectItem value="form">Form Data</SelectItem>
                    <SelectItem value="raw">Raw Text</SelectItem>
                    <SelectItem value="binary">Binary</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              {bodyType === 'json' && (
                <div className="space-y-2">
                  <Label className="text-sm font-semibold">JSON Body</Label>
                  <JSONEditor
                    value={typeof body === 'string' ? JSON.parse(body || '{}') : body}
                    onChange={(val) => onChange({ ...safeValue, body: val })}
                    height={200}
                    showValidation={true}
                    disabled={disabled}
                  />
                </div>
              )}

              {bodyType === 'raw' && (
                <div className="space-y-2">
                  <Label htmlFor="rawBody" className="text-sm font-semibold">
                    Raw Body
                  </Label>
                  <Textarea
                    id="rawBody"
                    value={body || ''}
                    onChange={(e) => onChange({ ...safeValue, body: e.target.value })}
                    rows={8}
                    className="font-mono text-xs"
                    disabled={disabled}
                  />
                </div>
              )}

              {bodyType === 'form' && (
                <div className="space-y-2">
                  <Label className="text-sm font-semibold">Form Data</Label>
                  <p className="text-xs text-muted-foreground">
                    Use JSON format with key-value pairs
                  </p>
                  <JSONEditor
                    value={typeof body === 'string' ? JSON.parse(body || '{}') : body}
                    onChange={(val) => onChange({ ...safeValue, body: val })}
                    height={200}
                    showValidation={true}
                    disabled={disabled}
                  />
                </div>
              )}
            </>
          )}
        </TabsContent>

        {/* Settings Tab */}
        <TabsContent value="settings" className="space-y-4 mt-4">
          <div className="space-y-2">
            <Label htmlFor="timeout" className="text-sm font-semibold">
              Timeout (ms)
            </Label>
            <Input
              id="timeout"
              type="number"
              value={timeout}
              onChange={(e) => onChange({ ...safeValue, timeout: Number(e.target.value) })}
              min={1000}
              max={300000}
              className="h-11"
              disabled={disabled}
            />
            <p className="text-xs text-muted-foreground">
              Request timeout in milliseconds (1s - 300s)
            </p>
          </div>

          <div className="flex items-center space-x-2">
            <Switch
              id="followRedirects"
              checked={followRedirects}
              onCheckedChange={(checked) => onChange({ ...safeValue, followRedirects: checked })}
              disabled={disabled}
            />
            <Label htmlFor="followRedirects" className="text-sm font-normal cursor-pointer">
              Follow Redirects
            </Label>
          </div>

          <div className="flex items-center space-x-2">
            <Switch
              id="validateSSL"
              checked={validateSSL}
              onCheckedChange={(checked) => onChange({ ...safeValue, validateSSL: checked })}
              disabled={disabled}
            />
            <Label htmlFor="validateSSL" className="text-sm font-normal cursor-pointer">
              Validate SSL Certificates
            </Label>
          </div>
        </TabsContent>
      </Tabs>
    </div>
  )
}
