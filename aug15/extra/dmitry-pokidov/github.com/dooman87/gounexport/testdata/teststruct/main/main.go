package main

import (
	"github.com/dooman87/gounexport/testdata/teststruct"
)

func main() {
	s := new(teststruct.UsedStruct)
	s.UsedField = "Hello"
	s.UsedFieldPointer = new(teststruct.ComplexType)
	s.UsedMethod()
}
