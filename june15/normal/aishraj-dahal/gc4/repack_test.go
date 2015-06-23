package main

import "testing"

func TestFitness(t *testing.T) {
	testBox := box{x: 0, y: 0, w: 1, l: 1, id: 1}
	_, shelves := makePallet(testBox)
	if len(shelves) > 1 {
		t.Errorf("The length of the shelves cannot exceed 1 after creating a single pallet. Current length is %v", len(shelves))
	}
	if len(shelves) <= 0 {
		t.Error("The length of shelves cannot be empty or less than 0.")
	}

	anotherBox := box{x: 0, y: 3, w: 2, l: 2, id: 2}
	boxFitness := findFit(anotherBox, &shelves)
	exepctedFitness := newShelfFit
	if boxFitness != exepctedFitness {
		t.Errorf("The box does not fit the shelf. Expected %v got %v", exepctedFitness, boxFitness)
	}
}

func TestPacking(t *testing.T) {
	demoTruck := truck{id: 4, pallets: nil}
	sampleBox1 := box{x: 0, y: 3, w: 2, l: 2, id: 2}
	sampleBox2 := box{x: 0, y: 2, w: 3, l: 3, id: 1}
	samplePallet := pallet{boxes: []box{sampleBox1, sampleBox2}}
	demoTruck.pallets = append(demoTruck.pallets, samplePallet)
	outPallets := shelfNF(&demoTruck)
	if len(outPallets.pallets) != 2 {
		t.Error("Exepected 2 pallets int truck. Got :", len(outPallets.pallets))
	}
}
