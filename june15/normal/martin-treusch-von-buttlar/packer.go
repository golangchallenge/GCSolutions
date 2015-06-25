package main

// repackMFFD uses a modified First Fit Decreasing algorithm to pack the boxes
// it speeds up FFD 3 times for a bit less packing efficiency.
func repackMFFD(b *Batch) int {
	boxes := 0

	// *pallet are used to remove full pallets
	// from the list without additional allocations
	pallets := make([]*pallet, len(b.pallets))
	for i := range b.pallets {
		pallets[i] = &b.pallets[i]
	}

	for _, size := range binSizes {
		lastPallet := 0
	box:
		for _, box := range b.bins[size] {
			startPallet := lastPallet
			for i, p := range pallets[startPallet:] {
				if p == nil {
					continue
				}

				// if the box is too big or too small, we assume that the next box
				// with the same size won't match either.
				if size > palletWidth || size <= 2 {
					lastPallet = startPallet + i
				}

				if fits, pg := p.FitBoxWithGrid(box); fits {
					boxes++
					if pg.IsFull() {
						pallets[i] = nil
					}
					continue box
				}
			}
		}
	}
	return boxes
}

// repackFFD uses a First Fit Decreasing algorithm to pack the boxes
func repackFFD(b *Batch) int {
	boxes := 0
	for _, size := range binSizes {
	box:
		for _, box := range b.bins[size] {
			for i, p := range b.pallets {
				if p.FitBox(box) {
					b.pallets[i] = p
					boxes++
					continue box
				}
			}
		}
	}
	return boxes
}
