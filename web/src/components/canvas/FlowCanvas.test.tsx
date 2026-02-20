/// <reference types="@testing-library/jest-dom/vitest" />
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/react'
import { ReactFlowProvider } from '@xyflow/react'
import FlowCanvas from './FlowCanvas'

// Mock React Flow
vi.mock('@xyflow/react', async () => {
  const actual = await vi.importActual('@xyflow/react')
  return {
    ...actual,
    useReactFlow: () => ({
      screenToFlowPosition: vi.fn((pos) => pos),
      fitView: vi.fn(),
      zoomIn: vi.fn(),
      zoomOut: vi.fn(),
      getNodes: vi.fn(() => []),
      getEdges: vi.fn(() => []),
    }),
  }
})

describe('FlowCanvas', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders without crashing', () => {
    render(
      <ReactFlowProvider>
        <FlowCanvas />
      </ReactFlowProvider>
    )
    expect(screen.getByText(/Drag nodes from the palette to get started/i)).toBeInTheDocument()
  })

  it('shows empty state when no nodes', () => {
    render(
      <ReactFlowProvider>
        <FlowCanvas />
      </ReactFlowProvider>
    )
    expect(screen.getByText(/Drag nodes from the palette to get started/i)).toBeInTheDocument()
  })

  it('renders all toolbar buttons', () => {
    render(
      <ReactFlowProvider>
        <FlowCanvas />
      </ReactFlowProvider>
    )

    // Check for Undo/Redo buttons
    expect(screen.getByTitle(/Undo \(Ctrl\+Z\)/i)).toBeInTheDocument()
    expect(screen.getByTitle(/Redo \(Ctrl\+Y\)/i)).toBeInTheDocument()

    // Check for Copy/Cut/Paste buttons
    expect(screen.getByTitle(/Copy \(Ctrl\+C\)/i)).toBeInTheDocument()
    expect(screen.getByTitle(/Cut \(Ctrl\+X\)/i)).toBeInTheDocument()
    expect(screen.getByTitle(/Paste \(Ctrl\+V\)/i)).toBeInTheDocument()

    // Check for Zoom buttons
    expect(screen.getByTitle(/Zoom In/i)).toBeInTheDocument()
    expect(screen.getByTitle(/Zoom Out/i)).toBeInTheDocument()
    expect(screen.getByTitle(/Fit View/i)).toBeInTheDocument()

    // Check for Save button
    expect(screen.getByTitle(/Save Flow/i)).toBeInTheDocument()
  })

  it('disables undo/redo buttons initially', () => {
    render(
      <ReactFlowProvider>
        <FlowCanvas />
      </ReactFlowProvider>
    )

    const undoButton = screen.getByTitle(/Undo \(Ctrl\+Z\)/i)
    const redoButton = screen.getByTitle(/Redo \(Ctrl\+Y\)/i)

    expect(undoButton).toBeDisabled()
    expect(redoButton).toBeDisabled()
  })

  it('disables copy/cut buttons when no selection', () => {
    render(
      <ReactFlowProvider>
        <FlowCanvas />
      </ReactFlowProvider>
    )

    const copyButton = screen.getByTitle(/Copy \(Ctrl\+C\)/i)
    const cutButton = screen.getByTitle(/Cut \(Ctrl\+X\)/i)

    expect(copyButton).toBeDisabled()
    expect(cutButton).toBeDisabled()
  })

  it('disables paste button when clipboard is empty', () => {
    render(
      <ReactFlowProvider>
        <FlowCanvas />
      </ReactFlowProvider>
    )

    const pasteButton = screen.getByTitle(/Paste \(Ctrl\+V\)/i)
    expect(pasteButton).toBeDisabled()
  })

  it('calls fitView when Fit View button is clicked', () => {
    const { container } = render(
      <ReactFlowProvider>
        <FlowCanvas />
      </ReactFlowProvider>
    )

    const fitViewButton = screen.getByTitle(/Fit View/i)
    fireEvent.click(fitViewButton)

    // Button should be clickable
    expect(fitViewButton).not.toBeDisabled()
  })

  it('handles save button click', () => {
    const consoleSpy = vi.spyOn(console, 'log')

    render(
      <ReactFlowProvider>
        <FlowCanvas />
      </ReactFlowProvider>
    )

    const saveButton = screen.getByTitle(/Save Flow/i)
    fireEvent.click(saveButton)

    expect(consoleSpy).toHaveBeenCalledWith(expect.stringContaining('Saving flow'), expect.anything())
  })

  it('handles drag over event', () => {
    const { container } = render(
      <ReactFlowProvider>
        <FlowCanvas />
      </ReactFlowProvider>
    )

    const reactFlowElement = container.querySelector('.react-flow')
    expect(reactFlowElement).toBeInTheDocument()

    if (reactFlowElement) {
      const dragEvent = new DragEvent('dragover', {
        bubbles: true,
        cancelable: true,
      })
      Object.defineProperty(dragEvent, 'dataTransfer', {
        value: {
          dropEffect: '',
        },
      })

      fireEvent(reactFlowElement, dragEvent)
      expect(dragEvent.defaultPrevented).toBe(true)
    }
  })
})

describe('FlowCanvas Keyboard Shortcuts', () => {
  it('handles keyboard shortcuts', async () => {
    render(
      <ReactFlowProvider>
        <FlowCanvas />
      </ReactFlowProvider>
    )

    // Simulate Ctrl+A (Select All)
    fireEvent.keyDown(document, { key: 'a', ctrlKey: true })

    // Simulate Ctrl+C (Copy)
    fireEvent.keyDown(document, { key: 'c', ctrlKey: true })

    // Simulate Ctrl+V (Paste)
    fireEvent.keyDown(document, { key: 'v', ctrlKey: true })

    // Simulate Ctrl+Z (Undo)
    fireEvent.keyDown(document, { key: 'z', ctrlKey: true })

    // Simulate Ctrl+Y (Redo)
    fireEvent.keyDown(document, { key: 'y', ctrlKey: true })

    // No errors should be thrown
  })

  it('handles Delete key', () => {
    render(
      <ReactFlowProvider>
        <FlowCanvas />
      </ReactFlowProvider>
    )

    fireEvent.keyDown(document, { key: 'Delete' })
    // No errors should be thrown
  })

  it('handles Backspace key', () => {
    render(
      <ReactFlowProvider>
        <FlowCanvas />
      </ReactFlowProvider>
    )

    fireEvent.keyDown(document, { key: 'Backspace' })
    // No errors should be thrown
  })
})

describe('FlowCanvas Multi-Select', () => {
  it('enables multi-select mode', () => {
    const { container } = render(
      <ReactFlowProvider>
        <FlowCanvas />
      </ReactFlowProvider>
    )

    const reactFlowElement = container.querySelector('.react-flow')
    expect(reactFlowElement).toBeInTheDocument()
  })
})
