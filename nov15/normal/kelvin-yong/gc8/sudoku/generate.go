package sudoku

import (
	"errors"
	"gc8/dlx"
	"math/rand"
	"time"
)

// generate a random 9x9 grid satisfying all the sudoku constraints.
func generateFullGrid() board {
	// u is an unsolved puzzled with all 0s
	var u board

	// fill up the first row
	for i := 0; i < 9; i++ {
		u[i] = byte(i + 1)
	}

	// shuffle the first row
	for i := 0; i < 9; i++ {
		j := rand.Intn(i + 1)
		u[i], u[j] = u[j], u[i]
	}

	// solve the above using DLX in random mode to get random grid
	var solved board
	d := dlx.NewDLX(324)
	addRows(d, &u)
	d.Solve(1, true)
	sol := d.Solutions[0]
	for _, rID := range sol {
		r := rID / 81
		c := (rID - r*81) / 9
		n := rID - r*81 - c*9
		solved[r*9+c] = byte(n + 1)
	}

	return solved
}

// GeneratePuzzle generates a Sudoku puzzle with a difficulty level.
// An error is returned if the difficulty level is invalid.
func GeneratePuzzle(level int) (*Sudoku, error) {
	if level < LevelAny || level > LevelEvil {
		return nil, errors.New("Difficulty level is invalid")
	}

	if level == LevelAny {
		return randomPuzzle(), nil
	}

	var worker = func(ch chan *Sudoku, quit chan struct{}) {
		for {
			s := randomPuzzle()
			if level == s.SolveHuman() {
				select {
				case ch <- s:
				default:
				}
				return
			}
			select {
			case <-quit:
				return
			default:
			}
		}
	}

	// number of workers.
	var n int

	if level == LevelEasy {
		n = 1
	} else {
		n = 4
	}

	ch, quit := make(chan *Sudoku), make(chan struct{})
	for i := 0; i < n; i++ {
		go worker(ch, quit)
	}
	s := <-ch
	close(quit)

	return s, nil
}

// gets a random sudoku puzzle
func randomPuzzle() *Sudoku {
	solved := generateFullGrid()

	// Pick a random permutation of the cells in the grid.
	toRemove := make([]int, 81)
	for i := 0; i < 81; i++ {
		toRemove[i] = i
	}
	for i := range toRemove {
		j := rand.Intn(i + 1)
		toRemove[i], toRemove[j] = toRemove[j], toRemove[i]
	}

	// Main idea to generate a puzzle with unique solution
	// For each cell in the permuted order: remove the number in that cell;
	// check to see if this has left the puzzle with multiple solutions;
	// if so, put the cell back.
	// "Binary search" method is faster by 3 times
	low, high := 0, 81
	var unsolved board
	for {
		mid := (low + high) / 2

		unsolved = solved

		if low == high {
			for _, cell := range toRemove[:mid-1] {
				unsolved[cell] = 0
			}
			break
		} else {
			for _, cell := range toRemove[:mid] {
				unsolved[cell] = 0
			}
		}

		d := dlx.NewDLX(324)
		addRows(d, &unsolved)
		d.Solve(2, false)

		if len(d.Solutions) != 1 {
			high = mid
		} else {
			low = mid + 1
		}
	}

	return &Sudoku{puzzle: unsolved}
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Straight forward method - slower than binary search method
/*
	unsolved1 := solved
	for i, cell := range toRemove {
		num := unsolved1[cell]
		unsolved1[cell] = 0
		if i < 9 {
			// can safely remove any 9 numbers
			continue
		}
		d := dlx.NewDLX(324)
		addRows(d, &unsolved1)
		d.Solve(2, false)
		if len(d.Solutions) == 2 {
			unsolved1[cell] = num
			break
		}
	}
*/
