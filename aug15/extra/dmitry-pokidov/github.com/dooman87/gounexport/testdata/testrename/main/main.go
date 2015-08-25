package main

import (
	"github.com/dooman87/gounexport/testdata/testrename"
)

func main() {
	s := new(testrename.UsedStruct)
	s.UsedField = "Hello"
	s.UsedMethod()
}
