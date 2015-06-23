package main

// A repacker repacks trucks.
type repacker interface {
	// repack repacks trucks from the input channel and
	// writes them to the output channel.  It should close
	// the out channel after it detects that the in channel
	// is closed.
	repack(in <-chan *truck, out chan<- *truck)
}

// oneBoxRepacker is the worst possible repacker, since it uses a new
// pallet for every box.
type oneBoxRepacker struct {
}

// This repacker is the worst possible, since it uses a new pallet for
// every box. Your job is to replace it with something better.
func (oneBoxRepacker) repack(t *truck) (out *truck) {
	out = &truck{id: t.id}
	for _, p := range t.pallets {
		for _, b := range p.boxes {
			b.x, b.y = 0, 0
			out.pallets = append(out.pallets, pallet{boxes: []box{b}})
		}
	}
	return
}

func newRepacker(in <-chan *truck, out chan<- *truck) repacker {
	//var r repacker = &shelfRepacker{}
	//r := newGuillotineRepacker()
	r := newMultiGuillotineRepacker(5)
	go r.repack(in, out)
	return r
}

func (b *box) isPortrait() bool {
	return b.l >= b.w
}

func (b *box) rotate() {
	b.l, b.w = b.w, b.l
}

func max(a, b uint8) uint8 {
	if a > b {
		return a
	}
	return b
}

func min(a, b uint8) uint8 {
	if a < b {
		return a
	}
	return b
}
