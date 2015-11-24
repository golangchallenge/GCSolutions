package main

import (
	"fmt"
)

// ErrUnsolvable indicates a Sudoku board is unsolvable
type ErrUnsolvable struct {
	msg string
}

// NewErrUnsolvable returns a new ErrUnsolvable error
func NewErrUnsolvable(format string, a ...interface{}) ErrUnsolvable {
	return ErrUnsolvable{msg: fmt.Sprintf(format, a...)}
}

func (e ErrUnsolvable) Error() string {
	return e.msg
}
