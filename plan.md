# Plan: Nines — Chess-Board Game App

## Context
Build a web-based two-player board game called "Nines" (9 pawns per side) played on an 8×8 chess board. Players race to move all 9 of their pawns to the opponent's starting area. The app supports real-time multiplayer (WebSocket) and singleplayer vs. AI (3 difficulty levels). Stack: Vue.js + Vuetify (frontend), Go (backend), MariaDB (database), Docker, GitHub Actions (CI/CD).

---

## Game Mechanics Summary

### Board & Coordinates
- 8×8 board, columns A–H, rows 1–8
- Internal representation: 0-indexed (col: 0=A…7=H, row: 0=1…7=8)

### Starting Positions
- **Whites** (H1:F3): F1,G1,H1, F2,G2,H2, F3,G3,H3 — bottom-right corner
- **Blacks** (A8:C6): A6,B6,C6, A7,B7,C7, A8,B8,C8 — top-left corner

### Movement Directions (relative per player)
| Player | Forward         | Left             |
|--------|----------------|-----------------|
| White  | row +1 (1→8)   | col -1 (H→A)    |
| Black  | row -1 (8→1)   | col +1 (A→H)    |

Valid single moves: forward (+0,+1 / +0,-1) or left (-1,0 / +1,0) — no diagonals.

### Hopping Rules
- A pawn may hop over any adjacent pawn (own or opponent) if:
  - The hopped pawn is directly adjacent in a forward or left direction
  - The landing square (2 steps in that direction) is empty
- Multiple hops in a single turn are allowed (chain); each hop must be forward or left
- No capturing — all 18 pawns always on the board

### Win Condition
First player to have all 9 of their pawns occupying the opponent's starting area wins.

---

## Architecture

### Project Structure
```
nines/
├── backend/
│   ├── cmd/server/main.go
│   ├── internal/
│   │   ├── game/       # Board, move validation, AI
│   │   ├── api/        # HTTP handlers (REST)
│   │   ├── ws/         # WebSocket hub + handler
│   │   └── db/         # MariaDB models + queries
│   ├── go.mod
│   └── Dockerfile
├── frontend/
│   ├── src/
│   │   ├── views/
│   │   │   ├── HomeView.vue      # Nickname + mode selection
│   │   │   ├── LobbyView.vue     # Create/join room (multiplayer)
│   │   │   └── GameView.vue      # Game board + UI
│   │   ├── components/
│   │   │   ├── GameBoard.vue     # 8×8 grid renderer
│   │   │   ├── BoardSquare.vue   # Single square (highlight, pawn)
│   │   │   ├── PlayerInfo.vue    # Name + move counter
│   │   │   └── WinDialog.vue     # End-game modal
│   │   ├── stores/               # Pinia: game state, player
│   │   ├── services/             # REST client, WebSocket client
│   │   └── router/
│   ├── package.json
│   └── Dockerfile
├── docker-compose.yml
├── .github/workflows/ci.yml
└── README.md
```

---

## Backend — Go

