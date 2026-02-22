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
	MoveNum    int            `db:"move_num"`
	WhiteScore int            `db:"white_score"`
	BlackScore int            `db:"black_score"`
	CreatedAt  time.Time      `db:"created_at"`
	UpdatedAt  time.Time      `db:"updated_at"`
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
		Board:      b,
		MoveNum:    r.MoveNum,
		WhiteScore: r.WhiteScore,
		BlackScore: r.BlackScore,
		CreatedAt:  r.CreatedAt,
		UpdatedAt:  r.UpdatedAt,
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

// LastMove holds the most recent move made in a game.
type LastMove struct {
	From   string   `json:"from"`
	Path   []string `json:"path"`
	Player string   `json:"player"`
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
	MoveNum    int        `json:"move_num"`
	WhiteScore int        `json:"white_score"`
	BlackScore int        `json:"black_score"`
	LastMove   *LastMove  `json:"last_move,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// CreateGame inserts a new game row.
func CreateGame(database *sqlx.DB, id, mode, whiteNick string, aiLevel, roomCode sql.NullString, b game.Board) error {
	boardJSON, err := json.Marshal(b)
	if err != nil {
		return fmt.Errorf("marshal board: %w", err)
	}
	status := "in_progress"
	if mode == "multiplayer" {
		status = "waiting"
	}
	_, err = database.Exec(
		`INSERT INTO games (id, mode, status, room_code, white_nick, ai_level, turn, board) VALUES (?,?,?,?,?,?,?,?)`,
		id, mode, status, roomCode, whiteNick, aiLevel, "white", boardJSON,
	)
	return err
}

// GetGameByRoomCode retrieves a game by its room code.
func GetGameByRoomCode(database *sqlx.DB, roomCode string) (*GameState, error) {
	var row GameRow
	err := database.Get(&row, `SELECT * FROM games WHERE room_code = ?`, roomCode)
	if err != nil {
		return nil, fmt.Errorf("get game by room code %s: %w", roomCode, err)
	}
	gs, err := row.ToGameState()
	if err != nil {
		return nil, err
	}
	gs.LastMove, _ = getLastMove(database, gs.ID)
	return gs, nil
}

// JoinGame sets the black player's nickname and sets status to in_progress.
func JoinGame(database *sqlx.DB, id, blackNick string) error {
	_, err := database.Exec(
		`UPDATE games SET black_nick=?, status='in_progress' WHERE id=?`,
		blackNick, id,
	)
	return err
}

// getLastMove fetches the most recent move for a game, returning nil if none exist.
func getLastMove(database *sqlx.DB, gameID string) (*LastMove, error) {
	var row struct {
		FromPos string `db:"from_pos"`
		Path    []byte `db:"path"`
		Player  string `db:"player"`
	}
	err := database.Get(&row, `SELECT from_pos, path, player FROM moves WHERE game_id = ? ORDER BY move_num DESC LIMIT 1`, gameID)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, nil
		}
		return nil, fmt.Errorf("get last move: %w", err)
	}
	var path []string
	if err := json.Unmarshal(row.Path, &path); err != nil {
		return nil, fmt.Errorf("unmarshal last move path: %w", err)
	}
	return &LastMove{From: row.FromPos, Path: path, Player: row.Player}, nil
}

// GetGame retrieves a game by ID.
func GetGame(database *sqlx.DB, id string) (*GameState, error) {
	var row GameRow
	err := database.Get(&row, `SELECT * FROM games WHERE id = ?`, id)
	if err != nil {
		return nil, fmt.Errorf("get game %s: %w", id, err)
	}
	gs, err := row.ToGameState()
	if err != nil {
		return nil, err
	}
	gs.LastMove, _ = getLastMove(database, id)
	return gs, nil
}

// UpdateGame saves the new board, turn, winner, status, move_num, and scores.
func UpdateGame(database *sqlx.DB, id string, b game.Board, turn, winner, status string, moveNum, whiteScore, blackScore int) error {
	boardJSON, err := json.Marshal(b)
	if err != nil {
		return fmt.Errorf("marshal board: %w", err)
	}
	var winnerVal sql.NullString
	if winner != "" {
		winnerVal = sql.NullString{String: winner, Valid: true}
	}
	_, err = database.Exec(
		`UPDATE games SET board=?, turn=?, winner=?, status=?, move_num=?, white_score=?, black_score=? WHERE id=?`,
		boardJSON, turn, winnerVal, status, moveNum, whiteScore, blackScore, id,
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
