package sudoku_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/husobee/sudoku"
)

const (
	goodPuzzleSolve string = `1 2 3 4 5 6 7 8 9
4 5 6 7 8 9 1 2 3
7 8 9 1 2 3 4 5 6
2 3 4 5 6 7 8 9 1
5 6 7 8 9 1 2 3 4
8 9 1 2 3 4 5 6 7
3 4 5 6 7 8 9 1 2
6 7 8 9 1 2 3 4 5
9 1 2 3 4 5 6 7 8
`
	goodPuzzle string = `1 _ 3 _ _ 6 _ 8 _
_ 5 _ _ 8 _ 1 2 _
7 _ 9 1 _ 3 _ 5 6
_ 3 _ _ 6 7 _ 9 _
5 _ 7 8 _ _ _ 3 _
8 _ 1 _ 3 _ 5 _ 7
_ 4 _ _ 7 8 _ 1 _
6 _ 8 _ _ 2 _ 4 _
_ 1 2 _ 4 5 _ 7 8
`
	NanPuzzle string = `a b c _ _ 6 _ 8 _
_ 5 _ _ 8 _ 1 2 _
7 _ 9 1 _ 3 _ 5 6
_ 3 _ _ 6 7 _ 9 _
5 _ 7 8 _ _ _ 3 _
8 _ 1 _ 3 _ 5 _ 7
_ 4 _ _ 7 8 _ 1 _
6 _ 8 _ _ 2 _ 4 _
_ 1 2 _ 4 5 _ 7 8
`
	BadSpacesPuzzle string = `123   _ _ 6 _ 8 _
_ 5 _ _ 8 _ 1 2 _
7 _ 9 1 _ 3 _ 5 6
_ 3 _ _ 6 7 _ 9 _
5 _ 7 8 _ _ _ 3 _
8 _ 1 _ 3 _ 5 _ 7
_ 4 _ _ 7 8 _ 1 _
6 _ 8 _ _ 2 _ 4 _
_ 1 2 _ 4 5 _ 7 8
`
	InvalidLengthPuzzle string = `1 2 3 _ _ 6 _ 8 _ 1 2 2
_ 5 _ _ 8 _ 1 2 _
7 _ 9 1 _ 3 _ 5 6
_ 3 _ _ 6 7 _ 9 _
5 _ 7 8 _ _ _ 3 _
8 _ 1 _ 3 _ 5 _ 7
_ 4 _ _ 7 8 _ 1 _
6 _ 8 _ _ 2 _ 4 _
_ 1 2 _ 4 5 _ 7 8
`
	InvalidRowsToManyPuzzle string = `1 2 3 _ _ 6 _ 8 _
_ 5 _ _ 8 _ 1 2 _
7 _ 9 1 _ 3 _ 5 6
_ 3 _ _ 6 7 _ 9 _
5 _ 7 8 _ _ _ 3 _
8 _ 1 _ 3 _ 5 _ 7
_ 4 _ _ 7 8 _ 1 _
6 _ 8 _ _ 2 _ 4 _
_ 1 2 _ 4 5 _ 7 8
_ 1 2 _ 4 5 _ 7 8
`
	InvalidRowsToFewPuzzle string = `1 2 3 _ _ 6 _ 8 _
_ 5 _ _ 8 _ 1 2 _
7 _ 9 1 _ 3 _ 5 6
_ 3 _ _ 6 7 _ 9 _
5 _ 7 8 _ _ _ 3 _
8 _ 1 _ 3 _ 5 _ 7
`
)

func TestDump(t *testing.T) {
	p, err := sudoku.ParsePuzzle(strings.NewReader(goodPuzzle))
	if err != nil {
		t.Errorf("failed to parse a good puzzle, err=%s", err.Error())
	}
	buf := bytes.NewBuffer([]byte{})
	p.Dump(buf)
	if buf.String() != goodPuzzle {
		t.Errorf("failed to dump puzzle correctly")
	}
}
func TestParsePuzzle(t *testing.T) {
	if _, err := sudoku.ParsePuzzle(strings.NewReader(goodPuzzle)); err != nil {
		t.Errorf("failed to parse a good puzzle, err=%s", err.Error())
	}
	if _, err := sudoku.ParsePuzzle(strings.NewReader(NanPuzzle)); err == nil {
		t.Errorf("failed to error on a bad puzzle")
	}
	if _, err := sudoku.ParsePuzzle(strings.NewReader(BadSpacesPuzzle)); err == nil {
		t.Errorf("failed to error on a bad puzzle")
	}
	if _, err := sudoku.ParsePuzzle(strings.NewReader(InvalidRowsToFewPuzzle)); err == nil {
		t.Errorf("failed to error on a bad puzzle")
	}
	if _, err := sudoku.ParsePuzzle(strings.NewReader(InvalidRowsToManyPuzzle)); err == nil {
		t.Errorf("failed to error on a bad puzzle")
	}
	if _, err := sudoku.ParsePuzzle(strings.NewReader(InvalidLengthPuzzle)); err == nil {
		t.Errorf("failed to error on a bad puzzle")
	}
}
func TestSolvePuzzleRecursionDepth(t *testing.T) {
	p, err := sudoku.ParsePuzzle(strings.NewReader(goodPuzzle))
	if err != nil {
		t.Errorf("failed to parse a good puzzle, err=%s", err.Error())
	}
	sudoku.SetRecursionDepth(10)
	err = p.BacktrackSolve()
	if err != sudoku.ErrSolveExceedRecursionDepth {
		t.Errorf("should have errored due to recursion depth")
	}
}

func TestSolvePuzzle(t *testing.T) {
	sudoku.SetRecursionDepth(-1)
	p, err := sudoku.ParsePuzzle(strings.NewReader(goodPuzzle))
	if err != nil {
		t.Errorf("failed to parse a good puzzle, err=%s", err.Error())
	}
	p.BacktrackSolve()
	buf := bytes.NewBuffer([]byte{})
	p.Dump(buf)
	if buf.String() != goodPuzzleSolve {
		t.Errorf("failed to solve puzzle correctly")
	}
}

func BenchmarkParsePuzzle(b *testing.B) {
	sudoku.SetRecursionDepth(-1)
	for i := 0; i < b.N; i++ {
		sudoku.ParsePuzzle(strings.NewReader(goodPuzzle))
	}
}

func BenchmarkSolvePuzzle(b *testing.B) {
	sudoku.SetRecursionDepth(-1)
	p, err := sudoku.ParsePuzzle(strings.NewReader(goodPuzzle))
	if err != nil {
		b.Error(err.Error())
	}
	for i := 0; i < b.N; i++ {
		var testPuzzle sudoku.Puzzle = p
		if err := testPuzzle.BacktrackSolve(); err != nil {
			b.Error(err.Error())
		}
	}
}
