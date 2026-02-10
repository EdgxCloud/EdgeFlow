import { useEffect, useRef } from 'react'
import {
  Copy,
  Scissors,
  ClipboardPaste,
  CopyPlus,
  Trash2,
  MousePointerClick,
  Maximize,
  Settings,
} from 'lucide-react'
import { XYPosition } from '@xyflow/react'

export interface ContextMenuState {
  visible: boolean
  x: number
  y: number
  type: 'pane' | 'node' | 'selection'
  nodeId?: string
  flowPosition?: XYPosition
}

interface CanvasContextMenuProps {
  state: ContextMenuState
  canPaste: boolean
  onAction: (action: string) => void
  onClose: () => void
}

interface MenuItem {
  label: string
  action: string
  icon: React.ReactNode
  shortcut?: string
  disabled?: boolean
  separator?: false
}

interface SeparatorItem {
  separator: true
}

type MenuEntry = MenuItem | SeparatorItem

export default function CanvasContextMenu({ state, canPaste, onAction, onClose }: CanvasContextMenuProps) {
  const menuRef = useRef<HTMLDivElement>(null)

  // Close on Escape or scroll
  useEffect(() => {
    if (!state.visible) return

    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose()
    }
    const handleScroll = () => onClose()

    document.addEventListener('keydown', handleEscape)
    document.addEventListener('scroll', handleScroll, true)
    return () => {
      document.removeEventListener('keydown', handleEscape)
      document.removeEventListener('scroll', handleScroll, true)
    }
  }, [state.visible, onClose])

  // Adjust position to stay within viewport
  useEffect(() => {
    if (!state.visible || !menuRef.current) return

    const menu = menuRef.current
    const rect = menu.getBoundingClientRect()
    const viewportWidth = window.innerWidth
    const viewportHeight = window.innerHeight

    let x = state.x
    let y = state.y

    if (x + rect.width > viewportWidth) {
      x = viewportWidth - rect.width - 8
    }
    if (y + rect.height > viewportHeight) {
      y = viewportHeight - rect.height - 8
    }

    menu.style.left = `${x}px`
    menu.style.top = `${y}px`
  }, [state.visible, state.x, state.y])

  if (!state.visible) return null

  const nodeMenuItems: MenuEntry[] = [
    { label: 'Copy', action: 'copy', icon: <Copy className="w-4 h-4" />, shortcut: 'Ctrl+C' },
    { label: 'Cut', action: 'cut', icon: <Scissors className="w-4 h-4" />, shortcut: 'Ctrl+X' },
    { label: 'Duplicate', action: 'duplicate', icon: <CopyPlus className="w-4 h-4" />, shortcut: 'Ctrl+D' },
    { label: 'Delete', action: 'delete', icon: <Trash2 className="w-4 h-4" />, shortcut: 'Del' },
    { separator: true },
    { label: 'Configure...', action: 'configure', icon: <Settings className="w-4 h-4" /> },
  ]

  const paneMenuItems: MenuEntry[] = [
    { label: 'Paste', action: 'paste', icon: <ClipboardPaste className="w-4 h-4" />, shortcut: 'Ctrl+V', disabled: !canPaste },
    { label: 'Select All', action: 'selectAll', icon: <MousePointerClick className="w-4 h-4" />, shortcut: 'Ctrl+A' },
    { separator: true },
    { label: 'Fit View', action: 'fitView', icon: <Maximize className="w-4 h-4" /> },
  ]

  const selectionMenuItems: MenuEntry[] = [
    { label: 'Copy', action: 'copy', icon: <Copy className="w-4 h-4" />, shortcut: 'Ctrl+C' },
    { label: 'Cut', action: 'cut', icon: <Scissors className="w-4 h-4" />, shortcut: 'Ctrl+X' },
    { label: 'Duplicate', action: 'duplicate', icon: <CopyPlus className="w-4 h-4" />, shortcut: 'Ctrl+D' },
    { label: 'Delete', action: 'delete', icon: <Trash2 className="w-4 h-4" />, shortcut: 'Del' },
  ]

  const items = state.type === 'pane' ? paneMenuItems : state.type === 'selection' ? selectionMenuItems : nodeMenuItems

  return (
    <div
      ref={menuRef}
      className="fixed z-50 min-w-[180px] bg-white dark:bg-gray-900 border border-gray-200 dark:border-gray-700 rounded-lg shadow-xl py-1 animate-in fade-in zoom-in-95 duration-100"
      style={{ left: state.x, top: state.y }}
      onClick={(e) => e.stopPropagation()}
    >
      {items.map((item, index) => {
        if ('separator' in item && item.separator) {
          return <div key={index} className="my-1 border-t border-gray-200 dark:border-gray-700" />
        }

        const menuItem = item as MenuItem
        return (
          <button
            key={menuItem.action}
            className="w-full flex items-center gap-3 px-3 py-1.5 text-sm text-gray-700 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-gray-800 disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
            onClick={() => !menuItem.disabled && onAction(menuItem.action)}
            disabled={menuItem.disabled}
          >
            <span className="text-gray-500 dark:text-gray-400">{menuItem.icon}</span>
            <span className="flex-1 text-left">{menuItem.label}</span>
            {menuItem.shortcut && (
              <span className="text-xs text-gray-400 dark:text-gray-500">{menuItem.shortcut}</span>
            )}
          </button>
        )
      })}
    </div>
  )
}
