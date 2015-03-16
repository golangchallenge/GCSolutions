// Package drum implements the decoding of .splice drum machine files.
// See golang-challenge.com/go-challenge1/ for more information
//
// The executable of the package drum allows to display
// the content file of splice files from the command line
//
// Example:
// $ ./drum fixtures/pattern_1.splice
package main

import (
	"flag"
	"fmt"

	"github.com/simcap/drum/decoder"
)

func main() {
	filepath := flag.String("path", "", "relative file path to a splice file")
	flag.Parse()

	p, err := decoder.DecodeFile(*filepath)
	if err != nil {
		fmt.Printf("Cannot process file '%s': %v", *filepath, err)
		return
	}
	fmt.Print(p)
}
