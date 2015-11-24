package main

import "errors"

const (
	nRows = 9
	nCols = 9
)

type Board [nRows][nCols]space

func NewBoard() (b Board) {
	for r := range b {
		for c := range b[r] {
			b[r][c] = emptySpace
		}
	}
	return
}

type posValue struct {
	pos   pos
	value int
}

// get the positions and values of all spaces on the board that have only one
// possible value. These spaces' usedForPrune field is marked to indicate
// that they have been returned from this function once already.
func (b *Board) getUnusedFixed() (res []posValue) {
	for r := range b {
		for c := range b[r] {
			if b[r][c].fixed && !b[r][c].usedForPrune {
				res = append(res, posValue{
					pos:   pos{row: r, col: c},
					value: b[r][c].value,
				})
				b[r][c].usedForPrune = true
			}
		}
	}
	return
}

var allFixedErr = errors.New("firstNonfixed: all spaces are fixed")

// findNonfixed returns the position and pointer to a space on the board that
// has the fewest possible values without being a solved space. This minimizes
// the branching required in the solver.
// A non-nil error is returned if the board is solved, meaning that
// no spaces have more than one possible value.
func (b *Board) findNonfixed() (p pos, s *space, err error) {
	minPossible := nVals + 1

	// will be overwritten to nil if any non-fixed spaces are found
	err = allFixedErr
	for r := range b {
		for c := range b[r] {
			if !b[r][c].fixed {
				if npossible := b[r][c].numPossible(); npossible < minPossible {
					minPossible = npossible
					p = pos{row: r, col: c}
					s = &b[r][c]
					err = nil
				}
			}
		}
	}
	return
}

// given a position on the board, return pointers to all spaces that are
// in the same row, column, or subgrid other than the original position
func (b *Board) getPeers(p pos) (res []*space) {
	for r := range b {
		for c := range b[r] {
			q := pos{row: r, col: c}
			if p.anyEq(q) && p != q {
				res = append(res, &b[r][c])
			}
		}
	}
	return
}
