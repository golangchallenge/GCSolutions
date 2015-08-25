package main

import (
	"fmt"

	"github.com/dooman87/gounexport/testdata/testvar"
)

func main() {
	fmt.Printf("Hello %s, %d", testvar.UsedVar, testvar.UsedConst)
}
