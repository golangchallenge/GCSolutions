package sudoku

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Write parses the input from an io.Reader to construct the initial
// sudoku puzzle. Returns an error in case of malformed input.
func (grid *Grid) Write(r io.Reader) error {
	scanner := bufio.NewScanner(r)
	i := 0
	for i < 9 && scanner.Scan() {
		line := scanner.Text()
		cells := strings.Fields(line)
		if len(cells) != 9 {
			return fmt.Errorf("sudoku input: row %v incorrect length", i+1)
		}
		for j, cell := range cells {
			val, err := strCellValue(cell)
			if err != nil {
				return err
			}
			(*grid)[i][j] = val
		}
		i++
	}
	if i < 9 {
		return fmt.Errorf("sudoku input: too few rows in input")
	}
	return nil
}

// String returns a string representation of the sudoku puzzle
func (g *Grid) String() string {
	var buff bytes.Buffer
	for r := 0; r < gridLen; r++ {
		for c := 0; c < gridLen; c++ {
			buff.WriteString(strconv.Itoa(g[r][c]))
			if c != gridLen-1 {
				buff.WriteString(" ")
			}
		}
		buff.WriteString("\n")
	}
	return buff.String()
}

func strCellValue(cell string) (int, error) {
	cellErr := fmt.Errorf("sudoku input: unacceptable cell value %v", cell)
	if len(cell) > 1 {
		return 0, cellErr
	}
	if cell == "_" {
		return 0, nil
	} else if cell > "0" || cell < "9" {
		val, err := strconv.Atoi(cell)
		return val, err
	} else {
		return 0, cellErr
	}
}
