package mosaic

import (
	// "fmt"
	"image"
	"image/color"
	"image/draw"
	"testing"
)

func TestAverageColor(t *testing.T) {
	var tests = []color.RGBA{
		color.RGBA{255, 0, 0, 255},
		color.RGBA{255, 255, 0, 255},
		color.RGBA{255, 255, 255, 255},
		color.RGBA{0, 0, 0, 255},
		color.RGBA{10, 10, 10, 255},
		color.RGBA{250, 1, 1, 255},
	}
	for _, c := range tests {
		m := image.NewRGBA(image.Rect(0, 0, 1000, 1000))
		draw.Draw(m, m.Bounds(), &image.Uniform{c}, image.ZP, draw.Src)
		got := averageColor(m)
		if got != c {
			t.Errorf("AverageColor(%v) = %v, want %v", c, got, c)
		}
		got = averageColorBound(m, 0, 0, 2, 2)
		if got != c {
			t.Errorf("AverageColorBound(%v) = %v, want %v", c, got, c)
		}
	}
}

func TestColorDistance(t *testing.T) {
	var tests = []struct {
		a, b color.RGBA
		want int
	}{
		{color.RGBA{0, 0, 0, 255}, color.RGBA{0, 0, 0, 255}, 0},
		{color.RGBA{255, 255, 255, 255}, color.RGBA{255, 255, 255, 255}, 0},
		{color.RGBA{0, 0, 0, 255}, color.RGBA{1, 1, 1, 255}, 1},
		{color.RGBA{0, 0, 0, 255}, color.RGBA{255, 255, 255, 255}, 441},
	}
	for _, c := range tests {
		got := int(colorDistance(c.a, c.b))
		if got != c.want {
			t.Errorf("distance(%v, %v) = %v, want %v", c.a, c.b, got, c.want)
		}
	}
}

func TestPixelSize(t *testing.T) {
	var tests = []struct {
		w, h         int
		size, w2, h2 int
	}{
		{2, 2, 1, 2, 2},
		{10, 10, 1, 10, 10},
		{450, 199, 1, 450, 199},
		{384, 250, 1, 384, 250},
		{1600, 1200, 8, 1600, 1200},
		{1601, 1200, 8, 1600, 1200},
		{635, 631, 4, 632, 628},
		{787, 361, 2, 786, 360},
	}
	for _, c := range tests {
		size, w2, h2 := pixelSize(c.w, c.h)
		if size != c.size || w2 != c.w2 || h2 != c.h2 {
			t.Errorf("pixelSize(%v, %v) = (%v, %v, %v) want (%v, %v, %v)", c.w, c.h, size, w2, h2, c.size, c.w2, c.h2)
		}
	}

}
