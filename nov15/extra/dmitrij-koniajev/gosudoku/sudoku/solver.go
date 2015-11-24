package sudoku

import "gosudoku/dlx"

// SolveFirst finds first solution of Sudoku in 81 character string format.
func SolveFirst(s string) (res string, ok bool) {
	Solve(s, func(solution string) bool {
		ok = true
		res = solution
		return true
	})
	return
}

// Solve finds all solutions of Sudoku in 81 character string format.
// Each solution is passed to accept function. It stops immediately when accept function returns true.
func Solve(s string, accept func(string) bool) {
	m := encodeConstraints(s)

	m.Solve(func(cs [][]int) bool {
		return accept(decodeExactCoverSolution(cs))
	})
}

func encodeConstraints(s string) *dlx.Matrix {
	m := dlx.NewMatrix(324)

	for row, position := 0, 0; row < 9; row++ {
		for column := 0; column < 9; column, position = column+1, position+1 {
			region := row/3*3 + column/3
			digit := int(s[position] - '1') // zero based digit
			if digit >= 0 && digit < 9 {
				m.AddRow([]int{position, 81 + row*9 + digit, 162 + column*9 + digit, 243 + region*9 + digit})
			} else {
				for digit = 0; digit < 9; digit++ {
					m.AddRow([]int{position, 81 + row*9 + digit, 162 + column*9 + digit, 243 + region*9 + digit})
				}
			}
		}
	}

	return m
}

func decodeExactCoverSolution(cs [][]int) string {
	b := make([]byte, len(cs))
	for _, row := range cs {
		position := row[0]
		value := row[1] % 9
		b[position] = byte(value) + '1'
	}
	return string(b)
}
