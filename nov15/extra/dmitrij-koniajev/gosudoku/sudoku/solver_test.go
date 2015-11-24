package sudoku

import (
	"sort"
	"strings"
	"testing"
)

var solverTable = []struct {
	input     string
	solutions []string
}{
	{
		".......2143.......6........2.15..........637...........68...4.....23........7....",
		[]string{
			"857349621432861597619752843271583964945126378386497215768915432194238756523674189",
		},
	},
	{
		"85...24..72......9..4.........1.7..23.5...9...4...........8..7..17..........36.4.",
		[]string{
			"859612437723854169164379528986147352375268914241593786432981675617425893598736241",
		},
	},
	{
		"....74316...6.384......85..7258...34....3..5......2798..894.....4..859..971326485",
		[]string{
			"589274316217653849463198527725819634896437152134562798658941273342785961971326485",
		},
	},
	{
		".8...9743.5...8.1..1.......8....5......8.4......3....6.......7..3.5...8.9724...5.",
		[]string{
			"286159743457638219319742568821965437693874125745321896568213974134597682972486351",
			"286159743357648219419723568821965437693874125745312896568231974134597682972486351",
			"286159743357648219419732568821965437693874125745321896568213974134597682972486351",
			"286159743457638912319742568891265437623874195745391826568913274134527689972486351",
			"286159743357648912419732568891265437623874195745391826568913274134527689972486351",
			"286159743354768912719243568823615497697824135145397826568931274431572689972486351",
			"286159743354768912719243568893615427627894135145327896568931274431572689972486351",
			"286159743354768912719243568893615427627894135145372896568931274431527689972486351",
		},
	},
	{
		".....5.8....6.1.43..........1.5........1.6...3.......553.....61........4.........",
		nil,
	},
}

func TestSolve(t *testing.T) {
	t.Parallel()

	for _, tc := range solverTable {
		var solutions []string
		Solve(tc.input, func(solution string) bool {
			solutions = append(solutions, solution)
			return false
		})

		solutionsStr := toString(solutions)
		expectedSolutionsStr := toString(tc.solutions)

		if solutionsStr != expectedSolutionsStr {
			t.Errorf("Failed to solve:\n%s\nExpected solutions:\n%s\nActual solutions:\n%s\n", tc.input, expectedSolutionsStr, solutionsStr)
		}
	}
}

func BenchmarkSolveTestSudoku(b *testing.B) {
	for n := 0; n < b.N; n++ {
		for _, tc := range solverTable {
			Solve(tc.input, func(solution string) bool {
				return false
			})
		}
	}
}

func toString(solutions []string) string {
	sort.Strings(solutions)
	return strings.Join(solutions, ",\n")
}
