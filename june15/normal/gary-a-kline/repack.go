package main

import "fmt"

// A repacker repacks trucks.
type repacker struct {
}

// repack will 'unpack' all boxes on the truck and then start filling pallets
// by placing the largest box first, followed by the largest boxes that can
// fit in the remaining space on the pallet. The pallet is filled from
// left to right, then down.
func repack(t *truck, boxQueue cargoQueue, queueTotal int) (*truck, cargoQueue, int) {
	var (
		candidate, availSpace int
		openLength, openWidth uint8
		moreBoxes, ok         bool
		err                   error
		cell                  openCell
		p                     pallet
		bx                    box
	)
	plan := new(packPlan)
	stack := make(openCellStack, 0, 10)

	out := &truck{id: t.id}

	// 'unpack' boxes into a queue
	for _, p := range t.pallets {
		for _, b := range p.boxes {
			b.turnVertical()
			boxQueue.addCargo(cargo{item: b, area: int(b.w * b.l)})
			queueTotal += int(b.w * b.l)
		}
	}
	boxQueue.sort() // sort the boxes by their area, in descending order

	// fill pallets until we have a half-pallet of boxes remaining, or
	// if this is the last truck, until we've emptied the box queue
	for (queueTotal > (palletWidth * (palletLength - 1))) || (t.id == idLastTruck && queueTotal > 0) {

		// initialize a fresh pallet
		p.boxes = make([]box, 0)
		plan.clear()
		stack.clear()
		cell = openCell{0, 0}
		openLength = palletLength
		openWidth = palletWidth
		availSpace = palletWidth * palletLength

		// fill the pallet until no more boxes will fit, or we run out of boxes
		candidate = 0
		moreBoxes = true
		for moreBoxes && availSpace > 0 {
			candidate, moreBoxes = boxQueue.findNextCandidate(int(openLength*openWidth), candidate)
			if moreBoxes {
				bx = boxQueue[candidate].item
				switch {
				case bx.l <= openLength && bx.w <= openWidth:
				case bx.l <= openWidth && bx.w <= openLength:
					bx.l, bx.w = bx.w, bx.l
				default:
					candidate++
					continue
				}

				bx.x, bx.y = cell.col, cell.row
				p.boxes = append(p.boxes, bx)

				// mark the plan with the box we just put on the pallet
				err = plan.markBox(bx.x, bx.y, bx.w)
				if err != nil {
					fmt.Println(err)
				}

				// save the open cells immediately below and immediately
				// to the right of the box, if they are empty
				if ok, _ = plan.isEmptyCell(bx.x, bx.y+bx.w); ok {
					stack.push(openCell{bx.x, bx.y + bx.w})
				}
				if ok, _ = plan.isEmptyCell(bx.x+bx.l, bx.y); ok {
					stack.push(openCell{bx.x + bx.l, bx.y})
				}

				availSpace -= boxQueue[candidate].area
				queueTotal -= boxQueue[candidate].area
				boxQueue.removeCargo(candidate)

				if availSpace > 0 {
					candidate = 0
					cell, err = stack.pop()
					if err != nil {
						// stack is empty, no more room on pallet
						availSpace = 0
						continue
					}
					openLength, openWidth = plan.getAvailSpace(cell)
				}
			}
		}
		// add the pallet to the truck
		out.pallets = append(out.pallets, p)
	}
	return out, boxQueue, queueTotal
}

func newRepacker(in <-chan *truck, out chan<- *truck) *repacker {
	boxQueue := make(cargoQueue, 0, 20)
	queueTotal := 0

	go func() {
		for t := range in {
			t, boxQueue, queueTotal = repack(t, boxQueue, queueTotal)
			out <- t
		}
		// The repacker must close channel out after it detects that
		// channel in is closed so that the driver program will finish
		// and print the stats.
		close(out)
	}()
	return &repacker{}
}
