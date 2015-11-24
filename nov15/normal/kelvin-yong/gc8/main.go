package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"gc8/sudoku"
	"os"
)

var levelNames = []string{"Unknown", "Easy", "Medium", "Hard", "Evil"}

func main() {
	generatePtr := flag.Int("generate", 0,
		"generate a puzzle with a specified level of difficulty "+
			"(1-easy, 2-medium, 3-hard, 4-evil)")
	flag.Parse()

	if *generatePtr == 0 {
		solveAndGrade()
	} else {
		generateSudoku(*generatePtr)
	}
}

func getPuzzleString() string {
	var buffer bytes.Buffer

	stdin := bufio.NewReader(os.Stdin)
	for {
		var s string
		_, err := fmt.Fscan(stdin, &s)
		if err != nil {
			break
		}
		buffer.WriteString(s)
		if buffer.Len() >= 81 {
			break
		}
	}
	if buffer.Len() >= 81 {
		return buffer.String()[0:81]
	}
	return buffer.String()
}

func solveAndGrade() {
	info, _ := os.Stdin.Stat()
	if (info.Mode() & os.ModeCharDevice) == os.ModeCharDevice {
		fmt.Println("Please enter the sudoku puzzle:")
	}

	s, err := sudoku.NewSudoku(getPuzzleString())
	if err == nil {
		numOfSol := s.Solve(2)
		if numOfSol > 0 {
			s.PrintSolution()
			if numOfSol > 1 {
				fmt.Println("This puzzle has more than 1 solution!")
			} else {
				level := s.SolveHuman()
				fmt.Println("Level of difficulty:", levelNames[level])
			}
		} else {
			err = errors.New("Sudoku is not solvable")
		}
	}

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func generateSudoku(level int) {
	if s, err := sudoku.GeneratePuzzle(level); err == nil {
		s.Solve(1)
		fmt.Println("Puzzle:", s.Puzzle())
		s.PrintPuzzle()
		fmt.Println("\nSolution:", s.Solution())
		s.PrintSolution()
		fmt.Println("Level of difficulty:", levelNames[level])
	} else {
		fmt.Println(err)
	}
}
