package main

import "testing"

func TestIsPortaitYes(t *testing.T) {
	b := &box{w: 10, l: 100}
	if !b.isPortrait() {
		t.Errorf("expected isPortrait to be true (w=%d, l=%d)\n", b.w, b.l)
	}
}

func TestIsPortraitNo(t *testing.T) {
	b := &box{w: 100, l: 10}
	if b.isPortrait() {
		t.Errorf("expected isPortrait to be false (w=%d, l=%d)\n", b.w, b.l)
	}
}

func TestIsPortraitRotate(t *testing.T) {
	// the box is initially portrait (l > w)
	b := &box{w: 10, l: 100}
	if !b.isPortrait() {
		t.Errorf("expected isPortrait to be true (w=%d, l=%d)\n", b.w, b.l)
	}

	b.rotate()

	// after rotating, the box should no longer be in portrait orientation
	if b.isPortrait() {
		t.Errorf("expected isPortrait to be false after rotation (w=%d, l=%d)\n", b.w, b.l)
	}
}

func TestSquareBoxOrientation(t *testing.T) {
	b := &box{w: 5, l: 5}
	// we don't really care whether a square is considered portrait or not,
	// we just care that we get the same answer before and after a rotate
	before := b.isPortrait()
	b.rotate()
	after := b.isPortrait()

	if before != after {
		t.Error("Square image orientation changed after a rotate")
	}
}

func TestShelfAddBoxTooWide(t *testing.T) {
	s := &shelf{Height: 1, MaxX: palletLength - 1}
	b := &box{l: 1, w: palletWidth * 2}
	if s.addBox(b) {
		t.Error("addBox should have failed (the box is wider than a pallet)")
	}
	if len(s.Boxes) != 0 {
		t.Errorf("shelf should be empty but has %d boxes\n", len(s.Boxes))
	}
}

func TestShelfAddBox(t *testing.T) {
	const h = 3
	s := &shelf{Height: h, MaxX: palletLength - 1}
	b := &box{l: 1, w: 1}
	succeeded := s.addBox(b)
	if !succeeded {
		t.Errorf("add to shelf (h=%d) should have succeeded (w=%d, l=%d)\n",
			s.Height, b.w, b.l)
	}
	if s.Height != h {
		t.Errorf("shelf height should not have changed (%d -> %d)\n", h, s.Height)
	}
}

func TestShelfAddMultipleBoxes(t *testing.T) {
	s := &shelf{Height: 1, MaxX: palletLength - 1}
	b0 := &box{l: 1, w: 1}
	b1 := &box{l: 1, w: palletWidth}
	b2 := &box{l: 1, w: palletWidth - b0.w}

	// first add should succeed
	if !s.addBox(b0) {
		t.Error("add first box failed")
	}
	// this box is too wide, add should fail
	if s.addBox(b1) {
		t.Error("add second box should have failed")
	}
	if s.NextY != 1 {
		t.Error("next X value didn't increment")
	}
	// there is just enough room to fit this box
	if !s.addBox(b2) {
		t.Error("add third box failed")
	}
	if len(s.Boxes) != 2 {
		t.Errorf("Expected 2 boxes on shelf, but got %d\n", len(s.Boxes))
	}
	// the shelf is full, subsequent adds should fail
	if s.addBox(b0) {
		t.Error("Add to full shelf succeeded but shoudn't have")
	}
}

// Ensure that an open shelf can grow when a box is added
func TestShelfAddBoxGrowHeight(t *testing.T) {
	const h = 1
	b0 := &box{w: 1, l: 2}
	b1 := &box{w: 2, l: 3}
	s := &shelf{Height: h, MaxX: palletLength - 1}
	if !s.addBox(b0) {
		t.Error("Add first box failed")
	}
	if s.Height != b0.l {
		t.Error("Shelf didn't grow after first box")
	}
	if !s.addBox(b1) {
		t.Error("Add second box failed")
	}
	if s.Height != b1.l {
		t.Error("Shelf didn't grow after second box")
	}
}

// Ensure that a closed shelf cannot grow in height
func TestClosedShelfDoesntGrow(t *testing.T) {
	s := &shelf{Height: 1, MaxX: palletLength - 1, Closed: true}
	b := &box{l: 2, w: 1}
	if s.addBox(b) {
		t.Error("Add should have failed")
	}
	if s.Height != 1 {
		t.Error("Closed shelf height should not have changed")
	}
}

func TestRepackerAddSingleBox(t *testing.T) {
	sr := shelfRepacker{}
	sr.reset()
	sr.addBox(box{w: 1, l: 1})
	if len(sr.Shelves) != 1 {
		t.Fatalf("Expected 1 shelf, but got %d\n", len(sr.Shelves))
	}
	if len(sr.Shelves[0].Boxes) != 1 {
		t.Fatalf("Expected 1 box in shelf, but got %d\n", len(sr.Shelves[0].Boxes))
	}
}

func TestRepackerAddUnitBoxes(t *testing.T) {
	sr := shelfRepacker{}
	sr.reset()
	const items = palletLength * palletWidth
	for i := 0; i < items; i++ {
		if !sr.addBox(box{id: uint32(i), w: 1, l: 1}) {
			t.Errorf("add box %d failed", i)
		}
	}
	p := pallet{boxes: sr.Boxes()}
	if err := p.IsValid(); err != nil {
		t.Error("Pallet invalid:", err)
	}
	if p.Items() != items {
		t.Errorf("Expected %d boxes, but got %d\n", items, p.Items())
	}
}

// Ensure that a new shelf is not added if there would not be room for the box
func TestShelfAddBoxNoMoreShelves(t *testing.T) {
	sr := shelfRepacker{}
	sr.reset()
	b := box{w: palletWidth / 4, l: 1}

	// fill the first shelf
	sr.addBox(b)
	sr.addBox(b)
	sr.addBox(b)
	sr.addBox(b)

	// almost fill the second shelf
	sr.addBox(b)
	sr.addBox(b)
	sr.addBox(b)

	if len(sr.Shelves) != 2 {
		t.Fatalf("Expected 2 shelves, but got %d\n", len(sr.Shelves))
	}

	// the next box won't fit on this shelf, and we don't have enough
	// room to add another shelf, so this next add should fail
	b.w, b.l = 3, 3
	if sr.addBox(b) {
		t.Fatal("Not enough room in pallet, but add succeeded anyway")
	}
}

func TestRepackerAddTwoBoxes(t *testing.T) {
	b0 := box{x: 2, y: 2, w: 1, l: 1, id: 2}
	b1 := box{x: 1, y: 2, w: 4, l: 2, id: 3}

	sr := shelfRepacker{}
	sr.reset()

	if !(sr.addBox(b0) && sr.addBox(b1)) {
		t.Errorf("couldn't add 2 boxes")
	}
	p := pallet{boxes: sr.Boxes()}
	if err := p.IsValid(); err != nil {
		t.Error("pallet invalid:", err)
		t.Log(p.String())
	}
}
