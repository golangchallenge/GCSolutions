package backtrack

import (
	"testing"

	"github.com/slimmy/go-challenge-8/puzzle"
	"github.com/stretchr/testify/assert"
)

// Just testing that we can create a solver and call Solve().
// Tests that involves solving boards is implemented in puzzle/solver_test.go
func TestBacktrack(t *testing.T) {
	for _, v := range []string{"easy", "medium", "hard"} {
		board, err := puzzle.Generate(v)
		assert.NotNil(t, board)
		assert.NoError(t, err)

		solvedBoard := NewSolver(board).Solve()
		assert.NotNil(t, solvedBoard)
	}
}
