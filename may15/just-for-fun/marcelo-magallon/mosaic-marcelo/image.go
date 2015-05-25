package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"time"
)

type LinearTransformation struct {
	x0, y0 int
	dx, dy int
}

func NewLinearTransformation(x0, y0, x1, y1 int) LinearTransformation {
	return LinearTransformation{
		x0: x0,
		y0: y0,
		dx: x1 - x0,
		dy: y1 - y0,
	}
}

func (l LinearTransformation) t(x int) int {
	// x - x0   x1 - x0                   y1 - y0
	// ------ = -------  =>  y = (x - x0) ------- + y0
	// y - y0   y1 - y0                   x1 - x0

	// return (x-l.x0)*l.dy/l.dx + l.y0

	return (2*(x-l.x0)*l.dy/l.dx+1)/2 + l.y0
}

type ClampedLinearTransformation struct {
	l      LinearTransformation
	y0, y1 int
}

func NewClampedLinearTransformation(x0, y0, x1, y1 int) ClampedLinearTransformation {
	return ClampedLinearTransformation{
		l:  NewLinearTransformation(x0, y0, x1, y1),
		y0: y0,
		y1: y1,
	}
}

func (c ClampedLinearTransformation) t(x int) int {
	y := c.l.t(x)
	switch {
	case y < c.y0:
		y = c.y0
	case y > c.y1:
		y = c.y1
	}
	return y
}

type RGBA128 struct {
	// internally uint32 are stored to prevent overflows, but the
	// data is actually uint16
	R, G, B, A uint32
}

func (c *RGBA128) Add(rhs color.Color) {
	r, g, b, a := rhs.RGBA()
	c.R += r
	c.G += g
	c.B += b
	c.A += a
}

func (c *RGBA128) AddRGBA(r, g, b, a uint8) {
	c.R += uint32(r) << 8
	c.G += uint32(g) << 8
	c.B += uint32(b) << 8
	c.A += uint32(a) << 8
}

func (c *RGBA128) Div(s uint32) *RGBA128 {
	c.R /= s
	c.G /= s
	c.B /= s
	c.A /= s
	return c
}

func (c RGBA128) ToRGBA() color.RGBA {
	return color.RGBA{
		uint8(c.R >> 8),
		uint8(c.G >> 8),
		uint8(c.B >> 8),
		uint8(c.A >> 8),
	}
}

func downscale(ii image.Image, or image.Rectangle) *image.RGBA {
	if false {
		defer func(start time.Time) {
			fmt.Println(time.Since(start))
		}(time.Now())
	}

	switch ii.(type) {
	case *image.RGBA:
		return downscaleRGBA(ii.(*image.RGBA), or)
	}

	return downscaleImage(ii, or)
}

func downscaleImage(ii image.Image, or image.Rectangle) *image.RGBA {
	// ii: input image,     ir: input rectangle
	// ti: temporary image, tr: temporary rectangle
	// oi: output image,    or: output rectangle
	//
	// ii (ir)           ti (tr)     oi (or)
	// +----------+  =>  +----+  =>  +----+
	// |          |      |    |      |    |
	// |          |      |    |      +----+
	// |          |      |    |
	// +----------+      +----+

	ir := ii.Bounds()

	tr := image.Rect(or.Min.X, ir.Min.Y, or.Max.X, ir.Max.Y)
	ti := image.NewRGBA(tr)

	oi := image.NewRGBA(or)

	l := NewLinearTransformation(tr.Min.X, ir.Min.X, tr.Max.X, ir.Max.X)

	for j := tr.Min.Y; j < tr.Max.Y; j++ {
		jp := j
		for i := tr.Min.X; i < tr.Max.X; i++ {
			i0, i1 := l.t(i), l.t(i+1)
			t := RGBA128{}
			for ip := i0; ip < i1; ip++ {
				t.Add(ii.At(ip, jp))
			}
			ti.SetRGBA(i, j, t.Div(uint32(i1-i0)).ToRGBA())
		}
	}

	l = NewLinearTransformation(or.Min.Y, tr.Min.Y, or.Max.Y, tr.Max.Y)

	for i := or.Min.X; i < or.Max.X; i++ {
		ip := i
		for j := or.Min.Y; j < or.Max.Y; j++ {
			j0, j1 := l.t(j), l.t(j+1)
			t := RGBA128{}
			for jp := j0; jp < j1; jp++ {
				t.Add(ti.At(ip, jp))
			}
			oi.SetRGBA(i, j, t.Div(uint32(j1-j0)).ToRGBA())
		}
	}

	return oi
}

func downscaleRGBA(ii *image.RGBA, or image.Rectangle) *image.RGBA {
	// ii: input image,     ir: input rectangle
	// ti: temporary image, tr: temporary rectangle
	// oi: output image,    or: output rectangle
	//
	// ii (ir)           ti (tr)     oi (or)
	// +----------+  =>  +----+  =>  +----+
	// |          |      |    |      |    |
	// |          |      |    |      +----+
	// |          |      |    |
	// +----------+      +----+

	ir := ii.Bounds()

	tr := image.Rect(or.Min.X, ir.Min.Y, or.Max.X, ir.Max.Y)
	ti := image.NewRGBA(tr)

	oi := image.NewRGBA(or)

	c := NewClampedLinearTransformation(tr.Min.X, ir.Min.X, tr.Max.X, ir.Max.X)

	for j := tr.Min.Y; j < tr.Max.Y; j++ {
		jp := j
		for i, i0, i1 := tr.Min.X, 0, c.t(tr.Min.X); i < tr.Max.X; i++ {
			i0, i1 = i1, c.t(i+1)
			o0, o1 := ii.PixOffset(i0, jp), ii.PixOffset(i1, jp)
			t := RGBA128{}
			for ip := o0; ip < o1; ip += 4 {
				t.AddRGBA(ii.Pix[ip+0], ii.Pix[ip+1], ii.Pix[ip+2], ii.Pix[ip+3])
			}
			ti.SetRGBA(i, j, t.Div(uint32(i1-i0)).ToRGBA())
		}
	}

	c = NewClampedLinearTransformation(or.Min.Y, tr.Min.Y, or.Max.Y, tr.Max.Y)

	for i := or.Min.X; i < or.Max.X; i++ {
		ip := i
		for j, j0, j1 := or.Min.Y, 0, c.t(or.Min.Y); j < or.Max.Y; j++ {
			j0, j1 = j1, c.t(j+1)
			o0, o1 := ti.PixOffset(ip, j0), ti.PixOffset(ip, j1)
			t := RGBA128{}
			for jp := o0; jp < o1; jp += ti.Stride {
				t.AddRGBA(ti.Pix[jp+0], ti.Pix[jp+1], ti.Pix[jp+2], ti.Pix[jp+3])
			}
			oi.SetRGBA(i, j, t.Div(uint32(j1-j0)).ToRGBA())
		}
	}

	return oi
}

func downscale_example() {
	r, err := os.Open("checker.png")
	defer r.Close()
	if err != nil {
		log.Fatal(err)
	}

	i0, _, err := image.Decode(r)
	if err != nil {
		log.Fatal(err)
	}

	s := image.Rectangle{
		Min: image.Point{X: 0, Y: 0},
		Max: image.Point{X: i0.Bounds().Dx() / 2, Y: i0.Bounds().Dy() / 2},
	}
	i1 := downscale(i0, s)

	w, err := os.Create("checker_s.png")
	if err != nil {
		log.Fatal(err)
	}
	defer w.Close()
	if err := png.Encode(w, i1); err != nil {
		log.Fatal(err)
	}
}
