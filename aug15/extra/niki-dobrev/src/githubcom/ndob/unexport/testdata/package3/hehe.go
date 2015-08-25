package package3

import (
	"fmt"

	"github.com/ndob/unexport/testdata/package1"
)

func main() {

	fmt.Println("hey", package1.Derp, package1.B)
	fmt.Println("hey", package1.Derp)

	package1.Asdf()
}
