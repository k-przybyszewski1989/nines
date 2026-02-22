const COL_VALUES: Record<string, number> = { A: 1, B: 1, C: 2, D: 3, E: 3, F: 2, G: 1, H: 1 }
const ROW_VALUES: Record<number, number> = { 1: 1, 2: 1, 3: 2, 4: 3, 5: 3, 6: 2, 7: 1, 8: 1 }

export function baseValue(pos: string): number {
  const col = COL_VALUES[pos[0]] ?? 1
  const row = ROW_VALUES[parseInt(pos[1])] ?? 1
  return Math.max(col, row)
}

export function midpoint(a: string, b: string): string {
  const col = String.fromCharCode((a.charCodeAt(0) + b.charCodeAt(0)) / 2)
  const row = String((parseInt(a[1]) + parseInt(b[1])) / 2)
  return col + row
}

export function isSimpleStep(from: string, to: string): boolean {
  const colDiff = Math.abs(from.charCodeAt(0) - to.charCodeAt(0))
  const rowDiff = Math.abs(parseInt(from[1]) - parseInt(to[1]))
  return colDiff <= 1 && rowDiff <= 1 && colDiff + rowDiff > 0
}

export function isDiagonalHop(from: string, to: string): boolean {
  const colDiff = Math.abs(from.charCodeAt(0) - to.charCodeAt(0))
  const rowDiff = Math.abs(parseInt(from[1]) - parseInt(to[1]))
  return colDiff === 2 && rowDiff === 2
}

export function calculateMoveScore(from: string, path: string[]): number {
  if (path.length === 1 && isSimpleStep(from, path[0])) {
    return 1
  }

  const positions = [from, ...path]
  let diagonalSum = 0
  let diagonalCount = 0
  let straightSum = 0

  for (let i = 0; i < positions.length - 1; i++) {
    const mid = midpoint(positions[i], positions[i + 1])
    const val = baseValue(mid)
    if (isDiagonalHop(positions[i], positions[i + 1])) {
      diagonalSum += val
      diagonalCount++
    } else {
      straightSum += val
    }
  }

  return Math.pow(diagonalSum, diagonalCount + 1) + straightSum
}
