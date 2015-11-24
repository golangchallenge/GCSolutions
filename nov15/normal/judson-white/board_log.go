package main

import "fmt"

type updateLog struct {
	pos      int
	oldHints uint
	newHints uint
}

var logLastHeader string
var logLastBoardWithHints string
var logLastStepReducedHints bool
var firstLog = true

func (b *board) AddLog(technique string, log *updateLog, format string, a ...interface{}) {
	if !b.showSteps {
		return
	}

	// int = pos, uint = hint list
	var coords coords
	var args []interface{}
	args = append(args, technique)
	for _, item := range a {
		if pos, ok := item.(int); ok {
			coords = getCoords(pos)
			hints := GetBitsString(b.blits[pos])
			args = append(args, fmt.Sprintf("%s(%s)", coords, hints))
		} else if hints, ok := item.(uint); ok {
			args = append(args, GetBitsString(hints))
		} else {
			args = append(args, item)
		}
	}

	if firstLog {
		if logLastBoardWithHints != "" {
			fmt.Print(logLastBoardWithHints)
		}
		firstLog = false
	}

	header := fmt.Sprintf("%s: "+format, args...)
	if header != logLastHeader {
		if logLastBoardWithHints != "" {
			if logLastStepReducedHints {
				fmt.Println()
				fmt.Print(logLastBoardWithHints)
			}
		}
		fmt.Println(header)
		logLastHeader = header
		logLastStepReducedHints = false
	}
	if log != nil && log.newHints != 0 {
		coords = getCoords(log.pos)
		fmt.Printf("    %-7s removed from %s      old: %-17s new: %-17s\n",
			GetBitsString(log.oldHints & ^log.newHints),
			coords,
			GetBitsString(log.oldHints),
			GetBitsString(log.newHints))
		logLastStepReducedHints = true
	}
	logLastBoardWithHints = b.GetTextBoardWithHints()
}
