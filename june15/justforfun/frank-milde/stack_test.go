//
// =========================================================================
//
//       Filename:  stack_test.go
//
//    Description:  Testing box stack.
//
//        Version:  1.0
//        Created:  06/16/2015 07:08:46 PM
//       Revision:  none
//       Compiler:  g++
//
//          Usage:  <+USAGE+>
//
//         Output:  <+OUTPUT+>
//
//         Author:  Frank Milde (FM), frank.milde (at) posteo.de
//        Company:
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

func Test_NewStack_LengthIsZero(t *testing.T) {
	s := NewStack()

	got := len(s)
	want := 0

	if got != want {
		t.Errorf("Got %d, want %d", got, want)
	}
}
func Test_NewStack_FrontIsNil(t *testing.T) {
	got := NewStack()
	if got != nil {
		t.Errorf("Got %v, want nil", got)
	}
}

func Test_IsEmpty_EmptyStack_ReturnTrue(t *testing.T) {
	s := NewStack()

	got := s.IsEmpty()
	want := true

	if got != want {
		t.Errorf("Got %d, want %d", got, want)
	}
}
func Test_IsEmpty_NonEmptyStack_ReturnFalse(t *testing.T) {
	s := NewStack()
	s.Push(box{0, 0, 1, 1, 100})

	got := s.IsEmpty()
	want := false

	if got != want {
		t.Errorf("Got %d, want %d", got, want)
	}
}

func Test_Front_NonEmptyStack_ReturnLastElement(t *testing.T) {
	s := NewStack()

	b1 := box{0, 0, 1, 1, 100}
	b2 := box{0, 0, 1, 1, 101}

	s.Push(b1)
	s.Push(b2)

	got := s.Front()
	want := b2

	if !BoxesAreEqual(got, want) {
		t.Errorf("Got %v, want %v", got, want)
	}
}
func Test_Front_EmptyStack_ReturnEmptyBox(t *testing.T) {
	s := NewStack()

	got := s.Front()
	want := emptybox

	if !BoxesAreEqual(got, want) {
		t.Errorf("Got %v, want %v", got, want)
	}
}

func Test_Push_AddBoxToEmptyStack(t *testing.T) {

	s := NewStack()
	b := box{0, 0, 1, 1, 100}

	s.Push(b)

	got := s[0]
	want := b

	if !BoxesAreEqual(got, want) {
		t.Errorf("Got %v, want %v", got, want)
	}
}
func Test_Push_AddBoxToNonEmptyStack(t *testing.T) {

	s := NewStack()
	b1 := box{0, 0, 1, 1, 100}
	b2 := box{0, 0, 1, 1, 101}

	s.Push(b1)
	s.Push(b2)

	got := s[1]
	want := b2

	if !BoxesAreEqual(got, want) {
		t.Errorf("Boxes: got %v, want %v", got, want)
	}
	if len(s) != 2 {
		t.Errorf("Length: got %d, want %d", len(s), 2)
	}
}

func Test_Pop_NonEmptyStackUntilEmpty_ReturnLastElement(t *testing.T) {
	s := NewStack()
	b1 := box{0, 0, 1, 1, 100}
	b2 := box{0, 0, 1, 1, 101}

	s.Push(b1)
	s.Push(b2)

	type wants struct {
		l int
		b box
	}

	tests := []struct {
		want wants
	}{
		{
			want: wants{1, b2},
		},
		{
			want: wants{0, b1},
		},
		{
			want: wants{0, emptybox},
		},
	} // end tests

	for i, test := range tests {
		gotb := s.Pop()
		gotl := len(s)
		if !BoxesAreEqual(gotb, test.want.b) {
			t.Errorf("Run %d", i)
			t.Errorf("Boxes: Got %v, want %v", gotb, test.want.b)
		}
		if gotl != test.want.l {
			t.Errorf("Run %d", i)
			t.Errorf("Length: Got %v, want %v", gotl, test.want.l)
		}
	}
}

