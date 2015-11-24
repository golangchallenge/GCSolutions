package main

import (
	"testing"
)

var quadTestCases = []struct {
	board  [][]byte
	coords []*Coord
}{
	{newSampleTestBoard(9), Quad0x0},
	{newSampleTestBoard(9), Quad0x1},
	{newSampleTestBoard(9), Quad0x2},
	{newSampleTestBoard(9), Quad1x0},
	{newSampleTestBoard(9), Quad1x1},
	{newSampleTestBoard(9), Quad1x2},
	{newSampleTestBoard(9), Quad2x0},
	{newSampleTestBoard(9), Quad2x1},
	{newSampleTestBoard(9), Quad2x2},
}

func TestQuadCoordinates(t *testing.T) {
	for _, testcase := range quadTestCases {
		for _, testcoord := range testcase.coords {
			result := Board(testcase.board).QuadCoords(testcoord)
			if !assertSameCoords(testcase.coords, result) {
				t.Logf("Failed getting Quad Coords: Expecting %+v, Received %+v", testcase.coords, result)
				t.Fail()
			}
		}
	}
}

/*
var challengeBoard = Board([][]byte{
	{1, 0, 3, 0, 0, 6, 0, 8, 0},
	{0, 5, 0, 0, 8, 0, 1, 2, 0},
	{7, 0, 9, 1, 0, 3, 0, 5, 6},
	{0, 3, 0, 0, 6, 7, 0, 9, 0},
	{5, 0, 7, 8, 0, 0, 0, 3, 0},
	{8, 0, 1, 0, 3, 0, 5, 0, 7},
	{0, 4, 0, 0, 7, 8, 0, 1, 0},
	{6, 0, 8, 0, 0, 2, 0, 4, 0},
	{0, 1, 2, 0, 4, 5, 0, 7, 8},
})
*/

func TestBoardString(t *testing.T) {
	output := challengeBoard.String()
	if output != expectedBoardOutput {
		t.Logf("Expecting Board \n[%v]\n but got \n[%v]\n", expectedBoardOutput, output)
		t.Fail()
	}
}

var availableValueTestCases = []struct {
	board  Board
	coord  *Coord
	values []byte
}{
	{challengeBoard, &Coord{0, 0}, []byte{}},
	{challengeBoard, &Coord{0, 1}, []byte{2}},
	{challengeBoard, &Coord{0, 2}, []byte{}},
	{challengeBoard, &Coord{0, 3}, []byte{4, 2, 5, 7, 9}},
	{challengeBoard, &Coord{0, 4}, []byte{5, 2, 9}},
	{challengeBoard, &Coord{0, 5}, []byte{}},
	{challengeBoard, &Coord{0, 6}, []byte{7, 4, 9}},
	{challengeBoard, &Coord{0, 7}, []byte{}},
	{challengeBoard, &Coord{0, 8}, []byte{9, 4}},
}

func TestAvailableValues(t *testing.T) {
	for _, testcase := range availableValueTestCases {

		actualValues := testcase.board.AvailableValuesAtCoordinate(testcase.coord)
		if len(actualValues) != len(testcase.values) {
			t.Logf("Values for coordinate %+v = %+v is different than expected %+v", testcase.coord, actualValues, testcase.values)
			t.Fail()
		} else {
		outer:
			for _, av := range actualValues {
				for _, ev := range testcase.values {
					if av == ev {
						continue outer
					}
				}
				t.Logf("Did not find match for actual value %v in received set %+v", av, testcase.values)
				t.Fail()
				break outer
			}
		}
	}
}

func TestClearExists(t *testing.T) {
	board := challengeBoard.Clone()
	x, y := 0, 0

	if board[x][y] == 0 {
		t.Logf("Expecting board to not be 0")
		t.Fail()
		return
	}
	board.Clear(&Coord{x, y})
	if board[x][y] != 0 {
		t.Logf("Expecting board to be 0")
		t.Fail()
		return
	}
}
func TestClearAlreadyZero(t *testing.T) {
	board := challengeBoard.Clone()
	x, y := 0, 1

	if board[x][y] != 0 {
		t.Logf("Expecting board to be 0")
		t.Fail()
		return
	}
	board.Clear(&Coord{x, y})
	if board[x][y] != 0 {
		t.Logf("Expecting board to be 0")
		t.Fail()
		return
	}
}

var expectedBoardOutput = `1 _ 3 _ _ 6 _ 8 _
_ 5 _ _ 8 _ 1 2 _
7 _ 9 1 _ 3 _ 5 6
_ 3 _ _ 6 7 _ 9 _
5 _ 7 8 _ _ _ 3 _
8 _ 1 _ 3 _ 5 _ 7
_ 4 _ _ 7 8 _ 1 _
6 _ 8 _ _ 2 _ 4 _
_ 1 2 _ 4 5 _ 7 8
`

