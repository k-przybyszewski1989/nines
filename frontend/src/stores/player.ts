import { defineStore } from 'pinia'
import { ref } from 'vue'

export const usePlayerStore = defineStore('player', () => {
  const nickname = ref('')
  const color = ref<'white' | 'black' | ''>('')
  const moveCount = ref(0)

  function setPlayer(nick: string, c: 'white' | 'black') {
    nickname.value = nick
    color.value = c
    moveCount.value = 0
  }

  function incrementMoves() {
    moveCount.value++
  }

  function reset() {
    nickname.value = ''
    color.value = ''
    moveCount.value = 0
  }

  return { nickname, color, moveCount, setPlayer, incrementMoves, reset }
})
