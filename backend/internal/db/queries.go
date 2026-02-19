package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/nines/backend/internal/game"
)

// GameRow mirrors the games table row.
type GameRow struct {
	ID        string         `db:"id"`
	Mode      string         `db:"mode"`
	Status    string         `db:"status"`
	RoomCode  sql.NullString `db:"room_code"`
	WhiteNick string         `db:"white_nick"`
	BlackNick sql.NullString `db:"black_nick"`
	AILevel   sql.NullString `db:"ai_level"`
	Turn      string         `db:"turn"`
	Winner    sql.NullString `db:"winner"`
	Board     []byte         `db:"board"`
	MoveNum   int            `db:"move_num"`
	CreatedAt time.Time      `db:"created_at"`
	UpdatedAt time.Time      `db:"updated_at"`
}

// ToGameState converts a GameRow to a GameState.
func (r GameRow) ToGameState() (*GameState, error) {
	var b game.Board
	if err := json.Unmarshal(r.Board, &b); err != nil {
		return nil, fmt.Errorf("unmarshal board: %w", err)
	}
	gs := &GameState{
		ID:        r.ID,
		Mode:      r.Mode,
		Status:    r.Status,
		WhiteNick: r.WhiteNick,
		Turn:      r.Turn,
		Board:     b,
		MoveNum:   r.MoveNum,
		CreatedAt: r.CreatedAt,
		UpdatedAt: r.UpdatedAt,
	}
	if r.RoomCode.Valid {
		gs.RoomCode = r.RoomCode.String
	}
	if r.BlackNick.Valid {
		gs.BlackNick = r.BlackNick.String
	}
	if r.AILevel.Valid {
		gs.AILevel = r.AILevel.String
	}
	if r.Winner.Valid {
		gs.Winner = r.Winner.String
	}
	return gs, nil
}

// GameState is the application-level representation of a game.
type GameState struct {
	ID        string     `json:"id"`
	Mode      string     `json:"mode"`
	Status    string     `json:"status"`
	RoomCode  string     `json:"room_code,omitempty"`
	WhiteNick string     `json:"white_nick"`
	BlackNick string     `json:"black_nick,omitempty"`
	AILevel   string     `json:"ai_level,omitempty"`
	Turn      string     `json:"turn"`
	Winner    string     `json:"winner,omitempty"`
	Board     game.Board `json:"board"`
	MoveNum   int        `json:"move_num"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// CreateGame inserts a new game row and returns its ID.
func CreateGame(database *sqlx.DB, id, mode, whiteNick string, aiLevel sql.NullString, b game.Board) error {
	boardJSON, err := json.Marshal(b)
	if err != nil {
		return fmt.Errorf("marshal board: %w", err)
	}
	_, err = database.Exec(
		`INSERT INTO games (id, mode, status, white_nick, ai_level, turn, board) VALUES (?,?,?,?,?,?,?)`,
		id, mode, "in_progress", whiteNick, aiLevel, "white", boardJSON,
	)
	return err
}

// GetGame retrieves a game by ID.
func GetGame(database *sqlx.DB, id string) (*GameState, error) {
	var row GameRow
	err := database.Get(&row, `SELECT * FROM games WHERE id = ?`, id)
	if err != nil {
		return nil, fmt.Errorf("get game %s: %w", id, err)
	}
	return row.ToGameState()
}

// UpdateGame saves the new board, turn, winner, status, and move_num.
func UpdateGame(database *sqlx.DB, id string, b game.Board, turn, winner, status string, moveNum int) error {
	boardJSON, err := json.Marshal(b)
	if err != nil {
		return fmt.Errorf("marshal board: %w", err)
	}
	var winnerVal sql.NullString
	if winner != "" {
		winnerVal = sql.NullString{String: winner, Valid: true}
	}
	_, err = database.Exec(
		`UPDATE games SET board=?, turn=?, winner=?, status=?, move_num=? WHERE id=?`,
		boardJSON, turn, winnerVal, status, moveNum, id,
	)
	return err
}

// RecordMove inserts a moves row.
func RecordMove(database *sqlx.DB, gameID, player string, moveNum int, fromPos string, path []string) error {
	pathJSON, err := json.Marshal(path)
	if err != nil {
		return fmt.Errorf("marshal path: %w", err)
	}
	_, err = database.Exec(
		`INSERT INTO moves (game_id, player, move_num, from_pos, path) VALUES (?,?,?,?,?)`,
		gameID, player, moveNum, fromPos, pathJSON,
	)
	return err
}
