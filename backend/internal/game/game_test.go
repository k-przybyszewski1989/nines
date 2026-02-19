package game

import (
	"testing"
)

func pos(s string) Position {
	p, err := ParsePosition(s)
	if err != nil {
		panic(err)
	}
	return p
}

// TestNewBoard verifies initial pawn placement.
func TestNewBoard(t *testing.T) {
	b := NewBoard()

	// White pawns: F1-H3
	for _, s := range []string{"F1", "G1", "H1", "F2", "G2", "H2", "F3", "G3", "H3"} {
		if b.At(pos(s)) != White {
			t.Errorf("expected White at %s", s)
		}
	}

	// Black pawns: A6-C8
	for _, s := range []string{"A6", "B6", "C6", "A7", "B7", "C7", "A8", "B8", "C8"} {
		if b.At(pos(s)) != Black {
			t.Errorf("expected Black at %s", s)
		}
	}

	// Some empty squares
	for _, s := range []string{"A1", "D4", "H8"} {
		if b.At(pos(s)) != Empty {
			t.Errorf("expected Empty at %s", s)
		}
	}
}

// TestParsePosition checks round-trip encoding.
func TestParsePosition(t *testing.T) {
	cases := []struct{ s string }{
		{"A1"}, {"H8"}, {"D4"}, {"F3"}, {"A8"},
	}
	for _, tc := range cases {
		p, err := ParsePosition(tc.s)
		if err != nil {
			t.Errorf("ParsePosition(%q): %v", tc.s, err)
			continue
		}
		if got := p.String(); got != tc.s {
			t.Errorf("ParsePosition(%q).String() = %q", tc.s, got)
		}
	}

	// Invalid cases
	for _, bad := range []string{"", "X", "A9", "I1", "a1"} {
		if _, err := ParsePosition(bad); err == nil {
			t.Errorf("ParsePosition(%q) should have failed", bad)
		}
	}
}

// TestWhiteForwardMove verifies a White pawn can move forward.
func TestWhiteForwardMove(t *testing.T) {
	b := NewBoard()
	// H1 White pawn — forward means row+1 → H2 (but H2 already has White!)
	// Try H3 → H4 (no pawn at H4)
	from := pos("H3")
	moves := ValidMoves(b, from, White)
	// H3 forward = H4 (empty ✓), H3 left = G3 (White — blocked)
	// So only one single-step: H4
	singleSteps := 0
	for _, m := range moves {
		if len(m) == 1 {
			singleSteps++
		}
	}
	if singleSteps == 0 {
		t.Errorf("H3 White should have at least one single-step move (H4)")
	}
	// Ensure H4 is among single-step moves
	found := false
	for _, m := range moves {
		if len(m) == 1 && m[0] == pos("H4") {
			found = true
		}
	}
	if !found {
		t.Errorf("H4 should be a valid forward move for White at H3")
	}
}

// TestBlackForwardMove verifies a Black pawn can move forward.
func TestBlackForwardMove(t *testing.T) {
	b := NewBoard()
	// A6 Black — forward means row-1 → A5 (empty ✓), left = col+1 → B6 (Black — blocked)
	from := pos("A6")
	moves := ValidMoves(b, from, Black)
	found := false
	for _, m := range moves {
		if len(m) == 1 && m[0] == pos("A5") {
			found = true
		}
	}
	if !found {
		t.Errorf("A5 should be a valid forward move for Black at A6")
	}
}

// TestNoBackwardMove ensures backward moves are not valid.
func TestNoBackwardMove(t *testing.T) {
	b := NewBoard()
	// Place a lone White pawn at D4
	b.Set(pos("D4"), White)
	// White backward = row-1, right = col+1 — neither should appear
	moves := ValidMoves(b, pos("D4"), White)
	for _, m := range moves {
		if len(m) == 1 {
			dest := m[0]
			if dest == pos("D3") {
				t.Error("D3 is backward for White — should not be valid")
			}
			if dest == pos("E4") {
				t.Error("E4 is right for White — should not be valid")
			}
		}
	}
}

// TestHopBasic tests a simple single hop.
func TestHopBasic(t *testing.T) {
	var b Board
	b.Set(pos("D4"), White)
	b.Set(pos("D5"), Black) // pawn to hop over (forward)
	// D6 must be empty for hop to work

	moves := ValidMoves(b, pos("D4"), White)
	found := false
	for _, m := range moves {
		if len(m) == 1 && m[0] == pos("D6") {
			found = true
		}
	}
	if !found {
		t.Error("White at D4 should be able to hop over Black at D5 to D6")
	}
}

// TestHopBlocked ensures hop is blocked when landing square is occupied.
func TestHopBlocked(t *testing.T) {
	var b Board
	b.Set(pos("D4"), White)
	b.Set(pos("D5"), Black) // pawn to hop over
	b.Set(pos("D6"), White) // landing square occupied

	moves := ValidMoves(b, pos("D4"), White)
	for _, m := range moves {
		if len(m) >= 1 && m[len(m)-1] == pos("D6") {
			t.Error("D6 is occupied — hop should be blocked")
		}
	}
}

