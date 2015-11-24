package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
)

type board struct {
	solved         [81]uint
	blits          [81]uint
	loading        bool
	changed        bool
	showSteps      bool
	countSolutions bool
	solutionCount  int
	maxSolutions   int
	difficulty     int
}

type coords struct {
	row int
	col int
	box int
	pos int
}

type inspector func(int, int) error
type containerOperator func(int, inspector) error

func getCoords(pos int) coords {
	boxRow := ((pos / 9) / 3)
	boxCol := ((pos % 9) / 3)
	box := boxRow*3 + boxCol

	return coords{row: pos / 9, col: pos % 9, box: box, pos: pos}
}

func (c coords) String() string {
	return fmt.Sprintf("%c%c", getTextRow(c.row), getTextCol(c.col))
}

func getTextRow(row int) int {
	if row == 8 {
		return 'J' // the letter "I" is skipped
	}
	return 'A' + row
}

func getTextCol(col int) int {
	return '1' + col
}

func (b *board) operateOnRow(pos int, op inspector) error {
	startRow := (pos / 9) * 9
	for r := startRow; r < startRow+9; r++ {
		if err := op(r, pos); err != nil {
			return err
		}
	}
	return nil
}

func (b *board) operateOnColumn(pos int, op inspector) error {
	for c := pos % 9; c < 81; c += 9 {
		if err := op(c, pos); err != nil {
			return err
		}
	}
	return nil
}

func (b *board) operateOnBox(pos int, op inspector) error {
	startRow := ((pos / 9) / 3) * 3
	startCol := ((pos % 9) / 3) * 3
	for r := startRow; r < startRow+3; r++ {
		for c := startCol; c < startCol+3; c++ {
			target := r*9 + c
			if err := op(target, pos); err != nil {
				return err
			}
		}
	}
	return nil
}

func (b *board) operateOnRCB(pos int, op inspector) error {
	if err := b.operateOnRow(pos, op); err != nil {
		return err
	}
	if err := b.operateOnColumn(pos, op); err != nil {
		return err
	}
	if err := b.operateOnBox(pos, op); err != nil {
		return err
	}
	return nil
}

func (b *board) operateOnCommon(pos1 int, pos2 int, op inspector) error {
	coords1 := getCoords(pos1)
	coords2 := getCoords(pos2)

	if coords1.row == coords2.row {
		if err := b.operateOnRow(pos1, op); err != nil {
			return err
		}
	}
	if coords1.col == coords2.col {
		if err := b.operateOnColumn(pos1, op); err != nil {
			return err
		}
	}
	if coords1.box == coords2.box {
		if err := b.operateOnBox(pos1, op); err != nil {
			return err
		}
	}
	return nil
}

func readBoard(rd io.Reader) ([]byte, error) {
	r := bufio.NewReader(rd)

	var line string
	var err error

	buf := bytes.NewBufferString("")

	for i := 0; i < 9; i++ {
		if line, err = r.ReadString('\n'); err != nil && err != io.EOF {
			return nil, err
		}
		line = strings.Trim(line, "\r\n")

		// small deviation from spec, accept compact 81-char format
		if i == 0 && len(line) == 81 {
			return []byte(line), nil
		}

		if len(line) != 17 {
			return nil, fmt.Errorf("line %d: expected: len=17, actual: len=%d. line: %q", i+1, len(line), line)
		}

		for j, r := range line {
			if j%2 == 0 {
				if r != '_' {
					var v int
					if v, err = strconv.Atoi(string(r)); err != nil || v == 0 {
						return nil, fmt.Errorf("line %d, pos %d: expected: digit > 0 or '_', actual: %q. line: %q", i+1, j+1, r, line)
					}
				}
				buf.WriteRune(r)
			} else {
				if r != ' ' {
					return nil, fmt.Errorf("line %d, pos %d: expected: \" \", actual: %q. line: %q", i+1, j+1, r, line)
				}
			}
		}
	}

	return buf.Bytes(), nil
}

func getBoard(fileName string) (*board, error) {
	b, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	return loadBoard(b)
}

func loadBoard(b []byte) (*board, error) {
	b = bytes.Replace(b, []byte{'\r'}, []byte{}, -1)
	b = bytes.Replace(b, []byte{'\n'}, []byte{}, -1)
	b = bytes.Replace(b, []byte{' '}, []byte{}, -1)

	if len(b) != 81 {
		return nil, fmt.Errorf("line needs to be 81 chars long. line: %q", string(b))
	}

	board := &board{loading: true}
	for i := 0; i < 81; i++ {
		board.blits[i] = 0x1FF
	}

	for i := 0; i < 81; i++ {
		// allow _ 0 . to indicate an empty cell
		if b[i] != '_' && b[i] != '0' && b[i] != '.' {
			val := uint(b[i] - 48)
			if err := board.SolvePosition(i, val); err != nil {
				return board, err
			}
		}
	}

	board.loading = false

	return board, nil
}

func (b *board) numSolved() int {
	num := 0
	for i := 0; i < 81; i++ {
		if b.solved[i] != 0 {
			num++
		}
	}
	return num
}

func (b *board) isSolved() bool {
	return b.numSolved() == 81
}

func (b *board) getVisibleCells(pos int) []int {
	var list []int
	coords := getCoords(pos)

	for i := 0; i < 81; i++ {
		if i == pos || b.solved[i] != 0 {
			continue
		}
		t := getCoords(i)
		if t.row == coords.row ||
			t.col == coords.col ||
			t.box == coords.box {
			list = append(list, i)
		}
	}

	return list
}

func (b *board) getVisibleCellsWithHint(pos int, hint uint) []int {
	var list []int
	coords := getCoords(pos)

	for i := 0; i < 81; i++ {
		if i == pos || b.solved[i] != 0 || b.blits[i]&hint != hint {
			continue
		}
		t := getCoords(i)
		if t.row == coords.row ||
			t.col == coords.col ||
			t.box == coords.box {
			list = append(list, i)
		}
	}

	return list
}
