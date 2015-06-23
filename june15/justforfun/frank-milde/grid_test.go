//
// =========================================================================
//
//       Filename:  stack_test.go
//
//    Description:  Testing box stack.
//
//        License:  GNU General Public License
//      Copyright:  Copyright (c) 2015, Frank Milde
//
// =========================================================================
//

package main

import (
	"testing"
)

func Test_NewGrid(t *testing.T) {
	g := NewGrid()

	if g != nil {
		t.Errorf("Wrong new grid")
	}
}

func Test_NewInitialGrid(t *testing.T) {
	g := NewInitialGrid()

	if !(g[0].x == 0 &&
		g[0].y == 0 &&
		g[0].w == 4 &&
		g[0].l == 4 &&
		g[0].size == 16 &&
		g[0].orient == SQUAREGRID) {
		t.Errorf("Wrong initial grid")
	}
}

func Test_NewSubGrid(t *testing.T) {
	init := GridElement{0, 0, 0, 0, 0, 0}
	g := NewSubGrid(init)

	var want uint8 = 0

	if !(g[0].x == want &&
		g[0].y == want &&
		g[0].w == want &&
		g[0].l == want &&
		g[0].size == int(want) &&
		g[0].orient == HORIZONTAL) {
		t.Errorf("Non zero")
	}

	init2 := GridElement{1, 2, 2, 2, 4, SQUAREGRID}
	g2 := NewSubGrid(init2)

	if !(g2[0].x == 1 &&
		g2[0].y == 2 &&
		g2[0].w == 2 &&
		g2[0].l == 2 &&
		g2[0].size == 4 &&
		g2[0].orient == SQUAREGRID) {
		t.Errorf("Non zero")
	}

}

func Test_SetProperties(t *testing.T) {

	e := GridElement{1, 2, 2, 2, 0, 0}

	e.SetProperties()

	if !(e.size == 4 && e.orient == SQUAREGRID) {
		t.Errorf("Settings Wrong")
	}
}

func Test_IsEmpty(t *testing.T) {
	g := NewGrid()

	if !g.IsEmpty() {
		t.Errorf("Grid not empty")
	}
}

func Test_UpdateFreeGrid_ReplaceLastElement(t *testing.T) {
	init := FreeGrid{
		GridElement{3, 3, 1, 1, 1, SQUAREGRID},
		GridElement{1, 1, 3, 3, 9, SQUAREGRID},
	}
	newFreeGrid := FreeGrid{
		GridElement{1, 2, 1, 2, 2, VERTICAL},
		GridElement{2, 1, 2, 1, 2, HORIZONTAL},
		GridElement{2, 2, 2, 2, 4, SQUAREGRID},
	}

	init.Update(newFreeGrid)

	want := FreeGrid{
		GridElement{3, 3, 1, 1, 1, SQUAREGRID},
		GridElement{1, 2, 1, 2, 2, VERTICAL},
		GridElement{2, 1, 2, 1, 2, HORIZONTAL},
		GridElement{2, 2, 2, 2, 4, SQUAREGRID},
	}

	if !FreeGridsAreEqual(init, want) {
		t.Errorf("Grids are not equal:")
		t.Errorf("got: \n%v", init)
		t.Errorf("want: \n%v", want)
	}
}
func Test_UpdateFreeGrid_EmptyGrid(t *testing.T) {
	init := FreeGrid{
		GridElement{3, 3, 1, 1, 1, SQUAREGRID},
	}
	newFreeGrid := FreeGrid{}

	init.Update(newFreeGrid)

	want := FreeGrid{}

	if !FreeGridsAreEqual(init, want) {
		t.Errorf("Grids are not equal:")
		t.Errorf("got: \n%v", init)
		t.Errorf("want: \n%v", want)
	}
}

func Test_GridElementsAreEqual_InputAreGridElements(t *testing.T) {
	type inputs struct {
		a GridElement
		b GridElement
	}

	tests := []struct {
		in   inputs
		want bool
	}{
		// two emptybox
		{
			in: inputs{
				emptygrid,
				emptygrid,
			},
			want: true,
		},
		// equal grids
		{
			in: inputs{
				GridElement{0, 0, 2, 2, 4, SQUAREGRID},
				GridElement{0, 0, 2, 2, 4, SQUAREGRID},
			},
			want: true,
		},
		// different id
		{
			in: inputs{
				GridElement{0, 0, 1, 4, 4, VERTICAL},
				GridElement{0, 0, 2, 2, 4, SQUAREGRID},
			},
			want: false,
		},
		// different origin
		{
			in: inputs{
				GridElement{0, 0, 2, 2, 4, SQUAREGRID},
				GridElement{1, 2, 2, 2, 4, SQUAREGRID},
			},
			want: false,
		},
		// one emptybox
		{
			in: inputs{
				GridElement{1, 2, 2, 2, 4, SQUAREGRID},
				emptygrid,
			},
			want: false,
		},
	}

	for _, test := range tests {
		got := GridElementsAreEqual(test.in.a, test.in.b)
		if got != test.want {
			t.Errorf("Comparing GridElements: \n %v \n      == \n %v \n want %t, got %t", test.in.a, test.in.b, test.want, got)
		}
	}
} // -----  end of function Test_BoxesAreEqual_InputAreBoxes  -----
func Test_FreeGridsAreEqual(t *testing.T) {
	type inputs struct {
		a FreeGrid
		b FreeGrid
	}
	tests := []struct {
		in   inputs
		want bool
	}{
		// two equal FreeGrids
		{
			in: inputs{
				FreeGrid{
					GridElement{0, 0, 2, 2, 4, SQUAREGRID},
					GridElement{1, 2, 2, 2, 4, SQUAREGRID},
				},
				FreeGrid{
					GridElement{0, 0, 2, 2, 4, SQUAREGRID},
					GridElement{1, 2, 2, 2, 4, SQUAREGRID},
				},
			},
			want: true,
		},
		// two different FreeGrids
		{
			in: inputs{
				FreeGrid{
					GridElement{0, 0, 2, 2, 4, SQUAREGRID},
					GridElement{1, 2, 3, 2, 6, HORIZONTAL},
				},
				FreeGrid{
					GridElement{0, 0, 2, 2, 4, SQUAREGRID},
					GridElement{1, 2, 2, 2, 4, SQUAREGRID},
				},
			},
			want: false,
		},
		// different number of FreeGrids
		{
			in: inputs{
				FreeGrid{
					GridElement{0, 0, 2, 2, 4, SQUAREGRID},
					GridElement{1, 2, 3, 2, 6, HORIZONTAL},
				},
				FreeGrid{
					GridElement{0, 0, 2, 2, 4, SQUAREGRID},
				},
			},
			want: false,
		},
		// case: two empty FreeGrids
		{
			in: inputs{
				FreeGrid{
					GridElement{},
				},
				FreeGrid{
					GridElement{},
				},
			},
			want: true,
		},
	}

	for _, test := range tests {
		got := FreeGridsAreEqual(test.in.a, test.in.b)
		if got != test.want {
			t.Errorf("got: %t, want: %t", got, test.want)
			t.Errorf("a: \n%v", test.in.a)
			t.Errorf("b: \n%v", test.in.b)
		}
	}
} // -----  end of function Test_PalletsAreEqual  -----

