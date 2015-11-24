package main

import (
	"fmt"
	"os"
)

/*0 1 2   3 4 5   6 7 8
0 1 0 3 | 0 0 6 | 0 8 0
1 0 5 0 | 0 8 0 | 1 2 0
2 7 0 9 | 1 0 3 | 0 5 6
  ------+-------+-------
3 0 3 0 | 0 6 7 | 0 9 0
4 5 0 7 | 8 0 0 | 0 3 0
5 8 0 1 | 0 3 0 | 5 0 7
  ------+-------+-------
6 0 4 0 | 0 7 8 | 0 1 0
7 6 0 8 | 0 0 2 | 0 4 0
8 0 1 2 | 0 4 5 | 0 7 8

1 _ 3 _ _ 6 _ 8 _ _ 5 _ _ 8 _ 1 2 _ 7 _ 9 1 _ 3 _ 5 6 _ 3 _ _ 6 7 _ 9 _ 5 _ 7 8 _ _ _ 3 _ 8 _ 1 _ 3 _ 5 _ 7 _ 4 _ _ 7 8 _ 1 _ 6 _ 8 _ _ 2 _ 4 _ _ 1 2 _ 4 5 _ 7 8
*/
func main() {
	fmt.Println("Enter init sudoku matrix (alow [\\s_1-9]):")
	b := &Board{}
	var err error
	err = b.Read(os.Stdin)
	if nil != err {
		fmt.Println("Error in input data, puzzles cannot be solved")
		return
	}
	s := NewSolver(b)
	s.Solve()
	if s.IsSolved() {
		fmt.Println(s.StringBoard())
		return
	}
	fmt.Println("This puzzle is to hard for my simple algoritm and can't be solved, sorry.")
}