var valueOfTests = []struct {
	board Board
	coord *Coord
	value int
}{
	{challengeBoard, &Coord{0, 0}, 1},
	{challengeBoard, &Coord{1, 0}, 0},
	{challengeBoard, &Coord{2, 0}, 7},
	{challengeBoard, &Coord{3, 0}, 0},
	{challengeBoard, &Coord{4, 0}, 5},
	{challengeBoard, &Coord{5, 0}, 8},
	{challengeBoard, &Coord{6, 0}, 0},
	{challengeBoard, &Coord{7, 0}, 6},
	{challengeBoard, &Coord{8, 0}, 0},

	{challengeBoard, &Coord{0, 1}, 0},
	{challengeBoard, &Coord{0, 2}, 3},
	{challengeBoard, &Coord{0, 3}, 0},
	{challengeBoard, &Coord{0, 4}, 0},
	{challengeBoard, &Coord{0, 5}, 6},
	{challengeBoard, &Coord{0, 6}, 0},
	{challengeBoard, &Coord{0, 7}, 8},
	{challengeBoard, &Coord{0, 8}, 0},

	{challengeBoard, &Coord{8, 8}, 8},
}

func TestBoardValueOf(t *testing.T) {
	for i, tc := range valueOfTests {
		actual := tc.board.ValueOf(tc.coord)
		if actual.value != byte(tc.value) {
			t.Logf("[%d] Expected value %v != Actual value %v", i, tc.value, actual.value)
			t.Fail()
		}
	}
}
func TestBoardBadValueOf(t *testing.T) {
	actual := challengeBoard.ValueOf(&Coord{100, 100})
	if actual != nil {
		t.Logf("Expecting nil value, but got %+v", actual)
		t.Fail()
	}
}

var conflictTests = []struct {
	board    Board
	coord    *Coord
	value    int
	conflict bool
}{
	{challengeBoard, &Coord{1, 0}, 4, false},
	{challengeBoard, &Coord{1, 0}, 1, true},
	{challengeBoard, &Coord{1, 0}, 2, true},
	{challengeBoard, &Coord{1, 0}, 3, true},

	{challengeBoard, &Coord{3, 0}, 2, false},
	{challengeBoard, &Coord{3, 0}, 6, true},
	{challengeBoard, &Coord{3, 0}, 5, true},
	{challengeBoard, &Coord{3, 0}, 1, true},
	{challengeBoard, &Coord{3, 0}, 9, true},

	{challengeBoard, &Coord{2, 6}, 4, false},
	{challengeBoard, &Coord{2, 6}, 9, true},
	{challengeBoard, &Coord{2, 6}, 2, true},
	{challengeBoard, &Coord{2, 6}, 5, true},
	{challengeBoard, &Coord{2, 6}, 8, true},

	{challengeBoard, &Coord{6, 6}, 6, false},
	{challengeBoard, &Coord{6, 6}, 2, false},
	{challengeBoard, &Coord{6, 6}, 4, true},
	{challengeBoard, &Coord{6, 6}, 1, true},
	{challengeBoard, &Coord{6, 6}, 7, true},
	{challengeBoard, &Coord{6, 6}, 8, true},
}

func TestBoardConflict(t *testing.T) {
	for i, tc := range conflictTests {
		testboard := tc.board.Clone()
		testboard[tc.coord.x][tc.coord.y] = byte(tc.value)
		if tc.conflict != testboard.Conflict() {
			t.Logf("[%d] Expected conflict at coord %+v with value %v to be %v", i, tc.coord, tc.value, tc.conflict)
			t.Fail()
		}
	}
}

var writeSafeTests = []struct {
	board   Board
	coord   *Coord
	value   int
	written bool
}{
	{challengeBoard, &Coord{0, 0}, 1, false},
	{challengeBoard, &Coord{1, 0}, 0, true},
	{challengeBoard, &Coord{2, 0}, 7, false},
	{challengeBoard, &Coord{3, 0}, 0, true},
	{challengeBoard, &Coord{4, 0}, 5, false},
	{challengeBoard, &Coord{5, 0}, 8, false},
	{challengeBoard, &Coord{6, 0}, 0, true},
	{challengeBoard, &Coord{7, 0}, 6, false},
	{challengeBoard, &Coord{8, 0}, 0, true},

	{challengeBoard, &Coord{0, 1}, 0, true},
	{challengeBoard, &Coord{0, 2}, 3, false},
	{challengeBoard, &Coord{0, 3}, 0, true},
	{challengeBoard, &Coord{0, 4}, 0, true},
	{challengeBoard, &Coord{0, 5}, 6, false},
	{challengeBoard, &Coord{0, 6}, 0, true},
	{challengeBoard, &Coord{0, 7}, 8, false},
	{challengeBoard, &Coord{0, 8}, 0, true},

	{challengeBoard, &Coord{8, 8}, 8, false},
	{challengeBoard, &Coord{9, 9}, 8, false},
}

func TestWriteSafe(t *testing.T) {
	for i, tc := range writeSafeTests {
		testboard := tc.board.Clone()
		actualError := testboard.WriteSafe(tc.coord, byte(tc.value))
		if tc.written && actualError != nil {
			t.Logf("[%d] Expected no write error at coord %+v with value %v, but %v", i, tc.coord, tc.value, actualError)
			t.Fail()
		} else if !tc.written && actualError == nil {
			t.Logf("[%d] Expected write error at coord %+v with value %v, but got none", i, tc.coord, tc.value)
			t.Fail()
		}
	}
}

func TestCoordXY(t *testing.T) {
	xy := &Coord{100, 200}
	newxy := CoordXY(xy)

	if xy != newxy {
		t.Logf("Expected xy == CoordXY: %v != %v", xy, newxy)
		t.Fail()
	}
}
