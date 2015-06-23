package main

import (
	"sync"
	"testing"
)

func TestPalletAdd(t *testing.T) {
	p := pallet{
		boxes: []box{
			{x: 0, y: 0, l: 4, w: 3},
			{x: 0, y: 3, l: 2, w: 1},
		},
	}

	if ok := p.Add(box{l: 1, w: 2}); !ok {
		t.Errorf("expected true, got %v", ok)
		t.FailNow() // next test depends on this one passing
	}
	if ok := p.Add(box{l: 1, w: 1}); ok {
		t.Errorf("expected false, got %v", ok)
	}
}

func TestPalletIsFull(t *testing.T) {
	p := pallet{
		boxes: []box{
			{x: 0, y: 0, l: 4, w: 3},
			{x: 0, y: 3, l: 2, w: 1},
		},
	}

	if ok := p.IsFull(); ok {
		t.Errorf("expected false, got %v", ok)
		t.FailNow() // next test depends on this one passing
	}

	p.Add(box{x: 2, y: 3, l: 2, w: 1})

	if ok := p.IsFull(); !ok {
		t.Errorf("expected true, got %v", ok)
	}
}

func TestBoxOverlap(t *testing.T) {
	b1 := box{x: 0, y: 0, l: 2, w: 1}
	b2 := box{x: 1, y: 0, l: 1, w: 2}
	b3 := box{x: 2, y: 0, l: 1, w: 2}

	if ok := b1.Overlap(b2); !ok {
		t.Errorf("expected true, got %v", ok)
	}
	if ok := b2.Overlap(b1); !ok {
		t.Errorf("expected true, got %v", ok)
	}
	if ok := b1.Overlap(b3); ok {
		t.Errorf("expected false, got %v", ok)
	}
	if ok := b3.Overlap(b1); ok {
		t.Errorf("expected false, got %v", ok)
	}
}

func TestRepacker(t *testing.T) {
	out := make(chan *truck)
	in := make(chan *truck)
	rp := newRepackerWorker(out, in, &sync.WaitGroup{})

	palletExcessCount := 3
	palletCount := excessPalletsMax + palletExcessCount

	boxesSeen := map[uint32]int{}

	p := pallet{}
	for id := uint32(0); id < uint32(palletCount); id++ {
		// for this test, make the boxes big enough so that a pallet can only hold 1
		p.boxes = append(p.boxes, box{id: id, l: 4, w: 3})
	}

	out <- &truck{id: 1, pallets: []pallet{p}}

	// Repacker runs in a goroutine, so we can't poke at it until it has finished
	// with the truck.
	// It also holds onto most recent truck, so lets send it an empty truck to
	// make sure the previous one has cleared.
	out <- &truck{id: 2}

	if len(rp.partialPallets) != excessPalletsMax {
		t.Errorf("expected %d excess pallets to be kept, actually kept %d", excessPalletsMax, len(rp.partialPallets))
	}

	trk := <-in
	if trk.id != 1 {
		t.Errorf("expected truck with id 1, received %d", trk.id)
	}
	if len(trk.pallets) != palletExcessCount {
		t.Errorf("expected %d pallets, received %d", palletExcessCount, len(trk.pallets))
	}
	for _, p := range trk.pallets {
		for _, b := range p.boxes {
			boxesSeen[b.id]++
		}
	}

	out <- &truck{id: 0}

	trk = <-in
	if trk.id != 2 {
		t.Errorf("expected truck with id 2, received %d", trk.id)
	}
	if len(trk.pallets) != 0 {
		t.Errorf("expected 0 pallets, received %d", len(trk.pallets))
	}
	for _, p := range trk.pallets {
		for _, b := range p.boxes {
			boxesSeen[b.id]++
		}
	}

	close(out)

	trk = <-in
	if trk.id != 0 {
		t.Errorf("expected truck with id 0, received %d", trk.id)
	}
	if len(trk.pallets) != excessPalletsMax {
		t.Errorf("expected %d pallets, received %d", excessPalletsMax, len(trk.pallets))
	}
	for _, p := range trk.pallets {
		for _, b := range p.boxes {
			boxesSeen[b.id]++
		}
	}

	trk, ok := <-in
	if ok {
		t.Errorf("expected repacker to close channel, but it sent something instead: %+v", trk)
	}

	if len(boxesSeen) != palletCount { // there should only be 1 box per pallet
		t.Errorf("expected %d boxes, received %d", palletCount, len(boxesSeen))
	}
	for id, count := range boxesSeen {
		if count > 1 {
			t.Errorf("received box %d %d times", id, count)
		}
	}
}
