package main

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
)

//Board contains init and result numbers
type Board struct {
	Box
	cell [9][9]int
}

//Rest set all cell to 0
func (b *Board) Rest() {
	for i := range b.cell {
		for j := range b.cell[i] {
			b.cell[i][j] = 0
		}
	}
}

// Read reads whitespace-separated ints from r. It returns an error value.
func (b *Board) Read(r io.Reader) error {
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanWords)
	i, x := 0, 0
	var err error
	for scanner.Scan() {
		t := scanner.Text()
		if t == "_" {
			x = 0
		} else {
			x, err = strconv.Atoi(t)
			if nil != err {
				return err
			}
		}
		err = b.SetCell(i%9, i/9, x)
		if nil != err {
			return err
		}
		i++
	}
	return scanner.Err()
}

//Show send board to stdout
func (b Board) Show() {
	fmt.Print(b.String())
}

//String transform board presentation to string
func (b Board) String() (s string) {
	for i := range b.cell {
		for j := range b.cell[i] {
			s += fmt.Sprintf("%d ", b.cell[i][j])
		}
		s += fmt.Sprintln()
	}
	return
}

//StringB do the same as String but added the lines for box visualization
func (b Board) StringB() (s string) {
	for i := range b.cell {
		for j := range b.cell[i] {
			s += fmt.Sprintf("%d ", b.cell[i][j])
			switch j {
			case 2, 5:
				s += fmt.Sprint("| ")
			case 8:
				s += fmt.Sprintln()
			}
		}
		if i == 2 || i == 5 {
			s += fmt.Sprintf("------+-------+-------\n")
		}
	}
	return
}

//IsInRow check is value exists in n-th row [0..8]
func (b Board) IsInRow(r, v int) bool {
	for j := range b.cell[r] {
		if b.cell[r][j] == v {
			return true
		}
	}
	return false
}

//IsInColumn check is value exists in n-th column [0..8]
func (b Board) IsInColumn(c, v int) bool {
	for i := range b.cell {
		if b.cell[i][c] == v {
			return true
		}
	}
	return false
}

//IsInBox check is value exixts in n-th box [0..8]
func (b Board) IsInBox(boxID, v int) bool {
	x, y := b.FirstXYBox(boxID)
	for i := y; i < y+3; i++ {
		for j := x; j < x+3; j++ {
			if b.cell[i][j] == v {
				return true
			}
		}
	}
	return false
}

//FreeCellsBox - use binaries mask to define is the cell empty
func (b Board) FreeCellsBox(boxID int) (mask int) {
	x, y := b.FirstXYBox(boxID)
	for i := y; i < y+3; i++ {
		for j := x; j < x+3; j++ {
			if b.cell[i][j] == 0 {
				mask |= (1 << uint((i-y)*3+(j-x)))
			}
		}
	}
	return
}

//LeftedBox return BoxIDs are not in the slice
func (b Board) LeftedBox(bSet []int) (mask int) {
	mask = BinaryMask9
	for _, b := range bSet {
		mask ^= (1 << uint(b))
	}
	return
}

//SetCell check if possible set number by coordinates and if of add it to board cell set
func (b *Board) SetCell(x, y, v int) error {
	switch true {
	case v < 0 || v > 9:
		return fmt.Errorf("The value out of range [1..9] or _ separated by whitespace: %d", v)
	case v > 0 && b.IsInRow(y, v):
		return fmt.Errorf("The value (%d) is duplicated with one in row #%d", v, y)
	case v > 0 && b.IsInColumn(x, v):
		return fmt.Errorf("The value (%d) is duplicated with one in column #%d", v, x)
	case v > 0 && b.IsInBox(b.BoxID(x, y), v):
		return fmt.Errorf("The value (%d) is duplicated with one in box #%d", v, b.BoxID(x, y))
	}
	b.cell[y][x] = v
	return nil
}
