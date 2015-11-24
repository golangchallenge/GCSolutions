package main

import (
	"errors"
	"fmt"
	"unicode"
)

const SIZE = 9

type Sudoku [SIZE][SIZE]int

func NewSudoku(in string) (*Sudoku, error) {
	var s Sudoku
	for i, c := range in {
		if c == '.' || c == '_' {
			s[i/SIZE][i%SIZE] = 0
		} else if unicode.IsDigit(c) {
			s[i/SIZE][i%SIZE] = int(c) - 48
		} else {
			return nil, errors.New("Invalid Sudoku")
		}
	}
	return &s, nil
}

func (s *Sudoku) Solve() error {
	var finished = false
	backtrack(s, &finished, 0, 0)

	for i := 0; i < SIZE; i++ {
		for j := 0; j < SIZE; j++ {
			if s[i][j] == 0 {
				return errors.New("Solution does not exist")
			}

		}
	}

	return nil
}

func backtrack(p *Sudoku, finished *bool, x int, y int) {
	var c [9]int
	for p[x][y] != 0 {
		y++
		if y == 9 {
			x++
			y = 0
		}
		if x >= 9 {
			break
		}
	}

	if x >= 9 {
		// process solution
		*finished = true
	} else {
		// found missing value
		construct_candidates(&c, x, y, p)
		for i := 0; i < 9; i++ {
			if c[i] == 0 {
				p[x][y] = i + 1
				backtrack(p, finished, x, y)
			}
			if *finished {
				break
			}
		}
		if *finished == false {
			p[x][y] = 0
		}
		//fmt.Println("finished")
	}
}

func construct_candidates(c *[9]int, m int, n int, p *Sudoku) {
	// get to nearest corner
	x := (m / 3) * 3
	y := (n / 3) * 3

	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if val := p[x+i][y+j]; val != 0 {
				c[val-1] = 1
			}
		}
	}

	for i := 0; i < SIZE; i++ {
		if val := p[m][i]; val != 0 {
			c[val-1] = 1
		}

		if val := p[i][n]; val != 0 {
			c[val-1] = 1
		}
	}
}

// Print prints the sudoku in 9x9 format
func (s *Sudoku) Print() {
	for i := 0; i < 9; i++ {
		for j := 0; j < 9; j++ {
			fmt.Printf("%d ", s[i][j])
		}
		fmt.Println()
	}
	fmt.Println()
}

// Validate checks if Sudoku is solved correctly
func (s *Sudoku) Validate() bool {
	var row, col, section [9]bool
	for i := 0; i < SIZE; i++ {
		// check ith row
		for j := 0; j < SIZE; j++ {
			digit := s[i][j]
			if digit < 1 || digit > 9 {
				return false
			}
			if row[digit-1] {
				return false
			}
			row[digit-1] = true

			digit = s[j][i]
			if digit < 1 || digit > 9 {
				return false
			}
			if col[digit-1] {
				return false
			}
			col[digit-1] = true
		}

		for i := range row {
			col[i] = false
			row[i] = false
		}
	}
	// check section
	for i := 0; i < SIZE; i += 3 {
		for j := 0; j < SIZE; j += 3 {
			for m := 0; m < 3; m++ {
				for n := 0; n < 3; n++ {
					digit := s[i+m][j+n]
					if digit < 1 || digit > 9 {
						return false
					}
					if section[digit-1] {
						return false
					}
					section[digit-1] = true
				}
			}
			for i := range row {
				section[i] = false
			}

		}
	}
	return true
}
