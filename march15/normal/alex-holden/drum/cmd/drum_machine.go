// The author disclaims copyright to this source code.  In place of
// a legal notice, here is a blessing:
//
//    May you do good and not evil.
//    May you find forgiveness for yourself and forgive others.
//    May you share freely, never taking more than you give.

// Package main is a CLI termbox client that provides an interface to it's
// parent 'drum' package. The directories containing the .splice files and audio
// samples needs to be specified via command line args '-sequences' and '-samples'
// respectively. They default to 'sequences' and 'samples' so if the CLI is run
// with the directory containing these directory, nothing needs to be specified.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"regexp/syntax"
	"strconv"
	"unicode"

	"../../drum"

	tb "github.com/nsf/termbox-go"
)

// entryMode determines how to interpret key presses
type entryMode int

const (
	cmdMode entryMode = iota
	tempoMode
	backupMode
	newModePickSample
	newModeEnterName
)

// terminal represent the overall termbox gui and contains misc metadata to
// keep track of its internal state.
type terminal struct {
	cur struct {
		row int
		col int
		sec int
	}
	ofs struct {
		seq int
		trk int
	}
	dm          *drum.Machine
	beat        int
	mode        entryMode
	input       bytes.Buffer
	sampleNames []string
	newSample   string
}

var term terminal

func setCellDefaultColor(x, y int, r rune) {
	tb.SetCell(x, y, r, tb.ColorWhite, tb.ColorBlack)
}

func drawHLines(ys ...int) {
	for i := 1; i < 79; i++ {
		for _, y := range ys {
			setCellDefaultColor(i, y, 0x2500)
		}
	}
}

type vLine struct{ x, y1, y2 int }

func drawVLines(lines ...vLine) {
	for _, l := range lines {
		for y := l.y1; y < l.y2; y++ {
			setCellDefaultColor(l.x, y, 0x2502)
		}
	}
}

func drawTopCorners(y int) {
	setCellDefaultColor(0, y, 0x250C)
	setCellDefaultColor(79, y, 0x2510)
}

func drawMidCorners(y int) {
	setCellDefaultColor(0, y, 0x251C)
	setCellDefaultColor(79, y, 0x2524)
}

func drawBottomCorners(y int) {
	setCellDefaultColor(0, y, 0x2514)
	setCellDefaultColor(79, y, 0x2518)
}

func tbPrintDef(x, y int, msg string) {
	tbPrint(x, y, tb.ColorWhite, tb.ColorBlack, msg)
}

func tbPrint(x, y int, fg, bg tb.Attribute, msg string) {
	for i := 0; i < len(msg); i++ {
		tb.SetCell(x+i, y, rune(msg[i]), fg, bg)
	}
}

func tbPrintf(x, y int, fg, bg tb.Attribute, format string, args ...interface{}) {
	s := fmt.Sprintf(format, args...)
	tbPrint(x, y, fg, bg, s)
}

func drawHeader() int {
	secSize := 9
	drawTopCorners(0)
	drawMidCorners(9)
	drawHLines(0, secSize)
	drawVLines(vLine{0, 1, secSize}, vLine{79, 1, secSize})

	tbPrintDef(2, 1, "hjkl: move around")
	tbPrintDef(2, 2, "w: toggle beat")
	tbPrintDef(2, 3, "q: clear track beats")
	tbPrintDef(2, 4, "e: enable/disable track")
	tbPrintDef(2, 5, "u/i: move track up/down")
	tbPrintDef(2, 6, "t: set tempo")
	tbPrintDef(2, 7, "r: decrease tempo by 2")
	tbPrintDef(2, 8, "y: increase tempo by 2")

	tbPrintDef(30, 1, "enter: load sequence")
	tbPrintDef(30, 2, "space: play/pause")
	tbPrintDef(30, 3, "n: create new sequence")
	tbPrintDef(30, 4, "s: save current sequence (overwrite file)")
	tbPrintDef(30, 5, "b: backup current sequence (create new file)")
	tbPrintDef(30, 8, "ctrl-c: quit")
	return secSize + 1
}

