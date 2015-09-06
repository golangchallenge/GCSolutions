package packb

import (
	"fmt"
	"github.com/isaiah/unexport/test_data/a"
)

type B struct {
	packa.A
	Z string
}

func NewB(i int) B {
	return B{packa.NewA(i), "Z"}
}

func (b *B) String() string {
	return fmt.Sprintf("#{B %d, %s}", b.A, b.Z)
}

func (b *B) Sum() int {
	return b.X + packa.UsedVar + packa.UsedConst
}

func (b *B) Dump() string {
	return b.String()
}

func Puts(dumper packa.D) string {
	return dumper.Dump()
}
