import React from 'react'

interface ResourceBarProps {
  used: number
  total: number
  type: 'memory' | 'disk'
  reserved?: number
}

const ResourceBar: React.FC<ResourceBarProps> = ({
  used,
  total,
  type,
  reserved = 0
}) => {
  const usedPercent = (used / total) * 100
  const reservedPercent = (reserved / total) * 100
  const availablePercent = 100 - usedPercent - reservedPercent

  const getColor = (percent: number) => {
    if (percent >= 90) return 'bg-red-500'
    if (percent >= 75) return 'bg-orange-500'
    if (percent >= 50) return 'bg-yellow-500'
    return 'bg-green-500'
  }

  const getReservedColor = () => {
    return type === 'memory' ? 'bg-blue-400' : 'bg-purple-400'
  }

  return (
    <div className="space-y-1">
      <div className="h-3 bg-gray-200 dark:bg-gray-700 rounded-full overflow-hidden flex">
        {/* Used */}
        <div
          className={`${getColor(usedPercent)} transition-all duration-300`}
          style={{ width: `${usedPercent}%` }}
          title={`Used: ${used}MB (${usedPercent.toFixed(1)}%)`}
        />
        {/* Reserved */}
        {reserved > 0 && (
          <div
            className={`${getReservedColor()} transition-all duration-300`}
            style={{ width: `${reservedPercent}%` }}
            title={`Reserved: ${reserved}MB (${reservedPercent.toFixed(1)}%)`}
          />
        )}
        {/* Available */}
        <div
          className="bg-gray-100 dark:bg-gray-800 transition-all duration-300"
          style={{ width: `${availablePercent}%` }}
          title={`Available: ${total - used - reserved}MB (${availablePercent.toFixed(1)}%)`}
        />
      </div>

      {/* Legend */}
      {reserved > 0 && (
        <div className="flex items-center gap-4 text-xs">
          <div className="flex items-center gap-1">
            <div className={`w-3 h-3 rounded ${getColor(usedPercent)}`} />
            <span className="text-gray-600 dark:text-gray-400">
              Used: {usedPercent.toFixed(1)}%
            </span>
          </div>
          <div className="flex items-center gap-1">
            <div className={`w-3 h-3 rounded ${getReservedColor()}`} />
            <span className="text-gray-600 dark:text-gray-400">
              Reserved: {reservedPercent.toFixed(1)}%
            </span>
          </div>
          <div className="flex items-center gap-1">
            <div className="w-3 h-3 rounded bg-gray-100 dark:bg-gray-800 border border-gray-300 dark:border-gray-600" />
            <span className="text-gray-600 dark:text-gray-400">
              Free: {availablePercent.toFixed(1)}%
            </span>
          </div>
        </div>
      )}
    </div>
  )
}

export default ResourceBar
