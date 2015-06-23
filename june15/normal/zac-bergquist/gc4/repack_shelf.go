package main

import (
	"bytes"
	"fmt"
	"io"
)

// shelfRepacker repacks trucks using a simple shelf first-fit algorithm.
type shelfRepacker struct {
	Shelves []shelf
}

func (sr *shelfRepacker) repack(in <-chan *truck, out chan<- *truck) {
	for t := range in {
		sr.reset()
		outt := &truck{id: t.id}
		for _, p := range t.pallets {
			for _, b := range p.boxes {
				// add boxes to a pallet until we find a box that won't fit,
				// then add that pallet to the truck and start a new pallet
				if !sr.addBox(b) {
					outt.pallets = append(outt.pallets, pallet{boxes: sr.Boxes()})
					sr.reset()
					// addBox should always succeed when starting a new pallet
					if !sr.addBox(b) {
						panic("addBox failed on a fresh pallet")
					}
				}
			}
		}
		// flush the last pallet if necessary
		last := sr.Boxes()
		if len(last) > 0 {
			outt.pallets = append(outt.pallets, pallet{boxes: last})
		}
		out <- outt
	}
	close(out)
}

func writeShelf(out io.Writer, s *shelf) {
	for _, b := range s.Boxes {
		fmt.Fprintf(out, "box%02d ", b.id)
	}
	fmt.Fprintf(out, "\n")
}

func (sr *shelfRepacker) String() string {
	var buf bytes.Buffer
	for i := len(sr.Shelves) - 1; i >= 0; i-- {
		writeShelf(&buf, &sr.Shelves[i])
	}
	return string(buf.Bytes())
}

func (sr *shelfRepacker) Boxes() []box {
	b := make([]box, 0, 256)
	for _, s := range sr.Shelves {
		b = append(b, s.Boxes...)
	}
	return b
}

func (sr *shelfRepacker) addBox(b box) bool {
	// first try fitting the box on a previously closed shelf
	// (this is the shelf first-fit algorithm)
	if l := len(sr.Shelves); l > 1 {
		for si := range sr.Shelves[:l-1] {
			if sr.Shelves[si].addBox(&b) {
				return true
			}
		}
	}

	// now try fitting the box on the top-most shelf
	s := &sr.Shelves[len(sr.Shelves)-1]
	if s.Closed {
		panic("last shelf should be open!!")
	}
	s.orientBox(&b)
	if !s.addBox(&b) {
		// the box didn't fit - close the current shelf
		s.Closed = true

		// if we opened a new shelf, would the box fit?
		maxHeight := 1 + s.MaxX - s.Height
		if b.l > maxHeight {
			// don't bother opening a new shelf (it wouldn't be tall enough)
			// [ this pallet is now complete ]
			return false
		}
		// make a new shelf
		s = &shelf{Height: 1, MaxX: s.MaxX - s.Height}
		if !s.addBox(&b) {
			panic("addBox on new shelf should succeed")
		}
		sr.Shelves = append(sr.Shelves, *s)
	}
	return true
}

// reset clears internal state and prepares to start packing a new pallet.
func (sr *shelfRepacker) reset() {
	// start with a single shelf
	sr.Shelves = make([]shelf, 1, palletLength)
	sr.Shelves[0] = shelf{Height: 1, MaxX: palletLength - 1}
}

// shelf is a sub-rectangle of a pallet.  The free area of the
// pallet is organized into shelves, bottom-up, in which the
// boxes are placed left to right.
type shelf struct {
	Boxes  []box
	Height uint8
	// MaxX is the X coordinate for the floor of the shelf.
	// Shelves are filled bottom up, and the bottom cells have
	// the highest X values.  (This is a strange coordinate
	// system - [0,0] is the top left corner, but x increases
	// as you move down and y increases as you move right
	MaxX  uint8
	NextY uint8
	// A shelf is either open or closed.  Only one shelf in a pallet
	// can be open, and it is always the top-most shelf.  The height
	// of the open shelf can be adjusted whenever a box is placed
	// on that shelf, because there are no shelves above it.
	Closed bool
}

// addBox attempts to add a box to a shelf.  The box should already be in the
// desired orientation.  If the shelf is open, its height may be modified to
// accomodate for the box.  This method returns true if the box was added.
func (s *shelf) addBox(b *box) bool {
	// is the box too wide?
	if s.NextY > palletWidth || b.w > palletWidth-s.NextY {
		return false
	}
	// increase height of shelf the shelf is open and we can
	// make it big enough to fit the box
	if !s.Closed && s.Height < b.l {
		newHeight := min(b.l, s.MaxX+1)
		if b.l > newHeight {
			// the box wouldn't fit even if we grew the shelf,
			// so don't bother
			return false
		}
		s.Height = newHeight
	}

	// is the box too tall (long)?
	if b.l > s.Height {
		return false
	}

	// the box fits!
	// (s.MaxX, s.NextY) is the coordinates for the bottom left of the box,
	// we need to convert that to the top left
	b.x, b.y = s.MaxX+1-b.l, s.NextY
	s.NextY += b.w
	s.Boxes = append(s.Boxes, *b)
	return true
}

// orientBox orients a box prior to placing it on a shelf.
func (s *shelf) orientBox(b *box) {
	b.x, b.y = 0, 0
	// if this will be the first box on a new open shelf,
	// store it sideways (to minimize the height of the new shelf)
	if !s.Closed && len(s.Boxes) == 0 {
		if !b.isPortrait() {
			b.rotate()
		}
		return
	}
	// if the rectangle fits upright, store it so
	// (to minimize wasted area between the top of the box and the shelf ceiling)
	if max(b.w, b.l) < s.Height {
		if b.isPortrait() {
			b.rotate()
		}
		return
	}
	// otherwise, store the rectangle sideways
	if !b.isPortrait() {
		b.rotate()
	}
}
