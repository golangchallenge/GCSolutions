package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

func readFromTestFile(path string) (Grid, error) {
	var g Grid
	f, err := os.Open(path)
	if err != nil {
		return g, err
	}
	defer f.Close()

	r := bufio.NewReader(f)
	for {
		var line string
		if line, err = r.ReadString('\n'); err != nil {
			break
		}
		if strings.TrimSpace(line) == "Puzzle:" {
			return ReadGrid(r)
		}
	}
	if err == io.EOF {
		err = fmt.Errorf("unable to find puzzle in %s", path)
	}

	return g, err
}

func readFromStdin() (Grid, error) {
	return ReadGrid(bufio.NewReader(os.Stdin))
}

func checkError(err error, msg string) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %s\n", msg, err)
		os.Exit(1)
	}
}

func main() {
	var verbose = flag.Bool("v", false, "be verbose")
	var ui = flag.Bool("ui", false, "show solution progress using a term-based UI")
	var testFile = flag.String("t", "", "read puzzle from a test file (for testing only)")
	var help = flag.Bool("h", false, "show help")

	flag.Parse()
	if *help {
		flag.Usage()
		os.Exit(0)
	}

	var puzzle Grid
	var err error
	if *testFile != "" {
		if *verbose {
			fmt.Printf("Reading puzzle from %s\n", *testFile)
		}
		puzzle, err = readFromTestFile(*testFile)
	} else {
		if *verbose {
			fmt.Println("Reading puzzle from stdin")
		}
		puzzle, err = readFromStdin()
	}
	checkError(err, "Failed to read puzzle")

	err = puzzle.Validate()
	checkError(err, "Invalid puzzle")

	if *ui {
		err = uiLoop(puzzle)
		checkError(err, "Failed to initialize UI")
	} else {
		if *verbose {
			fmt.Printf("Solving puzzle:\n%s\n\n", puzzle)
		}
		s := newSolver(puzzle)
		s.solve()
		checkError(s.err, "Failed to solve")
		if *verbose {
			fmt.Printf("Found solution:\n%s\n\n", s.g)
			fmt.Printf("Esitmated difficulty level: %d\n", s.level())
		} else {
			fmt.Println(s.g)
		}
	}
}
