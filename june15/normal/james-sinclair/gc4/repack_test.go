package main

import (
	"log"
	"strings"
	"testing"
)

const boxStorageTestTruck = `truck 1
0 0 1 1 101
1 1 1 1 102
2 2 1 1 103
3 0 4 1 104
3 0 1 4 105
0 0 1 1 106
0 0 1 1 107
endtruck
`

func makeTestStore(s string) *boxStorage {
	r := newTruckReader(strings.NewReader(s))
	store := new(boxStorage)

	truck, err := r.Next()
	if err != nil {
		log.Fatal("truck read:", err)
	}

	truck.unload(store)
	return store
}

func makeEmptyTruck() *truck {
	return new(truck)
}

func TestBoxStorage(t *testing.T) {
	store := makeTestStore(boxStorageTestTruck)

	// Test boxStorage.String format
	expStorage := "boxStorage{ 5@1x1 0@2x1 0@2x2 0@3x1 0@3x2 0@3x3 2@4x1 0@4x2 0@4x3 0@4x4 }"
	if store.String() != expStorage {
		t.Error("boxStorage format is wrong")
		t.Log(store.String())
	}

	// Test boxStorage.finds largest box with an empty palletMask
	expLargestBoxsize := boxlist{b4x1, b4x1, b1x1, b1x1, b1x1, b1x1, b1x1}
	for _, size := range expLargestBoxsize {
		_, foundBox, ok := store.findLargestBox(0)
		if !ok {
			t.Error("boxStorage returned not ok with empty palletMask")
		}
		if foundBox.size() != size {
			t.Error("boxStorage returned incorrect boxsize:", foundBox.size())
		}
	}
	// Test boxStorage.findLargestBox returns not ok when storage is empty
	_, _, ok := store.findLargestBox(0)
	if ok {
		t.Error("boxStorage returned ok when empty")
	}

	store = makeTestStore(boxStorageTestTruck)

	// Test fetching from storage
	boxes := store.fetch(b4x4, b3x4)
	if len(boxes) != 0 {
		t.Error("boxStorage.fetch returned the wrong number of boxes: expected 0 got", len(boxes))
	}
	boxes = store.fetch(b4x1, b1x1, b1x1)
	if len(boxes) != 3 {
		t.Error("boxStorage.fetch returned the wrong number of boxes: expected 3 got", len(boxes))
	}
	truck := makeEmptyTruck()
	truck.load(boxes)
	if len(truck.pallets) != 1 {
		t.Error("truck.load pallet count incorrect: expected 1 got", len(truck.pallets))
	}
	expected := `
| ! ! ! ! |
| @ #     |
|         |
|         |
`
	if truck.pallets[0].String() != expected {
		t.Error("truck.load built the wrong pallet")
		t.Log(truck.pallets[0].String())
	}
}

func TestBox(t *testing.T) {
	expected := `=========
|1 0 0 0|
|1 0 0 0|
|0 0 0 0|
|0 0 0 0|
=========`
	b := box{0, 0, 1, 2, 0}
	if b.mask().String() != expected {
		t.Error("box.mask returned incorrect mask")
		t.Logf("\n%v", b.mask())
	}
	if b.size() != b2x1 {
		t.Error("box.size did not return the correct box size:", b.size())
	}
	expected = `=========
|0 0 0 0|
|0 0 1 0|
|0 0 1 0|
|0 0 0 0|
=========`
	shifted := b.shift(1, 2)
	if shifted.mask().String() != expected {
		t.Error("box.mask returned incorrect mask after shift")
		t.Logf("\n%v", shifted.mask())
	}
}

func TestBoxsizze(t *testing.T) {
	// Test we get the canonical box correctly
	input := boxlist{b1x2, b1x3, b1x4, b2x3, b2x4, b3x4}
	output := boxlist{b2x1, b3x1, b4x1, b3x2, b4x2, b4x3}
	for i, b := range input {
		if b.canon() != output[i] {
			t.Errorf("boxsize.canon returned incorrect boxsize: expected %v got %v", output[i], b.canon())
		}
	}

	// Test converting a boxsize back to a box
	expected := box{0, 0, 2, 3, 10}
	boxsizeBox := b2x3.box(10)
	if boxsizeBox != expected {
		t.Error("")
	}
}
