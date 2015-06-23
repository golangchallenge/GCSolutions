package main

import (
	"errors"
	"sync"
)

// A area represent a area on a pallet, with a starting point (x, y) and a size (w, l)
type area struct {
	x, y, w, l uint8
}

// prepareBox prepares a box to fit in the area.
// The size of the area has to be equal or greater than the box
func prepareBox(a area, b box) box {
	if a.l < b.l || a.w < b.w {
		b.w, b.l = b.l, b.w
	}
	b.x, b.y = a.x, a.y

	return b
}

// palletSpace represents the pallet thru an array
// true = position is used
// false = position is free
type palletSpace [palletWidth][palletLength]bool

// freeAreas returns an array of unused areas
func (ps *palletSpace) freeAreas() []area {
	freeAreas := []area{}

	for x := 0; x < palletWidth; x++ {
		var a *area
		for y := 0; y < palletLength; y++ {
			if ps[x][y] == false {
				if a == nil {
					a = &area{w: 1, l: 1, x: uint8(x), y: uint8(y)}
				} else {
					a.w++
				}
			}
		}
		if a != nil {
			if len(freeAreas) > 0 && freeAreas[len(freeAreas)-1].w == a.w && freeAreas[len(freeAreas)-1].y == a.y {
				freeAreas[len(freeAreas)-1].l++
			} else {
				freeAreas = append(freeAreas, *a)
			}
		}
	}
	return freeAreas
}

// setUsed marks the space from the box as used
func (ps *palletSpace) setUsed(b box) {
	for x := b.x; x < b.x+b.l; x++ {
		for y := b.y; y < b.y+b.w; y++ {
			ps[x][y] = true
		}
	}
}

// A repacker recives and repacks trucks. Ist can be used simultaneously from multiple goroutines, while using the exported functions.
// Important: palletWidth has to be greater or equal than palletLength.
type repacker struct {
	truckIDsMu sync.Mutex
	truckIDs   []int
	boxesMu    sync.Mutex
	boxes      uint32
	boxIDsMu   [palletWidth][palletLength]sync.Mutex
	boxIDs     [palletWidth][palletLength][]uint32
}

// discharge stores all truck IDs and boxes recived from channel in, till idLastTruck recived
func (r *repacker) discharge(in <-chan *truck) {
	for t := range in {
		if t.id == idLastTruck {
			break
		}
		r.PushTruckID(t.id)
		for _, p := range t.pallets {
			for _, b := range p.boxes {
				r.PushBox(b)
			}
		}
	}
}

// repacker repacks all stored boxes, repacks it on pallets and send it out with trucks.
// All boxes will be repacked and all trucks will send out
func (r *repacker) repack(out chan<- *truck) {
	pallets := []pallet{}

	for {
		if r.Boxes() == 0 {
			break
		}
		p := pallet{}
		freeAreas := []area{area{w: 4, l: 4, x: 0, y: 0}}
		ps := palletSpace{}

	restart:
		for {
			for _, a := range freeAreas {
				b, err := r.PopBox(a.w, a.l)
				if err != nil {
					continue
				}
				b = prepareBox(a, b)
				p.boxes = append(p.boxes, b)
				ps.setUsed(b)
				freeAreas = ps.freeAreas()
				continue restart
			}
			pallets = append(pallets, p)
			if id, err := r.PopTruckID(false); err == nil {
				out <- &truck{id: id, pallets: pallets}
				pallets = []pallet{}
			}
			break
		}
	}
	for id, err := r.PopTruckID(true); err == nil; id, err = r.PopTruckID(true) {
		out <- &truck{id: id, pallets: pallets}
		pallets = []pallet{}
	}
}

// Boxes returns the number of currently stored boxes
func (r *repacker) Boxes() uint32 {
	r.boxesMu.Lock()
	boxes := r.boxes
	r.boxesMu.Unlock()
	return boxes
}

// PushBox stores a box. The x and y value of the box get lost.
func (r *repacker) PushBox(b box) {
	if b.l > b.w {
		b.w, b.l = b.l, b.w
	}
	b.w--
	b.l--
	r.boxIDsMu[b.w][b.l].Lock()
	if len(r.boxIDs[b.w][b.l]) == 0 {
		r.boxIDs[b.w][b.l] = make([]uint32, 0, 10000)
	}
	r.boxIDs[b.w][b.l] = append(r.boxIDs[b.w][b.l], b.id)
	r.boxIDsMu[b.w][b.l].Unlock()

	r.boxesMu.Lock()
	r.boxes++
	r.boxesMu.Unlock()
}

// PopBox removes the first box with the size w, l or smaller and returns the box.
// If no box with size w, l or smaller available, an error will returned.
func (r *repacker) PopBox(w, l uint8) (box, error) {
	if l > w {
		w, l = l, w
	}
	for x := w; x > 0; x-- {
		if x < l {
			l = x
		}
		for y := l; y > 0; y-- {
			r.boxIDsMu[x-1][y-1].Lock()
			if len(r.boxIDs[x-1][y-1]) == 0 {
				r.boxIDsMu[x-1][y-1].Unlock()
				continue
			}
			id := r.boxIDs[x-1][y-1][0]
			r.boxIDs[x-1][y-1] = r.boxIDs[x-1][y-1][1:]
			r.boxIDsMu[x-1][y-1].Unlock()

			r.boxesMu.Lock()
			r.boxes--
			r.boxesMu.Unlock()
			return box{id: id, w: x, l: y}, nil
		}
	}
	return box{}, errors.New("No BoxID available")
}

// PushTruckID stores a truck ID
func (r *repacker) PushTruckID(id int) {
	r.truckIDsMu.Lock()
	r.truckIDs = append(r.truckIDs, id)
	r.truckIDsMu.Unlock()
}

// PopTruckID removes the first truck ID and returns it.
// If no more truck ID available an error will returned.
// With last = false one truck ID can be hold back for later.
func (r *repacker) PopTruckID(last bool) (int, error) {
	minLen := 0
	if !last {
		minLen++
	}
	r.truckIDsMu.Lock()
	if len(r.truckIDs) == minLen {
		r.truckIDsMu.Unlock()
		return 0, errors.New("No truck ID's available")
	}
	truckID := r.truckIDs[0]
	r.truckIDs = r.truckIDs[1:]
	r.truckIDsMu.Unlock()
	return truckID, nil
}

// newRepacker returns an pointer to a repacker Object, discharge the trucks and send out the repacked trucks
func newRepacker(in <-chan *truck, out chan<- *truck) *repacker {
	r := repacker{}

	go func() {
		r.discharge(in)
		r.repack(out)
		close(out)
	}()

	return &r
}
