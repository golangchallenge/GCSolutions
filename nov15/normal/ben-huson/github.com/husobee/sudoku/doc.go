// Package sudoku - A sudoku solver implementation for "The Go Challenge 8"
// Requirements of the Challenge:
// 1. Your program should read a puzzle of this form from standard input:
//1 _ 3 _ _ 6 _ 8 _
//_ 5 _ _ 8 _ 1 2 _
//7 _ 9 1 _ 3 _ 5 6
//_ 3 _ _ 6 7 _ 9 _
//5 _ 7 8 _ _ _ 3 _
//8 _ 1 _ 3 _ 5 _ 7
//_ 4 _ _ 7 8 _ 1 _
//6 _ 8 _ _ 2 _ 4 _
//_ 1 2 _ 4 5 _ 7 8
//
//And it should write the solution to standard output:
//
//1 2 3 4 5 6 7 8 9
//4 5 6 7 8 9 1 2 3
//7 8 9 1 2 3 4 5 6
//2 3 4 5 6 7 8 9 1
//5 6 7 8 9 1 2 3 4
//8 9 1 2 3 4 5 6 7
//3 4 5 6 7 8 9 1 2
//6 7 8 9 1 2 3 4 5
//9 1 2 3 4 5 6 7 8
//
// It should reject malformed or invalid inputs and recognize and report puzzles that cannot be solved.
package sudoku
