package main;

import (
	"fmt"
	"os"
	"time"
	sodoku "sodoku"
	solutions "solutions"
)

func main() {
	
	start := time.Now()
	inputEntries := os.Args[1]

	solutionizer := &solutions.Solutionizer{}
	board := sodoku.GetPreDefinedBoard(inputEntries, 9)

	solutionizer.GetSodokuSolution(board)
	fmt.Println(board.GetStringFormat())
	fmt.Printf("Succesfully completed board = %v\n", board.IsBoardComplete())
	fmt.Printf("%v difficulty\n", solutionizer.Difficulty())

    elapsed := time.Since(start)
    fmt.Println("Solutionizer took %s", elapsed)
}
