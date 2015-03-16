package editor

import (
	"github.com/nsf/termbox-go"
)

// Open a popup Window with primitive prompt. Returns an empty string if the user closes
// the Prompt (i.e. pressing <ESC>).
func Prompt(title, msg, placeholder string) string {
	win := WindowNew(true)

	BoxNew(win, ScrHW, ScrHH, 40, 4).SetColor(CWhite, CMagenta)
	LabelNew(win, ScrHW, ScrHH-3, title).SetColor(CWhite|ABold, CMagenta)
	LabelNew(win, ScrHW, ScrHH, msg).SetColor(CWhite, CMagenta)
	li := LabelNew(win, ScrHW, ScrHH+2, placeholder).SetColor(CBlue|ABold, CMagenta)

	win.Listen("key", func(args ...interface{}) {
		key := args[0].(termbox.Key)
		ch := args[1].(rune)

		switch key {
		case termbox.KeyEsc:
			li.Text = ""
			win.Close()
		case termbox.KeyEnter:
			win.Close()
		case termbox.KeyBackspace:
			fallthrough
		case termbox.KeyBackspace2:
			li.Text = li.Text[:len(li.Text)-1]
		default:
			li.Text += string(ch)
		}
	})

	win.Loop()
	return li.Text
}
