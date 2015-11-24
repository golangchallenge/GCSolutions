// Package sudoku solves and grades sudoku puzzles. Also generate new puzzles.
package sudoku

import (
	"bytes"
	"errors"
	"fmt"
	"gc8/dlx"
	"regexp"
)

// The initial puzzle and the final solution is represented as
// an array of 81 bytes
type board [81]byte

// When solving via human-like strategies (i.e no backtracking/guessing),
// we need to keep track of possible candidates for each square.
type square struct {
	num        byte
	candidates map[byte]bool
}

type humanBoard struct {
	grid        []square
	newlySolved []int
}

// Sudoku represents a 9 x 9 Sudoku Game
type Sudoku struct {
	// represents the initial board and the final board
	puzzle   board
	solution board

	// how many squares remains unsolved
	remaining int

	// only useful for human solving
	hb *humanBoard
}

// NewSudoku returns an unsolved (partially filled) Sudoku puzzle.
func NewSudoku(u string) (*Sudoku, error) {
	if len(u) != 81 {
		return nil, errors.New("Sudoku input format is invalid")
	}

	s := Sudoku{}
	s.remaining = 81

	for i := 0; i < 81; i++ {
		n := byte(u[i] - '0')
		if n >= 1 && n <= 9 {
			s.puzzle[i] = n
			s.remaining--
		}
	}

	if err := checkValid(s.puzzle); err != nil {
		return nil, err
	}

	return &s, nil
}

// Solve searches for solution(s) of the puzzle using DLX.
// Returns the number of the solutions found. Only the first solution is
// saved.
func (s *Sudoku) Solve(max int) int {
	d := dlx.NewDLX(324) // there are 324 constraints in a 9x9 puzzle
	addRows(d, &s.puzzle)

	// solve for 1 solution, pick the first column that has the least s
	// instead of any random column with the least s
	d.Solve(max, false)
	if len(d.Solutions) != 0 {
		sol := d.Solutions[0]
		for _, rID := range sol {
			r := rID / 81
			c := (rID - r*81) / 9
			n := rID - r*81 - c*9
			s.solution[r*9+c] = byte(n + 1)
		}
		s.remaining = 0
	}

	return len(d.Solutions)
}

// Puzzle returns a string with length 81 to represents the puzzle.
// Solved squares contain '1' to '9' and unsolved squares contain '.'
func (s *Sudoku) Puzzle() string {
	return boardString(s.puzzle)
}

// Solution returns a string with length 81 to represents the puzzle.
// Contains only characters '1' to '9' if the Sudoku
// puzzle was fully solved. Otherwise, it will also include '.'
func (s *Sudoku) Solution() string {
	return boardString(s.solution)
}

// PrintPuzzle writes the unsolved puzzle to the standard output.
func (s *Sudoku) PrintPuzzle() {
	printBoardString(s.Puzzle())
}

// PrintSolution writes the solution to the standard output.
func (s *Sudoku) PrintSolution() {
	printBoardString(s.Solution())
}

// ========================
// Utility methods
// ========================

// Utility to add the rows that satisfies the Sudoku constraints.
// For solving via DLX
func addRows(d *dlx.DLX, arr *board) {
	// Each row will satisifies 4 constraints

	// There are 324 constraints for a 9x9 Sudoku:
	// A.   81 position constraint: A cell at (row r, col c) has a number
	//      offset + r * 9 + c, offset = 0 x 81
	// B.   81 row constraints: Row r has number n
	//       offset + r * 9 + n, offset = 1 x 81
	// C.   81 column constraints: Col c has number n
	//      offset + c * 9 + n, offset = 2 x 81
	// D.    81 box constraint: Box b has number n
	//      offset + b * 9 + n, offset = 3 x 81
	// The row id can be encoded as (row * 9 + col) * 9 + num
	// and subsequently decoded from the dlx solution
	for r, i := 0, 0; r < 9; r++ {
		for c := 0; c < 9; c, i = c+1, i+1 {
			b := r/3*3 + c/3

			n := int((*arr)[i])
			if n >= 1 && n <= 9 {
				n-- // adjust it to 0 based
				rID := i*9 + n
				d.AddRow(rID,
					[]int{i, 81 + r*9 + n, 162 + c*9 + n, 243 + b*9 + n})
			} else {
				for n = 0; n < 9; n++ {
					rID := i*9 + n
					d.AddRow(rID,
						[]int{i, 81 + r*9 + n, 162 + c*9 + n, 243 + b*9 + n})
				}
			}
		}
	}
}

// Utility to convert a board to a string
func boardString(b board) string {
	var buffer bytes.Buffer

	for _, n := range b {
		if n > 0 && n < 10 {
			buffer.WriteByte('0' + n)
		} else {
			buffer.WriteByte('.')
		}
	}
	return buffer.String()
}

var re = regexp.MustCompile("[^1-9]")

// Utility to pretty print the board as a grid
func printBoardString(s string) {
	s = re.ReplaceAllString(s, "_")
	for r, i := 0, 0; r < 9; r, i = r+1, i+9 {
		fmt.Printf("%c %c %c | %c %c %c | %c %c %c\n",
			s[i], s[i+1], s[i+2],
			s[i+3], s[i+4], s[i+5],
			s[i+6], s[i+7], s[i+8])
		if r == 2 || r == 5 {
			fmt.Println("------+-------+------")
		}
	}
}

// checks if the board is valid and returns and error if it is not
func checkValid(puzzle board) error {
	var checkGroups = func(title string, pa peersArray) error {
		// for a specific group of peers, is there any repeated numbers
		// in Sudoku, each row, col, or box must have distinct numbers
		for r := 0; r < 9; r++ {
			m := make(map[byte]bool)
			for _, id := range pa[r] {
				n := puzzle[id]
				if n == 0 {
					continue
				}
				_, found := m[n]
				if found {
					return fmt.Errorf("%d is duplicated in %s %d", n, title, r+1)
				}
				m[n] = true
			}
		}
		return nil
	}

	var types = []struct {
		title string
		pa    peersArray
	}{{"Row", peersForRow}, {"Column", peersForCol}, {"Box", peersForCol}}

	for _, t := range types {
		if err := checkGroups(t.title, t.pa); err != nil {
			return err
		}
	}

	return nil
}

// ========================
// Peer Arrays
// ========================

// peersArray type and the associated peersForXXX answers the question:
// Given a square with an id i with an associated row r, col c and box b,
// find me ids of all 9 squares that has the same row r (or col c or box b)
type peersArray [9][]int

var (
	peersForRow peersArray
	peersForCol peersArray
	peersForBox peersArray
)

func init() {
	bStart := [9]int{0, 3, 6, 27, 30, 33, 54, 57, 60}
	bDelta := [9]int{0, 1, 2, 9, 10, 11, 18, 19, 20}

	for i := 0; i < 9; i++ {
		peersForRow[i] = make([]int, 9)
		peersForCol[i] = make([]int, 9)
		peersForBox[i] = make([]int, 9)
		for p := 0; p < 9; p++ {
			peersForRow[i][p] = i*9 + p
			peersForCol[i][p] = i + p*9
			peersForBox[i][p] = bStart[i] + bDelta[p]
		}
	}
}
