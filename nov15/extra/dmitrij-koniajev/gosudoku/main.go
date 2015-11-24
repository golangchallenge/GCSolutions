package main

import (
	"fmt"
	"gosudoku/sudoku"
	"os"
)

func main() {
	s, err := sudoku.ParseReader(os.Stdin)
	if err != nil {
		fmt.Println(err)
		return
	}

	solution, solved := sudoku.SolveFirst(s)
	if !solved {
		fmt.Println("Sudoku cannot be solved")
	} else {
		printSudoku(solution)
	}
}

func printSudoku(s string) {
	for r, i := 0, 0; r < 9; r, i = r+1, i+9 {
		fmt.Printf("%c %c %c %c %c %c %c %c %c\n", s[i], s[i+1], s[i+2],
			s[i+3], s[i+4], s[i+5], s[i+6], s[i+7], s[i+8])
	}
}
