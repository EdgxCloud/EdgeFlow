import { useRef, useEffect } from 'react'
import Editor, { OnMount, Monaco } from '@monaco-editor/react'
import { useTheme } from '@/components/theme/theme-provider'
import { cn } from '@/lib/utils'

interface MonacoEditorProps {
  value: string
  onChange?: (value: string | undefined) => void
  language?: 'javascript' | 'python' | 'json' | 'html' | 'css' | 'typescript'
  height?: string
  readOnly?: boolean
  minimap?: boolean
  lineNumbers?: boolean
  className?: string
  options?: any
}

export default function MonacoEditor({
  value,
  onChange,
  language = 'javascript',
  height = '400px',
  readOnly = false,
  minimap = true,
  lineNumbers = true,
  className,
  options = {},
}: MonacoEditorProps) {
  const editorRef = useRef<any>(null)
  const { theme } = useTheme()

  const handleEditorDidMount: OnMount = (editor, monaco) => {
    editorRef.current = editor

    // Configure template variable suggestions
    if (language === 'javascript' || language === 'python') {
      monaco.languages.registerCompletionItemProvider(language, {
        provideCompletionItems: (model, position) => {
          const suggestions = [
            {
              label: 'msg.payload',
              kind: monaco.languages.CompletionItemKind.Property,
              insertText: 'msg.payload',
              documentation: 'The message payload',
            },
            {
              label: 'msg.topic',
              kind: monaco.languages.CompletionItemKind.Property,
              insertText: 'msg.topic',
              documentation: 'The message topic',
            },
            {
              label: 'msg._msgid',
              kind: monaco.languages.CompletionItemKind.Property,
              insertText: 'msg._msgid',
              documentation: 'The message ID',
            },
            {
              label: 'context.get',
              kind: monaco.languages.CompletionItemKind.Function,
              insertText: 'context.get("${1:key}")',
              insertTextRules: monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
              documentation: 'Get value from context',
            },
            {
              label: 'context.set',
              kind: monaco.languages.CompletionItemKind.Function,
              insertText: 'context.set("${1:key}", ${2:value})',
              insertTextRules: monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
              documentation: 'Set value in context',
            },
            {
              label: 'flow.get',
              kind: monaco.languages.CompletionItemKind.Function,
              insertText: 'flow.get("${1:key}")',
              insertTextRules: monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
              documentation: 'Get value from flow context',
            },
            {
              label: 'global.get',
              kind: monaco.languages.CompletionItemKind.Function,
              insertText: 'global.get("${1:key}")',
              insertTextRules: monaco.languages.CompletionItemInsertTextRule.InsertAsSnippet,
              documentation: 'Get value from global context',
            },
          ]

          return {
            suggestions: suggestions.map((s) => ({
              ...s,
              range: {
                startLineNumber: position.lineNumber,
                startColumn: position.column,
                endLineNumber: position.lineNumber,
                endColumn: position.column,
              },
            })),
          }
        },
      })
    }
  }

  const editorOptions = {
    readOnly,
    minimap: { enabled: minimap },
    lineNumbers: lineNumbers ? 'on' : 'off',
    fontSize: 14,
    fontFamily: 'JetBrains Mono, Consolas, Monaco, monospace',
    scrollBeyondLastLine: false,
    automaticLayout: true,
    tabSize: 2,
    wordWrap: 'on',
    renderWhitespace: 'selection',
    bracketPairColorization: { enabled: true },
    suggest: {
      showKeywords: true,
      showSnippets: true,
    },
    quickSuggestions: {
      other: true,
      comments: false,
      strings: true,
    },
    ...options,
  }

  return (
    <div className={cn('border border-border rounded-md overflow-hidden', className)}>
      <Editor
        height={height}
        language={language}
        value={value}
        onChange={onChange}
        onMount={handleEditorDidMount}
        theme={theme === 'dark' ? 'vs-dark' : 'light'}
        options={editorOptions}
        loading={
          <div className="flex items-center justify-center h-full">
            <div className="text-sm text-muted-foreground">Loading editor...</div>
          </div>
        }
      />
    </div>
  )
}
