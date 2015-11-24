package main

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"strings"
	"time"
)

const (
	FullStartingBoard  int = 8
	HalfStartingBoard      = 5
	EmptyStartingBoard     = 2
)

var logger = &LevelLogger{}
var random = rand.New(rand.NewSource(int64(time.Now().Nanosecond())))
var difficulty = HalfStartingBoard

func main() {

	fmt.Println(banner)

	config := ConfigFromCli()
	loglevel := setLogLevel(config)

	game := NewSudokuGame(config)

	if config.Finder == "rank" {
		game.SetCoordinateRanker(&RankedCoordFinder{})
	}

	board := game.Board()
	if board == nil {
		return
	}

	if config.ShowProgress {
		game.AddEventHandler(NewProgressEventHandler(config.ShowColors, config.ProgressTime))
	}
	if loglevel >= DebugLogLevel {
		game.AddEventHandler(NewLogEventHandler())
	}

	if config.InteractiveInput {
		if !readyToSolve(config) {
			return
		}
	}

	solution, err := game.Solve(board)
	if err != nil {
		fmt.Println("Could not solve board")
		fmt.Println(err)
	}

	if config.ShowColors {
		emptyCoords := board.EmptyCoordinates()
		fmt.Printf("\nFinished Solution\n\n%v", ColorizeBoard(solution, NewColorCoordSet(emptyCoords, InvertColor)))
	} else {
		fmt.Printf("\nFinished Solution\n\n%v", solution)
	}
}

func setLogLevel(config *Config) LogLevel {
	switch {
	case config.LogLevel == "trace":
		logger = NewTraceLevelLogger()
		return TraceLogLevel
	case config.LogLevel == "debug":
		logger = NewDebugLevelLogger()
		return DebugLogLevel
	case config.LogLevel == "off":
		logger = NewSilentLevelLogger()
		return SilentLogLevel
	case config.LogLevel == "error":
		logger = NewStandardLevelLogger()
		return ErrorLogLevel
	default:
		return SilentLogLevel
	}
}

func readyToSolve(config *Config) bool {
	if !config.AutoSolve {
		if !askUser(config.UserInput, "Ready to Solve? (y/n) ") {
			return false
		}
	}
	return true
}
func askUser(reader io.Reader, question string) bool {
	r := bufio.NewReader(reader)

	fmt.Print(question)

	line, err := r.ReadString('\n')
	if err != nil {
		return false
	}
	if strings.ContainsAny(line, "y Y") {
		return true
	}
	return false
}

/*
1 _ 3 _ _ 6 _ 8 _
_ 5 _ _ 8 _ 1 2 _
7 _ 9 1 _ 3 _ 5 6
_ 3 _ _ 6 7 _ 9 _
5 _ 7 8 _ _ _ 3 _
8 _ 1 _ 3 _ 5 _ 7
_ 4 _ _ 7 8 _ 1 _
6 _ 8 _ _ 2 _ 4 _
_ 1 2 _ 4 5 _ 7 8
*/
