package editor

import (
	"github.com/nsf/termbox-go"
)

// Represents a step in a Track.
type StepButton struct {
	Box
	stepRef *bool
}

// Create a new StepButton, with a reference to the step it is «assigned» to.
// The StepButton will also listen for mouse click events on the Window to handle toggling.
func StepButtonNew(w *Window, x, y int, stepRef *bool) *StepButton {
	b := &StepButton{Box: Box{X: x, Y: y, Width: 2, Height: 1}, stepRef: stepRef}

	w.Listen("draw", func(args ...interface{}) {
		if *stepRef {
			b.SetColor(CWhite, CBlue)
		} else {
			b.SetColor(CWhite, CBlack)
		}
		b.Draw()
		if *b.stepRef {
			termbox.SetCell(b.X, b.Y+1, '▬', CWhite, CBlue)
		} else {
			termbox.SetCell(b.X, b.Y+1, '▬', CBlue, CBlack)
		}
	})

	w.Listen("click", func(args ...interface{}) {
		x := args[0].(int)
		y := args[1].(int)
		if b.Inside(x, y) {
			*b.stepRef = !*b.stepRef
		}
	})

	return b
}
