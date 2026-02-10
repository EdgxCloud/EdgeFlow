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
  const wsRef = useRef<WebSocket | null>(null)
  const scrollRef = useRef<HTMLDivElement>(null)
  const inputRef = useRef<HTMLInputElement>(null)

  // Auto-scroll to bottom
  useEffect(() => {
    if (scrollRef.current) {
      scrollRef.current.scrollTop = scrollRef.current.scrollHeight
    }
  }, [lines])

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
          addLine('output', msg.output.replace(/\n$/, ''))
        } else if (msg.type === 'error' && msg.output) {
          addLine('error', msg.output.replace(/\n$/, ''))
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

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    const cmd = input.trim()
    if (!cmd) return

    // Add to history
    setHistory(prev => [...prev.filter(h => h !== cmd), cmd])
    setHistoryIndex(-1)

    // Show command in terminal
    addLine('input', `$ ${cmd}`)
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
      addLine('system', '  Any other command runs on the device via /bin/bash')
      return
    }

    // Send to backend
    if (wsRef.current?.readyState !== WebSocket.OPEN) {
      addLine('error', 'Not connected. Reconnecting...')
      connectWS()
      return
    }

    setIsRunning(true)
    wsRef.current.send(JSON.stringify({ type: 'command', command: cmd }))
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'ArrowUp') {
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
      <div ref={scrollRef} className="flex-1 overflow-auto p-3 space-y-0.5">
        {lines.map((line) => (
          <div key={line.id} className={getLineColor(line.type)} style={{ whiteSpace: 'pre-wrap', wordBreak: 'break-all' }}>
            {line.text}
          </div>
        ))}
      </div>

      <form onSubmit={handleSubmit} className="flex items-center px-3 py-2 border-t border-gray-700/50">
        <span className="text-blue-400 mr-1 shrink-0">{cwd}</span>
        <span className="text-gray-400 mr-2 shrink-0">$</span>
        <input
          ref={inputRef}
          type="text"
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={handleKeyDown}
          disabled={isRunning}
          className="flex-1 bg-transparent outline-none text-green-400 caret-green-400 placeholder-gray-600"
          placeholder={isRunning ? 'Running...' : 'Type a command...'}
          autoFocus
          spellCheck={false}
          autoComplete="off"
        />
      </form>
    </div>
  )
}
