package packa

import (
	"fmt"
)

var (
	// UnusedVar should show in the result
	UnusedVar = 1
	// UsedVar is used by github.com/isaiah/unexport/test_data/b
	UsedVar = 2
)

const (
	UsedConst   = 3
	UnusedConst = 4
	unusedConst = 5
)

type C interface {
	Count() int
}

type D interface {
	Dump() string
}

type A struct {
	X int
	B
}

type B struct {
	Y int
}

func NewA(i int) A {
	return A{X: i}
}

func (a *A) String() string {
	return fmt.Sprintf("a is %d", a.X)
}

func (a A) Count() int {
	return a.X
}

var _ C = &A{X: 0}
