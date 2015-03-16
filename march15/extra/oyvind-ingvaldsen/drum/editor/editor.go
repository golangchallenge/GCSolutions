// A quick (and dirty!) editor for SPLICE drum patterns.
package editor

import (
	"strconv"

	"oyvind/drum"

	"github.com/nsf/termbox-go"
)

const (
	CBlack   termbox.Attribute = termbox.ColorBlack
	CWhite   termbox.Attribute = termbox.ColorWhite
	CRed     termbox.Attribute = termbox.ColorRed
	CBlue    termbox.Attribute = termbox.ColorBlue
	CGreen   termbox.Attribute = termbox.ColorGreen
	CMagenta termbox.Attribute = termbox.ColorMagenta
	CYellow  termbox.Attribute = termbox.ColorYellow
	CDefault termbox.Attribute = termbox.ColorDefault
	ABold    termbox.Attribute = termbox.AttrBold
)

var (
	ScrW, ScrH   int
	ScrHW, ScrHH int
)

func setupPlayerWindow(p *drum.Pattern) *Window {
	var currStep int

	w := WindowNew(false)
	player, err := drum.PlayerNew(p)
	if err != nil {
		panic(err)
	}

	player.AddCallback(func(s int) {
		currStep = s
		w.Trigger("draw")
		termbox.Flush()
	})

	openFile := func() {
		player.Stop()
		path := Prompt("Open File", "Enter filename:", "fixtures/")
		if len(path) > 0 {
			p, err := drum.DecodeFile(path)
			if err != nil {
				MessageBox("File Open Failed!", err.Error(), CRed)
			} else {
				w.Close()
				w = setupPlayerWindow(p)
				w.Loop()
			}
		}
	}

	saveFile := func() {
		player.Stop()
		path := Prompt("Save File", "Enter filename:", "fixtures/")
		if len(path) > 0 {
			err := drum.EncodePattern(p, path)
			if err != nil {
				MessageBox("File Save Failed!", err.Error(), CRed)
			} else {
				MessageBox("File Save", "File saved!", CGreen)
			}
		}

	}

	changeTempo := func() {
		player.Stop()
		tempoString := Prompt("Change Tempo", "Enter new tempo:", strconv.FormatFloat(float64(p.Tempo), 'f', -1, 32))
		if len(tempoString) > 0 {
			tempo, err := strconv.ParseFloat(tempoString, 32)
			if err != nil {
				MessageBox("Invalid Tempo!", err.Error(), CRed)
			} else if tempo < 1 || tempo > 1000 {
				MessageBox("Invalid Tempo!", "Tempo must be between 1 and 1000", CRed)
			} else {
				p.Tempo = float32(tempo)
			}
		}
	}

	changeShuffle := func() {
		player.Stop()
		shuffleString := Prompt("Change Shuffle", "Enter new shuffle ratio (0.0 - 1.0):", strconv.FormatFloat(float64(player.Shuffle), 'f', -1, 32))
		if len(shuffleString) > 0 {
			shuffle, err := strconv.ParseFloat(shuffleString, 32)
			if err != nil {
				MessageBox("Invalid Shuffle Value!", err.Error(), CRed)
			} else if shuffle < 0.0 || shuffle > 1.0 {
				MessageBox("Invalid Shuffle Value!", "Must be between 0.0 and 1.0.", CRed)
			} else {
				player.Shuffle = shuffle
			}
		}

	}

	w.Listen("key", func(args ...interface{}) {
		key := args[0].(termbox.Key)
		ch := args[1].(rune)

		switch key {
		case termbox.KeyEsc:
			w.Close()
		case termbox.KeySpace:
			player.TogglePlay()
		default:
			switch ch {
			case 'c':
				p.ClearSteps()
			case 'o':
				openFile()
			case 's':
				saveFile()
			case 't':
				changeTempo()
			case 'u':
				changeShuffle()
			}
		}
	})

	LabelNew(w, ScrHW, 1, "SPLICE Editor v0.0")
	LabelNew(w, ScrHW, 2, "==================")
	LabelNew(w, ScrHW, ScrH-2, "Controls: <SPACE> to play/stop, <O> to open, <S> to save, <T> to change tempo, <U> to change shuffle, <C> to clear steps, <ESC> to exit")

	for j, t := range p.Tracks {
		y := 10 + (j * 4)
		x := ScrHW - 55

		LabelNew(w, x, y, t.Name)

		for i, _ := range t.Steps {
			x := (ScrHW - 45) + (i * 6) + int(i/4)

			if j == 0 {
				LabelNew(w, x, y-3, strconv.Itoa(i+1))
			}

			if j == len(p.Tracks)-1 {
				LedNew(w, x, y+3, i, &currStep)
			}

			StepButtonNew(w, x, y, &t.Steps[i])
		}
	}

	return w
}

// Open the provided pattern in a new editor.
func Editor(p *drum.Pattern) {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()
	termbox.SetInputMode(termbox.InputEsc | termbox.InputMouse)

	ScrW, ScrH = termbox.Size()
	ScrHW = ScrW / 2
	ScrHH = ScrH / 2

	w := setupPlayerWindow(p)
	w.Loop()
}
