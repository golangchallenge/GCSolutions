package sudoku

import (
	"strings"
	"testing"
)

var (
	tests = []struct {
		puzzle, solution string
		err              error
		DifficultyLevel
	}{{"2 _ _ 8 _ 4 _ _ 6\n" +
		"_ _ 6 _ _ _ 5 _ _\n" +
		"_ 7 4 _ _ _ 9 2 _\n" +
		"3 _ _ _ 4 _ _ _ 7\n" +
		"_ _ _ 3 _ 5 _ _ _\n" +
		"4 _ _ _ 6 _ _ _ 9\n" +
		"_ 1 9 _ _ _ 7 4 _\n" +
		"_ _ 8 _ _ _ 2 _ _\n" +
		"5 _ _ 6 _ 8 _ _ 1",
		"2 5 3 8 9 4 1 7 6\n" +
			"1 9 6 2 3 7 5 8 4\n" +
			"8 7 4 1 5 6 9 2 3\n" +
			"3 8 1 9 4 2 6 5 7\n" +
			"9 6 7 3 8 5 4 1 2\n" +
			"4 2 5 7 6 1 8 3 9\n" +
			"6 1 9 5 2 3 7 4 8\n" +
			"7 3 8 4 1 9 2 6 5\n" +
			"5 4 2 6 7 8 3 9 1", nil, DLEasy,
	}, { // Easy sudoku generated with http://www.websudoku.com/
		"_ 8 6 _ 2 _ _ _ 4\n" +
			"5 4 _ _ _ 1 8 7 _\n" +
			"_ 3 _ 6 _ 8 5 _ 2\n" +
			"4 _ _ _ _ _ 7 _ _\n" +
			"_ 9 5 _ _ _ 4 2 _\n" +
			"_ _ 7 _ _ _ _ _ 9\n" +
			"9 _ 8 5 _ 6 _ 4 _\n" +
			"_ 5 3 2 _ _ _ 9 7\n" +
			"2 _ _ _ 3 _ 6 8 _",

		"1 8 6 7 2 5 9 3 4\n" +
			"5 4 2 3 9 1 8 7 6\n" +
			"7 3 9 6 4 8 5 1 2\n" +
			"4 2 1 9 5 3 7 6 8\n" +
			"3 9 5 8 6 7 4 2 1\n" +
			"8 6 7 4 1 2 3 5 9\n" +
			"9 1 8 5 7 6 2 4 3\n" +
			"6 5 3 2 8 4 1 9 7\n" +
			"2 7 4 1 3 9 6 8 5", nil, DLEasy,
	}, { // Easy sudoku generated with http://www.websudoku.com/
		"_ 5 _ _ 8 1 _ 4 _\n" +
			"_ 4 _ 9 5 7 2 _ _\n" +
			"3 _ _ _ 6 2 1 _ _\n" +
			"_ _ 1 _ 2 6 _ _ 7\n" +
			"_ 8 _ _ _ _ _ 3 _\n" +
			"2 _ _ 8 3 _ 5 _ _\n" +
			"_ _ 3 1 9 _ _ _ 6\n" +
			"_ _ 6 2 4 5 _ 8 _\n" +
			"_ 1 _ 6 7 _ _ 2 _",

		"6 5 2 3 8 1 7 4 9\n" +
			"1 4 8 9 5 7 2 6 3\n" +
			"3 7 9 4 6 2 1 5 8\n" +
			"4 3 1 5 2 6 8 9 7\n" +
			"9 8 5 7 1 4 6 3 2\n" +
			"2 6 7 8 3 9 5 1 4\n" +
			"5 2 3 1 9 8 4 7 6\n" +
			"7 9 6 2 4 5 3 8 1\n" +
			"8 1 4 6 7 3 9 2 5", nil, DLEasy,
	}, { // Medium level sudoku generated with http://www.websudoku.com/
		"_ _ 9 6 _ _ 2 _ _\n" +
			"_ _ _ 2 _ 1 _ 9 3\n" +
			"_ _ _ _ 8 _ _ 6 _\n" +
			"_ 1 _ 3 2 _ 6 _ 5\n" +
			"_ 4 _ 1 _ 5 _ 3 _\n" +
			"3 _ 5 _ 9 6 _ 1 _\n" +
			"_ 6 _ _ 1 _ _ _ _\n" +
			"8 2 _ 7 _ 4 _ _ _\n" +
			"_ _ 1 _ _ 2 4 _ _",

		"1 3 9 6 5 7 2 8 4\n" +
			"6 7 8 2 4 1 5 9 3\n" +
			"4 5 2 9 8 3 1 6 7\n" +
			"9 1 7 3 2 8 6 4 5\n" +
			"2 4 6 1 7 5 8 3 9\n" +
			"3 8 5 4 9 6 7 1 2\n" +
			"7 6 4 5 1 9 3 2 8\n" +
			"8 2 3 7 6 4 9 5 1\n" +
			"5 9 1 8 3 2 4 7 6", nil, DLMedium,
	}, { // Medium level sudoku generated with http://www.websudoku.com/
		"_ _ _ _ 5 _ _ _ _\n" +
			"_ 6 _ 3 _ 4 _ _ 7\n" +
			"1 7 _ _ _ _ 8 _ _\n" +
			"_ 2 _ _ _ 5 _ 1 6\n" +
			"_ 8 _ 6 1 2 _ 3 _\n" +
			"6 9 _ 4 _ _ _ 2 _\n" +
			"_ _ 6 _ _ _ _ 7 2\n" +
			"4 _ _ 7 _ 1 _ 9 _\n" +
			"_ _ _ _ 4 _ _ _ _",
		"8 4 3 1 5 7 2 6 9\n" +
			"2 6 9 3 8 4 1 5 7\n" +
			"1 7 5 2 6 9 8 4 3\n" +
			"3 2 4 8 9 5 7 1 6\n" +
			"5 8 7 6 1 2 9 3 4\n" +
			"6 9 1 4 7 3 5 2 8\n" +
			"9 1 6 5 3 8 4 7 2\n" +
			"4 3 8 7 2 1 6 9 5\n" +
			"7 5 2 9 4 6 3 8 1",
		nil, DLMedium,
	}, { // Hard level sudoku generated with http://www.websudoku.com/
		"_ _ _ 3 4 _ _ _ 2\n" +
			"_ _ 8 _ _ 6 _ _ 5\n" +
			"_ 3 _ _ _ 9 _ 1 _\n" +
			"_ _ _ 1 _ _ _ 3 _\n" +
			"_ 4 2 _ 9 _ 5 7 _\n" +
			"_ 6 _ _ _ 7 _ _ _\n" +
			"_ 2 _ 5 _ _ _ 8 _\n" +
			"8 _ _ 9 _ _ 1 _ _\n" +
			"1 _ _ _ 7 4 _ _ _",
		"6 1 7 3 4 5 8 9 2\n" +
			"2 9 8 7 1 6 3 4 5\n" +
			"4 3 5 2 8 9 6 1 7\n" +
			"7 8 9 1 5 2 4 3 6\n" +
			"3 4 2 6 9 8 5 7 1\n" +
			"5 6 1 4 3 7 9 2 8\n" +
			"9 2 4 5 6 1 7 8 3\n" +
			"8 7 6 9 2 3 1 5 4\n" +
			"1 5 3 8 7 4 2 6 9", nil, DLHard,
	}, { // Hard level sudoku generated with http://www.websudoku.com/
		"_ _ _ _ _ 5 1 2 _\n" +
			"_ _ 4 _ 1 _ 7 _ _\n" +
			"_ _ _ 2 _ _ _ _ 5\n" +
			"5 _ _ _ 4 _ _ 1 _\n" +
			"3 8 _ _ 9 _ _ 6 4\n" +
			"_ 9 _ _ 3 _ _ _ 2\n" +
			"1 _ _ _ _ 3 _ _ _\n" +
			"_ _ 5 _ 6 _ 4 _ _\n" +
			"_ 3 9 1 _ _ _ _ _",
		"9 6 8 4 7 5 1 2 3\n" +
			"2 5 4 3 1 6 7 9 8\n" +
			"7 1 3 2 8 9 6 4 5\n" +
			"5 7 2 6 4 8 3 1 9\n" +
			"3 8 1 7 9 2 5 6 4\n" +
			"4 9 6 5 3 1 8 7 2\n" +
			"1 4 7 8 2 3 9 5 6\n" +
			"8 2 5 9 6 7 4 3 1\n" +
			"6 3 9 1 5 4 2 8 7", nil, DLHard,
	}, {
		"_ _ _ 3 5 _ _ _ _\n" +
			"_ _ _ 1 _ _ _ 3 _\n" +
			"1 5 3 _ _ 8 4 _ _\n" +
			"2 1 5 _ _ _ _ 4 _\n" +
			"3 _ _ _ _ _ _ _ 1\n" +
			"_ 7 _ _ _ _ 3 5 2\n" +
			"_ _ 4 8 _ _ 7 1 9\n" +
			"_ 9 _ _ _ 4 _ _ _\n" +
			"_ _ _ _ 6 2 _ _ _",
		"4 6 2 3 5 9 1 7 8\n" +
			"9 8 7 1 4 6 2 3 5\n" +
			"1 5 3 2 7 8 4 9 6\n" +
			"2 1 5 6 8 3 9 4 7\n" +
			"3 4 9 5 2 7 8 6 1\n" +
			"8 7 6 4 9 1 3 5 2\n" +
			"6 2 4 8 3 5 7 1 9\n" +
			"5 9 8 7 1 4 6 2 3\n" +
			"7 3 1 9 6 2 5 8 4", nil, DLMedium,
	}, {
		"_ _ _ _ 3 7 6 _ _\n" +
			"_ _ _ 6 _ _ _ 9 _\n" +
			"_ _ 8 _ _ _ _ _ 4\n" +
			"_ 9 _ _ _ _ _ _ 1\n" +
			"6 _ _ _ _ _ _ _ 9\n" +
			"3 _ _ _ _ _ _ 4 _\n" +
			"7 _ _ _ _ _ 8 _ _\n" +
			"_ 1 _ _ _ 9 _ _ _\n" +
			"_ _ 2 5 4 _ _ _ _",
		"9 5 4 1 3 7 6 8 2\n" +
			"2 7 3 6 8 4 1 9 5\n" +
			"1 6 8 2 9 5 7 3 4\n" +
			"4 9 5 7 2 8 3 6 1\n" +
			"6 8 1 4 5 3 2 7 9\n" +
			"3 2 7 9 6 1 5 4 8\n" +
			"7 4 9 3 1 2 8 5 6\n" +
			"5 1 6 8 7 9 4 2 3\n" +
			"8 3 2 5 4 6 9 1 7", nil, DLHard,
	}, { // Puzzle with empty row (extremely hard)
		"_ _ _ _ _ _ _ _ _\n" +
			"_ _ _ _ _ 3 _ 8 5\n" +
			"_ _ 1 _ 2 _ _ _ _\n" +
			"_ _ _ 5 _ 7 _ _ _\n" +
			"_ _ 4 _ _ _ 1 _ _\n" +
			"_ 9 _ _ _ _ _ _ _\n" +
			"5 _ _ _ _ _ _ 7 3\n" +
			"_ _ 2 _ 1 _ _ _ _\n" +
			"_ _ _ _ 4 _ _ _ 9",
		"9 8 7 6 5 4 3 2 1\n" +
			"2 4 6 1 7 3 9 8 5\n" +
			"3 5 1 9 2 8 7 4 6\n" +
			"1 2 8 5 3 7 6 9 4\n" +
			"6 3 4 8 9 2 1 5 7\n" +
			"7 9 5 4 6 1 8 3 2\n" +
			"5 1 9 2 8 6 4 7 3\n" +
			"4 7 2 3 1 9 5 6 8\n" +
			"8 6 3 7 4 5 2 1 9", nil, DLHard,
	}, { // Puzzle with empty column (evil)
		"_ 7 4 _ _ _ _ 9 3\n" +
			"_ _ 5 9 _ _ _ _ 4\n" +
			"_ _ _ _ _ 2 _ _ _\n" +
			"_ 2 _ 1 _ _ _ _ 7\n" +
			"_ _ 8 _ _ _ 4 _ _\n" +
			"1 _ _ _ _ 3 _ 2 _\n" +
			"_ _ _ 6 _ _ _ _ _\n" +
			"9 _ _ _ _ 7 8 _ _\n" +
			"5 8 _ _ _ _ 1 6 _",
		"2 7 4 8 5 1 6 9 3\n" +
			"8 3 5 9 7 6 2 1 4\n" +
			"6 9 1 4 3 2 7 5 8\n" +
			"3 2 6 1 4 5 9 8 7\n" +
			"7 5 8 2 6 9 4 3 1\n" +
			"1 4 9 7 8 3 5 2 6\n" +
			"4 1 2 6 9 8 3 7 5\n" +
			"9 6 3 5 1 7 8 4 2\n" +
			"5 8 7 3 2 4 1 6 9", nil, DLHard,
	}, {
		"1 _ 3 _ _ 6 _ 8 _\n" +
			"_ 5 _ _ 8 _ 1 2 _\n" +
			"7 _ 9 1 _ 3 _ 5 6\n" +
			"_ 3 _ _ 6 7 _ 9 _\n" +
			"5 _ 7 8 _ _ _ 3 _\n" +
			"8 _ 1 _ 3 _ 5 _ 7\n" +
			"_ 4 _ _ 7 8 _ 1 _\n" +
			"6 _ 8 _ _ 2 _ 4 _\n" +
			"_ 1 2 _ 4 5 _ 7 8", "",
		ErrMultipleSolutions, DLUnknown,
	}, {
		"_ _ _ _ _ 1 _ 2 _\n" +
			"_ _ _ 4 _ _ _ 5 _\n" +
			"_ _ _ _ _ _ _ 8 _\n" +
			"_ 7 _ _ _ 9 _ _ _\n" +
			"_ _ 9 8 _ _ _ _ 5\n" +
			"_ _ _ _ _ 5 _ _ _\n" +
			"_ _ _ 6 _ _ 1 _ _\n" +
			"_ _ _ 2 _ _ _ _ _\n" +
			"_ _ _ 7 _ _ 5 _ _",
		"", ErrMultipleSolutions, DLUnknown,
	}, {
		"3 _ _ _ _ 4 _ _ _\n" +
			"8 _ _ _ 9 2 _ 6 _\n" +
			"_ _ 1 _ 8 _ _ _ 4\n" +
			"_ 3 _ _ _ _ _ 1 _\n" +
			"1 _ 7 _ _ _ 5 _ 9\n" +
			"_ 2 _ _ _ _ _ 4 _\n" +
			"5 _ _ _ 4 _ 1 _ _\n" +
			"_ 9 _ 7 1 _ _ _ 5\n" +
			"_ _ _ 5 _ _ _ _ 8",
		"3 7 9 6 5 4 2 8 1\n" +
			"8 5 4 1 9 2 7 6 3\n" +
			"2 6 1 3 8 7 9 5 4\n" +
			"9 3 8 4 7 5 6 1 2\n" +
			"1 4 7 8 2 6 5 3 9\n" +
			"6 2 5 9 3 1 8 4 7\n" +
			"5 8 3 2 4 9 1 7 6\n" +
			"4 9 6 7 1 8 3 2 5\n" +
			"7 1 2 5 6 3 4 9 8", nil, DLHard,
	}, {
		"2 8 6 1 5 9 7 4 3\n" +
			"3 5 7 6 4 8 2 1 9\n" +
			"4 1 9 7 _ _ 5 6 8\n" +
			"8 2 1 9 6 5 4 _ 7\n" +
			"6 9 3 8 7 4 1 _ 5\n" +
			"7 4 5 3 _ _ 8 _ 6\n" +
			"5 6 8 2 _ _ 9 7 4\n" +
			"1 3 4 5 9 7 6 8 2\n" +
			"9 7 2 4 8 6 3 5 1",
		"", ErrMultipleSolutions, DLUnknown,
	}}

	emptyPuzzle = func() *Puzzle {
		p, _ := NewPuzzleFromReader(strings.NewReader(
			"_ _ _ _ _ _ _ _ _\n" +
				"_ _ _ _ _ _ _ _ _\n" +
				"_ _ _ _ _ _ _ _ _\n" +
				"_ _ _ _ _ _ _ _ _\n" +
				"_ _ _ _ _ _ _ _ _\n" +
				"_ _ _ _ _ _ _ _ _\n" +
				"_ _ _ _ _ _ _ _ _\n" +
				"_ _ _ _ _ _ _ _ _\n" +
				"_ _ _ _ _ _ _ _ _"))
		return p

	}()
)

