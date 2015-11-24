package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sudokusolver/sudoku"
)

const (
	fileFlag     = "file"
	generateFlag = "generate"
	levelFlag    = "level"

	fileFlagUsage     = "text file with sudoku puzzle"
	generateFlagUsage = "genetate puzzle"
	levelFlagUsage    = "set generated puzzle difficulty level (easy, medium, hard)"
)

var (
	filename = flag.String(fileFlag, "", fileFlagUsage)
	generate = flag.Bool(generateFlag, false, generateFlagUsage)
	level    = flag.String(levelFlag, "medium", levelFlagUsage)
)

func main() {
	flag.Parse()

	if *generate {
		dl := sudoku.DLMedium
		if *level == "easy" {
			dl = sudoku.DLEasy
		} else if *level == "hard" {
			dl = sudoku.DLHard
		}
		p := sudoku.Generate(dl)
		fmt.Println("Puzzle:", "\n"+p.String())

		return
	}

	var (
		puzzleReader io.ReadCloser
		err          error
		puzzle       *sudoku.Puzzle
	)

	if *filename == "" {
		puzzleReader = ioutil.NopCloser(os.Stdin)
	} else {
		puzzleReader, err = os.Open(*filename)
		if err != nil {
			panic(err)
		}
	}
	defer puzzleReader.Close()

	puzzle, err = sudoku.NewPuzzleFromReader(puzzleReader)
	if err != nil {
		fmt.Println("Could not read puzzle:", err)
		os.Exit(1)
	}

	puzzle, err = puzzle.Solve()
	if err != nil {
		fmt.Println("Could not solve puzzle:", err)
		os.Exit(1)
	}

	fmt.Println("Found unique solution.")
	fmt.Println(puzzle.String())
	fmt.Println("Difficulty level:", puzzle.Difficulty())

}
