package testfunc

import (
	"fmt"
)

//Unused is an example of unused function
func Unused() {
	Used()
	fmt.Printf("I'm unused function\n")
}

//Used is an example of used function
func Used() {
	fmt.Printf("Don't remove me please!\n")
}
