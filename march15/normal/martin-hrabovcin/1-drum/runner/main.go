package main

import (
	// "encoding/hex"
	"fmt"
	drum "github.com/mhrabovcin/go-challenge/1-drum"
	// "io/ioutil"
	"path/filepath"
)

func main() {

	fmt.Println("")

	// Get list of files
	files, _ := filepath.Glob("../fixtures/*.splice")

	// Loop through files
	for file := range files {
		fmt.Println(files[file])
		// data, _ := ioutil.ReadFile(files[file])
		// fmt.Println(hex.Dump(data[6:14]))
		// fmt.Println(hex.Dump(data[14:100]))

		pattern, err := drum.DecodeFile(files[file])
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(pattern)
		}
	}
}
