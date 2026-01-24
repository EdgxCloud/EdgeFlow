/**
 * Code Editor Component
 *
 * Monaco-based code editor with Python and JavaScript support
 */

import { useState, useEffect } from 'react'
import Editor from '@monaco-editor/react'
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Maximize2, Minimize2, Code2, AlertCircle } from 'lucide-react'
import { cn } from '@/lib/utils'

type Language = 'javascript' | 'python' | 'json' | 'yaml' | 'sql'

interface CodeEditorProps {
  value: string
  onChange: (value: string) => void
  label?: string
  language?: Language
  height?: number
  readOnly?: boolean
  showLanguageSelector?: boolean
  showLineNumbers?: boolean
  disabled?: boolean
}

const LANGUAGE_TEMPLATES: Record<Language, string> = {
  javascript: `// JavaScript code
// Available variables: msg, node, context, flow, global, env

return msg;`,
  python: `# Python code
# Available variables: msg, node, context

return msg`,
  json: '{}',
  yaml: '',
  sql: 'SELECT * FROM table;',
}

const LANGUAGE_SNIPPETS: Record<Language, Array<{ label: string; code: string }>> = {
  javascript: [
    {
      label: 'Basic Function',
      code: `// Process incoming message
const value = msg.payload;

// Modify payload
msg.payload = {
  value: value,
  timestamp: Date.now()
};

return msg;`,
    },
    {
      label: 'Multiple Outputs',
      code: `// Send to different outputs based on condition
if (msg.payload > 100) {
  return [msg, null]; // First output
} else {
  return [null, msg]; // Second output
}`,
    },
    {
      label: 'Context Storage',
      code: `// Store value in context
const count = context.get('counter') || 0;
context.set('counter', count + 1);

msg.payload = count + 1;
return msg;`,
    },
    {
      label: 'Async Function',
      code: `// Async function example
async function processData() {
  const response = await fetch('https://api.example.com/data');
  const data = await response.json();
  return data;
}

// Call async function
processData().then(data => {
  msg.payload = data;
  node.send(msg);
});

// Return null to prevent immediate send
return null;`,
    },
  ],
  python: [
    {
      label: 'Basic Function',
      code: `# Process incoming message
value = msg['payload']

# Modify payload
msg['payload'] = {
    'value': value,
    'timestamp': time.time()
}

return msg`,
    },
    {
      label: 'Multiple Outputs',
      code: `# Send to different outputs based on condition
if msg['payload'] > 100:
    return [msg, None]  # First output
else:
    return [None, msg]  # Second output`,
    },
    {
      label: 'Context Storage',
      code: `# Store value in context
count = context.get('counter', 0)
context.set('counter', count + 1)

msg['payload'] = count + 1
return msg`,
    },
    {
      label: 'Exception Handling',
      code: `# Error handling
try:
    value = int(msg['payload'])
    msg['payload'] = value * 2
    return msg
except ValueError as e:
    node.error(f"Invalid value: {e}")
    return None`,
    },
  ],
  json: [],
  yaml: [],
  sql: [],
}

