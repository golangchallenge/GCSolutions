package main

// A repacker repacks trucks.
type repacker struct {
	boxes    [][]box
	numBoxes uint64
}

func MakeRepacker() *repacker {
	// Initialize box storage with slots for each unique box shape
	numBoxShapes := (boxesIndex(palletWidth, palletLength) + 1)
	return &repacker{
		boxes:    make([][]box, numBoxShapes),
		numBoxes: 0,
	}
}

// To get the index of a box, calculate sum(0..w-1) + l-1
// There will be boxesIndex(palletWidth, palletLength) + 1 unique box shapes,
// once boxes are rotated such that the longest side is the width.
func boxesIndex(w, l uint8) uint8 {
	return ((w - 1) * ((w - 1) + 1) / 2) + (l - 1)
}

func (r *repacker) unloadTruck(t *truck) {
	for _, p := range t.pallets {
		for _, b := range p.boxes {
			if b.w < b.l {
				b.l, b.w = b.w, b.l
			}
			i := boxesIndex(b.w, b.l)
			r.boxes[i] = append(r.boxes[i], b)
			r.numBoxes += 1
		}
	}
}

func (r *repacker) takeLargestAvailableBox(maxW, maxL uint8) *box {
	for w := maxW; w > 0; w-- {
		for l := maxL; l > 0; l-- {
			if b := r.takeNextBox(w, l); b != nil {
				return b
			}
		}
	}
	return nil
}

func (r *repacker) takeNextBox(w, l uint8) *box {
	// boxes are stored horizontally. if a vertical box is requested, flip
	// the dimensions, and then rotate the box before returning
	flip := w < l
	if flip {
		w, l = l, w
	}

	i := boxesIndex(w, l)
	if len(r.boxes[i]) > 0 {
		// remove the first box from the boxes of that size
		b := r.boxes[i][0]
		r.boxes[i] = r.boxes[i][1:]
		r.numBoxes -= 1
		if flip {
			b.w, b.l = b.l, b.w
		}
		return &b
	}
	return nil
}

func (r *repacker) fillSpace(p *pallet, w, l, y, x uint8) {
	// grab the largest available box
	b := r.takeLargestAvailableBox(w, l)

	// if no box was found, there either aren't any pallets left, or this
	// space isn't fillable with what pallets are available
	if b == nil {
		return
	}

	// add box to pallet
	b.y, b.x = y, x
	p.boxes = append(p.boxes, *b)

	if b.w < w {
		// fill space to the right of the placed box (up to the length
		// of the placed box)
		r.fillSpace(p, w-b.w, b.l, y+b.w, x)
	}

	if b.l < l {
		// fill space below the placed box (including the space
		// below-right, beyond the length of the placed box)
		r.fillSpace(p, w, l-b.l, y, x+b.l)
	}
}

func (r *repacker) packEverything(id int) *truck {
	var pallets []pallet

	for r.numBoxes > 0 {
		p := &pallet{}
		r.fillSpace(p, palletWidth, palletLength, 0, 0)
		pallets = append(pallets, *p)
	}

	return &truck{id: id, pallets: pallets}
}

func newRepacker(in <-chan *truck, out chan<- *truck) *repacker {
	r := MakeRepacker()
	go func() {
		for t := range in {
			// The last truck is indicated by its id. You might
			// need to do something special here to make sure you
			// send all the boxes.
			r.unloadTruck(t)
			if t.id == idLastTruck {
				out <- r.packEverything(t.id)
			} else {
				out <- &truck{id: t.id}
			}

		}
		// The repacker must close channel out after it detects that
		// channel in is closed so that the driver program will finish
		// and print the stats.
		close(out)
	}()
	return r
}
