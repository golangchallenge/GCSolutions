# sugoku

Soduke solver using constaint propagation. Solution for the go challenge at http://golang-challenge.com/go-challenge8/

#### Usage:

The program will read the problem from stdin and solve it, if the format is correct, and if the sudoku has a solution

To solve the challenge in challenges/hard2.txt:

    go run sugoku.go < challenges/hard2.txt

If you want the program to rate the sudoku, just provide the rate flag

    go run sugoku.go --rate < challenges/hard2.txt
