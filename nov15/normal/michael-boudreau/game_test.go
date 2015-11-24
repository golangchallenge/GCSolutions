package main

import (
	"strings"
	"testing"
)

var gameTestcases = []struct {
	board  Board
	coord  *Coord
	expect byte
}{
	{
		board:  challengeBoard,
		coord:  &Coord{0, 0},
		expect: byte(0),
	},
	{
		board:  challengeBoard,
		coord:  &Coord{0, 1},
		expect: byte(2),
	},
}

func TestGameClosestCoordinateSolver(t *testing.T) {
	game := NewSudokuGame(&Config{})
	game.SetCoordinateRanker(&ClosestCoordFinder{})

	for i, test := range gameTestcases {
		for _, value := range numberSet {
			board := test.board.Clone()
			err := game.solveNextValue(board, test.coord, value)
			if test.expect == 0 && err == nil {
				t.Logf("[%d] Expecting coord %+v to already be filled", i, test.coord)
				t.Fail()
			} else if test.expect == value && err != nil {
				t.Logf("[%d] Expecting coord %+v to add value %v, but did not (%d,%d): %v", i, test.coord, test.expect, value, test.expect, err)
				t.Fail()
			} else if test.expect != value && err == nil {
				t.Logf("[%d] Expecting coord %+v to not allow value %v, but [%v] (%d,%d)", i, test.coord, value, test.expect, value, test.expect)
				t.Fail()
			}
		}
	}
}
func TestGameRankedCoordinateSolver(t *testing.T) {
	game := NewSudokuGame(&Config{})
	game.SetCoordinateRanker(&RankedCoordFinder{})

	for i, test := range gameTestcases {
		for _, value := range numberSet {
			board := test.board.Clone()
			err := game.solveNextValue(board, test.coord, value)
			if test.expect == 0 && err == nil {
				t.Logf("[%d] Expecting coord %+v to already be filled", i, test.coord)
				t.Fail()
			} else if test.expect == value && err != nil {
				t.Logf("[%d] Expecting coord %+v to add value %v, but did not (%d,%d): %v", i, test.coord, test.expect, value, test.expect, err)
				t.Fail()
			} else if test.expect != value && err == nil {
				t.Logf("[%d] Expecting coord %+v to not allow value %v, but [%v] (%d,%d)", i, test.coord, value, test.expect, value, test.expect)
				t.Fail()
			}
		}
	}
}

func BenchmarkChallengeBoardRankedCoordinateGameSolver(b *testing.B) {
	benchCoordinateSolverWithBoard(b, &RankedCoordFinder{}, challengeBoard.Clone())
}
func BenchmarkChallengeBoardClosestCoordinateGameSolver(b *testing.B) {
	benchCoordinateSolverWithBoard(b, &ClosestCoordFinder{}, challengeBoard.Clone())
}
func BenchmarkEasyOneRankedCoordinateGameSolver(b *testing.B) {
	benchCoordinateSolverWithBoardText(b, &RankedCoordFinder{}, easyBoardOne)
}
func BenchmarkEasyOneClosestCoordinateGameSolver(b *testing.B) {
	benchCoordinateSolverWithBoardText(b, &ClosestCoordFinder{}, easyBoardOne)
}
func BenchmarkEasyTwoRankedCoordinateGameSolver(b *testing.B) {
	benchCoordinateSolverWithBoardText(b, &RankedCoordFinder{}, easyBoardTwo)
}
func BenchmarkEasyTwoClosestCoordinateGameSolver(b *testing.B) {
	benchCoordinateSolverWithBoardText(b, &ClosestCoordFinder{}, easyBoardTwo)
}
func BenchmarkMediumOneRankedCoordinateGameSolver(b *testing.B) {
	benchCoordinateSolverWithBoardText(b, &RankedCoordFinder{}, mediumBoardOne)
}
func BenchmarkMediumOneClosestCoordinateGameSolver(b *testing.B) {
	benchCoordinateSolverWithBoardText(b, &ClosestCoordFinder{}, mediumBoardOne)
}
func BenchmarkMediumTwoRankedCoordinateGameSolver(b *testing.B) {
	benchCoordinateSolverWithBoardText(b, &RankedCoordFinder{}, mediumBoardTwo)
}
func BenchmarkMediumTwoClosestCoordinateGameSolver(b *testing.B) {
	benchCoordinateSolverWithBoardText(b, &ClosestCoordFinder{}, mediumBoardTwo)
}
func BenchmarkHardOneRankedCoordinateGameSolver(b *testing.B) {
	benchCoordinateSolverWithBoardText(b, &RankedCoordFinder{}, hardBoardOne)
}
func BenchmarkHardOneClosestCoordinateGameSolver(b *testing.B) {
	benchCoordinateSolverWithBoardText(b, &ClosestCoordFinder{}, hardBoardOne)
}
func BenchmarkHardTwoRankedCoordinateGameSolver(b *testing.B) {
	benchCoordinateSolverWithBoardText(b, &RankedCoordFinder{}, hardBoardTwo)
}
func BenchmarkHardTwoClosestCoordinateGameSolver(b *testing.B) {
	benchCoordinateSolverWithBoardText(b, &ClosestCoordFinder{}, hardBoardTwo)
}

