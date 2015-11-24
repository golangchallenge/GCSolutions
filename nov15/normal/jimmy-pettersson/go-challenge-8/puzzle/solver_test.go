package puzzle_test

import (
	"log"
	"testing"

	"github.com/slimmy/go-challenge-8/backtrack"
	"github.com/slimmy/go-challenge-8/dlx"
	"github.com/slimmy/go-challenge-8/puzzle"
	"github.com/stretchr/testify/assert"
)

var (
	testInput1 = `1 _ 3 _ _ 6 _ 8 _
_ 5 _ _ 8 _ 1 2 _
7 _ 9 1 _ 3 _ 5 6
_ 3 _ _ 6 7 _ 9 _
5 _ 7 8 _ _ _ 3 _
8 _ 1 _ 3 _ 5 _ 7
_ 4 _ _ 7 8 _ 1 _
6 _ 8 _ _ 2 _ 4 _
_ 1 2 _ 4 5 _ 7 8
`
	testInput2 = `_ _ _ _ _ 6 _ _ _
_ 5 9 _ _ _ _ _ 8
2 _ _ _ _ 8 _ _ _
_ 4 5 _ _ _ _ _ _
_ _ 3 _ _ _ _ _ _
_ _ 6 _ _ 3 _ 5 4
_ _ _ 3 2 5 _ _ 6
_ _ _ _ _ _ _ _ _
_ _ _ _ _ _ _ _ _
`
	testInput3 = `1 2 _ 4 5 6 7 8 9
_ _ 3 _ _ _ _ _ _
_ _ _ _ _ _ _ _ _
_ _ _ _ _ _ _ _ _
_ _ _ _ _ _ _ _ _
_ _ _ _ _ _ _ _ _
_ _ _ _ _ _ _ _ _
_ _ _ _ _ _ _ _ _
_ _ _ _ _ _ _ _ _
`
	testInput4 = `_ _ _ _ _ _ _ _ _
_ _ _ _ _ _ _ _ _
_ _ _ _ _ _ _ _ _
_ _ _ _ _ _ _ _ _
_ _ _ _ _ _ _ _ _
_ _ _ _ _ _ _ _ _
_ _ _ _ _ _ _ _ _
_ _ _ _ _ _ _ _ _
_ _ _ _ _ _ _ _ _
`
	testInput5 = `1 2 3 4 5 6 7 8 9
4 5 6 7 8 9 1 2 3
7 8 9 1 2 3 4 5 6
2 3 4 5 6 7 8 9 1
5 6 7 8 9 1 2 3 4
8 9 1 2 3 4 5 6 7
3 4 5 6 7 8 9 1 2
6 7 8 9 1 2 3 4 5
9 1 2 3 4 5 6 7 8
`
)

func TestSolvers(t *testing.T) {
	var tests = []struct {
		input       string
		shouldSolve bool
	}{
		{testInput1, true},
		{testInput2, true},
		{testInput3, false},
		{testInput4, true},
		{testInput5, true},
	}

	for _, test := range tests {
		b, err := puzzle.New([]byte(test.input))
		assert.NoError(t, err)
		assert.NotNil(t, b)

		for _, solver := range buildSolvers(b) {
			solvedBoard := solver.Solve()
			assert.Equal(t, test.shouldSolve, solvedBoard.Solved())
		}
	}
}

func BenchmarkBacktrack1(b *testing.B) {
	board, err := puzzle.New([]byte(testInput1))
	if err != nil {
		log.Fatalln(err)
	}

	benchmarkBacktrack(board, b)
}

func BenchmarkBacktrack2(b *testing.B) {
	board, err := puzzle.New([]byte(testInput2))
	if err != nil {
		log.Fatalln(err)
	}

	benchmarkBacktrack(board, b)
}

func BenchmarkBacktrack3(b *testing.B) {
	board, err := puzzle.New([]byte(testInput3))
	if err != nil {
		log.Fatalln(err)
	}

	benchmarkBacktrack(board, b)
}

func BenchmarkBacktrack4(b *testing.B) {
	board, err := puzzle.New([]byte(testInput4))
	if err != nil {
		log.Fatalln(err)
	}

	benchmarkBacktrack(board, b)
}

func BenchmarkBacktrack5(b *testing.B) {
	board, err := puzzle.New([]byte(testInput5))
	if err != nil {
		log.Fatalln(err)
	}

	benchmarkBacktrack(board, b)
}

func BenchmarkDLX1(b *testing.B) {
	board, err := puzzle.New([]byte(testInput1))
	if err != nil {
		log.Fatalln(err)
	}

	benchmarkDLX(board, b)
}

func BenchmarkDLX2(b *testing.B) {
	board, err := puzzle.New([]byte(testInput2))
	if err != nil {
		log.Fatalln(err)
	}

	benchmarkDLX(board, b)
}

func BenchmarkDLX3(b *testing.B) {
	board, err := puzzle.New([]byte(testInput3))
	if err != nil {
		log.Fatalln(err)
	}

	benchmarkDLX(board, b)
}

func BenchmarkDLX4(b *testing.B) {
	board, err := puzzle.New([]byte(testInput4))
	if err != nil {
		log.Fatalln(err)
	}

	benchmarkDLX(board, b)
}

func BenchmarkDLX5(b *testing.B) {
	board, err := puzzle.New([]byte(testInput5))
	if err != nil {
		log.Fatalln(err)
	}

	benchmarkDLX(board, b)
}

func buildSolvers(b *puzzle.Board) []puzzle.SudokuSolver {
	return []puzzle.SudokuSolver{backtrack.NewSolver(b), dlx.NewSolver(b)}
}

func benchmarkBacktrack(board *puzzle.Board, b *testing.B) {
	solver := backtrack.NewSolver(board)

	for i := 0; i < b.N; i++ {
		solver.Solve()
	}
}

func benchmarkDLX(board *puzzle.Board, b *testing.B) {
	solver := dlx.NewSolver(board)

	for i := 0; i < b.N; i++ {
		solver.Solve()
	}
}
