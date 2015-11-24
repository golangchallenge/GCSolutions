// Package sudoku provides types and functions to parse sudoku puzzles from different
// sources and functions to find solutions to those puzzles.
package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// Constants describing the size of the sudoku.
const (
	BlockSize  = 3                       // #cells in a sub square block
	SudokuSize = BlockSize * BlockSize   // #cells in row/column/square
	TotalCells = SudokuSize * SudokuSize // #cells in sudoku
)

// Error values returned by Parse().
var (
	ErrTooManyCols = errors.New("Too many columns.")
	ErrTooManyRows = errors.New("Too many rows.")
	ErrTooFewCols  = errors.New("Too few columns.")
	ErrTooFewRows  = errors.New("Too few rows.")
	ErrWrongEntry  = errors.New("Could not parse entry.")
)

// Error values returned by Solve().
var (
	ErrInvalidSudoku    = errors.New("Sudoku has contradicting entries")
	ErrUnsolvableSudoku = errors.New("Sudoku can not be solved.")
)

// pos is a named tuple representing a position in the sudoku.
type pos struct {
	row int
	col int
}
type options map[int]bool                      // avialable options for a cell
type optLookup [SudokuSize][SudokuSize]options // lookup table that maps each cell to its available options

// lookup table that maps each cell (the target)
// to the cells that are affected if the target cell changes
var cellLookup = initCellLookup()

// Sudoku represents a sudoku puzzle as a 2-dim array.
type Sudoku [SudokuSize][SudokuSize]int

// validCell checks if the value in the target cell
// is unique in its "scope" (row/column/block)
func (s Sudoku) validCell(row, col int) bool {
	valid := true
	val := s[row][col]

	// 0 values are not a problem
	if val != 0 {
		for _, p := range cellLookup[row][col] {
			i, j := p.row, p.col
			valid = valid && (val != s[i][j])
		}
	}

	return valid
}

// IsValid checks if a sudoku has invalid cells,
// i.e. cells that have entries that are already present in
// other cells of the same "scope" (same row/column/block).
func (s Sudoku) IsValid() bool {
	valid := true

	for i, c := range s {
		for j := range c {
			valid = valid && s.validCell(i, j)
		}
	}

	return valid
}

// getCells returns the cells that are affected if the target cell given by
// row and col is changed.
func getCells(row, col int) []pos {
	sqRow := row / BlockSize
	sqCol := col / BlockSize

	cells := []pos{}

	for i := 0; i < SudokuSize; i++ {
		if i != row {
			cells = append(cells, pos{i, col})
		}

		if i != col {
			cells = append(cells, pos{row, i})
		}
	}

	for i := 0; i < BlockSize; i++ {
		for j := 0; j < BlockSize; j++ {
			nRow := BlockSize*sqRow + i
			nCol := BlockSize*sqCol + j
			if nRow != row && nCol != col {
				cells = append(cells, pos{nRow, nCol})
			}
		}
	}

	return cells
}

// meant to initialize the package variable cellLookup
func initCellLookup() [SudokuSize][SudokuSize][]pos {
	lookup := [SudokuSize][SudokuSize][]pos{}

	for i, c := range lookup {
		for j := range c {
			lookup[i][j] = getCells(i, j)
		}
	}

	return lookup
}

// calcOptions returns the options for the target cell (row, col) that are available in the current state of
// the sudoku.
func (opts optLookup) calcOptions(row, col int, s Sudoku) options {
	o := options{}
	cells := cellLookup[row][col]

	// first inclue all options
	for i := 1; i <= SudokuSize; i++ {
		o[i] = true
	}

	// remove the ones that are already taken by any relevant cell
	for _, c := range cells {
		nRow, nCol := c.row, c.col
		delete(o, s[nRow][nCol])
	}

	return o
}

// initOpts initializes internall state for the solving process (the available
// options for each cell).
func (s Sudoku) initOpts() optLookup {
	opts := optLookup{}

	for i, c := range opts {
		for j := range c {
			opts[i][j] = opts.calcOptions(i, j, s)
		}
	}

	return opts
}

// removeOpts updates options for all affected cells  of the target cell (row,
// col). It returns a slice of all the cells whose options have been changed.
// This slice is the expected input to addOpts to reverse the action.
func (opts *optLookup) removeOpts(row, col int, opt int) []pos {
	result := []pos{}
	cells := cellLookup[row][col]

	for _, c := range cells {
		nRow, nCol := c.row, c.col
		if opts[nRow][nCol][opt] {
			delete(opts[nRow][nCol], opt)
			result = append(result, c)
		}
	}

	return result
}

// addOpts readds opt as option for the cells given by the pos structs in
// cells.
func (opts *optLookup) addOpts(cells []pos, opt int) {
	for _, i := range cells {
		row, col := i.row, i.col
		opts[row][col][opt] = true
	}
}

