package main

import (
	"testing"
)

var closestCoordTestCases = []struct {
	board         Board
	coord         *Coord
	expect        *Coord
	expectingMore bool
}{
	{testFinderBoardOne, &Coord{0, 0}, &Coord{0, 1}, true},
	{testFinderBoardOne, &Coord{5, 0}, &Coord{3, 0}, true},
	{testFinderBoardOne, &Coord{7, 7}, &Coord{8, 8}, true},
	{testFinderBoardOne, &Coord{5, 5}, &Coord{3, 4}, true},
	{testFinderBoardOne, &Coord{0, 5}, &Coord{0, 3}, true},
	{testFinderBoardOne, &Coord{0, 6}, &Coord{0, 6}, true},
}

func TestFindClosestCoordinate(t *testing.T) {
	finder := &ClosestCoordFinder{}

	for i, tc := range closestCoordTestCases {

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