func Test_StacksAreEqual_GetEmptyStacks_ReturnTrue(t *testing.T) {
	got := StacksAreEqual(NewStack(), NewStack())
	want := true
	if got != want {
		t.Errorf("Got %b, want %b", got, want)
	}
} // -----  end of function Test_StacksAreEqual  -----
func Test_StacksAreEqual_GetEqualStacks_ReturnTrue(t *testing.T) {
	s1 := NewStack()
	s2 := NewStack()

	b1 := box{0, 0, 1, 1, 101}
	b2 := box{0, 0, 1, 2, 102}
	b3 := box{0, 0, 1, 3, 103}
	b4 := box{0, 0, 4, 4, 110}

	boxes := []box{b1, b2, b3, b4}

	for _, box := range boxes {
		s1.Push(box)
		s2.Push(box)
	}
	got := StacksAreEqual(s1, s2)
	want := true
	if got != want {
		t.Errorf("Got %b, want %b", got, want)
	}
} // -----  end of function Test_StacksAreEqual  -----
func Test_StacksAreEqual_GetNonEqualStacks_ReturnFalse(t *testing.T) {
	s1 := NewStack()
	s2 := NewStack()

	b1 := box{0, 0, 1, 1, 101}
	b2 := box{0, 0, 1, 2, 102}
	b3 := box{0, 0, 1, 3, 103}
	b4 := box{0, 0, 4, 4, 110}

	boxes := []box{b1, b2, b3, b4}

	for _, box := range boxes {
		s1.Push(box)
		s2.Push(box)
	}
	s1.Push(b1)

	got := StacksAreEqual(s1, s2)
	want := false
	if got != want {
		t.Errorf("Got %b, want %b", got, want)
	}
} // -----  end of function Test_StacksAreEqual  -----

// === old ===

