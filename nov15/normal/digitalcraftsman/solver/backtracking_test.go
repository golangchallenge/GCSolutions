package solver

import (
	"reflect"
	"testing"
)

var (
	//
	//	Illegal inputs
	//

	inputIllegalChar = `1 _ 3 _ _ 6 _ 8 _
_ 5 _ _ 8 _ 1 2 _
7 _ 9 1 _ 3 _ 5 6
_ 3 _ _ 6 7 _ 9 _
5 _ 7 8 _ _ _ 3 _
8 _ 1 @ 3 _ 5 _ 7
_ 4 _ _ 7 8 _ 1 _
6 _ 8 _ _ 2 _ 4 _
_ 1 2 _ 4 5 _ 7 8`

	inputMissingRow = `1 _ 3 _ _ 6 _ 8 _
_ 5 _ _ 8 _ 1 2 _
7 _ 9 1 _ 3 _ 5 6
_ 3 _ _ 6 7 _ 9 _
5 _ 7 8 _ _ _ 3 _
8 _ 1 _ 3 _ 5 _ 7
_ 4 _ _ 7 8 _ 1 _
6 _ 8 _ _ 2 _ 4 _`

	inputShortRow = `1 _ 3 _ _ 6 _ 8 _
_ 5 _ _ 8 _ 1 2 _
_ 3 _ _ 6 7 _ 9 _
5 _ 7 8 _ _
8 _ 1 _ 3 _ 5 _ 7
_ 4 _ _ 7 8 _ 1 _
6 _ 8 _ _ 2 _ 4 _
_ 1 2 _ 4 5 _ 7 8`

	//
	//	Valid Sudokus to solve
	//

	// #1 difficulty: easy
	inputEasySudoku = `1 _ 3 _ _ 6 _ 8 _
_ 5 _ _ 8 _ 1 2 _
7 _ 9 1 _ 3 _ 5 6
_ 3 _ _ 6 7 _ 9 _
5 _ 7 8 _ _ _ 3 _
8 _ 1 _ 3 _ 5 _ 7
_ 4 _ _ 7 8 _ 1 _
6 _ 8 _ _ 2 _ 4 _
_ 1 2 _ 4 5 _ 7 8`

	boardEasySudoku = &Board{
		Cells: [9][9]int{
			{1, 0, 3, 0, 0, 6, 0, 8, 0},
			{0, 5, 0, 0, 8, 0, 1, 2, 0},
			{7, 0, 9, 1, 0, 3, 0, 5, 6},
			{0, 3, 0, 0, 6, 7, 0, 9, 0},
			{5, 0, 7, 8, 0, 0, 0, 3, 0},
			{8, 0, 1, 0, 3, 0, 5, 0, 7},
			{0, 4, 0, 0, 7, 8, 0, 1, 0},
			{6, 0, 8, 0, 0, 2, 0, 4, 0},
			{0, 1, 2, 0, 4, 5, 0, 7, 8},
		},
	}

	solutionEasySudoku = &Board{
		Cells: [9][9]int{
			{1, 2, 3, 4, 5, 6, 7, 8, 9},
			{4, 5, 6, 7, 8, 9, 1, 2, 3},
			{7, 8, 9, 1, 2, 3, 4, 5, 6},
			{2, 3, 4, 5, 6, 7, 8, 9, 1},
			{5, 6, 7, 8, 9, 1, 2, 3, 4},
			{8, 9, 1, 2, 3, 4, 5, 6, 7},
			{3, 4, 5, 6, 7, 8, 9, 1, 2},
			{6, 7, 8, 9, 1, 2, 3, 4, 5},
			{9, 1, 2, 3, 4, 5, 6, 7, 8},
		},
	}

	outputEasySudoku = `1 2 3 4 5 6 7 8 9
4 5 6 7 8 9 1 2 3
7 8 9 1 2 3 4 5 6
2 3 4 5 6 7 8 9 1
5 6 7 8 9 1 2 3 4
8 9 1 2 3 4 5 6 7
3 4 5 6 7 8 9 1 2
6 7 8 9 1 2 3 4 5
9 1 2 3 4 5 6 7 8`

	// #2 difficulty: hard
	// taken from apollon.issp.u-tokyo.ac.jp/~watanabe/sample/sudoku/
	inputHardSudoku = `_ 6 1 _ _ 7 _ _ 3
_ 9 2 _ _ 3 _ _ _
_ _ _ _ _ _ _ _ _
_ _ 8 5 3 _ _ _ _
_ _ _ _ _ _ 5 _ 4
5 _ _ _ _ 8 _ _ _
_ 4 _ _ _ _ _ _ 1
_ _ _ 1 6 _ 8 _ _
6 _ _ _ _ _ _ _ _`

	boardHardSudoku = &Board{
		Cells: [9][9]int{
			{0, 6, 1, 0, 0, 7, 0, 0, 3},
			{0, 9, 2, 0, 0, 3, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 0, 0, 0},
			{0, 0, 8, 5, 3, 0, 0, 0, 0},
			{0, 0, 0, 0, 0, 0, 5, 0, 4},
			{5, 0, 0, 0, 0, 8, 0, 0, 0},
			{0, 4, 0, 0, 0, 0, 0, 0, 1},
			{0, 0, 0, 1, 6, 0, 8, 0, 0},
			{6, 0, 0, 0, 0, 0, 0, 0, 0},
		},
	}

	solutionHardSudoku = &Board{
		Cells: [9][9]int{
			{4, 6, 1, 9, 8, 7, 2, 5, 3},
			{7, 9, 2, 4, 5, 3, 1, 6, 8},
			{3, 8, 5, 2, 1, 6, 4, 7, 9},
			{1, 2, 8, 5, 3, 4, 7, 9, 6},
			{9, 3, 6, 7, 2, 1, 5, 8, 4},
			{5, 7, 4, 6, 9, 8, 3, 1, 2},
			{8, 4, 9, 3, 7, 5, 6, 2, 1},
			{2, 5, 3, 1, 6, 9, 8, 4, 7},
			{6, 1, 7, 8, 4, 2, 9, 3, 5},
		},
	}

	outputHardSudoku = `4 6 1 9 8 7 2 5 3
7 9 2 4 5 3 1 6 8
3 8 5 2 1 6 4 7 9
1 2 8 5 3 4 7 9 6
9 3 6 7 2 1 5 8 4
5 7 4 6 9 8 3 1 2
8 4 9 3 7 5 6 2 1
2 5 3 1 6 9 8 4 7
6 1 7 8 4 2 9 3 5`
)

