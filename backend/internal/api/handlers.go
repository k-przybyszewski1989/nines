package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/nines/backend/internal/db"
	"github.com/nines/backend/internal/game"
	"github.com/nines/backend/internal/ws"
)

type Handler struct {
	DB        *sqlx.DB
	WSManager *ws.Manager
}

// POST /api/games
// Body: { "mode": "singleplayer"|"multiplayer", "nickname": "Alice", "ai_level": "medium" }
func (h *Handler) CreateGame(c *gin.Context) {
	var req struct {
		Mode     string `json:"mode" binding:"required"`
		Nickname string `json:"nickname" binding:"required"`
		AILevel  string `json:"ai_level"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Mode != "singleplayer" && req.Mode != "multiplayer" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "mode must be singleplayer or multiplayer"})
		return
	}
	if req.Mode == "singleplayer" {
		switch req.AILevel {
		case "easy", "medium", "hard":
		case "":
			req.AILevel = "easy"
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "ai_level must be easy, medium, or hard"})
			return
		}
	}

	id := uuid.New().String()
	board := game.NewBoard()
	aiLevel := sql.NullString{String: req.AILevel, Valid: req.AILevel != ""}

	var roomCode sql.NullString
	if req.Mode == "multiplayer" {
		roomCode = sql.NullString{String: generateRoomCode(), Valid: true}
	}

	if err := db.CreateGame(h.DB, id, req.Mode, req.Nickname, aiLevel, roomCode, board); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("create game: %v", err)})
		return
	}

	gs, err := db.GetGame(h.DB, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gs)
}

// GET /api/games/:id
func (h *Handler) GetGame(c *gin.Context) {
	id := c.Param("id")
	gs, err := db.GetGame(h.DB, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
		return
	}
	c.JSON(http.StatusOK, gs)
}

// POST /api/games/join
// Body: { "room_code": "ABC123", "nickname": "Bob" }
func (h *Handler) JoinGame(c *gin.Context) {
	var req struct {
		RoomCode string `json:"room_code" binding:"required"`
		Nickname string `json:"nickname" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	gs, err := db.GetGameByRoomCode(h.DB, req.RoomCode)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
		return
	}
	if gs.Status != "waiting" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "game is not waiting for players"})
		return
	}
	if req.Nickname == gs.WhiteNick {
		c.JSON(http.StatusBadRequest, gin.H{"error": "nickname already taken"})
		return
	}

	if err := db.JoinGame(h.DB, gs.ID, req.Nickname); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("join game: %v", err)})
		return
	}

	gs.BlackNick = req.Nickname
	gs.Status = "in_progress"

	// Notify white player if already connected via WebSocket.
	if h.WSManager != nil {
		if hub := h.WSManager.Get(gs.ID); hub != nil {
			type payload struct {
				BlackNick string `json:"black_nick"`
			}
			type envelope struct {
				Type    string  `json:"type"`
				Payload payload `json:"payload"`
			}
			if msg, merr := json.Marshal(envelope{
				Type:    "player_joined",
				Payload: payload{BlackNick: gs.BlackNick},
			}); merr == nil {
				hub.Broadcast(msg)
			}
		}
	}

	c.JSON(http.StatusOK, gs)
}

// POST /api/games/:id/move
// Body: { "from": "H3", "path": ["H4","H5"] }
func (h *Handler) MakeMove(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		From string   `json:"from" binding:"required"`
		Path []string `json:"path" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	gs, err := db.GetGame(h.DB, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
		return
	}

	if gs.Mode == "multiplayer" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "use WebSocket for multiplayer"})
		return
	}

	if gs.Status == "finished" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "game is already finished"})
		return
	}
	if gs.Turn != "white" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "it's not your turn"})
		return
	}

	fromPos, err := game.ParsePosition(req.From)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid from: %v", err)})
		return
	}

	path := make([]game.Position, len(req.Path))
	for i, s := range req.Path {
		p, perr := game.ParsePosition(s)
		if perr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid path[%d]: %v", i, perr)})
			return
		}
		path[i] = p
	}

	if !game.IsValidMove(gs.Board, fromPos, path, game.White) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid move"})
		return
	}

	// Apply human move.
	gs.Board = game.ApplyMove(gs.Board, fromPos, path)
	gs.MoveNum++
	gs.WhiteScore += game.CalculateMoveScore(req.From, req.Path)

	// Record human move.
	_ = db.RecordMove(h.DB, id, "white", gs.MoveNum, req.From, req.Path)

	// Check if human won.
	if winner, ok := game.CheckWin(gs.Board); ok {
		_ = db.UpdateGame(h.DB, id, gs.Board, "black", winner.String(), "finished", gs.MoveNum, gs.WhiteScore, gs.BlackScore)
		gs.Turn = "black"
		gs.Winner = winner.String()
		gs.Status = "finished"
		c.JSON(http.StatusOK, gs)
		return
	}

	// AI turn (singleplayer only).
	aiLevel := game.AILevel(gs.AILevel)
	aiMove, ok := game.AIMove(gs.Board, game.Black, aiLevel)
	if ok {
		aiPath := make([]string, len(aiMove.Path))
		for i, p := range aiMove.Path {
			aiPath[i] = p.String()
		}
		gs.Board = game.ApplyMove(gs.Board, aiMove.From, aiMove.Path)
		gs.MoveNum++
		gs.BlackScore += game.CalculateMoveScore(aiMove.From.String(), aiPath)
		_ = db.RecordMove(h.DB, id, "black", gs.MoveNum, aiMove.From.String(), aiPath)

		if winner, ok2 := game.CheckWin(gs.Board); ok2 {
			_ = db.UpdateGame(h.DB, id, gs.Board, "white", winner.String(), "finished", gs.MoveNum, gs.WhiteScore, gs.BlackScore)
			gs.Turn = "white"
			gs.Winner = winner.String()
			gs.Status = "finished"
			c.JSON(http.StatusOK, gs)
			return
		}
	}
	// After AI move, it's White's turn again.
	gs.Turn = "white"

	_ = db.UpdateGame(h.DB, id, gs.Board, gs.Turn, "", gs.Status, gs.MoveNum, gs.WhiteScore, gs.BlackScore)
	c.JSON(http.StatusOK, gs)
}

const roomCodeChars = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func generateRoomCode() string {
	b := make([]byte, 6)
	for i := range b {
		b[i] = roomCodeChars[rand.Intn(len(roomCodeChars))]
	}
	return string(b)
}
