<template>
  <v-dialog v-model="visible" max-width="400" persistent>
    <v-card rounded="xl" elevation="12">
      <v-card-title class="text-h4 text-center pa-6">
        🏆 Game Over
      </v-card-title>
      <v-card-text class="text-center text-h6 pb-4">
        <span :class="winnerColor === 'white' ? 'text-white' : 'text-deep-purple'">
          {{ winnerName }}
        </span>
        wins!
      </v-card-text>
      <v-card-actions class="justify-center pb-6">
        <v-btn color="primary" size="large" @click="playAgain">Play Again</v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'

const props = defineProps<{
  winner: string
  whiteNick: string
  blackNick: string
}>()

const router = useRouter()
const visible = computed(() => !!props.winner)
const winnerColor = computed(() => props.winner)
const winnerName = computed(() => props.winner === 'white' ? props.whiteNick : props.blackNick)

function playAgain() {
  router.push({ name: 'home' })
}
</script>
