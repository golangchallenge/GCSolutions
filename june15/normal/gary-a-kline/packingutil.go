// Packingutil.go uses constants and objects defined in the pallet.go file.
//
// The convention used for placing boxes on a pallet is as follows: the box length is
// added to the 'x' coordinate and the box width is added to the 'y' coordinate
// (see pallet.paint() function). Regarding box position, (0,0) is considered
// the 'top left' corner of the pallet, and box coordinates refer to the
// position of the box's 'top left' corner.
package main

import (
	"errors"
	"fmt"
	"sort"
)

const emptyCell = 0

type errBadCol int

func (e errBadCol) Error() string {
	return fmt.Sprintf("column value (%d) out of range [0..%d]", int(e), palletLength-1)
}

type errBadRow int

func (e errBadRow) Error() string {
	return fmt.Sprintf("row value (%d) out of range [0..%d]", int(e), palletWidth-1)
}

var errCellNotEmpty = errors.New("cell not empty")

// checkCellValid() returns nil if location (col, row) is valid, otherwise
// it returns an error string.
func checkCellValid(col, row uint8) error {
	if col < 0 || col >= palletLength {
		return errBadCol(col)
	}
	if row < 0 || row >= palletWidth {
		return errBadRow(row)
	}
	return nil
}

// A packPlan tracks the available pallet area as boxes are placed on a pallet.
// The packPlan consists of cells, each being one unit wide and one unit long.
type packPlan [palletWidth * palletLength]byte

// clear 'empties' the packPlan.
func (p *packPlan) clear() {
	for i := 0; i < len(p); i++ {
		p[i] = emptyCell
	}
}

// cell returns contents of a packPlan location.
func (p packPlan) cell(col, row uint8) (byte, error) {
	if err := checkCellValid(col, row); err != nil {
		return emptyCell, err
	}
	return p[row*palletLength+col], nil
}

// setCell sets the contents of a packPlan location.
func (p *packPlan) setCell(col uint8, row uint8, b byte) (byte, error) {
	if err := checkCellValid(col, row); err != nil {
		return emptyCell, err
	}
	p[row*palletLength+col] = b
	return p[row*palletLength+col], nil
}

// isEmptyCell returns true if cell(col, row) is a valid location and empty;
// otherwise it returns false. If the location is invalid,
// an error message is returned.
func (p packPlan) isEmptyCell(col, row uint8) (result bool, err error) {
	result = false
	b := byte(0)
	if b, err = p.cell(col, row); b == emptyCell && err == nil {
		result = true
	}
	return
}

// markBox marks the packPlan where a box will be placed.
// Since boxes are placed to the left and down,
// we only need to mark the left edge of the box.
func (p *packPlan) markBox(col, row, width uint8) error {

	isEmpty, err := p.isEmptyCell(col, row)
	switch {
	case err != nil:
		return err
	case !isEmpty:
		return errCellNotEmpty
	}

	isEmpty, err = p.isEmptyCell(col, row+width-1)
	switch {
	case err != nil:
		return err
	case !isEmpty:
		return fmt.Errorf("packPlan.markBox(%d,%d,%d) failed, ending cell is not empty.", col, row, width)
	}

	for i := row; i < row+width; i++ {
		if ok, _ := p.isEmptyCell(col, i); ok {
			p.setCell(col, i, byte('#'))
		}
	}
	return nil
}

// getAvailSpace returns the largest available space to the left and down
// of a given location on the packPlan. If cell(col, row) is not empty,
// or is an invalid location, 0 length and 0 width are returned.
func (p packPlan) getAvailSpace(cell openCell) (length, width uint8) {
	var test bool
	var i uint8

	if test, _ = p.isEmptyCell(cell.col, cell.row); !test {
		return 0, 0
	}

	for i, test = cell.col+1, true; test; i++ {
		test, _ = p.isEmptyCell(i, cell.row)
	}
	length = i - cell.col - 1

	// we will never 'undercut' a row, so width always extends
	// to the bottom edge of the pallet
	width = palletWidth - cell.row
	return
}

