package sudoku

import (
	"errors"
	"io"
	"strings"
	"testing"
)

var ErrBadReader = errors.New("Here comes some bad error")

type badTestReader struct {
	err int
}

func (b *badTestReader) Read([]byte) (int, error) {
	if b.err == 0 {
		return 0, ErrBadReader
	}
	return b.err, nil
}

func TestBadReaderCase(t *testing.T) {
	badReader := &badTestReader{err: 0}

	_, err := NewPuzzleFromReader(badReader)
	if err != ErrBadReader {
		t.Fatal("Bad reader should return 'ErrBadReader', but returns ",
			err)
	}

	badReader.err = 10
	_, err = NewPuzzleFromReader(badReader)
	if err != io.EOF {
		t.Fatal("Bad reader should return EOF, but returns ", err)
	}
}

func TestInvalidPuzzle(t *testing.T) {
	invalids := []string{
		"2 _ 6 8 _ 4 _ _ 6\n" +
			"_ _ 6 _ _ _ 5 _ _\n" +
			"_ 7 4 _ _ _ 9 2 _\n" +
			"3 _ _ _ 4 _ _ _ 7\n" +
			"_ _ _ 3 _ 5 _ _ _\n" +
			"4 _ _ _ 6 _ _ _ 9\n" +
			"_ 1 9 _ _ _ 7 4 _\n" +
			"_ _ 8 _ _ _ 2 _ _\n" +
			"5 _ _ 6 _ 8 _ _ 1",
		"2 _ _ 8 _ 4 _ _ 6\n" +
			"_ _ 6 _ _ _ 5 _ _\n" +
			"_ 7 4 _ _ _ 9 2 5\n" +
			"3 _ _ _ 4 _ _ _ 7\n" +
			"_ _ _ 3 _ 5 _ _ _\n" +
			"4 _ _ _ 6 _ _ _ 9\n" +
			"_ 1 9 _ _ _ 7 4 _\n" +
			"_ _ 8 _ _ _ 2 _ _\n" +
			"5 _ _ 6 _ 8 _ _ 1",
		"2 _ _ 8 _ 4 _ _ 6\n" +
			"_ _ 6 _ _ _ 5 _ _\n" +
			"_ 7 4 _ _ _ 9 2 _\n" +
			"3 _ _ _ 4 _ _ _ 7\n" +
			"_ _ _ 3 _ 5 _ _ _\n" +
			"4 _ 3 _ 6 _ _ _ 9\n" +
			"_ 1 9 _ _ _ 7 4 _\n" +
			"_ _ 8 _ _ _ 2 _ _\n" +
			"5 _ _ 6 _ 8 _ _ 1",
		"2 _ _ 8 _ 4 _ _ 6\n" +
			"_ _ 6 _ _ _ 5 _ _\n" +
			"_ 7 4 _ _ _ 9 2 _\n" +
			"3 _ _ _ 4 _ _ _ 7\n" +
			"_ _ _ 3 _ 5 _ _ _\n" +
			"4 _ _ _ 6 _ _ _ 9\n" +
			"2 1 9 _ _ _ 7 4 _\n" +
			"_ _ 8 _ _ _ 2 _ _\n" +
			"5 _ _ 6 _ 8 _ _ 1",
		"2 _ _ 8 _ 4 _ _ 6\n" +
			"_ _ 6 _ _ _ 5 _ _\n" +
			"_ a 4 _ _ _ 9 2 _\n" +
			"3 _ _ _ 4 _ _ _ 7\n" +
			"_ _ _ 3 _ 5 _ _ _\n" +
			"4 _ _ _ 6 _ _ _ 9\n" +
			"_ 1 9 _ _ _ 7 4 _\n" +
			"_ _ 8 _ _ _ 2 _ _\n" +
			"5 _ _ 6 _ 8 _ _ 1",
	}

	for i, invalidPuzzle := range invalids {
		_, err := NewPuzzleFromReader(strings.NewReader(invalidPuzzle))
		if err != ErrInvalidPuzzle {
			t.Fatal("Invalid puzzle:", i, "\n"+invalidPuzzle,
				"\nErr: ", err)
		}
	}
}

func TestRowChecker(t *testing.T) {
	row := puzzleRow{cells: [9]byte{
		'1', '2', '3', '4', '5', '6', '7', '8', '9'}}
	invalidRows := []puzzleRow{}
	tmpRow := row
	tmpRow.cells[1] = '3'
	invalidRows = append(invalidRows, tmpRow)

	tmpRow = row
	tmpRow.cells[2] = 'x'
	invalidRows = append(invalidRows, tmpRow)

	tmpRow = row
	tmpRow.cells[3] = '3'
	invalidRows = append(invalidRows, tmpRow)

	tmpRow = row
	tmpRow.cells[3] = emptyCell
	tmpRow.cells[4] = '3'
	invalidRows = append(invalidRows, tmpRow)

	for _, invalid := range invalidRows {
		if invalid.check() {
			t.Error("Row ", invalid, " is invalid")
		}
	}

	if !row.check() {
		t.Error("Row ", row, " is valid")
	}

	for i := range row.cells {
		row.cells[i] = emptyCell
		if !row.check() {
			t.Error("Row ", row, " is valid")
		}
	}
}

func TestStopOnError(t *testing.T) {
	puzzle := Puzzle{err: ErrInvalidPuzzle}

	if puzzle.IsSolved() {
		t.Error("Puzzle should not be solved!")
	}

	if _, e := puzzle.Solve(); e != ErrInvalidPuzzle {
		t.Error("Puzzle.Solve() should return ", ErrInvalidPuzzle)
	}
}
