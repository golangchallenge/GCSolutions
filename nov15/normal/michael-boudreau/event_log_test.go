package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestLogEventHandlerOnAttemptingCoord(t *testing.T) {
	ev := NewLogEventHandler()
	board := newSampleTestBoard(9)
	coord := &Coord{2, 2}
	testLogEventHandler(t, func() { ev.OnAttemptingCoord(board, coord) }, coord)
}
func TestLogEventHandlerOnBeforeClearCoord(t *testing.T) {
	ev := NewLogEventHandler()
	board := newSampleTestBoard(9)
	coord := &Coord{2, 2}
	testLogEventHandler(t, func() { ev.OnBeforeClearCoord(board, coord) }, coord)
}
func TestLogEventHandlerOnAfterClearCoord(t *testing.T) {
	ev := NewLogEventHandler()
	board := newSampleTestBoard(9)
	coord := &Coord{2, 2}
	testLogEventHandler(t, func() { ev.OnAfterClearCoord(board, coord) }, coord)
}
func TestLogEventHandlerOnSuccessfulCoord(t *testing.T) {
	ev := NewLogEventHandler()
	board := newSampleTestBoard(9)
	coord := &Coord{2, 2}
	testLogEventHandler(t, func() { ev.OnSuccessfulCoord(board, coord) }, coord)
}
func TestLogEventHandlerOnFailedCoord(t *testing.T) {
	ev := NewLogEventHandler()
	board := newSampleTestBoard(9)
	coord := &Coord{2, 2}
	testLogEventHandler(t, func() { ev.OnFailedCoord(board, coord) }, coord)
}

func testLogEventHandler(t *testing.T, eventFunc func(), coord XY) {
	buf := &bytes.Buffer{}
	log.SetOutput(buf)

	eventFunc()

	output := buf.String()
	expecting := fmt.Sprintf("%+v", coord)
	if !strings.Contains(output, strconv.Itoa(coord.X())) || !strings.Contains(output, strconv.Itoa(coord.Y())) {
		t.Logf("Expecting coordinate %+v to be present in string %v", expecting, output)
		t.Fail()
	}

	log.SetOutput(os.Stdout)
}
