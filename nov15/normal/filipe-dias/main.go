package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
)

// Type representing a sudoku game. A two dimensional array containing zeros for
// un-solved boxes
type SudokuGame [9][9]int

// Sudoku solver interface.
type SudokuSolver interface {
	// Tries to solve the passed in sudoku game and returns it, or err if not valid
	Solve(*SudokuGame) (*SudokuGame, error)
}

var (
	ErrNoSolution             = errors.New("Oups, this sudoku game is way too hard for me. Sorry.")
	ErrInvalidInput           = errors.New("Oh dear, here we go again. That's not a valid sudoku game.")
	ErrInvalidInputCharacters = errors.New("Oh my, input contained characters, which have long since been banned from sudoku land.")
	ErrInvalidInputTooShort   = errors.New("Well, I expected a little bit more from you.")
	ErrInvalidInputTooLong    = errors.New("Jeez, calm down. 81 is all I need.")
)

func main() {
	var err error
	var gm *SudokuGame

	if gm, err = readInGame(os.Stdin); err != nil {
		fmt.Println(err)
		return
	}

	solver := &BackTrackingSolver{}
	if gm, err = solver.Solve(gm); err == nil {
		fmt.Println(gm)
	} else {
		fmt.Println(err)
	}
}

// Implementation of the Sudoku solver using backtracking.
// Recursive implementation.
type BackTrackingSolver struct {
}

// Entry point, checks if game is valid
func (b *BackTrackingSolver) Solve(gm *SudokuGame) (*SudokuGame, error) {
	if !gm.IsValid() {
		return gm, ErrInvalidInput
	}

	if gm = b.solveRec(gm, 0, 0); !gm.Solved() {
		return gm, ErrNoSolution
	}

	return gm, nil
}

// Main part of the recursive algorithm. For each zero cell we first determine
// all the possible numbers which we could use (see findPossibleNumbers). We call
// the method recursively with each value until we found one which solved
// the game.
func (b *BackTrackingSolver) solveRec(gm *SudokuGame, row, col int) *SudokuGame {
	if col >= gm.Size() {
		col = 0
		row++
	}

	if row >= gm.Size() {
		return gm
	}

	if gm[row][col] != 0 {
		return b.solveRec(gm, row, col+1)
	}

	for _, n := range findPossibleNumbers(gm, row, col) {
		tmp := *gm
		tmp[row][col] = n
		if s := b.solveRec(&tmp, row, col); s.Solved() {
			return s
		}
	}

	return gm
}

// For the given cell find all the numbers which can be used
// without breaking the sudoku rules.
func findPossibleNumbers(gm *SudokuGame, row, col int) []int {
	usedNums := findUsedNums(gm, row, col)

	possible := []int{}
	for i := 1; i <= gm.Size(); i++ {
		if !contains(usedNums, i) {
			possible = append(possible, i)
		}
	}
	return possible
}

// Returns an array with all the numbers which are not available
// for the given cell
func findUsedNums(gm *SudokuGame, row, col int) []int {
	rs := []int{}
	for i := 0; i < gm.Size(); i++ {
		rs = appendIfNotExists(rs, gm[row][i])
		rs = appendIfNotExists(rs, gm[i][col])
	}

	rowCube := (row / 3) * 3
	colCube := (col / 3) * 3

	for i := rowCube; i < rowCube+3; i++ {
		for j := colCube; j < colCube+3; j++ {
			rs = appendIfNotExists(rs, gm[i][j])
		}
	}
	return rs
}

func appendIfNotExists(sl []int, value int) []int {
	if value == 0 || contains(sl, value) {
		return sl
	}

	return append(sl, value)
}

func contains(sl []int, val int) bool {
	for _, v := range sl {
		if v == val {
			return true
		}
	}
	return false
}

func checkIfUnique(max int, valProvider func(idx int) int) bool {
	occuredNums := []int{}
	for i := 0; i < max; i++ {
		val := valProvider(i)

		if val == 0 {
			continue
		}

		if contains(occuredNums, val) {
			return false
		}

		occuredNums = append(occuredNums, val)
	}
	return true
}

// Parse the sudoku game from the reader. Allowed characters are:
// 1) Numbers from 1 to 9
// 2) _ or . to indicate missing value
// 3) spaces, newlines or tabs are allowed but ignored
// All other character will result in ErrInvalidInputCharacters error
// Number of the valid characters 1), 2) must be exactly 81
func readInGame(reader io.Reader) (gm *SudokuGame, err error) {
	gm = &SudokuGame{}
	for r := 0; r < gm.Size(); r++ {
		for c := 0; c < gm.Size(); c++ {
			if el, err := nextEl(reader); err != nil {
				if err == io.EOF {
					return nil, ErrInvalidInputTooShort
				}
				return nil, err
			} else if el > '0' && el <= '9' {
				gm[r][c] = int(el - '0')
			} else if el != '_' && el != '.' {
				return nil, ErrInvalidInputCharacters
			}
		}
	}

	if _, err := nextEl(reader); err != io.EOF {
		return nil, ErrInvalidInputTooLong
	}

	return gm, nil
}

func nextEl(reader io.Reader) (el byte, err error) {
	for {
		if _, err = fmt.Fscanf(reader, "%c", &el); err != nil || (el != ' ' && el != '\n' && el != '\t') {
			return el, err
		}
	}
}

// only 9x9 games are supported but algorithms might be able to deal with other
// sizes..
func (s SudokuGame) Size() int {
	return 9
}

// Sudoku is considered solved if it has no more zeros. It does not check if it's valid.
func (s SudokuGame) Solved() bool {
	for _, r := range s {
		for _, c := range r {
			if c == 0 {
				return false
			}
		}
	}
	return true
}

// Checks that given sudoku game is valid
func (s SudokuGame) IsValid() bool {
	for i := 0; i < s.Size(); i++ {
		if !s.IsValidRow(i) || !s.IsValidCol(i) {
			return false
		}

		if i%3 == 0 {
			for j := 0; j < s.Size(); j += 3 {
				if !s.IsValidCube(i, j) {
					return false
				}
			}
		}
	}
	return true
}

func (s SudokuGame) IsValidRow(idx int) bool {
	return checkIfUnique(s.Size(), func(colIdx int) int {
		return s[idx][colIdx]
	})
}

func (s SudokuGame) IsValidCol(idx int) bool {
	return checkIfUnique(s.Size(), func(rowIdx int) int {
		return s[rowIdx][idx]
	})
}

func (s SudokuGame) IsValidCube(rowIdx, colIdx int) bool {
	return checkIfUnique(s.Size(), func(idx int) int {
		return s[rowIdx+(idx/3)][colIdx+(idx%3)]
	})
}

func (s SudokuGame) Equals(oth *SudokuGame) bool {
	for r, _ := range oth {
		for c, _ := range oth[r][:] {
			if oth[r][c] != s[r][c] {
				return false
			}
		}
	}
	return true
}

func (s SudokuGame) StringCmpr() string {
	return s.Printf("%c", ".", false)
}

func (s SudokuGame) String() string {
	return s.Printf(" %c", " _", true)
}

func (s SudokuGame) Printf(format, zeroStr string, newline bool) string {
	bf := bytes.NewBufferString("")
	for ir, r := range s {
		for _, c := range r {
			if c == 0 {
				fmt.Fprint(bf, zeroStr)
			} else {
				fmt.Fprintf(bf, format, '0'+c)
			}
		}
		if newline && ir < s.Size()-1 {
			fmt.Fprintln(bf)
		}
	}
	return bf.String()
}