const (
	possible1 = 1 << iota
	possible2
	possible3
	possible4
	possible5
	possible6
	possible7
	possible8
	possible9
)

func TestBlockNum(t *testing.T) {
	puzzle, _ := NewPuzzleFromReader(strings.NewReader(tests[0].puzzle))

	r := &puzzleRow{}
	puzzle.blockAsRow(0, r)
	if string(r.cells[:]) != "2____6_74" {
		t.Fatal("Wrong block #0: ", string(r.cells[:]))
	}

	puzzle.blockAsRow(3, r)
	if string(r.cells[:]) != "3_____4__" {
		t.Fatal("Wrong block #3: ", string(r.cells[:]))
	}

	puzzle.blockAsRow(5, r)
	if string(r.cells[:]) != "__7_____9" {
		t.Fatal("Wrong block #5: ", string(r.cells[:]))
	}
}

func TestRowSuggestions(t *testing.T) {
	puzzle, _ := NewPuzzleFromReader(strings.NewReader(tests[0].puzzle))

	suggestion := puzzle.rowsAvail[0]
	if suggestion != possible1|possible3|possible5|possible7|
		possible9 {
		t.Fatal(suggestion)
	}
}

func TestPuzzleSolver(t *testing.T) {
	for _, test := range tests {
		r := strings.NewReader(test.puzzle)
		puzzle, err := NewPuzzleFromReader(r)
		if err != nil {
			t.Fatal("Unexpected error: ", err)
		}

		if puzzle.String() != test.puzzle {
			t.Fatal("Puzzle is invalid\n\tExpected:\n",
				test.puzzle,
				"\n\tAcquired:", "\n"+puzzle.String())
		}

		if puzzle.IsSolved() {
			t.Fatal("Puzzle should not be solved")
		}

		solution, err := puzzle.Solve()

		if err == nil && err == test.err {
			if !solution.IsSolved() {
				t.Fatal("Puzzle is not solved:\n", solution)
			}
			if solution.String() != test.solution {
				t.Fatal("Wrong solution found\n\tExpected:",
					"\n"+test.solution,
					"\n\tAquired:", "\n"+solution.String())
			}

			if *solution == *puzzle {
				t.Fatal("Solve() function should not modify puzzle")
			}
		} else if err != test.err {
			t.Log("\n" + puzzle.String())
			t.Log("\n" + solution.String())
			t.Fatal("Could not solve the puzzle: ", err)
		}

		if dif := solution.Difficulty(); dif != test.DifficultyLevel {
			t.Log("\n" + puzzle.String())
			t.Log("Given cells: ", 81-solution.emptyCells,
				"(", solution.emptyCells, ")",
				"\nLower bound of given rows, cols, blocks: ",
				solution.lbRow, ", ", solution.lbCol, ", ",
				solution.lbBlock,
				"\nLoops counter: ", solution.loopCount)
			t.Fatal("Difficulty level should be ",
				test.DifficultyLevel, " but it is ", dif)
		}
	}
}

