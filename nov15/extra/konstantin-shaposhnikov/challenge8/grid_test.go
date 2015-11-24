package main

import (
	"bytes"
	"strings"
	"testing"
)

func newGrid(data [GridSize][GridSize]byte) Grid {
	var g Grid
	for r := 0; r < GridSize; r++ {
		for c := 0; c < GridSize; c++ {
			if data[r][c] != 0 {
				g[r][c].resolve(data[r][c])
			}
		}
	}
	return g
}

func TestReadWrite(t *testing.T) {
	testInput := `1 _ 3 _ _ 6 _ 8 _
_ 5 _ _ 8 _ 1 2 _
7 _ 9 1 _ 3 _ 5 6
_ 3 _ _ 6 7 _ 9 _
5 _ 7 8 _ _ _ 3 _
8 _ 1 _ 3 _ 5 _ 7
_ 4 _ _ 7 8 _ 1 _
6 _ 8 _ _ 2 _ 4 _
_ 1 2 _ 4 5 _ 7 8
`

	testGrid := newGrid([GridSize][GridSize]byte{
		{1, 0, 3, 0, 0, 6, 0, 8, 0},
		{0, 5, 0, 0, 8, 0, 1, 2, 0},
		{7, 0, 9, 1, 0, 3, 0, 5, 6},
		{0, 3, 0, 0, 6, 7, 0, 9, 0},
		{5, 0, 7, 8, 0, 0, 0, 3, 0},
		{8, 0, 1, 0, 3, 0, 5, 0, 7},
		{0, 4, 0, 0, 7, 8, 0, 1, 0},
		{6, 0, 8, 0, 0, 2, 0, 4, 0},
		{0, 1, 2, 0, 4, 5, 0, 7, 8},
	})

	g, err := ReadGrid(strings.NewReader(testInput))
	if err != nil {
		t.Fatalf("ReadFrom: unexpected err: %s", err)
	}
	if g != testGrid {
		t.Errorf("grids do not match\n")
	}

	s := bytes.NewBuffer(nil)
	n, err := g.WriteTo(s)
	if err != nil {
		t.Fatalf("WriteTo: unexpected err: %s\n", err)
	}
	if n != int64(s.Len()) {
		t.Errorf("n: want %d, got %d\n", len(testInput), n)
	}
	if want := strings.TrimSpace(testInput); want != s.String() {
		t.Errorf("WriteTo: want\n%s\ngot\n%s\n", want, s.String())
	}
	if s.String() != g.String() {
		t.Errorf("String: want\n%s\ngot\n%s\n", s.String(), g.String())
	}
}

func TestReadFromInvalidInput(t *testing.T) {
	tests := []string{
		// not enough rows
		`1 _ 3 _ _ 6 _ 8 _
_ 5 _ _ 8 _ 1 2 _`,

		// wrong charcters
		`1 x 3 _ _ 6 _ 8 _
_ 5 _ _ 8 _ 1 2 _
7 _ 9 1 _ 3 _ 5 6
_ 3 _ _ 6 7 _ 9 _
5 _ 7 8 _ _ _ 3 _
8 _ 1 _ 3 _ 5 _ 7
_ 4 _ _ 7 8 _ 1 _
6 _ 8 _ _ 2 _ 4 _
_ 1 2 _ 4 5 _ 7 8
`,

		// wrong charcters
		`1 _x3 _ _ 6 _ 8 _
_ 5 _ _ 8 _ 1 2 _
7 _ 9 1 _ 3 _ 5 6
_ 3 _ _ 6 7 _ 9 _
5 _ 7 8 _ _ _ 3 _
8 _ 1 _ 3 _ 5 _ 7
_ 4 _ _ 7 8 _ 1 _
6 _ 8 _ _ 2 _ 4 _
_ 1 2 _ 4 5 _ 7 8
`,

		// line too long
		`1 _ 3 _ _ 6 _ 8 _ 1 2
_ 5 _ _ 8 _ 1 2 _
7 _ 9 1 _ 3 _ 5 6
_ 3 _ _ 6 7 _ 9 _
5 _ 7 8 _ _ _ 3 _
8 _ 1 _ 3 _ 5 _ 7
_ 4 _ _ 7 8 _ 1 _
6 _ 8 _ _ 2 _ 4 _
_ 1 2 _ 4 5 _ 7 8
`,

		// line too short
		`1 _ 3 _ _
_ 5 _ _ 8 _ 1 2 _
7 _ 9 1 _ 3 _ 5 6
_ 3 _ _ 6 7 _ 9 _
5 _ 7 8 _ _ _ 3 _
8 _ 1 _ 3 _ 5 _ 7
_ 4 _ _ 7 8 _ 1 _
6 _ 8 _ _ 2 _ 4 _
_ 1 2 _ 4 5 _ 7 8
`,
	}

	for _, test := range tests {
		_, err := ReadGrid(strings.NewReader(test))
		t.Logf("error %s for grid:\n%s\n", err, test)
		if err == nil {
			t.Errorf("error expected for \n%s\n", test)
		}
	}
}

