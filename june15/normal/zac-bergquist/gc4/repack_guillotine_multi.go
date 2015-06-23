package main

// multiGuillotineRepacker uses multiple guillotineRepackers
// to repack boxes.
type multiGuillotineRepacker struct {
	repackers []*guillotineRepacker
}

func newMultiGuillotineRepacker(count int) *multiGuillotineRepacker {
	mgr := &multiGuillotineRepacker{
		repackers: make([]*guillotineRepacker, count),
	}
	for i := range mgr.repackers {
		mgr.repackers[i] = newGuillotineRepacker()
	}
	return mgr
}

// addBox attemps to add the box to the guillotineRepakcer at index i.
func (mgr *multiGuillotineRepacker) addBox(i int, b *box) bool {
	return mgr.repackers[i].addBox(b)
}

func (mgr *multiGuillotineRepacker) repack(in <-chan *truck, out chan<- *truck) {
	for t := range in {
		currentTruck := &truck{id: t.id}
		for _, p := range t.pallets {
			for _, b := range p.boxes {
				// try each repacker, hoping at least one of them can hold the box
				added := false
				for _, gr := range mgr.repackers {
					if gr.addBox(&b) {
						added = true
						// if we've filled up this pallet enough, then close it
						// (this didn't show any noticable improvements)

						// if fa := gr.freeArea(); fa < 1 {
						// 	currentTruck.pallets = append(currentTruck.pallets, pallet{boxes: mgr.repackers[0].boxes})
						// 	mgr.repackers[0].reset()
						// }
						break
					}
				}
				// if there wasn't room in any of the repackers, then we add
				// a pallet to a truck and start a fresh pallet
				if !added {
					// restart the first pallet (that way subsequent adds will likely succeed)
					currentTruck.pallets = append(currentTruck.pallets, pallet{boxes: mgr.repackers[0].boxes})
					mgr.repackers[0].reset()
					if !mgr.repackers[0].addBox(&b) {
						panic("addBox failed on empty repacker")
					}
				}
			}
		}
		out <- currentTruck
	}
	// no more trucks are coming, but we have to flush the unsent boxes
	lastTruck := &truck{}
	for _, gr := range mgr.repackers {
		if len(gr.boxes) > 0 {
			lastTruck.pallets = append(lastTruck.pallets, pallet{boxes: gr.boxes})
		}
	}
	out <- lastTruck

	close(out)
}
