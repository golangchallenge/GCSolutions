package main

import "sort"

// A repacker repacks trucks.
type repacker struct {
}

type someBoxes []box

func (b someBoxes) Len() int {
	return len(b)
}

func (b someBoxes) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b someBoxes) Less(i, j int) bool {
	return b[i].l > b[j].l
}

func (b someBoxes) arrangeLengthWise() someBoxes {
	for i, singleBox := range b {
		if singleBox.w < singleBox.l {
			singleBox.l, singleBox.w = singleBox.w, singleBox.l
		}

		b[i] = singleBox
	}

	return b
}

func findOpenSpot(pallets []pallet, newBox box) (uint8, int) {
	for idx, p := range pallets {
		// assume pallet is sorted by x-coordinate
		b := p.boxes[len(p.boxes)-1]

		xEdge := b.x + b.l
		if palletWidth >= xEdge && (palletWidth-xEdge) >= newBox.l {
			return xEdge, idx
		}
	}

	return 0, -1
}

func fitBoxesToPallet(t *truck) (out *truck) {
	out = &truck{id: t.id}
	var boxes []box

	for _, p := range t.pallets {
		b := someBoxes(p.boxes).arrangeLengthWise()
		boxes = append(boxes, b...)
	}

	sort.Sort(someBoxes(boxes))
	for _, b := range boxes {
		xEdge, openPalletIdx := findOpenSpot(out.pallets, b)

		b.x, b.y = xEdge, 0

		if openPalletIdx == -1 {
			out.pallets = append(out.pallets, pallet{boxes: []box{b}})
			continue
		}

		openPallet := out.pallets[openPalletIdx]
		openPallet.boxes = append(openPallet.boxes, b)

		out.pallets[openPalletIdx] = openPallet
	}

	return
}

func newRepacker(in <-chan *truck, out chan<- *truck) *repacker {
	go func() {
		for t := range in {
			// The last truck is indicated by its id. You might
			// need to do something special here to make sure you
			// send all the boxes.
			if t.id == idLastTruck {
			}

			// t = oneBoxPerPallet(t)
			t = fitBoxesToPallet(t)
			out <- t
		}
		// The repacker must close channel out after it detects that
		// channel in is closed so that the driver program will finish
		// and print the stats.
		close(out)
	}()
	return &repacker{}
}
