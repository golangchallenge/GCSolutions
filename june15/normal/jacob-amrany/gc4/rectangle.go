package main

import "fmt"

//Rectangle represented by its borders. The Right and Bottom borders are exclusive, while the Left and Top are inclusive. This is different from go's image/Rectangle, but makes this algorithm much easier to represent
type Rectangle struct {
	Left, Top, Right, Bottom uint8
}

//Dx is the delta of the width. Even though boxes use x as a vertical axis, I wanted to keep true to Go's image/Rectangle package. The width of the rectangle.
func (r *Rectangle) Dx() uint8 {
	return r.Right - r.Left
}

//Dy is the vertical height of the rectangle.
func (r *Rectangle) Dy() uint8 {
	return r.Bottom - r.Top
}

//Fits checks if the box fits inside the bounds of the rectangle
func (r *Rectangle) Fits(b *box) bool {
	dx, dy := r.Dx(), r.Dy()
	if b.w <= dx && b.l <= dy {
		return true
	}
	return false
}

//FitsR checks if the box fits after rotation
func (r *Rectangle) FitsR(b *box) bool {
	dx, dy := r.Dx(), r.Dy()

	if b.l <= dx && b.w <= dy {
		return true
	}
	return false
}

//PerfectFit returns true if the box fits perfectly into the rectangle
func (r *Rectangle) PerfectFit(b *box) bool {
	dx, dy := r.Dx(), r.Dy()
	if b.w == dx && b.l == dy {
		return true
	}
	return false
}

//PerfectFitR returns true if a perfect fit on the rotation
func (r *Rectangle) PerfectFitR(b *box) bool {
	dx, dy := r.Dx(), r.Dy()
	if b.l == dx && b.w == dy {
		return true
	}
	return false
}

//Pretty printing
func (r Rectangle) String() string {
	return fmt.Sprintf("l:%d, t:%d, r:%d, b:%d", r.Left, r.Top, r.Right, r.Bottom)
}
