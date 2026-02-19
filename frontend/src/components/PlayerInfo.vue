<template>
  <v-card variant="outlined" rounded="lg" class="pa-3" :class="isActive ? 'border-primary' : ''">
    <div class="d-flex align-center gap-3">
      <v-icon :color="color === 'white' ? 'white' : 'deep-purple'" size="28">
        mdi-chess-pawn
      </v-icon>
      <div>
        <div class="text-body-1 font-weight-bold">{{ name }}</div>
        <div class="text-caption text-medium-emphasis">
          {{ color === 'white' ? 'White' : 'Black' }} · {{ moveCount }} move{{ moveCount !== 1 ? 's' : '' }}
        </div>
      </div>
      <v-spacer />
      <v-chip v-if="isActive" color="primary" size="small">Your turn</v-chip>
    </div>
  </v-card>
</template>

<script setup lang="ts">
defineProps<{
  name: string
  color: 'white' | 'black'
  moveCount: number
  isActive: boolean
}>()
</script>
