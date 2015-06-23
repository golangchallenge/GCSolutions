//
// =========================================================================
//
//       Filename:  hash_test.go
//
//    Description:
//
//        License:  GNU General Public License
//      Copyright:  Copyright (c) 2015, Frank Milde
//
// =========================================================================
//

package main

import (
	//	"fmt"
	"errors"
	"testing"
)

func Test_Table_IsEmpty(t *testing.T) {
	table := NewGrid()

	if !table.IsEmpty() {
		t.Errorf("Table not empty")
	}
}

func Test_NewTable_InitialSizeIs17(t *testing.T) {

	store := NewTable()

	if len(store) != TABLESIZE {
		t.Errorf("New Table length is %v, wanted %v", len(store), TABLESIZE)
	}

	for i, s := range store {
		if len(s) != 0 {
			t.Errorf("Stack %d in New Table has length %d, want 0 ", i, len(s))
		}
		if s.Front() != emptybox {
			t.Errorf("Stack %d in New Table has Front %v, want emptybox %v", i, s.Front(), emptybox)
		}
	}

}

func Test_Add_SingleBoxPerStack(t *testing.T) {

	s := NewTable()

	b1 := box{0, 0, 1, 1, 101}
	b2 := box{0, 0, 1, 2, 102}
	b3 := box{0, 0, 1, 3, 103}
	b4 := box{0, 0, 1, 4, 104}
	b5 := box{0, 0, 2, 2, 105}
	b6 := box{0, 0, 2, 3, 106}
	b8 := box{0, 0, 2, 4, 107}
	b9 := box{0, 0, 3, 3, 108}
	b12 := box{0, 0, 3, 4, 109}
	b16 := box{0, 0, 4, 4, 110}

	boxes := []box{b1, b2, b3, b4, b5, b6, b8, b9, b12, b16}

	for _, box := range boxes {
		s.Add(box)
	}

	for i, stack := range s {
		got := stack.Pop()
		if !BoxesAreEqual(got, boxes[i]) {
			t.Errorf("%d Want (%v)", i, boxes[i])
			t.Errorf("%d Got  (%v) with len %d", i, got, len(stack))
		}
	}
}
func Test_Add_MultipleBoxesPerStack(t *testing.T) {

	s := NewTable()

	b1 := box{0, 0, 1, 1, 101}
	b2 := box{0, 0, 1, 1, 102}
	b3 := box{0, 0, 1, 1, 103}
	b4 := box{0, 0, 1, 1, 104}
	b5 := box{0, 0, 2, 4, 105}
	b6 := box{0, 0, 2, 4, 106}
	b8 := box{0, 0, 2, 4, 107}
	b9 := box{0, 0, 3, 3, 108}
	b12 := box{0, 0, 3, 3, 109}

	boxes := []box{b1, b2, b3, b4, b5, b6, b8, b9, b12}

	for _, box := range boxes {
		s.Add(box)
	}

	// all 1x1 boxes
	var s1 Stack = Stack{b1, b2, b3, b4}
	// all 2x4 boxes
	var s8 Stack = Stack{b5, b6, b8}
	// all 3x4 boxes
	var s9 Stack = Stack{b9, b12}

	if !StacksAreEqual(s[0], s1) {
		t.Errorf("%d Want (%v)", s1)
		t.Errorf("%d Got  (%v)", s[0])
	}
	if !StacksAreEqual(s[6], s8) {
		t.Errorf("%d Want (%v)", s8)
		t.Errorf("%d Got  (%v)", s[6])
	}
	if !StacksAreEqual(s[7], s9) {
		t.Errorf("%d Want (%v)", s9)
		t.Errorf("%d Got  (%v)", s[7])
	}
}
func Test_Add_InvalidInput_CheckErrMsg(t *testing.T) {

	s := NewTable()

	zeroWidth := box{0, 0, 0, 1, 101}
	tooLong := box{0, 0, 2, 5, 102}

	invSizeErr := errors.New("Add box to table: Box has invalid size.")

	boxes := []box{emptybox, zeroWidth, tooLong}
	want := invSizeErr

	errMsgs := make([]error, 3)

	// collect error messages
	for i, box := range boxes {
		errMsgs[i] = s.Add(box)
	}

	for _, got := range errMsgs {
		if got.Error() != want.Error() {
			t.Errorf("Got (%v), want (%v)", got, want)
		}
	}
}

