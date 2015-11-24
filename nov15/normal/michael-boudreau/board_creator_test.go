package main

import (
	"strings"
	"testing"
)

func TestReduceBoardToDifficultyEasy(t *testing.T) {
	for i := 0; i < 100; i++ {
		board := newSampleTestBoard(9)
		ReduceBoardToDifficulty(board, int(EasyBoard))
		count := getNumberOfFilledValues(board)
		if count < 32 || count > 36 {
			t.Logf("Expecting easy board to be in range of 32-36, instead got %d", count)
		}
	}

	// hard 21-25
	// medium 26-31
	// easy 32-36
}
func TestReduceBoardToDifficultyMedium(t *testing.T) {
	for i := 0; i < 100; i++ {
		board := newSampleTestBoard(9)
		ReduceBoardToDifficulty(board, int(MediumBoard))
		count := getNumberOfFilledValues(board)
		if count < 26 || count > 31 {
			t.Logf("Expecting medium board to be in range of 26-31, instead got %d", count)
		}
	}
}
func TestReduceBoardToDifficultyHard(t *testing.T) {
	for i := 0; i < 100; i++ {
		board := newSampleTestBoard(9)
		ReduceBoardToDifficulty(board, int(HardBoard))
		count := getNumberOfFilledValues(board)
		if count < 21 || count > 25 {
			t.Logf("Expecting hard board to be in range of 21-25, instead got %d", count)
		}
	}
}
func TestEstimateDifficultyEasyOne(t *testing.T) {
	if EstimateDifficulty(testDifficultyBoard33) != EasyBoard {
		t.Logf("Expecting Difficulty to be %v", EasyBoard)
		t.Fail()
	}
}
func TestEstimateDifficultyEasyTwo(t *testing.T) {
	if EstimateDifficulty(testDifficultyBoard32) != EasyBoard {
		t.Logf("Expecting Difficulty to be %v", EasyBoard)
		t.Fail()
	}
}
func TestEstimateDifficultyMediumOne(t *testing.T) {
	if EstimateDifficulty(testDifficultyBoard27) != MediumBoard {
		t.Logf("Expecting Difficulty to be %v", MediumBoard)
		t.Fail()
	}
}
func TestEstimateDifficultyMediumTwo(t *testing.T) {
	if EstimateDifficulty(testDifficultyBoard26) != MediumBoard {
		t.Logf("Expecting Difficulty to be %v", MediumBoard)
		t.Fail()
	}
}
func TestEstimateDifficultyHardOne(t *testing.T) {
	if EstimateDifficulty(testDifficultyBoard25) != HardBoard {
		t.Logf("Expecting Difficulty to be %v", HardBoard)
		t.Fail()
	}
}
func TestCreateInitialBoardWithStartingCountAndSize(t *testing.T) {
	board := CreateInitialBoardWithStartingCountAndSize(9, 10)
	count := getNumberOfFilledValues(board)

	if count != 10 {
		t.Logf("Expecting %d values created in new board, but found %d", 10, count)
		t.Fail()
	}
}

func getNumberOfFilledValues(board Board) int {
	count := 0
	for x := 0; x < 9; x++ {
		for y := 0; y < 9; y++ {
			if board[x][y] != 0 {
				count++
			}
		}
	}
	return count
}

func TestBoardFromReader(t *testing.T) {
	board, err := BoardFromReader(strings.NewReader(readerTestBoardText))

	if err != nil {
		t.Logf("Could not read board. err=%v", err)
		t.Fail()
		return
	}

	if assertBoardNotEqual(board, readerTestBoard) {
		t.Logf("Boards are not equal. Expecting \n[%v]\n to be equal to \n[%v]\n", board, readerTestBoard)
		t.Fail()
	}
}

var readerTestBoardText = `_ _ _ _ _ 7 _ 5 _
4 _ _ _ _ 2 _ 3 _
_ _ 9 _ _ _ 7 2 4
_ 3 _ _ 9 1 _ _ _
9 _ 1 _ _ 6 _ 7 5
_ _ 8 5 _ 4 6 9 _
5 _ 6 _ _ 8 _ _ _
3 _ _ 4 _ _ 5 1 _
1 _ _ 3 7 _ 2 _ _
`
var readerTestBoard = Board([][]byte{
	{0, 0, 0, 0, 0, 7, 0, 5, 0},
	{4, 0, 0, 0, 0, 2, 0, 3, 0},
	{0, 0, 9, 0, 0, 0, 7, 2, 4},
	{0, 3, 0, 0, 9, 1, 0, 0, 0},
	{9, 0, 1, 0, 0, 6, 0, 7, 5},
	{0, 0, 8, 5, 0, 4, 6, 9, 0},
	{5, 0, 6, 0, 0, 8, 0, 0, 0},
	{3, 0, 0, 4, 0, 0, 5, 1, 0},
	{1, 0, 0, 3, 7, 0, 2, 0, 0},
})

var testDifficultyBoard33 = Board([][]byte{
	{1, 9, 6, 5, 4, 3, 0, 0, 0},
	{0, 2, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 3, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 4, 6, 5, 3, 2, 1},
	{0, 0, 0, 0, 5, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 6, 0, 0, 0},
	{2, 5, 1, 0, 0, 0, 7, 4, 3},
	{4, 3, 7, 0, 0, 0, 6, 8, 2},
	{6, 8, 9, 0, 0, 0, 5, 1, 0},
})
var testDifficultyBoard32 = Board([][]byte{
	{1, 9, 6, 5, 4, 0, 0, 0, 0},
	{0, 2, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 3, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 4, 6, 5, 3, 2, 1},
	{0, 0, 0, 0, 5, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 6, 0, 0, 0},
	{2, 5, 1, 0, 0, 0, 7, 4, 3},
	{4, 3, 7, 0, 0, 0, 6, 8, 2},
	{6, 8, 9, 0, 0, 0, 5, 1, 0},
})
var testDifficultyBoard26 = Board([][]byte{
	{1, 9, 6, 5, 0, 0, 0, 0, 0},
	{0, 2, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 3, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 4, 0, 5, 3, 2, 1},
	{0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 6, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 7, 4, 3},
	{4, 3, 7, 0, 0, 0, 6, 8, 2},
	{6, 8, 9, 0, 0, 0, 5, 1, 0},
})
var testDifficultyBoard27 = Board([][]byte{
	{1, 9, 6, 5, 0, 0, 8, 0, 0},
	{0, 2, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 3, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 4, 0, 5, 3, 2, 1},
	{0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 6, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 7, 4, 3},
	{4, 3, 7, 0, 0, 0, 6, 8, 2},
	{6, 8, 9, 0, 0, 0, 5, 1, 0},
})
var testDifficultyBoard25 = Board([][]byte{
	{1, 9, 6, 5, 0, 0, 8, 0, 0},
	{0, 2, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 3, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 4, 0, 5, 3, 2, 1},
	{0, 0, 0, 0, 0, 0, 0, 0, 0},
	{0, 0, 0, 0, 0, 6, 0, 0, 0},
	{0, 0, 0, 0, 0, 0, 7, 4, 3},
	{4, 3, 7, 0, 0, 0, 0, 8, 2},
	{6, 8, 9, 0, 0, 0, 0, 1, 0},
})
