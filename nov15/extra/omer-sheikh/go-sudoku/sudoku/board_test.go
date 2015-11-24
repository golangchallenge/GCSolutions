package sudoku

import (
	"strings"
	"testing"
)

func TestSolve(t *testing.T) {
	var grid Grid
	s := `4 _ _ _ _ _ _ 5 _
  _ _ 3 _ _ 7 4 2 _
  _ 5 7 8 _ _ _ _ _
  3 4 _ 5 7 _ 2 _ 6
  _ _ 2 4 1 9 3 _ _
  8 _ 5 _ 6 2 _ 7 4
  _ _ _ _ _ 4 8 6 _
  _ 7 4 1 _ _ 9 _ _
  _ 2 _ _ _ _ _ _ 7`
	reader := strings.NewReader(s)
	grid.Write(reader)

	solution := "4 8 6 2 9 1 7 5 3\n" +
		"9 1 3 6 5 7 4 2 8\n" +
		"2 5 7 8 4 3 6 1 9\n" +
		"3 4 1 5 7 8 2 9 6\n" +
		"7 6 2 4 1 9 3 8 5\n" +
		"8 9 5 3 6 2 1 7 4\n" +
		"5 3 9 7 2 4 8 6 1\n" +
		"6 7 4 1 8 5 9 3 2\n" +
		"1 2 8 9 3 6 5 4 7\n"

	grid.Solve()
	if solution != grid.String() {
		t.Fail()
	}
}

// Test a puzzle with no solution
func TestSolveImpossible(t *testing.T) {
	var grid Grid
	s := `1 2 3 4 5 6 7 8 _
  _ _ _ _ _ _ _ _ _
  _ 5 7 8 _ _ _ _ _
  3 4 _ 5 7 _ 2 _ 9
  _ _ 2 4 1 9 3 _ _
  8 _ 5 _ 6 2 _ 7 4
  _ _ _ _ _ 4 8 6 _
  _ 7 4 1 _ _ 9 _ _
  _ 2 _ _ _ _ _ _ 7`
	reader := strings.NewReader(s)
	grid.Write(reader)

	solved := grid.Solve()
	if solved {
		t.Fail()
	}
}

func TestToggle(t *testing.T) {
	var grid Grid
	grid.togglePossibility(10)
	if grid[0][1] != 2 {
		t.Fail()
	}
	grid.togglePossibility(10)
	if grid[0][1] != 0 {
		t.Fail()
	}
}

func TestRankMultipleSolution(t *testing.T) {
	var grid Grid
	s := `1 _ 3 _ _ 6 _ 8 _
  _ 5 _ _ 8 _ 1 2 _
  7 _ 9 1 _ 3 _ 5 6
  _ 3 _ _ 6 7 _ 9 _
  5 _ 7 8 _ _ _ 3 _
  8 _ 1 _ 3 _ 5 _ 7
  _ 4 _ _ 7 8 _ 1 _
  6 _ 8 _ _ 2 _ 4 _
  _ 1 2 _ 4 5 _ 7 8`

	grid.Write(strings.NewReader(s))
	solved, message := grid.SolveAndRank()
	if !solved || message != rankMessage(unranked) {
		t.Fail()
	}
}
