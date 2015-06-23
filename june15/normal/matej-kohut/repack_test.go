package main

import (
	"math/rand"
	"runtime"
	"testing"
)

func TestGetAllEmtyRectangles(t *testing.T) {
	tests := []struct {
		p pallet
		r []box
	}{
		{
			pallet{[]box{}},
			[]box{box{0, 0, palletLength, palletWidth, 0}},
		},
		{
			pallet{[]box{box{0, 0, 1, 1, 0}}},
			[]box{box{0, 1, palletLength - 1, palletWidth, 0}, box{1, 0, palletLength, palletWidth - 1, 0}},
		},
		{
			pallet{[]box{box{0, 0, 1, 2, 0}}},
			[]box{box{0, 1, palletLength - 1, palletWidth, 0}, box{2, 0, palletLength, palletWidth - 2, 0}},
		},
		{
			pallet{[]box{box{0, 0, palletLength, palletWidth, 0}}},
			[]box{},
		},
		{
			pallet{[]box{box{0, 0, 2, 2, 0}, box{0, 2, 1, 1, 1}}},
			[]box{box{0, 3, palletLength - 3, palletWidth, 0}, box{1, 2, palletLength - 2, palletWidth - 1, 0}, box{2, 0, palletLength, palletWidth - 2, 0}},
		},
	}
	for _, test := range tests {
		_, err := test.p.paint()
		if err != nil {
			t.Log(test)
			t.Log(test.p.String())
			t.Error("wrong test case: ", test)
		}
		pg, _ := test.p.paint()
		rectangles := getAllEmptyRectangles(pg)
		ok := true
		if len(rectangles) != len(test.r) {
			ok = false
		} else {
			for i, r := range rectangles {
				if r != test.r[i] {
					ok = false
					break
				}
			}
		}
		if !ok {
			t.Log(test)
			t.Log(test.p.String())
			t.Errorf("wrong empty rectangles: %v", rectangles)
		}
	}
}

func TestBoxScoreOnPallet(t *testing.T) {
	tests := []struct {
		b box
		p pallet
		s float64
	}{
		{
			box{0, 0, 0, 0, 0}, pallet{[]box{}},
			0.0,
		},
		{
			box{0, 0, 1, 1, 0}, pallet{[]box{}},
			50.0,
		},
		{
			box{1, 0, 1, 1, 0}, pallet{[]box{}},
			25.0,
		},
		{
			box{1, 1, 1, 1, 0}, pallet{[]box{}},
			0.0,
		},
		{
			box{0, 0, palletLength, palletWidth, 0}, pallet{[]box{}},
			100.0,
		},
		{
			box{0, 0, palletLength - 1, palletWidth - 1, 0}, pallet{[]box{}},
			50.0,
		},
		{
			box{1, 1, palletLength - 1, palletWidth - 1, 0}, pallet{[]box{}},
			50.0,
		},
		{
			box{0, 0, 1, 1, 0}, pallet{[]box{box{1, 1, 1, 1, 1}}},
			50.0,
		},
		{
			box{0, 0, 1, 1, 0}, pallet{[]box{box{0, 1, 1, 1, 1}}},
			75.0,
		},
		{
			box{0, 0, 1, 1, 0}, pallet{[]box{box{1, 0, 1, 1, 1}}},
			75.0,
		},
		{
			box{0, 0, 1, 1, 0}, pallet{[]box{box{1, 0, 1, 1, 1}, box{0, 1, 1, 1, 2}}},
			100.0,
		},
		{
			box{0, 0, palletLength, 1, 0}, pallet{[]box{box{1, 0, palletLength, 1, 1}}},
			100.0,
		},
		{
			box{0, 0, 1, palletWidth, 0}, pallet{[]box{box{0, 1, 1, palletWidth, 1}}},
			100.0,
		},
	}

	for _, test := range tests {
		test.p.boxes = append(test.p.boxes, test.b)
		_, err := test.p.paint()
		if err != nil {
			t.Log(test)
			t.Log(test.p.String())
			t.Error("wrong test case: ", test)
		}
		test.p.boxes = test.p.boxes[:len(test.p.boxes)-1]
		pg, _ := test.p.paint()
		score := boxScoreOnPallet(test.b, pg)
		if score != test.s {
			t.Log(test)
			test.p.boxes = test.p.boxes[:len(test.p.boxes)+1]
			t.Log(test.p.String())
			t.Error("wrong score: ", score)
		}
	}
}

