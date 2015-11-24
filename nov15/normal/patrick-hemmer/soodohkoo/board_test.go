package main

import (
	"fmt"
	"strings"
	"testing"
)

type testAlgorithm struct {
	f func(b *Board, changes []uint8) bool
}

func (a testAlgorithm) Name() string { return "testAlgorithm" }

func (a testAlgorithm) Stats() *AlgorithmStats { return &AlgorithmStats{} }

func (a testAlgorithm) EvaluateChanges(b *Board, changes []uint8) bool {
	if a.f != nil {
		return a.f(b, changes)
	}
	return true
}

func TestTileIsKnown(t *testing.T) {
	for i := uint(0); i < 9; i++ {
		tv := Tile(1 << i)
		if !tv.isKnown() {
			t.Errorf("Tile(%09b).isKnown is false, expected true", tv)
		}
		tv |= (1 << 2) | (1 << 3)
		if tv.isKnown() {
			t.Errorf("Tile(%09b).isKnown is true, expected false", tv)
		}
	}
}

func TestTileNum(t *testing.T) {
	for i := uint8(0); i < 9; i++ {
		tv := Tile(1 << i)
		if tv.Num() != i+1 {
			t.Errorf("Tile(%09b).Num() is %d, expected %d", tv, tv.Num(), i+1)
		}
	}

	tv := Tile((1 << 0) | (1 << 1))
	if tv.Num() != 0 {
		t.Errorf("Tile(%09b).Num() is %d, expected %d", tv, tv.Num(), 0)
	}
}

func TestSet(t *testing.T) {
	b := NewBoard()

	var algoCalled bool
	ta := testAlgorithm{}
	ta.f = func(b *Board, changes []uint8) bool {
		algoCalled = true
		return true
	}

	b.Algorithms = []Algorithm{ta}

	if !b.Set(0, Tile(1)) {
		t.Errorf("Set returned false, expected true")
	}
	if !algoCalled {
		t.Errorf("algorithm was not evaluated")
	}
}

func TestRegionIndices(t *testing.T) {
	ris := RegionIndices[7]
	expected := [9]uint8{57, 58, 59, 66, 67, 68, 75, 76, 77}
	if ris != expected {
		t.Errorf("RegionIndices[7] is %v, expected %v", ris, expected)
	}
}
func TestRowIndices(t *testing.T) {
	ris := RowIndices[1]
	expected := [9]uint8{9, 10, 11, 12, 13, 14, 15, 16, 17}
	if ris != expected {
		t.Errorf("RowIndices[1] is %v, expected %v", ris, expected)
	}
}
func TestColumnIndices(t *testing.T) {
	cis := ColumnIndices[2]
	expected := [9]uint8{2, 11, 20, 29, 38, 47, 56, 65, 74}
	if cis != expected {
		t.Errorf("ColumnIndices[1] is %v, expected %v", cis, expected)
	}
}
func TestMaskBits(t *testing.T) {
	tv := Tile((1 << 0) | (1 << 2) | (1 << 3) | (1 << 7))
	mbs := MaskBits[tv]
	expected := []uint8{0, 2, 3, 7}
	if fmt.Sprintf("%v", mbs) != fmt.Sprintf("%v", expected) {
		t.Errorf("MaskBits[%09b] is %v, expected %v", tv, mbs, expected)
	}
}

func TestSet_invalidBoard(t *testing.T) {
	b := NewBoard()

	ta := testAlgorithm{}
	ta.f = func(b *Board, changes []uint8) bool {
		return false
	}

	b.Algorithms = []Algorithm{ta}

	if b.Set(0, Tile(1)) {
		t.Errorf("Set returned true, expected false")
	}
	if b.Tiles[0] != tAny {
		t.Errorf("b.Tiles[0] is %09b, expected %09b", b.Tiles[0], tAny)
	}
}

func TestLittleSet(t *testing.T) {
	b := NewBoard()
	b.Algorithms = []Algorithm{}

	b.Tiles[30] = (1 << 0) | (1 << 2)

	// ensure setting to a value not a subset of existing value fails
	if b.set(30, (1<<1)) != false {
		t.Errorf("set returned true, expected false")
	}

	// ensure setting to a value that is a superset of existing value discards extra bits
	if b.set(30, (1<<2)|(1<<3)) != true {
		t.Errorf("set returned false, expected true")
	}
	if b.Tiles[30] != (1 << 2) {
		t.Errorf("b.Tiles[30] is %09b, expected %09b", b.Tiles[30], (1 << 2))
	}

	if b.changeSet[1] != 1<<3 {
		t.Errorf("b.changeSet[30] is %027b, expected %027b", b.changeSet[1], 1<<3)
	}
}

func TestEvaluateAlgorithms(t *testing.T) {
	alg1i := uint8(0)
	alg2i := uint8(3)
	alg1AfterAlg2 := false
	alg1 := testAlgorithm{func(b *Board, changes []uint8) bool {
		if alg1i < 3 {
			// sets b.Tiles[0]=1, b.Tiles[1]=2, b.Tiles[2]=3
			b.set(alg1i, Tile(1<<alg1i))
			alg1i++
		}
		if alg2i != 3 {
			// to make sure that alg1 is re-called after alg2 makes a change
			alg1AfterAlg2 = true
		}

		return true
	}}
	alg2 := testAlgorithm{func(b *Board, changes []uint8) bool {
		if alg1i != 3 {
			// we shouldn't have gotten to alg2 until alg1 stopped making changes
			t.Errorf("alg1i is %d, expected %d", alg1i, 3)
		}

		if alg2i < 5 {
			// sets b.Tiles[3]=4, b.Tiles[4]=5
			b.set(alg2i, Tile(1<<alg2i))
			alg2i++
		}

		return true
	}}
	alg3Calls := 0
	alg3 := testAlgorithm{func(b *Board, changes []uint8) bool {
		alg3Calls++
		return true
	}}

	b := NewBoard()
	b.Algorithms = []Algorithm{alg1, alg2, alg3}

	b.set(80, 9)
	if b.evaluateAlgorithms() != true {
		t.Errorf("b.evaluateAlgorithms is false, expected true")
	}

	for i := uint8(0); i < 5; i++ {
		if b.Tiles[i] != 1<<i {
			t.Errorf("b.Tiles[%d] is %09b, expected %09b", i, b.Tiles[i], 1<<i)
		}
	}

	if alg3Calls != 1 {
		t.Errorf("alg3Calls is %d, expected %d", alg3Calls, 1)
	}
}

