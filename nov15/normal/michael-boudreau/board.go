package main

import (
	"fmt"
	"math"
)

type XY interface {
	X() int
	Y() int
}
type Coord struct {
	x, y int
}

func CoordXY(xy XY) *Coord {
	c, _ := xy.(*Coord)
	return c
}

type Value struct {
	*Coord
	value byte
}

func (c *Coord) X() int {
	return c.x
}
func (c *Coord) Y() int {
	return c.y
}

type Board [][]byte

func (b Board) String() string {
	var s string
	rows := len(b)
	cols := len(b[0])

	for x := 0; x < rows; x++ {
		for y := 0; y < cols; y++ {
			if y != 0 {
				s = fmt.Sprintf("%v ", s)
			}
			if b[x][y] == 0 {
				s = fmt.Sprintf("%v_", s)
			} else {
				s = fmt.Sprintf("%v%v", s, b[x][y])
			}
		}
		s = fmt.Sprintln(s)
	}
	return s
}

func (board Board) AvailableValuesAtCoordinate(coord XY) []byte {
	availableValues := make(map[byte]bool, 9)
	for i := 1; i <= 9; i++ {
		availableValues[byte(i)] = true
	}

	if board.ValueOf(coord).value != 0 {
		return []byte{}
	}
	// get available coords for quad
	quadValues := getQuadValues(board, coord)
	// get used values for quad and remove from availableValues
	for _, c := range quadValues {
		if c.value != 0 {
			availableValues[c.value] = false
		}
	}

	// get available coords for row
	rowValues := getRowValues(board, coord)
	// get used values for row and remove from availableValues
	for _, c := range rowValues {
		if c.value != 0 {
			availableValues[c.value] = false
		}
	}

	// get available coords for col
	colValues := getColValues(board, coord)
	// get used values for col and remove from availableValues
	for _, c := range colValues {
		if c.value != 0 {
			availableValues[c.value] = false
		}
	}

	var values []byte
	// range through available values
	for value, available := range availableValues {
		if !available {
			continue
		}

		values = append(values, value)
	}

	logger.Traceln("Available Values:", values)
	return values
}

func (board Board) WriteSafe(xy XY, value byte) error {
	if len(board) <= xy.X() || len(board) <= xy.Y() {
		return fmt.Errorf("Coordinate %+v is outside the size of the board", xy)
	}
	if board[xy.X()][xy.Y()] != 0 {
		return fmt.Errorf("Cannot overwrite existing value %v with new value %v at coordinate", board[xy.X()][xy.Y()], value, xy)
	}
	board[xy.X()][xy.Y()] = value
	return nil
}
func (board Board) Clear(xy XY) {
	board[xy.X()][xy.Y()] = 0
}
func (board Board) ValueOf(xy XY) *Value {
	if xy.X() >= len(board) || xy.Y() >= len(board) {
		return nil
	}
	return &Value{&Coord{x: xy.X(), y: xy.Y()}, board[xy.X()][xy.Y()]}
}

func (board Board) Conflict() bool {
	return board.conflictInRows() || board.conflictInCols() || board.conflictInQuad()
}
func (board Board) Clone() Board {
	newgrid := make([][]byte, len(board))
	for x := 0; x < len(board); x++ {
		newrow := make([]byte, len(board))
		oldrow := board[x]
		for y := 0; y < len(board); y++ {
			newrow[y] = oldrow[y]
		}
		newgrid[x] = newrow
	}
	return Board(newgrid)
}

func (board Board) conflictInRows() bool {
	for x := 0; x < len(board); x++ {
		row := make(map[byte]bool)
		for y := 0; y < len(board); y++ {
			val := board[x][y]
			if val == 0 {
				continue // blank cell, ignoring
			}

			if row[val] {
				return true // conflict found
			}
			row[val] = true
		}
	}
	return false // no conflict
}
func (board Board) conflictInCols() bool {
	for y := 0; y < len(board); y++ {
		col := make(map[byte]bool)
		for x := 0; x < len(board); x++ {
			val := board[x][y]
			if val == 0 {
				continue // blank cell, ignoring
			}

			if col[val] {
				return true // conflict found
			}
			col[val] = true
		}
	}
	return false // no conflict
}
func (board Board) conflictInQuad() bool {
	for x := 0; x < 3; x++ {
		for y := 0; y < 3; y++ {
			qx, qy := x*3, y*3
			quad := board.QuadCoords(&Coord{qx, qy})
			vals := make(map[byte]bool)
			for _, c := range quad {
				val := board[c.x][c.y]
				if val == 0 {
					continue // blank cell, ignoring
				}
				if vals[val] {
					return true // conflict found
				}
				vals[val] = true
			}
		}
	}

	return false // no conflict
}

func (board Board) QuadCoords(current XY) []*Coord {
	boardsize := len(board)
	quadsize := int(math.Sqrt(float64(boardsize)))

	xMin := (current.X() / quadsize) * quadsize
	xMax := xMin + quadsize
	yMin := (current.Y() / quadsize) * quadsize
	yMax := yMin + quadsize

	coords := make([]*Coord, boardsize)
	i := 0
	for x := xMin; x < xMax; x++ {
		for y := yMin; y < yMax; y++ {
			coords[i] = &Coord{x: x, y: y}
			i++
		}
	}

	return coords
}

func getQuadValues(board Board, current XY) []*Value {
	coords := board.QuadCoords(current)
	values := make([]*Value, len(coords))

	for i := 0; i < len(coords); i++ {
		c := coords[i]
		values[i] = board.ValueOf(c)
	}
	logger.Traceln("Values:", values)
	return values
}

func getRowValues(board Board, current XY) []*Value {
	boardsize := len(board)
	values := make([]*Value, boardsize)

	for y := 0; y < boardsize; y++ {
		values[y] = board.ValueOf(&Coord{x: current.X(), y: y})
	}
	return values
}

func getColValues(board Board, current XY) []*Value {
	boardsize := len(board)
	values := make([]*Value, boardsize)

	for x := 0; x < boardsize; x++ {
		values[x] = board.ValueOf(&Coord{x: x, y: current.Y()})
	}
	return values
}

func (board Board) EmptyCoordinates() []*Coord {
	var coords []*Coord
	for x := 0; x < len(board); x++ {
		for y := 0; y < len(board); y++ {
			if board[x][y] == 0 {
				coords = append(coords, &Coord{x, y})
			}
		}
	}
	return coords
}

/*
1 _ 3 _ _ 6 _ 8 _
_ 5 _ _ 8 _ 1 2 _
7 _ 9 1 _ 3 _ 5 6
_ 3 _ _ 6 7 _ 9 _
5 _ 7 8 _ _ _ 3 _
8 _ 1 _ 3 _ 5 _ 7
_ 4 _ _ 7 8 _ 1 _
6 _ 8 _ _ 2 _ 4 _
_ 1 2 _ 4 5 _ 7 8
*/
