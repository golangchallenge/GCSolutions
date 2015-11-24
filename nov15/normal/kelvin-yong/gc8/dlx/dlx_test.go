package dlx

import (
	"fmt"
	"sort"
	"testing"
)

// Node struct will be useful for us to check if a map of []int, A,
// is a sub-map of another map of [], B.
// B represents all known solutions, while A would be a subset of solutions
type node struct {
	v int
	c map[int]node
}

// addPath assumes that the path to be added is sorted
func (n *node) addPath(path []int) {
	if len(path) == 0 {
		return
	}
	v := path[0]
	child, ok := n.c[v]
	if !ok {
		child = node{v, make(map[int]node)}
		n.c[v] = child
	}
	child.addPath(path[1:])
}

func (n *node) pathIsComplete(path []int) bool {
	if len(path) == 0 {
		return len(n.c) == 0
	}
	v := path[0]
	child, ok := n.c[v]
	if !ok {
		return false
	}

	return child.pathIsComplete(path[1:])
}

func (n *node) print(ancestors []int) {
	if len(n.c) == 0 {
		fmt.Println(append(ancestors, n.v))
		return
	}

	for _, childNode := range n.c {
		path := append(ancestors, n.v)
		childNode.print(path)
	}
}

// TestPaths tests if the node structure is working as expected
func TestPaths(t *testing.T) {
	paths := [][]int{
		{0, 3, 4},
		{0, 3, 10},
		{0, 4, 9},
		{0, 9, 10},
		{3, 4, 6},
		{3, 6, 10},
		{4, 6, 9},
		{4, 9, 10},
	}
	root := node{-1, make(map[int]node)}
	for _, p := range paths {
		root.addPath(p)
	}
	if testing.Verbose() {
		fmt.Println("All paths")
		root.print([]int{})
	}

	for _, p := range paths {
		if !root.pathIsComplete(p) {
			t.Fatal("Path", p, "should be a complete path")
		}
		if root.pathIsComplete(append(p, 1000)) || root.pathIsComplete(p[:len(p)-1]) {
			t.Fatal("Invalid path wrongly identified as complete")
		}
	}
}

// Exact Cover problem - solving using DLX
//
// Given a matrix of 0s and 1s, does it have a set of rows containing
// exactly one 1 in each column?
// For example, the matrix with 7 columns (or constraints) shown
//             0 1 2 3 4 5 6
//             =============
//      row 0: 0 0 1 0 1 1 0   ==> {2, 4, 5}
//      row 1: 1 0 0 1 0 0 1
//      row 2: 0 1 1 0 0 1 0
//      row 3: 1 0 0 1 0 0 0
//      row 4: 0 1 0 0 0 0 1
//      row 5: 0 0 0 1 1 0 1
// has such a set (rows 0, 3, 4)

// matrix of 0 and 1s
var matrix = [][]int{
	{0, 0, 1, 0, 1, 1, 0},
	{1, 0, 0, 1, 0, 0, 1},
	{0, 1, 1, 0, 0, 1, 0},
	{1, 0, 0, 1, 0, 0, 0},
	{0, 1, 0, 0, 0, 0, 1},
	{0, 0, 0, 1, 1, 0, 1},
}

// rows will store a row of 0s and 1s as column positions
// eg {0, 0, 1, 0, 1, 1, 0} ==> {2, 4, 5}
var rows [][]int

func convertRows() {
	rows = make([][]int, len(matrix))
	for i, mr := range matrix {
		row := []int{}
		for j, v := range mr {
			if v == 1 {
				row = append(row, j)
			}
		}
		rows[i] = row
	}
}

// test cases
var coverTests = []struct {
	rows      []int
	solutions [][]int
}{
	{
		// no solutions
		[]int{0, 1, 2, 3, 5, 5},
		[][]int{},
	},
	{
		// only 1 solution
		[]int{0, 1, 2, 3, 4, 5},
		[][]int{
			{0, 3, 4},
		},
	},
	{
		// mulitple solutions
		[]int{0, 1, 2, 3, 4, 5, 0, 1, 2, 3},
		[][]int{
			{0, 3, 4},
			{6, 3, 4},
			{0, 9, 4},
			{6, 9, 4},
		},
	},
}

func init() {
	convertRows()
}

func TestExactCover(t *testing.T) {
	isRandom := false

	for _, test := range coverTests {
		root := node{-1, make(map[int]node)}
		for _, sol := range test.solutions {
			sort.Ints(sol)
			root.addPath(sol)
		}

		for _, times := range []int{0, 1, 2, 4, 10} {
			dlx := NewDLX(len(matrix[0]))
			for i, r := range test.rows {
				dlx.AddRow(i, rows[r])
			}
			// add dummy row. Should be nothing
			dlx.AddRow(1000, []int{})

			dlx.Solve(times, isRandom)
			isRandom = !isRandom

			//check the number of solutions is valid
			var expectedNumber int
			if times >= len(test.solutions) {
				expectedNumber = len(test.solutions)
			} else {
				expectedNumber = times
			}
			if len(dlx.Solutions) != expectedNumber {
				t.Fatal("Did not find the expected number of solutions")
			}

			// check the found solutions is a valid one in the solution space
			for _, sol := range dlx.Solutions {
				sort.Ints(sol)
				if !root.pathIsComplete(sol) {
					t.Fatal(sol, "is not valid in the solution set")
				}
			}
		}
	}
}
