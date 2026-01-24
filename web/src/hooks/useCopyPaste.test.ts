import { describe, it, expect, beforeEach } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { useCopyPaste } from './useCopyPaste'
import { Node, Edge } from '@xyflow/react'

describe('useCopyPaste', () => {
  let hook: ReturnType<typeof renderHook<ReturnType<typeof useCopyPaste>, unknown>>

  beforeEach(() => {
    hook = renderHook(() => useCopyPaste())
  })

  it('initializes with empty clipboard', () => {
    expect(hook.result.current.canPaste).toBe(false)
  })

  it('copies nodes successfully', () => {
    const nodes: Node[] = [
      {
        id: 'node-1',
        type: 'custom',
        position: { x: 100, y: 100 },
        data: { label: 'Test Node', nodeType: 'inject' },
      },
    ]
    const edges: Edge[] = []

    act(() => {
      hook.result.current.copy(nodes, edges)
    })

    expect(hook.result.current.canPaste).toBe(true)
  })

  it('pastes nodes with new IDs', () => {
    const nodes: Node[] = [
      {
        id: 'node-1',
        type: 'custom',
        position: { x: 100, y: 100 },
        data: { label: 'Test Node', nodeType: 'inject' },
      },
    ]
    const edges: Edge[] = []

    act(() => {
      hook.result.current.copy(nodes, edges)
    })

    let pasted: { nodes: Node[]; edges: Edge[] } | null = null
    act(() => {
      pasted = hook.result.current.paste()
    })

    expect(pasted).not.toBeNull()
    expect(pasted!.nodes).toHaveLength(1)
    expect(pasted!.nodes[0].id).not.toBe('node-1')
    expect(pasted!.nodes[0].data.label).toBe('Test Node')
  })

  it('pastes nodes with offset position', () => {
    const nodes: Node[] = [
      {
        id: 'node-1',
        type: 'custom',
        position: { x: 100, y: 100 },
        data: { label: 'Test Node', nodeType: 'inject' },
      },
    ]
    const edges: Edge[] = []

    act(() => {
      hook.result.current.copy(nodes, edges)
    })

    let pasted: { nodes: Node[]; edges: Edge[] } | null = null
    act(() => {
      pasted = hook.result.current.paste()
    })

    expect(pasted!.nodes[0].position.x).toBe(120) // Original 100 + 20 offset
    expect(pasted!.nodes[0].position.y).toBe(120)
  })

  it('pastes multiple nodes', () => {
    const nodes: Node[] = [
      {
        id: 'node-1',
        type: 'custom',
        position: { x: 100, y: 100 },
        data: { label: 'Node 1', nodeType: 'inject' },
      },
      {
        id: 'node-2',
        type: 'custom',
        position: { x: 200, y: 200 },
        data: { label: 'Node 2', nodeType: 'debug' },
      },
    ]
    const edges: Edge[] = []

    act(() => {
      hook.result.current.copy(nodes, edges)
    })

    let pasted: { nodes: Node[]; edges: Edge[] } | null = null
    act(() => {
      pasted = hook.result.current.paste()
    })

    expect(pasted!.nodes).toHaveLength(2)
    expect(pasted!.nodes[0].id).not.toBe('node-1')
    expect(pasted!.nodes[1].id).not.toBe('node-2')
  })

  it('pastes edges with remapped IDs', () => {
    const nodes: Node[] = [
      {
        id: 'node-1',
        type: 'custom',
        position: { x: 100, y: 100 },
        data: { label: 'Node 1', nodeType: 'inject' },
      },
      {
        id: 'node-2',
        type: 'custom',
        position: { x: 200, y: 200 },
        data: { label: 'Node 2', nodeType: 'debug' },
      },
    ]
    const edges: Edge[] = [
      {
        id: 'edge-1',
        source: 'node-1',
        target: 'node-2',
      },
    ]

    act(() => {
      hook.result.current.copy(nodes, edges)
    })

    let pasted: { nodes: Node[]; edges: Edge[] } | null = null
    act(() => {
      pasted = hook.result.current.paste()
    })

    expect(pasted!.edges).toHaveLength(1)
    expect(pasted!.edges[0].source).not.toBe('node-1')
    expect(pasted!.edges[0].target).not.toBe('node-2')
    expect(pasted!.edges[0].source).toBe(pasted!.nodes[0].id)
    expect(pasted!.edges[0].target).toBe(pasted!.nodes[1].id)
  })

  it('filters out edges with missing nodes', () => {
    const nodes: Node[] = [
      {
        id: 'node-1',
        type: 'custom',
        position: { x: 100, y: 100 },
        data: { label: 'Node 1', nodeType: 'inject' },
      },
    ]
    const edges: Edge[] = [
      {
        id: 'edge-1',
        source: 'node-1',
        target: 'node-2', // node-2 is not in copied nodes
      },
    ]

    act(() => {
      hook.result.current.copy(nodes, edges)
    })

    let pasted: { nodes: Node[]; edges: Edge[] } | null = null
    act(() => {
      pasted = hook.result.current.paste()
    })

    expect(pasted!.edges).toHaveLength(0) // Edge should be filtered out
  })

  it('can paste multiple times', () => {
    const nodes: Node[] = [
      {
        id: 'node-1',
        type: 'custom',
        position: { x: 100, y: 100 },
        data: { label: 'Test Node', nodeType: 'inject' },
      },
    ]
    const edges: Edge[] = []

    act(() => {
      hook.result.current.copy(nodes, edges)
    })

    let pasted1: { nodes: Node[]; edges: Edge[] } | null = null
    act(() => {
      pasted1 = hook.result.current.paste()
    })

    let pasted2: { nodes: Node[]; edges: Edge[] } | null = null
    act(() => {
      pasted2 = hook.result.current.paste()
    })

    expect(pasted1).not.toBeNull()
    expect(pasted2).not.toBeNull()
    expect(pasted1!.nodes[0].id).not.toBe(pasted2!.nodes[0].id)
  })

  it('cuts nodes by copying them', () => {
    const nodes: Node[] = [
      {
        id: 'node-1',
        type: 'custom',
        position: { x: 100, y: 100 },
        data: { label: 'Test Node', nodeType: 'inject' },
      },
    ]
    const edges: Edge[] = []

    act(() => {
      hook.result.current.cut(nodes, edges)
    })

    expect(hook.result.current.canPaste).toBe(true)
  })

  it('returns null when pasting from empty clipboard', () => {
    let pasted: { nodes: Node[]; edges: Edge[] } | null = null
    act(() => {
      pasted = hook.result.current.paste()
    })

    expect(pasted).toBeNull()
  })

  it('ignores copy when no nodes selected', () => {
    act(() => {
      hook.result.current.copy([], [])
    })

    expect(hook.result.current.canPaste).toBe(false)
  })

  it('pastes nodes with custom position', () => {
    const nodes: Node[] = [
      {
        id: 'node-1',
        type: 'custom',
        position: { x: 100, y: 100 },
        data: { label: 'Test Node', nodeType: 'inject' },
      },
    ]
    const edges: Edge[] = []

    act(() => {
      hook.result.current.copy(nodes, edges)
    })

    let pasted: { nodes: Node[]; edges: Edge[] } | null = null
    act(() => {
      pasted = hook.result.current.paste({ x: 500, y: 500 })
    })

    expect(pasted!.nodes[0].position.x).toBe(500)
    expect(pasted!.nodes[0].position.y).toBe(500)
  })
})