func Test_HashBox_InputRegularBoxes(t *testing.T) {
	b1 := box{0, 0, 1, 1, 101}
	b2 := box{0, 0, 1, 2, 102}
	b3 := box{0, 0, 1, 3, 103}
	b4 := box{0, 0, 1, 4, 104}
	b5 := box{0, 0, 2, 2, 105}
	b6 := box{0, 0, 2, 3, 106}
	b8 := box{0, 0, 2, 4, 107}
	b9 := box{0, 0, 3, 3, 108}
	b12 := box{0, 0, 3, 4, 109}
	b16 := box{0, 0, 4, 4, 110}

	boxes := []box{b1, b2, b3, b4, b5, b6, b8, b9, b12, b16}
	wants := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	for i, box := range boxes {
		got, _ := HashBox(&box)
		want := wants[i]
		if want != got {
			t.Errorf("Got %d, want %d", got, want)
		}
	}
}
func Test_HashBox_InputInvalidBoxes_CheckErrMsg(t *testing.T) {
	zeroWidth := box{0, 0, 0, 1, 101}
	tooLong := box{0, 0, 2, 5, 102}

	boxes := []box{emptybox, zeroWidth, tooLong}
	want := errors.New("hash: Invalid size.")

	for _, box := range boxes {
		_, got := HashBox(&box)
		if want.Error() != got.Error() {
			t.Errorf("Got (%v), want (%v)", got, want)
		}
	}
}

func Test_Hash_InputRegularBoxes(t *testing.T) {

	type inputs struct {
		size   int
		orient Orientation
	}

	tests := []struct {
		in   inputs
		want int
	}{
		// 2x2 box
		{
			in: inputs{
				4,
				SQUAREGRID,
			},
			want: 4,
		},
		{
			in: inputs{
				4,
				HORIZONTAL,
			},
			want: 3,
		},
		{
			in: inputs{
				16,
				HORIZONTAL,
			},
			want: 9,
		},
	}

	for _, test := range tests {
		got, _ := Hash(test.in.size, test.in.orient)
		if got != test.want {
			t.Errorf("Got: %v ", got)
			t.Errorf("Want: %v ", test.want)
		}
	}
}
func Test_Hash_InputInvalid_CheckErrMsg(t *testing.T) {
	type inputs struct {
		size   int
		orient Orientation
	}

	tests := []struct {
		in   inputs
		want error
	}{
		{
			in:   inputs{0, SQUAREGRID},
			want: errors.New("hash: Invalid size."),
		},
		{
			in:   inputs{10, SQUAREGRID},
			want: errors.New("hash: Invalid size."),
		},
		{
			in:   inputs{-1, SQUAREGRID},
			want: errors.New("hash: Invalid size."),
		},
		{
			in:   inputs{20, VERTICAL},
			want: errors.New("hash: Invalid size."),
		},
		{
			in:   inputs{2, 5},
			want: errors.New("hash: Invalid orientation."),
		},
	}

	for _, test := range tests {
		_, got := Hash(test.in.size, test.in.orient)
		if got.Error() != test.want.Error() {
			t.Errorf("Got: %v ", got.Error())
			t.Errorf("Want: %v ", test.want.Error())
		}
	}
}

