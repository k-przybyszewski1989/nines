<template>
  <v-container fluid class="pa-4">
    <v-row justify="center">
      <v-col cols="12" md="8" lg="6">

        <!-- Loading -->
        <div v-if="gameStore.loading" class="d-flex justify-center align-center" style="height: 80vh">
          <v-progress-circular indeterminate size="64" />
        </div>

        <!-- Error -->
        <v-alert v-else-if="gameStore.error && !gameStore.state" type="error" class="mb-4">
          {{ gameStore.error }}
        </v-alert>

        <template v-else-if="gameStore.state">
          <!-- Black player (top) -->
          <PlayerInfo
            class="mb-3"
            :name="gameStore.state.black_nick || 'AI'"
            color="black"
            :move-count="blackMoves"
            :is-active="gameStore.turn === 'black'"
          />

          <!-- Board -->
          <GameBoard
            class="my-2"
            :board="gameStore.board!"
            :player-color="playerStore.color as 'white' | 'black'"
            :is-my-turn="gameStore.turn === 'white' && gameStore.status === 'in_progress'"
            :game-status="gameStore.status"
            :selected-pawn="gameStore.selectedPawn"
            :current-path="gameStore.currentPath"
            :valid-paths="gameStore.validPaths"
            @pawn-selected="onPawnSelected"
            @move-commit="onMoveCommit"
            @path-extended="onPathExtended"
          />

          <!-- White player (bottom) -->
          <PlayerInfo
            class="mt-3"
            :name="playerStore.nickname"
            color="white"
            :move-count="playerStore.moveCount"
            :is-active="gameStore.turn === 'white'"
          />

          <!-- Status / turn indicator -->
          <v-row class="mt-3" justify="center">
            <v-col cols="auto">
              <v-chip
                v-if="gameStore.status === 'in_progress'"
                :color="gameStore.turn === 'white' ? 'white' : 'deep-purple'"
                size="large"
              >
                {{ gameStore.turn === 'white' ? 'Your turn' : 'AI thinking…' }}
              </v-chip>
            </v-col>
          </v-row>

          <v-alert v-if="gameStore.error" type="warning" class="mt-3" closable @click:close="gameStore.error = ''">
            {{ gameStore.error }}
          </v-alert>
        </template>

      </v-col>
    </v-row>

    <!-- Win dialog -->
    <WinDialog
      v-if="gameStore.state"
      :winner="gameStore.winner"
      :white-nick="gameStore.state.white_nick"
      :black-nick="gameStore.state.black_nick || 'AI'"
    />
  </v-container>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
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

// Count moves from move_num split between players (human = white moves = ceil(move_num/2)).
const blackMoves = computed(() => {
  if (!gameStore.state) return 0
  return Math.floor(gameStore.state.move_num / 2)
})

onMounted(async () => {
  // If game is already in store (just created), use it; otherwise fetch.
  if (!gameStore.state || gameStore.state.id !== gameId.value) {
    await gameStore.fetchGame(gameId.value)
    // Determine player color from state.
    if (gameStore.state && !playerStore.color) {
      playerStore.setPlayer(gameStore.state.white_nick, 'white')
    }
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
  await gameStore.submitMove(gameId.value, from, path)
  playerStore.incrementMoves()
}
</script>
