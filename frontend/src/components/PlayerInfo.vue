<template>
  <v-card
    rounded="lg"
    class="py-3 player-card"
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
      <span v-if="isActive" class="led-green" />
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
  padding-left: 23px !important;
  padding-right: 23px !important;
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
  margin-right: 18px;
}
.player-dot.white { background: #f5f5f5; box-shadow: 0 1px 3px rgba(0,0,0,0.4); }
.player-dot.black { background: #1a1a1a; box-shadow: 0 1px 3px rgba(0,0,0,0.4); }
.led-green {
  display: inline-block;
  width: 12px;
  height: 12px;
  border-radius: 50%;
  background: #4caf50;
  box-shadow: 0 0 6px 2px rgba(76, 175, 80, 0.7);
  animation: led-pulse 1.4s ease-in-out infinite;
}
@keyframes led-pulse {
  0%, 100% { box-shadow: 0 0 6px 2px rgba(76, 175, 80, 0.7); }
  50% { box-shadow: 0 0 10px 4px rgba(76, 175, 80, 0.4); }
}
</style>
