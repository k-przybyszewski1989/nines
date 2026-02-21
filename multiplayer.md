# Phase 2 — Multiplayer Implementation Plan

## Context
Phase 1 (singleplayer vs AI) is complete. Phase 2 adds real-time two-player multiplayer via WebSocket. Players create a room (get a code), share it, and play in real-time. No AI is involved. The existing REST `MakeMove` handler stays for singleplayer; multiplayer moves go through WebSocket only.

---

## Implementation Order

### Step 1 — Add gorilla/websocket
```bash
cd backend && go get github.com/gorilla/websocket@v1.5.3
```

---

### Step 2 — Update `backend/internal/db/queries.go`

**Change `CreateGame` signature** to accept `roomCode sql.NullString` and set correct initial `status`:
```go
func CreateGame(database *sqlx.DB, id, mode, whiteNick string, aiLevel, roomCode sql.NullString, b game.Board) error {
    status := "in_progress"
    if mode == "multiplayer" {
        status = "waiting"
    }
    // INSERT includes room_code column
}
```

**Add two new functions:**
```go
func GetGameByRoomCode(database *sqlx.DB, roomCode string) (*GameState, error)
func JoinGame(database *sqlx.DB, id, blackNick string) error  // sets black_nick + status='in_progress'
```

---

### Step 3 — Create `backend/internal/ws/hub.go`

```go
type Manager struct { mu sync.RWMutex; hubs map[string]*Hub }
func NewManager() *Manager
func (m *Manager) GetOrCreate(gameID string, db *sqlx.DB) *Hub
func (m *Manager) Get(gameID string) *Hub   // returns nil if not found

type Hub struct { gameID string; mu sync.Mutex; clients map[*Client]bool; db *sqlx.DB }
func (h *Hub) Register(c *Client)
func (h *Hub) Unregister(c *Client)   // closes c.send, removes from map
func (h *Hub) Broadcast(msg []byte)   // non-blocking sends to all c.send channels
func (h *Hub) ClientCount() int
```

Broadcast implementation: lock briefly to copy send channels, then send non-blocking (`select { case ch <- msg: default: }`).

---

### Step 4 — Create `backend/internal/ws/client.go`

```go
type Client struct {
    hub      *Hub
    conn     *websocket.Conn
    send     chan []byte  // buffered 256
    nickname string
    color    string  // "white" or "black"
}
func (c *Client) ReadPump(onMessage func([]byte))  // pong handler, 60s deadline, calls onMessage per frame
func (c *Client) WritePump()                        // ping every 54s, writes from send channel
```

Standard gorilla/websocket read/write pump pattern with ping/pong keepalives.

---

### Step 5 — Create `backend/internal/ws/handler.go`

```go
var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

func ServeWS(manager *Manager, database *sqlx.DB) gin.HandlerFunc
```

**Connection flow in ServeWS:**
1. Get `gameId` param, `nickname` query param
2. `db.GetGame` → validate mode=multiplayer, game not finished
3. Determine color: nickname==WhiteNick → "white"; nickname==BlackNick → "black"; else 403
4. `hub.ClientCount() >= 2` → 400 (full)
5. Upgrade HTTP → WebSocket
6. Create `Client`, `hub.Register(client)`
7. Send `{"type":"game_state","payload":gs}` immediately
8. If `color=="black" && gs.Status=="in_progress"`: broadcast `player_joined` to hub (handles late WS connect by black after REST join)
9. `go client.WritePump()`
10. `client.ReadPump(func(raw) { handleMessage(client, hub, database, raw) })`

**`handleMessage`** dispatches on `msg.type`:
- `"make_move"` → `handleMakeMove`

**`handleMakeMove`:**
1. Parse `{from, path}` from payload
2. Lock `hub.mu`
3. `db.GetGame` → validate status=in_progress, `gs.Turn==client.color`
4. Parse positions, `game.IsValidMove`
5. `game.ApplyMove`, increment MoveNum, `db.RecordMove`
6. `game.CheckWin` → if winner: `db.UpdateGame(status=finished, winner=...)`, broadcast `game_over`, unlock, return
7. Switch turn, `db.UpdateGame`
8. Collect client send channels (while still locked), unlock
9. Broadcast `{"type":"move_made","payload":{"from","path","player","state":gs}}`

Note: hold `hub.mu` during DB ops (game is 2-player, moves are sequential — performance not a concern).

---

### Step 6 — Update `backend/internal/api/handlers.go`

**Update `Handler` struct:**
```go
type Handler struct {
    DB        *sqlx.DB
    WSManager *ws.Manager
}
```

**Update `CreateGame`:**
- Generate room code for multiplayer: `generateRoomCode()` → 6-char alphanumeric (no ambiguous chars)
- Pass `roomCode sql.NullString` to `db.CreateGame`