func TestNormalPack(t *testing.T) {
	tests := []struct {
		b box
		p pallet
		r []box
	}{
		{
			box{0, 0, 1, 1, 0}, pallet{[]box{}},
			[]box{box{0, 0, 1, 1, 0}},
		},
		{
			box{1, 1, 1, 1, 0}, pallet{[]box{}},
			[]box{box{0, 0, 1, 1, 0}},
		},
		{
			box{0, 0, 5, 1, 0}, pallet{[]box{}},
			[]box{box{0, 0, 5, 1, 0}, box{0, 0, 1, 5, 0}},
		},
		{
			box{0, 0, 1, 5, 0}, pallet{[]box{}},
			[]box{box{0, 0, 1, 5, 0}, box{0, 0, 5, 1, 0}},
		},
	}

	for _, test := range tests {
		_, err := test.p.paint()
		if err != nil {
			t.Log(test)
			t.Log(test.p.String())
			t.Error("wrong test case: ", test)
		}
		placedBoxes := make(chan box)
		pg, _ := test.p.paint()
		go getNormalPlacements(test.b, pg, placedBoxes)
		boxes := []box{}
		for b := range placedBoxes {
			boxes = append(boxes, b)
		}
		ok := true
		if len(boxes) != len(test.r) {
			ok = false
		} else {
			for i, b := range boxes {
				if b != test.r[i] {
					ok = false
					break
				}
			}
		}
		if !ok {
			t.Log(test)
			t.Log(test.p.String())
			t.Errorf("wrong boxes: %v", boxes)
		}
	}
}

func initBoxes(n int, b *testing.B) []box {
	cpus := runtime.NumCPU()
	runtime.GOMAXPROCS(cpus)
	seed := 80085
	rand.Seed(int64(seed))

	boxes := make([]box, n)
	id = 1
	for i := 0; i < n; i++ {
		boxes[i] = genbox()
	}
	return boxes
}

func packBoxes(boxes []box, b *testing.B) {
	// benchmark repacker with one truck, with one pallet with 1000 boxes
	for n := 0; n < b.N; n++ {
		in := make(chan *truck)
		out := make(chan *truck)
		newRepacker(in, out)

		t := &truck{
			id: idLastTruck,
			pallets: []pallet{
				pallet{
					boxes: boxes,
				},
			},
		}

		go func() {
			defer close(in)
			in <- t
		}()

		// just get all trucks from repacker
		for _ = range out {
			// do not check if boxes are in good position, that has to be done in tests
		}
	}
}

func BenchmarkPack1Boxes(b *testing.B) {
	boxes := initBoxes(1, b)
	packBoxes(boxes, b)
}

func BenchmarkPack10Boxes(b *testing.B) {
	boxes := initBoxes(10, b)
	packBoxes(boxes, b)
}

func BenchmarkPack100Boxes(b *testing.B) {
	boxes := initBoxes(100, b)
	packBoxes(boxes, b)
}

func BenchmarkPack500Boxes(b *testing.B) {
	boxes := initBoxes(500, b)
	packBoxes(boxes, b)
}

func BenchmarkPack1000Boxes(b *testing.B) {
	boxes := initBoxes(1000, b)
	packBoxes(boxes, b)
}

/*
func BenchmarkPack10000Boxes(b *testing.B) {
	boxes := initBoxes(10000, b)
	packBoxes(boxes, b)
}

func BenchmarkPack20000Boxes(b *testing.B) {
	boxes := initBoxes(20000, b)
	packBoxes(boxes, b)
}

func BenchmarkPack50000Boxes(b *testing.B) {
  boxes := initBoxes(50000)
  packBoxes(boxes, b)
}
*/
