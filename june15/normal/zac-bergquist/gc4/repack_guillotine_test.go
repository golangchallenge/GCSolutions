package main

import "testing"

func TestMergeEmptyPallet(t *testing.T) {
	gr := newGuillotineRepacker()
	gr.reset()
	if len(gr.free) != 1 {
		t.Fatalf("New repacker has %d free rects, expected 1\n", len(gr.free))
	}
	gr.merge()
	if len(gr.free) != 1 {
		t.Fatalf("Repacker has %d free rects after merge, expected 1\n", len(gr.free))
	}
}

func TestMergeIncompatible(t *testing.T) {
	gr := newGuillotineRepacker()
	gr.free = []rectangle{
		rectangle{x: 0, y: 0, w: 2, l: 2},
		rectangle{x: 2, y: 0, w: 1, l: 1},
	}
	gr.merge()
	if len(gr.free) != 2 {
		t.Fatalf("Repacker has %d free rects after merge, expected 2\n", len(gr.free))
	}
}

func TestMergeHorizontally(t *testing.T) {
	gr := newGuillotineRepacker()
	gr.free = []rectangle{
		rectangle{x: 0, y: 0, w: 2, l: 3},
		rectangle{x: 0, y: 2, w: 1, l: 3},
	}
	gr.merge()
	if l := len(gr.free); l != 1 {
		t.Errorf("Expected 1 free rect after merge, got %d\n", l)
	}

	// now try the reversing the order
	gr.reset()
	gr.free = []rectangle{
		rectangle{x: 0, y: 2, w: 1, l: 3},
		rectangle{x: 0, y: 0, w: 2, l: 3},
	}
	gr.merge()
	if l := len(gr.free); l != 1 {
		t.Errorf("Expected 1 free rect after merge, got %d\n", l)
	}
}

func TestMergeVertically(t *testing.T) {
	gr := newGuillotineRepacker()
	gr.free = []rectangle{
		rectangle{x: 1, y: 1, w: 3, l: 2},
		rectangle{x: 3, y: 1, w: 3, l: 1},
	}
	gr.merge()
	if l := len(gr.free); l != 1 {
		t.Errorf("Expected 1 free rect after merge, got %d\n", l)
	}

	// now try the reversing the order
	gr.reset()
	gr.free = []rectangle{
		rectangle{x: 3, y: 1, w: 3, l: 1},
		rectangle{x: 1, y: 1, w: 3, l: 2},
	}
	gr.merge()
	if l := len(gr.free); l != 1 {
		t.Errorf("Expected 1 free rect after merge, got %d\n", l)
	}
}

func TestMergeWithExtraRects(t *testing.T) {
	// make sure merge works when we have additional free rects
	gr := newGuillotineRepacker()
	gr.free = []rectangle{
		rectangle{x: 0, y: 0, w: 2, l: 3},
		rectangle{x: 3, y: 3, w: 1, l: 1}, // not involved in merge
		rectangle{x: 0, y: 2, w: 1, l: 3},
	}
	gr.merge()
	if l := len(gr.free); l != 2 {
		t.Errorf("Expected 2 free rect after merge, got %d\n", l)
	}
}

func TestAddBoxWithEmptyPallet(t *testing.T) {
	gr := newGuillotineRepacker()
	b := box{l: palletLength, w: palletWidth}

	if !gr.addBox(&b) {
		t.Fatal("addBox failed on empty pallet")
	}
	if l := len(gr.boxes); l != 1 {
		t.Errorf("Expected 1 box on pallet, got %d boxes\n", l)
	}
	if l := len(gr.free); l != 0 {
		t.Errorf("Expected 0 free retangles after adding box, got %d\n", l)
	}
	// the pallet should be full, so adding another box should fail
	b = box{l: 1, w: 1}
	if gr.addBox(&b) {
		t.Error("addBox should have failed with full pallet, but didn't")
	}
}

