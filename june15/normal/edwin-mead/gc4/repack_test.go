package main

import (
	"strings"
	"testing"
)

func TestNewRepacker(t *testing.T) {
	testData := testTrucks()

	in := make(chan *truck)
	out := make(chan *truck)
	newRepacker(in, out)

	for palletIn, palletOut := range testData {
		r := newTruckReader(strings.NewReader(palletIn))
		truck, _ := r.Next()
		in <- truck
		repackedTruck := <-out
		if len(repackedTruck.pallets) != 1 {
			t.Error("Too many pallets")
		}
		if repackedTruck.pallets[0].String() != palletOut {
			t.Error("Bad packing")
		}
	}
	close(in)
}

// set up pallets with combinations of boxes to test repacking
func testTrucks() map[string]string {
	return map[string]string{
		`truck 1
0 0 4 4 2
endtruck
`: `
| ! ! ! ! |
| ! ! ! ! |
| ! ! ! ! |
| ! ! ! ! |
`,
		`truck 3
0 0 4 3 4
0 0 4 1 5
endtruck
`: `
| ! ! ! @ |
| ! ! ! @ |
| ! ! ! @ |
| ! ! ! @ |
`,
		`truck 6
0 0 4 2 7
0 0 4 2 8
endtruck
`: `
| ! ! @ @ |
| ! ! @ @ |
| ! ! @ @ |
| ! ! @ @ |
`,
		`truck 9  
0 0 4 2 10 
0 0 4 1 11 
endtruck
`: `
| ! ! @   |
| ! ! @   |
| ! ! @   |
| ! ! @   |
`,
		`truck 12 
0 0 4 2 13 
0 0 3 2 14 
0 0 1 1 15 
0 0 1 1 16 
endtruck
`: `
| ! ! @ @ |
| ! ! @ @ |
| ! ! @ @ |
| ! ! # $ |
`,
		`truck 17 
0 0 4 2 18 
0 0 2 1 19 
0 0 2 1 20 
0 0 2 1 21 
0 0 1 1 22 
0 0 1 1 23 
endtruck
`: `
| ! ! @ # |
| ! ! @ # |
| ! ! $ % |
| ! ! $ ^ |
`,
		`truck 24
0 0 4 1 25
0 0 3 1 26
0 0 1 1 27
0 0 3 2 28
0 0 2 1 29
endtruck
`: `
| ! @ $ $ |
| ! @ $ $ |
| ! @ $ $ |
| ! # % % |
`,
		`truck 30
0 0 3 3 31
0 0 3 1 32
0 0 1 2 33
0 0 2 1 34
endtruck
`: `
| ! ! ! # |
| ! ! ! # |
| ! ! ! $ |
| @ @ @ $ |
`,
		`truck 35
0 0 3 3 36
0 0 2 1 37
0 0 1 1 38
0 0 2 1 39
0 0 1 1 40
0 0 1 1 41
endtruck
`: `
| ! ! ! $ |
| ! ! ! $ |
| ! ! ! % |
| @ @ # ^ |
`,
		`truck 42
0 0 3 3 43
0 0 1 1 44
0 0 1 1 45
0 0 1 1 46
0 0 1 1 47
endtruck
`: `
| ! ! ! % |
| ! ! !   |
| ! ! !   |
| @ # $   |
`,
		`truck 48
0 0 3 2 49
0 0 2 1 50
0 0 3 1 51
0 0 1 1 52
endtruck
`: `
| ! ! #   |
| ! ! #   |
| ! ! #   |
| @ @ $   |
`,
		`truck 53
0 0 2 3 54
0 0 1 1 55
0 0 1 1 56
0 0 2 2 57
endtruck
`: `
| ! ! $ $ |
| ! ! $ $ |
| ! !     |
| @ #     |
`,
		`truck 58
0 0 3 1 59
0 0 1 1 60
endtruck
`: `
| !       |
| !       |
| !       |
| @       |
`,
		`truck 61 
0 0 1 1 62 
0 0 1 1 63 
0 0 1 1 64 
0 0 1 1 65 
endtruck
`: `
| ! @     |
| $ #     |
|         |
|         |
`,
		`truck 66
endtruck
`: `
|         |
|         |
|         |
|         |
`,
	}
}
