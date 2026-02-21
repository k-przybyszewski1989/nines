package ws

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/jmoiron/sqlx"
	"github.com/nines/backend/internal/db"
	"github.com/nines/backend/internal/game"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type wsMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// ServeWS is the Gin handler for GET /ws/:gameId.
func ServeWS(manager *Manager, database *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		gameID := c.Param("gameId")
		nickname := c.Query("nickname")
		if nickname == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "nickname query param required"})
			return
		}

		gs, err := db.GetGame(database, gameID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "game not found"})
			return
		}
		if gs.Mode != "multiplayer" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "not a multiplayer game"})
			return
		}
		if gs.Status == "finished" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "game is finished"})
			return
		}

		// Determine color.
		var color string
		switch {
		case nickname == gs.WhiteNick:
			color = "white"
		case nickname == gs.BlackNick:
			color = "black"
		default:
			c.JSON(http.StatusForbidden, gin.H{"error": "not a participant in this game"})
			return
		}

		hub := manager.GetOrCreate(gameID, database)
		if hub.ClientCount() >= 2 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "game is full"})
			return
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Printf("ws upgrade: %v", err)
			return
		}

		client := &Client{
			hub:      hub,
			conn:     conn,
			send:     make(chan []byte, 256),
			nickname: nickname,
			color:    color,
		}
		hub.Register(client)

		// Send current game state immediately.
		if msg, err := encodeMsg("game_state", gs); err == nil {
			client.send <- msg
		}

		// If black just connected and game is in_progress, broadcast player_joined.
		if color == "black" && gs.Status == "in_progress" {
			if msg, err := encodeMsg("player_joined", map[string]string{"black_nick": gs.BlackNick}); err == nil {
				hub.Broadcast(msg)
			}
		}

		go client.WritePump()
		client.ReadPump(func(raw []byte) {
			handleMessage(client, hub, database, raw)
		})
	}
}

func handleMessage(c *Client, h *Hub, database *sqlx.DB, raw []byte) {
	var msg wsMessage
	if err := json.Unmarshal(raw, &msg); err != nil {
		sendError(c, "invalid JSON")
		return
	}
	switch msg.Type {
	case "make_move":
		handleMakeMove(c, h, database, msg.Payload)
	default:
		sendError(c, fmt.Sprintf("unknown message type: %s", msg.Type))
	}
}

func handleMakeMove(c *Client, h *Hub, database *sqlx.DB, payload json.RawMessage) {
	var req struct {
		From string   `json:"from"`
		Path []string `json:"path"`
	}
	if err := json.Unmarshal(payload, &req); err != nil {
		sendError(c, "invalid make_move payload")
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	gs, err := db.GetGame(database, h.gameID)
	if err != nil {
		sendError(c, "game not found")
		return
	}
	if gs.Status != "in_progress" {
		sendError(c, "game is not in progress")
		return
	}
	if gs.Turn != c.color {
		sendError(c, "not your turn")
		return
	}

	fromPos, err := game.ParsePosition(req.From)
	if err != nil {
		sendError(c, fmt.Sprintf("invalid from: %v", err))
		return
	}

	path := make([]game.Position, len(req.Path))
	for i, s := range req.Path {
		p, perr := game.ParsePosition(s)
		if perr != nil {
			sendError(c, fmt.Sprintf("invalid path[%d]: %v", i, perr))
			return
		}
		path[i] = p
	}

	playerColor := game.White
	if c.color == "black" {
		playerColor = game.Black
	}

	if !game.IsValidMove(gs.Board, fromPos, path, playerColor) {
		sendError(c, "invalid move")
		return
	}

	gs.Board = game.ApplyMove(gs.Board, fromPos, path)
	gs.MoveNum++
	_ = db.RecordMove(database, h.gameID, c.color, gs.MoveNum, req.From, req.Path)

	if winner, ok := game.CheckWin(gs.Board); ok {
		winStr := winner.String()
		_ = db.UpdateGame(database, h.gameID, gs.Board, gs.Turn, winStr, "finished", gs.MoveNum)
		gs.Winner = winStr
		gs.Status = "finished"
		if msg, err := encodeMsg("game_over", map[string]string{"winner": winStr}); err == nil {
			broadcastLocked(h, msg)
		}
		return
	}

	// Switch turn.
	if gs.Turn == "white" {
		gs.Turn = "black"
	} else {
		gs.Turn = "white"
	}
	_ = db.UpdateGame(database, h.gameID, gs.Board, gs.Turn, "", gs.Status, gs.MoveNum)

	payload2, _ := json.Marshal(map[string]any{
		"from":   req.From,
		"path":   req.Path,
		"player": c.color,
		"state":  gs,
	})
	outMsg, _ := json.Marshal(wsMessage{Type: "move_made", Payload: payload2})
	broadcastLocked(h, outMsg)
}

// broadcastLocked sends to all clients while hub.mu is already held.
func broadcastLocked(h *Hub, msg []byte) {
	for client := range h.clients {
		select {
		case client.send <- msg:
		default:
		}
	}
}

func encodeMsg(msgType string, payload any) ([]byte, error) {
	p, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return json.Marshal(wsMessage{Type: msgType, Payload: p})
}

func sendError(c *Client, message string) {
	if msg, err := encodeMsg("error", map[string]string{"message": message}); err == nil {
		select {
		case c.send <- msg:
		default:
		}
	}
}
