/**
 * Terminal Panel
 * Real shell terminal connected to backend via WebSocket
 */

import { useState, useRef, useEffect, useCallback } from 'react'
import { WS_URL } from '@/services/api'

interface TerminalLine {
  id: number
  type: 'input' | 'output' | 'error' | 'system'
  text: string
}

let lineCounter = 0

export function TerminalPanel() {
  const [lines, setLines] = useState<TerminalLine[]>([
    { id: lineCounter++, type: 'system', text: 'EdgeFlow Terminal v1.0.0' },
    { id: lineCounter++, type: 'system', text: 'Type commands to execute on the device. Type "help" for info.\n' },
  ])
  const [input, setInput] = useState('')
  const [cwd, setCwd] = useState('~')
  const [isRunning, setIsRunning] = useState(false)
  const [history, setHistory] = useState<string[]>([])
  const [historyIndex, setHistoryIndex] = useState(-1)
  const [tabSuggestions, setTabSuggestions] = useState<string[]>([])
  const [tabIndex, setTabIndex] = useState(-1)
  const wsRef = useRef<WebSocket | null>(null)
  const scrollRef = useRef<HTMLDivElement>(null)
  const inputRef = useRef<HTMLInputElement>(null)

  // Auto-scroll to bottom
  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight
    }
  }, [lines])

  // Re-focus input when command finishes running
  useEffect(() => {
    if (!isRunning) {
      // Small delay to ensure DOM has updated after disabled state changes
      setTimeout(() => inputRef.current?.focus(), 50)
    }
  }, [isRunning])

  // Connect WebSocket
  const connectWS = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) return

    const wsUrl = WS_URL.replace('/ws', '/ws/terminal')
    const ws = new WebSocket(wsUrl)

    ws.onopen = () => {
      addLine('system', 'Connected to device.')
    }

    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data)
        if (msg.type === 'output' && msg.output) {
          const outputLines = msg.output.replace(/\n$/, '').split('\n')
          outputLines.forEach((line: string) => addLine('output', line))
        } else if (msg.type === 'error' && msg.output) {
          addLine('error', msg.output.replace(/\n$/, ''))
        } else if (msg.type === 'system' && msg.output) {
          addLine('system', msg.output.replace(/\n$/, ''))
        } else if (msg.type === 'completion' && msg.suggestions) {
          handleTabSuggestions(msg.suggestions, msg.partial)
        } else if (msg.type === 'done') {
          if (msg.cwd) setCwd(msg.cwd)
          setIsRunning(false)
        }
      } catch {
        // ignore parse errors
      }
    }

    ws.onerror = () => {
      addLine('error', 'Connection error. Is the backend running?')
      setIsRunning(false)
    }

    ws.onclose = () => {
      setIsRunning(false)
    }

    wsRef.current = ws
  }, [])

  // Connect on mount
  useEffect(() => {
    connectWS()
    return () => {
      wsRef.current?.close()
    }
  }, [connectWS])

  const addLine = (type: TerminalLine['type'], text: string) => {
    setLines(prev => [...prev, { id: lineCounter++, type, text }])
  }

  // Handle tab completion suggestions from backend
  const handleTabSuggestions = (suggestions: string[], partial: string) => {
    if (!suggestions || suggestions.length === 0) return

    if (suggestions.length === 1) {
      // Single match - auto-complete
      applyCompletion(suggestions[0], partial)
    } else {
      // Multiple matches - show them and find common prefix
      setTabSuggestions(suggestions)
      setTabIndex(0)
      addLine('system', suggestions.join('  '))

      // Apply longest common prefix
      const commonPrefix = findCommonPrefix(suggestions)
      if (commonPrefix.length > partial.length) {
        applyCompletion(commonPrefix, partial)
      }
    }
  }

  const applyCompletion = (completion: string, partial: string) => {
    setInput(prev => {
      // Find the last word boundary to replace partial with completion
      const parts = prev.split(/\s+/)
      parts[parts.length - 1] = completion
      const result = parts.join(' ')
      // Add trailing space if it's a single completion (not a prefix)
      return result + (tabSuggestions.length <= 1 ? ' ' : '')
    })
    setTabSuggestions([])
    setTabIndex(-1)
  }

  const findCommonPrefix = (strs: string[]): string => {
    if (strs.length === 0) return ''
    let prefix = strs[0]
    for (let i = 1; i < strs.length; i++) {
      while (!strs[i].startsWith(prefix)) {
        prefix = prefix.slice(0, -1)
        if (prefix === '') return ''
      }
    }
    return prefix
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    const cmd = input.trim()
    if (!cmd || isRunning) return

    // Clear tab state
    setTabSuggestions([])
    setTabIndex(-1)

    // Add to history
    setHistory(prev => [...prev.filter(h => h !== cmd), cmd])
    setHistoryIndex(-1)

    // Show command in terminal
    addLine('input', `${cwd} $ ${cmd}`)
    setInput('')

    // Handle local commands
    if (cmd === 'clear') {
      setLines([])
      return
    }
    if (cmd === 'help') {
      addLine('system', 'EdgeFlow Terminal - Commands are executed on the device.')
      addLine('system', '  clear     - Clear terminal')
      addLine('system', '  help      - Show this help')
      addLine('system', '  Tab       - Auto-complete file/directory names')
      addLine('system', '  Any other command runs on the device via shell')
      return
    }

    // Send to backend
    const ws = wsRef.current
    if (!ws || ws.readyState !== WebSocket.OPEN) {
      addLine('error', 'Not connected. Reconnecting...')
      connectWS()
      return
    }

    setIsRunning(true)
    try {
      ws.send(JSON.stringify({ type: 'command', command: cmd }))
    } catch (err) {
      addLine('error', `Failed to send command: ${err}`)
      setIsRunning(false)
    }
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Tab') {
      e.preventDefault()
      if (isRunning) return

      // Send tab completion request to backend
      const ws = wsRef.current
      if (!ws || ws.readyState !== WebSocket.OPEN) return

      const currentInput = input
      // Extract the last word as the partial to complete
      const parts = currentInput.split(/\s+/)
      const partial = parts[parts.length - 1] || ''

      try {
        ws.send(JSON.stringify({
          type: 'complete',
          command: currentInput,
          partial: partial,
        }))
      } catch {
        // ignore
      }
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      if (history.length > 0) {
        const newIndex = historyIndex < history.length - 1 ? historyIndex + 1 : historyIndex
        setHistoryIndex(newIndex)
        setInput(history[history.length - 1 - newIndex] || '')
      }
    } else if (e.key === 'ArrowDown') {
      e.preventDefault()
      if (historyIndex > 0) {
        const newIndex = historyIndex - 1
        setHistoryIndex(newIndex)
        setInput(history[history.length - 1 - newIndex] || '')
      } else {
        setHistoryIndex(-1)
        setInput('')
      }
    } else {
      // Any other key clears tab suggestions
      if (tabSuggestions.length > 0) {
        setTabSuggestions([])
        setTabIndex(-1)
      }
    }
  }

  const getLineColor = (type: TerminalLine['type']) => {
    switch (type) {
      case 'input': return 'text-cyan-400'
      case 'error': return 'text-red-400'
      case 'system': return 'text-yellow-400'
      default: return 'text-green-400'
    }
  }

  return (
    <div
      className="h-full flex flex-col bg-[#1a1a2e] text-green-400 font-mono text-xs"
      onClick={() => inputRef.current?.focus()}
    >
      <div ref={scrollRef} className="flex-1 min-h-0 overflow-auto p-3 space-y-0.5">
        {lines.map((line) => (
          <div key={line.id} className={getLineColor(line.type)} style={{ whiteSpace: 'pre-wrap', wordBreak: 'break-all' }}>
            {line.text}
          </div>
        ))}
      </div>

      <form onSubmit={handleSubmit} className="shrink-0 flex items-center px-3 py-2 border-t border-gray-700/50 bg-[#1a1a2e]">
        <span className="text-blue-400 mr-1 shrink-0">{cwd}</span>
        <span className="text-gray-400 mr-2 shrink-0">$</span>
        <input
          ref={inputRef}
          type="text"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={handleKeyDown}
          readOnly={isRunning}
          className={`flex-1 bg-transparent outline-none text-green-400 caret-green-400 placeholder-gray-600 ${isRunning ? 'opacity-50' : ''}`}
          placeholder={isRunning ? 'Running...' : 'Type a command...'}
          autoFocus
          spellCheck={false}
          autoComplete="off"
        />
        {isRunning && (
          <span className="text-yellow-400 ml-2 animate-pulse">...</span>
        )}
      </form>
    </div>
  )
}
