package tile

import (
	"image"
	"testing"
)

// Slices must be in the same order
func SlicesEqual(a, b []image.Rectangle) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestByCount(t *testing.T) {
	cases := []struct {
		bounds     image.Rectangle
		wide, high int
		result     []image.Rectangle
	}{
		{
			image.Rect(0, 0, 10, 10), 1, 1,
			[]image.Rectangle{image.Rect(0, 0, 10, 10)},
		},
		{
			image.Rect(0, 0, 10, 10), 2, 2,
			[]image.Rectangle{
				image.Rect(0, 0, 5, 5),
				image.Rect(5, 0, 10, 5),
				image.Rect(0, 5, 5, 10),
				image.Rect(5, 5, 10, 10)},
		},
		{
			image.Rect(10, 10, 20, 20), 2, 2,
			[]image.Rectangle{
				image.Rect(10, 10, 15, 15),
				image.Rect(15, 10, 20, 15),
				image.Rect(10, 15, 15, 20),
				image.Rect(15, 15, 20, 20)},
		},
	}
	for _, c := range cases {
		got, _ := ByCount(c.bounds, c.wide, c.high)
		if !SlicesEqual(c.result, got) {
			t.Errorf("ByCount(%q, %v, %v) = %q, wanted %q", c.bounds, c.wide,
				c.high, got, c.result)
		}
	}
}

func TestByPixel(t *testing.T) {
	cases := []struct {
		bounds     image.Rectangle
		wide, high int
		result     []image.Rectangle
	}{
		{
			image.Rect(0, 0, 10, 10), 10, 10,
			[]image.Rectangle{image.Rect(0, 0, 10, 10)},
		},
		{
			image.Rect(0, 0, 10, 10), 5, 5,
			[]image.Rectangle{
				image.Rect(0, 0, 5, 5),
				image.Rect(5, 0, 10, 5),
				image.Rect(0, 5, 5, 10),
				image.Rect(5, 5, 10, 10)},
		},
	}
	for _, c := range cases {
		got, _ := ByPixel(c.bounds, c.wide, c.high)
		if !SlicesEqual(c.result, got) {
			t.Errorf("ByCount(%q, %v, %v) = %q, wanted %q", c.bounds, c.wide,
				c.high, got, c.result)
		}
	}
}
