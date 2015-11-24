// author: Jayant Ameta
// https://github.com/wittyameta

// Package datatypes provides the data objects required by the sudoku-solver
package datatypes

import (
	"fmt"
	"sync"
)

// Position is the position of each cell of the sudoku grid. X, Y denote row, column respectively.
// Top-left corner is positioned at {0,0}, and bottom-right at {8,8}.
type Position struct {
	X int
	Y int
}

// Value contains a pointer to integer value, and a map of possible values. The key for 'Possible' map can be from 1-9.
// Val points to 0 if the value is not finalized yet, else it points to the exact value.
type Value struct {
	Val      *int
	Possible map[int]bool
}

// InitValue creates a Value object where the 'Possible' map contains all 1-9 values. Val points to 0.
func InitValue() *Value {
	possible := make(map[int]bool)
	for i := 1; i < 10; i++ {
		possible[i] = true
	}
	val := 0
	v := Value{&val, possible}
	return &v
}

// SetValue creates a Value object from int param val, where the 'Possible' map contains all only val as the key.
// Val points to val.
func SetValue(val int) *Value {
	possible := make(map[int]bool)
	possible[val] = true
	v := Value{&val, possible}
	return &v
}

// CopyValue creates a Value object from Value param value. All the fields are copied to the new Value created.
func CopyValue(value Value) *Value {
	possible := make(map[int]bool)
	for key := range value.Possible {
		possible[key] = true
	}
	val := *value.Val
	v := Value{&val, possible}
	return &v
}

// Cell is the unit from which the sudoku grid is created.
// Cell contains a pointer to integer value which is the most recent value, a Mutex, and a map of Value.
// The IterationValues map gives the Value of this cell for each iteration count.
// Iteration count starts with 0, and is incremented (if required) with each guess for a cell.
type Cell struct {
	Val             *int
	IterationValues map[int]Value
	Mutex           sync.Mutex
}

// NewCell creates a Cell object wih Val pointing to 0, and the iteration map entry for 0th iteration as InitValue.
func NewCell(x int, y int) *Cell {
	value := *InitValue()
	iterationValues := make(map[int]Value)
	iterationValues[0] = value
	val := 0
	cell := Cell{&val, iterationValues, sync.Mutex{}}
	return &cell
}

// Grid is the representation of the sudoku grid. It is a 2 dimensional array of Cell.
type Grid [9][9]Cell

// InitGrid crates a Grid object where each cell is created as NewCell.
func InitGrid() *Grid {
	grid := Grid{}
	for i := 0; i < 9; i++ {
		for j := 0; j < 9; j++ {
			grid[i][j] = *NewCell(i, j)
		}
	}
	return &grid
}

// Print prints the sudoku grid as a matrix.
func (grid *Grid) Print() {
	fmt.Println()
	for i := 0; i < 9; i++ {
		for j := 0; j < 9; j++ {
			fmt.Printf("%d ", *grid[i][j].Val)
		}
		fmt.Println()
	}
	fmt.Println()
}
