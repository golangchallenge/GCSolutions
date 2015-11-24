package main

import (
	"strings"
	"testing"
)

var sBoard = `1 _ 3 _ _ 6 _ 8 _
_ 5 _ _ 8 _ 1 2 _
7 _ 9 1 _ 3 _ 5 6
_ 3 _ _ 6 7 _ 9 _
5 _ 7 8 _ _ _ 3 _
8 _ 1 _ 3 _ 5 _ 7
_ 4 _ _ 7 8 _ 1 _
6 _ 8 _ _ 2 _ 4 _
_ 1 2 _ 4 5 _ 7 8`
var b Board

func TestMain(t *testing.T) {
	b := Board{}
	b.Read(strings.NewReader(sBoard))
	t.Log(b.String())
}

func TestRead(t *testing.T) {
	err := b.Read(strings.NewReader(sBoard))
	if err != nil {
		t.Errorf("%s, \n%s\n", err, b)
	}

	strBoard := `1 _ 3 _ _ 3 _ 8 _
_ 5 _ _ 8 _ 1 2 _
7 _ 9 1 _ 3 _ 5 6
_ 3 _ _ 6 7 _ 9 _
5 _ 7 8 _ _ _ 3 _
8 _ 1 _ 3 _ 5 _ 7
_ 4 _ _ 7 8 _ 1 _
6 _ 8 _ _ 2 _ 4 _
_ 1 2 _ 4 5 _ 7 8`
	err = b.Read(strings.NewReader(strBoard))
	errStr := "The value is duplicated with one in row, column or box: 3"
	if err == nil || strings.Compare(err.Error(), errStr) != -1 {
		t.Errorf("get: %s\nexpect: %s\n", err, errStr)
	}
}

func TestSetCell(t *testing.T) {
	testSet := []struct{ x, y, v int }{
		{1, 2, 8},
		{6, 3, 8},
		{1, 7, 7},
		{6, 0, 7},
		{3, 1, 7},
		{4, 0, 1},
	}
	for _, test := range testSet {
		err := b.SetCell(test.x, test.y, test.v)
		if test.x != 4 && err != nil {
			t.Errorf("error: %s, while set %d at (%d,%d)\n", err.Error(), test.v, test.x, test.y)
		}
		if test.x == 4 && err == nil {
			t.Errorf("expect error: The value (1) is duplicated with one in row #0, while set %d at (%d,%d)\n", test.v, test.x, test.y)
		}
	}
}

func TestFreeCellsBox(t *testing.T) {
	testSet := []struct{ v, b int }{
		{171, 1},
		{149, 3},
		{173, 5},
		{109, 8},
	}
	b.Rest()
	b.Read(strings.NewReader(sBoard))
	for _, test := range testSet {
		cellStatBox := b.FreeCellsBox(test.b)
		if test.v != cellStatBox {
			t.Errorf("board: %d expect: %b (%d) get: %b (%d)\n", test.b, test.v, test.v, cellStatBox, cellStatBox)
		}
	}
}

func TestLeftedBox(t *testing.T) {
	testSet := []struct {
		bSet []int
		v    int
	}{
		{
			[]int{0, 2, 1, 3, 8, 6},
			176,
		},
		{
			[]int{0, 1, 2, 3, 4},
			480,
		},
		{
			[]int{0, 2, 3, 5, 7},
			338,
		},
		{
			[]int{1, 2, 3, 4, 6, 7, 8},
			33,
		},
	}
	for _, test := range testSet {
		bMask := b.LeftedBox(test.bSet)
		if bMask != test.v {
			t.Errorf("expect: %b (%d) get: %b (%d)\n", test.v, test.v, bMask, bMask)
		}
	}
}

func TestGetBoxMask(t *testing.T) {
	testSet := []struct{ bID, m int }{
		{0, 341},
		{2, 410},
		{4, 142},
		{6, 426},
	}
	getFullCellInBox := func(boxID int) (mask int) {
		x, y := b.FirstXYBox(boxID)
		for i := range b.cell {
			for j := range b.cell[i] {
				if b.cell[i][j] == 0 || i < y || i > y+2 || j < x || j > x+2 {
					continue
				}
				cMask := (1 << uint((i%3)*3+j%3))
				t.Logf("for pos (%d,%d) with value: %d use mask %b (%d)", i, j, b.cell[i][j], cMask, cMask)
				mask |= cMask
			}
		}
		return
	}
	for _, test := range testSet {
		bMask := getFullCellInBox(test.bID)
		if bMask != test.m {
			t.Errorf("expect: %b (%d) get: %b (%d)\n", test.m, test.m, bMask, bMask)
		}
	}
}
