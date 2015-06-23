package main

import "testing"

func TestFindBiggestSpace(t *testing.T) {
	var basement [palletWidth][palletLength]boxInfo
	bs := boxStorage{basement}
	p, _ := palletFromString("0 0 1 1 1")

	for _, b := range p.boxes {
		bs.addBox(b)
	}

	ps := palletStorage{false, p}
	biggestSpace := ps.findBiggestSpace()
	expected := freeSpace{1, 0, 3, 4}

	if biggestSpace != expected {
		t.Error(biggestSpace, "is not the biggest space")
	}
}

func TestIsFreeSpace(t *testing.T) {
	var basement [palletWidth][palletLength]boxInfo
	bs := boxStorage{basement}
	p, _ := palletFromString("1 0 2 2 1")

	for _, b := range p.boxes {
		bs.addBox(b)
	}

	ps := palletStorage{false, p}
	expected := true
	x, y := 3, 0
	result := ps.isFreeSpace(x, y)
	if result != expected {
		if result {
			t.Error("x:", x, "y:", y, "is not a freespace")
		} else {
			t.Error("x:", x, "y:", y, "is a freespace")
		}
	}
}

func TestGetBiggestBox(t *testing.T) {
	var basement [palletWidth][palletLength]boxInfo
	bs := boxStorage{basement}
	bs.addBox(box{0, 0, 1, 1, 1})
	bs.addBox(box{0, 0, 2, 2, 2})
	bs.addBox(box{0, 0, 2, 1, 3})
	result := bs.findBiggestBox(4, 4)
	expected := box{0, 0, 2, 2, 2}
	if result != expected {
		t.Error(result, "is not the biggest box")
	}
}
