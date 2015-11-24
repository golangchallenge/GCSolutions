package sudoku

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"
)

type SudokuTest struct {
	puzzle   string
	solution string
}

var (
	unsolvable        = "000005080000601043000000000010500000000106000300000005530000061000000004000000000"
	invalidTests      = loadTestPuzzles("testdata/p_invalid.txt")
	generalTests      = loadTestPuzzles("testdata/p_general.txt")
	nakedSingleTests  = loadTestPuzzles("testdata/p_method_nakedsingle.txt")
	hiddenSingleTests = loadTestPuzzles("testdata/p_method_nakedsingle.txt")
	nakedPairTests    = loadTestPuzzles("testdata/p_method_nakedpair.txt")
	lockedTypeTests   = loadTestPuzzles("testdata/p_method_lockedtype.txt")

	easyTests   = loadTestPuzzles("testdata/p_level_easy.txt")
	mediumTests = loadTestPuzzles("testdata/p_level_medium.txt")
	hardTests   = loadTestPuzzles("testdata/p_level_hard.txt")
	evilTests   = loadTestPuzzles("testdata/p_level_evil.txt")
)

// ========================
// Test for Invalid Puzzles
// ========================
func TestInvalidPuzzle(t *testing.T) {
	for _, st := range invalidTests {
		_, err := NewSudoku(st.puzzle)
		if err == nil {
			t.Fatal("Accepted invalid puzzle:", st.puzzle)
		}
	}
}

func TestUnsolvablePuzzle(t *testing.T) {
	s, err := NewSudoku(unsolvable)
	if err != nil {
		t.Fatal(err)
	}
	if s.Solve(1) != 0 {
		t.Fatal("DLX should not solve unsolvable")
	}
}

// ========================
// Test for solving by DLX and human strategies
// ========================
func TestDLX_All(t *testing.T) {
	allTests := []SudokuTest{}
	for _, tests := range [][]SudokuTest{
		generalTests, nakedSingleTests, hiddenSingleTests, nakedPairTests,
	} {
		allTests = append(allTests, tests...)
	}
	for _, st := range allTests {
		s, err := NewSudoku(st.puzzle)
		if err != nil {
			t.Fatal(err)
		}
		if s.Solve(1) != 1 {
			t.Fatal("DLX cannot solve:", st.puzzle)
		}
	}
}

func humanTesting(t *testing.T, testCases []SudokuTest, strategies []strategy) {
	for _, st := range testCases {
		s, _ := NewSudoku(st.puzzle)
		s.SolveHuman()
		if s.Solution() != st.solution {
			t.Fatal("Human cannot solve:", st.puzzle)
		}
	}
}

func TestHuman_NakedSingle(t *testing.T) {
	humanTesting(t, nakedSingleTests, []strategy{stNakedSingle})
}

func TestHuman_HiddenSingle(t *testing.T) {
	humanTesting(t, hiddenSingleTests, []strategy{stNakedSingle, stHiddenSingle})
}

func TestHuman_NakedPair(t *testing.T) {
	humanTesting(t, nakedPairTests, []strategy{stNakedSingle, stHiddenSingle, stNakedPair})
}

func TestHuman_LockedType(t *testing.T) {
	humanTesting(t, lockedTypeTests, []strategy{stNakedSingle, stHiddenSingle, stLockedType})
}

// ========================
// Test for puzzle grading
// ========================
func TestGrading(t *testing.T) {
	allSt := []strategy{stNakedSingle, stHiddenSingle, stLockedType, stNakedPair}
	var testCases []SudokuTest
	var total, correct, harderByOneLevel, easierByOneLevel int
	var solved int

	for level := LevelEasy; level <= LevelEvil; level++ {
		switch level {
		case LevelEasy:
			testCases = easyTests
		case LevelMedium:
			testCases = mediumTests
		case LevelHard:
			testCases = hardTests
		case LevelEvil:
			testCases = evilTests
		}

		for _, st := range testCases {
			total++
			s, _ := NewSudoku(st.puzzle)
			stats := s.solveWith(allSt)
			if stats.EmptyEnd == 0 {
				solved++
			}
			gradedLevel := gradeDifficulty(stats)
			switch gradedLevel - level {
			case 0:
				correct++
			case 1:
				harderByOneLevel++
			case -1:
				easierByOneLevel++
			}
		}
	}
	fmt.Printf("Total: %d, Solvable: %d, Correct:%d, Off by one: %d\n",
		total, solved, correct, easierByOneLevel+harderByOneLevel)

	if (float32(correct) / float32(total)) < 0.8 {
		t.Fatal("Grader graded correctly less than 0.8 of sample puzzles.")
	}

	veryIncorrect := total - correct - easierByOneLevel - harderByOneLevel
	if (float32(veryIncorrect) / float32(total)) > 0.05 {
		t.Fatal("Grader graded very incorrectly over  0.05 of sample puzzles.")
	}
}

