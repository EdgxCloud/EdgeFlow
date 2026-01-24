/**
 * JSON Editor Component
 *
 * Monaco-based JSON editor with validation
 */

import { useState, useEffect } from 'react'
import Editor from '@monaco-editor/react'
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'
import { AlertCircle, Check, Maximize2, Minimize2 } from 'lucide-react'
import { cn } from '@/lib/utils'

interface JSONEditorProps {
  value: string | object
  onChange: (value: any) => void
  label?: string
  height?: number
  readOnly?: boolean
  showValidation?: boolean
  disabled?: boolean
}

export function JSONEditor({
  value,
  onChange,
  label,
  height = 200,
  readOnly = false,
  showValidation = true,
  disabled = false,
}: JSONEditorProps) {
  const [internalValue, setInternalValue] = useState('')
  const [isValid, setIsValid] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [isExpanded, setIsExpanded] = useState(false)

  // Initialize internal value
  useEffect(() => {
    const stringValue = typeof value === 'string' ? value : JSON.stringify(value, null, 2)
    setInternalValue(stringValue)
    validateJSON(stringValue)
  }, [value])

  const validateJSON = (jsonString: string) => {
    if (!jsonString.trim()) {
      setIsValid(true)
      setError(null)
      return true
    }

    try {
      JSON.parse(jsonString)
      setIsValid(true)
      setError(null)
      return true
    } catch (err) {
      setIsValid(false)
      setError(err instanceof Error ? err.message : 'Invalid JSON')
      return false
    }
  }

  const handleEditorChange = (newValue: string | undefined) => {
    const value = newValue || ''
    setInternalValue(value)

    if (validateJSON(value)) {
      try {
        const parsed = JSON.parse(value)
        onChange(parsed)
      } catch {
        // If empty or whitespace, pass empty object
        if (!value.trim()) {
          onChange({})
        }
      }
    }
  }

  const handleFormat = () => {
    try {
      const parsed = JSON.parse(internalValue)
      const formatted = JSON.stringify(parsed, null, 2)
      setInternalValue(formatted)
      onChange(parsed)
      setIsValid(true)
      setError(null)
    } catch {
      // Already invalid, do nothing
    }
  }

  const handleMinify = () => {
    try {
      const parsed = JSON.parse(internalValue)
      const minified = JSON.stringify(parsed)
      setInternalValue(minified)
      onChange(parsed)
      setIsValid(true)
      setError(null)
    } catch {
      // Already invalid, do nothing
    }
  }

  const editorHeight = isExpanded ? 400 : height

  return (
    <div className="space-y-2">
      {label && (
        <div className="flex items-center justify-between">
          <Label className="text-sm font-semibold">{label}</Label>
          <div className="flex items-center gap-2">
            {showValidation && (
              <div
                className={cn(
                  'flex items-center gap-1 text-xs',
                  isValid ? 'text-green-600' : 'text-red-600'
                )}
              >
                {isValid ? (
                  <>
                    <Check className="w-3 h-3" />
                    Valid
                  </>
                ) : (
                  <>
                    <AlertCircle className="w-3 h-3" />
                    Invalid
                  </>
                )}
              </div>
            )}
            {!readOnly && (
              <>
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  onClick={handleFormat}
                  disabled={disabled || !isValid}
                  className="h-7 text-xs"
                >
                  Format
                </Button>
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  onClick={handleMinify}
                  disabled={disabled || !isValid}
                  className="h-7 text-xs"
                >
                  Minify
                </Button>
              </>
            )}
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={() => setIsExpanded(!isExpanded)}
              className="h-7 px-2"
            >
              {isExpanded ? (
                <Minimize2 className="w-4 h-4" />
              ) : (
                <Maximize2 className="w-4 h-4" />
              )}
            </Button>
          </div>
        </div>
      )}

      <div
        className={cn(
          'border rounded-lg overflow-hidden',
          !isValid && 'border-red-500',
          disabled && 'opacity-50 pointer-events-none'
        )}
      >
        <Editor
          height={editorHeight}
          defaultLanguage="json"
          value={internalValue}
          onChange={handleEditorChange}
          options={{
            readOnly: readOnly || disabled,
            minimap: { enabled: editorHeight > 300 },
            scrollBeyondLastLine: false,
            fontSize: 13,
            lineNumbers: 'on',
            folding: true,
            automaticLayout: true,
            formatOnPaste: true,
            formatOnType: true,
            tabSize: 2,
          }}
          theme="vs-dark"
        />
      </div>

      {error && showValidation && (
        <div className="flex items-start gap-2 p-3 bg-red-50 dark:bg-red-950/20 border border-red-200 dark:border-red-900 rounded-lg">
          <AlertCircle className="w-4 h-4 text-red-600 dark:text-red-400 mt-0.5 flex-shrink-0" />
          <div>
            <p className="text-sm font-medium text-red-900 dark:text-red-100">JSON Error</p>
            <p className="text-xs text-red-700 dark:text-red-300 mt-1">{error}</p>
          </div>
        </div>
      )}
    </div>
  )
}
