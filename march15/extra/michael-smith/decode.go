package main

import (
	"drum"
	"fmt"
	"os"
)

// Parses and prints .splice file(s)
//
// Usage:
//   decode [file ...]
func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s [file ...]\n", os.Args[0])
		os.Exit(1)
	}

	paths := os.Args[1:]
	for _, path := range paths {
		p, err := drum.DecodeFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("%s:\n\n", path)
		fmt.Println(p)
		fmt.Println()
	}
}
