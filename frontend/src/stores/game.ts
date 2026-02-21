import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { api, type GameState } from '@/services/api'
import { wsService } from '@/services/ws'

export const useGameStore = defineStore('game', () => {
  const state = ref<GameState | null>(null)
  const loading = ref(false)
  const error = ref('')
  const wsConnected = ref(false)

  // Selected pawn for move UI.
  const selectedPawn = ref<string | null>(null)
  // Accumulated path of hop destinations in current move sequence.
  const currentPath = ref<string[]>([])
  // Valid destinations for the selected pawn (flat list of full paths from ValidMoves).
  const validPaths = ref<string[][]>([])
  // Squares involved in the last opponent move (from + to), for highlighting.
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

  function setState(gs: GameState) {
    state.value = gs
  }

  async function submitMove(id: string, from: string, path: string[]) {
    loading.value = true
    error.value = ''
    try {
      const prevBoard = state.value?.board
      state.value = await api.makeMove(id, { from, path })
      clearSelection()
      lastAISquares.value = prevBoard ? diffColorSquares(prevBoard, state.value!.board, 'black') : []
    } catch (e: any) {
      error.value = e?.response?.data?.error ?? 'Invalid move'
      throw e
    } finally {
      loading.value = false
    }
  }

  function connectWS(gameId: string, nickname: string, myColor: 'white' | 'black'): void {
    wsService.connect(gameId, nickname)
    wsService.onMessage((msg) => {
      switch (msg.type) {
        case 'game_state':
          state.value = msg.payload
          break
        case 'player_joined':
          if (state.value) {
            state.value = { ...state.value, black_nick: msg.payload.black_nick, status: 'in_progress' }
          }
          break
        case 'move_made': {
          const prev = state.value?.board
          state.value = msg.payload.state
          if (prev && msg.payload.player !== myColor) {
            lastAISquares.value = diffColorSquares(prev, state.value!.board, msg.payload.player)
          }
          clearSelection()
          break
        }
        case 'game_over':
          if (state.value) {
            state.value = { ...state.value, winner: msg.payload.winner, status: 'finished' }
          }
          break
        case 'error':
          error.value = msg.payload.message
          break
      }
    })
    wsConnected.value = true
  }

  function submitMoveWS(from: string, path: string[]): void {
    error.value = ''
    wsService.sendMove(from, path)
    clearSelection()
  }

  function disconnectWS(): void {
    wsService.disconnect()
    wsConnected.value = false
  }

  // Returns squares that changed for `color` between two boards (from + to of the move).
  function diffColorSquares(prev: string[][], next: string[][], color: string): string[] {
    const changed: string[] = []
    for (let row = 0; row < 8; row++) {
      for (let col = 0; col < 8; col++) {
        if (prev[row][col] !== next[row][col] && (prev[row][col] === color || next[row][col] === color)) {
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
    wsConnected,
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
    setState,
    submitMove,
    connectWS,
    submitMoveWS,
    disconnectWS,
    selectPawn,
    clearSelection,
  }
})
