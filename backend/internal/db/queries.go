package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/nines/backend/internal/game"
)

// GameRow mirrors the games table row.
type GameRow struct {
	ID         string    `gorm:"column:id;primaryKey;type:char(36)"`
	Mode       string    `gorm:"column:mode;type:enum('singleplayer','multiplayer');not null"`
	Status     string    `gorm:"column:status;type:enum('waiting','in_progress','finished');not null;default:'waiting'"`
	RoomCode   *string   `gorm:"column:room_code;size:8;uniqueIndex"`
	WhiteNick  string    `gorm:"column:white_nick;size:50;not null"`
	BlackNick  *string   `gorm:"column:black_nick;size:50"`
	AILevel    *string   `gorm:"column:ai_level;type:enum('easy','medium','hard')"`
	Turn       string    `gorm:"column:turn;type:enum('white','black');not null;default:'white'"`
	Winner     *string   `gorm:"column:winner;type:enum('white','black')"`
	Board      []byte    `gorm:"column:board;type:json;not null"`
	MoveNum    int       `gorm:"column:move_num;not null;default:0"`
	WhiteScore int       `gorm:"column:white_score;not null;default:0"`
	BlackScore int       `gorm:"column:black_score;not null;default:0"`
	CreatedAt  time.Time `gorm:"column:created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at"`
}

func (GameRow) TableName() string { return "games" }

// MoveRow mirrors the moves table row.
type MoveRow struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement"`
	GameID    string    `gorm:"column:game_id;type:char(36);not null;index"`
	Game      GameRow   `gorm:"foreignKey:GameID;references:ID;constraint:OnDelete:CASCADE"`
	Player    string    `gorm:"column:player;type:enum('white','black');not null"`
	MoveNum   int       `gorm:"column:move_num;not null"`
	FromPos   string    `gorm:"column:from_pos;size:2;not null"`
	Path      []byte    `gorm:"column:path;type:json;not null"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (MoveRow) TableName() string { return "moves" }

// ToGameState converts a GameRow to a GameState.
func (r GameRow) ToGameState() (*GameState, error) {
	var b game.Board
	if err := json.Unmarshal(r.Board, &b); err != nil {
		return nil, fmt.Errorf("unmarshal board: %w", err)
	}
	gs := &GameState{
		ID:         r.ID,
		Mode:       r.Mode,
		Status:     r.Status,
		WhiteNick:  r.WhiteNick,
		Turn:       r.Turn,
		Board:      b,
		MoveNum:    r.MoveNum,
		WhiteScore: r.WhiteScore,
		BlackScore: r.BlackScore,
		CreatedAt:  r.CreatedAt,
		UpdatedAt:  r.UpdatedAt,
	}
	if r.RoomCode != nil {
		gs.RoomCode = *r.RoomCode
	}
	if r.BlackNick != nil {
		gs.BlackNick = *r.BlackNick
	}
	if r.AILevel != nil {
		gs.AILevel = *r.AILevel
	}
	if r.Winner != nil {
		gs.Winner = *r.Winner
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
	ID         string     `json:"id"`
	Mode       string     `json:"mode"`
	Status     string     `json:"status"`
	RoomCode   string     `json:"room_code,omitempty"`
	WhiteNick  string     `json:"white_nick"`
	BlackNick  string     `json:"black_nick,omitempty"`
	AILevel    string     `json:"ai_level,omitempty"`
	Turn       string     `json:"turn"`
	Winner     string     `json:"winner,omitempty"`
	Board      game.Board `json:"board"`
	MoveNum    int        `json:"move_num"`
	WhiteScore int        `json:"white_score"`
	BlackScore int        `json:"black_score"`
	LastMove   *LastMove  `json:"last_move,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// CreateGame inserts a new game row.
func CreateGame(database *gorm.DB, id, mode, whiteNick string, aiLevel, roomCode *string, b game.Board) error {
	boardJSON, err := json.Marshal(b)
	if err != nil {
		return fmt.Errorf("marshal board: %w", err)
	}
	status := "in_progress"
	if mode == "multiplayer" {
		status = "waiting"
	}
	row := GameRow{
		ID:        id,
		Mode:      mode,
		Status:    status,
		RoomCode:  roomCode,
		WhiteNick: whiteNick,
		AILevel:   aiLevel,
		Turn:      "white",
		Board:     boardJSON,
	}
	return database.Create(&row).Error
}

// GetGameByRoomCode retrieves a game by its room code.
func GetGameByRoomCode(database *gorm.DB, roomCode string) (*GameState, error) {
	var row GameRow
	if err := database.First(&row, "room_code = ?", roomCode).Error; err != nil {
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
func JoinGame(database *gorm.DB, id, blackNick string) error {
	return database.Model(&GameRow{}).Where("id = ?", id).Updates(map[string]any{
		"black_nick": blackNick,
		"status":     "in_progress",
	}).Error
}

// getLastMove fetches the most recent move for a game, returning nil if none exist.
func getLastMove(database *gorm.DB, gameID string) (*LastMove, error) {
	var row struct {
		FromPos string `gorm:"column:from_pos"`
		Path    []byte `gorm:"column:path"`
		Player  string `gorm:"column:player"`
	}
	err := database.Model(&MoveRow{}).
		Select("from_pos, path, player").
		Where("game_id = ?", gameID).
		Order("move_num DESC").
		First(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
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
func GetGame(database *gorm.DB, id string) (*GameState, error) {
	var row GameRow
	if err := database.First(&row, "id = ?", id).Error; err != nil {
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
func UpdateGame(database *gorm.DB, id string, b game.Board, turn, winner, status string, moveNum, whiteScore, blackScore int) error {
	boardJSON, err := json.Marshal(b)
	if err != nil {
		return fmt.Errorf("marshal board: %w", err)
	}
	updates := map[string]any{
		"board":       boardJSON,
		"turn":        turn,
		"status":      status,
		"move_num":    moveNum,
		"white_score": whiteScore,
		"black_score": blackScore,
		"winner":      nil,
	}
	if winner != "" {
		updates["winner"] = winner
	}
	return database.Model(&GameRow{}).Where("id = ?", id).Updates(updates).Error
}

// RecordMove inserts a moves row.
func RecordMove(database *gorm.DB, gameID, player string, moveNum int, fromPos string, path []string) error {
	pathJSON, err := json.Marshal(path)
	if err != nil {
		return fmt.Errorf("marshal path: %w", err)
	}
	row := MoveRow{
		GameID:  gameID,
		Player:  player,
		MoveNum: moveNum,
		FromPos: fromPos,
		Path:    pathJSON,
	}
	return database.Create(&row).Error
}
