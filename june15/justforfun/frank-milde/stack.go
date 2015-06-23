//
// =========================================================================
//
//       Filename:  stack.go
//
//    Description:  Implements a stack by using slices not lists.
//
//        License:  GNU General Public License
//      Copyright:  Copyright (c) 2015, Frank Milde
//
// =========================================================================
//

package main

import (
	"fmt"
)

type Stack []box

func NewStack() Stack {
	var s []box
	return s
}

func (s Stack) IsEmpty() bool { return len(s) == 0 }

// Front returns last element of slice, which is the front of the stack
func (s Stack) Front() box {
	if s.IsEmpty() {
		return emptybox
	}
	return s[len(s)-1]
}

// Push appends box b to stack pointer sp
func (sp *Stack) Push(b box) {
	*sp = append(*sp, b)
}

// Pop returns the last box of stack. If stack is empty, emptybox is
// returned.
func (sp *Stack) Pop() box {
	if (*sp).IsEmpty() {
		return emptybox
	}

	last := len(*sp) - 1
	b := (*sp)[last]
	//	s[last] = nil // or the zero value of T
	(*sp) = (*sp)[:last]
	return b
}

// StacksAreEqual compares two stacks s1,s2 and returns true if s1 and
// s2 have the same length and the same boxes at the same positions.
func StacksAreEqual(s1, s2 Stack) bool {
	if len(s1) != len(s2) {
		return false
	}

	for i, b := range s1 {
		if !BoxesAreEqual(b, s2[i]) {
			return false
		}
	}
	return true
}

func (s Stack) String() string {

	var outstring string
	for _, b := range s {
		outstring += fmt.Sprintf("(%v)  ", b)
	}
	return outstring
}
