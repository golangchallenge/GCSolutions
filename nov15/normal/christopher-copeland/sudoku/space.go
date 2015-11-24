package main

import "fmt"

const (
	nVals  = 9
	nGrids = 3
)

type pos struct {
	row, col int
}

func (x pos) rowEq(y pos) bool {
	return x.row == y.row
}

func (x pos) colEq(y pos) bool {
	return x.col == y.col
}

func (x pos) subgridEq(y pos) bool {
	xsub := pos{x.row / nGrids, x.col / nGrids}
	ysub := pos{y.row / nGrids, y.col / nGrids}
	return xsub.colEq(ysub) && xsub.rowEq(ysub)
}

func (x pos) anyEq(y pos) bool {
	return x.rowEq(y) || x.colEq(y) || x.subgridEq(y)
}

type space struct {
	possible     [nVals]bool
	value        int
	fixed        bool
	usedForPrune bool
}

var emptySpace = space{
	fixed:        false,
	value:        -1,
	possible:     [nVals]bool{true, true, true, true, true, true, true, true, true},
	usedForPrune: false,
}

// Set fixes the space's value to v and eliminates other possibilities
func (s *space) Set(v int) {
	s.fixed = true
	s.value = v
	for i := range s.possible {
		s.possible[i] = i == v
	}
}

// numPossible returns the number of possible values that the
// space could be
func (s *space) numPossible() int {
	n := 0
	for _, p := range s.possible {
		if p {
			n++
		}
	}
	return n
}

// prune eliminates the possibility of v in the space it returns true when this
// change actually narrowed the possibilities for the space, and false if the
// value was already eliminated a non-nil error is returned if pruning v from
// the space would leave the space with no possible values
func (s *space) prune(v int) (bool, error) {
	if s.fixed && s.value == v {
		return false, fmt.Errorf("can't prune %v", v+1)
	}
	if s.fixed {
		return false, nil
	}
	if s.possible[v] {
		s.possible[v] = false
		if s.numPossible() == 1 {
			s.fixed = true
			for v := range s.possible {
				if s.possible[v] {
					s.value = v
				}
			}
		}
		return true, nil
	}
	return false, nil
}
