package package2

import (
	"fmt"

	"github.com/ndob/unexport/testdata/package1"
)

func main() {
	fmt.Println("hey", package1.Derp)
	package1.Asdf()

	var c package1.StructTest2
	fmt.Println("a", c.Abc)

	p := package1.NewInnerStruct()
	fmt.Println("hey", package1.Derp, package1.B)
	fmt.Println("haha:", p.Fridge)

	p.Foo()
}
