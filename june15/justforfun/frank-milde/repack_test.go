//
// =========================================================================
//
//       Filename:  repack_test.go
//
//    Description:
//
//        Version:  1.0
//        Created:  06/21/2015 10:48:05 PM
//       Revision:  none
//       Compiler:  g++
//
//          Usage:  <+USAGE+>
//
//         Output:  <+OUTPUT+>
//
//         Author:  Frank Milde (FM), frank.milde (at) posteo.de
//        Company:
//
//        License:  GNU General Public License
//      Copyright:  Copyright (c) 2015, Frank Milde
//
// =========================================================================
//

package main

import (
	"fmt"
	"strings"
	"testing"
)

func Test_betterPacker(t *testing.T) {

	Truck1 := `truck 1
2 2 1 1 2,1 2 4 2 3
1 0 3 4 5,2 2 4 1 6
0 2 2 1 8,0 0 4 1 9
1 0 2 3 10,0 1 3 2 11,0 1 2 2 12
0 2 2 2 14
1 1 2 3 15
0 1 1 3 17
0 1 1 4 18
0 0 3 2 19,1 1 4 2 20
2 2 2 1 22
2 0 4 3 23
2 2 2 1 25,0 2 3 1 26,2 0 1 3 27
1 2 2 2 29,1 0 2 1 30,0 1 4 2 31
0 0 1 1 429,0 1 2 1 430
0 2 4 4 431
2 1 3 1 433,0 2 3 1 434,0 1 3 1 435
2 2 4 1 436,0 0 3 2 437
1 2 2 4 439,0 2 4 2 440
2 0 4 3 442,1 0 2 2 443
0 0 4 2 445
0 1 4 2 446
1 1 2 2 447,2 0 2 4 448,2 0 3 1 449
1 1 3 4 451,2 2 1 4 452
1 2 2 3 454,2 0 2 3 455,0 0 1 2 456
1 0 3 3 458
0 2 4 1 459
endtruck
`
	r := newTruckReader(strings.NewReader(Truck1))

	truck1, err := r.Next()
	if err != nil {
		t.Fatal("truck read:", err)
	}
	store := NewTable()
	got := betterPacker(truck1, &store)
	/*
		b1 := box{0, 1, 2, 1, 103}
		p1 := pallet{[]box{b1}}

		b2 := box{1, 0, 1, 2, 103}
		p2 := pallet{[]box{b2}}

		b3 := box{1, 1, 2, 2, 103}
		p3 := pallet{[]box{b3}}

		fmt.Println("x: ", b1.x, "  y: ", b1.y, "  w: ", b1.w, "  l: ", b1.l)
		fmt.Println(p1)
		fmt.Println("x: ", b2.x, "  y: ", b2.y, "  w: ", b2.w, "  l: ", b2.l)
		fmt.Println(p2)
		fmt.Println("x: ", b3.x, "  y: ", b3.y, "  w: ", b3.w, "  l: ", b3.l)
		fmt.Println(p3)
	*/

	fmt.Println("output truck: \n", got)
	/*
		type inputs struct {
			store Table
			Truck *truck
		}
		type outputs struct {
			store Table
			Truck *truck
		}
		tests := []struct {
			in   inputs
			want outputs
		}{
			// filled Store
			{
				in: inputs{
					Table{
						Stack{
							box{0, 0, 1, 1, 101},
							box{0, 0, 1, 1, 111},
						},
						Stack{
							box{0, 0, 1, 2, 102},
							box{0, 0, 1, 2, 112},
						},
						Stack{
							box{0, 0, 1, 3, 103},
							box{0, 0, 1, 3, 113},
						},
						Stack{
							box{0, 0, 1, 4, 104},
							box{0, 0, 1, 4, 114},
						},
						Stack{
							box{0, 0, 2, 2, 105},
							box{0, 0, 2, 2, 115},
						},
						Stack{
							box{0, 0, 2, 3, 106},
							box{0, 0, 2, 3, 116},
						},
						Stack{
							box{0, 0, 2, 4, 107},
							box{0, 0, 2, 4, 117},
						},
						Stack{},
						Stack{
							box{0, 0, 3, 4, 109},
							box{0, 0, 3, 4, 119},
						},
						Stack{},
					},
					&truck{
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
						},
					},
				},
				want: outputs{
					Table{
						Stack{
							box{0, 0, 1, 1, 101},
							box{0, 0, 1, 1, 111},
						},
						Stack{
							box{0, 0, 1, 2, 102},
							box{0, 0, 1, 2, 112},
						},
						Stack{
							box{0, 0, 1, 3, 103},
							box{0, 0, 1, 3, 113},
						},
						Stack{
							box{0, 0, 1, 4, 104},
							box{0, 0, 1, 4, 114},
						},
						Stack{
							box{0, 0, 2, 2, 105},
							box{0, 0, 2, 2, 115},
						},
						Stack{
							box{0, 0, 2, 3, 106},
						},
						Stack{
							box{0, 0, 2, 4, 107},
							box{0, 0, 2, 4, 117},
						},
						Stack{},
						Stack{
							box{0, 0, 3, 4, 109},
							box{0, 0, 3, 4, 119},
						},
						Stack{},
					},
					&truck{
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
						},
					},
				},
			},
		}

		for _, test := range tests {
			got := betterPacker(test.in.Truck, test.in.store)
			if !TablesAreEqual(test.in.store, test.want.store) {
				t.Errorf("Table error")
				t.Errorf("Got: \n%v ", test.in.store)
				t.Errorf("Want:\n%v ", test.want.store)
				t.Errorf("\n")
			}
			if !TrucksAreEqual(*got, *test.want.Truck) {
				t.Errorf("Pallets error")
				t.Errorf("Got: \n%v ", got)
				t.Errorf("Want:\n%v ", test.want.Truck)
				t.Errorf("\n")
			}
		}
	*/
}
