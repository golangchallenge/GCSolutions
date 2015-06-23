package main

import (
	"fmt"
	"os"
)

// A repacker repacks trucks.
// Accept trucks in for repacking (newRepacker).
// Store the boxes (storeBoxes) taken off each truck.
// Loads space optimized pallets (packPallet) onto trucks (packTruck).
// Send repacked trucks out (newRepacker)

// repacker structure provides storage for boxes removed from trucks
type repacker struct {
	// array has an element for each possible box size,
	//	each element is a stack of box id's for that size
	boxes [palletWidth][palletLength]boxStack
}

// storeBoxes takes boxes off the given truck and puts them onto boxStacks
func (r *repacker) storeBoxes(t *truck) {
	defer func() {
		if x := recover(); x != nil {
			fmt.Printf("panic in storeBoxes: truck %v\nerror: %v\n", t, x)
			os.Exit(1)
		}
	}()

	// push boxes onto stacks, there is one stack for each dimension combination
	var x, y uint8
	for _, p := range t.pallets {
		for _, b := range p.boxes {
			// check box size is legit
			if !checkBounds(b.w, b.l) {
				fmt.Printf("box %d (%v) on truck %d has an invalid size. The box is discarded.\n", b.id, b, t.id)
				continue
			}
			// adapt dimensions to zero based arrays
			y, x = b.w-1, b.l-1
			// only use half of the array, avoid duplicate sized elements
			if b.w < b.l {
				y, x = x, y
			}
			r.boxes[x][y].push(b.id)
		}
	}
	return
}

// packPallet goes through stored boxes and tries to fill a pallet with the boxes.
// if no boxes are available an empty pallet is returned.
func (r *repacker) packPallet() (p pallet) {
	defer func() {
		if x := recover(); x != nil {
			fmt.Printf("panic in packPallet: error: %v\n", x)
		}
	}()
	// create a pallet to return, preallocating storage for a reasonable number of boxes
	p = pallet{boxes: make([]box, 0, palletWidth*palletLength/2)}
	// we divide a pallet into spaces and try to match the spaces to stored boxes
	// start out with a single space that matches the full size of the pallet
	spaces := make([]space, 0, palletWidth*palletLength/2)
	spaces = append(spaces, space{x: 0, y: 0, w: palletWidth, l: palletLength})
	// loop while there are spaces 
NEXTSPACE:
	for len(spaces) > 0 {
		// the current space we are trying to fill
		s := spaces[0]
		// start with a cutout matching the current space and then shrink it until there is a match
		cutout := s
		// try different size boxes, working our way down in size
		// the cutout can never be larger than the space so use current space dimensions as upper limits
		for cutout.w = s.w; cutout.w > 0; cutout.w-- {
			for cutout.l = s.l; cutout.l > 0; cutout.l-- {
				// box storage is a zero based 2d array
				y, x := cutout.w-1, cutout.l-1
				// we only use the part of the array where y >= x
				if cutout.w < cutout.l {
					y, x = x, y
				}
				// do we have a box of the cutout size in storage
				if r.boxes[x][y].count > 0 {
					// get the position where the box goes and any leftover spaces
					// (space above and space to the right of the box) that still
					// need to be filled
					cutSpaces, fits, valid := s.carveSpace(cutout)
					if fits {
						// is the space allocated for the box valid
						// invalid means no space available
						if valid[0] {
							// get the box from storage
							boxid, boxvalid := r.boxes[x][y].pop()
							if boxvalid {
								// create the box and put it on the output pallet at the correct position
								b := box{x: cutSpaces[0].x, y: cutSpaces[0].y, w: cutSpaces[0].w, l: cutSpaces[0].l, id: boxid}
								p.boxes = append(p.boxes, b)
								// are the leftover spaces valid, if so put in spaces to be filled
								if valid[1] {
									spaces = append(spaces, cutSpaces[1])
								}
								if valid[2] {
									spaces = append(spaces, cutSpaces[2])
								}
								// more spaces left to fill or just the current?
								if len(spaces) == 1 {
									return p
								}
								// remove the space we just filled
								spaces = spaces[1:]
								continue NEXTSPACE
							}
						}
					}
				}
			} // l for loop
		} // w for loop
		// didn't find any matching boxes
		break
	}
	return p
}

// packTruck creates a truck with the given id and tries to fill it
// with pallets. If it cannot create more pallets it places the
// truck, with it's current pallets, in the out channel.
func (r *repacker) packTruck(tid int, out chan<- *truck) {
	defer func() {
		if x := recover(); x != nil {
			fmt.Printf("panic in packTruck, truck id: %d, error: %v\n", tid, x)
		}
	}()
	// last truck is just a signal, don't pack it
	if tid == idLastTruck {
		return
	}
	t := &truck{id: tid}
	for {
		p := r.packPallet()
		// add non-empty pallets to the truck
		if len(p.boxes) > 0 {
			t.pallets = append(t.pallets, p)
		} else {
			break
		}
	}
	out <- t
	return
}

// newRepacker accepts packed trucks via an input channel 'in', and
// sends out repacked tracks on channel 'out'. A repacker struct
// pointer is returned to the caller.
func newRepacker(in <-chan *truck, out chan<- *truck) *repacker {
	defer func() {
		if x := recover(); x != nil {
			fmt.Printf("panic in newRepacker error: %v\n", x)
			os.Exit(1)
		}
	}()
	// prep work before processing anything
	rp := &repacker{}
	// preallocate some storage space for boxes taken from trucks
	for l := 0; l < palletLength; l++ {
		for w := l; w < palletWidth; w++ {
			rp.boxes[l][w].stack = make([]uint32, 0, 1024)
		}
	}
	// channel for sending truckids to truck packer
	packTids := make(chan int, 10)

	// unpack the incoming trucks and store the boxes
	go func() {
		for t := range in {
			// The last truck t.id signal means no more trucks to unpack.
			if t.id == idLastTruck {
				close(packTids)
				break
			}
			// store the truck boxes, put t.id into packTids channel to pass to next step
			rp.storeBoxes(t)
			packTids <- t.id
			// done with this incoming truck, can discard
		}
	}()

	// re-pack the trucks and send them out
	go func() {
		for tid := range packTids {
			// pack truck and send it to driver program via out channel
			rp.packTruck(tid, out)
		}
		// if here we've packed the last truck
		close(out)
	}()
	// this routine is done but the go routines we launched are still running
	return rp
}
