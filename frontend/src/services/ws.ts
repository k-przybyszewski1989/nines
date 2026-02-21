type WsMessageType = 'game_state' | 'player_joined' | 'move_made' | 'game_over' | 'error'
type WsMessage = { type: WsMessageType; payload: unknown }
type MessageHandler = (msg: WsMessage) => void

class WsService {
  private ws: WebSocket | null = null
  private handlers: MessageHandler[] = []

  connect(gameId: string, nickname: string): void {
    this.disconnect()
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const host = window.location.host
    const url = `${protocol}//${host}/ws/${gameId}?nickname=${encodeURIComponent(nickname)}`
    this.ws = new WebSocket(url)

    this.ws.onmessage = (event) => {
      try {
        const msg: WsMessage = JSON.parse(event.data)
        this.handlers.forEach(h => h(msg))
      } catch {
        // ignore malformed frames
      }
    }

    this.ws.onerror = (err) => {
      console.error('WebSocket error:', err)
    }
  }

  disconnect(): void {
    if (this.ws) {
      this.ws.onmessage = null
      this.ws.onerror = null
      this.ws.close()
      this.ws = null
    }
  }

  onMessage(handler: MessageHandler): () => void {
    this.handlers.push(handler)
    return () => {
      this.handlers = this.handlers.filter(h => h !== handler)
    }
  }

  sendMove(from: string, path: string[]): void {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) return
    this.ws.send(JSON.stringify({
      type: 'make_move',
      payload: { from, path },
    }))
  }

  get isConnected(): boolean {
    return this.ws !== null && this.ws.readyState === WebSocket.OPEN
  }
}

export const wsService = new WsService()
