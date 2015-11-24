package dlx

import "testing"

// Dummy matrix for testing
// 1 1 0
// 0 0 1
// 0 1 1
func buildDummy() (*Node, [][]int) {
	root := NewRoot()
	c := make([]*Node, 3)
	for i := range c {
		c[i] = AddHeader(root)
	}
	matrix := [][]int{{1, 1, 0}, {0, 0, 1}, {0, 1, 1}}
	for i, r := range matrix {
		var row []*Node
		for j, v := range r {
			if v != 0 {
				node := AddNode(i, c[j])
				row = append(row, node)
			}
		}
		BuildRow(row)
	}
	return root, matrix
}

func TestMatrix(t *testing.T) {
	root, matrix := buildDummy()

	// Test Matrix()
	m := Matrix(root, 3, 3)
	for i, row := range m {
		for j := range row {
			if m[i][j] != matrix[i][j] {
				t.Fail()
			}
		}
	}
}

// Test covering on r1c1 of the matrix
// Only r2c3 should remain
func TestCover(t *testing.T) {
	root, matrix := buildDummy()
	newMatrix := [][]int{{0}, {1}, {0}}
	r1c1 := root.right.down

	// Test Cover()
	Cover(r1c1)
	// Count remaining headers
	cur := root.right
	cols := 0
	for cur != root {
		cur = cur.right
		cols++
	}
	if cols != 1 {
		t.Fatal("Cover did not remove appropriate number of columns")
	}
	m := Matrix(root, 3, 1)
	for i, row := range m {
		for j := range row {
			if m[i][j] != newMatrix[i][j] {
				t.Fatal("error in Cover function")
			}
		}
	}

	// Test Uncover()
	Uncover(r1c1)
	m = Matrix(root, 3, 3)
	for i, row := range m {
		for j := range row {
			if m[i][j] != matrix[i][j] {
				t.Fatal("error in Cover or Uncover function")
			}
		}
	}
}

func TestFind(t *testing.T) {
	root, _ := buildDummy()
	r2c3 := root.left.down
	if Find(r2c3.possibility, root) == nil {
		t.Fail()
	}
	if Find(10, root) != nil {
		t.Fail()
	}
}

func TestSmallestColumn(t *testing.T) {
	root, _ := buildDummy()
	col, size := smallestColumn(root)
	if col != root.right || size != 1 {
		t.Fail()
	}
}

func TestSolve(t *testing.T) {
	root, _ := buildDummy()
	sol := make(chan int, 10)
	solved := Solve(root, sol)
	if <-sol != 1 {
		t.Error("expexted solution not found")
	}
	if <-sol != 0 {
		t.Error("expexted solution not found")
	}
	if !solved {
		t.Error("solution not found")
	}
	if root.right != root || root.left != root {
		t.Error("matrix not actually emptied")
	}
}
