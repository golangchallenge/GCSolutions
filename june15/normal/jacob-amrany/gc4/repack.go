package main

import (
	"runtime"
	"sync"
)

// A repacker repacks trucks.
type repacker struct {
	numroutines int
	wg          sync.WaitGroup
}

const boxBuffCount = 100

func (r *repacker) treeSubdivision(t *truck) (out *truck) {
	out = &truck{id: t.id}

	var trees []*Tree

	boxes := make([]box, 0, boxBuffCount)

	for _, p := range t.pallets {
		for _, b := range p.boxes {
			boxes = append(boxes, b)
			if len(boxes) == boxBuffCount {
				for _, val := range boxes {
					trees = Insert(trees, &val)
				}

				//Reset the boxes array for the next block
				boxes = make([]box, 0, boxBuffCount)
			}
		}
	}

	//Any leftover, insert them
	if len(boxes) > 0 {
		for _, b := range boxes {
			trees = Insert(trees, &b)
		}
	}

	//Put all pallets on the truck
	for _, t := range trees {
		out.pallets = append(out.pallets, t.pal)
	}
	return
}

func newRepacker(in <-chan *truck, out chan<- *truck) *repacker {
	n := runtime.GOMAXPROCS(-1)
	r := &repacker{numroutines: n}
	for i := 0; i < r.numroutines; i++ {
		r.wg.Add(1)
		go r.Repack(in, out)
	}

	go func() {
		r.wg.Wait()
		close(out)
	}()
	return r
}

func (r *repacker) Repack(in <-chan *truck, out chan<- *truck) {
	for t := range in {
		t = r.treeSubdivision(t)
		out <- t
	}
	r.wg.Done()
}