### REST API
| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/games` | Create game (mode, nickname, ai_level) |
| GET | `/api/games/:id` | Get full game state |
| POST | `/api/games/join` | Join by room code + nickname |

### WebSocket
- Endpoint: `GET /ws/:gameId?nickname=...`
- Events **server → client**: `game_state`, `move_made`, `game_over`, `player_joined`, `error`
- Events **client → server**: `make_move` `{ pawn: "H1", path: ["H2","H3"] }`

### Database Schema (MariaDB)
```sql
CREATE TABLE games (
  id         CHAR(36) PRIMARY KEY,   -- UUID
  mode       ENUM('singleplayer','multiplayer') NOT NULL,
  status     ENUM('waiting','in_progress','finished') NOT NULL DEFAULT 'waiting',
  room_code  VARCHAR(8) UNIQUE,      -- multiplayer only
  white_nick VARCHAR(50) NOT NULL,
  black_nick VARCHAR(50),            -- filled when second player joins
  ai_level   ENUM('easy','medium','hard'), -- singleplayer only
  turn       ENUM('white','black') NOT NULL DEFAULT 'white',
  winner     ENUM('white','black'),
  board      JSON NOT NULL,          -- serialized board state
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE moves (
  id         BIGINT AUTO_INCREMENT PRIMARY KEY,
  game_id    CHAR(36) NOT NULL REFERENCES games(id),
  player     ENUM('white','black') NOT NULL,
  move_num   INT NOT NULL,
  from_pos   VARCHAR(2) NOT NULL,    -- e.g. "H1"
  path       JSON NOT NULL,          -- e.g. ["H2","H3"]
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Game Engine (`internal/game/`)
- `Board` struct: 8×8 grid with pawn colors
- `ValidMoves(board, pawn, player) [][]Position` — returns all legal move paths (single + multi-hop)
- `ApplyMove(board, path) Board`
- `CheckWin(board) (winner, bool)`
- `AIMove(board, player, level) []Position`:
  - Easy: random valid move for random pawn
  - Medium: greedy — pick move maximising sum of pawn progress toward goal
  - Hard: minimax with alpha-beta pruning (depth 4)

### WebSocket Hub (`internal/ws/`)
- One Hub per game (map[gameID]*Hub)
- Hub holds 2 client connections max (multiplayer) or 1 (singleplayer)
- On `make_move`: validate → apply → broadcast `move_made` + new state → if singleplayer trigger AI turn

---

## Frontend — Vue.js + Vuetify

### Views
1. **HomeView** — Nickname input (v-text-field) + mode buttons (Singleplayer / Multiplayer)
2. **LobbyView** (multiplayer only) — "Create Room" → display room code; "Join Room" → enter code input
3. **GameView** — main game screen

### GameView layout
```
┌─────────────────────────────────────────┐
│ [PlayerInfo: Black] [Turn indicator]    │
│                                         │
│         8×8 GameBoard                   │
│         (centered, square cells)        │
│                                         │
│ [PlayerInfo: White] [Move counter]      │
└─────────────────────────────────────────┘
```

### Interaction flow
1. Player clicks a pawn → highlight it, fetch/compute valid destinations
2. Valid single-step squares: shown with green overlay
3. Valid hop landing squares: shown with blue overlay
4. Player clicks destination → if multi-hop possible from there, show next hops
5. When no further hops, clicking destination commits the move via WebSocket
6. On opponent's move: board animates to new state

### Pinia Stores
- `usePlayerStore` — nickname, color assigned, move count
- `useGameStore` — board state, turn, game status, room code

### WebSocket Service
- Connects on game start
- Dispatches incoming events to Pinia store actions
- Sends `make_move` events

---

## Docker

### docker-compose.yml services
- `db` — MariaDB (volume mounted)
- `backend` — Go binary, depends on db, exposes :8080
- `frontend` — Nginx serving built Vue app, proxies /api and /ws to backend

---

## CI/CD — GitHub Actions (`.github/workflows/ci.yml`)

### Triggers: push to `main` and PRs

### Jobs
1. **test-backend** — `go test ./...` inside backend/
2. **test-frontend** — `npm run test:unit` inside frontend/
3. **build-docker** — build both Docker images (on main branch only)
4. **lint** — golangci-lint (backend), eslint (frontend)

---

---

## Phase 1 — Singleplayer ✅ COMPLETE

### Goal
A human player can open the app, enter a nickname, select Singleplayer, pick a difficulty, and play a full game against the AI.

### Scope
- No room codes, no WebSocket required — turn loop is request/response via REST
- AI turn is triggered automatically by the backend after each human move

### Backend endpoints (Singleplayer)
| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/games` | Create game `{ mode:"singleplayer", nickname, ai_level }` → returns game state |
| GET | `/api/games/:id` | Get current game state |
| POST | `/api/games/:id/move` | Submit human move `{ from, path }` → backend applies move, runs AI, returns updated state |

### Frontend flow (Singleplayer)
1. **HomeView** — Nickname + "Singleplayer" → difficulty picker (Easy / Medium / Hard) → POST `/api/games`
2. **GameView** — one-shot fetch after each move; human is always White; AI is Black
3. After human submits a move, backend responds with AI move already applied (full new state)
4. WinDialog shown when `winner` field is set

### AI Implementation (`internal/game/ai.go`)
- **Easy**: pick a random pawn with valid moves; pick a random valid path
- **Medium**: for each pawn + path, score = total forward+left progress of all own pawns; pick max
- **Hard**: minimax with alpha-beta pruning, depth 4; evaluation = sum of (progress score for own pawns) − sum (for opponent)

### Implementation Status
| File | Status |
|------|--------|
| `backend/internal/game/board.go` | ✅ Done |
| `backend/internal/game/moves.go` | ✅ Done |
| `backend/internal/game/ai.go` | ✅ Done |
| `backend/internal/game/game_test.go` | ✅ Done (15/15 tests pass) |
| `backend/internal/db/db.go` | ✅ Done |
| `backend/internal/db/queries.go` | ✅ Done |
| `backend/internal/api/handlers.go` | ✅ Done |
| `backend/cmd/server/main.go` | ✅ Done |
| `backend/Dockerfile` | ✅ Done |
| `frontend/src/views/HomeView.vue` | ✅ Done |
| `frontend/src/views/GameView.vue` | ✅ Done |
| `frontend/src/components/GameBoard.vue` | ✅ Done |
| `frontend/src/components/BoardSquare.vue` | ✅ Done |
| `frontend/src/components/PlayerInfo.vue` | ✅ Done |
| `frontend/src/components/WinDialog.vue` | ✅ Done |
| `frontend/src/stores/game.ts` | ✅ Done |
| `frontend/src/stores/player.ts` | ✅ Done |
| `frontend/src/services/api.ts` | ✅ Done |
| `frontend/Dockerfile` | ✅ Done |
| `docker-compose.yml` | ✅ Done |
| `.github/workflows/ci.yml` | ✅ Done |

### Singleplayer Verification
- Human move → AI responds in same HTTP response, board updates
- Win condition triggers WinDialog
- All 3 difficulty levels produce legal moves

---

## Phase 2 — Multiplayer

### Goal
Two human players can play against each other in real-time via WebSocket, with one creating a room and sharing the code with the other.

### Scope
- Requires WebSocket hub
- Room code system for pairing players
- Both players see moves instantly; no AI involved

### Backend additions (Multiplayer)
| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/games` | Create game `{ mode:"multiplayer", nickname }` → returns game + `room_code` |
| POST | `/api/games/join` | Join game `{ room_code, nickname }` → returns game state |
| GET | `/ws/:gameId?nickname=...` | Upgrade to WebSocket |

### WebSocket Events
| Direction | Event | Payload |
|-----------|-------|---------|
| server → client | `game_state` | Full board state (sent on connect) |
| server → client | `player_joined` | Black player nickname |
| server → client | `move_made` | Path + new board state |
| server → client | `game_over` | Winner |
| server → client | `error` | Validation error message |
| client → server | `make_move` | `{ from, path }` |

### WebSocket Hub (`internal/ws/`)
- One Hub per game stored in an in-memory map
- Accepts max 2 connections; rejects others
- On `make_move`: validate → apply → persist → broadcast new state to both clients
- On disconnect: broadcast error, mark game as abandoned

### Frontend additions (Multiplayer)
1. **LobbyView** — two options:
   - "Create Room" → POST `/api/games` → display room code in large text + "Waiting for opponent…" spinner
   - "Join Room" → code input field → POST `/api/games/join` → navigate to GameView
2. **GameView** — connects WebSocket on mount; dispatches events to Pinia game store
3. Players can be either White or Black (White = creator, Black = joiner)

### Multiplayer Verification
- Two browser tabs: one creates, one joins with code; both reach GameView
- Move made in tab A appears in tab B in real-time
- Only the player whose turn it is can submit a move (others get error event)
- Win condition triggers WinDialog in both tabs

---

## Shared Implementation Steps

1. **Backend core** — game engine (board, move validation, win check) with unit tests
2. **Database** — schema migrations, db layer (SQLX or GORM)
3. **Docker** — docker-compose with `db`, `backend`, `frontend` (Nginx proxy)
4. **CI/CD** — GitHub Actions: lint + test backend, lint + test frontend, build Docker images

---

## Verification (Full Stack)

- Unit tests cover all valid/invalid move scenarios (forward, left, hop, multi-hop, boundary, blocked)
- `docker-compose up` starts entire stack cleanly
- Singleplayer game completable end-to-end against all 3 AI levels
- Multiplayer game completable end-to-end between two browser sessions
- CI passes on a clean branch
