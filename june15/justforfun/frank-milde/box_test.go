//
// =========================================================================
//
//       Filename:  box_test.go
//
//    Description:  Unit test for the box.go file. Test should have the form
//                  Test_unitOfWork_scenario_expectedBehavior()
//
//           TODO:  Transform all tests to the form above.
//
//        License:  GNU General Public License
//      Copyright:  Copyright (c) 2015, Frank Milde
//
// =========================================================================
//

package main

import (
	"errors"
	"sort"
	"testing"
)

func Test_ValidCoordinates_inputIsOn4x4Grid_returnTrue(t *testing.T) {
	type inputs struct {
		x uint8
		y uint8
	}

	tests := []struct {
		in   inputs
		want bool
	}{
		{
			in:   inputs{0, 0},
			want: true,
		},
		{
			in:   inputs{3, 3},
			want: true,
		},
	}
	for _, test := range tests {
		got := ValidCoordinates(test.in.x, test.in.y)
		if got != test.want {
			t.Errorf("ValidCoordinates(%v,%v) == %t, want %t", test.in.x, test.in.y, got, test.want)
		}
	}
} // -----  end of function Test_ValidCoordinates_inputIsOn4x4Grid_returnTrue  -----
func Test_ValidCoordinates_inputIsNOTon4x4Grid_returnFalse(t *testing.T) {
	type inputs struct {
		x uint8
		y uint8
	}

	tests := []struct {
		in   inputs
		want bool
	}{
		// x out of bound
		{
			in:   inputs{4, 0},
			want: false,
		},
		// y out of bound
		{
			in:   inputs{1, 4},
			want: false,
		},
		// x,y out of bound
		{
			in:   inputs{4, 4},
			want: false,
		},
	}
	for _, test := range tests {
		got := ValidCoordinates(test.in.x, test.in.y)
		if got != test.want {
			t.Errorf("ValidCoordinates(%v,%v) == %t, want %t", test.in.x, test.in.y, got, test.want)
		}
	}
} // -----  end of function Test_ValidCoordinates_inputIsOn4x4Grid_returnTrue  -----

func Test_Size_normalInput_returnSizeOfBox(t *testing.T) {
	tests := []struct {
		in   *box
		want uint8
	}{
		{
			in:   &box{0, 0, 1, 1, 101},
			want: 1,
		},
		{
			in:   &box{3, 0, 4, 1, 104},
			want: 4,
		},
	}

	for _, test := range tests {
		got := test.in.Size()
		if got != test.want {
			t.Errorf("(%v).Size() == %d, want %d", test.in, got, test.want)
		}
	}
} // -----  end of function Test_Size_normalInput_returnSizeOfBox  -----
func Test_Size_emptyBoxInput_returnZero(t *testing.T) {
	tests := []struct {
		in   *box
		want uint8
	}{
		{
			in:   &box{},
			want: 0,
		},
		{
			in:   &emptybox,
			want: 0,
		},
	}

	for _, test := range tests {
		got := test.in.Size()
		if got != test.want {
			t.Errorf("(%v).Size() == %d, want %d", test.in, got, test.want)
		}
	}
} // -----  end of function Test_Size_emptyBoxInput_returnZero  -----

