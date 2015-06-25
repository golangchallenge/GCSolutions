package main

import (
	"log"
	"math/rand"
)

const (
	// how many trucks are set aside for processing after the first timeout
	// given in percent
	restpostenFactor = 50
)

// A Packline repacks trucks in multiple batches
type Packline struct {
	// signal channel for shutdown
	shutdown chan struct{}
	// number of trucks sent to the out channel
	sentTrucks int
	// restposten holds the trucks to be processed after the first timeout
	restposten []*truck
}

// NewPackline initializes a new Packline
func NewPackline() *Packline {
	pl := &Packline{
		shutdown: make(chan struct{}),
		// make restposten reasonably large to avoid allocations during processing
		restposten: make([]*truck, 0, 50000),
	}
	return pl
}

// ProcessTrucks takes trucks from in and sends them repacked into out.
// runs in it's own GoRoutine
// Receival of the last (empty) truck is signaled via ltReceived
func (pl *Packline) ProcessTrucks(in <-chan *truck, out chan<- *truck, ltReceived chan<- struct{}) {

	b := NewBatch()
	for {
		select {
		case t := <-in:
			if t == nil {
				continue
			}
			if len(t.pallets) == 0 {
				ltReceived <- struct{}{}
				// the last truck is empty, so we don't need to unpack it, or send it back
				continue
			}

			// rand was intentionally not seeded,
			// as it is not important to have a truly random decision here
			// as multiple Packlines read from the in channel
			if rand.Intn(100) > 100-restpostenFactor {
				pl.restposten = append(pl.restposten, t)
				continue
			}

			b.UnpackTruck(t)
			if b.IsFull() {
				b.RepackPallets()
				pl.sentTrucks += b.SendTrucks(out)
				b = NewBatch()
			}

		case <-pl.shutdown:
			trucks := pl.sentTrucks
			for _, t := range pl.restposten {
				b.UnpackTruck(t)
				if b.IsFull() {
					b.RepackPallets()
					pl.sentTrucks += b.SendTrucks(out)
					b = NewBatch()
				}
			}

			// send out the remaining trucks
			b.RepackPallets()
			pl.sentTrucks += b.SendTrucks(out)
			log.Printf("Processed %d trucks (%d/%d)",
				pl.sentTrucks, trucks, pl.sentTrucks-trucks)
			close(pl.shutdown)
			return
		}
	}
}

// Shutdown shuts down the packline in a two phase way.
// Shutdown first signals the request for shutdown and then
// waits for the packline to clean up.
func (pl *Packline) Shutdown() {
	pl.shutdown <- struct{}{}
	for range pl.shutdown {
	}
}
