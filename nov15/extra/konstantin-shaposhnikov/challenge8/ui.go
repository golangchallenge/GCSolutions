package main

import (
	"fmt"

	"github.com/nsf/termbox-go"
)

const hLine = '─'
const vLine = '│'
const topLeftCorner = '┌'
const topRightCorner = '┐'
const bottomLeftCorner = '└'
const bottomRightCorner = '┘'
const leftCross = '├'
const rightCross = '┤'
const topCross = '┬'
const bottomCross = '┴'
const centerCross = '┼'

const cellSz = RegionSize + 1

func displayMessage(msg string, fg termbox.Attribute) {
	const y = cellSz*GridSize + 3
	const maxX = 80

	for x := 0; x < maxX; x++ {
		termbox.SetCell(x, y, ' ', termbox.ColorDefault, termbox.ColorDefault)
	}

	x := 0
	for _, ch := range msg {
		if x >= maxX {
			break
		}
		termbox.SetCell(x, y, ch, fg, termbox.ColorDefault)
		x++
	}

	_ = termbox.Flush()
}

func displayGrid(g Grid) {
	const maxX = cellSz * GridSize

	for r := 0; r <= GridSize; r++ {
		y := r * cellSz
		for x := 0; x <= maxX; x++ {
			termbox.SetCell(x, y, hLine, termbox.ColorDefault, termbox.ColorDefault)
		}
	}

	for c := 0; c <= GridSize; c++ {
		x := c * cellSz
		for y := 0; y <= maxX; y++ {
			var ch = vLine
			switch {
			case x == 0 && y == 0:
				ch = topLeftCorner
			case x == 0 && y == maxX:
				ch = bottomLeftCorner
			case x == maxX && y == 0:
				ch = topRightCorner
			case x == maxX && y == maxX:
				ch = bottomRightCorner
			case y == 0:
				ch = topCross
			case y == maxX:
				ch = bottomCross
			case y%cellSz == 0 && x == 0:
				ch = leftCross
			case y%cellSz == 0 && x == maxX:
				ch = rightCross
			case y%cellSz == 0:
				ch = centerCross
			}

			termbox.SetCell(x, y, ch, termbox.ColorDefault, termbox.ColorDefault)
		}
	}

	for r := 0; r < GridSize; r++ {
		for c := 0; c < GridSize; c++ {
			x := c*cellSz + 1
			y := r*cellSz + 1
			cell := g[r][c]
			v := cell.value()
			for i := 1; i <= GridSize; i++ {
				ch := rune(i + '0')
				fg := termbox.ColorYellow
				if v != 0 {
					if i == 5 {
						ch = rune(v + '0')
					} else {
						ch = ' '
					}
					fg = termbox.ColorGreen | termbox.AttrBold
				} else {
					if !cell.contains(byte(i)) {
						ch = ' '
					}
				}
				termbox.SetCell(x+(i-1)%3, y+(i-1)/3, ch, fg, termbox.ColorDefault)
			}
		}
	}

	_ = termbox.Flush()
}

func displayDiff(from Grid, to Grid) {
	for r := 0; r < GridSize; r++ {
		for c := 0; c < GridSize; c++ {
			fCell := from[r][c]
			v := fCell.value()
			if v != 0 {
				continue
			}
			tCell := to[r][c]

			x := c*cellSz + 1
			y := r*cellSz + 1
			for i := 1; i <= GridSize; i++ {
				d := byte(i)
				ch := rune(i + '0')
				fg := termbox.ColorYellow
				bg := termbox.ColorDefault
				if !fCell.contains(d) {
					ch = ' '
				}
				if fCell.contains(d) && !tCell.contains(d) {
					fg = termbox.ColorWhite | termbox.AttrBold
					bg = termbox.ColorRed
				}
				termbox.SetCell(x+(i-1)%3, y+(i-1)/3, ch, fg, bg)
			}
		}
	}

	_ = termbox.Flush()
}

func uiLoop(g Grid) error {
	if err := termbox.Init(); err != nil {
		return err
	}
	defer termbox.Close()

	s := newSolver(g)

	for !s.solved {
		displayMessage("", termbox.ColorDefault)
		displayGrid(s.g)
		termbox.PollEvent()

		lastG := s.g
		s.step()
		if s.err != nil {
			break
		}
		displayMessage(s.msg, termbox.ColorDefault)
		displayDiff(lastG, s.g)
		termbox.PollEvent()
	}

	if s.err != nil {
		displayMessage(fmt.Sprintf("Unable to solve: %s", s.err), termbox.ColorRed)
	} else {
		displayGrid(s.g)
		displayMessage(fmt.Sprintf("Solved! Level: %d", s.level()), termbox.ColorGreen)
	}

	termbox.PollEvent()

	return nil
}
