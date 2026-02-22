import { defineStore } from 'pinia'
import { ref } from 'vue'

export const usePlayerStore = defineStore('player', () => {
  const nickname = ref('')
  const color = ref<'white' | 'black' | ''>('')
  const totalScore = ref(0)
  const aiScore = ref(0)

  function setPlayer(nick: string, c: 'white' | 'black') {
    nickname.value = nick
    color.value = c
    totalScore.value = 0
    aiScore.value = 0
  }

  function addScore(points: number) {
    totalScore.value += points
  }

  function addAiScore(points: number) {
    aiScore.value += points
  }

  function reset() {
    nickname.value = ''
    color.value = ''
    totalScore.value = 0
    aiScore.value = 0
  }

  return { nickname, color, totalScore, aiScore, setPlayer, addScore, addAiScore, reset }
})