func TestValidate(t *testing.T) {
	testGrid := newGrid([GridSize][GridSize]byte{
		{1, 0, 3, 0, 0, 6, 0, 8, 0},
		{0, 5, 0, 0, 8, 0, 1, 2, 0},
		{7, 0, 9, 1, 0, 3, 0, 5, 6},
		{0, 3, 0, 0, 6, 7, 0, 9, 0},
		{5, 0, 7, 8, 0, 0, 0, 3, 0},
		{8, 0, 1, 0, 3, 0, 5, 0, 7},
		{0, 4, 0, 0, 7, 8, 0, 1, 0},
		{6, 0, 8, 0, 0, 2, 0, 4, 0},
		{0, 1, 2, 0, 4, 5, 0, 7, 8},
	})
	if err := testGrid.Validate(); err != nil {
		t.Errorf("unexpected error '%s' for grid\n%s", err, testGrid)
	}

}

func TestValidateInvalid(t *testing.T) {
	tests := [][GridSize][GridSize]byte{
		// duplicate number 2 in row 2
		{
			{1, 0, 3, 0, 0, 6, 0, 8, 0},
			{2, 5, 0, 0, 8, 0, 1, 2, 0},
			{7, 0, 9, 1, 0, 3, 0, 5, 6},
			{0, 3, 0, 0, 6, 7, 0, 9, 0},
			{5, 0, 7, 8, 0, 0, 0, 3, 0},
			{8, 0, 1, 0, 3, 0, 5, 0, 7},
			{0, 4, 0, 0, 7, 8, 0, 1, 0},
			{6, 0, 8, 0, 0, 2, 0, 4, 0},
			{0, 1, 2, 0, 4, 5, 0, 7, 8},
		},
		// duplicate number 1 in column 4
		{
			{1, 0, 3, 0, 0, 6, 0, 8, 0},
			{0, 5, 0, 0, 8, 0, 1, 2, 0},
			{7, 0, 9, 1, 0, 3, 0, 5, 6},
			{0, 3, 0, 0, 6, 7, 0, 9, 0},
			{5, 0, 7, 8, 0, 0, 0, 3, 0},
			{8, 0, 1, 0, 3, 0, 5, 0, 7},
			{0, 4, 0, 0, 7, 8, 0, 1, 0},
			{6, 0, 8, 1, 0, 2, 0, 4, 0},
			{0, 1, 2, 0, 4, 5, 0, 7, 8},
		},
		// duplicate number 6 in region (2,2)
		{
			{1, 0, 3, 0, 0, 6, 0, 8, 0},
			{0, 5, 0, 0, 8, 0, 1, 2, 0},
			{7, 0, 9, 1, 0, 3, 0, 5, 6},
			{0, 3, 0, 0, 6, 7, 0, 9, 0},
			{5, 0, 7, 8, 0, 0, 0, 3, 0},
			{8, 0, 1, 6, 3, 0, 5, 0, 7},
			{0, 4, 0, 0, 7, 8, 0, 1, 0},
			{6, 0, 8, 0, 0, 2, 0, 4, 0},
			{0, 1, 2, 0, 4, 5, 0, 7, 8},
		},
	}

	for _, data := range tests {
		g := newGrid(data)
		err := g.Validate()
		if err == nil {
			t.Errorf("expected error for\n%s", g)
			continue
		}
		t.Logf("error '%s' for grid\n%s", err, g)
	}
}