func copyBoard(cells [N][N]int) *Board {
	return &Board{Cells: cells}
}

func TestBoard_isDigitValid(t *testing.T) {
	for i, check := range []struct {
		row, col, digit int
		expect          bool
	}{
		{4, 6, 1, false},
		{4, 6, 2, true},
		{4, 6, 3, false},
		{4, 6, 4, true},
		{4, 6, 5, false},
		{4, 6, 6, true},
		{4, 6, 7, false},
		{4, 6, 8, false},
		{4, 6, 9, false},
	} {
		if boardEasySudoku.isDigitValid(check.row, check.col, check.digit) != check.expect {
			t.Errorf("[%d] did not expect to find digit %d in column %d, row %d or in the corresponding 3x3 section",
				i, check.digit, check.col, check.row)
		}
	}
}

func TestBoard_Backtrack(t *testing.T) {
	for i, check := range []struct {
		board, expect *Board
	}{
		{boardEasySudoku, solutionEasySudoku},
		{boardHardSudoku, solutionHardSudoku},
	} {
		// create copy the board since Backtrack() modifies the orignal values
		copy := copyBoard(check.board.Cells)

		if !copy.Backtrack() || !reflect.DeepEqual(copy, check.expect) {
			t.Errorf("[%d] the given board wasn't solved as expected", i)
		}
	}
}

func TestBoard_findEmptyCell(t *testing.T) {
	for i, check := range []struct {
		board           *Board
		row, col        int
		expectEmptyCell bool
	}{
		{boardEasySudoku, 0, 1, true},
		{boardHardSudoku, 0, 0, true},
		{solutionEasySudoku, 0, 0, false}, // solutions are already filled
		{solutionHardSudoku, 0, 0, false},
	} {
		nextRow, nextCol, foundEmptyCell := check.board.findEmptyCell()

		if foundEmptyCell != check.expectEmptyCell || nextRow != check.row || nextCol != check.col {
			t.Errorf("[%d] expect to find empty cell in row %d and col %d, got row %d, col %d",
				i, check.row, check.col, nextRow, nextCol)
		}
	}
}

func BenchmarkBacktrackEasySudoku(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// prevent the modification of the orignal board
		copy := copyBoard(boardEasySudoku.Cells)
		b.StartTimer()
		copy.Backtrack()
	}
}

func BenchmarkBacktrackHardSudoku(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		// prevent the modification of the orignal board
		copy := copyBoard(boardHardSudoku.Cells)
		b.StartTimer()
		copy.Backtrack()
	}
}