func TestPopCountAndCTZ(t *testing.T) {
	tests := []struct {
		n       uint16
		pc, ctz byte
	}{
		{1, 1, 0},
		{2, 1, 1},
		{3, 2, 0},
		{4, 1, 2},
		{allPossible, 9, 0},
		{0x1F0, 5, 4},
		{0x034, 3, 2},
		{0x100, 1, 8},
	}

	for _, test := range tests {
		pc, ctz := popCount9Bit(test.n), ctz16Bits(test.n)
		if pc != test.pc {
			t.Error("Wrong pop count for", test.n, ":", pc)
		}
		if ctz != test.ctz {
			t.Error("Wrong ctz for", test.n, ":", ctz)
		}
	}
}

var benchPuzzle *Puzzle

func BenchmarkSolveSimpleEmpty(b *testing.B) {
	for i := 0; i < b.N; i++ {
		p := *emptyPuzzle
		p.simpleSolve()
		if p.err == nil || p.err == errSolutionNotFound {
			benchPuzzle = &p
		}
	}

	b.ReportAllocs()
}

func BenchmarkSolveSimpleSolved(b *testing.B) {
	b.StopTimer()
	puzzle, _ := NewPuzzleFromReader(strings.NewReader(
		tests[0].solution,
	))
	b.ResetTimer()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		p := *puzzle
		p.simpleSolve()
		if p.err == nil || p.err == errSolutionNotFound {
			benchPuzzle = &p
		}
	}

	b.ReportAllocs()
}

