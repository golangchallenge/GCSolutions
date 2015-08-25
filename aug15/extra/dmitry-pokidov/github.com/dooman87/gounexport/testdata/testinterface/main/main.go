package main

import (
	"fmt"
	"sort"

	"github.com/dooman87/gounexport/testdata/testinterface"
)

type impl struct {
	s string
}

func (i impl) SayHello() string {
	fmt.Printf("Hello %s", i.s)
	return i.s
}

func main() {
	//testing internal interface
	var i impl
	i.s = "Bob"
	hello(i)
	/*
	   TODO:
	*/
	//testing external interface
	s := new(testinterface.SortImpl)
	s.Arr = []int{3, 2, 1}
	sort.Sort(s)
}

func hello(s testinterface.UsedInterface) {
	s.SayHello()
}
