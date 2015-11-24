package main

import (
	"challenge8/solver"
	"log"
	"os"
)

func main() {
	/*
		// Read input from file
		f, err := os.Open("test.in")
		if err != nil {
			log.Fatalf("fail to open file: %v", err)
		}
		defer f.Close()
	*/

	// Read input from standard input
	problem, err := solver.LoadInput(os.Stdin)
	if err != nil {
		log.Fatalf("fail to read input: %v", err)
	}
	solver.ReportSolution(solver.Solve(problem))
}
