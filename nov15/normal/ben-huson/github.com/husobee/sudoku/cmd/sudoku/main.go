package main

import (
	"flag"
	"log"
	"os"

	"github.com/husobee/sudoku"
)

// command line recursion depth flag
var recursionDepth int

func init() {
	// command line recursion depth flag, default to -1, or Unlimited
	flag.IntVar(&recursionDepth, "r", -1, "set the recursion depth allowable to solve")
	flag.Parse()
}

func main() {
	// set the recursion depth allowable for our backtrack solving algorithm
	sudoku.SetRecursionDepth(recursionDepth)
	// take stdin and make a Puzzle
	p, err := sudoku.ParsePuzzle(os.Stdin)
	if err != nil {
		// bad input
		log.Fatalf("Invalid Puzzle: %s", err.Error())
	}
	// attempt to solve the puzzle
	if err := p.BacktrackSolve(); err != nil {
		// couldn't solve the puzzle
		log.Fatalf("Error solving Puzzle: %s", err.Error())
	}
	// dump out the solution to stdout
	p.Dump(os.Stdout)
}
