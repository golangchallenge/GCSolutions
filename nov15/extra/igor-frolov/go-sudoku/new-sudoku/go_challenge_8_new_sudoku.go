package main

import (
	"fmt"
	"go-sudoku/sudoku"
)

func main() {
	var complexity string
	for true {
		fmt.Print("Enter sudoku complexity (easy/medium/hardcore): ")
		fmt.Scanf("%s", &complexity)
		if complexity == "easy" || complexity == "medium" ||
			complexity == "hardcore" {
			break
		}
	}
	var sudk sudoku.SudokuField
	switch complexity {
	case "easy":
		sudk = sudoku.SudokuNew(0)
	case "medium":
		for true {
			sudk = sudoku.SudokuNew(5)
			sudokuCopy := sudk.Clone()
			_, compl := sudokuCopy.Solve(0)
			if compl < 7 {
				break
			}
		}
	case "hardcore":
		sudk = sudoku.SudokuNew(10)
	}
	sudk.Print()
}
