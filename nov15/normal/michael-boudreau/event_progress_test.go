package main

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"testing"
)

func TestProgressEventHandlerOnAttemptingCoord(t *testing.T) {
	ev := NewProgressEventHandler(true, 0)
	board := challengeBoard
	coord := &Coord{3, 3}
	testProgressEventHandler(t, func() { ev.OnAttemptingCoord(board, coord) }, ev, coord, true)
}
func TestProgressEventHandlerOnBeforeClearCoord(t *testing.T) {
	ev := NewProgressEventHandler(true, 0)
	board := challengeBoard
	coord := &Coord{3, 3}
	testProgressEventHandler(t, func() { ev.OnBeforeClearCoord(board, coord) }, ev, coord, false)
}
func TestProgressEventHandlerOnAfterClearCoord(t *testing.T) {
	ev := NewProgressEventHandler(true, 0)
	board := challengeBoard
	coord := &Coord{3, 3}
	testProgressEventHandler(t, func() { ev.OnAfterClearCoord(board, coord) }, ev, coord, false)
}
func TestProgressEventHandlerOnSuccessfulCoord(t *testing.T) {
	ev := NewProgressEventHandler(true, 0)
	board := challengeBoard
	coord := &Coord{3, 3}
	testProgressEventHandler(t, func() { ev.OnSuccessfulCoord(board, coord) }, ev, coord, true)
}
func TestProgressEventHandlerOnFailedCoord(t *testing.T) {
	ev := NewProgressEventHandler(true, 0)
	board := challengeBoard
	coord := &Coord{3, 3}
	testProgressEventHandler(t, func() { ev.OnFailedCoord(board, coord) }, ev, coord, true)
}

func testProgressEventHandler(t *testing.T, eventFunc func(), ev *ProgressEventHandler, coord XY, hasOutput bool) {
	buf := &bytes.Buffer{}
	ev.Writer = buf

	eventFunc()

	output := buf.String()
	expecting := fmt.Sprintf("%+v", coord)
	if !hasOutput {
		if output != "" {
			t.Logf("Expecing no output, but got back %v", output)
		}
	} else {
		if !strings.Contains(output, strconv.Itoa(coord.X())) || !strings.Contains(output, strconv.Itoa(coord.Y())) {
			t.Logf("Expecting coordinate %+v to be present in string %v", expecting, output)
			t.Fail()
		}
	}
}
