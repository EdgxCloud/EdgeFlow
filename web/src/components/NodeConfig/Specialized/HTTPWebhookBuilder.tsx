/**
 * HTTP Webhook Builder
 *
 * Comprehensive HTTP Webhook/Endpoint configuration editor based on Node-RED and n8n patterns
 * Supports URL path, HTTP methods, authentication, CORS, and response configuration
 */

import { useState } from 'react'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'
import { Switch } from '@/components/ui/switch'
import { Textarea } from '@/components/ui/textarea'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from '@/components/ui/accordion'
import { Globe, Shield, Settings2, Send, Eye, EyeOff, Copy, Check, Plus, Trash2 } from 'lucide-react'
import { cn } from '@/lib/utils'

// HTTP Methods supported for webhooks
export type WebhookMethod = 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE' | 'OPTIONS' | 'HEAD' | 'ALL'

// Authentication types
export type WebhookAuthType = 'none' | 'basic' | 'header' | 'jwt' | 'apiKey'

// Response content types
export type ResponseContentType = 'application/json' | 'text/plain' | 'text/html' | 'application/xml' | 'auto'

// Response mode
export type ResponseMode = 'onReceived' | 'lastNode' | 'responseNode'

interface HeaderItem {
  id: string
  key: string
  value: string
}

interface WebhookConfig {
  // Endpoint configuration
  path?: string
  method?: WebhookMethod
  // Authentication
  authType?: WebhookAuthType
  authConfig?: {
    username?: string
    password?: string
    headerName?: string
    headerValue?: string
    jwtSecret?: string
    jwtAlgorithm?: 'HS256' | 'HS384' | 'HS512' | 'RS256' | 'RS384' | 'RS512'
    apiKeyName?: string
    apiKeyValue?: string
    apiKeyLocation?: 'header' | 'query'
  }
  // Request options
  acceptFileUploads?: boolean
  rawBody?: boolean
  maxBodySize?: string
  // Response configuration
  responseMode?: ResponseMode
  responseContentType?: ResponseContentType
  responseStatusCode?: number
  responseHeaders?: HeaderItem[]
  noResponseBody?: boolean
  // Security
  ipWhitelist?: string
  rateLimitEnabled?: boolean
  rateLimitRequests?: number
  rateLimitWindow?: number
  // CORS
  corsEnabled?: boolean
  corsOrigin?: string
  corsMethods?: string
  corsHeaders?: string
  corsCredentials?: boolean
}

interface HTTPWebhookBuilderProps {
  value: WebhookConfig
  onChange: (value: WebhookConfig) => void
  disabled?: boolean
}

// Common webhook path presets
const PATH_PRESETS = [
  { name: 'Generic', path: '/webhook' },
  { name: 'API', path: '/api/v1/webhook' },
  { name: 'Events', path: '/events' },
  { name: 'Notifications', path: '/notifications' },
  { name: 'Callbacks', path: '/callback' },
]

