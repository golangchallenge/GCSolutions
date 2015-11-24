package main

import (
	"strings"
	"testing"
)

var emptySudoku = `_ _ _ _ _ _ _ _ _
_ _ _ _ _ _ _ _ _
_ _ _ _ _ _ _ _ _
_ _ _ _ _ _ _ _ _
_ _ _ _ _ _ _ _ _
_ _ _ _ _ _ _ _ _
_ _ _ _ _ _ _ _ _
_ _ _ _ _ _ _ _ _
_ _ _ _ _ _ _ _ _
`

var testSudoku = `1 _ 3 _ _ 6 _ 8 _
_ 5 _ _ 8 _ 1 2 _
7 _ 9 1 _ 3 _ 5 6
_ 3 _ _ 6 7 _ 9 _
5 _ 7 8 _ _ _ 3 _
8 _ 1 _ 3 _ 5 _ 7
_ 4 _ _ 7 8 _ 1 _
6 _ 8 _ _ 2 _ 4 _
_ 1 2 _ 4 5 _ 7 8
`

var benchmarkSudoku = `
_ _ _ _ _ _ _ 1 2
5 _ _ _ _ 8 _ _ _
_ _ _ 7 _ _ _ _ _
6 _ _ 1 2 _ _ _ _
7 _ _ _ _ _ 4 5 _
_ _ _ _ 3 _ _ _ _
_ 3 _ _ _ _ 8 _ _
_ _ _ 5 _ _ 7 _ _
_ 2 _ _ _ _ _ _ _
`

var invalidSudoku = `1 _ 3 _ _ 6 _ 8 _
_ 5 _ _ 8 _ 1 2 _
7 _ 9 1 _ 3 _ 5 6
_ 3 _ _ 6 7 _ 9 _
5 _ 7 8 _ _ _ 3 _
8 _ 1 _ 3 _ 5 _ 7
_ 4 _ _ 7 8 _ 1 _
6 _ 8 _ _ 2 _ 4 _
_ 1 2 _ 4 5 _ 7 6
`

var unsolvableSudoku = `1 2 3 4 5 6 7 8 _
_ _ _ _ _ _ _ _ 9
_ _ _ _ _ _ _ _ _
_ _ _ _ _ _ _ _ _
_ _ _ _ _ _ _ _ _
_ _ _ _ _ _ _ _ _
_ _ _ _ _ _ _ _ _
_ _ _ _ _ _ _ _ _
_ _ _ _ _ _ _ _ _
`

var solvedSudoku = `1 2 3 4 5 6 7 8 9
4 5 6 7 8 9 1 2 3
7 8 9 1 2 3 4 5 6
2 3 4 5 6 7 8 9 1
5 6 7 8 9 1 2 3 4
8 9 1 2 3 4 5 6 7
3 4 5 6 7 8 9 1 2
6 7 8 9 1 2 3 4 5
9 1 2 3 4 5 6 7 8
`
var hardSudoku = `_ _ _ _ 6 _ _ 8 _
_ 2 _ _ _ _ _ _ _
_ _ 1 _ _ _ _ _ _
_ 7 _ _ _ _ 1 _ 2
5 _ _ _ 3 _ _ _ _
_ _ _ _ _ _ 4 _ _
_ _ 4 2 _ 1 _ _ _
3 _ _ 7 _ _ 6 _ _
_ _ _ _ _ _ _ 5 _`