func drawSequences() int {
	st := term.ofs.seq
	en := st + len(term.dm.Sequences) + 1

	if term.dm.Curr != nil {
		drawMidCorners(en)
	} else {
		drawBottomCorners(en)
	}
	drawHLines(en)
	drawVLines(vLine{0, st, en}, vLine{79, st, en})

	tbPrint(2, st, tb.ColorWhite, tb.ColorBlack, "sequences")
	if term.mode == backupMode {
		msg := fmt.Sprintf(" backup name = %v", term.input.String())
		tbPrintDef(10, st, msg)
		tbPrint(10+len(msg), st, tb.ColorBlack, tb.ColorYellow, " ")
	}

	for i, seq := range term.dm.Sequences {
		if term.dm.Curr != nil && term.dm.Sequences[i] == term.dm.Curr {
			setCellDefaultColor(2, st+i+1, '*')
		}
		seqStr := fmt.Sprintf("%s (%s)", seq.Name, seq.Version)
		if term.cur.sec == 0 && term.cur.row == i {
			tbPrint(4, st+i+1, tb.ColorWhite, tb.ColorGreen, seqStr)
		} else {
			tbPrint(4, st+i+1, tb.ColorWhite, tb.ColorBlack, seqStr)
		}
	}

	return en + 1
}

func drawSamples() int {
	st := term.ofs.seq
	en := st + len(term.dm.Samples) + 1

	if term.dm.Curr != nil && len(term.dm.Curr.Sections) > 0 {
		drawMidCorners(en)
	} else {
		drawBottomCorners(en)
	}
	drawHLines(en)
	drawVLines(vLine{0, st, en}, vLine{79, st, en})

	tbPrint(2, st, tb.ColorWhite, tb.ColorBlack, "samples")
	if term.mode == newModeEnterName {
		msg := fmt.Sprintf("sequence name = %v", term.input.String())
		tbPrintDef(2, st, msg)
		tbPrint(2+len(msg), st, tb.ColorBlack, tb.ColorYellow, " ")
	}

	for i, name := range term.sampleNames {
		if term.cur.sec == 0 && term.cur.row == i {
			tbPrint(4, st+i+1, tb.ColorWhite, tb.ColorGreen, name)
		} else {
			tbPrint(4, st+i+1, tb.ColorWhite, tb.ColorBlack, name)
		}
	}

	return en + 1
}

func drawSections() {
	if term.dm.Curr == nil || len(term.dm.Curr.Sections) == 0 {
		return
	}

	st := term.ofs.trk
	en := st + len(term.dm.Curr.Sections) + 3

	drawBottomCorners(en)
	drawHLines(en)
	drawVLines(vLine{0, st, en}, vLine{79, st, en}, vLine{28, st, en})

	// pattern tempo
	if term.mode == tempoMode {
		msg := fmt.Sprintf("tracks, bpm = %v", term.input.String())
		tbPrintDef(2, st, msg)
		tbPrint(16+len(msg), st, tb.ColorBlack, tb.ColorYellow, " ")
	} else {
		msg := fmt.Sprintf("tracks, bpm = %v", term.dm.Curr.Tempo)
		tbPrintDef(2, st, msg)
	}

	// beat numbers
	for i := 0; i < 16; i++ {
		if term.beat == i {
			tbPrintf(i*3+30, st, tb.ColorBlue, tb.ColorBlack, "%2d", i+1)
		} else {
			tbPrintf(i*3+30, st, tb.ColorWhite, tb.ColorBlack, "%2d", i+1)
		}
	}

	// sequence sections
	for i, sec := range term.dm.Curr.Sections {
		tbPrintf(2, st+2+i, tb.ColorWhite, tb.ColorBlack, "%4d", sec.ID)
		if sec.Enabled {
			tbPrintf(8, st+2+i, tb.ColorGreen, tb.ColorBlack, "%18s", sec.Name)
		} else {
			tbPrintf(8, st+2+i, tb.ColorWhite, tb.ColorBlack, "%18s", sec.Name)
		}
		if term.cur.sec == 1 && term.cur.row == i {
			tbPrint(27, st+2+i, tb.ColorWhite, tb.ColorBlack, "*")
		}
		for j, beat := range sec.Beats {
			conb := term.cur.sec == 1 && term.cur.row == i && term.cur.col == j
			if beat && conb {
				tb.SetCell(j*3+31, st+2+i, 'x', tb.ColorMagenta, tb.ColorYellow)
			} else if beat && j == term.beat {
				tb.SetCell(j*3+31, st+2+i, 'x', tb.ColorBlue, tb.ColorBlack)
			} else if beat {
				setCellDefaultColor(j*3+31, st+2+i, 'x')
			} else if conb {
				tb.SetCell(j*3+31, st+2+i, ' ', tb.ColorDefault, tb.ColorGreen)
			}
		}
	}
}

