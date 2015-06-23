package main

import (
	"fmt"
	"strings"
	"testing"
)

func tempTruck(convoyNumber int) *truck {
	var convoy = make([]string, 3)
	convoy[0] = `truck 97
0 0 1 1 101,1 1 1 1 102,2 2 1 1 103,0 3 1 4 104
0 0 1 1 105,1 0 1 1 106
0 0 4 4 107
endtruck
`
	convoy[1] = `truck 108
0 0 1 1 111,1 1 1 1 112,2 2 1 1 113,0 3 1 3 114
0 0 1 1 115,1 0 1 1 116
endtruck
`
	convoy[2] = `truck 0
endtruck
`

	r := newTruckReader(strings.NewReader(convoy[convoyNumber]))
	t, _ := r.Next()
	return t
}

func TestRepack(t *testing.T) {
	var out *truck
	boxQueue := make(cargoQueue, 0, 20)
	queueTotal := 0

	actualBox := ""
	expectedBox := ""
	in := tempTruck(0)
	out, boxQueue, queueTotal = repack(in, boxQueue, queueTotal)

	// confirm truck id
	expectedID := 97
	if out.id != expectedID {
		t.Errorf("repack() returns truck with id &v; expected truck.id %v", out.id, expectedID)
	}

	// confirm correct count of outgoing pallets
	actualPalletCount := len(out.pallets)
	expectedPalletCount := 1
	if actualPalletCount != expectedPalletCount {
		t.Errorf("repack() returned %v pallets on truck id %v; expected %v pallets", actualPalletCount, out.id, expectedPalletCount)
	}

	// confirm correct first box on first pallet
	actualBox = out.pallets[0].boxes[0].String()
	expectedBox = "0 0 4 4 107"
	if actualBox != expectedBox {
		t.Errorf("repack() returned first box on first pallet as %v; expected box %v", actualBox, expectedBox)
	}

	// confirm correct first box on remaining queue
	actualBox = boxQueue[0].item.String()
	expectedBox = "0 3 4 1 104"
	if actualBox != expectedBox {
		t.Errorf("first box in queue is %v; expected box %v", actualBox, expectedBox)
	}

	// confirm queueTotal
	expectedTotal := 9
	if queueTotal != expectedTotal {
		t.Errorf("queueTotal is %v; expected %v", queueTotal, expectedTotal)
	}

	// process second truck
	in = tempTruck(1)
	out, boxQueue, queueTotal = repack(in, boxQueue, queueTotal)

	// confirm truck id
	expectedID = 108
	if out.id != expectedID {
		t.Errorf("repack() returns truck with id %v; expected truck.id %v", out.id, expectedID)
	}

	// confirm correct count of outgoing pallets
	actualPalletCount = len(out.pallets)
	expectedPalletCount = 1
	if actualPalletCount != expectedPalletCount {
		t.Errorf("repack() returned %v pallets on truck id %v; expected %v pallets", actualPalletCount, out.id, expectedPalletCount)
	}

	// confirm first box on first pallet
	actualBox = out.pallets[0].boxes[0].String()
	expectedBox = "0 0 4 1 104"
	if actualBox != expectedBox {
		t.Errorf("repack() truck id %v returned first box on first pallet as %v; expected box %v", out.id, actualBox, expectedBox)
	}

	// confirm second box on first pallet
	actualBox = out.pallets[0].boxes[1].String()
	expectedBox = "1 0 3 1 114"
	if actualBox != expectedBox {
		t.Errorf("repack() truck id %v returned second box on first pallet as %v; expected box %v", out.id, actualBox, expectedBox)
	}

	// confirm first box on remaining queue has area of 1
	actualArea := boxQueue[0].area
	expectedArea := 1
	if actualArea != expectedArea {
		t.Errorf("first box in queue has area of %v; expected %v", actualArea, expectedArea)
	}

	// confirm queueTotal
	expectedTotal = 1
	if queueTotal != expectedTotal {
		t.Errorf("queueTotal is %v; expected %v", queueTotal, expectedTotal)
		fmt.Printf(" %v \n", boxQueue)
		fmt.Printf(" %v \n", out.pallets[0])
	}

	// process last truck
	in = tempTruck(2)
	out, boxQueue, queueTotal = repack(in, boxQueue, queueTotal)

	// confirm truck id
	expectedID = 0
	if out.id != expectedID {
		t.Errorf("repack() returns truck with id %v; expected truck.id %v", out.id, expectedID)
	}

	// confirm correct count of outgoing pallets
	actualPalletCount = len(out.pallets)
	expectedPalletCount = 1
	if actualPalletCount != expectedPalletCount {
		t.Errorf("repack() returned %v pallets on truck id %v; expected %v pallets", actualPalletCount, out.id, expectedPalletCount)
	}

	// confirm remaining queue is empty
	expectedLength := 0
	if len(boxQueue) != expectedLength {
		t.Errorf("queue has length %v; expected %v", len(boxQueue), expectedLength)
	}

}
