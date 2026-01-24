import { describe, it, expect, beforeEach } from 'vitest'
import { renderHook, act } from '@testing-library/react'
import { useUndoRedo } from './useUndoRedo'

describe('useUndoRedo', () => {
  it('initializes with initial state', () => {
    const initialState = { value: 0 }
    const { result } = renderHook(() => useUndoRedo(initialState))

    const [state, actions] = result.current
    expect(state).toEqual(initialState)
    expect(actions.canUndo).toBe(false)
    expect(actions.canRedo).toBe(false)
  })

  it('pushes new state to history', () => {
    const initialState = { value: 0 }
    const { result } = renderHook(() => useUndoRedo(initialState))

    act(() => {
      result.current[1].push({ value: 1 })
    })

    const [state, actions] = result.current
    expect(state).toEqual({ value: 1 })
    expect(actions.canUndo).toBe(true)
    expect(actions.canRedo).toBe(false)
  })

  it('undoes to previous state', () => {
    const initialState = { value: 0 }
    const { result } = renderHook(() => useUndoRedo(initialState))

    act(() => {
      result.current[1].push({ value: 1 })
    })

    act(() => {
      result.current[1].undo()
    })

    const [state, actions] = result.current
    expect(state).toEqual({ value: 0 })
    expect(actions.canUndo).toBe(false)
    expect(actions.canRedo).toBe(true)
  })

  it('redoes to next state', () => {
    const initialState = { value: 0 }
    const { result } = renderHook(() => useUndoRedo(initialState))

    act(() => {
      result.current[1].push({ value: 1 })
    })

    act(() => {
      result.current[1].undo()
    })

    act(() => {
      result.current[1].redo()
    })

    const [state, actions] = result.current
    expect(state).toEqual({ value: 1 })
    expect(actions.canUndo).toBe(true)
    expect(actions.canRedo).toBe(false)
  })

  it('handles multiple undo operations', () => {
    const initialState = { value: 0 }
    const { result } = renderHook(() => useUndoRedo(initialState))

    act(() => {
      result.current[1].push({ value: 1 })
      result.current[1].push({ value: 2 })
      result.current[1].push({ value: 3 })
    })

    act(() => {
      result.current[1].undo()
    })

    let [state] = result.current
    expect(state).toEqual({ value: 2 })

    act(() => {
      result.current[1].undo()
    })

    ;[state] = result.current
    expect(state).toEqual({ value: 1 })

    act(() => {
      result.current[1].undo()
    })

    ;[state] = result.current
    expect(state).toEqual({ value: 0 })
  })

  it('handles multiple redo operations', () => {
    const initialState = { value: 0 }
    const { result } = renderHook(() => useUndoRedo(initialState))

    act(() => {
      result.current[1].push({ value: 1 })
      result.current[1].push({ value: 2 })
      result.current[1].push({ value: 3 })
    })

    act(() => {
      result.current[1].undo()
      result.current[1].undo()
      result.current[1].undo()
    })

    act(() => {
      result.current[1].redo()
    })

    let [state] = result.current
    expect(state).toEqual({ value: 1 })

    act(() => {
      result.current[1].redo()
    })

    ;[state] = result.current
    expect(state).toEqual({ value: 2 })

    act(() => {
      result.current[1].redo()
    })

    ;[state] = result.current
    expect(state).toEqual({ value: 3 })
  })

  it('clears future when pushing after undo', () => {
    const initialState = { value: 0 }
    const { result } = renderHook(() => useUndoRedo(initialState))

    act(() => {
      result.current[1].push({ value: 1 })
      result.current[1].push({ value: 2 })
    })

    act(() => {
      result.current[1].undo()
    })

    act(() => {
      result.current[1].push({ value: 3 })
    })

    const [state, actions] = result.current
    expect(state).toEqual({ value: 3 })
    expect(actions.canRedo).toBe(false)
  })

  it('resets to new state', () => {
    const initialState = { value: 0 }
    const { result } = renderHook(() => useUndoRedo(initialState))

    act(() => {
      result.current[1].push({ value: 1 })
      result.current[1].push({ value: 2 })
    })

    act(() => {
      result.current[1].reset({ value: 10 })
    })

    const [state, actions] = result.current
    expect(state).toEqual({ value: 10 })
    expect(actions.canUndo).toBe(false)
    expect(actions.canRedo).toBe(false)
  })

  it('limits history to MAX_HISTORY entries', () => {
    const initialState = { value: 0 }
    const { result } = renderHook(() => useUndoRedo(initialState))

    // Push 60 states (more than MAX_HISTORY of 50)
    act(() => {
      for (let i = 1; i <= 60; i++) {
        result.current[1].push({ value: i })
      }
    })

    // Should only be able to undo 50 times max
    let undoCount = 0
    act(() => {
      while (result.current[1].canUndo && undoCount < 100) {
        result.current[1].undo()
        undoCount++
      }
    })

    expect(undoCount).toBeLessThanOrEqual(50)
  })

  it('does nothing when undoing at start of history', () => {
    const initialState = { value: 0 }
    const { result } = renderHook(() => useUndoRedo(initialState))

    act(() => {
      result.current[1].undo()
    })

    const [state, actions] = result.current
    expect(state).toEqual(initialState)
    expect(actions.canUndo).toBe(false)
  })

  it('does nothing when redoing at end of history', () => {
    const initialState = { value: 0 }
    const { result } = renderHook(() => useUndoRedo(initialState))

    act(() => {
      result.current[1].push({ value: 1 })
    })

    act(() => {
      result.current[1].redo()
    })

    const [state, actions] = result.current
    expect(state).toEqual({ value: 1 })
    expect(actions.canRedo).toBe(false)
  })

  it('handles complex state objects', () => {
    const initialState = {
      nodes: [{ id: '1', data: 'node1' }],
      edges: [{ id: 'e1', source: '1', target: '2' }],
    }
    const { result } = renderHook(() => useUndoRedo(initialState))

    act(() => {
      result.current[1].push({
        nodes: [
          { id: '1', data: 'node1' },
          { id: '2', data: 'node2' },
        ],
        edges: [{ id: 'e1', source: '1', target: '2' }],
      })
    })

    act(() => {
      result.current[1].undo()
    })

    const [state] = result.current
    expect(state.nodes).toHaveLength(1)
    expect(state.edges).toHaveLength(1)
  })
})
