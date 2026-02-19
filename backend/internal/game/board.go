package game

import (
	"encoding/json"
	"fmt"
)

// Color represents a cell state on the board.
type Color int

const (
	Empty Color = iota
	White
	Black
)

func (c Color) String() string {
	switch c {
	case White:
		return "white"
	case Black:
		return "black"
	default:
		return "empty"
	}
}

func ColorFromString(s string) (Color, error) {
	switch s {
	case "white":
		return White, nil
	case "black":
		return Black, nil
	case "empty":
		return Empty, nil
	default:
		return Empty, fmt.Errorf("unknown color: %s", s)
	}
}

// Position is a 0-indexed board coordinate (col: 0=A…7=H, row: 0=1…7=8).
type Position struct {
	Col int
	Row int
}

// String returns the algebraic notation e.g. "H1".
func (p Position) String() string {
	return fmt.Sprintf("%c%d", rune('A'+p.Col), p.Row+1)
}

// ParsePosition parses "H1" → Position{7, 0}.
func ParsePosition(s string) (Position, error) {
	if len(s) != 2 {
		return Position{}, fmt.Errorf("invalid position %q: must be 2 chars", s)
	}
	col := int(s[0] - 'A')
	row := int(s[1] - '1')
	if col < 0 || col > 7 || row < 0 || row > 7 {
		return Position{}, fmt.Errorf("position %q out of range", s)
	}
	return Position{Col: col, Row: row}, nil
}

// Board is an 8×8 grid. board[row][col] holds the occupant.
type Board [8][8]Color

// NewBoard returns the initial board with pawns in starting positions.
//
// Whites occupy F1:H3 → cols 5–7, rows 0–2.
// Blacks occupy A6:C8 → cols 0–2, rows 5–7.
func NewBoard() Board {
	var b Board
	for col := 5; col <= 7; col++ {
		for row := 0; row <= 2; row++ {
			b[row][col] = White
		}
	}
	for col := 0; col <= 2; col++ {
		for row := 5; row <= 7; row++ {
			b[row][col] = Black
		}
	}
	return b
}

// At returns the occupant at the given position.
func (b *Board) At(p Position) Color {
	return b[p.Row][p.Col]
}

// Set places a color at the given position.
func (b *Board) Set(p Position, c Color) {
	b[p.Row][p.Col] = c
}

// MarshalJSON serialises the board as a 2D array of strings.
func (b Board) MarshalJSON() ([]byte, error) {
	grid := [8][8]string{}
	for r := 0; r < 8; r++ {
		for c := 0; c < 8; c++ {
			grid[r][c] = b[r][c].String()
		}
	}
	return json.Marshal(grid)
}

// UnmarshalJSON deserialises the board from a 2D array of strings.
func (b *Board) UnmarshalJSON(data []byte) error {
	var grid [8][8]string
	if err := json.Unmarshal(data, &grid); err != nil {
		return err
	}
	for r := 0; r < 8; r++ {
		for c := 0; c < 8; c++ {
			col, err := ColorFromString(grid[r][c])
			if err != nil {
				return fmt.Errorf("cell [%d][%d]: %w", r, c, err)
			}
			b[r][c] = col
		}
	}
	return nil
}

// Opponent returns the opposite player color.
func Opponent(c Color) Color {
	if c == White {
		return Black
	}
	return White
}

// WhiteHome is the target zone for White pawns (Black's start): cols 0–2, rows 5–7.
var WhiteHome = homePositions(0, 2, 5, 7)

// BlackHome is the target zone for Black pawns (White's start): cols 5–7, rows 0–2.
var BlackHome = homePositions(5, 7, 0, 2)

func homePositions(colMin, colMax, rowMin, rowMax int) []Position {
	var ps []Position
	for col := colMin; col <= colMax; col++ {
		for row := rowMin; row <= rowMax; row++ {
			ps = append(ps, Position{Col: col, Row: row})
		}
	}
	return ps
}