func TestIsSolved(t *testing.T) {
	solvedGrid := newGrid([GridSize][GridSize]byte{
		{1, 2, 3, 4, 5, 6, 7, 8, 9},
		{4, 5, 6, 7, 8, 9, 1, 2, 3},
		{7, 8, 9, 1, 2, 3, 4, 5, 6},
		{2, 3, 4, 5, 6, 7, 8, 9, 1},
		{5, 6, 7, 8, 9, 1, 2, 3, 4},
		{8, 9, 1, 2, 3, 4, 5, 6, 7},
		{3, 4, 5, 6, 7, 8, 9, 1, 2},
		{6, 7, 8, 9, 1, 2, 3, 4, 5},
		{9, 1, 2, 3, 4, 5, 6, 7, 8},
	})
	if !solvedGrid.IsSolved() {
		t.Errorf("expected solved for grid\n%s", solvedGrid)
	}

	unsolvedGrid := newGrid([GridSize][GridSize]byte{
		{1, 2, 3, 4, 5, 6, 7, 8, 9},
		{4, 5, 6, 7, 8, 9, 1, 2, 3},
		{7, 8, 9, 1, 2, 3, 4, 5, 6},
		{2, 3, 4, 5, 6, 7, 8, 9, 1},
		{5, 6, 7, 8, 9, 1, 2, 3, 4},
		{8, 9, 1, 2, 3, 4, 5, 6, 7},
		{0, 4, 5, 6, 7, 8, 9, 1, 2},
		{6, 7, 8, 9, 1, 2, 3, 4, 5},
		{9, 1, 2, 3, 4, 5, 6, 7, 8},
	})
	if unsolvedGrid.IsSolved() {
		t.Errorf("expected unsolved for grid\n%s", unsolvedGrid)
	}
}

func TestCell(t *testing.T) {
	s := Cell{}
	if s.empty() {
		t.Errorf("zero cell should not be empty")
	}
	if s.resolved() {
		t.Errorf("zero cell should not be resolved")
	}
	if v := s.value(); v != 0 {
		t.Errorf("zero cell value: got %d, want 0", v)
	}
	if s.containsOnly(1, 5) {
		t.Errorf("zero cell shouldn't contain only 1 and 5")
	}

	for d := byte(1); d <= GridSize; d++ {
		if !s.contains(d) {
			t.Errorf("zero cell should contain %d", d)
		}
	}

	// remove all values except 5
	for d := byte(1); d <= GridSize; d++ {
		if d != 5 {
			s.remove(d)
			if s.contains(d) {
				t.Errorf("'%s' should not contain %d", s, d)

			}
		}
	}

	if s.empty() {
		t.Errorf("'%s' should not be empty", s)
	}
	if !s.resolved() {
		t.Errorf("'%s' should be resolved", s)
	}
	if v := s.value(); v != 5 {
		t.Errorf("'%s' value: got %d, want 5", s, v)
	}
	if s.containsOnly(1, 5) {
		t.Errorf("resolved cell shouldn't contain only 1 and 5")
	}

	resolved5 := Cell{}
	resolved5.resolve(5)
	if s != resolved5 {
		t.Errorf("resolve(5) value: got %s, want '%s'", resolved5, s)
	}

	s.set([]byte{1, 5})
	if !s.containsOnly(1, 5) {
		t.Errorf("'%s' contains only 1 and 5", s)
	}
}
