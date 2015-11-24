package puzzle

import (
	"bytes"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
)

// A sudoku board should take up 162 bytes with the given input from the challenge (18*9 bytes)
const expectedInputSize = 162

// Board represents a 9x9 Sudoku board consisting of 81 cells
type Board struct {
	cells [81]int
}

// New returs a new *Board from the parsed input bytes
func New(inBytes []byte) (*Board, error) {
	if len(inBytes) != expectedInputSize {
		log.Fatalf("Expected input size of %d bytes, got %d bytes\n", expectedInputSize, len(inBytes))
	}

	board := &Board{}

	for ix, b := range inBytes {
		// This pos should contain either a space or a "\n"
		if ix%2 != 0 {
			if b != 10 && b != 32 {
				return nil, fmt.Errorf("Unexpected byte character %q (%d)", string(b), b)
			}

			continue
		}

		// Check for unexpected bytes (should be 1-9 or _)
		if (b < 49 || b > 57) && b != 95 {
			return nil, fmt.Errorf("Unexpected byte character %q (%d)", string(b), b)
		}

		row := ix / 18
		col := (ix / 2) % 9

		// We got a "_"
		if b == 95 {
			board.cells[row*9+col] = 0
			continue
		}

		i, err := strconv.Atoi(string(b))
		if err != nil {
			return nil, err
		}

		if !board.Allowed(row, col, i) {
			return nil, fmt.Errorf("Constraint violated at row %d col %d with value %d", row, col, i)
		}

		board.cells[row*9+col] = i
	}

	return board, nil
}

// ValueAt returns the value at b[row][col]
// returns 0 for invalid indices
func (b *Board) ValueAt(row, col int) int {
	cell := row*9 + col
	if cell < 0 || cell > 80 {
		return 0
	}

	return b.cells[cell]
}

// Allowed checks if the value can be inserted at b[row][col]
// given the constraints of Sudoku
func (b *Board) Allowed(row, col, val int) bool {
	if row < 0 || row > 8 {
		return false
	}

	if col < 0 || col > 8 {
		return false
	}

	if val < 1 || col > 9 {
		return false
	}

	return !contains(val, b.row(row)) &&
		!contains(val, b.col(col)) &&
		!contains(val, b.zone(3*(row/3)+col/3))
}

// Solved returns true if the board is solved
func (b *Board) Solved() bool {
	// Perform naive check first
	for _, v := range b.cells {
		if v == 0 {
			return false
		}
	}

	return b.validate()
}

// SetAndCopy creates a copy of the board with the new value inserted.
func (b *Board) SetAndCopy(row, col, val int) *Board {
	newBoard := &Board{b.cells}
	newBoard.SetValue(row, col, val)
	return newBoard
}

// SetValue sets the cell at cells[row][col] to val
func (b *Board) SetValue(row, col, val int) {
	if b.Allowed(row, col, val) {
		b.cells[row*9+col] = val
	}
}

// String prints the current Board state to stdout
func (b *Board) String() string {
	var buffer bytes.Buffer

	for i, b := range b.cells {
		if (i+1)%9 == 0 {
			buffer.WriteString(fmt.Sprintf("%d\n", b))
			continue
		}

		buffer.WriteString(fmt.Sprintf("%d ", b))
	}

	return strings.TrimSpace(buffer.String())
}

// Difficulty returns the estimated diffuculty of the current board
func (b *Board) Difficulty() string {
	var filled int
	for _, v := range b.cells {
		if v != 0 {
			filled++
		}
	}

	if filled >= 32 {
		return "easy"
	} else if filled >= 28 {
		return "medium"
	}

	return "hard"
}

// col returns a copy of the i-th column (0-8)
func (b *Board) col(c int) []int {
	col := make([]int, 9)
	for j := 0; j < 9; j++ {
		col[j] = b.cells[c+j*9]
	}

	return col
}

// row returns a copy of the i-th row (0-8)
func (b *Board) row(r int) []int {
	row := make([]int, 9)
	copy(row, b.cells[r*9:r*9+9])
	return row
}

// zone returns a copy of the i-th zone (0-8)
func (b *Board) zone(z int) []int {
	zone := make([]int, 9)

	// startingIndex is the upper left cell within the zone z
	startingIndex := 27*(z/3) + 3*(z-3*(z/3))

	copy(zone[0:3], b.cells[startingIndex:startingIndex+3])
	copy(zone[3:6], b.cells[startingIndex+9:startingIndex+12])
	copy(zone[6:9], b.cells[startingIndex+18:startingIndex+21])

	return zone
}

// validate checks that a board with all cells set to {1-9}
// fulfills the constraints of a solved board
func (b *Board) validate() bool {
	for i := 0; i < 9; i++ {
		row := b.row(i)
		col := b.col(i)
		zone := b.zone(i)

		sort.Ints(row)
		sort.Ints(col)
		sort.Ints(zone)

		// Check duplicates
		for i := 0; i < 8; i++ {
			if row[i] == row[i+1] {
				return false
			}
			if col[i] == col[i+1] {
				return false
			}
			if zone[i] == zone[i+1] {
				return false
			}
		}
	}

	return true
}

// helper functions
func contains(needle int, haystack []int) bool {
	for _, i := range haystack {
		if i == needle {
			return true
		}
	}

	return false
}