**Add `JoinGame` handler (POST /api/games/join):**
1. Bind `{room_code, nickname}`
2. `db.GetGameByRoomCode` → 404 if not found
3. Validate `gs.Status == "waiting"` → 400 if not
4. Validate nickname != WhiteNick → 400 (collision)
5. `db.JoinGame(gs.ID, nickname)`
6. If `h.WSManager != nil`: `hub := h.WSManager.Get(gs.ID)` → if hub != nil: broadcast `player_joined`
7. Return updated gs (mutate BlackNick + Status fields before returning)

**Update `MakeMove`:** add early return for multiplayer:
```go
if gs.Mode == "multiplayer" {
    c.JSON(http.StatusBadRequest, gin.H{"error": "use WebSocket for multiplayer"})
    return
}
```

Add `generateRoomCode()` private function (math/rand, charset without 0/O/1/I).

---

### Step 7 — Update `backend/cmd/server/main.go`

```go
wsManager := ws.NewManager()
h := &api.Handler{DB: database, WSManager: wsManager}
// ...
apiGroup.POST("/games/join", h.JoinGame)
// Outside apiGroup (no /api prefix):
r.GET("/ws/:gameId", ws.ServeWS(wsManager, database))
```

Import `github.com/nines/backend/internal/ws`.

---

### Step 8 — Update `frontend/src/services/api.ts`

Add `JoinGameRequest` interface and `joinGame` method:
```typescript
export interface JoinGameRequest { room_code: string; nickname: string }
// in api object:
joinGame(req: JoinGameRequest): Promise<GameState>  // POST /games/join
```

---

### Step 9 — Create `frontend/src/services/ws.ts`

```typescript
type WsMessageType = 'game_state' | 'player_joined' | 'move_made' | 'game_over' | 'error'
type WsMessage = { type: WsMessageType; payload: any }
type MessageHandler = (msg: WsMessage) => void

class WsService {
  private ws: WebSocket | null = null
  private handlers: MessageHandler[] = []

  connect(gameId: string, nickname: string): void
    // ws:// or wss:// based on window.location.protocol
    // URL: `/ws/${gameId}?nickname=${encodeURIComponent(nickname)}`
    // on message: JSON.parse → dispatch to handlers

  disconnect(): void

  onMessage(handler: MessageHandler): () => void  // returns unsubscribe fn

  sendMove(from: string, path: string[]): void
    // sends {"type":"make_move","payload":{from,path}}

  get isConnected(): boolean
}

export const wsService = new WsService()
```

---

### Step 10 — Update `frontend/src/stores/game.ts`

**Update `diffBlackSquares` → `diffColorSquares`** (accepts color param):
```typescript
function diffColorSquares(prev: string[][], next: string[][], color: string): string[]
```
Update `submitMove` call: `diffColorSquares(prevBoard, ..., 'black')`.

**Add WS state and actions:**
```typescript
const wsConnected = ref(false)

function connectWS(gameId: string, nickname: string, myColor: 'white' | 'black'): void {
  wsService.connect(gameId, nickname)
  wsService.onMessage((msg) => {
    switch (msg.type) {
      case 'game_state':   state.value = msg.payload; break
      case 'player_joined':
        if (state.value) state.value = { ...state.value, black_nick: msg.payload.black_nick, status: 'in_progress' }
        break
      case 'move_made':
        const prev = state.value?.board
        state.value = msg.payload.state
        if (prev && msg.payload.player !== myColor)
          lastAISquares.value = diffColorSquares(prev, state.value!.board, msg.payload.player)
        clearSelection()
        break
      case 'game_over':
        if (state.value) state.value = { ...state.value, winner: msg.payload.winner, status: 'finished' }
        break
      case 'error': error.value = msg.payload.message; break
    }
  })
  wsConnected.value = true
}

function submitMoveWS(from: string, path: string[]): void {
  error.value = ''
  wsService.sendMove(from, path)
  clearSelection()
}

function disconnectWS(): void {
  wsService.disconnect()
  wsConnected.value = false
}
```

Expose `wsConnected`, `connectWS`, `submitMoveWS`, `disconnectWS` from the store.

---

### Step 11 — Create `frontend/src/views/LobbyView.vue`

Two-panel card (tabs or toggle):

**Create Room:**
- Shows nickname (passed as query param or re-entered)
- "Create Room" button → `gameStore.createGame('multiplayer', nickname)` → `playerStore.setPlayer(nick, 'white')` → navigate to `/game/:id?color=white`

**Join Room:**
- Nickname input + room code input (6-char, auto uppercase)
- "Join" button → `api.joinGame({room_code, nickname})` → `gameStore.setState(result)` → `playerStore.setPlayer(nick, 'black')` → navigate to `/game/:id?color=black`

Add `gameStore.setState` action (sets `state.value = gs` directly, for the join flow where we already have the state from REST).

---

### Step 12 — Update `frontend/src/views/GameView.vue`

