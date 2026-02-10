import { useCallback, useRef } from 'react'
import { Node, Edge, XYPosition } from '@xyflow/react'

interface CopyPasteHook {
  copy: (selectedNodes: Node[], selectedEdges: Edge[]) => void
  paste: (position?: XYPosition) => { nodes: Node[], edges: Edge[] } | null
  cut: (selectedNodes: Node[], selectedEdges: Edge[]) => void
  canPaste: boolean
}

const PASTE_OFFSET = 20

/**
 * Custom hook for copy/paste functionality
 * Stores copied nodes and edges in a ref
 */
export function useCopyPaste(): CopyPasteHook {
  const clipboard = useRef<{ nodes: Node[], edges: Edge[] } | null>(null)

  const copy = useCallback((selectedNodes: Node[], selectedEdges: Edge[]) => {
    if (selectedNodes.length === 0) {
      return
    }

    // Deep clone the selected nodes and edges
    clipboard.current = {
      nodes: JSON.parse(JSON.stringify(selectedNodes)),
      edges: JSON.parse(JSON.stringify(selectedEdges)),
    }

    console.log('📋 Copied', selectedNodes.length, 'nodes and', selectedEdges.length, 'edges')
  }, [])

  const paste = useCallback((position?: XYPosition) => {
    if (!clipboard.current) {
      return null
    }

    const { nodes: copiedNodes, edges: copiedEdges } = clipboard.current

    // Create ID mapping for pasted nodes
    const idMap = new Map<string, string>()
    const now = Date.now()

    // Generate new IDs and adjust positions
    const pastedNodes = copiedNodes.map((node, index) => {
      const newId = `${node.data?.nodeType || 'node'}-${now}-${index}`
      idMap.set(node.id, newId)

      return {
        ...node,
        id: newId,
        position: position
          ? {
              x: position.x + (index * PASTE_OFFSET),
              y: position.y + (index * PASTE_OFFSET)
            }
          : {
              x: node.position.x + PASTE_OFFSET,
              y: node.position.y + PASTE_OFFSET,
            },
        selected: true, // Select pasted nodes
      }
    })

    // Update edge IDs to match new node IDs
    const pastedEdges = copiedEdges
      .filter(edge => {
        // Only include edges where both source and target nodes were copied
        return idMap.has(edge.source) && idMap.has(edge.target)
      })
      .map(edge => ({
        ...edge,
        id: `e${idMap.get(edge.source)}-${idMap.get(edge.target)}`,
        source: idMap.get(edge.source)!,
        target: idMap.get(edge.target)!,
        selected: true, // Select pasted edges
      }))

    console.log('📌 Pasted', pastedNodes.length, 'nodes and', pastedEdges.length, 'edges')

    return { nodes: pastedNodes, edges: pastedEdges }
  }, [])

  const cut = useCallback((selectedNodes: Node[], selectedEdges: Edge[]) => {
    copy(selectedNodes, selectedEdges)
    // The actual deletion is handled by the component using this hook
    console.log('✂️ Cut', selectedNodes.length, 'nodes and', selectedEdges.length, 'edges')
  }, [copy])

  return {
    copy,
    paste,
    cut,
    canPaste: clipboard.current !== null,
  }
}

/**
 * Hook for keyboard shortcuts
 */
export function useKeyboardShortcuts(
  onCopy: () => void,
  onPaste: () => void,
  onCut: () => void,
  onDelete: () => void,
  onUndo?: () => void,
  onRedo?: () => void,
  onSelectAll?: () => void,
  onDuplicate?: () => void
) {
  const handleKeyDown = useCallback(
    (event: KeyboardEvent) => {
      // Check if we're in an input/textarea
      const target = event.target as HTMLElement
      if (
        target.tagName === 'INPUT' ||
        target.tagName === 'TEXTAREA' ||
        target.isContentEditable
      ) {
        return
      }

      const isMac = navigator.platform.toUpperCase().indexOf('MAC') >= 0
      const modifierKey = isMac ? event.metaKey : event.ctrlKey

      if (modifierKey) {
        switch (event.key.toLowerCase()) {
          case 'c':
            event.preventDefault()
            onCopy()
            break
          case 'v':
            event.preventDefault()
            onPaste()
            break
          case 'x':
            event.preventDefault()
            onCut()
            break
          case 'd':
            event.preventDefault()
            onDuplicate?.()
            break
          case 'z':
            event.preventDefault()
            if (event.shiftKey) {
              onRedo?.()
            } else {
              onUndo?.()
            }
            break
          case 'y':
            if (!isMac) {
              event.preventDefault()
              onRedo?.()
            }
            break
          case 'a':
            event.preventDefault()
            onSelectAll?.()
            break
        }
      } else if (event.key === 'Delete' || event.key === 'Backspace') {
        event.preventDefault()
        onDelete()
      }
    },
    [onCopy, onPaste, onCut, onDelete, onUndo, onRedo, onSelectAll, onDuplicate]
  )

  return handleKeyDown
}
