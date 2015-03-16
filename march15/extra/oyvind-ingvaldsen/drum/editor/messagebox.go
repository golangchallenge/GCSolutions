package editor

import (
	"github.com/nsf/termbox-go"
)

// Open a popup Window with a message.
func MessageBox(title, msg string, bg termbox.Attribute) {
	win := WindowNew(true)
	BoxNew(win, ScrHW, ScrHH, 40, 4).SetColor(CWhite, bg)
	LabelNew(win, ScrHW, ScrHH-3, title).SetColor(CWhite|ABold, bg)
	LabelNew(win, ScrHW, ScrHH, msg).SetColor(CWhite, bg)

	win.Listen("key", func(args ...interface{}) {
		win.Close()
	})

	win.Loop()
}
