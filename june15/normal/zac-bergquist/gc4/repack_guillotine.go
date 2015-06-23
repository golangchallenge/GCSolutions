package main

import (
	"fmt"
	"io"
)

// guillotineRepacker repacks trucks using a guillotine algorithm.
// The algorithm maintains a list of one or more "free rectangles"
// where boxes can be placed.
type guillotineRepacker struct {
	free  []rectangle
	boxes []box
}

// rectangle represents free space on a pallet
type rectangle box

func newGuillotineRepacker() *guillotineRepacker {
	// start with a single free rectangle (the size of the pallet)
	return &guillotineRepacker{
		free: []rectangle{{w: palletWidth, l: palletLength}},
	}
}

func (gr *guillotineRepacker) repack(in <-chan *truck, out chan<- *truck) {
	// for now, each input truck corresponds to a single output truck,
	// but we may use less (or more, but hoepfully less) pallets on the output
	for t := range in {
		// start a fresh pallet for each truck
		currentTruck := &truck{id: t.id}
		gr.reset()

		for _, p := range t.pallets {
			for _, b := range p.boxes {
				if !gr.addBox(&b) {
					// this box won't fit on the pallet
					// add the current pallet to the truck and start a new one
					currentTruck.pallets = append(currentTruck.pallets, pallet{boxes: gr.boxes})
					gr.reset()

					if !gr.addBox(&b) {
						panic("addBox failed on a fresh pallet")
					}
				}
			}
		}
		// flush the last pallet if necessary
		if len(gr.boxes) > 0 {
			currentTruck.pallets = append(currentTruck.pallets, pallet{boxes: gr.boxes})
		}

		out <- currentTruck
	}
	close(out)
}

// addBox adds a box to the current pallet.  It returns true if the box
// was succesfully added, or false if there wasn't room for the box.
func (gr *guillotineRepacker) addBox(b *box) bool {
	// determine which free rectangle to place b into
	// if no such rectangle, restart with new pallet
	idx := gr.getFreeRectangle(b)
	if idx == -1 {
		return false
	}

	rect := &gr.free[idx]

	// place box at bottom left of free rect
	b.x = rect.x + rect.l - b.l
	b.y = rect.y
	gr.boxes = append(gr.boxes, *b)

	// use guillotine split to divide free rect
	gr.splitFreeRect(idx, b)

	// the rectangle merge improvement helps to defragment
	// free space by merging free rectangles
	gr.merge()
	return true
}

// reset clears any state and prepares the repacker to start
// packing a new pallet
func (gr *guillotineRepacker) reset() {
	gr.boxes = make([]box, 0, 16)
	gr.free = []rectangle{{w: palletWidth, l: palletLength}}
}

// freeArea gets the amount of free space remaining on the current pallet.
func (gr *guillotineRepacker) freeArea() int {
	area := 0
	for bi := range gr.free {
		area += int(gr.free[bi].w * gr.free[bi].l)
	}
	return area
}

// splitFreeRect splits the free retangle at the specified index into
// two.  b is the box that was just added to the free rectangle.
func (gr *guillotineRepacker) splitFreeRect(index int, b *box) {
	// here we use the "min area split" rule to determine
	// which axis to cut along
	toparea := gr.free[index].w * (gr.free[index].l - b.l)
	rightarea := gr.free[index].l * (gr.free[index].w - b.w)

	if rightarea >= toparea {
		// split vertically
		right := rectangle{
			x: gr.free[index].x,
			y: gr.free[index].y + b.w,
			w: gr.free[index].w - b.w,
			l: gr.free[index].l,
		}

		// add the new right rectangle only if it has non-zero area
		if right.w > 0 && right.l > 0 {
			gr.free = append(gr.free, right)
		}

		// trim the original free rect to become the top
		// (we'll remove it if we end up shrinking it to 0 area)
		gr.free[index].w = b.w
		gr.free[index].l -= b.l

	} else {
		//split horiztonally
		top := rectangle{
			x: gr.free[index].x,
			y: gr.free[index].y,
			w: gr.free[index].w,
			l: gr.free[index].l - b.l,
		}

		// add the new top rectangle only if it has non-zero area
		if top.w > 0 && top.l > 0 {
			gr.free = append(gr.free, top)
		}

		// trim the original free rect to become the right
		// (we'll remove it if we end up shrinking it to 0 area)
		gr.free[index].l -= top.l
		gr.free[index].w -= b.w
		gr.free[index].y += b.w
		gr.free[index].x += top.l
	}

	// if by trimming the original free rectangle, we reduced it to size 0,
	// we can remove it from the set of free rectangles
	if gr.free[index].w == 0 || gr.free[index].l == 0 {
		gr.free = removeSliceElement(gr.free, index)
	}
}

