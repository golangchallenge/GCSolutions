package main

import (
	"bytes"
	"fmt"
	"io"
)

// GridSize is a size of Sudoku grid.
const GridSize = 9

// RegionSize is a size of a region (a block of nine adjacent Sudoku cells).
const RegionSize = 3

// Cell represents a sudoku cell as a set of candidate numbers for a single cell.
//
// Zero value corresponds to an unknown cell that can contain any number.
type Cell struct {
	removed byte           // number of removed candidates
	sv      byte           // a single value or 0
	mask    [GridSize]bool // true - impossible, false - possible
	// note: using uint16 and bit operations doesn't make it faster
}

// Grid represents Sudoku grid as 2-d array of cells
type Grid [GridSize][GridSize]Cell

// Group is one of row, column or region of a Grid.
type group interface {
	// valid values for i: [0..PageSize)
	cell(i int) *Cell

	coord(i int) (r int, c int)

	hasCell(r int, c int) bool
}

func (c Cell) String() string {
	var buf [GridSize]byte
	for i := byte(0); i < GridSize; i++ {
		if c.mask[i] {
			buf[i] = '_'
		} else {
			buf[i] = '1' + i
		}
	}
	return string(buf[:])
}

func (c *Cell) set(candidates []byte) {
	for i := 0; i < GridSize; i++ {
		c.mask[i] = true
	}
	for _, d := range candidates {
		c.mask[d-1] = false

	}
	// TODO: check for dups in candidates
	c.removed = byte(GridSize - len(candidates))
	if len(candidates) == 1 {
		c.sv = candidates[0]
	} else {
		c.sv = 0
	}
}

func (c *Cell) get() []byte {
	var candidates []byte
	for i := 0; i < GridSize; i++ {
		if !c.mask[i] {
			candidates = append(candidates, byte(i+1))
		}
	}
	return candidates
}

// remove removes d from the set of possible candidates
func (c *Cell) remove(d byte) {
	if !c.mask[d-1] {
		c.mask[d-1] = true
		c.removed++
	}
	if c.resolved() {
		c.sv = c.value()
	}
}

// contains checks if d is in the set of possible candidates
func (c *Cell) contains(d byte) bool {
	return !c.mask[d-1]
}

// containsOnly checks if the only candidates in a cell are d1 and d2
func (c *Cell) containsOnly(d1 byte, d2 byte) bool {
	if c.removed != GridSize-2 {
		return false
	}
	all := c.get()
	return len(all) == 2 &&
		((all[0] == d1 && all[1] == d2) || (all[0] == d2 && all[1] == d1))
}

// resolve updates a cell to contain a single possible candidate.
func (c *Cell) resolve(d byte) {
	for i := byte(0); i < GridSize; i++ {
		c.mask[i] = i != (d - 1)
	}
	c.removed = GridSize - 1
	c.sv = d
}

// resolved checks if a single possible value has been found for a cell.
func (c *Cell) resolved() bool {
	return c.removed == GridSize-1
}

// value returns a resolved number for a given cell or 0 if there are more than
// one or zero candidates.
func (c *Cell) value() byte {
	if c.sv != 0 {
		return c.sv
	}
	if !c.resolved() {
		return 0
	}
	var sv byte
	for i := byte(0); i < GridSize; i++ {
		if !c.mask[i] {
			if sv != 0 {
				return 0
			}
			sv = i + 1
		}
	}
	return sv
}

// empty checks if there are no possible candidates left for a cell.
func (c *Cell) empty() bool {
	return c.removed == GridSize
}

// NewGrid creates new empty Sudoku grid
func NewGrid() Grid {
	return Grid([GridSize][GridSize]Cell{})
}

// implements group
type row struct {
	g *Grid
	n int
}

func (r row) cell(i int) *Cell {
	return &r.g[r.n][i]
}

func (r row) coord(i int) (int, int) {
	return r.n, i
}

func (r row) hasCell(rr int, cc int) bool {
	return r.n == rr
}

// implements group
type column struct {
	g *Grid
	n int
}

func (c column) cell(i int) *Cell {
	return &c.g[i][c.n]
}

func (c column) coord(i int) (int, int) {
	return i, c.n
}

func (c column) hasCell(rr int, cc int) bool {
	return c.n == cc
}

// implements group
type region struct {
	g *Grid
	r int
	c int
}

func (rgn region) cell(i int) *Cell {
	return &rgn.g[rgn.r+i/3][rgn.c+i%3]
}

func (rgn region) coord(i int) (int, int) {
	return rgn.r + i/3, rgn.c + i%3
}

func (rgn region) hasCell(r int, c int) bool {
	return r >= rgn.r && r < rgn.r+RegionSize && c >= rgn.c && c < rgn.c+RegionSize
}

func (g *Grid) row(r int) row {
	return row{g, r}
}

func (g *Grid) column(c int) column {
	return column{g, c}
}

