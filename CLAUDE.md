# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Nines** is a two-player board game (8×8 grid, 9 pawns per side) where each player races to move all pawns to the opponent's starting corner. Built as a full-stack web app with singleplayer (vs AI) and planned multiplayer modes.

## Commands

### Backend (Go)
ORM: GORM

```bash
cd backend
go mod download
go build ./cmd/server        # Build server binary
go test ./... -race -count=1 # Run all tests
go test ./internal/game/...  # Run game logic tests only
golangci-lint run            # Lint
```

### Frontend (Vue/TypeScript)
```bash
cd frontend
npm ci
npm run dev         # Dev server with /api proxy to localhost:8080
npm run build       # Production bundle → dist/
npm run test:unit   # Vitest unit tests
npm run lint        # ESLint
npx vue-tsc --noEmit # Type check
```

### Full Stack
```bash
docker-compose up   # Start MariaDB + Backend + Frontend (port 8090)
```

## Architecture

### Backend (`backend/`)
- **Entry:** `cmd/server/main.go` — Gin HTTP server, CORS setup
- **Game engine:** `internal/game/` — board representation, move validation, AI
  - `board.go`: 8×8 array of Color (Empty/White/Black), position string parsing (e.g. "H2")
  - `moves.go`: `ValidMoves()` — generates all legal moves (forward, left, multi-hop over opponent)
  - `ai.go`: Easy (random), Medium (greedy progress score), Hard (minimax depth-4 with alpha-beta)
  - `game_test.go`: 15 unit tests covering all mechanics
- **API:** `internal/api/handlers.go` — three endpoints: `POST /api/games`, `GET /api/games/:id`, `POST /api/games/:id/move`
- **DB:** `internal/db/` — MariaDB via SQLx; `db.go` auto-migrates schema on startup; `queries.go` handles game/move CRUD

### Frontend (`frontend/src/`)
- **Views:** `HomeView.vue` (nickname + mode select) → `GameView.vue` (board + player info + win dialog)
- **Components:** `GameBoard.vue` (8×8 grid, client-side move highlighting), `BoardSquare.vue`, `PlayerInfo.vue`, `WinDialog.vue`
- **Stores (Pinia):** `game.ts` (board state, turn, selected pawn, valid moves), `player.ts` (nickname, color, move count)
- **API client:** `services/api.ts` — Axios + TypeScript interfaces (`GameState`, request/response types)
- **Routing:** `/` → HomeView, `/game/:id` → GameView

### Key Design Decisions
- **Client-side move validation mirror:** `GameBoard.vue` replicates backend move logic to highlight valid destinations without a server round-trip. Keep both in sync when changing move rules.
- **Single HTTP round-trip for AI:** `POST /api/games/:id/move` applies the human move, runs the AI turn, and returns the full updated state — no WebSocket needed for singleplayer.
- **Move paths as JSON arrays:** Paths stored as `["H2", "H3"]` strings in the database `moves` table.

### Environment Variables (Backend)
```
DB_HOST, DB_PORT, DB_USER, DB_PASS, DB_NAME, PORT (default 8080)
```

### Infrastructure
- **Docker Compose:** MariaDB 11, Go backend, Nginx frontend (port 8090)
- **CI:** `.github/workflows/ci.yml` — tests + lint for both backend/frontend, Docker builds on main/master

## Current Status
- **Phase 1 (Singleplayer):** Complete and tested
- **Phase 2 (Multiplayer):** Planned (WebSocket hub, room codes, `LobbyView`) — not yet implemented
