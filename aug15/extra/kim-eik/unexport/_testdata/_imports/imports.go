package _imports

import "fmt"
import "github.com/netbrain/unexport/_testdata/_imports/foo"

func main() {
	fmt.Println(foo.FooConst)
	fmt.Println(foo.FooVar)
	fmt.Println(foo.Foo())

	f1 := &foo.FooStruct{}
	fmt.Println(f1.FooFunc())
	fmt.Println(f1.FooPtrFunc())
}
