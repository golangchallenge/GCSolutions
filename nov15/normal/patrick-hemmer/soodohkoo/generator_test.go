package main

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
)

func TestAlgoGeneratorShuffle(t *testing.T) {
	a := algoGenerateShuffle{rand.New(rand.NewSource(0))}
	mbs := fmt.Sprintf("%v", MaskBits[numsTile(1, 3, 5, 7)])
	a.EvaluateChanges(nil, nil)
	mbs2 := fmt.Sprintf("%v", MaskBits[numsTile(1, 3, 5, 7)])
	if mbs == mbs2 {
		t.Errorf("MaskBits[%09b] did not change", numsTile(1, 3, 5, 7))
	}
}

func TestNewRandomBoard(t *testing.T) {
	b := NewRandomBoard(5)
	unknownTiles := 0
	for _, tv := range b.Tiles {
		if !tv.isKnown() {
			unknownTiles++
		}
	}
	if unknownTiles != 5 {
		t.Errorf("unknownTiles is %d, expected %d", unknownTiles, 5)
	}

	if !b.Solve() {
		t.Errorf("b.Solve() is false, expected true")
	}
}

func TestDropRandomTile(t *testing.T) {
	b := NewBoard()
	b.ReadFrom(strings.NewReader(`8 1 _ _ _ _ _ _ _
_ _ 3 6 _ _ _ _ _
_ 7 _ _ 9 _ 2 _ _
_ 5 _ _ _ 7 _ _ _
_ _ _ _ 4 5 7 _ _
_ _ _ 1 _ _ _ 3 _
_ _ 1 _ _ _ _ 6 8
_ _ 8 5 _ _ _ 1 _
_ 9 _ _ _ _ 4 _ _
`))
	if b.dropRandomTile(rand.New(rand.NewSource(0))) == false {
		// the tile at 1,0 can be dropped
		t.Errorf("b.dropRandomTile() is false, expected true")
	}

	if b.dropRandomTile(rand.New(rand.NewSource(0))) == true {
		// the tile at 1,0 was the only tile that can be dropped
		t.Errorf("b.dropRandomTile() is true, expected false")
	}
}
