package main

import (
	"runtime"
	"sync"
)

// Add tries to add a box to the pallet. Returns true if successful.
//
// Add will try rotating, and moving the box around.
func (p *pallet) Add(b box) bool {
	if p.addNoRotate(b) {
		return true
	}
	l := b.l
	b.l = b.w
	b.w = l
	return p.addNoRotate(b)
}
func (p *pallet) addNoRotate(b box) bool {
	for b.y = 0; b.y+b.w <= palletLength; b.y++ {
	X:
		for b.x = 0; b.x+b.l <= palletWidth; b.x++ {
			for i := range p.boxes {
				if p.boxes[i].Overlap(b) {
					continue X
				}
			}

			p.boxes = append(p.boxes, b)
			return true
		}
	}

	return false
}

// IsFull returns true if all spaces on the pallet are occupied.
func (p pallet) IsFull() bool {
	size := palletWidth * palletLength
	used := 0
	for i := range p.boxes {
		used += int(p.boxes[i].l) * int(p.boxes[i].w)
	}
	return size == used
}

// Overlap returns true if the source box overlaps with the given box.
func (b box) Overlap(b2 box) bool {
	if b2.x >= b.x+b.l || b2.x+b2.l <= b.x {
		// b2's left edge is right of b's right edge
		// or b2's right edge is left of b's left edge
		return false
	}

	if b2.y >= b.y+b.w || b2.y+b2.w <= b.y {
		// b2's top edge is below b's bottom edge
		// or b2's bottom edge is above b's top edge
		return false
	}

	return true
}

// repacker manages the process of accepting incoming trucks, and repacking the
// boxes on those trucks to more efficiently use pallets.
type repacker struct {
	partialPallets []*pallet // pointers are measurably faster
	fullPallets    []*pallet
}

// newRepacker constructs a new repacker group.
// It takes a chan of trucks to read from, and a chan of trucks to send out.
// It will spawn off GOMAXPROCS number of goroutine workers.
func newRepacker(in <-chan *truck, out chan<- *truck) {
	wg := &sync.WaitGroup{}
	for i := 0; i < runtime.GOMAXPROCS(0); i++ {
		newRepackerWorker(in, out, wg)
	}
}
func newRepackerWorker(in <-chan *truck, out chan<- *truck, wg *sync.WaitGroup) *repacker {
	rp := &repacker{}
	wg.Add(1)
	go rp.repack(in, out, wg)
	return rp
}

// add finds a pallet to place a box on. If the box cannot fit on an existing
// pallet, a new one is created.
func (rp *repacker) add(b box) {
	for i, p := range rp.partialPallets {
		if !p.Add(b) {
			continue
		}

		if p.IsFull() {
			rp.fullPallets = append(rp.fullPallets, p)
			rp.partialPallets = append(rp.partialPallets[:i], rp.partialPallets[i+1:]...)
		}
		return
	}

	// can't fix on an existing pallet. create one
	b.x = 0
	b.y = 0
	p := &pallet{
		boxes: []box{b},
	}
	if p.IsFull() {
		rp.fullPallets = append(rp.fullPallets, p)
	} else {
		rp.partialPallets = append(rp.partialPallets, p)
	}
}

const excessPalletsMax = 10

// getExcessPallets removes all but 10 partial pallets, and returns the ones
// removed.
func (rp *repacker) getExcessPallets() []pallet {
	n := len(rp.partialPallets) - excessPalletsMax
	if n <= 0 {
		return []pallet(nil)
	}

	pallets := make([]pallet, n)
	for i, p := range rp.partialPallets[:n] {
		pallets[i] = *p
	}
	rp.partialPallets = rp.partialPallets[n:]
	return pallets
}

// repack starts the repacking process.
func (rp *repacker) repack(in <-chan *truck, out chan<- *truck, wg *sync.WaitGroup) {
	var prevTruck *truck
	for t := range in {
		if prevTruck != nil {
			// We hold onto the previous truck and release it when a new one comes in.
			// This is so that when the channel is closed, we can put all partial
			// pallets onto the last truck.
			out <- prevTruck
		}
		for _, p := range t.pallets {
			for i := range p.boxes {
				rp.add(p.boxes[i])
			}
		}

		t.pallets = make([]pallet, len(rp.fullPallets))
		for i := range rp.fullPallets {
			t.pallets[i] = *rp.fullPallets[i]
		}
		rp.fullPallets = ([]*pallet)(nil)

		t.pallets = append(t.pallets, rp.getExcessPallets()...)

		prevTruck = t
	}

	if len(rp.partialPallets) > 0 {
		for _, p := range rp.partialPallets {
			prevTruck.pallets = append(prevTruck.pallets, *p)
		}
		out <- prevTruck
	}

	wg.Done()

	if prevTruck != nil {
		if prevTruck.id == idLastTruck {
			// we have truck 0.
			// wait for all other repackers to finish, then close the channel.
			wg.Wait()
			close(out)
		}
	}
}