func Test_HasValidDimensions_validBoxInput_returnTrue(t *testing.T) {
	tests := []struct {
		in   *box
		want bool
	}{
		{
			in:   &box{0, 0, 1, 1, 101},
			want: true,
		},
		{
			in:   &box{0, 0, 3, 1, 103},
			want: true,
		},
		{
			in:   &box{1, 2, 4, 4, 103},
			want: true,
		},
	}

	for _, test := range tests {
		got := test.in.HasValidDimensions()
		if got != test.want {
			t.Errorf("(%v).HasValidDimensions() == %t, want %t", test.in, got, test.want)
		}
	}
} // -----  end of function Test_HasValidDimensions  -----
func Test_HasValidDimensions_boxIsToBig_returnFalse(t *testing.T) {
	tests := []struct {
		in   *box
		want bool
	}{
		// to big in first input
		{
			in:   &box{3, 0, 4, 6, 104},
			want: false,
		},
		// to big in second input
		{
			in:   &box{3, 3, 8, 1, 104},
			want: false,
		},
		// to big in both inputs
		{
			in:   &box{3, 3, 5, 5, 104},
			want: false,
		},
	}

	for _, test := range tests {
		got := test.in.HasValidDimensions()
		if got != test.want {
			t.Errorf("(%v).HasValidDimensions() == %t, want %t", test.in, got, test.want)
		}
	}
} // -----  end of function Test_HasValidDimensions  -----
func Test_HasValidDimensions_boxHasZeroLengthOrWidth_returnFalse(t *testing.T) {
	tests := []struct {
		in   *box
		want bool
	}{
		// to big in first input
		{
			in:   &box{3, 1, 0, 3, 104},
			want: false,
		},
		// to big in second input
		{
			in:   &box{3, 3, 2, 0, 104},
			want: false,
		},
		// to big in both inputs
		{
			in:   &box{3, 2, 0, 0, 104},
			want: false,
		},
	}

	for _, test := range tests {
		got := test.in.HasValidDimensions()
		if got != test.want {
			t.Errorf("(%v).HasValidDimensions() == %t, want %t", test.in, got, test.want)
		}
	}
} // -----  end of function Test_HasValidDimensions  -----
func Test_HasValidDimensions_boxSizeIsValidButCoordinateAreOutOfBound_returnTrue(t *testing.T) {
	tests := []struct {
		in   *box
		want bool
	}{
		{
			in:   &box{2, 7, 2, 2, 104},
			want: true,
		},
		{
			in:   &box{7, 1, 2, 2, 104},
			want: true,
		},
		{
			in:   &box{7, 7, 2, 2, 104},
			want: true,
		},
	}

	for _, test := range tests {
		got := test.in.HasValidDimensions()
		if got != test.want {
			t.Errorf("(%v).HasValidDimensions() == %t, want %t", test.in, got, test.want)
		}
	}
} // -----  end of function Test_HasValidDimensions  -----
func Test_HasValidDimensions_emptyBoxInput_returnFalse(t *testing.T) {
	tests := []struct {
		in   *box
		want bool
	}{
		{
			in:   &emptybox,
			want: false,
		},
		{
			in:   &box{},
			want: false,
		},
	}

	for _, test := range tests {
		got := test.in.HasValidDimensions()
		if got != test.want {
			t.Errorf("(%v).HasValidDimensions() == %t, want %t", test.in, got, test.want)
		}
	}
} // -----  end of function Test_HasValidDimensions  -----

func Test_HasValidCoordinates(t *testing.T) {
	tests := []struct {
		in   *box
		want bool
	}{
		// valid box
		{
			in:   &box{0, 0, 1, 1, 101},
			want: true,
		},
		{
			in:   &box{3, 3, 1, 1, 101},
			want: true,
		},
		// too far in x
		{
			in:   &box{4, 0, 4, 6, 104},
			want: false,
		},
		// too far in y
		{
			in:   &box{2, 4, 4, 6, 104},
			want: false,
		},
		// too far in x and y
		{
			in:   &box{4, 4, 4, 6, 104},
			want: false,
		},
		// empty box
		{
			in:   &emptybox,
			want: true,
		},
	}

	for _, test := range tests {
		got := test.in.HasValidCoordinates()
		if got != test.want {
			t.Errorf("(%v).HasValidCoordinates() == %t, want %t", test.in, got, test.want)
		}
	}
} // -----  end of function Test_HasValidCoordinates  -----

