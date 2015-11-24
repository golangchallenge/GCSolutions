package main

import (
	"fmt"
	"os"
)

func main() {
	os.Exit(mainExitCode())
}

func mainExitCode() int {
	b, err := BoardFromReader(os.Stdin)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	solved := b.Solve()
	if solved {
		fmt.Print(b)
	} else {
		fmt.Fprintln(os.Stderr, "the given board cannot be solved")
		return 1
	}
	return 0
}