func benchCoordinateSolverWithBoardText(b *testing.B, finder CoordinateFinder, boardtext string) {
	game := NewSudokuGame(&Config{})
	game.SetCoordinateRanker(finder)

	board, err := BoardFromReader(strings.NewReader(boardtext))
	if err != nil {
		b.Logf("Caught Error:", err)
		b.Fail()
		return
	}

	benchCoordinateSolverWithBoard(b, finder, board)
}
func benchCoordinateSolverWithBoard(b *testing.B, finder CoordinateFinder, board Board) {
	game := NewSudokuGame(&Config{})
	game.SetCoordinateRanker(finder)

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		game.Solve(board)
	}
}

var (
	easyBoardOne = `_ _ _ _ _ 7 _ 5 _
4 _ _ _ _ 2 _ 3 _
_ _ 9 _ _ _ 7 2 4
_ 3 _ _ 9 1 _ _ _
9 _ 1 _ _ 6 _ 7 5
_ _ 8 5 _ 4 6 9 _
5 _ 6 _ _ 8 _ _ _
3 _ _ 4 _ _ 5 1 _
1 _ _ 3 7 _ 2 _ _
`
	easyBoardTwo = `_ _ _ _ _ _ 7 _ _
_ 5 _ 4 8 _ _ _ _
8 9 4 1 6 7 5 _ _
3 1 _ _ _ _ 2 4 _
_ _ 5 _ 1 _ _ _ 9
4 8 7 3 _ _ _ _ _
_ _ 1 _ 3 8 _ _ 6
_ _ _ 7 4 6 _ _ _
6 4 2 _ 5 _ 8 _ _
`
	mediumBoardOne = `8 _ _ _ _ _ 6 _ _
_ 2 3 _ _ _ _ 7 _
_ _ 4 9 _ 3 _ _ _
_ _ 2 7 _ _ 8 _ 3
5 _ _ 4 3 2 _ _ _
_ _ _ _ 8 _ 4 _ 5
_ _ _ _ _ _ _ _ _
_ _ 5 _ 9 1 3 _ _
2 _ _ 3 4 _ _ _ _
`
	mediumBoardTwo = `_ _ _ 3 _ _ _ _ 6
_ 7 3 _ _ _ 9 _ _
_ 8 _ 2 _ _ 3 1 7
_ _ _ _ _ _ _ 2 _
_ _ 1 4 5 2 _ _ 8
_ _ _ _ _ 9 1 _ _
_ _ _ 9 3 _ 6 _ _
1 _ _ _ 2 _ _ _ 5
8 _ _ _ 6 _ _ _ _
`
	hardBoardOne = `_ _ _ _ _ _ 1 _ _
_ _ 5 _ _ _ _ _ _
8 6 _ 4 _ 7 _ _ 2
1 _ 3 9 _ 4 _ _ _
_ _ 6 _ _ _ _ _ _
9 _ _ _ _ _ 3 _ _
_ _ _ _ _ _ _ _ 9
_ 1 _ _ 4 _ _ _ 3
_ 8 _ _ _ _ 4 _ _
`
	hardBoardTwo = `_ _ 6 2 5 _ _ 3 8
_ 5 8 _ _ _ _ 7 _
_ _ _ _ _ _ _ 5 _
5 _ _ 4 _ _ _ _ _
_ 3 _ _ _ _ _ _ _
1 _ 7 _ _ _ _ _ _
_ _ _ 3 _ _ _ 4 _
_ 9 3 _ 2 _ _ 6 _
_ 8 _ _ _ 4 _ _ _
`
)
