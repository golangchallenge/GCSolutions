package sudoku

import (
	"testing"
)

func BenchmarkSolve(b *testing.B) {
	inkala_sudoku := SudokuField{
		8, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 3, 6, 0, 0, 0, 0, 0,
		0, 7, 0, 0, 9, 0, 2, 0, 0,
		0, 5, 0, 0, 0, 7, 0, 0, 0,
		0, 0, 0, 0, 4, 5, 7, 0, 0,
		0, 0, 0, 1, 0, 0, 0, 3, 0,
		0, 0, 1, 0, 0, 0, 0, 6, 8,
		0, 0, 8, 5, 0, 0, 0, 1, 0,
		0, 9, 0, 0, 0, 0, 4, 0, 0,
	}
	for n := 0; n < b.N; n++ {
		inkala_sudoku.Solve(0)
	}
}

func benchmarkSudokuNew(difficulty int, b *testing.B) {
	for n := 0; n < b.N; n++ {
		SudokuNew(difficulty)
	}
}

func BenchmarkSudokuNewEasy(b *testing.B)     { benchmarkSudokuNew(0, b) }
func BenchmarkSudokuNewMedium(b *testing.B)   { benchmarkSudokuNew(5, b) }
func BenchmarkSudokuNewHardcore(b *testing.B) { benchmarkSudokuNew(10, b) }

type SolveTest struct {
	sudoku       SudokuField
	solvedSudoku SudokuField
	solved       bool
}

var solveTests = []SolveTest{
	{SudokuField{
		8, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 3, 6, 0, 0, 0, 0, 0,
		0, 7, 0, 0, 9, 0, 2, 0, 0,
		0, 5, 0, 0, 0, 7, 0, 0, 0,
		0, 0, 0, 0, 4, 5, 7, 0, 0,
		0, 0, 0, 1, 0, 0, 0, 3, 0,
		0, 0, 1, 0, 0, 0, 0, 6, 8,
		0, 0, 8, 5, 0, 0, 0, 1, 0,
		0, 9, 0, 0, 0, 0, 4, 0, 0,
	},
		SudokuField{
			8, 1, 2, 7, 5, 3, 6, 4, 9,
			9, 4, 3, 6, 8, 2, 1, 7, 5,
			6, 7, 5, 4, 9, 1, 2, 8, 3,
			1, 5, 4, 2, 3, 7, 8, 9, 6,
			3, 6, 9, 8, 4, 5, 7, 2, 1,
			2, 8, 7, 1, 6, 9, 5, 3, 4,
			5, 2, 1, 9, 7, 4, 3, 6, 8,
			4, 3, 8, 5, 2, 6, 9, 1, 7,
			7, 9, 6, 3, 1, 8, 4, 5, 2,
		},
		true,
	},
	{SudokuField{
		8, 2, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 3, 6, 0, 0, 0, 0, 0,
		0, 7, 0, 0, 9, 0, 2, 0, 0,
		0, 5, 0, 0, 0, 7, 0, 0, 0,
		0, 0, 0, 0, 4, 5, 7, 0, 0,
		0, 0, 0, 1, 0, 0, 0, 3, 0,
		0, 0, 1, 0, 0, 0, 0, 6, 8,
		0, 0, 8, 5, 0, 0, 0, 1, 0,
		0, 9, 0, 0, 0, 0, 4, 0, 0,
	},
		SudokuField{
			8, 2, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 3, 6, 0, 0, 0, 0, 0,
			0, 7, 0, 0, 9, 0, 2, 0, 0,
			0, 5, 0, 0, 0, 7, 0, 0, 0,
			0, 0, 0, 0, 4, 5, 7, 0, 0,
			0, 0, 0, 1, 0, 0, 0, 3, 0,
			0, 0, 1, 0, 0, 0, 0, 6, 8,
			0, 0, 8, 5, 0, 0, 0, 1, 0,
			0, 9, 0, 0, 0, 0, 4, 0, 0,
		},
		false,
	},
}

func TestSolve(t *testing.T) {
	for _, test := range solveTests {
		sudoku := (test.sudoku).Clone()
		solved, _ := sudoku.Solve(0)
		if test.solved != solved {
			t.Errorf("%v\n sudoku has solution: %v", test.sudoku, test.solved)
		}
		if test.solved == true {
			sudoku = (test.sudoku).Clone()
			(sudoku).Solve(0)
			if sudoku != test.solvedSudoku {
				t.Errorf("%v\n sudoku has solution \n %v", test.sudoku, test.solvedSudoku)
			}
		}
	}
}

func TestNewSudoku(t *testing.T) {
	//new easy sudoku
	sudoku := SudokuNew(0)
	hasSolution, complexity := sudoku.Solve(0)
	if hasSolution == false {
		t.Errorf("new easy sudoku must have solution")
	} else if complexity > 0 {
		t.Errorf("easy sudoku must have compl=0")
	}
	//new hardcore sudoku
	sudoku = SudokuNew(10)
	hasSolution, _ = sudoku.Solve(0)
	if hasSolution == false {
		t.Errorf("new hard sudoku must have solution")
	}
}
