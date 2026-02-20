import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api, type GameState } from '@/services/api'

export const useGameStore = defineStore('game', () => {
  const state = ref<GameState | null>(null)
  const loading = ref(false)
  const error = ref('')

  // Selected pawn for move UI.
  const selectedPawn = ref<string | null>(null)
  // Accumulated path of hop destinations in current move sequence.
  const currentPath = ref<string[]>([])
  // Valid destinations for the selected pawn (flat list of full paths from ValidMoves).
  const validPaths = ref<string[][]>([])
  // Squares involved in the last AI move (from + to), for highlighting.
  const lastAISquares = ref<string[]>([])

  const board = computed(() => state.value?.board ?? null)
  const turn = computed(() => state.value?.turn ?? '')
  const status = computed(() => state.value?.status ?? '')
  const winner = computed(() => state.value?.winner ?? '')

  async function createGame(mode: 'singleplayer' | 'multiplayer', nickname: string, aiLevel?: string) {
    loading.value = true
    error.value = ''
    try {
      state.value = await api.createGame({ mode, nickname, ai_level: aiLevel as any })
    } catch (e: any) {
      error.value = e?.response?.data?.error ?? 'Failed to create game'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function fetchGame(id: string) {
    loading.value = true
    error.value = ''
    try {
      state.value = await api.getGame(id)
    } catch (e: any) {
      error.value = e?.response?.data?.error ?? 'Failed to load game'
      throw e
    } finally {
      loading.value = false
    }
  }

  async function submitMove(id: string, from: string, path: string[]) {
    loading.value = true
    error.value = ''
    try {
      const prevBoard = state.value?.board
      state.value = await api.makeMove(id, { from, path })
      clearSelection()
      lastAISquares.value = prevBoard ? diffBlackSquares(prevBoard, state.value!.board) : []
    } catch (e: any) {
      error.value = e?.response?.data?.error ?? 'Invalid move'
      throw e
    } finally {
      loading.value = false
    }
  }

  // Returns the squares that changed for black between two boards (from + to of AI move).
  function diffBlackSquares(prev: string[][], next: string[][]): string[] {
    const changed: string[] = []
    for (let row = 0; row < 8; row++) {
      for (let col = 0; col < 8; col++) {
        if (prev[row][col] !== next[row][col] && (prev[row][col] === 'black' || next[row][col] === 'black')) {
          changed.push(String.fromCharCode('A'.charCodeAt(0) + col) + (row + 1))
        }
      }
    }
    return changed
  }

  function selectPawn(pos: string, paths: string[][]) {
    selectedPawn.value = pos
    currentPath.value = []
    validPaths.value = paths
  }

  function clearSelection() {
    selectedPawn.value = null
    currentPath.value = []
    validPaths.value = []
  }

  return {
    state,
    loading,
    error,
    selectedPawn,
    currentPath,
    validPaths,
    lastAISquares,
    board,
    turn,
    status,
    winner,
    createGame,
    fetchGame,
    submitMove,
    selectPawn,
    clearSelection,
  }
})
