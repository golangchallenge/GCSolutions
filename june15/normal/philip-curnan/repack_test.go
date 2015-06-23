package main

import (
	"testing"
	"time"
)

// create a box of each possible size and store it
func TestStoreBoxes(t *testing.T) {
	rp := &repacker{}

	// create and store a box with each length and width combination
	boxes := []box{}
	boxid := uint32(1)
	for w := uint8(1); w <= palletWidth; w++ {
		for l := uint8(1); l <= palletLength; l++ {
			boxes = append(boxes, box{x: 0, y: 0, w: w, l: l, id: boxid})
			boxid++
		}
	}
	testTruck := truck{id: 1, pallets: []pallet{{boxes: boxes}}}
	rp.storeBoxes(&testTruck)

	// test we only used half the array
	for x := 0; x < palletLength; x++ {
		for y := x; y < palletWidth; y++ {
			// matrix axis should have one box on each stack
			if x == y {
				if rp.boxes[x][y].count != 1 {
					t.Fatalf("rp.boxes[%d][%d].count is %d expected 1", x, y, rp.boxes[x][y].count)
				}
			} else {
				if rp.boxes[x][y].count != 2 {
					t.Fatalf("rp.boxes[%d][%d].count is %d expected 2", x, y, rp.boxes[x][y].count)
				}
			}
		}
	}
}

// store two thousand boxes of each size and remove them
func TestStore2000(t *testing.T) {
	rp := &repacker{}

	// create and store a box with each length and width combination
	boxes := []box{}
	boxid := uint32(1)
	for w := uint8(1); w <= palletWidth; w++ {
		for l := uint8(1); l <= palletLength; l++ {
			for n := 0; n < 2000; n++ {
				boxes = append(boxes, box{x: 0, y: 0, w: w, l: l, id: boxid})
				boxid++
			}
		}
	}
	testTruck := truck{id: 1, pallets: []pallet{{boxes: boxes}}}
	rp.storeBoxes(&testTruck)

	// test we only use half the array
	for x := 0; x < palletLength; x++ {
		for y := x; y < palletWidth; y++ {
			// matrix axis should have one box on each stack
			if x == y {
				if rp.boxes[x][y].count != 2000 {
					t.Fatalf("rp.boxes[%d][%d].count is %d expected 1000", x, y, rp.boxes[x][y].count)
				}
			} else {
				if rp.boxes[x][y].count != 4000 {
					t.Fatalf("rp.boxes[%d][%d].count is %d expected 2000", x, y, rp.boxes[x][y].count)
				}
			}
		}
	}
}

// test we can completely fill a pallet (with current palletWidth and palletLength)
func TestFullPallet(t *testing.T) {

	// fill a pallet with 1x1 boxes
	bxs := make([]box, 0, palletWidth+palletLength)
	for n := uint32(2); n < 18; n++ {
		bxs = append(bxs, box{x: 0, y: 0, w: 1, l: 1, id: n})
	}
	trk := &truck{id: 1, pallets: []pallet{{boxes: bxs}}}
	rp := &repacker{}
	rp.storeBoxes(trk)

	out := make(chan *truck)

	go func() {
		rp.packTruck(2, out)
	}()
	go func() {
		for t1 := range out {
			if len(t1.pallets) != 1 {
				t.Fatalf("expected pallet count 1 got %d", len(t1.pallets))
			}
		}
	}()
	time.Sleep(100 * time.Millisecond)
	close(out)

	// fill a pallet with 1 3x3 & 1x1 boxes
	bxs = make([]box, 0, 8)
	bxs = append(bxs, box{x: 0, y: 0, w: 3, l: 3, id: 3})
	for n := uint32(4); n < 4+palletWidth+palletLength-9; n++ {
		bxs = append(bxs, box{x: 0, y: 0, w: 1, l: 1, id: n})
	}
	trk = &truck{id: 2, pallets: []pallet{{boxes: bxs}}}
	rp = &repacker{}
	rp.storeBoxes(trk)

	out = make(chan *truck)

	go func() {
		rp.packTruck(2, out)
	}()
	go func() {
		for t1 := range out {
			if len(t1.pallets) != 1 {
				t.Fatalf("expected pallet count 1 got %d", len(t1.pallets))
			}
		}
	}()
	time.Sleep(100 * time.Millisecond)
	close(out)

}
