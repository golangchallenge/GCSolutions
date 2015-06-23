package main

import (
	"fmt"
	"sync"
)

//go:generate stringer -type boxsize
type boxsize int

const (
	b1x1 boxsize = iota
	b1x2
	b1x3
	b1x4
	b2x1
	b2x2
	b2x3
	b2x4
	b3x1
	b3x2
	b3x3
	b3x4
	b4x1
	b4x2
	b4x3
	b4x4
)

type (
	repacker   struct{}
	boxlist    []boxsize
	palletMask uint16
	boxStorage struct {
		boxes [16][]uint32
		count uint32
	}
)

var boxesBySize = boxlist{
	b4x4,       // 16
	b4x3, b3x4, // 12
	b3x3,       // 9
	b4x2, b2x4, // 8
	b3x2, b2x3, // 6
	b2x2, b4x1, b1x4, // 4
	b3x1, b1x3, // 3
	b2x1, b1x2, // 2
	b1x1, // 1
}

func (t *truck) load(b []box) {
	p := makePallet(b)
	t.pallets = append(t.pallets, p)
}
func (t *truck) unload(s *boxStorage) {
	for _, p := range t.pallets {
		for _, b := range p.boxes {
			s.store(b)
		}
	}
}

func (b box) size() boxsize {
	t := b.canon()
	return boxsize((t.w-1)<<2 | (t.l - 1))
}
func (b box) shift(d, r uint8) (out box) {
	out = b
	out.x += d
	out.y += r
	return
}
func (b box) mask() (pm palletMask) {
	for x := b.x; x < b.x+b.l; x++ {
		for y := b.y; y < b.y+b.w; y++ {
			if x > 3 || y > 3 {
				return 0
			}
			pm = pm | 1<<(15-(4*x+y))
		}
	}
	return
}

func (s boxsize) box(id uint32) (b box) {
	b.w = uint8((s>>2)&3) + 1
	b.l = uint8(s&3) + 1
	b.id = id
	return
}
func (s boxsize) canon() boxsize {
	switch s {
	case b1x2:
		return b2x1
	case b1x3:
		return b3x1
	case b1x4:
		return b4x1
	case b2x3:
		return b3x2
	case b2x4:
		return b4x2
	case b3x4:
		return b4x3
	default:
		return s
	}
}

func (s *boxStorage) store(b box) {
	b = b.canon()
	t := b.size()
	s.boxes[t] = append(s.boxes[t], b.id)
	s.count++
}
func (s *boxStorage) fetch(sizes ...boxsize) (b []box) {
	asked := [b4x4 + 1]int{}
	var size boxsize
	var boxID uint32
	for _, size = range sizes {
		asked[size.canon()]++
	}
	for i, count := range asked {
		if count == 0 {
			continue
		}
		if len(s.boxes[i]) < count {
			return
		}
	}
	for _, size := range sizes {
		c := size.canon()
		boxID, s.boxes[c] = s.boxes[c][0], s.boxes[c][1:]
		b = append(b, size.box(boxID))
	}
	s.count -= uint32(len(b))
	return
}
func (s *boxStorage) String() string {
	return fmt.Sprint(
		"boxStorage{ ",
		len(s.boxes[b1x1]), "@1x1 ", len(s.boxes[b2x1]), "@2x1 ",
		len(s.boxes[b2x2]), "@2x2 ", len(s.boxes[b3x1]), "@3x1 ",
		len(s.boxes[b3x2]), "@3x2 ", len(s.boxes[b3x3]), "@3x3 ",
		len(s.boxes[b4x1]), "@4x1 ", len(s.boxes[b4x2]), "@4x2 ",
		len(s.boxes[b4x3]), "@4x3 ", len(s.boxes[b4x4]), "@4x4 ",
		"}",
	)
}
func (s *boxStorage) findLargestBox(pm palletMask) (pmOut palletMask, b box, ok bool) {
	for _, size := range boxesBySize {
		boxes := s.fetch(size)
		if len(boxes) == 0 {
			continue
		}
		b = boxes[0]
		pmOut, b, ok = pm.placeBox(b)
		if ok {
			return pmOut, b, true
		}
		s.store(b)
	}
	return pm, box{}, false
}

func (p palletMask) placeBox(b box) (outP palletMask, outB box, ok bool) {
	var row, col uint8
outer:
	for row = 0; row < 4; row++ {
		for col = 0; col < 4; col++ {
			outB = b.shift(row, col)
			bMask := outB.mask()
			if bMask > 0 && p&bMask == 0 {
				outP = p | bMask
				ok = true
				break outer
			}
		}
	}
	return
}
func (p palletMask) String() string {
	return fmt.Sprintf(
		"=========\n|%d %d %d %d|\n|%d %d %d %d|\n|%d %d %d %d|\n|%d %d %d %d|\n=========",
		(p>>15)&1, (p>>14)&1, (p>>13)&1, (p>>12)&1,
		(p>>11)&1, (p>>10)&1, (p>>9)&1, (p>>8)&1,
		(p>>7)&1, (p>>6)&1, (p>>5)&1, (p>>4)&1,
		(p>>3)&1, (p>>2)&1, (p>>1)&1, (p>>0)&1,
	)
}

func makePallet(in []box) pallet {
	var (
		pMask palletMask
		newB  box
		ok    bool
	)
	out := []box{}
	for _, b := range in {
		pMask, newB, ok = pMask.placeBox(b)
		if ok {
			out = append(out, newB)
		}
	}
	return pallet{out}
}

func biggestBoxesFirst(t *truck, outChan chan<- *truck, wg *sync.WaitGroup) {
	var (
		store = new(boxStorage)
		boxes = []box{}

		pm palletMask
		b  box
		ok bool
	)
	t.unload(store)
	out := &truck{id: t.id}
	for {
		pm, b, ok = store.findLargestBox(pm)
		if ok {
			boxes = append(boxes, b)
		} else {
			out.load(boxes)
			boxes = []box{}
			pm = 0
		}
		if store.count == 0 {
			out.load(boxes)
			break
		}
	}
	outChan <- out
	wg.Done()
}

func newRepacker(in <-chan *truck, out chan<- *truck) *repacker {
	wg := new(sync.WaitGroup)
	go func() {
		for t := range in {
			wg.Add(1)
			go biggestBoxesFirst(t, out, wg)
		}
		wg.Wait()
		close(out)
	}()
	return &repacker{}
}