func Test_GetBoxThatFitsOrIsEmpty_FullTableRequestedBoxIsAvaiable_getBox(t *testing.T) {

	// Fill table
	s := NewTable()

	boxes := []box{
		box{0, 0, 1, 1, 101},
		box{0, 0, 1, 1, 111},
		box{0, 0, 1, 2, 102},
		box{0, 0, 1, 2, 112},
		box{0, 0, 1, 3, 103},
		box{0, 0, 1, 3, 113},
		box{0, 0, 1, 4, 104},
		box{0, 0, 1, 4, 114},
		box{0, 0, 2, 2, 105},
		box{0, 0, 2, 2, 115},
		box{0, 0, 2, 3, 106},
		box{0, 0, 2, 3, 116},
		box{0, 0, 2, 4, 107},
		box{0, 0, 2, 4, 117},
		box{0, 0, 3, 3, 108},
		box{0, 0, 3, 3, 118},
		box{0, 0, 3, 4, 109},
		box{0, 0, 3, 4, 119},
		box{0, 0, 4, 4, 110},
		box{0, 0, 4, 4, 120},
	}

	for _, box := range boxes {
		s.Add(box)
	}
	// end of Fill up

	type inputs struct {
		size   int
		orient Orientation
	}
	type wants struct {
		b box
		t Table
	}

	tests := []struct {
		in   inputs
		want wants
	}{
		// 2x2 box
		{
			in: inputs{
				4,
				SQUAREGRID,
			},
			want: wants{
				box{0, 0, 2, 2, 115},
				Table{
					Stack{
						box{0, 0, 1, 1, 101},
						box{0, 0, 1, 1, 111},
					},
					Stack{
						box{0, 0, 1, 2, 102},
						box{0, 0, 1, 2, 112},
					},
					Stack{
						box{0, 0, 1, 3, 103},
						box{0, 0, 1, 3, 113},
					},
					Stack{
						box{0, 0, 1, 4, 104},
						box{0, 0, 1, 4, 114},
					},
					Stack{box{0, 0, 2, 2, 105}},
					Stack{
						box{0, 0, 2, 3, 106},
						box{0, 0, 2, 3, 116},
					},
					Stack{
						box{0, 0, 2, 4, 107},
						box{0, 0, 2, 4, 117},
					},
					Stack{
						box{0, 0, 3, 3, 108},
						box{0, 0, 3, 3, 118},
					},
					Stack{
						box{0, 0, 3, 4, 109},
						box{0, 0, 3, 4, 119},
					},
					Stack{
						box{0, 0, 4, 4, 110},
						box{0, 0, 4, 4, 120},
					},
				},
			},
		},
	}

	for _, test := range tests {
		got, _ := s.GetBoxThatFitsOrIsEmpty(test.in.size, test.in.orient)
		if !BoxesAreEqual(got, test.want.b) {
			t.Errorf("Box error")
			t.Errorf("Got: %v ", got)
			t.Errorf("Want: %v ", test.want.b)
		}
		if !TablesAreEqual(s, test.want.t) {
			t.Errorf("Table error")
			t.Errorf("Got: \n%v ", s)
			t.Errorf("Want:\n%v ", test.want.t)
		}
	}
}
func Test_GetBoxThatFitsOrIsEmpty_RequestedBoxIsUnavailable_getNextSmallerBox(t *testing.T) {

	type inputs struct {
		table  Table
		size   int
		orient Orientation
	}
	type wants struct {
		b box
		t Table
	}

	tests := []struct {
		in   inputs
		want wants
	}{
		// request 3x3, not present, ignore 4x2 box and pick 3x2 box
		{
			in: inputs{
				Table{
					Stack{
						box{0, 0, 1, 1, 101},
						box{0, 0, 1, 1, 111},
					},
					Stack{
						box{0, 0, 1, 2, 102},
						box{0, 0, 1, 2, 112},
					},
					Stack{
						box{0, 0, 1, 3, 103},
						box{0, 0, 1, 3, 113},
					},
					Stack{
						box{0, 0, 1, 4, 104},
						box{0, 0, 1, 4, 114},
					},
					Stack{
						box{0, 0, 2, 2, 105},
						box{0, 0, 2, 2, 115},
					},
					Stack{
						box{0, 0, 2, 3, 106},
						box{0, 0, 2, 3, 116},
					},
					Stack{
						box{0, 0, 2, 4, 107},
						box{0, 0, 2, 4, 117},
					},
					Stack{},
					Stack{
						box{0, 0, 3, 4, 109},
						box{0, 0, 3, 4, 119},
					},
					Stack{},
				},
				9,
				SQUAREGRID,
			},
			want: wants{
				box{0, 0, 2, 3, 116},
				Table{
					Stack{
						box{0, 0, 1, 1, 101},
						box{0, 0, 1, 1, 111},
					},
					Stack{
						box{0, 0, 1, 2, 102},
						box{0, 0, 1, 2, 112},
					},
					Stack{
						box{0, 0, 1, 3, 103},
						box{0, 0, 1, 3, 113},
					},
					Stack{
						box{0, 0, 1, 4, 104},
						box{0, 0, 1, 4, 114},
					},
					Stack{
						box{0, 0, 2, 2, 105},
						box{0, 0, 2, 2, 115},
					},
					Stack{
						box{0, 0, 2, 3, 106},
					},
					Stack{
						box{0, 0, 2, 4, 107},
						box{0, 0, 2, 4, 117},
					},
					Stack{},
					Stack{
						box{0, 0, 3, 4, 109},
						box{0, 0, 3, 4, 119},
					},
					Stack{},
				},
			},
		},
		// request 4x4, not present, pick 4x3
		{
			in: inputs{
				Table{
					Stack{
						box{0, 0, 1, 1, 101},
						box{0, 0, 1, 1, 111},
					},
					Stack{
						box{0, 0, 1, 2, 102},
						box{0, 0, 1, 2, 112},
					},
					Stack{
						box{0, 0, 1, 3, 103},
						box{0, 0, 1, 3, 113},
					},
					Stack{
						box{0, 0, 1, 4, 104},
						box{0, 0, 1, 4, 114},
					},
					Stack{
						box{0, 0, 2, 2, 105},
						box{0, 0, 2, 2, 115},
					},
					Stack{
						box{0, 0, 2, 3, 106},
						box{0, 0, 2, 3, 116},
					},
					Stack{
						box{0, 0, 2, 4, 107},
						box{0, 0, 2, 4, 117},
					},
					Stack{
						box{0, 0, 3, 3, 108},
						box{0, 0, 3, 3, 118},
					},
					Stack{
						box{0, 0, 3, 4, 109},
						box{0, 0, 3, 4, 119},
					},
					Stack{},
				},
				16,
				SQUAREGRID,
			},
			want: wants{
				box{0, 0, 3, 4, 119},
				Table{
					Stack{
						box{0, 0, 1, 1, 101},
						box{0, 0, 1, 1, 111},
					},
					Stack{
						box{0, 0, 1, 2, 102},
						box{0, 0, 1, 2, 112},
					},
					Stack{
						box{0, 0, 1, 3, 103},
						box{0, 0, 1, 3, 113},
					},
					Stack{
						box{0, 0, 1, 4, 104},
						box{0, 0, 1, 4, 114},
					},
					Stack{
						box{0, 0, 2, 2, 105},
						box{0, 0, 2, 2, 115},
					},
					Stack{
						box{0, 0, 2, 3, 106},
						box{0, 0, 2, 3, 116},
					},
					Stack{
						box{0, 0, 2, 4, 107},
						box{0, 0, 2, 4, 117},
					},
					Stack{
						box{0, 0, 3, 3, 108},
						box{0, 0, 3, 3, 118},
					},
					Stack{
						box{0, 0, 3, 4, 109},
					},
					Stack{},
				},
			},
		},
		// request 2x2, not present, ignore 4x1,3x1 box and pick 2x1 box
		{
			in: inputs{
				Table{
					Stack{
						box{0, 0, 1, 1, 101},
						box{0, 0, 1, 1, 111},
					},
					Stack{
						box{0, 0, 1, 2, 102},
						box{0, 0, 1, 2, 112},
					},
					Stack{
						box{0, 0, 1, 3, 103},
						box{0, 0, 1, 3, 113},
					},
					Stack{
						box{0, 0, 1, 4, 104},
						box{0, 0, 1, 4, 114},
					},
					Stack{},
					Stack{
						box{0, 0, 2, 3, 106},
						box{0, 0, 2, 3, 116},
					},
					Stack{
						box{0, 0, 2, 4, 107},
						box{0, 0, 2, 4, 117},
					},
					Stack{},
					Stack{
						box{0, 0, 3, 4, 109},
						box{0, 0, 3, 4, 119},
					},
					Stack{},
				},
				4,
				SQUAREGRID,
			},
			want: wants{
				box{0, 0, 1, 2, 112},
				Table{
					Stack{
						box{0, 0, 1, 1, 101},
						box{0, 0, 1, 1, 111},
					},
					Stack{
						box{0, 0, 1, 2, 102},
					},
					Stack{
						box{0, 0, 1, 3, 103},
						box{0, 0, 1, 3, 113},
					},
					Stack{
						box{0, 0, 1, 4, 104},
						box{0, 0, 1, 4, 114},
					},
					Stack{},
					Stack{
						box{0, 0, 2, 3, 106},
						box{0, 0, 2, 3, 116},
					},
					Stack{
						box{0, 0, 2, 4, 107},
						box{0, 0, 2, 4, 117},
					},
					Stack{},
					Stack{
						box{0, 0, 3, 4, 109},
						box{0, 0, 3, 4, 119},
					},
					Stack{},
				},
			},
		},
	}

	for _, test := range tests {
		got, _ := test.in.table.GetBoxThatFitsOrIsEmpty(test.in.size, test.in.orient)
		if !BoxesAreEqual(got, test.want.b) {
			t.Errorf("Box error")
			t.Errorf("Got: %v ", got)
			t.Errorf("Want: %v ", test.want.b)
			t.Errorf("\n")
		}
		if !TablesAreEqual(test.in.table, test.want.t) {
			t.Errorf("Table error")
			t.Errorf("Got: \n%v ", test.in.table)
			t.Errorf("Want:\n%v ", test.want.t)
			t.Errorf("\n")
		}
	}
}

