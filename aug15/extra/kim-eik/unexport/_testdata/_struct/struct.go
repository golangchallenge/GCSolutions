package _struct

type Struct struct {
	A int
	b string
	S *Struct
}

func (*Struct) Foo() {}

func (Struct) Bar() {}
