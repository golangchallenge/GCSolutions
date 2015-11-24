# Sudoku Solver in Golang

This is a concurrent solver for Sudoku puzzles written in Golang.

### Solution strategy

- Repeat until fixed point:
  - For each space with a known value:
    - Eliminate that value from the possibilities of its peers (other spaces in
      the same row, column, or subgrid).
- If the board is solved:
  - Return.
- Otherwise:
  - Choose an unsolved space with the fewest possibilities.
  - For each possibility:
    - Create a copy of the board where the space is fixed to that possibility
      and spawn a new goroutine to solve that board with this strategy. The
      first goroutine to return a solved board will terminate the search. If
      none of the goroutines can find a solution, then an error message is
      emitted to indicate that the board is unsolvable.

### Example puzzles

The `puzzles` directory contains several puzzles that were obtained
directly from or indirectly via
[Peter Norvig's Sudoku page](http://norvig.com/sudoku.html).

### To build

- [Obtain Golang](https://golang.org/dl/)
- `go build`

### To run

`./sudoku < puzzles/top95_1.txt`
