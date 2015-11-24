package main

import (
	"strings"
	"testing"
)

func TestColorizeBoardGreen(t *testing.T) {
	coords := []*Coord{&Coord{0, 0}}
	cc := NewColorCoordSet(coords, GreenColor)
	output := ColorizeBoard(challengeBoard, cc)

	if !strings.Contains(output, GreenColor) {
		t.Logf("Expecting green color code in colorize output")
		t.Fail()
	}
	if strings.Contains(output, RedColor) {
		t.Logf("Not expecting red color code in colorize output")
		t.Fail()
	}
	if strings.Contains(output, YellowColor) {
		t.Logf("Not expecting yellow color code in colorize output")
		t.Fail()
	}
}
func TestColorizeBoardRed(t *testing.T) {
	coords := []*Coord{&Coord{0, 0}}
	cc := NewColorCoordSet(coords, RedColor)
	output := ColorizeBoard(challengeBoard, cc)

	if !strings.Contains(output, RedColor) {
		t.Logf("Expecting red color code in colorize output")
		t.Fail()
	}
	if strings.Contains(output, GreenColor) {
		t.Logf("Not expecting green color code in colorize output")
		t.Fail()
	}
	if strings.Contains(output, YellowColor) {
		t.Logf("Not expecting yellow color code in colorize output")
		t.Fail()
	}
}
