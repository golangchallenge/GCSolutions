package main

import (
	"sync"
)

// Small pallet size allows us to make a seperate pile for each possible box size.
type repacker struct {
	boxes4x4   []uint32
	boxes4x3   []uint32
	boxes3x3   []uint32
	boxes4x2   []uint32
	boxes3x2   []uint32
	boxes4x1   []uint32
	boxes2x2   []uint32
	boxes3x1   []uint32
	boxes2x1   []uint32
	boxes1x1   []uint32
	boxesReady uint32
}

const numDocks = 4

func newRepacker(in <-chan *truck, out chan<- *truck) {
	wg := sync.WaitGroup{}
	// Lets say our warehouse has four docks. Each one manages itself independently.
	manageDock := func() {
		defer wg.Done()
		r := repacker{}
		done := false
		pals := []pallet{}
		var firstTruck *truck
		for {
			t := <-in
			if t == nil || t.id == idLastTruck {
				done = true
			} else {
				r.Unload(t)
				if firstTruck == nil {
					// Remember first truck we see on this dock. We will load it up at the end with all pallets.
					firstTruck = t
				} else {
					// All other trucks get send out empty
					t.pallets = nil
					out <- t
				}
			}

			// once we are done receiving trucks, pack all the boxes onto the last truck and send it off.
			if done {
				for r.boxesReady > 0 {
					p := r.PackPallet()
					pals = append(pals, *p)
				}
				firstTruck.pallets = pals
				out <- firstTruck
				return
			}
		}
	}
	for i := 0; i < numDocks; i++ {
		wg.Add(1)
		go manageDock()
	}
	go func() {
		wg.Wait()
		close(out)
	}()
}

func (r *repacker) Unload(t *truck) {
	for _, p := range t.pallets {
		for _, b := range p.boxes {
			r.boxesReady++
			switch b.w * b.l {
			case 4 * 4:
				r.boxes4x4 = append(r.boxes4x4, b.id)
			case 4 * 3:
				r.boxes4x3 = append(r.boxes4x3, b.id)
			case 3 * 3:
				r.boxes3x3 = append(r.boxes3x3, b.id)
			case 4 * 2:
				r.boxes4x2 = append(r.boxes4x2, b.id)
			case 3 * 2:
				r.boxes3x2 = append(r.boxes3x2, b.id)
			case 4:
				// 2x2 and 4x1 have same area.
				if b.w == 2 {
					r.boxes2x2 = append(r.boxes2x2, b.id)
				} else {
					r.boxes4x1 = append(r.boxes4x1, b.id)
				}
			case 3 * 1:
				r.boxes3x1 = append(r.boxes3x1, b.id)
			case 2 * 1:
				r.boxes2x1 = append(r.boxes2x1, b.id)
			case 1 * 1:
				r.boxes1x1 = append(r.boxes1x1, b.id)
			default:
				panic("Unrecognized box size")
			}
		}
	}
}

// add a box to the given pallet at the specified coordinates
// take the id from the front of the id list and cut off the used element
// pointer used to avoid repetition at the call site.
func (r *repacker) addBox(p *pallet, x, y, w, l uint8, ids *[]uint32) {
	p.boxes = append(p.boxes, box{x, y, l, w, (*ids)[0]})
	r.boxesReady--
	*ids = (*ids)[1:]
}

// This approach takes advantage of the fact that 4x4 is just small enough
// to have a fairly finite number of possible arrangements.
// I abandon the generic approach to pack 4x4 pallets in a very fast and near-optimal way.
func (r *repacker) PackPallet() *pallet {
	pal := &pallet{[]box{}}

	if len(r.boxes4x4) > 0 {
		r.addBox(pal, 0, 0, 4, 4, &r.boxes4x4)
	} else if len(r.boxes4x3) > 0 {
		// XXXX
		// XXXX
		// XXXX
		// 1111
		r.addBox(pal, 0, 0, 4, 3, &r.boxes4x3)
		r.Pack4x1(pal, 3)
	} else if len(r.boxes3x3) > 0 {
		// XXX1
		// XXX1
		// XXX1
		// 2222
		r.addBox(pal, 0, 0, 3, 3, &r.boxes3x3)
		r.Pack1x3(pal)
		r.Pack4x1(pal, 3)
	} else if len(r.boxes4x2) > 0 {
		// XXXX
		// XXXX
		// 1111
		// 1111
		r.addBox(pal, 0, 0, 4, 2, &r.boxes4x2)
		r.Pack4x2(pal)
	} else if len(r.boxes3x2) > 0 {
		// XXX1
		// XXX1
		// 2222
		// 2222
		r.addBox(pal, 0, 0, 3, 2, &r.boxes3x2)
		r.Pack1x2(pal, 3, 0)
		r.Pack4x2(pal)
	} else if len(r.boxes2x2) > 0 {
		// XX11
		// XX11
		// 2222
		// 2222
		r.addBox(pal, 0, 0, 2, 2, &r.boxes2x2)
		r.Pack2x2(pal, 2, 0)
		r.Pack4x2(pal)
	} else { //everything else fits in a single row
		// 1111
		// 2222
		// 3333
		// 4444
		r.Pack4x1(pal, 0)
		r.Pack4x1(pal, 1)
		r.Pack4x1(pal, 2)
		r.Pack4x1(pal, 3)
	}
	return pal
}

