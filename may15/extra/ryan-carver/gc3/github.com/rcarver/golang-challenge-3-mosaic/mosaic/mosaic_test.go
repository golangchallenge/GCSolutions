package mosaic

import (
	"image"
	"image/color"
	"image/color/palette"
	"testing"
)

func TestComposeSquare(t *testing.T) {
	// Non-square image becomes square.
	in := image.NewRGBA(image.Rect(0, 0, 100, 200))
	pal := NewSolidPalette(palette.WebSafe)
	// Output bounds are units * unit size in both dimensions.
	m := ComposeSquare(in, 10, 75, pal)
	if got, want := m.Bounds().Dx(), 750; got != want {
		t.Errorf("Dx got %d, want %d", got, want)
	}
	if got, want := m.Bounds().Dy(), 750; got != want {
		t.Errorf("Dy got %d, want %d", got, want)
	}
}

func TestCompose(t *testing.T) {
	// Non-square image bounds have unspecified behavior right now.
	in := image.NewRGBA(image.Rect(0, 0, 100, 200))
	pal := NewSolidPalette(palette.WebSafe)
	// Output bounds are units * unit size in both dimensions.
	m := Compose(in, 10, 20, 35, 55, pal)
	if got, want := m.Bounds().Dx(), 350; got != want {
		t.Errorf("Dx got %d, want %d", got, want)
	}
	if got, want := m.Bounds().Dy(), 1100; got != want {
		t.Errorf("Dy got %d, want %d", got, want)
	}
}

func TestShrink(t *testing.T) {
	in := image.NewRGBA(image.Rect(0, 0, 100, 500))
	out := Shrink(in, .5)
	if got, want := out.Bounds().Dx(), 50; got != want {
		t.Errorf("Dx got %d, want %d", got, want)
	}
	if got, want := out.Bounds().Dy(), 250; got != want {
		t.Errorf("Dy got %d, want %d", got, want)
	}
}

func TestMosiac_Dither(t *testing.T) {
	in := solidImg(image.Rect(0, 0, 100, 100), color.White)
	mos := Mosaic{UnitsX: 10, UnitsY: 10, img: in}
	out := mos.Dither(palette.WebSafe)
	if _, ok := out.(*image.Paletted); !ok {
		t.Fatalf("want a Paletted image")
	}
	if got, want := out.Bounds().Dx(), 10; got != want {
		t.Errorf("x got %d, want %d", got, want)
	}
	if got, want := out.Bounds().Dy(), 10; got != want {
		t.Errorf("y got %d, want %d", got, want)
	}
}

func TestMosiac_Compose(t *testing.T) {
	in := solidImg(image.Rect(0, 0, 500, 500), color.White)
	mos := Mosaic{10, 10, 10, 10, in}
	pal := NewSolidPalette(palette.WebSafe)
	out := mos.Compose(pal)
	if got, want := out.Bounds().Dx(), 100; got != want {
		t.Errorf("x got %d, want %d", got, want)
	}
	if got, want := out.Bounds().Dy(), 100; got != want {
		t.Errorf("y got %d, want %d", got, want)
	}
}

func Test_dither(t *testing.T) {
	m := solidImg(image.Rect(0, 0, 100, 100), color.White)
	o := dither(m, palette.WebSafe)
	if _, ok := o.(*image.Paletted); !ok {
		t.Fatalf("want a Paletted image")
	}
}

func Test_downsample(t *testing.T) {
	c := color.RGBA{100, 120, 140, 255}
	m := solidImg(image.Rect(0, 0, 500, 500), c)
	dm := downsample(m, 100, 100, 1, 0.5)
	if got, want := dm.Bounds().Dx(), 100; got != want {
		t.Errorf("x got %d, want %d", got, want)
	}
	if got, want := dm.Bounds().Dy(), 100; got != want {
		t.Errorf("y got %d, want %d", got, want)
	}
}

func Test_average(t *testing.T) {
	c := color.RGBA64{100, 120, 140, 65535}
	m := image.NewUniform(c)
	a := average(m, image.Rect(0, 0, 100, 100), 20)
	if a != c {
		t.Errorf("average() got %v, want %v", a, c)
	}
}
