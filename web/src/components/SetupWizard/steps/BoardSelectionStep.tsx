/**
 * Board Selection Step
 *
 * Select the IoT board model
 */

import { Cpu, Check, CircuitBoard, Zap, HardDrive } from 'lucide-react'
import { cn } from '@/lib/utils'
import { BoardType, BoardInfo, SUPPORTED_BOARDS } from '../types'

interface BoardSelectionStepProps {
  selectedBoard: BoardType | null
  onBoardSelect: (board: BoardType) => void
}

export function BoardSelectionStep({
  selectedBoard,
  onBoardSelect,
}: BoardSelectionStepProps) {
  // Group boards by type
  const raspberryPiBoards = SUPPORTED_BOARDS.filter((b) =>
    b.id.startsWith('rpi')
  )
  const otherBoards = SUPPORTED_BOARDS.filter(
    (b) => !b.id.startsWith('rpi') && b.id !== 'custom'
  )
  const customBoard = SUPPORTED_BOARDS.find((b) => b.id === 'custom')

  return (
    <div className="space-y-6">
      <div className="text-center space-y-2">
        <h2 className="text-2xl font-bold text-gray-900 dark:text-white">
          Select Your Board
        </h2>
        <p className="text-muted-foreground">
          Choose your IoT board to optimize EdgeFlow configuration
        </p>
      </div>

      {/* Raspberry Pi Boards */}
      <div className="space-y-3">
        <h3 className="text-sm font-semibold text-muted-foreground uppercase tracking-wide flex items-center gap-2">
          <CircuitBoard className="w-4 h-4" />
          Raspberry Pi
        </h3>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          {raspberryPiBoards.map((board) => (
            <BoardCard
              key={board.id}
              board={board}
              selected={selectedBoard === board.id}
              onSelect={() => onBoardSelect(board.id)}
            />
          ))}
        </div>
      </div>

      {/* Other Boards */}
      <div className="space-y-3">
        <h3 className="text-sm font-semibold text-muted-foreground uppercase tracking-wide flex items-center gap-2">
          <Cpu className="w-4 h-4" />
          Other Boards
        </h3>
        <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
          {otherBoards.map((board) => (
            <BoardCard
              key={board.id}
              board={board}
              selected={selectedBoard === board.id}
              onSelect={() => onBoardSelect(board.id)}
            />
          ))}
        </div>
      </div>

      {/* Custom Board */}
      {customBoard && (
        <div className="space-y-3">
          <h3 className="text-sm font-semibold text-muted-foreground uppercase tracking-wide flex items-center gap-2">
            <HardDrive className="w-4 h-4" />
            Custom
          </h3>
          <BoardCard
            board={customBoard}
            selected={selectedBoard === customBoard.id}
            onSelect={() => onBoardSelect(customBoard.id)}
          />
        </div>
      )}

      {/* Selected Board Info */}
      {selectedBoard && (
        <div className="mt-6 p-4 bg-primary/5 border border-primary/20 rounded-xl">
          <div className="flex items-start gap-3">
            <div className="w-10 h-10 bg-primary rounded-lg flex items-center justify-center flex-shrink-0">
              <Check className="w-5 h-5 text-primary-foreground" />
            </div>
            <div>
              <h4 className="font-semibold text-gray-900 dark:text-white">
                {SUPPORTED_BOARDS.find((b) => b.id === selectedBoard)?.name}
              </h4>
              <p className="text-sm text-muted-foreground mt-1">
                EdgeFlow will be optimized for this board's capabilities
              </p>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

interface BoardCardProps {
  board: BoardInfo
  selected: boolean
  onSelect: () => void
}

function BoardCard({ board, selected, onSelect }: BoardCardProps) {
  return (
    <button
      onClick={onSelect}
      className={cn(
        'w-full text-left p-4 rounded-xl border-2 transition-all duration-200',
        'hover:shadow-md hover:border-primary/50',
        selected
          ? 'border-primary bg-primary/5 shadow-md'
          : 'border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800/50'
      )}
    >
      <div className="flex items-start justify-between">
        <div className="flex-1">
          <div className="flex items-center gap-2">
            <h4 className="font-semibold text-gray-900 dark:text-white">
              {board.name}
            </h4>
            {selected && (
              <span className="w-5 h-5 bg-primary rounded-full flex items-center justify-center">
                <Check className="w-3 h-3 text-primary-foreground" />
              </span>
            )}
          </div>
          <p className="text-sm text-muted-foreground mt-1">
            {board.description}
          </p>
        </div>
      </div>

      {/* Board Specs */}
      <div className="mt-3 flex flex-wrap gap-2">
        <span className="inline-flex items-center gap-1 text-xs px-2 py-1 bg-gray-100 dark:bg-gray-700 rounded-md">
          <Cpu className="w-3 h-3" />
          {board.cpu}
        </span>
        <span className="inline-flex items-center gap-1 text-xs px-2 py-1 bg-gray-100 dark:bg-gray-700 rounded-md">
          <HardDrive className="w-3 h-3" />
          {board.ram}
        </span>
        {board.gpioCount > 0 && (
          <span className="inline-flex items-center gap-1 text-xs px-2 py-1 bg-green-100 dark:bg-green-900/30 text-green-700 dark:text-green-300 rounded-md">
            <Zap className="w-3 h-3" />
            {board.gpioCount} GPIO
          </span>
        )}
      </div>
    </button>
  )
}
