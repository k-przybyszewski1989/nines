<template>
  <div
    class="board-square"
    :class="[squareColor, highlightClass, { 'cursor-pointer': isClickable, 'sq-last-ai': isLastAiMove && highlight === 'none' }]"
    @click="$emit('click')"
  >
    <!-- Pawn -->
    <div
      v-if="pawn !== 'empty'"
      class="pawn"
      :class="pawn"
    />

    <!-- Highlight indicator dot -->
    <div
      v-else-if="highlight !== 'none'"
      class="highlight-dot"
      :class="highlight"
    />
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  col: number
  row: number
  pawn: 'white' | 'black' | 'empty'
  highlight: 'none' | 'selected' | 'step' | 'hop'
  isLastAiMove: boolean
}>()

defineEmits<{
  click: []
}>()

// Classic chess board checkerboard coloring.
const squareColor = computed(() =>
  (props.col + props.row) % 2 === 0 ? 'sq-light' : 'sq-dark'
)

const highlightClass = computed(() => {
  if (props.highlight === 'selected') return 'sq-selected'
  if (props.highlight === 'step') return 'sq-step'
  if (props.highlight === 'hop') return 'sq-hop'
  return ''
})

const isClickable = computed(() =>
  props.highlight !== 'none' || props.pawn !== 'empty'
)
</script>

<style scoped>
.board-square {
  width: 100%;
  aspect-ratio: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  transition: background-color 0.15s;
}
.sq-light { background-color: #f0d9b5; }
.sq-dark  { background-color: #b58863; }

.sq-selected  { background-color: #f6f669 !important; }
.sq-step      { background-color: #7fc97f !important; }
.sq-hop       { background-color: #7bb8e8 !important; }
.sq-last-ai   { opacity: 0.45; }

.cursor-pointer { cursor: pointer; }

.pawn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 80%;
  height: 80%;
  border-radius: 50%;
  transition: transform 0.15s;
}
.pawn:hover { transform: scale(1.1); }
.pawn.white { background: #f5f5f5; box-shadow: 0 1px 4px rgba(0,0,0,0.5); }
.pawn.black { background: #1a1a1a; box-shadow: 0 1px 4px rgba(0,0,0,0.5); }

.highlight-dot {
  width: 30%;
  height: 30%;
  border-radius: 50%;
  opacity: 0.7;
}
.highlight-dot.step { background: #2e7d32; }
.highlight-dot.hop  { background: #1565c0; }
</style>