func drawDrumMachine() {
	tb.Clear(tb.ColorDefault, tb.ColorDefault)
	term.ofs.seq = drawHeader()
	if term.mode == newModePickSample || term.mode == newModeEnterName {
		term.ofs.trk = drawSamples()
	} else {
		term.ofs.trk = drawSequences()
		drawSections()
	}
	tb.Flush()
}

// handleEvent processes keyboard events an processes valid input accordingly.
func handleEvent(ev tb.Event) {
	switch term.mode {
	case cmdMode:
		if term.dm.Curr != nil {
			if ev.Ch == 'q' && term.cur.sec == 1 {
				term.dm.ClearBeats(term.cur.row)
			} else if ev.Ch == 't' {
				term.mode = tempoMode
			} else if ev.Ch == 'r' {
				term.dm.ChangeTempo(term.dm.Curr.Tempo - 2)
			} else if ev.Ch == 'y' {
				term.dm.ChangeTempo(term.dm.Curr.Tempo + 2)
			} else if ev.Ch == 'e' && term.cur.sec == 1 {
				term.dm.EnableSection(term.cur.row)
			} else if ev.Ch == 'w' && term.cur.sec == 1 {
				term.dm.ToggleBeat(term.cur.row, term.cur.col)
			} else if ev.Ch == 's' {
				term.dm.SaveCurrentSequence(term.dm.Curr.Name)
			} else if ev.Ch == 'u' {
				cr := term.cur.row
				if term.cur.sec == 1 && cr < len(term.dm.Curr.Sections)-1 {
					s := term.dm.Curr.Sections
					s[cr], s[cr+1] = s[cr+1], s[cr]
					term.cur.row++
				}
			} else if ev.Ch == 'i' {
				cr := term.cur.row
				if term.cur.sec == 1 && cr > 0 {
					s := term.dm.Curr.Sections
					s[cr], s[cr-1] = s[cr-1], s[cr]
					term.cur.row--
				}
			} else if ev.Ch == 'b' {
				term.mode = backupMode
			} else if ev.Key == tb.KeySpace {
				term.dm.TogglePlayPause()
			}
		}
		if ev.Ch == 'n' {
			term.mode = newModePickSample
			term.cur.sec = 0
			term.cur.row = 0
		} else if ev.Key == tb.KeyEnter {
			if term.cur.sec == 0 {
				term.dm.LoadSequence(term.cur.row)
			}
		} else {
			handleMovement(ev)
		}
	default:
		if ev.Ch != 0 {
			if term.mode == tempoMode {
				if unicode.IsDigit(ev.Ch) || ev.Ch == '.' {
					term.input.WriteRune(ev.Ch)
				}
			} else if term.mode == newModePickSample {
				handleMovement(ev)
			} else {
				if syntax.IsWordChar(ev.Ch) {
					term.input.WriteRune(ev.Ch)
				}
			}
		} else if ev.Key == tb.KeyEnter {
			if term.mode == tempoMode {
				f, err := strconv.ParseFloat(term.input.String(), 32)
				if err == nil {
					term.dm.ChangeTempo(float32(f))
				}
			} else if term.mode == backupMode {
				term.dm.SaveCurrentSequence(term.input.String())
			} else if term.mode == newModeEnterName {
				term.dm.NewSequence(term.input.String(), term.newSample)
			}
			term.input.Reset()
			if term.mode == newModePickSample {
				term.newSample = term.sampleNames[term.cur.row]
				term.mode = newModeEnterName
			} else {
				term.mode = cmdMode
			}
		} else if ev.Key == tb.KeyBackspace || ev.Key == tb.KeyBackspace2 {
			if term.input.Len() > 0 {
				term.input.Truncate(term.input.Len() - 1)
			}
		} else if ev.Key == tb.KeyEsc {
			term.input.Reset()
			term.mode = cmdMode
		}
	}

	drawDrumMachine()
}

