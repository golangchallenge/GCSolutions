package puzzle

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

const (
	Easy = iota
	Medium
	Hard
)

// Generate generates a new Sudoku board for a given difficulty
func Generate(d string) (*Board, error) {
	var difficulty int

	switch strings.ToLower(d) {
	case "easy":
		difficulty = Easy
	case "medium":
		difficulty = Medium
	case "hard":
		difficulty = Hard
	default:
		return nil, fmt.Errorf("Invalid difficulty for generating sudoku board: %q", d)
	}

	return generate(difficulty), nil
}

func generate(difficulty int) *Board {
	rand.Seed(time.Now().UnixNano())
	b := &Board{}
	n := cellsToFill(difficulty)

	var filledCells int
	for filledCells < n {
		row := rand.Intn(9)
		col := rand.Intn(9)
		val := 1 + rand.Intn(9)
		if b.ValueAt(row, col) != 0 {
			continue
		}

		if b.Allowed(row, col, val) {
			b.SetValue(row, col, val)
			filledCells++
		}
	}

	return b
}

// Returns the number of cells to fill for the given diffuculty
func cellsToFill(difficulty int) int {
	switch difficulty {
	case Easy:
		return 32 + rand.Intn(5)
	case Medium:
		return 28 + rand.Intn(4)
	case Hard:
		return 26 + rand.Intn(2)
	}

	// We shouldnt end up here, but defaulting to an easy board
	return cellsToFill(Easy)
}
