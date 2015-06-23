package main

import "testing"

var table = []struct {
	r    Rectangle
	edx  uint8
	edy  uint8
	b    box
	fits bool
}{
	{Rectangle{0, 0, 4, 4}, 4, 4, box{0, 0, 4, 5, 1}, false},
	{Rectangle{0, 0, 2, 2}, 2, 2, box{0, 0, 2, 2, 1}, true},
	{Rectangle{1, 1, 2, 2}, 1, 1, box{0, 0, 2, 2, 1}, false},
	{Rectangle{2, 2, 4, 4}, 2, 2, box{0, 0, 1, 1, 1}, true},
	{Rectangle{3, 3, 4, 4}, 1, 1, box{0, 0, 2, 2, 1}, false},
	{Rectangle{3, 3, 4, 4}, 1, 1, box{0, 0, 1, 1, 1}, true},
	{Rectangle{4, 4, 4, 4}, 0, 0, box{0, 0, 1, 1, 1}, false},
	{Rectangle{1, 1, 3, 3}, 2, 2, box{0, 0, 1, 1, 1}, true},
	{Rectangle{1, 1, 1, 1}, 0, 0, box{0, 0, 1, 1, 1}, false},
	{Rectangle{0, 1, 4, 4}, 4, 3, box{0, 1, 4, 1, 1}, true},
}

func TestDx(t *testing.T) {
	for _, val := range table {
		if val.r.Dx() != val.edx {
			t.Errorf("Expected: %d, Got: %d\n", val.edx, val.r.Dx())
		}
	}
}

func TestDy(t *testing.T) {
	for _, val := range table {
		if val.r.Dy() != val.edy {
			t.Errorf("Expected: %d, Got: %d\n", val.edy, val.r.Dy())
		}
	}
}

func TestFits(t *testing.T) {
	for _, val := range table {
		if val.r.Fits(&val.b) != val.fits {
			t.Errorf("Mismatch in fit: %s, %s\n", val.r, val.b)
		}
	}
}

var tabler = []struct {
	r    Rectangle
	b    box
	fits bool
}{
	{Rectangle{0, 1, 1, 1}, box{0, 1, 1, 4, 1}, false},
	{Rectangle{0, 1, 4, 4}, box{0, 1, 1, 4, 1}, true},
}

func TestFitsR(t *testing.T) {
	for _, val := range tabler {
		if val.r.FitsR(&val.b) != val.fits {
			t.Errorf("Mismatch in fit: %s, %s\n", val.r, val.b)
		}
	}
}
