package game

import (
	"math"
	"math/rand"
)

// AILevel represents the AI difficulty.
type AILevel string

const (
	Easy   AILevel = "easy"
	Medium AILevel = "medium"
	Hard   AILevel = "hard"
)

// Move bundles a pawn starting position with its chosen path.
type Move struct {
	From Position
	Path []Position
}

// AIMove returns the AI's chosen move given the board, the AI's player color, and difficulty.
func AIMove(b Board, player Color, level AILevel) (Move, bool) {
	switch level {
	case Medium:
		return aiMedium(b, player)
	case Hard:
		return aiHard(b, player)
	default:
		return aiEasy(b, player)
	}
}

// aiEasy picks a random pawn with valid moves and a random valid path.
func aiEasy(b Board, player Color) (Move, bool) {
	pawns := Pawns(b, player)
	rand.Shuffle(len(pawns), func(i, j int) { pawns[i], pawns[j] = pawns[j], pawns[i] })
	for _, pawn := range pawns {
		paths := ValidMoves(b, pawn, player)
		if len(paths) > 0 {
			path := paths[rand.Intn(len(paths))]
			return Move{From: pawn, Path: path}, true
		}
	}
	return Move{}, false
}

// aiMedium picks the move that maximises the sum of progress scores of all own pawns
// after the move is applied.
func aiMedium(b Board, player Color) (Move, bool) {
	pawns := Pawns(b, player)
	bestScore := math.MinInt32
	var bestMove Move
	found := false

	for _, pawn := range pawns {
		paths := ValidMoves(b, pawn, player)
		for _, path := range paths {
			nb := ApplyMove(b, pawn, path)
			score := progressScore(nb, player)
			if score > bestScore {
				bestScore = score
				bestMove = Move{From: pawn, Path: path}
				found = true
			}
		}
	}
	return bestMove, found
}

// progressScore computes the sum of progress of all `player` pawns toward their goal.
// Progress is measured in steps already completed toward the target corner.
func progressScore(b Board, player Color) int {
	score := 0
	for _, p := range Pawns(b, player) {
		score += pawnProgress(p, player)
	}
	return score
}

// pawnProgress returns how many steps a pawn has advanced toward its goal zone.
//
//   White goal: cols 0–2, rows 5–7. Progress = (7 - col) + row
//   Black goal: cols 5–7, rows 0–2. Progress = col + (7 - row)
func pawnProgress(p Position, player Color) int {
	if player == White {
		return (7 - p.Col) + p.Row
	}
	return p.Col + (7 - p.Row)
}

// aiHard uses minimax with alpha-beta pruning (depth 4).
func aiHard(b Board, player Color) (Move, bool) {
	pawns := Pawns(b, player)
	bestScore := math.MinInt32
	var bestMove Move
	found := false

	for _, pawn := range pawns {
		paths := ValidMoves(b, pawn, player)
		for _, path := range paths {
			nb := ApplyMove(b, pawn, path)
			score := minimax(nb, 3, math.MinInt32, math.MaxInt32, false, player)
			if score > bestScore {
				bestScore = score
				bestMove = Move{From: pawn, Path: path}
				found = true
			}
		}
	}
	return bestMove, found
}

// minimax returns the evaluation score for the given board state.
// maximising=true means it's the AI's turn (maximising player).
func minimax(b Board, depth, alpha, beta int, maximising bool, aiPlayer Color) int {
	if winner, ok := CheckWin(b); ok {
		if winner == aiPlayer {
			return 1000000 + depth // prefer faster wins
		}
		return -1000000 - depth
	}
	if depth == 0 {
		return progressScore(b, aiPlayer) - progressScore(b, Opponent(aiPlayer))
	}

	curPlayer := aiPlayer
	if !maximising {
		curPlayer = Opponent(aiPlayer)
	}

	pawns := Pawns(b, curPlayer)

	if maximising {
		best := math.MinInt32
		for _, pawn := range pawns {
			for _, path := range ValidMoves(b, pawn, curPlayer) {
				nb := ApplyMove(b, pawn, path)
				val := minimax(nb, depth-1, alpha, beta, false, aiPlayer)
				if val > best {
					best = val
				}
				if val > alpha {
					alpha = val
				}
				if beta <= alpha {
					return best
				}
			}
		}
		return best
	}

	best := math.MaxInt32
	for _, pawn := range pawns {
		for _, path := range ValidMoves(b, pawn, curPlayer) {
			nb := ApplyMove(b, pawn, path)
			val := minimax(nb, depth-1, alpha, beta, true, aiPlayer)
			if val < best {
				best = val
			}
			if val < beta {
				beta = val
			}
			if beta <= alpha {
				return best
			}
		}
	}
	return best
}