func TestEmptyPalletSplitCount(t *testing.T) {
	gr := newGuillotineRepacker()
	b := box{l: 1, w: 1}
	if !gr.addBox(&b) {
		t.Fatal("addBox failed on empty pallet")
	}
	// the empty space in the pallet should be split into
	// 2 free rectangles
	if l := len(gr.free); l != 2 {
		t.Errorf("Expected 2 free rects, got %d\n", l)
	}
}

func TestGuillotine(t *testing.T) {
	r := newGuillotineRepacker()
	in := make(chan *truck)
	out := make(chan *truck, 16)
	go r.repack(in, out)

	tid := 0
	bid := uint32(0)

	in <- &truck{
		id: tid,
		pallets: []pallet{{
			[]box{
				{x: 0, y: 0, w: 2, l: 1, id: bid},
				{x: 0, y: 0, w: 1, l: 2, id: bid},
				{x: 0, y: 0, w: 2, l: 1, id: bid},
				{x: 0, y: 0, w: 1, l: 2, id: bid},
			},
		}},
	}
	close(in)

	for tr := range out {
		for _, p := range tr.pallets {
			if err := p.IsValid(); err != nil {
				t.Errorf("Pallet packed incorrectly")
				t.Log(p.String())
			}
		}
	}
}

func TestRemoveSliceElementMiddle(t *testing.T) {
	r := []rectangle{
		{x: 0, y: 0, w: 1, l: 1},
		{x: 1, y: 1, w: 1, l: 1},
		{x: 2, y: 2, w: 1, l: 1},
	}
	rnew := removeSliceElement(r, 1)
	l0 := len(r)
	l1 := len(rnew)
	if l1 != l0-1 {
		t.Errorf("Expected length to decrease after removal (before %d, after %d)\n", l0, l1)
	}
	for _, rect := range rnew {
		if rect.x == 1 {
			t.Error("rectangle with x==1 should have been removed")
		}
	}
}

func TestRemoveSliceElementBeginning(t *testing.T) {
	r := []rectangle{
		{x: 0, y: 0, w: 1, l: 1},
		{x: 1, y: 1, w: 1, l: 1},
	}
	rnew := removeSliceElement(r, 0)
	l0 := len(r)
	l1 := len(rnew)
	if l1 != l0-1 {
		t.Errorf("Expected length to decrease after removal (before %d, after %d)\n", l0, l1)
	}
	for _, rect := range rnew {
		if rect.x == 0 {
			t.Error("rectangle with x==0 should have been removed")
		}
	}
}

func TestRemoveSliceElementEnd(t *testing.T) {
	r := []rectangle{
		{x: 0, y: 0, w: 1, l: 1},
		{x: 1, y: 1, w: 1, l: 1},
	}
	rnew := removeSliceElement(r, 1)
	l0 := len(r)
	l1 := len(rnew)
	if l1 != l0-1 {
		t.Errorf("Expected length to decrease after removal (before %d, after %d)\n", l0, l1)
	}
	for _, rect := range rnew {
		if rect.x == 1 {
			t.Error("rectangle with x==0 should have been removed")
		}
	}
}

func TestRemoveSliceElementEnd2(t *testing.T) {
	gr := newGuillotineRepacker()
	gr.free = []rectangle{
		{x: 0, y: 0, w: 1, l: 1},
		{x: 1, y: 1, w: 1, l: 1},
		{x: 2, y: 2, w: 1, l: 1},
	}
	l0 := len(gr.free)
	gr.free = removeSliceElement(gr.free, 2)

	l1 := len(gr.free)
	if l1 != l0-1 {
		t.Errorf("Expected length to decrease after removal (before %d, after %d)\n", l0, l1)
	}
	for _, rect := range gr.free {
		if rect.x == 2 {
			t.Error("rectangle with x==2 should have been removed")
		}
	}
}

func TestRemoveOnlySliceElement(t *testing.T) {
	r := []rectangle{
		{x: 0, y: 0, w: 1, l: 1},
	}
	rnew := removeSliceElement(r, 0)
	if len(rnew) != 0 {
		t.Errorf("Expected slice to be empty after removal, but size is %d\n", len(rnew))
	}
}