// getFreeRectangle returns the idnex of the free rectangle to pack b into.
// It returns -1 if there is no free rectangle that can hold b.
// This method may rotate the box in order to find the best fit.
func (gr *guillotineRepacker) getFreeRectangle(b *box) int {
	// here we use a "best area fit" algorithm
	// (try to minimize the narrow strips of wasted space)
	minArea := uint8(palletWidth*palletLength) + 1
	minRect := -1
	needsrotate := false

	for ri := range gr.free {
		rect := &gr.free[ri]
		if rect.w >= b.w && rect.l >= b.l {
			// the box fits in its default orientation
			if a := uint8(rect.w * rect.l); a < minArea {
				needsrotate = false
				minArea = a
				minRect = ri
			}
		} else if rect.w >= b.l && rect.l >= b.w {
			// the box fits, but must be rotated first
			if a := uint8(rect.w * rect.l); a < minArea {
				needsrotate = true
				minArea = a
				minRect = ri
			}
		}
	}
	if needsrotate {
		b.rotate()
	}
	return minRect
}

// merge attempts to reduce framentation among free rectangles by
// merging smaller rectangles into larger ones.  This is important
// because the guillotine algorithm never places a box that would
// straddle two free rectangles.
func (gr *guillotineRepacker) merge() {
	for i := 0; i < len(gr.free); i++ {
		r0 := &gr.free[i]

		// if two free rectangles r0 and r1 can be merged,
		// merge r1 into r0, then remove r1
		for j := i + 1; j < len(gr.free); j++ {
			r1 := &gr.free[j]

			if r0.x == r1.x && r0.l == r1.l {
				// here we're joining two rectangles of equal height
				// the two subcases are for r0 on the left and r0 on the right
				if r0.y == r1.y+r1.w {
					// r0 is on the right, so it has to be shifted left
					r0.y = r1.y
					r0.w += r1.w
					gr.free = removeSliceElement(gr.free, j)
					j--
					r0 = &gr.free[i]
					r1 = &gr.free[j]
				} else if r0.y+r0.w == r1.y {
					// r0 is on the left, so just update its width
					r0.w += r1.w
					gr.free = removeSliceElement(gr.free, j)
					j--
					r0 = &gr.free[i]
					r1 = &gr.free[j]
				}
			} else if r0.y == r1.y && r0.w == r1.w {
				// here we join two rectangles of equal width
				// the two subcases are for r0 on the top and r0 on the bottom
				if r0.x == r1.x+r1.l {
					// r1 is on top, so r0 has to be shifted up
					r0.x = r1.x
					r0.l += r1.l
					gr.free = removeSliceElement(gr.free, j)
					j--
					r0 = &gr.free[i]
					r1 = &gr.free[j]
				} else if r0.x+r0.l == r1.x {
					// r0 is on top, so we just need to update its height
					r0.l += r1.l
					gr.free = removeSliceElement(gr.free, j)
					j--
					r0 = &gr.free[i]
					r1 = &gr.free[j]
				}
			}
		}
	}
}

func (gr *guillotineRepacker) writeFreeRects(out io.Writer) {
	for ri := range gr.free {
		rect := &gr.free[ri]
		fmt.Fprintf(out, "free rect @ (%d,%d) size (%dx%d)\n",
			rect.x, rect.y, rect.w, rect.l)
	}
}

// remove r[idx] from r
func removeSliceElement(r []rectangle, idx int) []rectangle {
	// swap with the last and remove the last one
	last := len(r) - 1
	if idx != last {
		r[idx], r[last] = r[last], r[idx]
	}
	out := r[:last]
	return out
}