func Test_TablesAreEqual(t *testing.T) {
	t1 := NewTable()
	t2 := NewTable()

	b1 := box{0, 0, 1, 1, 101}
	b2 := box{0, 0, 1, 1, 102}
	b3 := box{0, 0, 1, 1, 103}
	b4 := box{0, 0, 1, 1, 104}
	b5 := box{0, 0, 2, 4, 105}
	b6 := box{0, 0, 2, 4, 106}
	b8 := box{0, 0, 2, 4, 107}
	b9 := box{0, 0, 3, 3, 108}
	b12 := box{0, 0, 3, 3, 109}

	boxes := []box{b1, b2, b3, b4, b5, b6, b8, b9, b12}

	for _, box := range boxes {
		t1.Add(box)
		t2.Add(box)
	}
	t1.Add(b1)

	got := TablesAreEqual(t1, t2)
	want := false

	if got != want {
		t.Errorf("%v ", t1)
		t.Errorf("%v ", t2)
	}

}

func Test_String(t *testing.T) {
	s := NewTable()

	b1 := box{0, 0, 1, 1, 101}
	b2 := box{0, 0, 1, 2, 102}
	b3 := box{0, 0, 1, 3, 103}
	b4 := box{0, 0, 1, 4, 104}
	b5 := box{0, 0, 2, 2, 105}
	b6 := box{0, 0, 2, 3, 106}
	b8 := box{0, 0, 2, 4, 107}
	b9 := box{0, 0, 3, 3, 108}
	b12 := box{0, 0, 3, 4, 109}
	b16 := box{0, 0, 4, 4, 110}

	boxes := []box{b1, b2, b3, b4, b5, b6, b8, b9, b12, b16}

	for _, box := range boxes {
		s.Add(box)
	}

	got := s.String()

	want := `
[ 1]  -->  (0 0 1 1 101)  
[ 2]  -->  (0 0 1 2 102)  
[ 3]  -->  (0 0 1 3 103)  
[ 4]  -->  (0 0 1 4 104)  
[4s]  -->  (0 0 2 2 105)  
[ 6]  -->  (0 0 2 3 106)  
[ 8]  -->  (0 0 2 4 107)  
[ 9]  -->  (0 0 3 3 108)  
[12]  -->  (0 0 3 4 109)  
[16]  -->  (0 0 4 4 110)  
`
	if got != want {
		t.Errorf("String is wrong")
		t.Errorf("got:\n%s", got)
		t.Errorf("want\n%s", want)
	}

}