export function CodeEditor({
  value,
  onChange,
  label,
  language = 'javascript',
  height = 300,
  readOnly = false,
  showLanguageSelector = true,
  showLineNumbers = true,
  disabled = false,
}: CodeEditorProps) {
  const [internalValue, setInternalValue] = useState(value)
  const [currentLanguage, setCurrentLanguage] = useState<Language>(language)
  const [isExpanded, setIsExpanded] = useState(false)
  const [selectedSnippet, setSelectedSnippet] = useState<string>('')

  useEffect(() => {
    setInternalValue(value)
  }, [value])

  useEffect(() => {
    setCurrentLanguage(language)
  }, [language])

  const handleEditorChange = (newValue: string | undefined) => {
    const val = newValue || ''
    setInternalValue(val)
    onChange(val)
  }

  const handleLanguageChange = (newLang: Language) => {
    setCurrentLanguage(newLang)
    // If code is empty or matches template, set new template
    if (!internalValue || internalValue === LANGUAGE_TEMPLATES[currentLanguage]) {
      const template = LANGUAGE_TEMPLATES[newLang]
      setInternalValue(template)
      onChange(template)
    }
  }

  const handleSnippetInsert = (snippetCode: string) => {
    setInternalValue(snippetCode)
    onChange(snippetCode)
    setSelectedSnippet('')
  }

  const editorHeight = isExpanded ? 600 : height

  const snippets = LANGUAGE_SNIPPETS[currentLanguage] || []

  return (
    <div className="space-y-2">
      {label && (
        <div className="flex items-center justify-between">
          <Label className="text-sm font-semibold">{label}</Label>
          <div className="flex items-center gap-2">
            {showLanguageSelector && (
              <Select value={currentLanguage} onValueChange={handleLanguageChange}>
                <SelectTrigger className="w-32 h-8 text-xs">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="javascript">JavaScript</SelectItem>
                  <SelectItem value="python">Python</SelectItem>
                  <SelectItem value="json">JSON</SelectItem>
                  <SelectItem value="yaml">YAML</SelectItem>
                  <SelectItem value="sql">SQL</SelectItem>
                </SelectContent>
              </Select>
            )}

            {snippets.length > 0 && (
              <Select value={selectedSnippet} onValueChange={handleSnippetInsert}>
                <SelectTrigger className="w-40 h-8 text-xs">
                  <Code2 className="w-3 h-3 mr-1" />
                  <SelectValue placeholder="Insert snippet" />
                </SelectTrigger>
                <SelectContent>
                  {snippets.map((snippet, index) => (
                    <SelectItem key={index} value={snippet.code}>
                      {snippet.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            )}

            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={() => setIsExpanded(!isExpanded)}
              className="h-8 px-2"
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
          disabled && 'opacity-50 pointer-events-none'
        )}
      >
        <Editor
          height={editorHeight}
          language={currentLanguage}
          value={internalValue}
          onChange={handleEditorChange}
          options={{
            readOnly: readOnly || disabled,
            minimap: { enabled: editorHeight > 400 },
            scrollBeyondLastLine: false,
            fontSize: 13,
            lineNumbers: showLineNumbers ? 'on' : 'off',
            folding: true,
            automaticLayout: true,
            formatOnPaste: true,
            formatOnType: true,
            tabSize: currentLanguage === 'python' ? 4 : 2,
            insertSpaces: true,
            wordWrap: 'on',
            suggestOnTriggerCharacters: true,
            quickSuggestions: true,
            parameterHints: { enabled: true },
          }}
          theme="vs-dark"
        />
      </div>

      {/* Help Text */}
      <div className="p-3 bg-muted/50 rounded-lg border border-border">
        <div className="flex items-start gap-2">
          <AlertCircle className="w-4 h-4 text-muted-foreground mt-0.5 flex-shrink-0" />
          <div className="text-xs text-muted-foreground">
            {currentLanguage === 'javascript' && (
              <>
                <p className="font-semibold mb-1">Available JavaScript variables:</p>
                <ul className="list-disc list-inside space-y-0.5">
                  <li>
                    <code className="bg-background px-1 rounded">msg</code> - Input message
                    object
                  </li>
                  <li>
                    <code className="bg-background px-1 rounded">node</code> - Node instance
                    (send, warn, error methods)
                  </li>
                  <li>
                    <code className="bg-background px-1 rounded">context</code> - Flow context
                    (get, set methods)
                  </li>
                  <li>
                    <code className="bg-background px-1 rounded">flow</code> - Flow context
                    (alias)
                  </li>
                  <li>
                    <code className="bg-background px-1 rounded">global</code> - Global context
                  </li>
                  <li>
                    <code className="bg-background px-1 rounded">env</code> - Environment
                    variables
                  </li>
                </ul>
              </>
            )}
            {currentLanguage === 'python' && (
              <>
                <p className="font-semibold mb-1">Available Python variables:</p>
                <ul className="list-disc list-inside space-y-0.5">
                  <li>
                    <code className="bg-background px-1 rounded">msg</code> - Input message
                    dictionary
                  </li>
                  <li>
                    <code className="bg-background px-1 rounded">node</code> - Node instance
                    (send, warn, error methods)
                  </li>
                  <li>
                    <code className="bg-background px-1 rounded">context</code> - Flow context
                    (get, set methods)
                  </li>
                </ul>
                <p className="mt-2">
                  <strong>Note:</strong> Python execution requires the Python node to be properly
                  configured with a Python interpreter.
                </p>
              </>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}
