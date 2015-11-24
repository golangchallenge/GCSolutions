package sudoku_test

import (
	"bitbucket.org/jrozansk/go-challenge8/sudoku"
	"testing"
)

type testCase struct {
	input   string
	correct string
}

func TestSolvePositive(t *testing.T) {
	cases := positiveTestCases()
	for _, testCase := range cases {
		sut, _ := sudoku.InitSolver(testCase.input)
		if !sut.Solve() {
			t.Errorf("Sudoku didn't solved but solution was expected for %v", testCase.input)
		}
		result := sut.GetSolution()
		if result != testCase.correct {
			t.Errorf("Solver returned wrong value from Solve.\nExpected:\n %v\nGot:\n %v", testCase.correct, result)
		}
	}
}

func TestInitSolverBadInput(t *testing.T) {
	cases := negativeTestCases()
	for i, testCase := range cases {
		sut, _ := sudoku.InitSolver(testCase.input)
		if sut != nil {
			t.Errorf("InitSolver given corrupted grid %v (testCase %v) \nshould return nil", testCase.input, i)
		}
	}
}

func positiveTestCases() []testCase {
	return []testCase{
		testCase{
			input:   "103006080050080120709103056030067090507800030801030507040078010608002040012045078",
			correct: "123456789456789123789123456234567891567891234891234567345678912678912345912345678",
		},
		testCase{
			input:   "004800073010200000500003000020000007000030150081000209000950010090060500000400000",
			correct: "264815973913274865578693421325189647649732158781546239436958712192367584857421396",
		},
		testCase{
			input:   "805000030030900000406030000600010900050308070009040001000020308000009020070000504",
			correct: "815674239732951486496832715687215943154398672329746851941527368563489127278163594",
		},
		testCase{
			input:   "080009743050008010010000000800005000000804000000300006000000070030500080972400050",
			correct: "286159743354768912719243568823615497697824135145397826568931274431572689972486351",
		},
	}
}

func negativeTestCases() []testCase {
	return []testCase{
		testCase{input: "111"},
		testCase{input: "103006080050080120709103056030067090507800030801030507040078010608002040012045076"}, //the same values in one column
	}
}