/*
func Test_Len(t *testing.T) {
	tests := []struct {
		in   *Stack
		want uint
	}{
		{
			in:   &Stack{Element{}, 0},
			want: 0,
		},
		{
			in:   &Stack{Element{nil, &Stack{}, &box{0, 0, 1, 1, 100}}, 1},
			want: 1,
		},
	} // -----  end of tests  -----

	for _, test := range tests {
		got := test.in.Len()
		if got != test.want {
			t.Errorf("Got %d, want %d", got, test.want)
		}
	} // -----  end of for  -----
} // -----  end of function Test_Len  -----

func Test_NewStack_CreateNewStack_GetEmptyStack(t *testing.T) {

	got := NewStack()
	want := &Stack{}

	if (got.root != want.root) && (got.length != want.length) {
		t.Errorf("Got %v, want %v", &got, &want)
	}
} // -----  end of function Test_NewStack  -----

func Test_Next(t *testing.T) {
	in := &Element{
		&Element{nil, &Stack{}, &box{0, 0, 1, 1, 101}},
		&Stack{},
		&box{0, 0, 1, 1, 100},
	}
	want := &Element{nil, &Stack{}, &box{0, 0, 1, 1, 101}}

	// it is easier to compare the nils
	if in.next.next != want.next {
		t.Errorf("Next pointers: got (%v), want (%v)", in.next, want)
	}
	got := in.Next()
	if !BoxesAreEqual(*got.b, *want.b) {
		t.Errorf("Boxes: got (%v), want (%v)", got.b, want.b)
	}
} // -----  end of function Test_Next  -----
func Test_Next_LastElement_ReturnNil(t *testing.T) {
	s := NewStack()
	b := box{0, 0, 1, 1, 100}
	c := box{1, 1, 2, 2, 101}

	s.Push(&b)
	s.Push(&c)

	if s.Front().Next() == nil {
		t.Errorf("got %v, wanted non-nil", s.Front().Next())
	}
	gotB := *s.Front().Next().b
	if !BoxesAreEqual(gotB, b) {
		t.Errorf("Boxes: got (%v), want (%v)", gotB, b)
	}
} // -----  end of function Test_Next  -----
func Test_Next_UsingElementListTakeLastElement_ReturnNil(t *testing.T) {
	in := &Element{
		&Element{nil, &Stack{}, &box{0, 0, 1, 1, 101}},
		&Stack{},
		&box{0, 0, 1, 1, 100},
	}
	var want *Element = nil

	if in.Next().Next() != want {
		t.Errorf("Next pointers: got (%v), want (%v)", in.next, want)
	}
} // -----  end of function Test_Next  -----

func Test_Front_NonEmptyStack_ReturnFirstElement(t *testing.T) {

	s := NewStack()
	b := box{0, 0, 1, 1, 100}
	c := box{1, 1, 2, 2, 101}

	s.Push(&b)
	s.Push(&c)

	var want uint = 2

	if s.Len() != want {
		t.Errorf("Got s.Len() = %d, want %d", s.Len(), want)
	}
	if !BoxesAreEqual(*s.Front().b, c) {
		t.Errorf("Boxes got:  s.b = (%v)", s.Front().b)
		t.Errorf("Boxes want:   b = (%v)", c)
	}
} // -----  end of function Test_Push_AddBoxToNonEmptyStack  -----
func Test_Front_EmptyStack_ReturnNil(t *testing.T) {

	s := NewStack()

	var want uint = 0

	if s.Len() != want {
		t.Errorf("Got s.Len() = %d, want %d", s.Len(), want)
	}
	if s.Front() != nil {
		t.Errorf("Front got:  s.Front = (%v)", s.Front())
		t.Errorf("Front want:   b = (%v)", nil)
	}
} // -----  end of function Test_Front_EmptyStack_ReturnNil  -----

func Test_Box_GetCorrectBox(t *testing.T) {

	e := &Element{nil, &Stack{}, &box{0, 0, 1, 1, 100}}

	got := e.Box()
	want := box{0, 0, 1, 1, 100}

	if !BoxesAreEqual(*got, want) {
		t.Errorf("got (%v), want (%v)", got, want)
	}

} // -----  end of function Test_Box  -----

func Test_Box_GetNil(t *testing.T) {

	e := &Element{nil, &Stack{}, nil}

	got := e.Box()
	want := box{0, 0, 1, 1, 100}

	if got != nil {
		t.Errorf("got (%v), want (%v)", got, want)
	}

} // -----  end of function Test_Box  -----

func Test_StacksAreEqual_GetEmptyStacks_ReturnTrue(t *testing.T) {
	got := StacksAreEqual(NewStack(), NewStack())
	want := true
	if got != want {
		t.Errorf("Got %b, want %b", got, want)
	}
} // -----  end of function Test_StacksAreEqual  -----
func Test_StacksAreEqual_GetEqualStacks_ReturnTrue(t *testing.T) {
	type inputs struct {
		a, b *Stack
	}
	tests := []struct {
		in   inputs
		want bool
	}{
		{
			in: inputs{
				&Stack{
					Element{
						&Element{nil, &Stack{}, &box{0, 0, 1, 1, 101}},
						&Stack{},
						&box{0, 0, 1, 1, 100},
					},
					2,
				},
				&Stack{
					Element{
						&Element{nil, &Stack{}, &box{0, 0, 1, 1, 101}},
						&Stack{},
						&box{0, 0, 1, 1, 100},
					},
					2,
				},
			}, // -----  end of inputs  -----
			want: true,
		},
	} // -----  end of tests  -----

	for _, test := range tests {
		got := StacksAreEqual(test.in.a, test.in.b)
		if got != test.want {
			t.Errorf("Got %t, want %t", got, test.want)
		}
	}
} // -----  end of function Test_StacksAreEqual  -----
func Test_StacksAreEqual_GetUnEqualStacks_ReturnFalse(t *testing.T) {
	type inputs struct {
		a, b *Stack
	}
	tests := []struct {
		in   inputs
		want bool
	}{
		// id is wrong
		{
			in: inputs{
				&Stack{
					Element{
						&Element{nil, &Stack{}, &box{0, 0, 1, 1, 101}},
						&Stack{},
						&box{0, 0, 1, 1, 102},
					},
					2,
				},
				&Stack{
					Element{
						&Element{nil, &Stack{}, &box{0, 0, 1, 1, 101}},
						&Stack{},
						&box{0, 0, 1, 1, 100},
					},
					2,
				},
			}, // -----  end of inputs  -----
			want: false,
		},
		// number elements is wrong
		{
			in: inputs{
				&Stack{
					Element{
						&Element{nil, &Stack{}, &box{0, 0, 1, 1, 101}},
						&Stack{},
						&box{0, 0, 1, 1, 102},
					},
					2,
				},
				&Stack{
					Element{nil, &Stack{}, &box{0, 0, 1, 1, 101}},
					1,
				},
			}, // -----  end of inputs  -----
			want: false,
		},
	} // -----  end of tests  -----

	for _, test := range tests {
		got := StacksAreEqual(test.in.a, test.in.b)
		if got != test.want {
			t.Errorf("Got %t, want %t", got, test.want)
		}
	}
} // -----  end of function Test_StacksAreEqual  -----

func Test_NewStack(t *testing.T) {
	got := NewStack()

	if got.Len() != 0 {
		t.Errorf("s.Len(): got %d, want 0", got.Len())
	}
	if got.root.next != &got.root {
		t.Errorf("root pointers are not equal: %v != %v", got.root.next, &got.root)
	}
} // -----  end of function Test_NewStack  -----

func Test_Push_AddBoxToEmptyStack(t *testing.T) {

	s := NewStack()
	b := box{0, 0, 1, 1, 100}

	s.Push(&b)
	var want uint = 1

	if s.Len() != want {
		t.Errorf("Got s.Len() = %d, want %d", s.Len(), want)
	}
	if !BoxesAreEqual(*s.Front().b, b) {
		t.Errorf("Boxes: s.b = (%v)", s.Front(), s.Len())
		t.Errorf("Boxes:   b = (%v)", b)
	}
} // -----  end of function Test_Push  -----
func Test_Push_AddBoxToNonEmptyStack(t *testing.T) {

	s := NewStack()
	b := box{0, 0, 1, 1, 100}
	c := box{1, 1, 2, 2, 101}

	s.Push(&b)
	s.Push(&c)

	var want uint = 2

	if s.Len() != want {
		t.Errorf("Got s.Len() = %d, want %d", s.Len(), want)
	}
	if !BoxesAreEqual(*s.Front().b, c) {
		t.Errorf("Boxes: s.b = (%v)", s.root.b, s.Len())
		t.Errorf("Boxes:   b = (%v)", b)
	}
} // -----  end of function Test_Push_AddBoxToNonEmptyStack  -----

func Test_Pop_DeleteBoxFromNonEmptyStack_GetBox(t *testing.T) {
	s := NewStack()
	i := box{0, 0, 1, 1, 100}
	j := box{1, 1, 2, 2, 101}

	s.Push(&i)
	s.Push(&j)

	type wants struct {
		l uint
		b box
	}

	tests := []struct {
		want wants
	}{
		{
			want: wants{1, j},
		},
		{
			want: wants{0, i},
		},
	} // end tests

	for _, test := range tests {
		gotB := *s.Pop()
		wantB := test.want.b
		gotL := s.Len()
		wantL := test.want.l

		if gotL != wantL {
			t.Errorf("Got s.Len() = %d, want %d", gotL, wantL)
		}
		if !BoxesAreEqual(gotB, wantB) {
			t.Errorf("Boxes: s.Pop() = (%v)", gotB)
			t.Errorf("Boxes:       b = (%v)", wantB)
		}
	}
} // -----  end of function Test_Pop_DeleteBoxFromNonEmptyStack  -----
func Test_Pop_DeleteBoxFromEmptyStack_ReturnNil(t *testing.T) {
	s := NewStack()

	got := s.Pop()
	var want *box = nil

	if got != want {
		t.Errorf("s.Pop() got %v, want %v", got, want)
	}
} // -----  end of function Test_Pop_DeleteBoxFromEmptyStack  -----
*/
