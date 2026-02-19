<template>
  <div class="board-wrapper">
    <div class="board-grid">
      <!-- Row labels (8→1) on the left -->
      <div class="label-corner" />
      <div v-for="c in 8" :key="'col-label-' + c" class="col-label">
        {{ colLabel(c - 1) }}
      </div>

      <template v-for="r in 8" :key="'row-' + r">
        <!-- Row number label -->
        <div class="row-label">{{ 9 - r }}</div>
        <!-- 8 squares in this row (displayed top→bottom = row 8→1 = index 7→0) -->
        <BoardSquare
          v-for="c in 8"
          :key="'sq-' + r + '-' + c"
          :col="c - 1"
          :row="7 - (r - 1)"
          :pawn="cellPawn(c - 1, 7 - (r - 1))"
          :highlight="cellHighlight(c - 1, 7 - (r - 1))"
          @click="onSquareClick(c - 1, 7 - (r - 1))"
        />
      </template>
    </div>

    <!-- Confirm button shown when mid-hop and more hops available -->
    <v-btn
      v-if="canCommitEarly"
      color="primary"
      class="mt-3"
      @click="commitMove"
    >
      Confirm move here
    </v-btn>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import BoardSquare from './BoardSquare.vue'

const props = defineProps<{
  board: string[][]  // board[row][col], row 0=rank1, col 0=A
  playerColor: 'white' | 'black'
  isMyTurn: boolean
  gameStatus: string
  // Move state passed from parent (GameView manages via gameStore)
  selectedPawn: string | null
  currentPath: string[]
  validPaths: string[][]
}>()

const emit = defineEmits<{
  pawnSelected: [pos: string, paths: string[][]]
  moveCommit: [from: string, path: string[]]
  pathExtended: [newPath: string[]]
}>()

function colLabel(col: number): string {
  return String.fromCharCode('A'.charCodeAt(0) + col)
}

function posStr(col: number, row: number): string {
  return colLabel(col) + (row + 1)
}

function parsePos(s: string): { col: number; row: number } {
  return { col: s.charCodeAt(0) - 'A'.charCodeAt(0), row: parseInt(s[1]) - 1 }
}

function cellPawn(col: number, row: number): 'white' | 'black' | 'empty' {
  if (!props.board) return 'empty'
  const v = props.board[row]?.[col]
  if (v === 'white') return 'white'
  if (v === 'black') return 'black'
  return 'empty'
}

// ------------------------------------------------------------------
// Move calculation (client-side mirror of backend game/moves.go)
// ------------------------------------------------------------------

function directions(player: 'white' | 'black'): { col: number; row: number }[] {
  return player === 'white'
    ? [{ col: 0, row: 1 }, { col: -1, row: 0 }]
    : [{ col: 0, row: -1 }, { col: 1, row: 0 }]
}

function inBounds(col: number, row: number): boolean {
  return col >= 0 && col <= 7 && row >= 0 && row <= 7
}

function computeValidPaths(board: string[][], fromStr: string, player: 'white' | 'black'): string[][] {
  const from = parsePos(fromStr)
  const result: string[][] = []
  const dirs = directions(player)

  // Single-step moves.
  for (const dir of dirs) {
    const nc = from.col + dir.col
    const nr = from.row + dir.row
    if (inBounds(nc, nr) && board[nr][nc] === 'empty') {
      result.push([posStr(nc, nr)])
    }
  }

  // Hop moves.
  const visited = new Set<string>([fromStr])
  const hopFrom = (cc: number, cr: number, path: string[]) => {
    for (const dir of dirs) {
      const mc = cc + dir.col; const mr = cr + dir.row
      const lc = cc + 2 * dir.col; const lr = cr + 2 * dir.row
      if (!inBounds(mc, mr) || !inBounds(lc, lr)) continue
      if (board[mr][mc] === 'empty') continue
      if (board[lr][lc] !== 'empty') continue
      const ls = posStr(lc, lr)
      if (visited.has(ls)) continue
      const newPath = [...path, ls]
      result.push(newPath)
      visited.add(ls)
      hopFrom(lc, lr, newPath)
      visited.delete(ls)
    }
  }
  hopFrom(from.col, from.row, [])

  return result
}

// ------------------------------------------------------------------
// Next-step destinations for highlighting.
// Returns the next squares reachable given the current accumulated path.
// ------------------------------------------------------------------
function nextDestinations(validPaths: string[][], currentPath: string[]): string[] {
  const dests = new Set<string>()
  for (const path of validPaths) {
    if (path.length <= currentPath.length) continue
    let matches = true
    for (let i = 0; i < currentPath.length; i++) {
      if (path[i] !== currentPath[i]) { matches = false; break }
    }
    if (matches) dests.add(path[currentPath.length])
  }
  return [...dests]
}