func BenchmarkSolveSimple(b *testing.B) {
	b.StopTimer()
	puzzle, _ := NewPuzzleFromReader(strings.NewReader(
		tests[0].puzzle,
	))
	b.ResetTimer()
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		p := *puzzle
		p.simpleSolve()
		if p.err == nil || p.err == errSolutionNotFound {
			benchPuzzle = &p
		}
	}

	b.ReportAllocs()
}

var benchPopCount byte

func BenchmarkPopCount(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for j := uint16(allPossible); j != 0; j-- {
			pc := popCount9Bit(j)
			if pc != 1 {
				benchPopCount = pc
			}
		}
	}
	b.ReportAllocs()
	b.SetBytes(2 * allPossible)
}

func BenchmarkEmptyPuzzleSolve(b *testing.B) {
	for i := 0; i < b.N; i++ {
		solution, err := emptyPuzzle.Solve()
		emptyPuzzle.err = nil
		if err == nil {
			benchPuzzle = solution
		}
	}
	b.ReportAllocs()
}

func BenchmarkExtraHardPuzzleSolve(b *testing.B) {
	b.StopTimer()
	puzzle, _ := NewPuzzleFromReader(strings.NewReader(
		"_ _ _ _ _ _ _ _ _\n" +
			"_ _ _ _ _ 3 _ 8 5\n" +
			"_ _ 1 _ 2 _ _ _ _\n" +
			"_ _ _ 5 _ 7 _ _ _\n" +
			"_ _ 4 _ _ _ 1 _ _\n" +
			"_ 9 _ _ _ _ _ _ _\n" +
			"5 _ _ _ _ _ _ 7 3\n" +
			"_ _ 2 _ 1 _ _ _ _\n" +
			"_ _ _ _ 4 _ _ _ 9",
	))
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		s, _ := puzzle.Solve()
		if s != nil {
			benchPuzzle = s
		}
	}
	b.ReportAllocs()
}