func Test_BoxesAreEqual_InputAreBoxes(t *testing.T) {
	type inputs struct {
		a box
		b box
	}

	tests := []struct {
		in   inputs
		want bool
	}{
		// two emptybox
		{
			in: inputs{
				emptybox,
				emptybox,
			},
			want: true,
		},
		// equal boxes
		{
			in: inputs{
				box{0, 0, 1, 1, 101},
				box{0, 0, 1, 1, 101},
			},
			want: true,
		},
		// different id
		{
			in: inputs{
				box{0, 0, 1, 1, 102},
				box{0, 0, 1, 1, 101},
			},
			want: false,
		},
		// different origin
		{
			in: inputs{
				box{1, 0, 1, 1, 101},
				box{0, 0, 1, 1, 101},
			},
			want: false,
		},
		// one emptybox
		{
			in: inputs{
				box{1, 0, 1, 1, 101},
				emptybox,
			},
			want: false,
		},
	}

	for _, test := range tests {
		got := BoxesAreEqual(test.in.a, test.in.b)
		if got != test.want {
			t.Errorf("Comparing boxes: \n %v \n      == \n %v \n want %t, got %t", test.in.a, test.in.b, test.want, got)
		}
	}
} // -----  end of function Test_BoxesAreEqual_InputAreBoxes  -----
func Test_BoxesAreEqual_InputAreBoxPointers(t *testing.T) {
	type inputs struct {
		a *box
		b *box
	}
	tests := []struct {
		in   inputs
		want bool
	}{
		// equal boxes
		{
			in: inputs{
				&box{0, 0, 1, 1, 101},
				&box{0, 0, 1, 1, 101},
			},
			want: true,
		},
		// different id
		{
			in: inputs{
				&box{0, 0, 1, 1, 102},
				&box{0, 0, 1, 1, 101},
			},
			want: false,
		},
		// different origin
		{
			in: inputs{
				&box{1, 0, 1, 1, 101},
				&box{0, 0, 1, 1, 101},
			},
			want: false,
		},
		// emptybox
		{
			in: inputs{
				&box{1, 0, 1, 1, 101},
				&emptybox,
			},
			want: false,
		},
		// invalid pointers
		//	{
		//		in: inputs{
		//			nil,
		//			nil,
		//		},
		//		want: true,
		//	},
	}

	for _, test := range tests {
		got := BoxesAreEqual(*test.in.a, *test.in.b)
		if got != test.want {
			t.Errorf("Comparing boxes:")
			t.Errorf("%v", test.in.a)
			t.Errorf("    == ")
			t.Errorf("%v", test.in.b)
			t.Errorf("want %t", test.want)
			t.Errorf("got  %t", got)
		}
	}
} // -----  end of function Test_BoxesAreEqual_InputAreBoxPointers  -----

func Test_PalletsAreEqual(t *testing.T) {
	type inputs struct {
		a pallet
		b pallet
	}
	tests := []struct {
		in   inputs
		want bool
	}{
		// two equal pallets
		{
			in: inputs{
				pallet{
					[]box{
						box{0, 0, 1, 1, 101},
						box{0, 0, 1, 1, 101},
					},
				},
				pallet{
					[]box{
						box{0, 0, 1, 1, 101},
						box{0, 0, 1, 1, 101},
					},
				},
			},
			want: true,
		},
		// two different pallets
		{
			in: inputs{
				pallet{
					[]box{
						box{0, 0, 1, 1, 101},
						box{0, 0, 1, 1, 101},
					},
				},
				pallet{
					[]box{
						box{0, 0, 1, 1, 101},
						box{1, 0, 1, 1, 101},
					},
				},
			},
			want: false,
		},
		// different number of pallets
		{
			in: inputs{
				pallet{
					[]box{
						box{0, 0, 1, 1, 101},
					},
				},
				pallet{
					[]box{
						box{0, 0, 1, 1, 101},
						box{1, 0, 1, 1, 101},
					},
				},
			},
			want: false,
		},
		// case: two empty pallets
		{
			in: inputs{
				pallet{
					[]box{},
				},
				pallet{
					[]box{},
				},
			},
			want: true,
		},
	}

	for _, test := range tests {
		got := PalletsAreEqual(test.in.a, test.in.b)
		if got != test.want {
			t.Errorf("Comparing pallets \n %v \n            ==\n %v\n want %t, got %t", test.in.a.boxes, test.in.b.boxes, test.want, got)
		}
	}
} // -----  end of function Test_PalletsAreEqual  -----

func Test_BoxArraysAreEqual(t *testing.T) {
	type inputs struct {
		a []box
		b []box
	}
	tests := []struct {
		in   inputs
		want bool
	}{
		// two equal pallets
		{
			in: inputs{
				[]box{
					box{0, 0, 1, 1, 101},
					box{0, 0, 1, 1, 101},
				},
				[]box{
					box{0, 0, 1, 1, 101},
					box{0, 0, 1, 1, 101},
				},
			},
			want: true,
		},
		// two different pallets
		{
			in: inputs{
				[]box{
					box{0, 0, 1, 1, 101},
					box{0, 0, 1, 1, 101},
				},
				[]box{
					box{0, 0, 1, 1, 101},
					box{1, 0, 1, 1, 101},
				},
			},
			want: false,
		},
		// different number of pallets
		{
			in: inputs{
				[]box{
					box{0, 0, 1, 1, 101},
				},
				[]box{
					box{0, 0, 1, 1, 101},
					box{1, 0, 1, 1, 101},
				},
			},
			want: false,
		},
		// case: two empty pallets
		{
			in: inputs{
				[]box{},
				[]box{},
			},
			want: true,
		},
	}

	for _, test := range tests {
		got := BoxArraysAreEqual(test.in.a, test.in.b)
		if got != test.want {
			t.Errorf("Comparing boxlist \n %v \n            ==\n %v\n want %t, got %t", test.in.a, test.in.b, test.want, got)
		}
	}
} // -----  end of function Test_BoxArraysAreEqual  -----