var tooManyLines = `1 _ 3 _ _ 6 _ 8 _
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
var tooFewLines = `1 _ 3 _ _ 6 _ 8 _
_ 5 _ _ 8 _ 1 2 _
7 _ 9 1 _ 3 _ 5 6
_ 3 _ _ 6 7 _ 9 _
5 _ 7 8 _ _ _ 3 _
8 _ 1 _ 3 _ 5 _ 7
_ 4 _ _ 7 8 _ 1 _
6 _ 8 _ _ 2 _ 4 _
`
var tooManyCols = `1 _ 3 _ _ 6 _ 8 _ 4
_ 5 _ _ 8 _ 1 2 _ 4
7 _ 9 1 _ 3 _ 5 6 4
_ 3 _ _ 6 7 _ 9 _ 4
5 _ 7 8 _ _ _ 3 _ 4
8 _ 1 _ 3 _ 5 _ 7 4
_ 4 _ _ 7 8 _ 1 _ 4
6 _ 8 _ _ 2 _ 4 _ 4
`

var tooFewCols = `1 _ 3 _ _ 6 _ 8
_ 5 _ _ 8 _ 1 2
7 _ 9 1 _ 3 _ 5
_ 3 _ _ 6 7 _ 9
5 _ 7 8 _ _ _ 3
8 _ 1 _ 3 _ 5 _
_ 4 _ _ 7 8 _ 1
6 _ 8 _ _ 2 _ 4
`

var incorrectFormat = `ä ö ü ß ! " °? a b
1 2 3 4 5 6 7 8 9`

var wrongNumbers = `1 30 3 4 5 6 7 8 9
4 5 6 7 8 9 1 2 3
7 8 9 1 2 3 4 5 6
2 3 4 5 6 7 8 9 1
5 6 7 8 9 1 2 3000 4
8 9 1 2 3 4 5 6 7
3 4 5 6 7 8 9 1 2
6 7 8 9 1 2 3 4 5
9 1 2 3 4 5 6 7 8
`

func validNums(nums map[int]bool) bool {
	res := true
	for i := 1; i < 10; i++ {
		res = res && nums[i]
	}
	return res
}

func isValidRow(s Sudoku, row int, col int) bool {
	nums := make(map[int]bool)

	for i := 0; i < SudokuSize; i++ {
		nums[s[i][col]] = true
	}

	return validNums(nums)
}

func isValidCol(s Sudoku, row int, col int) bool {
	nums := make(map[int]bool)

	for i := 0; i < SudokuSize; i++ {
		nums[s[row][i]] = true
	}

	return validNums(nums)
}

func isValidSquare(s Sudoku, row int, col int) bool {
	nums := make(map[int]bool)

	sqRow := row / 3
	sqCol := col / 3

	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			nums[s[sqRow*3+i][sqCol*3+j]] = true
		}
	}

	return validNums(nums)
}

func isValidSolution(s Sudoku) bool {
	for i := 0; i < SudokuSize; i++ {
		for j := 0; j < SudokuSize; j++ {
			if !(isValidSquare(s, i, j) && isValidRow(s, i, j) && isValidCol(s, i, j)) {
				return false
			}
		}
	}

	return true
}

func TestIsValid(t *testing.T) {
	s, _ := parseSudoku(testSudoku)
	if !s.IsValid() {
		t.Errorf("%v\n should be valid", s)
	}

	s, _ = parseSudoku(invalidSudoku)
	if s.IsValid() {
		t.Errorf("%v\n should not be valid", s)
	}
}

func TestIsValidCell(t *testing.T) {
	s, _ := parseSudoku(invalidSudoku)

	row, col := 8, 8
	if s.validCell(row, col) {
		t.Errorf("(%v,%v) should not be a valid cell in \n%v\n", row, col, s)
	}

	row, col = 0, 0
	if !s.validCell(row, col) {
		t.Errorf("(%v,%v) should be a valid cell in \n%v\n", row, col, s)
	}
}

func TestParse(t *testing.T) {
	testStrs := [...]string{emptySudoku, testSudoku, solvedSudoku}

	for _, str := range testStrs {
		r := strings.NewReader(str)
		s, e := Parse(r)
		if e != nil {
			t.Errorf("Could not parse Sudoku:\n%v", str)
		}
		if s.String() != str {
			t.Errorf("Expected output:\n%v\nactual output:\n%v", str, s.String())
		}
	}
}

func TestParseErrors(t *testing.T) {
	testStrs := [...]string{tooManyLines, tooFewLines, tooManyCols, incorrectFormat, wrongNumbers, tooFewCols}
	for _, str := range testStrs {
		r := strings.NewReader(str)
		_, e := Parse(r)
		if e == nil {
			t.Errorf("Parsing \n%v\n gave no error!", str)
		}
	}
}

func TestSolveErrors(t *testing.T) {
	testStrs := [...]string{invalidSudoku, unsolvableSudoku}
	for _, str := range testStrs {
		s, _ := parseSudoku(str)
		_, e := s.Solve()

		if e == nil {
			t.Errorf("Solving the sudoku \n%v\n should give an error!", s)
		}
	}
}

func TestSolve(t *testing.T) {
	testStrs := [...]string{testSudoku, emptySudoku, benchmarkSudoku}

	for _, str := range testStrs {
		r := strings.NewReader(str)
		s, e := Parse(r)
		sol, e := s.Solve()
		if e != nil {
			t.Errorf("Could not solve solvable Sudoku.")
		}

		if !isValidSolution(sol) {
			t.Errorf("Wrong solution \n%v\n for\n%v\n", sol.String(), str)
		}
	}
}

func parseSudoku(str string) (Sudoku, error) {
	r := strings.NewReader(str)
	s, e := Parse(r)
	return s, e
}

func index(p pos) int {
	return 9*p.row + p.col
}

func TestGetCells(t *testing.T) {
	expected := []int{4, 36, 13, 37, 22, 38, 31, 39, 49, 41, 58, 42, 67, 43, 76, 44, 30, 32, 48, 50}
	actual := getCells(4, 4)
	if len(expected) != len(actual) {
		t.Error("Got wrong length")
	}
	for i, val := range actual {
		if expected[i] != index(val) {
			t.Errorf("Calculated wrong cells: expected %v\n got %v\n", expected, actual)
		}
	}
}

func BenchmarkSolve(b *testing.B) {
	s, _ := parseSudoku(benchmarkSudoku)
	for i := 0; i < b.N; i++ {
		s.Solve()
	}
}
