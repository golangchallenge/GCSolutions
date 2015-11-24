package main

import (
	"testing"
)

var rankedCoordTestCases = []struct {
	board         Board
	coord         *Coord
	expect        *Coord
	expectingMore bool
}{
	{testFinderBoardOne, &Coord{0, 0}, &Coord{8, 8}, true},
	{testFinderBoardOne, &Coord{5, 0}, &Coord{8, 8}, true},
	{testFinderBoardOne, &Coord{7, 7}, &Coord{8, 8}, true},
	{testFinderBoardOne, &Coord{5, 5}, &Coord{8, 8}, true},
	{testFinderBoardOne, &Coord{0, 5}, &Coord{8, 8}, true},
}

func TestFindRankedCoordinate(t *testing.T) {
	finder := &RankedCoordFinder{}

	for i, tc := range rankedCoordTestCases {

		nextCoord, hasMore := finder.NextOpenCoordinate(tc.board, tc.coord)
		if tc.expect.x != nextCoord.x || tc.expect.y != nextCoord.y {
			t.Logf("[%v] Coordinate mismatch. Expecting %+v, but got %+v", i, tc.expect, nextCoord)
			t.Fail()
		}
		if hasMore != tc.expectingMore {
			t.Logf("[%v] Expectation for more mismatch. Expecting %v, but got %v", i, tc.expectingMore, hasMore)
			t.Fail()
		}
	}
}
