package main

type ClosestCoordFinder struct {
}

func (f *ClosestCoordFinder) NextOpenCoordinate(board Board, coord XY) (*Coord, bool) {
	// look through current quad
	quadCoords := board.QuadCoords(coord)
	for i := 0; i < len(quadCoords); i++ {
		c := quadCoords[i]
		if board[c.x][c.y] == 0 {
			return c, true
		}
	}

	// look down row
	for x := coord.X(); x < len(board); x++ {
		if board[x][coord.Y()] == 0 {
			return &Coord{x: x, y: coord.Y()}, true
		}
	}

	// look down col
	for y := coord.Y(); y < len(board); y++ {
		if board[coord.X()][y] == 0 {
			return &Coord{x: coord.X(), y: y}, true
		}
	}

	// if we get here, then go down from next entire row
	for x := coord.X() + 1; x < len(board); x++ {
		for y := 0; y < len(board); y++ {
			if board[x][y] == 0 {
				return &Coord{x: x, y: y}, true
			}
		}
	}

	// if we get here, then go up from the previous entire row
	for x := coord.X() - 1; x >= 0; x-- {
		for y := 0; y < len(board); y++ {
			if board[x][y] == 0 {
				return &Coord{x: x, y: y}, true
			}
		}
	}

	// if we get here, we are done!
	return nil, false
}