func TestGuess(t *testing.T) {
	// very simple test, just make sure it generates a valid board
	b := NewBoard()
	if b.guess() != true {
		t.Errorf("b.guess() is false, expected true")
	}

	if !b.Solved() {
		t.Errorf("b.Solved() is false, expected true")
	}

	b2 := NewBoard()
	b2.Algorithms = []Algorithm{&algoKnownValueElimination{}}
	for ti, tv := range b.Tiles {
		if b2.Set(uint8(ti), tv) == false {
			t.Errorf("b2.Set(%d, %09b) is false, expected true", ti, tv)
		}
	}
}

func TestReadFrom(t *testing.T) {
	boardReader := strings.NewReader(`_ 8 _ _ 6 _ _ _ _
5 4 _ _ _ 7 _ 3 _
_ _ _ 1 _ _ 8 6 7
_ _ 9 _ 3 _ _ _ 6
_ _ 5 _ _ _ 3 _ _
3 _ _ _ 4 _ 2 _ _
7 5 4 _ _ 6 _ _ _
_ 2 _ 4 _ _ _ 7 9
_ _ _ _ 2 _ _ 8 _
`)
	b := NewBoard()
	n, err := b.ReadFrom(boardReader)
	if n != boardReader.Size() {
		t.Errorf("n is %d, expected %d", n, boardReader.Size())
	}
	if err != nil {
		t.Errorf("b.ReadFrom() returned error when none expected: %s", err)
	}

	// spot check the board
	checks := []struct {
		ti    uint8
		value uint8
	}{
		{1, 8},
		{4, 6},
		{9, 5},
		{38, 5},
		{54, 7},
		{79, 8},
	}
	for _, chk := range checks {
		if b.Tiles[chk.ti].Num() != chk.value {
			t.Errorf("b.Tiles[%d] is %d, expected %d", b.Tiles[chk.ti].Num(), chk.value)
		}
	}
}

func TestArt(t *testing.T) {
	boardString := `_ 8 _ _ 6 _ _ _ _
5 4 _ _ _ 7 _ 3 _
_ _ _ 1 _ _ 8 6 7
_ _ 9 _ 3 _ _ _ 6
_ _ 5 _ _ _ 3 _ _
3 _ _ _ 4 _ 2 _ _
7 5 4 _ _ 6 _ _ _
_ 2 _ 4 _ _ _ 7 9
_ _ _ _ 2 _ _ 8 _
`
	b := NewBoard()
	b.ReadFrom(strings.NewReader(boardString))
	art := b.Art()
	if string(art[:]) != boardString {
		t.Errorf("b.Art() does not match expected value\nb.Art(): %s\nExpected: %s\n",
			string(art[:]),
			boardString,
		)
	}
}

func BenchmarkSolve(b *testing.B) {
	// test the same board rotated a few different ways
	boardReader1 := strings.NewReader(`_ 8 _ _ 6 _ _ _ _
5 4 _ _ _ 7 _ 3 _
_ _ _ 1 _ _ 8 6 7
_ _ 9 _ 3 _ _ _ 6
_ _ 5 _ _ _ 3 _ _
3 _ _ _ 4 _ 2 _ _
7 5 4 _ _ 6 _ _ _
_ 2 _ 4 _ _ _ 7 9
_ _ _ _ 2 _ _ 8 _
`)
	boardReader2 := strings.NewReader(`_ _ 7 6 _ _ _ 9 _
_ 3 6 _ _ _ _ 7 8
_ _ 8 _ 3 2 _ _ _
_ 7 _ _ _ _ 6 _ _
6 _ _ 3 _ 4 _ _ 2
_ _ 1 _ _ _ _ 4 _
_ _ _ 9 5 _ 4 _ _
8 4 _ _ _ _ 5 2 _
_ 5 _ _ _ 3 7 _ _
`)
	boardReader3 := strings.NewReader(`_ 8 _ _ 2 _ _ _ _
9 7 _ _ _ 4 _ 2 _
_ _ _ 6 _ _ 4 5 7
_ _ 2 _ 4 _ _ _ 3
_ _ 3 _ _ _ 5 _ _
6 _ _ _ 3 _ 9 _ _
7 6 8 _ _ 1 _ _ _
_ 3 _ 7 _ _ _ 4 5
_ _ _ _ 6 _ _ 8 _
`)
	for i := 0; i < b.N; i++ {
		board := NewBoard()
		boardReader1.Seek(0, 0)
		board.ReadFrom(boardReader1)
		board.Solve()

		board = NewBoard()
		boardReader2.Seek(0, 0)
		board.ReadFrom(boardReader2)
		board.Solve()

		board = NewBoard()
		boardReader3.Seek(0, 0)
		board.ReadFrom(boardReader3)
		board.Solve()
	}
}
