package main

import (
	"fmt"
	"strconv"
)

type ColorCode string

const (
	InvertColor           ColorCode = "\033[7m"
	BlackColor                      = "\033[1;30m"
	GreenColor                      = "\033[1;32m"
	RedColor                        = "\033[1;31m"
	YellowColor                     = "\033[1;33m"
	BlueColor                       = "\033[1;34m"
	LightGrayColor                  = "\033[37m"
	WhiteColor                      = "\033[97m"
	BlackBGColor                    = "\033[40m"
	RedBGColor                      = "\033[41m"
	GreenBGColor                    = "\033[42m"
	YellowBGColor                   = "\033[43m"
	ResetColor                      = "\033[0m"
	BoldColor                       = "\033[1m"
	StatusSuccessColor              = "\033[1;30;42m" // bold; FG=black; BG=green
	StatusFailedColor               = "\033[1;37;41m" // bold; FG=light gray; BG=red
	StatusInProgressColor           = "\033[1;30;43m" // bold; FG=black; BG=yellow

	ClearConsole string = "\033[2J"
	ResetCursor         = "\033[H"
)

type ColorCoordSet struct {
	coords []*Coord
	codes  []ColorCode
}

func NewColorCoordSet(coords []*Coord, codes ...ColorCode) *ColorCoordSet {
	return &ColorCoordSet{coords, codes}
}

func ColorizeBoard(b Board, colors ...*ColorCoordSet) string {
	var s string
	rows := len(b)
	cols := len(b[0])

	for x := 0; x < rows; x++ {
		for y := 0; y < cols; y++ {
			val := strconv.Itoa(int(b[x][y]))
			if b[x][y] == 0 {
				val = "_"
			}

			var colorcode string
			for _, colorset := range colors {
				if inCoordSet(x, y, colorset.coords) {
					colorcode = fmt.Sprintf("%v%v", colorcode, buildColorString(colorset.codes...))
				}
			}
			if colorcode == "" {
				colorcode = ResetColor
			}

			s = fmt.Sprintf("%v%v %v %v", s, colorcode, val, ResetColor)
		}
		s = fmt.Sprintln(s)
	}
	return s
}

func inCoordSet(x int, y int, coords []*Coord) bool {
	for _, c := range coords {
		if c == nil {
			continue
		}
		if c.x == x && c.y == y {
			return true
		}
	}
	return false
}

func buildColorString(codes ...ColorCode) string {
	var s string
	for _, c := range codes {
		s = fmt.Sprintf("%v%v", s, c)
	}
	return s
}
