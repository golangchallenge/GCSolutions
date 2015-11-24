package main

import (
	"reflect"
	"testing"
)

var bin = &Binary{}

func TestConvBit2Slice(t *testing.T) {
	testSet := []struct {
		mask int
		bits []int
	}{
		{362, []int{1, 3, 5, 6, 8}},
		{142, []int{1, 2, 3, 7}},
		{338, []int{1, 4, 6, 8}},
		{426, []int{1, 3, 5, 7, 8}},
	}

	for _, test := range testSet {
		get := bin.ConvBit2Slice(test.mask)
		if !reflect.DeepEqual(get, test.bits) {
			t.Errorf("\nexpext: %v (%b)\nget: %v\n", test.bits, test.mask, get)
		}
	}
}

func TestBoxID(t *testing.T) {
	testSet := []struct{ r, c, b int }{
		{2, 0, 0},
		{3, 5, 4},
		{4, 2, 3},
		{5, 8, 5},
		{6, 4, 7},
		{8, 7, 8},
	}
	b := Box{}
	for _, test := range testSet {
		if b.BoxID(test.c, test.r) != test.b {
			t.Errorf("get boxID: %d, expect: %d for row: %d and column: %d\n", b.BoxID(test.c, test.r), test.b, test.r, test.c)
		}
	}
}

func TestFirstXYBox(t *testing.T) {
	testSet := []struct{ x, y, b int }{
		{0, 0, 0},
		{3, 0, 1},
		{6, 0, 2},
		{3, 3, 4},
		{0, 3, 3},
		{6, 3, 5},
		{3, 6, 7},
		{6, 6, 8},
	}
	b := Box{}
	for _, test := range testSet {
		x, y := b.FirstXYBox(test.b)
		if x != test.x || y != test.y {
			t.Errorf("for boxID: %d, expect: (%d,%d) and get: (%d,%d)\n", test.b, test.x, test.y, x, y)
		}
	}
}

func TestXYBox(t *testing.T) {
	testSet := []struct{ x, y, b, i int }{
		{0, 0, 0, 0},
		{5, 2, 1, 8},
		{7, 2, 2, 7},
		{3, 4, 4, 3},
		{1, 3, 3, 1},
		{7, 4, 5, 4},
		{4, 6, 7, 1},
		{6, 8, 8, 6},
	}
	b := Box{}
	for _, test := range testSet {
		x, y := b.XYBox(test.b, test.i)
		if x != test.x || y != test.y {
			t.Errorf("for boxID: %d, post: %d, expect: (%d,%d) and get: (%d,%d)\n", test.b, test.i, test.x, test.y, x, y)
		}
	}
}

func TestBoxMinFreeCell(t *testing.T) {
	b := Box{}
	test := [9]int{341, 340, 410, 362, 142, 338, 426, 422, 402}
	bID := b.BoxMinFreeCell(test)
	if 0 != bID {
		t.Errorf("in init min free cell in almost all boxes is 4: first one (%b) id: 0; get: %d", test[bID], bID)
	}
	test[2] = 411
	bID = b.BoxMinFreeCell(test)
	if 2 != bID {
		t.Errorf("after we set on sell new box minimum - 3 cell (%b), id: 2; get: %d", test[bID], bID)
	}
}
