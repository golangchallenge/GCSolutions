package main

import (
	"testing"
)

var mainTestcases = []struct {
	board  Board
	coord  *Coord
	expect byte
}{
	{
		board:  challengeBoard,
		coord:  &Coord{0, 0},
		expect: byte(0),
	},
	{
		board:  challengeBoard,
		coord:  &Coord{0, 1},
		expect: byte(2),
	},
}

func TestMainFlow(t *testing.T) {
}
