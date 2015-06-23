package main

import "testing"

func TestFindOpenSpotWithRoom(t *testing.T) {
	boxOne, _ := boxFromString("0 0 4 2 1")
	boxTwo, _ := boxFromString("0 0 4 2 2")

	p := pallet{boxes: []box{boxOne}}

	x, idx := findOpenSpot([]pallet{p}, boxTwo)

	if idx != 0 || x != 2 {
		t.Fatalf("returned pallet idx %v & x edge %v, expected 0, 2", idx, x)
	}
}

func TestFindOpenSpotWithoutRoom(t *testing.T) {
	boxOne, _ := boxFromString("0 0 4 4 1")
	boxTwo, _ := boxFromString("0 0 4 2 2")

	p := pallet{boxes: []box{boxOne}}

	x, idx := findOpenSpot([]pallet{p}, boxTwo)

	if idx != -1 || x != 0 {
		t.Fatalf("returned pallet idx %v & x edge %v, expected -1, 0", idx, x)
	}
}
