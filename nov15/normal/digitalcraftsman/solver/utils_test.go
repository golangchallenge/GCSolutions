package solver

import (
	"reflect"
	"strings"
	"testing"
)

func TestGetBoardFrom(t *testing.T) {
	for i, check := range []struct {
		input  string
		expect *Board
	}{
		{inputIllegalChar, nil},
		{inputMissingRow, nil},
		{inputShortRow, nil},
		{inputEasySudoku, boardEasySudoku},
		{inputHardSudoku, boardHardSudoku},
	} {
		// convert the string into an io.Reader
		reader := strings.NewReader(check.input)
		got, err := GetBoardFrom(reader)

		// invalid boards have no solution
		if err != nil && check.expect != nil {
			t.Fatalf("[%d] %s", i, err)
		}

		// if a solution exists check equality
		if check.expect != nil {
			if !reflect.DeepEqual(got, check.expect) {
				t.Errorf("[%d] the input board wasn't parsed correctly", i)
			}
		}
	}
}

func TestvalidateInputCell(t *testing.T) {
	for i, check := range []struct {
		input  string
		expect int
	}{
		{"0", 0},
		{"1", 1},
		{"4", 4},
		{"9", 9},
		{"_", 0},
		{"@", 0},
		{" ", 0},
		{".", 0},
		{"10", 0},
	} {
		digit, err := validateInputCell(check.input)
		if digit != check.expect {
			t.Errorf("[%d] got %d but expected %d - %s",
				i, digit, check.expect, err)
		}
	}
}

func TestBoard_String(t *testing.T) {
	for i, check := range []struct {
		board  *Board
		expect string
	}{
		{solutionEasySudoku, outputEasySudoku},
		{solutionHardSudoku, outputHardSudoku},
	} {
		if check.board.String() != check.expect {
			t.Errorf("[%d] the board wasn't printed correctly", i)
		}
	}
}
