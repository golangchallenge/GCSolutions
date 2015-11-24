package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/slimmy/go-challenge-8/backtrack"
	"github.com/slimmy/go-challenge-8/dlx"
	"github.com/slimmy/go-challenge-8/puzzle"
)

var (
	errUnsolvable  = errors.New("Could not solve board")
	generateFlag   string
	printDifficuly bool
)

func main() {
	flag.StringVar(
		&generateFlag,
		"generate",
		"",
		"Generate a Sudoku board, accepts inputs: 'easy', 'medium' 'hard'",
	)
	flag.BoolVar(
		&printDifficuly,
		"print-difficulty",
		false,
		"Prints the difficulty of the input board",
	)
	flag.Parse()

	// Generate a Sudoku board and print it to stdout
	if len(generateFlag) > 0 {
		for {
			// generate a random board
			board, err := puzzle.Generate(generateFlag)
			if err != nil {
				log.Fatalln(err)
			}

			// make sure it can be solved
			// the backtrack algorithm is used to solve the board since it's
			// faster (in most cases) than DLX to detect unsolvable boards
			solvedBoard := backtrack.NewSolver(board).Solve()
			if solvedBoard.Solved() {
				fmt.Println(strings.Replace(board.String(), "0", "_", -1))
				return
			}
		}
	}

	// Read Sudoku board from stdin
	b, err := ioutil.ReadAll(bufio.NewReader(os.Stdin))
	if err != nil {
		log.Fatalln(err)
	}

	// Build the board
	board, err := puzzle.New(b)
	if err != nil {
		log.Fatalln(err)
	}

	// Just print the difficulty of the board if the caller asks for it
	if printDifficuly {
		fmt.Println(board.Difficulty())
		return
	}

	// Solve the board
	dlx := dlx.NewSolver(board)
	solvedBoard := dlx.Solve()
	if !solvedBoard.Solved() {
		log.Fatalln(errUnsolvable)
	}

	fmt.Println(solvedBoard)
}
