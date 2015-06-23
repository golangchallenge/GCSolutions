//
// =========================================================================
//
//       Filename:  trucks_test.go
//
//    Description:  Unit test for the trucks.go file.
//
//        License:  GNU General Public License
//      Copyright:  Copyright (c) 2015, Frank Milde
//
// =========================================================================
//

package main

import (
	"testing"
)

func Test_Unload_TruckWithSomePallets(t *testing.T) {

	truck1 := &truck{
		1, []pallet{
			pallet{
				[]box{
					box{0, 0, 1, 1, 101},
				},
			},
			pallet{
				[]box{
					box{1, 1, 1, 1, 102},
					box{2, 2, 1, 1, 103},
				},
			},
			pallet{
				[]box{
					box{3, 0, 4, 1, 104},
					box{0, 0, 1, 1, 105},
					box{0, 0, 1, 1, 106},
					box{0, 0, 3, 4, 107},
				},
			},
		},
	}

	want := Table{
		Stack{ // 1
			box{0, 0, 1, 1, 101},
			box{0, 0, 1, 1, 102},
			box{0, 0, 1, 1, 103},
			box{0, 0, 1, 1, 105},
			box{0, 0, 1, 1, 106},
		},
		Stack{}, // 2
		Stack{}, // 3
		Stack{ // 4
			box{0, 0, 4, 1, 104},
		},
		Stack{}, // 5
		Stack{}, // 6
		Stack{}, // 8
		Stack{}, // 9
		Stack{ // 12
			box{0, 0, 4, 3, 107},
		},
		Stack{}, // 16
	} // end Stack

	gotTable := NewTable()

	gotNrPallets := truck1.Unload(gotTable)
	if !TablesAreEqual(gotTable, want) {
		t.Errorf("Comparing Tables:\n")
		t.Errorf("Got: \n%v ", gotTable)
		t.Errorf("Want:\n%v ", want)
	}
	if gotNrPallets != 3 {
		t.Errorf("Nr pallets: Got %d, want %d.", gotNrPallets, 3)
	}
}

func Test_Unload_TruckWithEmptyPallets(t *testing.T) {

	truck1 := &truck{
		1, []pallet{
			pallet{
				[]box{
					box{0, 0, 1, 1, 101},
				},
			},
			pallet{
				[]box{
					box{1, 1, 1, 1, 102},
					box{2, 2, 1, 1, 103},
				},
			},
			pallet{
				[]box{},
			},
			pallet{},
		},
	}
	want := Table{
		Stack{ // 1
			box{0, 0, 1, 1, 101},
			box{0, 0, 1, 1, 102},
			box{0, 0, 1, 1, 103},
		},
		Stack{}, // 2
		Stack{}, // 3
		Stack{}, // 4
		Stack{}, // 5
		Stack{}, // 6
		Stack{}, // 8
		Stack{}, // 9
		Stack{}, // 12
		Stack{}, // 16
	} // end Stack

	gotTable := NewTable()

	gotNrPallets := truck1.Unload(gotTable)
	if !TablesAreEqual(gotTable, want) {
		t.Errorf("Comparing Tables:\n")
		t.Errorf("Got: \n%v ", gotTable)
		t.Errorf("Want:\n%v ", want)
	}
	if gotNrPallets != 4 {
		t.Errorf("Nr pallets: Got %d, want %d.", gotNrPallets, 3)
	}
}
