package editor

import (
	"github.com/nsf/termbox-go"
)

// Led is used to indicate which step the editor is on.
type Led struct {
	X        int
	Y        int
	step     int
	currStep *int
}

// Create a new Led, assiged to the provided step.
func LedNew(w *Window, x, y, step int, currStep *int) *Led {
	l := &Led{X: x, Y: y, step: step, currStep: currStep}
	w.Listen("draw", func(args ...interface{}) {
		l.Draw()
	})
	return l
}

// Draw the Led. It will «light» if the current step is the step this Led is assigned to.
func (l *Led) Draw() {
	if *l.currStep == l.step {
		termbox.SetCell(l.X, l.Y, '▬', CWhite, CDefault)
	} else {
		termbox.SetCell(l.X, l.Y, '▬', CBlack, CDefault)
	}
}
