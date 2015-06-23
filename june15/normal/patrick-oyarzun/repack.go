package main

import (
	"sort"
	"sync"
)

// A repacker repacks trucks.
type repacker struct {
}

// byArea implements sort.Interface to sort boxes by decreasing area
type byArea []box

func (b byArea) Len() int {
	return len(b)
}

func (b byArea) Swap(i, j int) {
	b[i], b[j] = b[j], b[i]
}

func (b byArea) Less(i, j int) bool {
	x := b[i]
	y := b[j]
	return x.w*x.l > y.w*y.l
}

type node struct {
	used        bool
	down, right *node
	x, y, w, h  uint8
}

type binaryTreePacker struct {
	root node
}

// pack fits as many boxes as it can into the pallet
// and returns the remainder along with the pallet
func (b *binaryTreePacker) pack(blocks []box) ([]box, pallet) {
	var p pallet
	var used []int
	for i, block := range blocks {
		// attempt to fit the box
		n, ok := b.search(&b.root, block.w, block.l)
		if !ok {
			// attempt with the box rotated
			n, ok = b.search(&b.root, block.l, block.w)
			if !ok {
				continue
			}
			block.l, block.w = block.w, block.l
		}
		pack := b.subdivide(n, block.w, block.l)
		used = append(used, i)
		block.x = pack.y
		block.y = pack.x
		p.boxes = append(p.boxes, block)
	}
	for i := len(used) - 1; i >= 0; i-- {
		blocks = append(blocks[:used[i]], blocks[used[i]+1:]...)
	}
	return blocks, p
}

// search attempts to find a big enough space on the pallet.
// returns false as the second value if there is no such space.
func (b *binaryTreePacker) search(root *node, w, h uint8) (*node, bool) {
	if root.used {
		if x, ok := b.search(root.right, w, h); ok {
			return x, true
		}
		return b.search(root.down, w, h)
	} else if w <= root.w && h <= root.h {
		return root, true
	}
	return nil, false
}

// subdivide splits a node into three sections:
//  - The region containing a wxh box
//  - The region to the right of that region
//  - The region below that region
// Boxes are always placed in the 0,0 corner of a region.
func (b *binaryTreePacker) subdivide(n *node, w, h uint8) *node {
	n.used = true
	n.down = &node{x: n.x, y: n.y + h, w: n.w, h: n.h - h}
	n.right = &node{x: n.x + w, y: n.y, w: n.w - w, h: h}
	return n
}

// newBinaryTreePacker contructs a packer suitable for pallets.
func newBinaryTreePacker() *binaryTreePacker {
	return &binaryTreePacker{root: node{
		x: 0, y: 0, w: palletWidth, h: palletLength,
	}}
}

// packList applies a heuristic algorithm to pack
// as many boxes as it can onto a pallet. Returns
// the remaining boxes and a pallet.
func packList(boxes []box) (rest []box, p pallet) {
	sort.Sort(byArea(boxes))
	packer := newBinaryTreePacker()
	return packer.pack(boxes)
}

// unpackTrucks converts a stream of trucks into a stream of boxes
// and a stream of empty trucks. Because the packers are very good
// at their job, a lot of trucks are simply sent out empty since
// there aren't enough pallets after compacting to fill all trucks.
func unpackTrucks(in <-chan *truck, out chan<- *truck) (boxes chan box, emptyTrucks chan *truck) {
	boxes = make(chan box)
	emptyTrucks = make(chan *truck, 20)
	go func() {
		for t := range in {
			for _, p := range t.pallets {
				for _, b := range p.boxes {
					boxes <- b
				}
			}

			// fill up to 20 trucks if possible, otherwise
			// send out empty trucks
			select {
			case emptyTrucks <- t:
			default:
				t.pallets = t.pallets[:0]
				out <- t
			}
		}
		close(boxes)
		close(emptyTrucks)
	}()
	return boxes, emptyTrucks
}

// palletizeWorker is called in a go routine by palletize
// each instance pulls boxes into pallets and signals the waitgroup
// when done.
func palletizeWorker(boxes <-chan box, pallets chan<- pallet, wg *sync.WaitGroup) {
	const (
		// Controls the number of boxes sent to the binaryTreePacker at a time.
		// This value was determined through experimentation to be approximately
		// optimal. Too low of a value doesn't provide the packer with enough
		// information. Too high of a value causes large overhead in the sort step
		// since at most 16 boxes are consumed per pallet and the rest must be
		// processed again.
		bufferSize = 25
	)
	go func() {
		var buf = make([]box, 0, bufferSize)
		var p pallet
		// send bufferSize chunks to the packer
		for b := range boxes {
			buf = append(buf, b)
			if len(buf) == bufferSize {
				buf, p = packList(buf)
				pallets <- p
			}
		}
		// palletize any remainder less than the buffer size
		for len(buf) > 0 {
			buf, p = packList(buf)
			pallets <- p
		}
		wg.Done()
	}()
}

// palletize turns a stream of boxes into a stream of pallets
func palletize(boxes <-chan box) (pallets chan pallet) {
	const (
		numWorkers = 4
	)
	pallets = make(chan pallet)
	go func() {
		wg := new(sync.WaitGroup)
		wg.Add(numWorkers)
		for i := 0; i < numWorkers; i++ {
			go palletizeWorker(boxes, pallets, wg)
		}
		wg.Wait()
		close(pallets)
	}()
	return pallets
}

// newRepacker reads trucks from in and sends repacked trucks
// to out. Many trucks end up empty after repacking, with their
// boxes moved to other trucks instead.
func newRepacker(in <-chan *truck, out chan<- *truck) *repacker {
	go func() {
		boxes, emptyTrucks := unpackTrucks(in, out)
		pallets := palletize(boxes)

		for truck := range emptyTrucks {
			out <- fillTruck(truck, pallets)
		}
		close(out)
	}()
	return &repacker{}
}

// fillTruck fills a truck with pallets from the channel
// until all original pallets are replaced. This will
// sometimes end up putting more boxes on the truck than
// the truck showed up with.
func fillTruck(t *truck, pallets <-chan pallet) *truck {
	n := len(t.pallets)
	t.pallets = t.pallets[:0]
	for len(t.pallets) < n {
		p, ok := <-pallets
		if !ok {
			return t
		}
		t.pallets = append(t.pallets, p)
	}
	return t
}