func Test_AddToPallet(t *testing.T) {
	type inputs struct {
		b box
		p *pallet
	}

	tests := []struct {
		in   inputs
		want pallet
	}{
		// empty box on empty pallet
		{
			in: inputs{
				emptybox,
				&pallet{},
			},
			want: pallet{},
		},
		// box on empty pallet
		{
			in: inputs{
				box{0, 0, 1, 1, 100},
				&pallet{},
			},
			want: pallet{
				[]box{
					box{0, 0, 1, 1, 100},
				},
			},
		},
		// box on filled pallet
		{
			in: inputs{
				box{1, 1, 1, 1, 101},
				&pallet{
					[]box{
						box{0, 0, 1, 1, 100},
					},
				},
			},
			want: pallet{
				[]box{
					box{0, 0, 1, 1, 100},
					box{1, 1, 1, 1, 101},
				},
			},
		},
		{
			in: inputs{
				box{0, 1, 1, 1, 101},
				&pallet{
					[]box{
						box{0, 0, 4, 1, 100},
					},
				},
			},
			want: pallet{
				[]box{
					box{0, 0, 4, 1, 100},
					box{0, 1, 1, 1, 101},
				},
			},
		},
		// box with invalid coordinates on filled pallet
		{
			in: inputs{
				box{4, 5, 1, 1, 100},
				&pallet{
					[]box{
						box{0, 0, 1, 1, 100},
					},
				},
			},
			want: pallet{
				[]box{
					box{0, 0, 1, 1, 100},
				},
			},
		},
	}

	for _, test := range tests {
		test.in.b.AddToPallet(test.in.p)
		if !PalletsAreEqual(*test.in.p, test.want) {
			t.Errorf("Comparing pallets \n   %v \n!=\n   %v", test.in.p.boxes, test.want.boxes)
		}
	}
} // -----  end of function Test_AddToPallet  -----

func Test_Sort(t *testing.T) {
	tests := []struct {
		in   []box
		want []box
	}{
		{
			// all boxes at 0,0
			in: []box{
				box{0, 0, 4, 4, 101},
				box{0, 0, 2, 2, 102},
				box{0, 0, 2, 1, 103},
				box{0, 0, 3, 2, 104},
			},
			want: []box{
				box{0, 0, 2, 1, 103},
				box{0, 0, 2, 2, 102},
				box{0, 0, 3, 2, 104},
				box{0, 0, 4, 4, 101},
			},
		},
		{
			// boxes at different coordinates
			in: []box{
				box{0, 0, 4, 4, 101},
				box{1, 1, 2, 2, 102},
				box{3, 1, 2, 1, 103},
				box{0, 2, 3, 2, 104},
			},
			want: []box{
				box{3, 1, 2, 1, 103},
				box{1, 1, 2, 2, 102},
				box{0, 2, 3, 2, 104},
				box{0, 0, 4, 4, 101},
			},
		},
		{
			// two equivalent boxes
			in: []box{
				box{0, 0, 4, 4, 101},
				box{1, 1, 2, 2, 102},
				box{3, 1, 2, 1, 103},
				box{0, 2, 2, 2, 104},
			},
			want: []box{
				box{3, 1, 2, 1, 103},
				box{1, 1, 2, 2, 102},
				box{0, 2, 2, 2, 104},
				box{0, 0, 4, 4, 101},
			},
		},
	}

	for _, test := range tests {
		original := test.in
		sort.Sort(BySize(test.in))

		if !BoxArraysAreEqual(test.in, test.want) {
			t.Errorf("Sorting     %v", original)
			t.Errorf("Resulted in %v", test.in)
			t.Errorf("Should be   %v", test.want)
		}
	}
} // -----  end of function Test_Sort  -----