func (g *Grid) region(r int) region {
	i := (r / 3) * RegionSize
	j := (r % 3) * RegionSize
	return region{g, i, j}
}

// Validate return an error if g is not a valid Sudoku grid, e.g. if there are
// duplicate numbers in rows/columns/regions.
func (g Grid) Validate() error {
	// validate rows
	for r := 0; r < GridSize; r++ {
		var exist [GridSize + 1]bool
		for c := 0; c < GridSize; c++ {
			d := g[r][c].value()
			if d != 0 && exist[d] {
				return fmt.Errorf("duplicate number %d in row %d, col %d", d, r+1, c+1)
			}
			exist[d] = true
		}
	}

	// validate columns
	for c := 0; c < GridSize; c++ {
		var exist [GridSize + 1]bool
		for r := 0; r < GridSize; r++ {
			d := g[r][c].value()
			if d != 0 && exist[d] {
				return fmt.Errorf("duplicate number %d in row %d, col %d", d, r+1, c+1)
			}
			exist[d] = true
		}
	}

	// validate regions
	for i := 0; i < GridSize/RegionSize; i++ {
		for j := 0; j < GridSize/RegionSize; j++ {
			var exist [GridSize + 1]bool
			for r := i * RegionSize; r < (i+1)*RegionSize; r++ {
				for c := j * RegionSize; c < (j+1)*RegionSize; c++ {
					d := g[r][c].value()
					if d != 0 && exist[d] {
						return fmt.Errorf("duplicate number %d in at region (%d, %d), row %d, col %d",
							d, 3*i+1, 3*j+1, r+1, c+1)
					}
					exist[d] = true
				}
			}

		}
	}

	return nil
}

// IsSovled reports whether a Sudoku game is complete, i.e. there are no missing
// numbers and the Sudoku rule is satisfied.
//
// This method assumees that g.Validate() == nil
func (g Grid) IsSolved() bool {
	for r := 0; r < GridSize; r++ {
		for c := 0; c < GridSize; c++ {
			if !g[r][c].resolved() {
				return false
			}
		}
	}
	return true
}

// WriteTo writes Grid g to w.
//
// Each row is written on a new line, each number in a row is separated by
// space. A missing numbers are represented by '_' symbol. No new line is
// written after the last row.
//
// Implements io.WriterTo
func (g Grid) WriteTo(w io.Writer) (n int64, err error) {
	var buf []byte = []byte{0, 0}
	for r := 0; r < GridSize; r++ {
		for c := 0; c < GridSize; c++ {
			buf[0] = '_'
			d := g[r][c].value()
			if d != 0 {
				buf[0] = d + '0'
			}
			if c != GridSize-1 {
				buf[1] = ' '
			} else {
				buf[1] = '\n'
			}
			var k int
			if r == GridSize-1 && c == GridSize-1 {
				// do not write new line for the last row
				k, err = w.Write(buf[0:1])
			} else {
				k, err = w.Write(buf)
			}
			n += int64(k)
			if err != nil {
				return n, fmt.Errorf("failed to write sudoku grid: %s", err)
			}
		}
	}
	return n, nil
}

// String converts a Grid to a string representation described in WriteTo
// method.
func (g Grid) String() string {
	b := bytes.NewBuffer(nil)
	if _, err := g.WriteTo(b); err != nil {
		return fmt.Sprintf("Error: %s", err)
	}
	return b.String()
}

// ReadGrid reads a grid from s using the same format as WriteTo.
func ReadGrid(s io.ByteScanner) (Grid, error) {
	g := NewGrid()
	for r := 0; r < GridSize; r++ {
		for c := 0; c < GridSize; c++ {
			// read number
			b, err := s.ReadByte()
			if err != nil {
				return g, fmt.Errorf("failed to read sudoku grid row %d: %s", r+1, err)
			}

			if b == '_' {
				g[r][c] = Cell{}
			} else if b >= '1' && b <= '9' {
				g[r][c].resolve(b - '0')
			} else {
				return g, fmt.Errorf("fot a number %c at row %d", b, r+1)
			}

			if c != GridSize-1 {
				// read space
				b, err = s.ReadByte()
				if err != nil {
					return g, fmt.Errorf("failed to read sudoku grid row %d: %s", r+1, err)
				}
				if b != ' ' {
					return g, fmt.Errorf("unexpected character '%c' at row %d", b, r+1)
				}
			} else {
				// read newline
				b, err = s.ReadByte()
				if r == GridSize-1 && err == io.EOF {
					break // TODO: return EOF here?
				}
				if err != nil {
					return g, fmt.Errorf("failed to read sudoku grid row %d: %s", r+1, err)
				}
				if b != '\n' {
					// TODO: support Windows and MAC new lines
					return g, fmt.Errorf("unexpected character '%c' at row %d", b, r+1)
				}
			}
		}
	}
	return g, nil
}
