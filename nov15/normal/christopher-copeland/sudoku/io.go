package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

func (b Board) String() string {
	s := ""
	for r := range b {
		for c := range b[r] {
			if b[r][c].fixed {
				s += strconv.Itoa(int(b[r][c].value) + 1)
			} else {
				s += "_"
			}
			if c < nCols-1 {
				s += " "
			}
		}
		s += "\n"
	}
	return s
}

func BoardFromReader(r io.Reader) (Board, error) {
	b := NewBoard()
	br := bufio.NewReader(r)
	for r := range b {
		line, err := br.ReadString('\n')
		if err != nil {
			return b, err
		}
		fields := strings.Fields(line)
		if len(fields) != nCols {
			return b, fmt.Errorf("row %v does not contain %v columns: %q", r, nCols, line)
		}
		for c := range b[r] {
			if fields[c] != "_" { // leave '_' fields open
				val, err := strconv.Atoi(fields[c])
				if err != nil || val < 1 || val > nVals {
					return b, fmt.Errorf("row %v column %v is invalid: %q", r, c, fields[c])
				}
				b[r][c].Set(val - 1)
			}
		}
	}
	return b, nil
}
