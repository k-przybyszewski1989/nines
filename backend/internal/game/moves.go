package game

// directions returns the two valid move directions for a player:
// forward and left (relative to the player's perspective).
//
//   White: forward = row+1 (toward row 8), left = col−1 (toward col A)
//   Black: forward = row-1 (toward row 1), left = col+1 (toward col H)
func directions(player Color) [2]Position {
	if player == White {
		return [2]Position{{Col: 0, Row: 1}, {Col: -1, Row: 0}}
	}
	return [2]Position{{Col: 0, Row: -1}, {Col: 1, Row: 0}}
}

func inBounds(p Position) bool {
	return p.Col >= 0 && p.Col <= 7 && p.Row >= 0 && p.Row <= 7
}

func add(a, b Position) Position {
	return Position{Col: a.Col + b.Col, Row: a.Row + b.Row}
}

// ValidMoves returns all legal move paths for a pawn at `from` belonging to `player`.
// Each path is a slice of landing positions (not including the starting position).
// A path of length 1 is a single step; longer paths are multi-hop sequences.
func ValidMoves(b Board, from Position, player Color) [][]Position {
	var result [][]Position
	dirs := directions(player)

	// Single-step moves (no hopping).
	for _, dir := range dirs {
		next := add(from, dir)
		if inBounds(next) && b.At(next) == Empty {
			result = append(result, []Position{next})
		}
	}

	// Multi-hop moves (chained jumps).
	visited := map[Position]bool{from: true}
	var hopFrom func(cur Position, path []Position)
	hopFrom = func(cur Position, path []Position) {
		for _, dir := range dirs {
			mid := add(cur, dir)
			land := add(mid, dir)
			if !inBounds(mid) || !inBounds(land) {
				continue
			}
			if b.At(mid) == Empty {
				continue // nothing to hop over
			}
			if b.At(land) != Empty {
				continue // landing square occupied
			}
			if visited[land] {
				continue // already visited in this chain
			}
			newPath := make([]Position, len(path)+1)
			copy(newPath, path)
			newPath[len(path)] = land

			result = append(result, newPath)
			visited[land] = true
			hopFrom(land, newPath)
			visited[land] = false // backtrack so other chains can use this square
		}
	}
	hopFrom(from, nil)

	return result
}

// IsValidMove checks whether `path` is a legal move for pawn at `from` by `player`.
func IsValidMove(b Board, from Position, path []Position, player Color) bool {
	if b.At(from) != player {
		return false
	}
	valid := ValidMoves(b, from, player)
	for _, vp := range valid {
		if pathsEqual(vp, path) {
			return true
		}
	}
	return false
}

func pathsEqual(a, b []Position) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// ApplyMove applies a move: removes the pawn from `from` and places it at the last position of `path`.
// The board is modified in place and also returned for convenience.
func ApplyMove(b Board, from Position, path []Position) Board {
	color := b.At(from)
	b.Set(from, Empty)
	b.Set(path[len(path)-1], color)
	return b
}

// CheckWin returns (winner, true) if someone has won, otherwise (Empty, false).
//
//   White wins by filling Black's home (cols 0–2, rows 5–7).
//   Black wins by filling White's home (cols 5–7, rows 0–2).
func CheckWin(b Board) (Color, bool) {
	if allColor(b, WhiteHome, White) {
		return White, true
	}
	if allColor(b, BlackHome, Black) {
		return Black, true
	}
	return Empty, false
}

func allColor(b Board, positions []Position, c Color) bool {
	for _, p := range positions {
		if b.At(p) != c {
			return false
		}
	}
	return true
}

// Pawns returns all positions occupied by the given player.
func Pawns(b Board, player Color) []Position {
	var ps []Position
	for r := 0; r < 8; r++ {
		for c := 0; c < 8; c++ {
			if b[r][c] == player {
				ps = append(ps, Position{Col: c, Row: r})
			}
		}
	}
	return ps
}