/* Test not working since length and width had to be switched
func Test_Put_3x2on4x4_returnsTopRightTopright(t *testing.T) {
	//  | e e e e |
	//  | e e e e |
	//  | e e e e |
	//  | e e e e |
	e := GridElement{0, 0, 4, 4, 16, SQUAREGRID}

	//  | 1 1 3 3 |
	//  | b b 2 2 |
	//  | b b 2 2 |
	//  | b b 2 2 |
	b := box{0, 0, 2, 3, 100}

	g := Put(&b, e)

	want := FreeGrid{
		GridElement{3, 0, 1, 2, 2, HORIZONTAL},
		GridElement{3, 2, 1, 2, 2, HORIZONTAL},
		GridElement{0, 2, 3, 2, 6, VERTICAL},
	}

	if !FreeGridsAreEqual(g, want) {
		t.Errorf("Spliting wrong")
		t.Errorf("got:  \n%v", g)
		t.Errorf("want: \n%v", want)
	}
}
func Test_Put_3x2on3x2_gridIsFilled_returnsEmpty(t *testing.T) {
	//  |         |
	//  |         |
	//  | e e e   |
	//  | e e e   |
	e := GridElement{0, 0, 2, 3, 6, HORIZONTAL}

	//  |         |
	//  |         |
	//  | b b b   |
	//  | b b b   |
	b := box{0, 0, 2, 3, 100}

	g := Put(&b, e)

	want := FreeGrid{}

	if !FreeGridsAreEqual(g, want) {
		t.Errorf("Spliting wrong")
		t.Errorf("got:  \n%v", g)
		t.Errorf("want: \n%v", want)
	}
}
func Test_Put_3x2on4x2_returnsRight(t *testing.T) {
	//  | e e e e |
	//  | e e e e |
	//  |         |
	//  |         |
	e := GridElement{0, 2, 2, 4, 8, HORIZONTAL}

	//  | b b b 2 |
	//  | b b b 2 |
	//  |         |
	//  |         |
	b := box{0, 0, 2, 3, 101}

	g := Put(&b, e)

	want := FreeGrid{
		GridElement{3, 2, 1, 2, 2, HORIZONTAL},
	}

	if !FreeGridsAreEqual(g, want) {
		t.Errorf("Spliting wrong")
		t.Errorf("got:  \n%v", g)
		t.Errorf("want: \n%v", want)
	}
}
func Test_Put_4x1on4x3_returnsTop(t *testing.T) {
	//  | e e e e |
	//  | e e e e |
	//  | e e e e |
	//  |         |
	e := GridElement{0, 1, 3, 4, 12, HORIZONTAL}

	//  | 1 1 1 1 |
	//  | 1 1 1 1 |
	//  | b b b b |
	//  |         |
	b := box{0, 0, 1, 4, 100}

	g := Put(&b, e)

	want := FreeGrid{
		GridElement{0, 2, 4, 2, 8, VERTICAL},
	}

	if !FreeGridsAreEqual(g, want) {
		t.Errorf("Spliting wrong")
		t.Errorf("got:  \n%v", g)
		t.Errorf("want: \n%v", want)
	}
}
func Test_Put_1x1on3x3_returnsTopRightTopRight(t *testing.T) {
	//  |   e e e |
	//  |   e e e |
	//  |   e e e |
	//  |         |
	e := GridElement{1, 1, 3, 3, 9, SQUAREGRID}

	//  |   1 3 3 |
	//  |   1 3 3 |
	//  |   b 2 2 |
	//  |         |
	b := box{0, 0, 1, 1, 100}

	g := Put(&b, e)

	want := FreeGrid{
		GridElement{1, 2, 1, 2, 2, HORIZONTAL},
		GridElement{2, 1, 2, 1, 2, VERTICAL},
		GridElement{2, 2, 2, 2, 4, SQUAREGRID},
	}

	if !FreeGridsAreEqual(g, want) {
		t.Errorf("Spliting wrong")
		t.Errorf("got:  \n%v", g)
		t.Errorf("want: \n%v", want)
	}
}
*/
