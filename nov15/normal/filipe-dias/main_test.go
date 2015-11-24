package main

import (
	"strings"
	"testing"
)

func TestSolve(t *testing.T) {
	data := []struct {
		desc     string
		intput   string
		expected string
		err      error
	}{
		{
			"Already solved game",
			"974236158638591742125487936316754289742918563589362417867125394253649871491873625",
			"974236158638591742125487936316754289742918563589362417867125394253649871491873625",
			nil,
		},
		{
			"One missing number",
			"2564891733746159829817234565932748617128.6549468591327635147298127958634849362715",
			"256489173374615982981723456593274861712836549468591327635147298127958634849362715",
			nil,
		},
		{
			"Easy game",
			"3.542.81.4879.15.6.29.5637485.793.416132.8957.74.6528.2413.9.655.867.192.965124.8",
			"365427819487931526129856374852793641613248957974165283241389765538674192796512438",
			nil,
		},
		{
			"Medium game",
			"..2.3...8.....8....31.2.....6..5.27..1.....5.2.4.6..31....8.6.5.......13..531.4..",
			"672435198549178362831629547368951274917243856254867931193784625486592713725316489",
			nil,
		},
		{
			"Sparsely filled",
			"..2.3.........8.....1.2.....6....2...........2.4.....1......6.5.......13....1....",
			"452136789376498152891527346165349278739281564284765931913874625647952813528613497",
			nil,
		},
		{
			"Invalid input: same number in one cube",
			"..9.7...5..21..9..1...28....7...5..1..851.....5....3.......3..68........21.....87",
			"..9.7...5..21..9..1...28....7...5..1..851.....5....3.......3..68........21.....87",
			ErrInvalidInput,
		},
		{
			"Invalid input: same number in one column",
			"6.159.....9..1............4.7.314..6.24.....5..3....1...6.....3...9.2.4......16..",
			"6.159.....9..1............4.7.314..6.24.....5..3....1...6.....3...9.2.4......16..",
			ErrInvalidInput,
		},
		{
			"Invalid input: same number in one row",
			".4.1..35.............2.5......4.89..26.....12.5.3....7..4...16.6....7....1..8..2.",
			".4.1..35.............2.5......4.89..26.....12.5.3....7..4...16.6....7....1..8..2.",
			ErrInvalidInput,
		},
		{
			"No solution",
			".9.3....1....8..46......8..4.5.6..3...32756...6..1.9.4..1......58..2....2....7.6.",
			".9.3....1....8..46......8..4.5.6..3...32756...6..1.9.4..1......58..2....2....7.6.",
			ErrNoSolution,
		},
		{
			"No solution 2",
			"..9.287..8.6..4..5..3.....46.........2.71345.........23.....5..9..4..8.7..125.3..",
			"..9.287..8.6..4..5..3.....46.........2.71345.........23.....5..9..4..8.7..125.3..",
			ErrNoSolution,
		},
		{
			"No solution 3",
			"9..1....4.14.3.8....3....9....7.8..18....3..........3..21....7...9.4.5..5...16..3",
			"9..1....4.14.3.8....3....9....7.8..18....3..........3..21....7...9.4.5..5...16..3",
			ErrNoSolution,
		},
		{
			"Multiple solutions",
			"....9....6..4.7..8.4.812.3.7.......5..4...9..5..371..4.5..6..4.2.17.85.9.........",
			"178693452623457198945812736716984325384526917592371684857169243231748569469235871",
			nil,
		},
	}

	for i, testCase := range data {
		solver := &BackTrackingSolver{}
		if gm, err := solver.Solve(buidGame(t, testCase.intput)); err != testCase.err || !buidGame(t, testCase.expected).Equals(gm) {
			t.Errorf("Test case %v/%v failed.\nexpected.gm: %v\nexpected.err: %v\n\nactual.gm: %v\nactual.err: %v", i, testCase.desc, testCase.expected, testCase.err, gm.StringCmpr(), err)
		}
	}
}

func TestReadInGame(t *testing.T) {
	data := []struct {
		desc   string
		intput string
		err    error
	}{
		{
			"Valid input only numbers",
			"974236158638591742125487936316754289742918563589362417867125394253649871491873625",
			nil,
		},
		{
			"Formatted representation",
			`
			1 _ 3 _ _ 6 _ 8 _
			_ 5 _ _ 8 _ 1 2 _
			7 _ 9 1 _ 3 _ 5 6
			_ 3 _ _ 6 7 _ 9 _
			5 _ 7 8 _ _ _ 3 _
			8 _ 1 _ 3 _ 5 _ 7
			_ 4 _ _ 7 8 _ 1 _
			6 _ 8 _ _ 2 _ 4 _
			_ 1 2 _ 4 5 _ 7 8
			`,
			nil,
		},
		{
			"Compressed representation",
			"3.5_2.81.4.__.15.6.29.56.74.5.7.3.41.1_2.8_57._4.6528.2413.9.6_5.867.1_2.96_1_4.8",
			nil,
		},
		{
			"Too short",
			"..2.3........8....31.2.....6..5.27..1.....5.2.4.6..31....8.6.5.......13..531.4..",
			ErrInvalidInputTooShort,
		},
		{
			"Too long",
			"..2.3........8....31.2.....6..5.27..1.....5.2.4.6..31....8.6.5.......13..531.4..9.4.5",
			ErrInvalidInputTooLong,
		},
		{
			"Invalid characters",
			"1.2.3.An' here I go again on my own... NOOOOOOO, STOOOOOOOOP..",
			ErrInvalidInputCharacters,
		},
	}

	for i, testCase := range data {
		if _, err := readInGame(strings.NewReader(testCase.intput)); err != testCase.err {
			t.Errorf("Test case %v/%v failed.\nexpected.err: %v\n\nactual.err: %v", i, testCase.desc, testCase.err, err)
		}
	}
}

func buidGame(t *testing.T, str string) *SudokuGame {
	if gm, err := readInGame(strings.NewReader(str)); err != nil {
		t.Fatal("Error in test case", err)
	} else {
		return gm
	}
	return nil
}
