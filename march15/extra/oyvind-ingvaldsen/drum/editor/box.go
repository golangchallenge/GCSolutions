package editor

import (
	"github.com/nsf/termbox-go"
)

// A simple Box.
type Box struct {
	X          int
	Y          int
	Width      int
	Height     int
	Foreground termbox.Attribute
	Background termbox.Attribute
}

// Create a new Box (and listen to callbacks on the provided Window).
func BoxNew(win *Window, x, y, w, h int) *Box {
	b := &Box{X: x, Y: y, Width: w, Height: h, Foreground: CDefault, Background: CDefault}

	win.Listen("draw", func(args ...interface{}) {
		b.Draw()
	})

	return b
}

// Set the foreground and background color of the Box.
func (b *Box) SetColor(fg, bg termbox.Attribute) *Box {
	b.Foreground = fg
	b.Background = bg
	return b
}

// Check if a point is inside the Box.
func (b *Box) Inside(x, y int) bool {
	return x >= b.X-b.Width && x <= b.X+b.Width && y >= b.Y-b.Height && y <= b.Y+b.Height
}

// Draw the Box (by filling cells with spaces).
func (b *Box) Draw() {
	for y := b.Y - b.Height; y <= b.Y+b.Height; y++ {
		for x := b.X - b.Width; x <= b.X+b.Width; x++ {
			termbox.SetCell(x, y, ' ', b.Foreground, b.Background)
		}
	}
}