// findNextCell returns the next empty cell with the least available options.
// If all cells are filled, it returns an error.
func (s Sudoku) findNextCell(opts optLookup) (int, int, error) {
	row, col := SudokuSize, SudokuSize
	min := SudokuSize + 1

	// find minimum of available option
	for i, c := range s {
		for j, v := range c {
			k := len(opts[i][j])
			if v == 0 && k < min {
				min = k
				row, col = i, j
			}
		}
	}

	if row == SudokuSize || col == SudokuSize {
		return row, col, errors.New("All cells are filled.")
	}

	return row, col, nil
}

// solve is the internal implementation of Solve, that solves the sudoku using recursive backtracking
// opts holds the current state of available options for each cell
func (s *Sudoku) solve(opts optLookup) (Sudoku, error) {
	row, col, e := s.findNextCell(opts)

	// if no next cell is available, the puzzle is solved
	if e != nil {
		return *s, nil
	}

	o := opts[row][col]
	// temporarily assign empty options
	// s.t. any calls to solve will not change the loop
	opts[row][col] = options{}

	for i, v := range o {
		if v {
			s[row][col] = i

			// update available options &
			// save cells whose options were changed
			changed := opts.removeOpts(row, col, i)

			sol, e := s.solve(opts)

			// solution found/
			if e == nil {
				return sol, nil
			}

			// restore previous state
			opts.addOpts(changed, i)
		}
	}

	// revert all changes made
	opts[row][col] = o
	s[row][col] = 0
	return *s, ErrUnsolvableSudoku
}

// Solve returns a solution to the sudoku if it is solvable
// else it returns an error.
func (s Sudoku) Solve() (Sudoku, error) {
	if !s.IsValid() {
		return s, ErrInvalidSudoku
	}

	opts := s.initOpts()
	return s.solve(opts)
}

// Parse parses a sudoku from a reader.
// The input should be formatted by the following rules:
// Each row of the sudoku is on a seperate line
// Each line contains exactly 9 entries seperated by exactly one space.
// No trailing spaces are allowed.
// Empty entries are represented by "_".
// A sudoku needs to have exactly 9 rows.
//
// example:
//	1 _ 3 _ _ 6 _ 8 _
//	_ 5 _ _ 8 _ 1 2 _
//	7 _ 9 1 _ 3 _ 5 6
//	_ 3 _ _ 6 7 _ 9 _
//	5 _ 7 8 _ _ _ 3 _
//	8 _ 1 _ 3 _ 5 _ 7
//	_ 4 _ _ 7 8 _ 1 _
//	6 _ 8 _ _ 2 _ 4 _
//	_ 1 2 _ 4 5 _ 7 8
//
func Parse(r io.Reader) (Sudoku, error) {
	s := Sudoku{}
	l := 0

	scanner := bufio.NewScanner(r)
	// scan input line by line
	for scanner.Scan() && l < SudokuSize {
		line := scanner.Text()
		nums := strings.Split(line, " ")

		if len(nums) > SudokuSize {
			return s, ErrTooManyCols
		} else if len(nums) < SudokuSize {
			return s, ErrTooFewCols
		}

		for i, str := range nums {
			if str != "_" {
				// try number conversion
				n, err := strconv.ParseInt(str, 10, 8)

				if err != nil {
					return s, ErrWrongEntry
				}

				if 0 < n && n <= SudokuSize {
					s[l][i] = int(n)
				} else {
					return s, ErrWrongEntry
				}
			}
		}
		l++
	}

	if l < SudokuSize {
		return s, ErrTooFewRows
	}

	if err := scanner.Err(); err == nil && scanner.Text() != "" {
		return s, ErrTooManyRows
	}

	return s, scanner.Err()
}

// String returns a formatted version of the sudoku.
// It returns the same format that is expected by Parse().
func (s Sudoku) String() string {
	lines := []string{}

	for _, row := range s {

		// convert []int to []string
		lineSlice := []string{}
		for _, val := range row {
			lineSlice = append(lineSlice, strconv.FormatInt(int64(val), 10))
		}

		line := strings.Join(lineSlice, " ")
		line = strings.Replace(line, "0", "_", -1)
		lines = append(lines, line)
	}

	lines = append(lines, "") // in order to append newline to the end of the string

	return strings.Join(lines, "\n")
}

// main reads a sudoku from stdin and prints its solution to stdout
func main() {
	s, e := Parse(os.Stdin)

	if e != nil {
		fmt.Println("There was an error while parsing your Sudoku, please check your input.")
		fmt.Println(e)
		return
	}
	sol, e := s.Solve()

	if e != nil {
		fmt.Println("The sudoku has no solution.")
		return
	}

	fmt.Println(sol)
}