export function HTTPWebhookBuilder({
  value,
  onChange,
  disabled = false,
}: HTTPWebhookBuilderProps) {
  // Handle undefined or null value
  const safeValue = value || {}

  // Endpoint configuration
  const path = safeValue.path || '/webhook'
  const method = safeValue.method || 'POST'

  // Authentication
  const authType = safeValue.authType || 'none'
  const authConfig = safeValue.authConfig || {}

  // Request options
  const acceptFileUploads = safeValue.acceptFileUploads ?? false
  const rawBody = safeValue.rawBody ?? false
  const maxBodySize = safeValue.maxBodySize || '1mb'

  // Response configuration
  const responseMode = safeValue.responseMode || 'onReceived'
  const responseContentType = safeValue.responseContentType || 'application/json'
  const responseStatusCode = safeValue.responseStatusCode ?? 200
  const responseHeaders = safeValue.responseHeaders || []
  const noResponseBody = safeValue.noResponseBody ?? false

  // Security
  const ipWhitelist = safeValue.ipWhitelist || ''
  const rateLimitEnabled = safeValue.rateLimitEnabled ?? false
  const rateLimitRequests = safeValue.rateLimitRequests ?? 100
  const rateLimitWindow = safeValue.rateLimitWindow ?? 60

  // CORS
  const corsEnabled = safeValue.corsEnabled ?? false
  const corsOrigin = safeValue.corsOrigin || '*'
  const corsMethods = safeValue.corsMethods || 'GET,POST,PUT,DELETE,OPTIONS'
  const corsHeaders = safeValue.corsHeaders || 'Content-Type,Authorization'
  const corsCredentials = safeValue.corsCredentials ?? false

  // UI state
  const [showPassword, setShowPassword] = useState(false)
  const [copied, setCopied] = useState(false)
  const [activeAccordion, setActiveAccordion] = useState<string[]>(['endpoint'])

  // Generate full webhook URL (preview)
  const baseUrl = typeof window !== 'undefined' ? window.location.origin : 'http://localhost:8080'
  const fullWebhookUrl = `${baseUrl}/api/v1/webhooks${path.startsWith('/') ? path : '/' + path}`

  // Copy URL to clipboard
  const copyUrl = async () => {
    try {
      await navigator.clipboard.writeText(fullWebhookUrl)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch (err) {
      console.error('Failed to copy URL:', err)
    }
  }

  // Update auth config
  const updateAuthConfig = (field: string, fieldValue: any) => {
    onChange({
      ...safeValue,
      authConfig: {
        ...authConfig,
        [field]: fieldValue,
      },
    })
  }

  // Add response header
  const addResponseHeader = () => {
    const newHeaders = [
      ...responseHeaders,
      { id: `${Date.now()}`, key: '', value: '' },
    ]
    onChange({ ...safeValue, responseHeaders: newHeaders })
  }

  // Update response header
  const updateResponseHeader = (id: string, field: 'key' | 'value', fieldValue: string) => {
    const newHeaders = responseHeaders.map((h) =>
      h.id === id ? { ...h, [field]: fieldValue } : h
    )
    onChange({ ...safeValue, responseHeaders: newHeaders })
  }

  // Remove response header
  const removeResponseHeader = (id: string) => {
    const newHeaders = responseHeaders.filter((h) => h.id !== id)
    onChange({ ...safeValue, responseHeaders: newHeaders })
  }

  return (
    <div className="space-y-4">
      <Accordion
        type="multiple"
        value={activeAccordion}
        onValueChange={setActiveAccordion}
        className="space-y-2"
      >
        {/* Endpoint Configuration Section */}
        <AccordionItem value="endpoint" className="border rounded-lg px-4">
          <AccordionTrigger className="hover:no-underline">
            <div className="flex items-center gap-2">
              <Globe className="w-4 h-4 text-blue-500" />
              <span className="font-semibold">Endpoint Configuration</span>
              {path && (
                <code className="text-xs bg-muted px-2 py-0.5 rounded ml-2">
                  {method} {path}
                </code>
              )}
            </div>
          </AccordionTrigger>
          <AccordionContent className="space-y-4 pt-4">
            {/* Path Presets */}
            <div className="space-y-2">
              <Label className="text-xs text-muted-foreground">Quick Paths</Label>
              <div className="flex flex-wrap gap-2">
                {PATH_PRESETS.map((preset) => (
                  <Button
                    key={preset.name}
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={() => onChange({ ...safeValue, path: preset.path })}
                    className="h-7 text-xs"
                    disabled={disabled}
                  >
                    {preset.name}
                  </Button>
                ))}
              </div>
            </div>

            {/* HTTP Method */}
            <div className="space-y-2">
              <Label htmlFor="method" className="text-sm font-semibold">
                HTTP Method
              </Label>
              <Select
                value={method}
                onValueChange={(val) => onChange({ ...safeValue, method: val as WebhookMethod })}
                disabled={disabled}
              >
                <SelectTrigger className="h-11" id="method">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="GET">GET - Retrieve data</SelectItem>
                  <SelectItem value="POST">POST - Submit data</SelectItem>
                  <SelectItem value="PUT">PUT - Update/Replace</SelectItem>
                  <SelectItem value="PATCH">PATCH - Partial update</SelectItem>
                  <SelectItem value="DELETE">DELETE - Remove data</SelectItem>
                  <SelectItem value="OPTIONS">OPTIONS - Preflight</SelectItem>
                  <SelectItem value="HEAD">HEAD - Headers only</SelectItem>
                  <SelectItem value="ALL">ALL - Any method</SelectItem>
                </SelectContent>
              </Select>
            </div>

            {/* URL Path */}
            <div className="space-y-2">
              <Label htmlFor="path" className="text-sm font-semibold">
                URL Path
              </Label>
              <Input
                id="path"
                value={path}
                onChange={(e) => onChange({ ...safeValue, path: e.target.value })}
                placeholder="/webhook"
                className="h-11 font-mono"
                disabled={disabled}
              />
              <p className="text-xs text-muted-foreground">
                Use path parameters like /users/:id or /events/:type/:id
              </p>
            </div>

            {/* Full URL Preview */}
            <div className="p-3 bg-muted rounded-lg border">
              <Label className="text-xs text-muted-foreground">Webhook URL:</Label>
              <div className="flex items-center gap-2 mt-1">
                <code className="flex-1 text-sm font-mono break-all">
                  {fullWebhookUrl}
                </code>
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  onClick={copyUrl}
                  className="h-8 w-8 p-0 flex-shrink-0"
                  disabled={disabled}
                >
                  {copied ? (
                    <Check className="w-4 h-4 text-green-500" />
                  ) : (
                    <Copy className="w-4 h-4" />
                  )}
                </Button>
              </div>
            </div>

            {/* Request Options */}
            <div className="space-y-3 pt-2">
              <Label className="text-sm font-semibold">Request Options</Label>

              <div className="flex items-center justify-between py-2 px-3 bg-muted/50 rounded-lg">
                <div>
                  <Label htmlFor="acceptFileUploads" className="text-sm font-medium cursor-pointer">
                    Accept File Uploads
                  </Label>
                  <p className="text-xs text-muted-foreground">
                    Enable multipart/form-data file uploads
                  </p>
                </div>
                <Switch
                  id="acceptFileUploads"
                  checked={acceptFileUploads}
                  onCheckedChange={(checked) => onChange({ ...safeValue, acceptFileUploads: checked })}
                  disabled={disabled}
                />
              </div>

              <div className="flex items-center justify-between py-2 px-3 bg-muted/50 rounded-lg">
                <div>
                  <Label htmlFor="rawBody" className="text-sm font-medium cursor-pointer">
                    Raw Body
                  </Label>
                  <p className="text-xs text-muted-foreground">
                    Don't parse body, receive as raw buffer
                  </p>
                </div>
                <Switch
                  id="rawBody"
                  checked={rawBody}
                  onCheckedChange={(checked) => onChange({ ...safeValue, rawBody: checked })}
                  disabled={disabled}
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="maxBodySize" className="text-xs">Max Body Size</Label>
                <Input
                  id="maxBodySize"
                  value={maxBodySize}
                  onChange={(e) => onChange({ ...safeValue, maxBodySize: e.target.value })}
                  placeholder="1mb"
                  className="h-9"
                  disabled={disabled}
                />
                <p className="text-xs text-muted-foreground">
                  Examples: 100kb, 1mb, 10mb
                </p>
              </div>
            </div>
          </AccordionContent>
        </AccordionItem>

        {/* Authentication Section */}
        <AccordionItem value="auth" className="border rounded-lg px-4">
          <AccordionTrigger className="hover:no-underline">
            <div className="flex items-center gap-2">
              <Shield className="w-4 h-4 text-green-500" />
              <span className="font-semibold">Authentication</span>
              {authType !== 'none' && (
                <span className="text-xs bg-green-100 dark:bg-green-900 text-green-700 dark:text-green-300 px-2 py-0.5 rounded ml-2">
                  {authType}
                </span>
              )}
            </div>
          </AccordionTrigger>
          <AccordionContent className="space-y-4 pt-4">
            {/* Auth Type */}
            <div className="space-y-2">
              <Label htmlFor="authType" className="text-sm font-semibold">
                Authentication Type
              </Label>
              <Select
                value={authType}
                onValueChange={(val) => onChange({ ...safeValue, authType: val as WebhookAuthType, authConfig: {} })}
                disabled={disabled}
              >
                <SelectTrigger className="h-11" id="authType">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="none">
                    <div>
                      <div className="font-medium">None</div>
                      <div className="text-xs text-muted-foreground">No authentication required</div>
                    </div>
                  </SelectItem>
                  <SelectItem value="basic">
                    <div>
                      <div className="font-medium">Basic Auth</div>
                      <div className="text-xs text-muted-foreground">Username and password</div>
                    </div>
                  </SelectItem>
                  <SelectItem value="header">
                    <div>
                      <div className="font-medium">Header Auth</div>
                      <div className="text-xs text-muted-foreground">Custom header validation</div>
                    </div>
                  </SelectItem>
                  <SelectItem value="jwt">
                    <div>
                      <div className="font-medium">JWT Auth</div>
                      <div className="text-xs text-muted-foreground">JSON Web Token validation</div>
                    </div>
                  </SelectItem>
                  <SelectItem value="apiKey">
                    <div>
                      <div className="font-medium">API Key</div>
                      <div className="text-xs text-muted-foreground">API key in header or query</div>
                    </div>
                  </SelectItem>
                </SelectContent>
              </Select>
            </div>

            {/* Basic Auth Config */}
            {authType === 'basic' && (
              <div className="grid grid-cols-2 gap-3">
                <div className="space-y-2">
                  <Label htmlFor="authUsername" className="text-xs">Username</Label>
                  <Input
                    id="authUsername"
                    value={authConfig.username || ''}
                    onChange={(e) => updateAuthConfig('username', e.target.value)}
                    placeholder="Username"
                    className="h-9"
                    disabled={disabled}
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="authPassword" className="text-xs">Password</Label>
                  <div className="relative">
                    <Input
                      id="authPassword"
                      type={showPassword ? 'text' : 'password'}
                      value={authConfig.password || ''}
                      onChange={(e) => updateAuthConfig('password', e.target.value)}
                      placeholder="Password"
                      className="h-9 pr-9"
                      disabled={disabled}
                    />
                    <button
                      type="button"
                      onClick={() => setShowPassword(!showPassword)}
                      className="absolute right-2 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                      disabled={disabled}
                    >
                      {showPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                    </button>
                  </div>
                </div>
              </div>
            )}

            {/* Header Auth Config */}
            {authType === 'header' && (
              <div className="grid grid-cols-2 gap-3">
                <div className="space-y-2">
                  <Label htmlFor="headerName" className="text-xs">Header Name</Label>
                  <Input
                    id="headerName"
                    value={authConfig.headerName || ''}
                    onChange={(e) => updateAuthConfig('headerName', e.target.value)}
                    placeholder="X-Auth-Token"
                    className="h-9"
                    disabled={disabled}
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="headerValue" className="text-xs">Expected Value</Label>
                  <div className="relative">
                    <Input
                      id="headerValue"
                      type={showPassword ? 'text' : 'password'}
                      value={authConfig.headerValue || ''}
                      onChange={(e) => updateAuthConfig('headerValue', e.target.value)}
                      placeholder="Secret token"
                      className="h-9 pr-9"
                      disabled={disabled}
                    />
                    <button
                      type="button"
                      onClick={() => setShowPassword(!showPassword)}
                      className="absolute right-2 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                      disabled={disabled}
                    >
                      {showPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                    </button>
                  </div>
                </div>
              </div>
            )}

            {/* JWT Auth Config */}
            {authType === 'jwt' && (
              <div className="space-y-3">
                <div className="space-y-2">
                  <Label htmlFor="jwtSecret" className="text-xs">JWT Secret / Public Key</Label>
                  <Textarea
                    id="jwtSecret"
                    value={authConfig.jwtSecret || ''}
                    onChange={(e) => updateAuthConfig('jwtSecret', e.target.value)}
                    placeholder="Enter secret key or PEM public key"
                    rows={3}
                    className="font-mono text-xs"
                    disabled={disabled}
                  />
                </div>
                <div className="space-y-2">
                  <Label htmlFor="jwtAlgorithm" className="text-xs">Algorithm</Label>
                  <Select
                    value={authConfig.jwtAlgorithm || 'HS256'}
                    onValueChange={(val) => updateAuthConfig('jwtAlgorithm', val)}
                    disabled={disabled}
                  >
                    <SelectTrigger className="h-9" id="jwtAlgorithm">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="HS256">HS256 (HMAC SHA-256)</SelectItem>
                      <SelectItem value="HS384">HS384 (HMAC SHA-384)</SelectItem>
                      <SelectItem value="HS512">HS512 (HMAC SHA-512)</SelectItem>
                      <SelectItem value="RS256">RS256 (RSA SHA-256)</SelectItem>
                      <SelectItem value="RS384">RS384 (RSA SHA-384)</SelectItem>
                      <SelectItem value="RS512">RS512 (RSA SHA-512)</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </div>
            )}

            {/* API Key Auth Config */}
            {authType === 'apiKey' && (
              <div className="space-y-3">
                <div className="grid grid-cols-2 gap-3">
                  <div className="space-y-2">
                    <Label htmlFor="apiKeyName" className="text-xs">Key Name</Label>
                    <Input
                      id="apiKeyName"
                      value={authConfig.apiKeyName || ''}
                      onChange={(e) => updateAuthConfig('apiKeyName', e.target.value)}
                      placeholder="X-API-Key"
                      className="h-9"
                      disabled={disabled}
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="apiKeyLocation" className="text-xs">Location</Label>
                    <Select
                      value={authConfig.apiKeyLocation || 'header'}
                      onValueChange={(val) => updateAuthConfig('apiKeyLocation', val)}
                      disabled={disabled}
                    >
                      <SelectTrigger className="h-9" id="apiKeyLocation">
                        <SelectValue />
                      </SelectTrigger>
                      <SelectContent>
                        <SelectItem value="header">Header</SelectItem>
                        <SelectItem value="query">Query Parameter</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                </div>
                <div className="space-y-2">
                  <Label htmlFor="apiKeyValue" className="text-xs">Expected Value</Label>
                  <div className="relative">
                    <Input
                      id="apiKeyValue"
                      type={showPassword ? 'text' : 'password'}
                      value={authConfig.apiKeyValue || ''}
                      onChange={(e) => updateAuthConfig('apiKeyValue', e.target.value)}
                      placeholder="Your API key"
                      className="h-9 pr-9 font-mono"
                      disabled={disabled}
                    />
                    <button
                      type="button"
                      onClick={() => setShowPassword(!showPassword)}
                      className="absolute right-2 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground"
                      disabled={disabled}
                    >
                      {showPassword ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                    </button>
                  </div>
                </div>
              </div>
            )}

            {authType === 'none' && (
              <div className="p-3 bg-amber-50 dark:bg-amber-950/20 border border-amber-200 dark:border-amber-900 rounded-lg">
                <p className="text-xs text-amber-700 dark:text-amber-300">
                  <strong>Warning:</strong> Without authentication, anyone with the webhook URL can send requests to this endpoint.
                </p>
              </div>
            )}
          </AccordionContent>
        </AccordionItem>

        {/* Response Configuration Section */}
        <AccordionItem value="response" className="border rounded-lg px-4">
          <AccordionTrigger className="hover:no-underline">
            <div className="flex items-center gap-2">
              <Send className="w-4 h-4 text-purple-500" />
              <span className="font-semibold">Response Configuration</span>
            </div>
          </AccordionTrigger>
          <AccordionContent className="space-y-4 pt-4">
            {/* Response Mode */}
            <div className="space-y-2">
              <Label htmlFor="responseMode" className="text-sm font-semibold">
                Response Mode
              </Label>
              <Select
                value={responseMode}
                onValueChange={(val) => onChange({ ...safeValue, responseMode: val as ResponseMode })}
                disabled={disabled}
              >
                <SelectTrigger className="h-11" id="responseMode">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="onReceived">
                    <div>
                      <div className="font-medium">On Received</div>
                      <div className="text-xs text-muted-foreground">Respond immediately when request is received</div>
                    </div>
                  </SelectItem>
                  <SelectItem value="lastNode">
                    <div>
                      <div className="font-medium">Last Node</div>
                      <div className="text-xs text-muted-foreground">Respond with output from last node</div>
                    </div>
                  </SelectItem>
                  <SelectItem value="responseNode">
                    <div>
                      <div className="font-medium">Response Node</div>
                      <div className="text-xs text-muted-foreground">Use a separate HTTP Response node</div>
                    </div>
                  </SelectItem>
                </SelectContent>
              </Select>
            </div>

            {/* Response Content Type */}
            <div className="space-y-2">
              <Label htmlFor="responseContentType" className="text-xs">Content-Type</Label>
              <Select
                value={responseContentType}
                onValueChange={(val) => onChange({ ...safeValue, responseContentType: val as ResponseContentType })}
                disabled={disabled}
              >
                <SelectTrigger className="h-9" id="responseContentType">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="application/json">application/json</SelectItem>
                  <SelectItem value="text/plain">text/plain</SelectItem>
                  <SelectItem value="text/html">text/html</SelectItem>
                  <SelectItem value="application/xml">application/xml</SelectItem>
                  <SelectItem value="auto">Auto-detect</SelectItem>
                </SelectContent>
              </Select>
            </div>

            {/* Response Status Code */}
            <div className="space-y-2">
              <Label htmlFor="responseStatusCode" className="text-xs">Status Code</Label>
              <Input
                id="responseStatusCode"
                type="number"
                value={responseStatusCode}
                onChange={(e) => onChange({ ...safeValue, responseStatusCode: Number(e.target.value) })}
                min={100}
                max={599}
                className="h-9"
                disabled={disabled}
              />
            </div>

            {/* No Response Body */}
            <div className="flex items-center justify-between py-2 px-3 bg-muted/50 rounded-lg">
              <div>
                <Label htmlFor="noResponseBody" className="text-sm font-medium cursor-pointer">
                  No Response Body
                </Label>
                <p className="text-xs text-muted-foreground">
                  Send response without body content
                </p>
              </div>
              <Switch
                id="noResponseBody"
                checked={noResponseBody}
                onCheckedChange={(checked) => onChange({ ...safeValue, noResponseBody: checked })}
                disabled={disabled}
              />
            </div>

            {/* Response Headers */}
            <div className="space-y-2">
              <div className="flex items-center justify-between">
                <Label className="text-xs">Response Headers</Label>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={addResponseHeader}
                  className="h-7 text-xs"
                  disabled={disabled}
                >
                  <Plus className="w-3 h-3 mr-1" />
                  Add Header
                </Button>
              </div>
              {responseHeaders.length > 0 && (
                <div className="space-y-2">
                  {responseHeaders.map((header) => (
                    <div key={header.id} className="flex items-center gap-2">
                      <Input
                        value={header.key}
                        onChange={(e) => updateResponseHeader(header.id, 'key', e.target.value)}
                        placeholder="Header name"
                        className="h-8 text-xs"
                        disabled={disabled}
                      />
                      <Input
                        value={header.value}
                        onChange={(e) => updateResponseHeader(header.id, 'value', e.target.value)}
                        placeholder="Header value"
                        className="h-8 text-xs"
                        disabled={disabled}
                      />
                      <Button
                        type="button"
                        variant="ghost"
                        size="sm"
                        onClick={() => removeResponseHeader(header.id)}
                        className="h-8 w-8 p-0 flex-shrink-0"
                        disabled={disabled}
                      >
                        <Trash2 className="w-3 h-3" />
                      </Button>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </AccordionContent>
        </AccordionItem>

        {/* Security & CORS Section */}
        <AccordionItem value="security" className="border rounded-lg px-4">
          <AccordionTrigger className="hover:no-underline">
            <div className="flex items-center gap-2">
              <Settings2 className="w-4 h-4 text-orange-500" />
              <span className="font-semibold">Security & CORS</span>
            </div>
          </AccordionTrigger>
          <AccordionContent className="space-y-4 pt-4">
            {/* IP Whitelist */}
            <div className="space-y-2">
              <Label htmlFor="ipWhitelist" className="text-xs">IP Whitelist</Label>
              <Input
                id="ipWhitelist"
                value={ipWhitelist}
                onChange={(e) => onChange({ ...safeValue, ipWhitelist: e.target.value })}
                placeholder="192.168.1.1, 10.0.0.0/8"
                className="h-9 font-mono"
                disabled={disabled}
              />
              <p className="text-xs text-muted-foreground">
                Comma-separated list of allowed IPs or CIDR ranges. Leave empty to allow all.
              </p>
            </div>

            {/* Rate Limiting */}
            <div className="space-y-3">
              <div className="flex items-center justify-between py-2 px-3 bg-muted/50 rounded-lg">
                <div>
                  <Label htmlFor="rateLimitEnabled" className="text-sm font-medium cursor-pointer">
                    Rate Limiting
                  </Label>
                  <p className="text-xs text-muted-foreground">
                    Limit requests per time window
                  </p>
                </div>
                <Switch
                  id="rateLimitEnabled"
                  checked={rateLimitEnabled}
                  onCheckedChange={(checked) => onChange({ ...safeValue, rateLimitEnabled: checked })}
                  disabled={disabled}
                />
              </div>

              {rateLimitEnabled && (
                <div className="grid grid-cols-2 gap-3">
                  <div className="space-y-2">
                    <Label htmlFor="rateLimitRequests" className="text-xs">Max Requests</Label>
                    <Input
                      id="rateLimitRequests"
                      type="number"
                      value={rateLimitRequests}
                      onChange={(e) => onChange({ ...safeValue, rateLimitRequests: Number(e.target.value) })}
                      min={1}
                      className="h-9"
                      disabled={disabled}
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="rateLimitWindow" className="text-xs">Window (seconds)</Label>
                    <Input
                      id="rateLimitWindow"
                      type="number"
                      value={rateLimitWindow}
                      onChange={(e) => onChange({ ...safeValue, rateLimitWindow: Number(e.target.value) })}
                      min={1}
                      className="h-9"
                      disabled={disabled}
                    />
                  </div>
                </div>
              )}
            </div>

            {/* CORS Configuration */}
            <div className="space-y-3">
              <div className="flex items-center justify-between py-2 px-3 bg-muted/50 rounded-lg">
                <div>
                  <Label htmlFor="corsEnabled" className="text-sm font-medium cursor-pointer">
                    Enable CORS
                  </Label>
                  <p className="text-xs text-muted-foreground">
                    Allow cross-origin requests
                  </p>
                </div>
                <Switch
                  id="corsEnabled"
                  checked={corsEnabled}
                  onCheckedChange={(checked) => onChange({ ...safeValue, corsEnabled: checked })}
                  disabled={disabled}
                />
              </div>

              {corsEnabled && (
                <div className="space-y-3">
                  <div className="space-y-2">
                    <Label htmlFor="corsOrigin" className="text-xs">Allowed Origins</Label>
                    <Input
                      id="corsOrigin"
                      value={corsOrigin}
                      onChange={(e) => onChange({ ...safeValue, corsOrigin: e.target.value })}
                      placeholder="* or https://example.com"
                      className="h-9 font-mono"
                      disabled={disabled}
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="corsMethods" className="text-xs">Allowed Methods</Label>
                    <Input
                      id="corsMethods"
                      value={corsMethods}
                      onChange={(e) => onChange({ ...safeValue, corsMethods: e.target.value })}
                      placeholder="GET,POST,PUT,DELETE,OPTIONS"
                      className="h-9 font-mono"
                      disabled={disabled}
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="corsHeaders" className="text-xs">Allowed Headers</Label>
                    <Input
                      id="corsHeaders"
                      value={corsHeaders}
                      onChange={(e) => onChange({ ...safeValue, corsHeaders: e.target.value })}
                      placeholder="Content-Type,Authorization"
                      className="h-9 font-mono"
                      disabled={disabled}
                    />
                  </div>
                  <div className="flex items-center justify-between py-2 px-3 bg-muted/50 rounded-lg">
                    <div>
                      <Label htmlFor="corsCredentials" className="text-sm font-medium cursor-pointer">
                        Allow Credentials
                      </Label>
                      <p className="text-xs text-muted-foreground">
                        Include cookies in CORS requests
                      </p>
                    </div>
                    <Switch
                      id="corsCredentials"
                      checked={corsCredentials}
                      onCheckedChange={(checked) => onChange({ ...safeValue, corsCredentials: checked })}
                      disabled={disabled}
                    />
                  </div>
                </div>
              )}
            </div>

            {/* Help Section */}
            <div className="p-4 bg-blue-50 dark:bg-blue-950/20 border border-blue-200 dark:border-blue-900 rounded-lg">
              <p className="text-sm font-semibold text-blue-900 dark:text-blue-100 mb-2">
                Webhook Security Tips
              </p>
              <ul className="list-disc list-inside text-xs text-blue-700 dark:text-blue-300 space-y-1">
                <li>Always use authentication in production</li>
                <li>Use HTTPS for encrypted communication</li>
                <li>Implement rate limiting to prevent abuse</li>
                <li>Whitelist IPs when possible</li>
                <li>Validate incoming data before processing</li>
              </ul>
            </div>
          </AccordionContent>
        </AccordionItem>
      </Accordion>
    </div>
  )
}
