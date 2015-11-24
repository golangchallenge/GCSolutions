package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/oijazsh/go-sudoku/sudoku"
)

func main() {
	var rank bool
	flag.BoolVar(&rank, "rank", false,
		"Rank the puzzle in addition to solving it. Also test for uniqueness. False by default.")

	flag.Parse()

	var grid sudoku.Grid
	err := grid.Write(os.Stdin)
	if err != nil {
		log.Fatal(err)
	}

	ok, message := false, ""
	if rank {
		ok, message = grid.SolveAndRank()
	} else {
		ok = grid.Solve()
	}
	if !ok {
		log.Fatal("no solution exists for given sudoku puzzle")
	}
	fmt.Print(grid.String())
	if message != "" {
		fmt.Println(message)
	}
}
