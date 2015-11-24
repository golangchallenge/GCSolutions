package parser

import (
	"bufio"
	"errors"
	"io"
	"log"
	"strconv"
	"strings"
)

//GetInput reads from an io.Reader until the end of file and returns the grid for sudoku.
func GetInput(r io.Reader) ([9][9]int, error) {
	var grid [9][9]int
	i := 0
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		inputLine := scanner.Text()
		if err := scanner.Err(); err != nil {
			log.Println("Error reading from given ioreader. Error is: ", err)
			return grid, err
		}
		inputLine = strings.TrimSpace(inputLine)
		gridLine, err := parseLine(inputLine)
		if err != nil {
			log.Println("Error parsing input line")
			return grid, err
		}
		grid[i] = gridLine
		i++
	}
	return grid, nil
}

func parseLine(line string) ([9]int, error) {
	lineItems := strings.Split(line, " ")
	var retValues [9]int
	for ind, item := range lineItems {
		item = strings.TrimSpace(item)
		if strings.Compare(item, "_") == 0 {
			retValues[ind] = 0
			continue
		}
		if len(item) < 1 {
			continue
		}
		squareVal, err := strconv.Atoi(item)
		if err != nil {
			return retValues, err
		}
		if squareVal > 9 || squareVal < 1 {
			return retValues, errors.New("Expected a digit between 1 and 9 in the input cell.")
		}
		retValues[ind] = squareVal
	}
	return retValues, nil
}
