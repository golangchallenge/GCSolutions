package main

import (
	"log"
	"runtime"
	"sync"
)

// repacker controls the repacking of trucks.
// it starts multiple packLines and manages the lifecycle
type repacker struct {
	// packing is split into multiple packLines
	lines []*Packline
	// ltReceived signals the arrival of the last truck (empty one)
	ltReceived chan struct{}
	out        chan<- *truck
	in         <-chan *truck
}

// waitForQuitSignal shuts the packLines down in parallel, when the quit signal is received
func (r *repacker) waitForQuitSignal() {
	<-r.ltReceived
	log.Println("Shutting down packlines")
	wg := sync.WaitGroup{}
	for _, pl := range r.lines {
		pl := pl // circumvent loop variable reuse
		wg.Add(1)
		go func() {
			pl.Shutdown()
			wg.Done()
		}()
	}
	wg.Wait()
	log.Println("shutdown completed")
	close(r.out)
}

// startPacklines creates the given number of packLines
func (r *repacker) startPacklines(numberOfLines int) {
	log.Printf("starting %d packLines.\n", numberOfLines)
	for i := 0; i < numberOfLines; i++ {
		pl := NewPackline()
		r.lines = append(r.lines, pl)
		go pl.ProcessTrucks(r.in, r.out, r.ltReceived)
	}
}

// newRepacker creates a new repacker and starts the packLines
func newRepacker(in <-chan *truck, out chan<- *truck) *repacker {
	r := &repacker{
		ltReceived: make(chan struct{}),
		in:         in,
		out:        out,
	}
	r.startPacklines(runtime.GOMAXPROCS(-1))
	go r.waitForQuitSignal()

	return r
}
