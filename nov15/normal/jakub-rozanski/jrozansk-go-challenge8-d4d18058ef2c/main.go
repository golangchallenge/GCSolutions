package main

import (
	"bitbucket.org/jrozansk/go-challenge8/sudoku"
	"bufio"
	"fmt"
	"io"
	"os"
)

func main() {
	input := readGrid(os.Stdin)
	solver, err := sudoku.InitSolver(input)
	if err != nil || !solver.Solve() {
		fmt.Println("Couldn't solve given sudoku: ", err)
		return
	}
	sudoku.PrintReadableGrid(solver.GetSolution())
}

func readGrid(reader io.Reader) string {
	scanner := bufio.NewScanner(reader)
	scanner.Split(bufio.ScanWords)
	grid := make([]byte, sudoku.NumOfCells)
	for i := 0; i < len(grid); i++ {
		if !scanner.Scan() || len(scanner.Bytes()) == 0 {
			fmt.Println("Reading sudoku grid failed!")
			return ""
		}
		if scanner.Bytes()[0] == '_' {
			grid[i] = '0'
			continue
		}
		grid[i] = scanner.Bytes()[0]
	}
	return string(grid)
}
