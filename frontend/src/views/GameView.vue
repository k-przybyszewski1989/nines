<template>
  <v-container
    fluid
    class="pa-4"
  >
    <v-row justify="center">
      <v-col
        cols="12"
        md="8"
        lg="6"
      >
        <!-- Loading -->
        <div
          v-if="gameStore.loading"
          class="d-flex justify-center align-center"
          style="height: 80vh"
        >
          <v-progress-circular
            indeterminate
            size="64"
          />
        </div>

        <!-- Error -->
        <v-alert
          v-else-if="gameStore.error && !gameStore.state"
          type="error"
          class="mb-4"
        >
          {{ gameStore.error }}
        </v-alert>

        <template v-else-if="gameStore.state">
          <!-- Waiting for opponent (multiplayer only) -->
          <template v-if="gameStore.status === 'waiting'">
            <v-card class="text-center pa-8 mb-4">
              <div class="text-h6 mb-3">
                Waiting for opponent...
              </div>
              <div class="text-subtitle-2 mb-2">
                Share this room code:
              </div>
              <div
                class="text-h3 font-weight-bold mb-4"
                style="letter-spacing: 0.2em"
              >
                {{ gameStore.state.room_code }}
              </div>
              <v-progress-circular
                indeterminate
                color="primary"
              />
            </v-card>
          </template>

          <template v-else>
            <!-- Black player (top) -->
            <PlayerInfo
              class="mb-3"
              :name="blackDisplayName"
              color="black"
              :move-count="blackMoves"
              :is-active="gameStore.turn === 'black'"
            />

            <!-- Board -->
            <GameBoard
              class="my-2"
              :board="gameStore.board!"
              :player-color="playerColor"
              :is-my-turn="isMyTurn"
              :game-status="gameStore.status"
              :selected-pawn="gameStore.selectedPawn"
              :current-path="gameStore.currentPath"
              :valid-paths="gameStore.validPaths"
              :last-ai-squares="gameStore.lastAISquares"
              @pawn-selected="onPawnSelected"
              @move-commit="onMoveCommit"
              @path-extended="onPathExtended"
            />

            <!-- White player (bottom) -->
            <PlayerInfo
              class="mt-3"
              :name="gameStore.state?.white_nick || playerStore.nickname"
              color="white"
              :move-count="playerStore.moveCount"
              :is-active="gameStore.turn === 'white'"
            />

            <!-- Status / turn indicator -->
            <v-row
              class="mt-3"
              justify="center"
            >
              <v-col cols="auto">
                <v-chip
                  v-if="gameStore.status === 'in_progress'"
                  :color="gameStore.turn === 'white' ? 'white' : 'deep-purple'"
                  size="large"
                >
                  {{ turnLabel }}
                </v-chip>
              </v-col>
            </v-row>

            <v-alert
              v-if="gameStore.error"
              type="warning"
              class="mt-3"
              closable
              @click:close="gameStore.error = ''"
            >
              {{ gameStore.error }}
            </v-alert>
          </template>
        </template>
      </v-col>
    </v-row>

    <!-- Win dialog -->
    <WinDialog
      v-if="gameStore.state"
      :winner="gameStore.winner"
      :white-nick="gameStore.state.white_nick"
      :black-nick="blackDisplayName"
    />
  </v-container>
</template>

<script setup lang="ts">
import { computed, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vue-router'
import GameBoard from '@/components/GameBoard.vue'
import PlayerInfo from '@/components/PlayerInfo.vue'
import WinDialog from '@/components/WinDialog.vue'
import { useGameStore } from '@/stores/game'
import { usePlayerStore } from '@/stores/player'

const route = useRoute()
const gameStore = useGameStore()
const playerStore = usePlayerStore()

const gameId = computed(() => route.params.id as string)
const myColor = computed(() => (route.query.color as string || 'white') as 'white' | 'black')
const playerColor = computed(() => playerStore.color as 'white' | 'black')

const blackDisplayName = computed(() => {
  if (!gameStore.state) return 'AI'
  if (gameStore.state.mode === 'multiplayer') {
    return gameStore.state.black_nick || 'Waiting...'
  }
  return gameStore.state.black_nick || 'AI'
})

// Count moves from move_num split between players.
const blackMoves = computed(() => {
  if (!gameStore.state) return 0
  return Math.floor(gameStore.state.move_num / 2)
})

const isMyTurn = computed(() => {
  if (gameStore.status !== 'in_progress') return false
  return gameStore.turn === playerStore.color
})

const turnLabel = computed(() => {
  if (gameStore.state?.mode === 'multiplayer') {
    return gameStore.turn === playerStore.color ? 'Your turn' : "Opponent's turn"
  }
  return gameStore.turn === 'white' ? 'Your turn' : 'AI thinking…'
})

onMounted(async () => {
  if (!gameStore.state || gameStore.state.id !== gameId.value) {
    await gameStore.fetchGame(gameId.value)
  }
  if (gameStore.state && !playerStore.color) {
    const nick = gameStore.state.mode === 'multiplayer'
      ? (myColor.value === 'white' ? gameStore.state.white_nick : gameStore.state.black_nick ?? '')
      : gameStore.state.white_nick
    playerStore.setPlayer(nick, myColor.value)
  }
  if (gameStore.state?.mode === 'multiplayer') {
    gameStore.connectWS(gameId.value, playerStore.nickname, playerStore.color as 'white' | 'black')
  }
})

onUnmounted(() => {
  if (gameStore.state?.mode === 'multiplayer') {
    gameStore.disconnectWS()
  }
})

function onPawnSelected(pos: string, paths: string[][]) {
  if (!pos) {
    gameStore.clearSelection()
    return
  }
  gameStore.selectPawn(pos, paths)
}

function onPathExtended(newPath: string[]) {
  gameStore.currentPath.splice(0, gameStore.currentPath.length, ...newPath)
}

async function onMoveCommit(from: string, path: string[]) {
  if (gameStore.state?.mode === 'multiplayer') {
    gameStore.submitMoveWS(from, path)
  } else {
    await gameStore.submitMove(gameId.value, from, path)
  }
  playerStore.incrementMoves()
}
</script>
