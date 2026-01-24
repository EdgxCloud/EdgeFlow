/**
 * Position Conversion Utilities
 *
 * Converts between n8n-style [x, y] array format and ReactFlow {x, y} object format
 */

// n8n-style position: [x, y] tuple
export type ArrayPosition = [number, number]

// ReactFlow position: {x, y} object
export interface ObjectPosition {
  x: number
  y: number
}

// Union type for either format
export type AnyPosition = ArrayPosition | ObjectPosition | undefined | null

/**
 * Check if position is in array format [x, y]
 */
export function isArrayPosition(position: AnyPosition): position is ArrayPosition {
  return Array.isArray(position) && position.length === 2 &&
         typeof position[0] === 'number' && typeof position[1] === 'number'
}

/**
 * Check if position is in object format {x, y}
 */
export function isObjectPosition(position: AnyPosition): position is ObjectPosition {
  return position !== null &&
         typeof position === 'object' &&
         !Array.isArray(position) &&
         'x' in position &&
         'y' in position &&
         typeof position.x === 'number' &&
         typeof position.y === 'number'
}

/**
 * Convert any position format to ReactFlow object format {x, y}
 * Used when loading flows for ReactFlow canvas
 */
export function toObjectPosition(position: AnyPosition, defaultPos: ObjectPosition = { x: 0, y: 0 }): ObjectPosition {
  if (!position) return defaultPos

  if (isArrayPosition(position)) {
    return { x: position[0], y: position[1] }
  }

  if (isObjectPosition(position)) {
    return position
  }

  return defaultPos
}

/**
 * Convert any position format to n8n-style array format [x, y]
 * Used when saving flows for storage/API
 */
export function toArrayPosition(position: AnyPosition, defaultPos: ArrayPosition = [0, 0]): ArrayPosition {
  if (!position) return defaultPos

  if (isArrayPosition(position)) {
    return position
  }

  if (isObjectPosition(position)) {
    return [position.x, position.y]
  }

  return defaultPos
}

/**
 * Round position to grid snap (optional)
 */
export function snapToGrid(position: ObjectPosition, gridSize: number = 15): ObjectPosition {
  return {
    x: Math.round(position.x / gridSize) * gridSize,
    y: Math.round(position.y / gridSize) * gridSize
  }
}

/**
 * Round array position to grid snap
 */
export function snapArrayToGrid(position: ArrayPosition, gridSize: number = 15): ArrayPosition {
  return [
    Math.round(position[0] / gridSize) * gridSize,
    Math.round(position[1] / gridSize) * gridSize
  ]
}
