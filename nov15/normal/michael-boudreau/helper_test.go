package main

var numberSet = []byte{1, 2, 3, 4, 5, 6, 7, 8, 9}

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

var (
	Quad0x0 []*Coord = []*Coord{&Coord{0, 0}, &Coord{0, 1}, &Coord{0, 2}, &Coord{1, 0}, &Coord{1, 1}, &Coord{1, 2}, &Coord{2, 0}, &Coord{2, 1}, &Coord{2, 2}}
	Quad0x1          = []*Coord{&Coord{0, 3}, &Coord{0, 4}, &Coord{0, 5}, &Coord{1, 3}, &Coord{1, 4}, &Coord{1, 5}, &Coord{2, 3}, &Coord{2, 4}, &Coord{2, 5}}
	Quad0x2          = []*Coord{&Coord{0, 6}, &Coord{0, 7}, &Coord{0, 8}, &Coord{1, 6}, &Coord{1, 7}, &Coord{1, 8}, &Coord{2, 6}, &Coord{2, 7}, &Coord{2, 8}}
	Quad1x0          = []*Coord{&Coord{3, 0}, &Coord{3, 1}, &Coord{3, 2}, &Coord{4, 0}, &Coord{4, 1}, &Coord{4, 2}, &Coord{5, 0}, &Coord{5, 1}, &Coord{5, 2}}
	Quad1x1          = []*Coord{&Coord{3, 3}, &Coord{3, 4}, &Coord{3, 5}, &Coord{4, 3}, &Coord{4, 4}, &Coord{4, 5}, &Coord{5, 3}, &Coord{5, 4}, &Coord{5, 5}}
	Quad1x2          = []*Coord{&Coord{3, 6}, &Coord{3, 7}, &Coord{3, 8}, &Coord{4, 6}, &Coord{4, 7}, &Coord{4, 8}, &Coord{5, 6}, &Coord{5, 7}, &Coord{5, 8}}
	Quad2x0          = []*Coord{&Coord{6, 0}, &Coord{6, 1}, &Coord{6, 2}, &Coord{7, 0}, &Coord{7, 1}, &Coord{7, 2}, &Coord{8, 0}, &Coord{8, 1}, &Coord{8, 2}}
	Quad2x1          = []*Coord{&Coord{6, 3}, &Coord{6, 4}, &Coord{6, 5}, &Coord{7, 3}, &Coord{7, 4}, &Coord{7, 5}, &Coord{8, 3}, &Coord{8, 4}, &Coord{8, 5}}
	Quad2x2          = []*Coord{&Coord{6, 6}, &Coord{6, 7}, &Coord{6, 8}, &Coord{7, 6}, &Coord{7, 7}, &Coord{7, 8}, &Coord{8, 6}, &Coord{8, 7}, &Coord{8, 8}}
)

func assertSameCoords(c1s []*Coord, c2s []*Coord) bool {
	if len(c1s) != len(c2s) {
		return false
	}
outer:
	for _, c1 := range c1s {
		for _, c2 := range c2s {
			if assertSameCoord(c1, c2) {
				continue outer
			}
		}
		// did not find match
		return false
	}

	return true
}

func assertBoardNotEqual(b1, b2 Board) bool {
	return !assertBoardEqual(b1, b2)
}
func assertBoardEqual(b1, b2 Board) bool {
	if len(b1) != len(b2) {
		return false
	}

	for x := 0; x < len(b1); x++ {
		for y := 0; y < len(b1); y++ {
			if b1[x][y] != b2[x][y] {
				return false
			}
		}
	}

	return true
}
func assertSameCoord(c1 *Coord, c2 *Coord) bool {
	return c1.x == c2.x && c1.y == c2.y
}

func newSampleTestBoard(size int) Board {
	board := make([][]byte, size)
	for x := 0; x < size; x++ {
		row := make([]byte, size)
		for y := 0; y < size; y++ {
			row[y] = byte(y * x)
		}
		board[x] = row
	}
	return board
}

func testRow(row, max int) []*Coord {
	coords := make([]*Coord, max)
	for y := 0; y < max; y++ {
		coords[y] = &Coord{x: row, y: y}
	}
	return coords
}
func testCol(col, max int) []*Coord {
	coords := make([]*Coord, max)
	for x := 0; x < max; x++ {
		coords[x] = &Coord{x: x, y: col}
	}
	return coords
}
