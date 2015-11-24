package main

import (
	"fmt"
	"log"
)

func main() {
	var c rune
	var in string
	for i := 0; i < 9; i++ {
		for j := 0; j < 9; j++ {
			fmt.Scanf("%c", &c)
			in += string(c)
		}
		fmt.Scanf(" ")
	}
	s, err := NewSudoku(in)
	if err != nil {
		log.Fatal(err)
	}
	err = s.Solve()
	if err != nil {
		log.Fatal(err)
	}
	s.Print()
}