// handleMovement updates the cursor position when hjkl keys are pressed.
func handleMovement(ev tb.Event) {
	if term.mode == newModePickSample {
		if ev.Ch == 'j' {
			term.cur.row++
			if term.cur.row >= len(term.sampleNames) {
				term.cur.row = 0
			}
		} else if ev.Ch == 'k' {
			term.cur.row--
			if term.cur.row < 0 {
				term.cur.row = len(term.sampleNames) - 1
			}
		}
		return
	}

	if ev.Ch == 'j' { // move down
		if term.cur.sec == 0 {
			if term.cur.row < len(term.dm.Sequences)-1 {
				term.cur.row++
			} else {
				term.cur.row = 0
				if term.dm.Curr != nil && len(term.dm.Curr.Sections) > 0 {
					term.cur.sec++
				}
			}
		} else if term.cur.sec == 1 {
			if term.cur.row < len(term.dm.Curr.Sections)-1 {
				term.cur.row++
			}
		}
	} else if ev.Ch == 'k' { // move up
		if term.cur.sec == 0 {
			if term.cur.row > 0 {
				term.cur.row--
			}
		} else if term.cur.sec == 1 {
			if term.cur.row > 0 {
				term.cur.row--
			} else {
				term.cur.row = len(term.dm.Sequences) - 1
				term.cur.sec--
			}
		}
	} else if ev.Ch == 'h' && term.cur.sec == 1 { // move left
		if term.cur.col == 0 {
			term.cur.col = 15
		} else {
			term.cur.col--
		}
	} else if ev.Ch == 'l' && term.cur.sec == 1 { // move right
		if term.cur.col == 15 {
			term.cur.col = 0
		} else {
			term.cur.col++
		}
	}
}

func beatChange(beat int) {
	term.beat = beat
	drawDrumMachine()
}

var sequenceDir = flag.String("sequences", "sequences", "directory containing .splice files")
var sampleDir = flag.String("samples", "samples", "directory containing audio sample version directories")

func main() {
	flag.Parse()
	dm, err := drum.NewMachine(*sequenceDir, *sampleDir)
	if err != nil {
		fmt.Println("an error occurred loading the drum machine:", err)
		return
	}
	term.dm = dm
	defer term.dm.Close()
	term.dm.SetBeatChangeCB(beatChange)
	for name := range term.dm.Samples {
		term.sampleNames = append(term.sampleNames, name)
	}
	term.mode = cmdMode

	if err := tb.Init(); err != nil {
		fmt.Println("an error occurred initializing termbox:", err)
		return
	}
	defer tb.Close()

	drawDrumMachine()
	for { // main run loop
		ev := tb.PollEvent()
		if ev.Key == tb.KeyCtrlC {
			break
		} else if ev.Type == tb.EventResize {
			drawDrumMachine()
		} else if ev.Type == tb.EventError {
			fmt.Println("an error occurred in termbox:", err)
			return
		} else if ev.Type == tb.EventKey {
			handleEvent(ev)
		}
	}
}
