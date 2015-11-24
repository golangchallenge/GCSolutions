package main

import (
	"bytes"
	"log"
	"os"
	"testing"
)

func TestNoEventHandlerOnAttemptingCoord(t *testing.T) {
	ev := &NoEventHandler{}
	board := newSampleTestBoard(9)
	coord := &Coord{2, 2}
	testNoEventHandler(t, func() { ev.OnAttemptingCoord(board, coord) }, coord)
}
func TestNoEventHandlerOnBeforeClearCoord(t *testing.T) {
	ev := &NoEventHandler{}
	board := newSampleTestBoard(9)
	coord := &Coord{2, 2}
	testNoEventHandler(t, func() { ev.OnBeforeClearCoord(board, coord) }, coord)
}
func TestNoEventHandlerOnAfterClearCoord(t *testing.T) {
	ev := &NoEventHandler{}
	board := newSampleTestBoard(9)
	coord := &Coord{2, 2}
	testNoEventHandler(t, func() { ev.OnAfterClearCoord(board, coord) }, coord)
}
func TestNoEventHandlerOnSuccessfulCoord(t *testing.T) {
	ev := &NoEventHandler{}
	board := newSampleTestBoard(9)
	coord := &Coord{2, 2}
	testNoEventHandler(t, func() { ev.OnSuccessfulCoord(board, coord) }, coord)
}
func TestNoEventHandlerOnFailedCoord(t *testing.T) {
	ev := &NoEventHandler{}
	board := newSampleTestBoard(9)
	coord := &Coord{2, 2}
	testNoEventHandler(t, func() { ev.OnFailedCoord(board, coord) }, coord)
}

func testNoEventHandler(t *testing.T, eventFunc func(), coord XY) {
	buf := &bytes.Buffer{}
	log.SetOutput(buf)

	eventFunc()

	output := buf.String()
	if output != "" {
		t.Logf("Expecting no output, but found %v", output)
		t.Fail()
	}

	log.SetOutput(os.Stdout)
}
