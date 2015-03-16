package main

import (
	"fmt"
	"os"

	"github.com/jstemmer/go-challenge1/drum"
)

func main() {
	if len(os.Args) != 2 {
		dief("Usage: %s <splice file>\n", os.Args[0])
	}

	pattern, err := drum.DecodeFile(os.Args[1])
	if err != nil {
		dief("error loading splice file: %s\n", err)
	}

	fmt.Printf("Loaded pattern:\n%s\n", pattern.String())

	// add more cowbell
	for i, track := range pattern.Tracks {
		if track.Name == "cowbell" {
			for j := 0; j < 8; j++ {
				pattern.Tracks[i].Steps[j*2] = true
			}
		}
	}

	fmt.Printf("Modified pattern:\n%s\n", pattern.String())
	if err := drum.EncodeFile(*pattern, "out.splice"); err != nil {
		dief("error encoding pattern: %s\n", err)
	}
}

func dief(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format, args...)
	os.Exit(1)
}
