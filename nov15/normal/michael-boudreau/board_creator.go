package main

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

type BoardDifficulty int

const (
	EasyBoard   BoardDifficulty = 1
	MediumBoard                 = 3
	HardBoard                   = 5
)

var boardRandomizer = rand.New(rand.NewSource(int64(time.Now().Nanosecond())))

func ReduceBoardToDifficulty(board Board, difficulty int) {
	starting := 35 - (difficulty * 3)                            // range of 21-32
	numberOfStartingValues := starting + boardRandomizer.Intn(5) // 21-25... 32-36

	boardsize := len(board) * len(board)
	numberToRemove := boardsize - numberOfStartingValues

	for i := 0; i < numberToRemove; i++ {
		for attempts := 0; ; attempts++ {
			randomCoord := getCoordinateForAttempt(attempts)
			if val := board.ValueOf(randomCoord).value; val != 0 {
				board.Clear(randomCoord)
				break
			}
		}
	}
}

func getCoordinateForAttempt(attempt int) *Coord {
	if attempt%2 == 0 {
		boardRandomizer.Seed(int64(time.Now().Nanosecond()))
	}
	return &Coord{boardRandomizer.Intn(9), boardRandomizer.Intn(9)}
}

func CreateInitialBoardWithStartingCountAndSize(boardSize, startingCount int) Board {
	grid := make([][]byte, boardSize)
	for i := 0; i < len(grid); i++ {
		grid[i] = make([]byte, boardSize)
	}

	board := Board(grid)
	placeValues(board, startingCount)

	return board
}

func placeValues(board Board, count int) {
	for i := 0; i < count; {
		randomCoord := &Coord{boardRandomizer.Intn(9), boardRandomizer.Intn(9)}
		if board.ValueOf(randomCoord).value != 0 {
			continue
		}
		vals := board.AvailableValuesAtCoordinate(randomCoord)
		if len(vals) == 0 {
			continue
		}
		randomIndex := boardRandomizer.Intn(len(vals))
		if err := board.WriteSafe(randomCoord, vals[randomIndex]); err == nil && !board.Conflict() {
			i++
		}
	}
}

func EstimateDifficulty(board Board) BoardDifficulty {
	empty := board.EmptyCoordinates()
	filled := (len(board) * len(board)) - len(empty)

	switch {
	case filled < 26:
		return HardBoard
	case filled < 32:
		return MediumBoard
	default:
		return EasyBoard
	}
}

func BoardFromReader(reader io.Reader) (Board, error) {
	r := bufio.NewReader(reader)

	var grid [][]byte
	var length int

	for {
		line, err := r.ReadString('\n')
		if err != nil {
			return nil, err
		}

		row, err := convertLineToRow(line)
		if err != nil {
			return nil, err
		}

		if len(grid) == 0 {
			length = len(row)
		}

		if length != len(row) {
			return nil, fmt.Errorf("Inconsistent row lengths %v != %v", length, len(row))
		}

		grid = append(grid, row)
		if len(grid) == length {
			return Board(grid), nil
		}
	}
}

func convertLineToRow(line string) ([]byte, error) {
	line = strings.TrimSuffix(line, "\n")
	parts := strings.Split(line, " ")
	result := make([]byte, len(parts))
	for i := 0; i < len(parts); i++ {
		if parts[i] == "-" || parts[i] == "_" {
			result[i] = 0
		} else {
			myInt, err := strconv.Atoi(parts[i])
			if err != nil {
				return nil, fmt.Errorf("Could not convert %v to int: %v", parts[i], err)
			}
			result[i] = byte(myInt)
		}
	}
	return result, nil
}
