package main

import (
	"fmt"
	"math/rand"
)

func generate(n, seed int) {
	rand.Seed(int64(seed))

	for i := 0; i < n; i++ {
		t := gentruck()

		fmt.Println("truck", t.id)
		for _, p := range t.pallets {
			fmt.Println(p.OneLine())
		}
		fmt.Println("endtruck")
	}
}

var id = 1

func nextid() (newid int) {
	newid = id
	id++
	return
}

func gentruck() *truck {
	t := &truck{id: nextid()}
	np := randRange(5, 10)

	for i := 0; i < np; i++ {
		t.pallets = append(t.pallets, genpal())
	}
	return t
}

func genpal() (p pallet) {
	maxsq := palletWidth * palletLength

	nb := randRange(1, 5)

	sq := 0
	for i := 0; i < nb; i++ {
		b := genbox()
		sq += int(b.w) * int(b.l)
		// Stop once the surface area of the boxes is more than the pallet.
		if sq > maxsq {
			break
		}
		p.boxes = append(p.boxes, b)
	}

	return
}

func randRange(low, high int) int {
	return rand.Intn(high) + low
}

func genbox() (b box) {
	b.x = uint8(randRange(0, 3))
	b.y = uint8(randRange(0, 3))
	b.w = uint8(randRange(1, 4))
	b.l = uint8(randRange(1, 4))
	b.id = uint32(nextid())
	return
}