// Pack the bottom half of the pallet
func (r *repacker) Pack4x2(pal *pallet) {
	if len(r.boxes4x2) > 0 {
		r.addBox(pal, 0, 2, 4, 2, &r.boxes4x2)
	} else if len(r.boxes3x2) > 0 {
		// XXX1
		// XXX1
		r.addBox(pal, 0, 2, 3, 2, &r.boxes3x2)
		r.Pack1x2(pal, 3, 2)
	} else if len(r.boxes2x2) > 0 {
		// XX11
		// XX11
		r.addBox(pal, 0, 2, 2, 2, &r.boxes2x2)
		r.Pack2x2(pal, 2, 2)
	} else {
		// 1111
		// 2222
		r.Pack4x1(pal, 2)
		r.Pack4x1(pal, 3)
	}
}

// Pack a single row
func (r *repacker) Pack4x1(pal *pallet, y uint8) {
	if len(r.boxes4x1) > 0 {
		r.addBox(pal, 0, y, 4, 1, &r.boxes4x1)
	} else if len(r.boxes3x1) > 0 {
		// XXX1
		r.addBox(pal, 0, y, 3, 1, &r.boxes3x1)
		r.Pack1x1(pal, 3, y)
	} else if len(r.boxes2x1) > 0 {
		// XX11
		r.addBox(pal, 0, y, 2, 1, &r.boxes2x1)
		r.Pack2x1(pal, 2, y)
	} else if len(r.boxes1x1) > 0 {
		// X123
		r.addBox(pal, 0, y, 1, 1, &r.boxes1x1)
		r.Pack1x1(pal, 1, y)
		r.Pack1x1(pal, 2, y)
		r.Pack1x1(pal, 3, y)
	}
}

// Pack top-right 1x3 block
func (r *repacker) Pack1x3(pal *pallet) {
	if len(r.boxes3x1) > 0 {
		r.addBox(pal, 3, 0, 1, 3, &r.boxes3x1)
	} else if len(r.boxes2x1) > 0 {
		// X
		// X
		// 1
		r.addBox(pal, 3, 0, 1, 2, &r.boxes2x1)
		r.Pack1x1(pal, 3, 2)
	} else if len(r.boxes1x1) > 0 {
		// X
		// 1
		// 2
		r.addBox(pal, 3, 0, 1, 1, &r.boxes1x1)
		r.Pack1x1(pal, 3, 1)
		r.Pack1x1(pal, 3, 2)
	}
}

// Pack a vertical 1x2 region
func (r *repacker) Pack1x2(pal *pallet, x, y uint8) {
	if len(r.boxes2x1) > 0 {
		r.addBox(pal, x, y, 1, 2, &r.boxes2x1)
	} else if len(r.boxes1x1) > 0 {
		// X
		// 1
		r.addBox(pal, x, y, 1, 1, &r.boxes1x1)
		r.Pack1x1(pal, x, y+1)
	}
}

// Pack a 2x2 square
func (r *repacker) Pack2x2(pal *pallet, x, y uint8) {
	if len(r.boxes2x2) > 0 {
		r.addBox(pal, x, y, 2, 2, &r.boxes2x2)
	} else {
		// 11
		// 22
		r.Pack2x1(pal, x, y)
		r.Pack2x1(pal, x, y+1)
	}
}

// Pack a 2x1 region with the given coordinates
func (r *repacker) Pack2x1(pal *pallet, x, y uint8) {
	if len(r.boxes2x1) > 0 {
		// XX
		r.addBox(pal, x, y, 2, 1, &r.boxes2x1)
	} else if len(r.boxes1x1) > 0 {
		// X1
		r.addBox(pal, x, y, 1, 1, &r.boxes1x1)
		r.Pack1x1(pal, x+1, y)
	}
}

//Pack a 1x1 square if possible
func (r *repacker) Pack1x1(pal *pallet, x, y uint8) {
	if len(r.boxes1x1) > 0 {
		r.addBox(pal, x, y, 1, 1, &r.boxes1x1)
	}
}
