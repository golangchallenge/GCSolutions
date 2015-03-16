package editor

import (
	"github.com/nsf/termbox-go"
)

// A simple text label.
type Label struct {
	X          int
	Y          int
	Text       string
	Foreground termbox.Attribute
	Background termbox.Attribute
}

// Create a new Label (and listen for events on the provided Window).
func LabelNew(w *Window, x, y int, text string) *Label {
	l := &Label{X: x, Y: y, Text: text}
	l.SetColor(CDefault, CDefault)

	w.Listen("draw", func(args ...interface{}) {
		l.Draw()
	})

	return l
}

// Set foreground and background color of the Label.
func (l *Label) SetColor(fg, bg termbox.Attribute) *Label {
	l.Foreground = fg
	l.Background = bg
	return l
}

// Draw the Label. Note that Label's text is centered.
func (l *Label) Draw() {
	hl := len(l.Text) / 2
	for i := 0; i < len(l.Text); i++ {
		termbox.SetCell(l.X+i-hl, l.Y, rune(l.Text[i]), l.Foreground, l.Background)
	}
}
