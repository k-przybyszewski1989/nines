package game

import "math"

var colScoreValues = map[byte]int{
	'A': 1, 'B': 1, 'C': 2, 'D': 3, 'E': 3, 'F': 2, 'G': 1, 'H': 1,
}

var rowScoreValues = map[byte]int{
	'1': 1, '2': 1, '3': 2, '4': 3, '5': 3, '6': 2, '7': 1, '8': 1,
}

func squareBaseValue(pos string) int {
	col := colScoreValues[pos[0]]
	row := rowScoreValues[pos[1]]
	if col > row {
		return col
	}
	return row
}

func scoreMidpoint(a, b string) string {
	return string([]byte{
		byte((int(a[0]) + int(b[0])) / 2),
		byte((int(a[1]) + int(b[1])) / 2),
	})
}

func scoreAbsDiff(a, b byte) int {
	if a > b {
		return int(a - b)
	}
	return int(b - a)
}

func isScoreSimpleStep(from, to string) bool {
	colDiff := scoreAbsDiff(from[0], to[0])
	rowDiff := scoreAbsDiff(from[1], to[1])
	return colDiff <= 1 && rowDiff <= 1 && colDiff+rowDiff > 0
}

func isScoreDiagonalHop(from, to string) bool {
	return scoreAbsDiff(from[0], to[0]) == 2 && scoreAbsDiff(from[1], to[1]) == 2
}

// CalculateMoveScore returns the score for a move.
// from and path elements are algebraic notation strings (e.g. "H2").
func CalculateMoveScore(from string, path []string) int {
	if len(path) == 1 && isScoreSimpleStep(from, path[0]) {
		return 1
	}

	positions := make([]string, 0, 1+len(path))
	positions = append(positions, from)
	positions = append(positions, path...)

	var diagonalSum, diagonalCount, straightSum int
	for i := 0; i < len(positions)-1; i++ {
		mid := scoreMidpoint(positions[i], positions[i+1])
		val := squareBaseValue(mid)
		if isScoreDiagonalHop(positions[i], positions[i+1]) {
			diagonalSum += val
			diagonalCount++
		} else {
			straightSum += val
		}
	}

	return int(math.Pow(float64(diagonalSum), float64(diagonalCount+1))) + straightSum
}
