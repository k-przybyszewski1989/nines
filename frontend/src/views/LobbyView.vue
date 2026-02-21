<template>
  <v-container class="fill-height" fluid>
    <v-row justify="center" align="center" class="fill-height">
      <v-col cols="12" sm="9" md="6" lg="5">
        <v-card elevation="8" rounded="xl">
          <v-card-title class="text-h5 text-center pa-6 font-weight-bold">
            ♟ Multiplayer
          </v-card-title>

          <v-card-text class="pa-6">
            <v-tabs v-model="tab" grow class="mb-6">
              <v-tab value="create">Create Room</v-tab>
              <v-tab value="join">Join Room</v-tab>
            </v-tabs>

            <v-window v-model="tab">
              <!-- Create Room -->
              <v-window-item value="create">
                <v-text-field
                  v-model="createNickname"
                  label="Your nickname"
                  variant="outlined"
                  prepend-inner-icon="mdi-account"
                  class="mb-4"
                />
                <v-btn
                  block
                  color="primary"
                  size="large"
                  :loading="createLoading"
                  :disabled="!createNickname.trim()"
                  @click="createRoom"
                >
                  Create Room
                </v-btn>
                <v-alert v-if="createError" type="error" class="mt-3">{{ createError }}</v-alert>
              </v-window-item>

              <!-- Join Room -->
              <v-window-item value="join">
                <v-text-field
                  v-model="joinNickname"
                  label="Your nickname"
                  variant="outlined"
                  prepend-inner-icon="mdi-account"
                  class="mb-4"
                />
                <v-text-field
                  v-model="roomCode"
                  label="Room code"
                  variant="outlined"
                  prepend-inner-icon="mdi-key"
                  class="mb-4"
                  :rules="[v => v.length === 6 || 'Must be 6 characters']"
                  @input="roomCode = roomCode.toUpperCase()"
                />
                <v-btn
                  block
                  color="secondary"
                  size="large"
                  :loading="joinLoading"
                  :disabled="!joinNickname.trim() || roomCode.length !== 6"
                  @click="joinRoom"
                >
                  Join Room
                </v-btn>
                <v-alert v-if="joinError" type="error" class="mt-3">{{ joinError }}</v-alert>
              </v-window-item>
            </v-window>

            <v-btn variant="text" block class="mt-4" @click="router.push('/')">← Back</v-btn>
          </v-card-text>
        </v-card>
      </v-col>
    </v-row>
  </v-container>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { api } from '@/services/api'
import { useGameStore } from '@/stores/game'
import { usePlayerStore } from '@/stores/player'

const router = useRouter()
const route = useRoute()
const gameStore = useGameStore()
const playerStore = usePlayerStore()

const tab = ref('create')

// Pre-fill nickname from query param if coming from HomeView.
const prefill = (route.query.nickname as string) ?? ''
const createNickname = ref(prefill)
const joinNickname = ref(prefill)
const roomCode = ref('')

const createLoading = ref(false)
const createError = ref('')
const joinLoading = ref(false)
const joinError = ref('')

async function createRoom() {
  if (!createNickname.value.trim()) return
  createLoading.value = true
  createError.value = ''
  try {
    await gameStore.createGame('multiplayer', createNickname.value.trim())
    playerStore.setPlayer(createNickname.value.trim(), 'white')
    router.push({ name: 'game', params: { id: gameStore.state!.id }, query: { color: 'white' } })
  } catch (e: any) {
    createError.value = e?.response?.data?.error ?? 'Failed to create room'
  } finally {
    createLoading.value = false
  }
}

async function joinRoom() {
  if (!joinNickname.value.trim() || roomCode.value.length !== 6) return
  joinLoading.value = true
  joinError.value = ''
  try {
    const gs = await api.joinGame({ room_code: roomCode.value, nickname: joinNickname.value.trim() })
    gameStore.setState(gs)
    playerStore.setPlayer(joinNickname.value.trim(), 'black')
    router.push({ name: 'game', params: { id: gs.id }, query: { color: 'black' } })
  } catch (e: any) {
    joinError.value = e?.response?.data?.error ?? 'Failed to join room'
  } finally {
    joinLoading.value = false
  }
}
</script>