func Test_Rotate(t *testing.T) {
	tests := []struct {
		in   *box
		want *box
	}{
		{
			in:   &box{0, 0, 1, 2, 100},
			want: &box{0, 0, 2, 1, 100},
		},
		{
			in:   &box{0, 0, 1, 1, 100},
			want: &box{0, 0, 1, 1, 100},
		},
		// invalid length
		{
			in:   &box{0, 0, 1, 5, 100},
			want: &box{0, 0, 5, 1, 100},
		},
		// invalid coordinates
		{
			in:   &box{0, 5, 1, 3, 100},
			want: &box{0, 5, 3, 1, 100},
		},
	}
	for _, test := range tests {
		original := *test.in
		test.in.Rotate()
		if !BoxesAreEqual(*test.want, *test.in) {
			space := "       "
			t.Errorf("Rotate %v\n %s Got    %v\n %s want   %v", original, space, test.in, space, test.want)
		} // -----  end if  -----
	} // -----  end for  -----
} // -----  end of function Test_Rotate  -----

func Test_IsSquare(t *testing.T) {
	tests := []struct {
		in   *box
		want bool
	}{
		// square box
		{
			in:   &box{0, 0, 1, 1, 101},
			want: true,
		},
		// rectangular box
		{
			in:   &box{3, 0, 4, 1, 104},
			want: false,
		},
		// square box at undefined coordinates
		{
			in:   &box{2, 7, 2, 2, 104},
			want: true,
		},
	}

	for _, test := range tests {
		got := test.in.IsSquare()
		if got != test.want {
			t.Errorf("(%v).IsSquare() == %t, want %t", test.in, got, test.want)
		}
	}
} // -----  end of function Test_IsSquare  -----

func Test_IsWithinBounds(t *testing.T) {
	type inputs struct {
		b    *box
		x, y uint8
	}
	tests := []struct {
		in   inputs
		want bool
	}{
		// box is ok
		{
			in:   inputs{&box{0, 0, 1, 1, 100}, 1, 1},
			want: true,
		},
		{
			in:   inputs{&box{0, 0, 1, 1, 100}, 3, 3},
			want: true,
		},
		{
			in:   inputs{&box{0, 0, 4, 4, 100}, 0, 0},
			want: true,
		},
		// box too big
		{
			in:   inputs{&box{0, 0, 3, 3, 100}, 2, 2},
			want: false,
		},
		{
			in:   inputs{&box{0, 0, 2, 1, 100}, 3, 3},
			want: false,
		},
	} // -----  end tests  -----

	for _, test := range tests {
		got := test.in.b.IsWithinBounds(test.in.x, test.in.y)
		if got != test.want {
			t.Errorf("(%v).IsWithinBounds(%d,%d) == %t, want %t", test.in.b, test.in.x, test.in.y, got, test.want)
		} // -----  end if  -----
	} // -----  end for  -----
} // -----  end of function Test_IsWithinBounds  -----

