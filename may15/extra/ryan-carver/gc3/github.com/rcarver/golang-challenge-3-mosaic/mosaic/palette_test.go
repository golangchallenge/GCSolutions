package mosaic

import (
	"image"
	"image/color"
	"image/color/palette"
	"image/draw"
	"testing"
)

func TestNewImagePalette(t *testing.T) {
	ip := NewImagePalette(256)
	m := ip.AtColor(color.RGBA{0, 0, 255, 255})
	if m != nil {
		t.Fatalf("want nil image")
	}
}

func TestNewSolidImagePalette(t *testing.T) {
	ip := NewSolidPalette(palette.WebSafe)
	blue := color.RGBA{0, 0, 255, 255}
	m := ip.AtColor(blue)
	if m == nil {
		t.Fatalf("want non-nil image")
	}
	c := m.At(0, 0)
	if c != blue {
		t.Fatalf("got %v, want blue", c)
	}
}

func solidImg(box image.Rectangle, c color.Color) image.Image {
	m := image.NewRGBA(box)
	draw.Draw(m, m.Bounds(), &image.Uniform{c}, image.ZP, draw.Over)
	return m
}

func TestImagePalette_Add(t *testing.T) {
	ip := NewImagePalette(3)
	box := image.Rect(0, 0, 100, 100)

	if ip.NumColors() != 0 {
		t.Errorf("want NumColors zero")
	}
	if ip.NumImages() != 0 {
		t.Errorf("want NumImages 0")
	}

	// Test that colors are slotted, then accumulated.
	tests := []struct {
		numColors int
		numImages int
		color     color.Color
	}{
		{1, 1, color.RGBA{0, 0, 255, 255}},
		{1, 2, color.RGBA{0, 0, 255, 255}},
		{2, 3, color.RGBA{0, 255, 255, 255}},
		{2, 4, color.RGBA{0, 255, 255, 255}},
		{3, 5, color.RGBA{255, 255, 255, 255}},
		{3, 6, color.RGBA{255, 255, 255, 255}},
		{3, 7, color.RGBA{25, 25, 25, 255}},
	}
	for i, want := range tests {
		m := solidImg(box, want.color)
		ip.Add(m)
		if got, want := ip.NumColors(), want.numColors; got != want {
			t.Errorf("%d got size %d, want %d", i, got, want)
		}
		if got, want := ip.NumImages(), want.numImages; got != want {
			t.Errorf("%d got NumImages %d, want %d", i, got, want)
		}
	}
}

func TestImagePalette_AtColor(t *testing.T) {
	ip := NewImagePalette(3)
	box := image.Rect(0, 0, 100, 100)

	// Fill the three color slots, then add more blues.
	colors := []color.Color{
		color.RGBA{0, 0, 255, 255},
		color.RGBA{0, 255, 255, 255},
		color.RGBA{255, 255, 255, 255},
		color.RGBA{0, 0, 254, 255},
		color.RGBA{0, 0, 253, 255},
	}
	for _, c := range colors {
		m := solidImg(box, c)
		ip.Add(m)
	}

	// Test color input / output.
	tests := []struct {
		in  color.Color
		out color.Color
	}{
		// Test that blue cycles through the options.
		{color.RGBA{0, 0, 255, 255}, color.RGBA{0, 0, 255, 255}},
		{color.RGBA{0, 0, 255, 255}, color.RGBA{0, 0, 254, 255}},
		{color.RGBA{0, 0, 255, 255}, color.RGBA{0, 0, 253, 255}},
		{color.RGBA{0, 0, 255, 255}, color.RGBA{0, 0, 255, 255}},
		// Test that other colors return closest.
		{color.RGBA{244, 244, 244, 255}, color.RGBA{255, 255, 255, 255}},
		{color.RGBA{244, 244, 244, 255}, color.RGBA{255, 255, 255, 255}},
	}
	for i, want := range tests {
		m := ip.AtColor(want.in)
		gotColor := m.At(0, 0)
		wantColor := want.out
		if gotColor != wantColor {
			t.Errorf("%d AtColor(%v) got %v, want %v", i, want.in, gotColor, wantColor)
		}
	}
}