// TestMultiHop tests a two-hop chain.
func TestMultiHop(t *testing.T) {
	var b Board
	b.Set(pos("D4"), White)
	b.Set(pos("D5"), Black) // first hop target
	b.Set(pos("D7"), Black) // second hop target
	// D6 and D8 are empty landing squares

	moves := ValidMoves(b, pos("D4"), White)
	// Expect path [D6, D8]
	found := false
	for _, m := range moves {
		if len(m) == 2 && m[0] == pos("D6") && m[1] == pos("D8") {
			found = true
		}
	}
	if !found {
		t.Error("Expected two-hop path [D6, D8] for White at D4")
	}
}

// TestApplyMove verifies pawn is moved correctly.
func TestApplyMove(t *testing.T) {
	b := NewBoard()
	from := pos("H3")
	to := pos("H4")
	b2 := ApplyMove(b, from, []Position{to})
	if b2.At(from) != Empty {
		t.Errorf("expected %s to be empty after move", from)
	}
	if b2.At(to) != White {
		t.Errorf("expected %s to have White after move", to)
	}
}

// TestCheckWin verifies win detection.
func TestCheckWin(t *testing.T) {
	// Set up White winning: fill Black's home (A6:C8) with White
	var b Board
	for _, p := range WhiteHome {
		b.Set(p, White)
	}
	winner, ok := CheckWin(b)
	if !ok || winner != White {
		t.Errorf("expected White win, got %v %v", winner, ok)
	}

	// Set up Black winning: fill White's home (F1:H3) with Black
	var b2 Board
	for _, p := range BlackHome {
		b2.Set(p, Black)
	}
	winner2, ok2 := CheckWin(b2)
	if !ok2 || winner2 != Black {
		t.Errorf("expected Black win, got %v %v", winner2, ok2)
	}

	// No winner on initial board
	b3 := NewBoard()
	_, ok3 := CheckWin(b3)
	if ok3 {
		t.Error("no winner expected on initial board")
	}
}

// TestBoundaryMoves ensures moves off the board are rejected.
func TestBoundaryMoves(t *testing.T) {
	var b Board
	// White at A4 — forward ok (A5), left would be off-board (col -1)
	b.Set(pos("A4"), White)
	moves := ValidMoves(b, pos("A4"), White)
	for _, m := range moves {
		if len(m) == 1 {
			dest := m[0]
			if dest.Col < 0 || dest.Col > 7 || dest.Row < 0 || dest.Row > 7 {
				t.Errorf("out-of-bounds move to %v", dest)
			}
		}
	}
	// Expect only A5 (forward), not off-board left
	found := false
	for _, m := range moves {
		if len(m) == 1 && m[0] == pos("A5") {
			found = true
		}
	}
	if !found {
		t.Error("White at A4 should have A5 as valid forward move")
	}
}

// TestIsValidMove checks move validation.
func TestIsValidMove(t *testing.T) {
	b := NewBoard()
	from := pos("H3")
	path := []Position{pos("H4")}
	if !IsValidMove(b, from, path, White) {
		t.Error("H3→H4 should be valid for White")
	}
	// Wrong player
	if IsValidMove(b, from, path, Black) {
		t.Error("H3 is White's pawn, should fail for Black player")
	}
	// Invalid destination
	if IsValidMove(b, from, []Position{pos("H2")}, White) {
		t.Error("H3→H2 is backward — should be invalid")
	}
}

// TestAIEasyProducesLegalMove ensures the Easy AI always makes a legal move.
func TestAIEasyProducesLegalMove(t *testing.T) {
	b := NewBoard()
	move, ok := AIMove(b, Black, Easy)
	if !ok {
		t.Fatal("Easy AI should find a move on initial board")
	}
	if !IsValidMove(b, move.From, move.Path, Black) {
		t.Errorf("Easy AI made illegal move: %v → %v", move.From, move.Path)
	}
}

// TestAIMediumProducesLegalMove ensures the Medium AI always makes a legal move.
func TestAIMediumProducesLegalMove(t *testing.T) {
	b := NewBoard()
	move, ok := AIMove(b, Black, Medium)
	if !ok {
		t.Fatal("Medium AI should find a move on initial board")
	}
	if !IsValidMove(b, move.From, move.Path, Black) {
		t.Errorf("Medium AI made illegal move: %v → %v", move.From, move.Path)
	}
}

// TestAIHardProducesLegalMove ensures the Hard AI always makes a legal move.
func TestAIHardProducesLegalMove(t *testing.T) {
	b := NewBoard()
	move, ok := AIMove(b, Black, Hard)
	if !ok {
		t.Fatal("Hard AI should find a move on initial board")
	}
	if !IsValidMove(b, move.From, move.Path, Black) {
		t.Errorf("Hard AI made illegal move: %v → %v", move.From, move.Path)
	}
}
