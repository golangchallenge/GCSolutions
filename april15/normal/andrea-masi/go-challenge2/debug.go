// +build debug

package main

import (
	"fmt"
	"os"
)

func debugPrintln(args ...interface{}) {
	// we use Stderr to not break any axample tests which evaluate Stdout
	fmt.Fprintln(os.Stderr, args...)
}
