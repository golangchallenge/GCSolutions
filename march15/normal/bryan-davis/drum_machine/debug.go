package drum // package main

import (
	"fmt"
	"path"
)

var files = [5]string{
	"pattern_1.splice",
	"pattern_2.splice",
	"pattern_3.splice",
	"pattern_4.splice",
	"pattern_5.splice",
}

func Debug() {
	for _, v := range files {
		p, _ := DecodeFile(path.Join("fixtures", v))
		fmt.Println(p)
	}
}