**`onMounted` changes:**
```typescript
const myColor = (route.query.color as string || 'white') as 'white' | 'black'
if (!gameStore.state || gameStore.state.id !== gameId.value) {
  await gameStore.fetchGame(gameId.value)
}
if (!playerStore.color) {
  const nick = gameStore.state?.mode === 'multiplayer'
    ? (myColor === 'white' ? gameStore.state.white_nick : gameStore.state.black_nick ?? '')
    : gameStore.state?.white_nick ?? ''
  playerStore.setPlayer(nick, myColor)
}
if (gameStore.state?.mode === 'multiplayer') {
  gameStore.connectWS(gameId.value, playerStore.nickname, playerStore.color as 'white' | 'black')
}
```

**`onUnmounted`:** `if (gameStore.state?.mode === 'multiplayer') gameStore.disconnectWS()`

**`isMyTurn` computed (update the prop):**
```typescript
const isMyTurn = computed(() => {
  if (gameStore.status !== 'in_progress') return false
  return gameStore.turn === playerStore.color
})
```
Pass this to `GameBoard :is-my-turn="isMyTurn"` (remove the hardcoded `=== 'white'` check).

**`onMoveCommit`:**
```typescript
async function onMoveCommit(from: string, path: string[]) {
  if (gameStore.state?.mode === 'multiplayer') {
    gameStore.submitMoveWS(from, path)
  } else {
    await gameStore.submitMove(gameId.value, from, path)
  }
  playerStore.incrementMoves()
}
```

**Add waiting state display** (before the Black PlayerInfo, inside `v-else-if="gameStore.state"`):
```vue
<template v-if="gameStore.status === 'waiting'">
  <v-card class="text-center pa-8 mb-4">
    <div class="text-h6 mb-3">Waiting for opponent...</div>
    <div class="text-subtitle-2 mb-2">Share this room code:</div>
    <div class="text-h3 font-weight-bold letter-spacing-widest mb-4">
      {{ gameStore.state.room_code }}
    </div>
    <v-progress-circular indeterminate color="primary" />
  </v-card>
</template>
<template v-else>
  <!-- existing Black PlayerInfo + Board + White PlayerInfo -->
</template>
```

**Turn chip text:**
```typescript
const turnLabel = computed(() => {
  if (gameStore.state?.mode === 'multiplayer') {
    return gameStore.turn === playerStore.color ? 'Your turn' : `Opponent's turn`
  }
  return gameStore.turn === 'white' ? 'Your turn' : 'AI thinking…'
})
```

**Black player display name:**
- In multiplayer: `gameStore.state.black_nick || 'Waiting...'`
- In singleplayer: `gameStore.state.black_nick || 'AI'`

**Bottom PlayerInfo name:** use `playerStore.nickname` (unchanged).

**Player move counts for multiplayer:** Since move_num increments for both players, compute separately:
- White moves = ceil(move_num / 2) when white is first
- Black moves = floor(move_num / 2)

These are already computed correctly for singleplayer and work for multiplayer too.

---

### Step 13 — Update `frontend/src/views/HomeView.vue`

Enable the multiplayer button and navigate to `/lobby`:
```vue
<v-btn
  block color="secondary" size="large" variant="outlined"
  :disabled="!nickname.trim()"
  @click="goMultiplayer"
>
  <v-icon start>mdi-account-multiple</v-icon>
  Multiplayer
</v-btn>
```
```typescript
function goMultiplayer() {
  if (!nickname.value.trim()) return
  router.push({ name: 'lobby', query: { nickname: nickname.value.trim() } })
}
```

---

### Step 14 — Update `frontend/src/router/index.ts`

```typescript
{
  path: '/lobby',
  name: 'lobby',
  component: () => import('@/views/LobbyView.vue'),
},
```

---

### Step 15 — Update `frontend/vite.config.ts`

```typescript
proxy: {
  '/api': 'http://backend:8080',
  '/ws': {
    target: 'ws://backend:8080',
    ws: true,
  },
},
```

---

## Critical Files

| File | Role |
|------|------|
| `backend/internal/ws/hub.go` | Core WS infrastructure; all other WS files depend on it |
| `backend/internal/ws/handler.go` | Move processing for multiplayer; mirrors REST MakeMove logic |
| `backend/internal/api/handlers.go` | JoinGame handler + WSManager injection |
| `frontend/src/services/ws.ts` | All frontend WS communication flows through here |
| `frontend/src/views/GameView.vue` | Most complex frontend change: WS lifecycle + multiplayer branching |

---

## Verification

1. **Backend builds:** `cd backend && go build ./...`
2. **Backend tests still pass:** `go test ./... -race -count=1`
3. **Frontend type-checks:** `cd frontend && npx vue-tsc --noEmit`
4. **End-to-end (Docker):**
   - `docker-compose up`
   - Tab A: nickname "Alice" → Multiplayer → Create Room → note room code
   - Tab B: open `localhost:8090` → nickname "Bob" → Multiplayer → Join Room → enter code
   - Both tabs show the board; Alice (white) moves first
   - Move in Tab A appears in Tab B instantly (via WS broadcast)
   - Only the active player can move; attempt by inactive player shows error
   - Game plays to completion; WinDialog appears in both tabs
5. **Singleplayer regression:** Play a full singleplayer game to confirm unaffected
