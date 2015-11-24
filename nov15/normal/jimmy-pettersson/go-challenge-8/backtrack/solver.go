package backtrack

import "github.com/slimmy/go-challenge-8/puzzle"

// Solver implements the SudokuSolver interface
type Solver struct {
	board *puzzle.Board
}

// NewSolver returns a new solver for a given board
func NewSolver(board *puzzle.Board) *Solver {
	return &Solver{board: board}
}

// Solve tries to solve the board using the backtrack algorithm
func (s *Solver) Solve() *puzzle.Board {
	bb, _ := s.backtrack(s.board)
	return bb
}

// Reference implementation of a Sudoku solver
// https://en.wikipedia.org/wiki/Sudoku_solving_algorithms#Backtracking
func (s *Solver) backtrack(b *puzzle.Board) (*puzzle.Board, bool) {
	if b.Solved() {
		return b, true
	}

	for row := 0; row < 9; row++ {
		for col := 0; col < 9; col++ {
			if b.ValueAt(row, col) == 0 {
				for i := 1; i < 10; i++ {
					if b.Allowed(row, col, i) {
						if bb, s := s.backtrack(b.SetAndCopy(row, col, i)); s {
							return bb, s
						}
					}
				}

				// None of the numbers [1,9] can be inserted in this empty cell
				// so the board is unsolvable
				return b, false
			}
		}
	}

	return b, false
}
