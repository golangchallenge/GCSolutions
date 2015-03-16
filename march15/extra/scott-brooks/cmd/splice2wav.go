package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/ScottBrooks/challenge1"
)

var length = flag.Int("length", 20, "Length in seconds of sample")

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "----------------------------------\n")
		fmt.Fprintf(os.Stderr, "splice2wav [Arguments] input.splice output.wav\n")
		fmt.Fprintf(os.Stderr, "----------------------------------\n\n")
		fmt.Fprintf(os.Stderr, "Arguments\n")

		flag.PrintDefaults()
	}

	flag.Parse()
	if flag.NArg() != 2 {
		fmt.Fprintf(os.Stderr, "Invalid command line arguments\n")
		flag.Usage()
		os.Exit(2)
	}

	decoded, err := drum.DecodeFile(flag.Arg(0))
	if err != nil {
		log.Fatalf("Error loading splice: %+v", err)
	}

	w, err := os.Create(flag.Arg(1))
	if err != nil {
		log.Fatalf("Error creating output file: %+v", err)
	}
	defer w.Close()

	err = decoded.WriteWav(w, *length)
	if err != nil {
		log.Fatalf("Error writing wav: %+v", err)
	}
	log.Printf("Converted %s to %s\n", flag.Arg(0), flag.Arg(1))

}
