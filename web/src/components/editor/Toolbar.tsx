import { Save, Play, Square, Settings, Undo, Redo } from 'lucide-react'

interface ToolbarProps {
  onSave: () => void
  onDeploy: () => void
  onRun: () => void
  onStop: () => void
  isRunning?: boolean
  canUndo?: boolean
  canRedo?: boolean
  onUndo?: () => void
  onRedo?: () => void
}

export default function Toolbar({
  onSave,
  onDeploy,
  onRun,
  onStop,
  isRunning = false,
  canUndo = false,
  canRedo = false,
  onUndo,
  onRedo,
}: ToolbarProps) {
  return (
    <div className="h-14 bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 flex items-center justify-between px-4">
      {/* Left actions */}
      <div className="flex items-center space-x-2 ">
        <button
          onClick={onSave}
          className="flex items-center space-x-1  px-3 py-1.5 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors text-sm"
        >
          <Save className="w-4 h-4" />
          <span>Save</span>
        </button>

        <button
          onClick={onDeploy}
          className="flex items-center space-x-1  px-3 py-1.5 bg-green-600 hover:bg-green-700 text-white rounded-lg transition-colors text-sm"
        >
          <span>Deploy</span>
        </button>

        <div className="w-px h-6 bg-gray-300 dark:bg-gray-600 mx-1"></div>

        {isRunning ? (
          <button
            onClick={onStop}
            className="flex items-center space-x-1  px-3 py-1.5 bg-red-100 dark:bg-red-900/20 text-red-700 dark:text-red-400 hover:bg-red-200 dark:hover:bg-red-900/30 rounded-lg transition-colors text-sm"
          >
            <Square className="w-4 h-4" />
            <span>Stop</span>
          </button>
        ) : (
          <button
            onClick={onRun}
            className="flex items-center space-x-1  px-3 py-1.5 bg-green-100 dark:bg-green-900/20 text-green-700 dark:text-green-400 hover:bg-green-200 dark:hover:bg-green-900/30 rounded-lg transition-colors text-sm"
          >
            <Play className="w-4 h-4" />
            <span>Run</span>
          </button>
        )}
      </div>

      {/* Center actions */}
      <div className="flex items-center space-x-2 ">
        <button
          onClick={onUndo}
          disabled={!canUndo}
          className={`p-2 rounded-lg transition-colors ${
            canUndo
              ? 'hover:bg-gray-100 dark:hover:bg-gray-700 text-gray-700 dark:text-gray-300'
              : 'text-gray-400 dark:text-gray-600 cursor-not-allowed'
          }`}
        >
          <Undo className="w-4 h-4" />
        </button>

        <button
          onClick={onRedo}
          disabled={!canRedo}
          className={`p-2 rounded-lg transition-colors ${
            canRedo
              ? 'hover:bg-gray-100 dark:hover:bg-gray-700 text-gray-700 dark:text-gray-300'
              : 'text-gray-400 dark:text-gray-600 cursor-not-allowed'
          }`}
        >
          <Redo className="w-4 h-4" />
        </button>
      </div>

      {/* Right actions */}
      <div className="flex items-center space-x-2 ">
        <button className="p-2 hover:bg-gray-100 dark:hover:bg-gray-700 rounded-lg transition-colors text-gray-700 dark:text-gray-300">
          <Settings className="w-4 h-4" />
        </button>
      </div>
    </div>
  )
}
