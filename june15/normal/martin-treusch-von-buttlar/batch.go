package main

import (
	"fmt"
	"sort"
)

const (
	// MaxBoxCount limits the number of boxes per batch
	MaxBoxCount = 250
	// truckFactor is the upper limit of trucks needed to reach MaxBoxCount boxes
	truckFactor = MaxBoxCount / 15
)

// Batch stores and packs a few truck loads worth of pallets and boxes
type Batch struct {
	bins        map[int][]box    // boxes are stored with their area as key
	boxCount    int              // number of boxes already seen
	openPallets int              // pallets are allocated only once, this is the running count
	pallets     []pallet         // this slice holds the pallets during packing
	trucks      []*truck         // this slice holds the trucks unpacked for this batch
	repacker    func(*Batch) int // holds a reference to the repacking function used
}

// NewBatch creates a new, empty batch
func NewBatch() *Batch {
	b := &Batch{
		bins:     make(map[int][]box, palletLength*palletWidth),
		trucks:   make([]*truck, 0, truckFactor),
		repacker: repackMFFD,
	}
	for _, size := range binSizes {
		b.bins[size] = make([]box, 0, 100)
	}
	return b
}

// IsFull returns whether this batch should be repacked and send out or not
func (b *Batch) IsFull() bool {
	return b.boxCount > MaxBoxCount
}

// RepackPallets delegates the heavy box lifting to the repacker function
func (b *Batch) RepackPallets() int {
	b.pallets = make([]pallet, b.openPallets)
	repackedBoxes := b.repacker(b)
	if repackedBoxes != b.boxCount {
		panic(fmt.Sprintf("%d of %d boxes lost.", repackedBoxes, b.boxCount))
	}
	return repackedBoxes
}

// SendTrucks fills the trucks with pallets and sends them on their way into out
func (b *Batch) SendTrucks(out chan<- *truck) int {
	sentTrucks := 0
	if len(b.trucks) == 0 {
		return sentTrucks
	}

	// extract filled pallets
	pallets := make([]pallet, 0, b.openPallets)
	for _, p := range b.pallets {
		if p.Items() > 0 {
			pallets = append(pallets, p)
		}
	}

	// put all non empty pallets on the first truck
	b.trucks[0].pallets = pallets
	sentTrucks++
	out <- b.trucks[0]

	// send out all the other trucks empty
	for _, t := range b.trucks[1:] {
		t.pallets = []pallet{}
		out <- t
		sentTrucks++
	}
	return sentTrucks
}

// UnpackTruck unpacks all boxes from a single truck
func (b *Batch) UnpackTruck(t *truck) {
	b.openPallets += len(t.pallets)
	for _, p := range t.pallets {
		b.boxCount += len(p.boxes)
		for j, box := range p.boxes {
			sq := int(box.l) * int(box.w)
			b.bins[sq] = append(b.bins[sq], p.boxes[j])
		}
	}
	b.trucks = append(b.trucks, t)
}

// binSizes holds a list of all box areas in descending order
// this slice is only read from multiple GoRoutines
var binSizes []int

// init initializes binSizes
func init() {
	bins := make(map[int]bool)

	for w := 1; w <= palletWidth; w++ {
		for l := 1; l <= palletLength; l++ {
			bins[w*l] = true
		}
	}
	binSizes = make([]int, 0, len(bins))

	for k := range bins {
		binSizes = append(binSizes, k)
	}

	sort.Sort(sort.Reverse(sort.IntSlice(binSizes)))
}
