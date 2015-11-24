package dlx

import (
	"fmt"
	"sort"
	"strings"
	"testing"
)

var table = []struct {
	nColumns  int
	rows      [][]int
	solutions [][][]int
}{
	{ // 0
		nColumns: 4,
		rows: [][]int{
			[]int{2, 3},
		},
		solutions: nil,
	},
	{ // 1
		nColumns: 7,
		rows: [][]int{
			[]int{2, 4, 5},
			[]int{0, 3, 6},
			[]int{1, 2, 5},
			[]int{0, 3},
			[]int{1, 6},
			[]int{3, 4, 6},
		},
		solutions: [][][]int{
			[][]int{
				[]int{2, 4, 5},
				[]int{0, 3},
				[]int{1, 6},
			},
		},
	},
	{ // 2
		nColumns: 4,
		rows: [][]int{
			[]int{0, 1},
			[]int{0, 2},
			[]int{1, 2},
		},
		solutions: nil,
	},
	{ // 3
		nColumns: 4,
		rows: [][]int{
			[]int{0, 1, 2},
			[]int{0, 2},
			[]int{1},
			[]int{3},
		},
		solutions: [][][]int{
			[][]int{
				[]int{0, 1, 2},
				[]int{3},
			},
			[][]int{
				[]int{0, 2},
				[]int{1},
				[]int{3},
			},
		},
	},
}

func TestSolve(t *testing.T) {
	t.Parallel()

	for i, v := range table {
		t.Logf("Test-case %d", i)

		m := NewMatrix(v.nColumns)

		for _, r := range v.rows {
			m.AddRow(r)
		}

		var solutions [][][]int
		m.Solve(func(cs [][]int) bool {
			solutions = append(solutions, cs)
			return false
		})

		solutionsStr := toString(solutions)
		expectedSolutionsStr := toString(v.solutions)

		if solutionsStr != expectedSolutionsStr {
			t.Errorf("Failed test-case %d:\nExpected solutions:\n%s\nActual solutions:\n%s\n", i, expectedSolutionsStr, solutionsStr)
		}
	}
}

func toString(solutions [][][]int) string {
	var solutionsStrs []string
	for _, solutions := range solutions {
		var solutionStrs []string
		for _, row := range solutions {
			rowStr := fmt.Sprintf("  %v", row)
			solutionStrs = append(solutionStrs, rowStr)
		}
		sort.Strings(solutionStrs)
		solutionStr := "[\n" + strings.Join(solutionStrs, ",\n") + "\n]"
		solutionsStrs = append(solutionsStrs, solutionStr)
	}
	sort.Strings(solutionsStrs)
	solutionsStr := strings.Join(solutionsStrs, ",\n")
	return solutionsStr
}