func Test_SetOrigin_ValidInputCoord_returnNoErr(t *testing.T) {
	type inputs struct {
		b    *box
		x, y uint8
	}
	type outputs struct {
		b   *box
		err error
	}

	tests := []struct {
		in   inputs
		want outputs
	}{
		{
			in:   inputs{&box{0, 0, 1, 1, 100}, 0, 0},
			want: outputs{&box{0, 0, 1, 1, 100}, nil},
		},
		{
			in:   inputs{&box{0, 0, 1, 1, 100}, 1, 1},
			want: outputs{&box{1, 1, 1, 1, 100}, nil},
		},
		{
			in:   inputs{&box{0, 0, 1, 1, 100}, 3, 3},
			want: outputs{&box{3, 3, 1, 1, 100}, nil},
		},
		{
			in:   inputs{&box{3, 3, 1, 1, 100}, 1, 1},
			want: outputs{&box{1, 1, 1, 1, 100}, nil},
		},
		{
			in:   inputs{&box{3, 3, 1, 1, 100}, 1, 1},
			want: outputs{&box{1, 1, 1, 1, 100}, nil},
		},
		{
			in:   inputs{&box{3, 3, 4, 4, 100}, 0, 0},
			want: outputs{&box{0, 0, 4, 4, 100}, nil},
		},
		{
			in:   inputs{&box{3, 3, 4, 4, 100}, 0, 0},
			want: outputs{&box{0, 0, 4, 4, 100}, nil},
		},
		{
			in:   inputs{&box{3, 3, 3, 3, 100}, 1, 1},
			want: outputs{&box{1, 1, 3, 3, 100}, nil},
		},
		{
			in:   inputs{&box{1, 1, 1, 4, 100}, 0, 3},
			want: outputs{&box{0, 3, 1, 4, 100}, nil},
		},
	} // -----  end tests  -----
	for _, test := range tests {
		got := test.in.b.SetOrigin(test.in.x, test.in.y)
		if (got != test.want.err) || !(BoxesAreEqual(*test.in.b, *test.want.b)) {
			t.Errorf("Got  (%v) and %v", test.in.b, got)
			t.Errorf("Want (%v) and %v", test.want.b, test.want.err)
		} // -----  end if  -----
	} // -----  end for  -----
} // -----  end of function Test_SetOrigin  -----
func Test_SetOrigin_InvalidInputCoord_returnErr(t *testing.T) {
	type inputs struct {
		b    *box
		x, y uint8
	}
	type outputs struct {
		b   *box
		err error
	}

	err_outOfBound := errors.New("box: Origin coordinates out of bounds.")

	tests := []struct {
		in   inputs
		want outputs
	}{
		{
			in:   inputs{&box{0, 0, 1, 1, 100}, 4, 4},
			want: outputs{&box{0, 0, 1, 1, 100}, err_outOfBound},
		},
		{
			in:   inputs{&box{0, 0, 1, 1, 100}, 4, 4},
			want: outputs{&box{0, 0, 1, 1, 100}, err_outOfBound},
		},
		{
			in:   inputs{&box{5, 5, 1, 1, 100}, 4, 4},
			want: outputs{&box{5, 5, 1, 1, 100}, err_outOfBound},
		},
	} // -----  end tests  -----
	for _, test := range tests {
		got := test.in.b.SetOrigin(test.in.x, test.in.y)
		if got.Error() != test.want.err.Error() {
			t.Errorf("Got  %v", got)
			t.Errorf("want %v", test.want.err)
		} // -----  end if  -----
		if !(BoxesAreEqual(*test.in.b, *test.want.b)) {
			t.Errorf("Got  (%v)", test.in.b)
			t.Errorf("Want (%v)", test.want.b)
		} // -----  end if  -----
	} // -----  end for  -----
} // -----  end of function Test_SetOrigin_InvalidInputCoord_returnErr  -----
func Test_SetOrigin_InvalidSizeForCoord_returnErr(t *testing.T) {
	type inputs struct {
		b    *box
		x, y uint8
	}
	type outputs struct {
		b   *box
		err error
	}

	err_hangOver := errors.New("box.SetOrigin: Hangs over pallet edge. Unable to place box on grid")
	err_invalidSize := errors.New("box: Has invalid size.")

	tests := []struct {
		in   inputs
		want outputs
	}{
		{
			in:   inputs{&box{0, 0, 2, 2, 100}, 3, 3},
			want: outputs{&box{0, 0, 2, 2, 100}, err_hangOver},
		},
		{
			in:   inputs{&box{2, 2, 4, 4, 100}, 2, 2},
			want: outputs{&box{2, 2, 4, 4, 100}, err_hangOver},
		},
		//box to big
		{
			in:   inputs{&box{0, 0, 5, 5, 100}, 0, 0},
			want: outputs{&box{0, 0, 5, 5, 100}, err_invalidSize},
		},
		//box has zero size
		{
			in:   inputs{&box{0, 0, 2, 0, 100}, 0, 0},
			want: outputs{&box{0, 0, 2, 0, 100}, err_invalidSize},
		},
	} // -----  end tests  -----

	for _, test := range tests {
		got := test.in.b.SetOrigin(test.in.x, test.in.y)
		if got.Error() != test.want.err.Error() {
			t.Errorf("Got  %v", got)
			t.Errorf("want %v", test.want.err)
		} // -----  end if  -----
		if !(BoxesAreEqual(*test.in.b, *test.want.b)) {
			t.Errorf("Got  (%v)", test.in.b)
			t.Errorf("Want (%v)", test.want.b)
		} // -----  end if  -----
	} // -----  end for  -----

} // -----  end of function Test_SetOrigin_InvalidSizeForCoord_returnErr  -----
