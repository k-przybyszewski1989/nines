<template>
  <v-card
    rounded="lg"
    class="pa-3 player-card"
    :class="isActive ? 'player-card--active' : 'player-card--inactive'"
    :elevation="isActive ? 4 : 0"
  >
    <div class="d-flex align-center gap-3">
      <div
        class="player-dot"
        :class="color"
      />
      <div>
        <div class="text-body-1 font-weight-bold">
          {{ name }}
        </div>
        <div class="text-caption text-medium-emphasis">
          {{ color === 'white' ? 'White' : 'Black' }} · {{ moveCount }} move{{ moveCount !== 1 ? 's' : '' }}
        </div>
      </div>
      <v-spacer />
      <v-chip
        v-if="isActive"
        color="primary"
        size="small"
      >
        Your turn
      </v-chip>
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

<style scoped>
.player-card {
  transition: box-shadow 0.25s ease, background-color 0.25s ease, border-color 0.25s ease;
  border: 2px solid transparent;
}
.player-card--active {
  border-color: rgb(var(--v-theme-primary));
  background-color: rgba(var(--v-theme-primary), 0.06);
}
.player-card--inactive {
  border-color: rgba(var(--v-border-color), var(--v-border-opacity));
}
.player-dot {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  flex-shrink: 0;
  margin-right: 12px;
}
.player-dot.white { background: #f5f5f5; box-shadow: 0 1px 3px rgba(0,0,0,0.4); }
.player-dot.black { background: #1a1a1a; box-shadow: 0 1px 3px rgba(0,0,0,0.4); }
</style>