// ========================
// Test for puzzle generation can generate puzzles
// ========================
func TestGeneratePuzzle(t *testing.T) {
	levels := []int{LevelAny, LevelEasy, LevelMedium, LevelHard, LevelMedium}
	for _, level := range levels {
		_, err := GeneratePuzzle(level)
		if err != nil {
			t.Fatal("Failed to generate puzzle")
		}
	}

	_, err := GeneratePuzzle(5)
	if err == nil {
		t.Fatal("Puzzle should not be generated.")
	}
}

// ========================
// Benchmarks for DLX
// ========================
func dlxBenchmarking(b *testing.B, puzzle string) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s, _ := NewSudoku(puzzle)
		s.Solve(1)
	}
}

func BenchmarkDLX_Simple(b *testing.B) {
	dlxBenchmarking(b, generalTests[0].puzzle)
}

func BenchmarkDLX_Hard(b *testing.B) {
	dlxBenchmarking(b, generalTests[1].puzzle)
}

func BenchmarkDLX_NakedSingle(b *testing.B) {
	dlxBenchmarking(b, nakedSingleTests[0].puzzle)
}

func BenchmarkDLX_HiddenSingle(b *testing.B) {
	dlxBenchmarking(b, hiddenSingleTests[0].puzzle)
}

func BenchmarkDLX_NakedPair(b *testing.B) {
	dlxBenchmarking(b, nakedPairTests[0].puzzle)
}

func BenchmarkDLX_LockedType(b *testing.B) {
	dlxBenchmarking(b, lockedTypeTests[0].puzzle)
}

func BenchmarkDLX_Unsolvable(b *testing.B) {
	dlxBenchmarking(b, unsolvable)
}

func BenchmarkDLX_AllTop95(b *testing.B) {
	data, err := ioutil.ReadFile("testdata/top95.txt")
	if err != nil {
		log.Fatal(err)
	}
	puzzles := strings.Split(string(data), "\n")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, puzzle := range puzzles {
			s, _ := NewSudoku(puzzle)
			s.Solve(1)
		}
	}
}

// ========================
// Benchmarks for strategies
// ========================
func humanBenchmarking(b *testing.B, puzzle string, strategies []strategy) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s, _ := NewSudoku(puzzle)
		s.SolveHuman()
	}
}

func BenchmarkHuman_NakedSingle(b *testing.B) {
	humanBenchmarking(b, nakedSingleTests[0].puzzle,
		[]strategy{stNakedSingle})
}

func BenchmarkHuman_HiddenSingle(b *testing.B) {
	humanBenchmarking(b, hiddenSingleTests[0].puzzle,
		[]strategy{stHiddenSingle, stHiddenSingle})
}

func BenchmarkHuman_NakedPair(b *testing.B) {
	humanBenchmarking(b, nakedPairTests[0].puzzle,
		[]strategy{stHiddenSingle, stHiddenSingle, stNakedPair})
}

func BenchmarkHuman_LockedType(b *testing.B) {
	humanBenchmarking(b, lockedTypeTests[0].puzzle,
		[]strategy{stHiddenSingle, stHiddenSingle, stNakedPair})
}

func genPuzzleBenchmarking(b *testing.B, level int) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		GeneratePuzzle(level)
	}
}

// ========================
// Benchmarks for puzzle generation
// ========================
func BenchmarkGeneratePuzzle_Easy(b *testing.B) {
	genPuzzleBenchmarking(b, LevelEasy)
}

func BenchmarkGeneratePuzzle_Medium(b *testing.B) {
	genPuzzleBenchmarking(b, LevelMedium)
}

func BenchmarkGeneratePuzzle_Hard(b *testing.B) {
	genPuzzleBenchmarking(b, LevelHard)
}

func BenchmarkGeneratePuzzle_Evil(b *testing.B) {
	genPuzzleBenchmarking(b, LevelEvil)
}

// ========================
// Utility
// ========================
func loadTestPuzzles(filepath string) (testCases []SudokuTest) {
	file, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	line := 1
	var puzzle string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if line%2 == 1 {
			puzzle = scanner.Text()
		} else {
			solution := scanner.Text()
			testCases = append(testCases, SudokuTest{puzzle, solution})
		}
		line++
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return
}
