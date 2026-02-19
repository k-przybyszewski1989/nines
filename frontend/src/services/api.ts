import axios from 'axios'

const client = axios.create({
  baseURL: '/api',
  headers: { 'Content-Type': 'application/json' },
})

export interface GameState {
  id: string
  mode: string
  status: string
  room_code?: string
  white_nick: string
  black_nick?: string
  ai_level?: string
  turn: string
  winner?: string
  board: string[][]
  move_num: number
  created_at: string
  updated_at: string
}

export interface CreateGameRequest {
  mode: 'singleplayer' | 'multiplayer'
  nickname: string
  ai_level?: 'easy' | 'medium' | 'hard'
}

export interface MakeMoveRequest {
  from: string
  path: string[]
}

export const api = {
  createGame(req: CreateGameRequest): Promise<GameState> {
    return client.post<GameState>('/games', req).then(r => r.data)
  },
  getGame(id: string): Promise<GameState> {
    return client.get<GameState>(`/games/${id}`).then(r => r.data)
  },
  makeMove(id: string, req: MakeMoveRequest): Promise<GameState> {
    return client.post<GameState>(`/games/${id}/move`, req).then(r => r.data)
  },
}
