<template>
  <v-card
    variant="outlined"
    rounded="lg"
    class="pa-3"
    :class="isActive ? 'border-primary' : ''"
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