// String returns an ASCII diagram of the packPlan as a string.
func (p packPlan) String() string {
	var b byte
	var err error
	var s string
	s = "\n"

	for row := 0; row < palletWidth; row++ {
		for col := 0; col < palletLength; col++ {

			if b, err = p.cell(uint8(col), uint8(row)); err != nil {
				s += fmt.Sprintf("%v, \n", err)
				return s
			}

			if b == emptyCell {
				s += "_"
			} else {
				s += string(b)
			}
		}
		s += "\n"
	}
	return s
}

var errStackEmpty = errors.New("stack is empty")

// An openCell stores the upper left corner of an open space on a pallet.
type openCell struct {
	col, row uint8
}

// An openCellStack stores a stack of the current open spaces on a pallet.
type openCellStack []openCell

// peek returns the openCell at the top of the stack without popping it.
// If the stack is empty, return invalid cell coordinates and an error.
func (s openCellStack) peek() (openCell, error) {
	if len(s) == 0 {
		return openCell{palletLength, palletWidth}, errStackEmpty
	}
	return s[len(s)-1], nil
}

// push will place valid openCells onto the stack,
// ignoring invalid cell locations.
func (s *openCellStack) push(c openCell) {
	if err := checkCellValid(c.col, c.row); err == nil {
		(*s) = append((*s), c)
	}
}

// pop removes the top openCell from the stack and returns it.
// If the stack is empty, return invalid cell coordinates and an error.
func (s *openCellStack) pop() (openCell, error) {
	if len(*s) == 0 {
		return openCell{palletLength, palletWidth}, errStackEmpty
	}
	c := (*s)[len(*s)-1]
	(*s) = (*s)[:len(*s)-1]
	return c, nil
}

// clear empties the stack.
func (s *openCellStack) clear() {
	(*s) = (*s)[:0]
}

// A cargo record holds box information, including its area.
type cargo struct {
	item box
	area int
}

type errCargoIndexOutOfRange int

func (e errCargoIndexOutOfRange) Error() string {
	return fmt.Sprintf("cargoQueue index (%d) out of range", int(e))
}

// String returns string representation of a cargo record.
func (c cargo) String() string {
	return fmt.Sprintf("<ID: %d, %d x %d, %d area>", c.item.id, c.item.l, c.item.w, c.area)
}

// A cargoQueue implements sort.Interface for []cargo based on the area field.
type cargoQueue []cargo

func (q cargoQueue) Len() int           { return len(q) }
func (q cargoQueue) Swap(i, j int)      { q[i], q[j] = q[j], q[i] }
func (q cargoQueue) Less(i, j int) bool { return q[i].area < q[j].area }

// sort reorders the cargoQueue's boxes by descending area.
func (q cargoQueue) sort() {
	sort.Sort(sort.Reverse(q))
}

// addCargo appends a cargo record to the queue.
func (q *cargoQueue) addCargo(c cargo) {
	(*q) = append((*q), c)
}

// removeCargo removes a given cargo record from cargoQueue.
func (q *cargoQueue) removeCargo(i int) error {
	if i < 0 || i >= len(*q) {
		return errCargoIndexOutOfRange(i)
	}

	for j := i; j < len(*q)-1; j++ {
		(*q)[j], (*q)[j+1] = (*q)[j+1], (*q)[j]
	}
	(*q) = (*q)[:len(*q)-1]
	return nil
}

// findNextCandidate looks for the next box in the queue with an area
// equal to or smaller than an available space on the pallet. If a candidate
// is found, the function returns the index of the candidate, and ok is set
// to true. If not, the function returns 0 for the index, and ok set to false.
func (q cargoQueue) findNextCandidate(target, curr int) (int, bool) {
	if curr < 0 {
		curr = 0
	}
	for ; curr < len(q); curr++ {
		if q[curr].area <= target {
			return curr, true
		}
	}
	return 0, false
}

// turnVertical turns the box so its longest side is its width.
func (b *box) turnVertical() {
	if b.l > b.w {
		b.w, b.l = b.l, b.w
	}
}
