//
// =========================================================================
//
//       Filename:  hash.go
//
//    Description:  Implements the hash table to store the box stacks in.
//
//        License:  GNU General Public License
//      Copyright:  Copyright (c) 2015, Frank Milde
//
// =========================================================================
//

package main

import (
	"errors"
	"fmt"
)

const (
	TABLESIZE = 10
	SQUAREBOX = 5
)

type Table []Stack

type HashError int

var ErrSize error = errors.New("hash: Invalid size.")
var ErrOrient error = errors.New("hash: Invalid orientation.")
var ErrHash error = errors.New("hash: Invalid hash.")

func (t Table) IsEmpty() bool {
	for _, stack := range t {
		if !stack.IsEmpty() {
			return false
		}
	}
	return true
}

// NewTable returns a new Table of capazity TABLESIZE = 10
func NewTable() Table {
	store := make([]Stack, TABLESIZE)
	// In case we change the stack to work with *box we need to initialize the
	// individual stacks
	//	for i := 0; i != TABLESIZE; i++ {
	//		store[i].Init()
	//	}
	return store
}

// HashBox returns the hash [0-9] of box b from its size s=b.Size(). If the box
// has invalid dimensions or the size s is wrong, an error is returned.
func HashBox(b *box) (int, error) {

	var errVal int = 10

	if !b.HasValidDimensions() {
		return errVal, ErrSize
	}

	var hash int
	s := int(b.Size())

	switch s {
	case 1, 2, 3, 6:
		hash = s - 1
	case 4:
		if b.IsSquare() {
			hash = s
		} else {
			hash = s - 1
		}
	case 8:
		hash = 6
	case 9:
		hash = 7
	case 12:
		hash = 8
	case 16:
		hash = 9
	default:
		return errVal, ErrSize
	}

	return hash, nil
}

// Add pushes a box b to the appropriate box stack in Table t according to
// b's size. An error is returned when input is invalid box.
func (t Table) Add(b box) error {

	// this also covers the case of an emptybox
	if !b.HasValidDimensions() {
		return errors.New("Add box to table: Box has invalid size.")
	}

	hash, errHash := HashBox(&b)

	if errHash == nil {
		t[hash].Push(b)
		return nil
	}
	return errHash
}

// Hash takes a size s and orientation o and returns the hash [0-9] for a
// corresponding box. If an invalid size or orientation is given an error is
// returned.
func Hash(s int, o Orientation) (int, error) {

	var errVal int = 10

	if !(s >= 0 && s <= palletWidth*palletLength) {
		return errVal, ErrSize
	}
	if o != HORIZONTAL && o != VERTICAL && o != SQUAREGRID {
		return errVal, ErrOrient
	}

	var hash int

	switch s {
	case 1, 2, 3, 6:
		hash = s - 1
	case 4:
		if o == SQUAREGRID {
			hash = s
		} else {
			hash = s - 1
		}
	case 8:
		hash = 6
	case 9:
		hash = 7
	case 12:
		hash = 8
	case 16:
		hash = 9
	default:
		return errVal, ErrSize
	}

	return hash, nil
}

// GetBoxThatFitsOrIsEmpty will return the largest box b that fits in a grid of size s and
// orientation o from Table p. If no box is found in t, an emptybox is
// returned. If wrong size/orientation is given an error is returned.
// TODO: Proper error handling
func (t Table) GetBoxThatFitsOrIsEmpty(s int, o Orientation) (box, error) {

	hash, err := Hash(s, o)
	if err != nil {
		return emptybox, err
	}

	b := emptybox
	stackNr := hash

	// Start checking the stack at t[stackNr=hash] for box. If stack is empty,
	// check the next lower stack in table for a box until box is found or
	// stackNr == 0.
	// However, the layout of the table does not allow for every input size/
	// box type to simply fit into the next lower one:
	// If a 3x3 is requested, but not found then the next smaller sized box types
	// 3x2,3x1,2x2 will fit. But due to its geometry a 1x4 box will NOT FIT,
	// although it has a smaller size than 3x3.
	// So we have to carefully check the requested box type and exclude the
	// smaller boxes that do not  geometrically fit, if requested box is not
	// found.
	switch hash {
	case 9, 8, 6, 3, 2, 1, 0:
		for b == emptybox && stackNr >= 0 {
			b = t[stackNr].Pop() // for these box types all smaller boxes fit.
			stackNr--
		}
	case 7:
		for b == emptybox && stackNr >= 0 {
			if stackNr != 6 { // exclude [6] = 4x2. It does not fit into [7] = 3x3
				b = t[stackNr].Pop()
			}
			stackNr--
		}
	case 5:
		for b == emptybox && stackNr >= 0 {
			if stackNr != 3 { // exclude [3] = 4x1. It does not fit into [5] = 3x2
				b = t[stackNr].Pop()
			}
			stackNr--
		}
	case 4:
		for b == emptybox && stackNr >= 0 {
			// exclude [3] = 4x1 and [2] = 3x1. They do not fit into [4] = 2x2.
			if stackNr != 3 && stackNr != 2 {
				b = t[stackNr].Pop()
			}
			stackNr--
		}
	default:
		return emptybox, ErrHash
	}

	return b, nil
}

//
//
//
// TablesAreEqual returns true if Table t1 and t2 have the same length and
// their stacks are equal.
func TablesAreEqual(t1, t2 Table) bool {
	if len(t1) != len(t2) {
		return false
	}

	for i, s := range t1 {
		if !StacksAreEqual(s, t2[i]) {
			return false
		}
	}
	return true
}

// String interface to pretty print a Table
func (t Table) String() string {
	total := fmt.Sprintf("\n")
	for i, stack := range t {
		var label string
		switch i {
		case 0, 1, 2, 3, 5:
			label = fmt.Sprintf(" %d", i+1)
		case 4:
			label = fmt.Sprintf("4s")
		case 6:
			label = fmt.Sprintf(" %d", 8)
		case 7:
			label = fmt.Sprintf(" %d", 9)
		case 8:
			label = fmt.Sprintf("%d", 12)
		case 9:
			label = fmt.Sprintf("%d", 16)
		default:
			fmt.Println("default")
		}
		total += fmt.Sprintf("[%s]  -->  %v\n", label, stack)
	}
	return total
}
