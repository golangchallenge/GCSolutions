package main

import (
	"fmt"
)

// A repacker repacks trucks.
type repacker struct {
}

// This repacker is the worst possible, since it uses a new pallet for
// every box. Your job is to replace it with something better.
func oneBoxPerPallet(t *truck) (out *truck) {
	out = &truck{id: t.id}
	for _, p := range t.pallets {
		for _, b := range p.boxes {
			b.x, b.y = 0, 0
			out.pallets = append(out.pallets, pallet{boxes: []box{b}})
		}
	}
	return
}

func betterPacker(t *truck, store *Table) (out *truck) {
	out = &truck{id: t.id}

	nrPallets := t.Unload(*store)

	for i := 0; i < nrPallets && !store.IsEmpty(); i++ {
		var p pallet
		// grid will track the free space on pallet
		freeGridSpace := NewInitialGrid()

		for !freeGridSpace.IsEmpty() && !store.IsEmpty() {

			// grab last element of g
			last := len(freeGridSpace) - 1
			e := freeGridSpace[last]

			b, _ := store.GetBoxThatFitsOrIsEmpty(e.size, e.orient)

			if b == emptybox {
				break
			}

			newFreeGridElements := Put(&b, e)
			b.AddToPallet(&p)
			freeGridSpace.Update(newFreeGridElements)
		} // end loop

		fmt.Println("Repacked Pallet:\n", p)
		out.pallets = append(out.pallets, p)
	} //  end for pallets
	fmt.Println("Store end: ", store)
	return
}

func newRepacker(in <-chan *truck, out chan<- *truck) *repacker {
	store := NewTable()
	go func() {
		for t := range in {
			// The last truck is indicated by its id. You might
			// need to do something special here to make sure you
			// send all the boxes.
			if t.id == idLastTruck {
				//				emptytruck := truck{}
				//				out <- &emptyTruck
			}

			t = betterPacker(t, &store)
			out <- t
		}
		// The repacker must close channel out after it detects that
		// channel in is closed so that the driver program will finish
		// and print the stats.
		close(out)
	}()
	return &repacker{}
}