// Whether currentPath is itself a complete valid path (can commit here).
const currentPathIsComplete = computed(() => {
  if (!props.currentPath.length) return false
  return props.validPaths.some(p =>
    p.length === props.currentPath.length &&
    p.every((v, i) => v === props.currentPath[i])
  )
})

// Are there further hops from the current position?
const hasMoreHops = computed(() => {
  return nextDestinations(props.validPaths, props.currentPath).length > 0
})

// Show "Confirm here" button when we're mid-hop and CAN stop here AND CAN continue.
const canCommitEarly = computed(() =>
  currentPathIsComplete.value && hasMoreHops.value && props.selectedPawn
)

// ------------------------------------------------------------------
// Highlight computation
// ------------------------------------------------------------------
function cellHighlight(col: number, row: number): 'none' | 'selected' | 'step' | 'hop' {
  const pos = posStr(col, row)

  // Currently selected pawn.
  if (props.selectedPawn === pos) return 'selected'

  if (!props.selectedPawn) return 'none'

  const nexts = nextDestinations(props.validPaths, props.currentPath)
  if (!nexts.includes(pos)) return 'none'

  // Is this a single-step or hop destination?
  // A destination is "step" if it appears in a path of length 1 (single step from original pawn).
  // A destination is "hop" if it only appears in paths of length > 1.
  const nextIndex = props.currentPath.length // index in the path for this next destination

  const isStep = props.validPaths.some(p =>
    p.length === nextIndex + 1 &&
    p[nextIndex] === pos &&
    (nextIndex === 0 ? isSingleStep(props.selectedPawn!, pos) : true)
  )
  const isHop = props.validPaths.some(p =>
    p.length > nextIndex + 1 &&
    (props.currentPath.every((v, i) => p[i] === v)) &&
    p[nextIndex] === pos
  )

  if (props.currentPath.length > 0) {
    // In hop chain: all next options are hops.
    return 'hop'
  }
  if (isStep && !isHop) return 'step'
  return 'hop'
}

function isSingleStep(from: string, to: string): boolean {
  const f = parsePos(from)
  const t = parsePos(to)
  const dc = Math.abs(t.col - f.col)
  const dr = Math.abs(t.row - f.row)
  return (dc === 0 && dr === 1) || (dc === 1 && dr === 0)
}

// ------------------------------------------------------------------
// Click handler
// ------------------------------------------------------------------
function onSquareClick(col: number, row: number) {
  if (!props.isMyTurn || props.gameStatus !== 'in_progress') return

  const pos = posStr(col, row)
  const pawn = cellPawn(col, row)

  // Clicking own pawn: select it.
  if (pawn === props.playerColor) {
    const paths = computeValidPaths(props.board, pos, props.playerColor)
    emit('pawnSelected', pos, paths)
    return
  }

  // If a pawn is selected, check if this is a valid next destination.
  if (props.selectedPawn) {
    const nexts = nextDestinations(props.validPaths, props.currentPath)
    if (!nexts.includes(pos)) {
      // Click on empty/invalid square: deselect.
      emit('pawnSelected', '', [])
      return
    }

    const newPath = [...props.currentPath, pos]

    // Are there further hops from this position?
    const furtherHops = nextDestinations(props.validPaths, newPath)

    if (furtherHops.length === 0) {
      // No more hops: commit immediately.
      emit('moveCommit', props.selectedPawn, newPath)
    } else {
      // More hops possible: extend the path and let player choose to continue or confirm.
      emit('pathExtended', newPath)
    }
  }
}

function commitMove() {
  if (props.selectedPawn && props.currentPath.length > 0) {
    emit('moveCommit', props.selectedPawn, props.currentPath)
  }
}
</script>

<style scoped>
.board-wrapper {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.board-grid {
  display: grid;
  grid-template-columns: 24px repeat(8, 1fr);
  grid-template-rows: 24px repeat(8, 1fr);
  width: min(80vw, 560px);
  height: min(80vw, 560px);
  border: 2px solid #555;
  border-radius: 4px;
  overflow: hidden;
}

.label-corner { background: #222; }

.col-label {
  background: #222;
  color: #ccc;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 11px;
  font-weight: bold;
}

.row-label {
  background: #222;
  color: #ccc;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 11px;
  font-weight: bold;
}
</style>
