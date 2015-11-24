package main

import (
	"fmt"
	"go-sudoku/sudoku" //in case if "go-sudoku" dir situated in $GOPATH/src
)

func main() {
	var data sudoku.SudokuField
	data.Input()
	fmt.Println("You input this sudoku matrix:")
	data.Print()
	fmt.Println()
	hasSolution, complexity := data.Solve(0)
	if hasSolution {
		switch {
		case complexity == 0:
			fmt.Println("Complexity: easy")
		case 0 < complexity && complexity < 7:
			fmt.Println("Complexity: medium")
		case complexity >= 7:
			fmt.Println("Complexity: hardcore")
		}
		data.Print()
	} else {
		fmt.Println("Has no solution")
	}
}
