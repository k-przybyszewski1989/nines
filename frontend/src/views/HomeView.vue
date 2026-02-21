<template>
  <v-container class="fill-height" fluid>
    <v-row justify="center" align="center" class="fill-height">
      <v-col cols="12" sm="8" md="5" lg="4">
        <v-card elevation="8" rounded="xl">
          <v-card-title class="text-h4 text-center pa-6 font-weight-bold">
            ♟ Nines
          </v-card-title>
          <v-card-text class="pa-6">
            <v-text-field
              v-model="nickname"
              label="Your nickname"
              variant="outlined"
              prepend-inner-icon="mdi-account"
              :rules="[v => !!v || 'Nickname is required']"
              class="mb-4"
              @keydown.enter="step === 'mode' ? null : null"
            />

            <!-- Step: choose mode -->
            <template v-if="step === 'mode'">
              <div class="text-subtitle-1 mb-3">Select game mode</div>
              <v-btn
                block
                color="primary"
                size="large"
                class="mb-3"
                :disabled="!nickname.trim()"
                @click="step = 'difficulty'"
              >
                <v-icon start>mdi-robot</v-icon>
                Singleplayer
              </v-btn>
              <v-btn
                block
                color="secondary"
                size="large"
                variant="outlined"
                :disabled="!nickname.trim()"
                @click="goMultiplayer"
              >
                <v-icon start>mdi-account-multiple</v-icon>
                Multiplayer
              </v-btn>
            </template>

            <!-- Step: choose difficulty -->
            <template v-else-if="step === 'difficulty'">
              <div class="text-subtitle-1 mb-3">Select difficulty</div>
              <v-btn
                block
                color="green"
                size="large"
                class="mb-3"
                :loading="loading && aiLevel === 'easy'"
                @click="startGame('easy')"
              >
                Easy
              </v-btn>
              <v-btn
                block
                color="orange"
                size="large"
                class="mb-3"
                :loading="loading && aiLevel === 'medium'"
                @click="startGame('medium')"
              >
                Medium
              </v-btn>
              <v-btn
                block
                color="red"
                size="large"
                class="mb-3"
                :loading="loading && aiLevel === 'hard'"
                @click="startGame('hard')"
              >
                Hard
              </v-btn>
              <v-btn variant="text" block @click="step = 'mode'">← Back</v-btn>
            </template>

            <v-alert v-if="error" type="error" class="mt-3">{{ error }}</v-alert>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useGameStore } from '@/stores/game'
import { usePlayerStore } from '@/stores/player'

const router = useRouter()
const gameStore = useGameStore()
const playerStore = usePlayerStore()

const nickname = ref('')
const step = ref<'mode' | 'difficulty'>('mode')
const aiLevel = ref('')
const loading = ref(false)
const error = ref('')

function goMultiplayer() {
  if (!nickname.value.trim()) return
  router.push({ name: 'lobby', query: { nickname: nickname.value.trim() } })
}

async function startGame(level: string) {
  if (!nickname.value.trim()) return
  aiLevel.value = level
  loading.value = true
  error.value = ''
  try {
    await gameStore.createGame('singleplayer', nickname.value.trim(), level)
    playerStore.setPlayer(nickname.value.trim(), 'white')
    router.push({ name: 'game', params: { id: gameStore.state!.id } })
  } catch (e: any) {
    error.value = e?.response?.data?.error ?? 'Failed to start game'
  } finally {
    loading.value = false
  }
}
</script>
