// Go Challenge #4: Parker & Packer Unpacking & Repacking, Limited

package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"sync"
	"time"
)

type result struct {
	profit, items int
	fail          bool
}

type accounting struct {
	trucksMu, boxesMu sync.Mutex
	trucks            map[int]int
	boxes             map[box]bool
}

func process(doneTime time.Time, r io.Reader, a *accounting, resultChan chan result) {
	defer close(resultChan)

	tr := newTruckReader(r)
	in := make(chan *truck)
	out := make(chan *truck)

	// Construct the repacker.
	newRepacker(in, out)

	// A goroutine to read and send trucks
	go func() {
		defer close(in)

		for {
			done := time.Now().After(doneTime)

			t, err := tr.Next()
			if done || err != nil {
				if done {
					fmt.Println("timeout")
				}

				if err != nil && err != io.EOF {
					fmt.Println("truck reading error: ", err)
				}

				// Send one more empty truck as a signal that they now
				// need to send out any stored boxes.
				a.trucksMu.Lock()
				a.trucks[idLastTruck] = 0
				a.trucksMu.Unlock()
				in <- &truck{id: idLastTruck}

				return
			}

			a.boxesMu.Lock()
			for _, p := range t.pallets {
				for _, b := range p.boxes {
					a.boxes[b.canon()] = true
				}
			}
			a.boxesMu.Unlock()

			// Remember how many pallets were in the truck.
			a.trucksMu.Lock()
			a.trucks[t.id] = len(t.pallets)
			a.trucksMu.Unlock()

			in <- t
		}
	}()

	// Receive the trucks and check them.
	for t := range out {
		r := result{}

		// Only correctly packed pallets count
		for pn, p := range t.pallets {
			for _, b := range p.boxes {
				if !a.boxOk(b) {
					log.Printf("box %v in truck %d was not in the input", b.id, t.id)
					r.fail = true
				}
			}
			if err := p.IsValid(); err == nil {
				r.items += p.Items()
			} else {
				log.Printf("pallet %v in truck %d is not correctly packed: %v", pn, t.id, err)
				r.fail = true
			}
		}

		// Calculate the profit (or loss!) of pallets.
		a.trucksMu.Lock()
		if _, ok := a.trucks[t.id]; ok {
			r.profit = a.trucks[t.id] - len(t.pallets)
		} else {
			log.Printf("truck %v unknown", t.id)
			r.fail = true
		}
		a.trucksMu.Unlock()

		resultChan <- r
	}

	a.boxesMu.Lock()
	if len(a.boxes) != 0 {
		log.Printf("%v boxes not seen in the departing trucks", len(a.boxes))
		resultChan <- result{fail: true}
	}
	a.boxesMu.Unlock()
}

// boxOk checks if the box was in the input, and if so it deletes it from the map.
func (a *accounting) boxOk(b box) (ok bool) {
	b0 := b.canon()
	a.boxesMu.Lock()
	if _, ok = a.boxes[b0]; ok {
		delete(a.boxes, b0)
	}
	a.boxesMu.Unlock()

	return
}

func main() {
	limit := flag.Duration("limit", 2*time.Second, "How long to repack before stopping.")
	ngen := flag.Int("generate", 0, "How many trucks to generate.")
	seed := flag.Int("seed", 1337, "The seed to use for generation (optional).")
	flag.Parse()

	// If asked to generate trucks, do that and then exit.
	if *ngen > 0 {
		generate(*ngen, *seed)
		return
	}

	runtime.GOMAXPROCS(4)

	// This needs to be a local so that the functions in repack.go can't
	// cheat and mess with it. :)
	acc := &accounting{
		trucks: make(map[int]int),
		boxes:  make(map[box]bool),
	}

	profit := 0
	trucks := 0
	items := 0
	fail := false
	resultChan := make(chan result)

	// A goroutine to load trucks, repack them, check the results
	// and send them back to us. This needs to be in a goroutine, so that it's
	// behavior (blocking, taking a long time for certain repacks, etc)
	// never prevents the final timeout from firing.
	go process(time.Now().Add(*limit), os.Stdin, acc, resultChan)

	// The final timeout is 2*limit, so that you have time to work on
	// packing the final truck.
	finalLimit := *limit + *limit
	finalTimeout := time.After(finalLimit)

done:
	for {
		select {
		case <-finalTimeout:
			fmt.Println("final timeout")
			break done
		case r, open := <-resultChan:
			if !open {
				break done
			}
			trucks++
			profit += r.profit
			items += r.items
			if r.fail {
				fail = true
			}
		}
	}

	if fail {
		log.Fatal("Trucks were not repacked correctly.")
	}
	fmt.Println("trucks repacked:", trucks)
	fmt.Println("items repacked:", items)
	fmt.Println("profit:", profit)
}
