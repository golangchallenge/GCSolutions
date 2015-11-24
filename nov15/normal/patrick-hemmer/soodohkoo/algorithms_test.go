package main

import (
	"strings"
	"testing"
)

func TestAlgoKnownValueElimination(t *testing.T) {
	b := NewBoard()
	b.set(0, Tile(1))
	a := algoKnownValueElimination{}
	if a.EvaluateChanges(&b, b.changes()) != true {
		t.Errorf("EvaluateChanges is false, expected true")
	}

	neighborSets := [][9]uint8{
		RowIndices[0],
		ColumnIndices[0],
		RegionIndices[0],
	}
	for _, ns := range neighborSets {
		for _, ti := range ns {
			if b.Tiles[ti] != ^Tile(1)&tAny {
				if ti == 0 {
					continue
				}
				t.Errorf("b.Tiles[%d] is %09b, expected %09b", ti, b.Tiles[ti], ^Tile(1)&tAny)
			}
		}
	}
}

func TestAlgoOnePossibleTile(t *testing.T) {
	b := NewBoard()
	b.ReadFrom(strings.NewReader(`1 2 3 _ _ _ _ _ _
4 5 6 _ _ _ _ _ _
7 8 _ _ _ _ _ _ _
_ _ _ _ _ _ _ _ _
_ _ _ _ _ _ _ _ _
_ _ _ _ _ _ _ _ _
_ _ _ _ _ _ _ _ _
_ _ _ _ _ _ _ _ _
_ _ _ _ _ _ _ _ _
`))

	a := algoOnePossibleTile{}
	if a.EvaluateChanges(&b, b.changes()) != true {
		t.Errorf("EvaluateChanges is false, expected true")
	}

	if b.Tiles[xyToIndex(2, 2)] != Tile(1<<8) {
		t.Errorf("b.Tiles[%d] is %09b, expected %09b", xyToIndex(2, 2), b.Tiles[xyToIndex(2, 2)], 1<<8)
	}
}

func TestAlgoOnlyRow(t *testing.T) {
	b := NewBoard()
	b.ReadFrom(strings.NewReader(`_ _ _ _ _ _ _ _ _
_ _ _ _ _ _ _ _ _
_ _ _ _ 1 _ _ _ _
_ _ _ 2 _ 3 _ _ _
_ _ _ _ _ _ _ _ _
_ _ _ 4 _ 5 _ _ _
_ _ _ _ _ _ _ _ _
_ _ _ _ _ _ _ _ _
_ _ _ _ _ _ _ _ _
`))
	// invoke algoKnownValueElimination as an easy way to prep the board
	(algoKnownValueElimination{}).EvaluateChanges(&b, b.changes())

	a := algoOnlyRow{}
	if a.EvaluateChanges(&b, b.changes()) != true {
		t.Errorf("EvaluateChanges is false, expected true")
	}

	for _, ti := range []uint8{36, 37, 38, 42, 43, 44} {
		if b.Tiles[ti]&1<<0 != 0 {
			t.Errorf("b.Tiles[%d] is %09b which includes %09b, expected it not to", ti, b.Tiles[ti], 1<<0)
		}
	}
}

func numsTile(nums ...uint8) Tile {
	tv := Tile(0)
	for _, n := range nums {
		tv |= 1 << (n - 1)
	}
	return tv
}

func TestAlgoNakedSubset(t *testing.T) {
	b := NewBoard()
	b.set(0, numsTile(4, 7))
	b.set(9, numsTile(1))
	b.set(27, numsTile(2, 4, 6, 7))
	b.set(36, numsTile(4, 7))
	b.set(45, numsTile(2, 4, 7))
	b.set(63, numsTile(1, 2, 3, 4, 6, 7))

	a := algoNakedSubset{}
	if a.EvaluateChanges(&b, b.changes()) != true {
		t.Errorf("EvaluateChanges is false, expected true")
	}

	for _, ti := range []uint8{0, 36} {
		if b.Tiles[ti] != numsTile(4, 7) {
			t.Errorf("b.Tiles[%d] is %09b, expected %09b", ti, b.Tiles[ti], numsTile(4, 7))
		}
	}
	for _, ti := range []uint8{9, 27, 45, 63} {
		if b.Tiles[ti]&numsTile(4, 7) != 0 {
			t.Errorf("b.Tiles[%d] is %09b which includes %09b, expected it not to",
				ti,
				b.Tiles[ti],
				numsTile(4, 7),
			)
		}
	}
}

func TestAlgoHiddenSubset(t *testing.T) {
	b := NewBoard()
	// remove the values 1 & 4 from all tiles in row 0 except tiles 0 & 1
	for ti := uint8(2); ti <= 9; ti++ {
		b.set(ti, ^numsTile(1, 4))
	}

	a := algoHiddenSubset{}
	if a.EvaluateChanges(&b, b.changes()) != true {
		t.Errorf("EvaluateChanges is false, expected true")
	}

	for _, ti := range []uint8{0, 1} {
		if b.Tiles[ti] != numsTile(1, 4) {
			t.Errorf("b.Tiles[%d] is %09b, expected %09b", ti, b.Tiles[ti], numsTile(1, 4))
		}
	}
}
