package package1

import "fmt"

func Asdf() {
	fmt.Println("asdf")
}

type innerStruct struct {
	Fridge string
}

type StructTest struct {
	local bool
	Abc   string
	inner innerStruct
}

type StructTest2 struct {
	aaa int
	*StructTest
}

type StructSlice []StructTest

type someHandler interface {
	ServeHTTP(int, string)
	serveTCP(int, int)
}

func (t *innerStruct) Foo() {}
func (t StructTest) foo2()  {}

const Derp = "fii"

var a, B = 2, 4.3

const leaf = "asfd"

func NewInnerStruct() innerStruct {
	var p innerStruct
	return p
}
