Go language based sudoku solver for 9x9 grid.
Main is located in solver.go file.

To build:
```
go build datatypes/datatypes.go
go build solver.go
```

To run:
* Input format is a 9x9 matrix where each element in a row is space delimited. Allowed elements are 1-9, and _ for blanks.
* Output shows the solved grid, with number of solutions, and the difficulty level.

```
./solver
_ _ _ _ 4 5 _ _ _
8 _ _ _ _ _ 2 _ 7
_ _ 2 _ _ _ _ _ 4
_ _ 6 _ _ _ 3 _ 2
_ _ _ 1 _ _ _ _ _
2 _ 7 4 _ _ 6 _ _
6 4 _ _ 9 8 _ _ _
7 9 _ _ _ 4 _ _ _
_ _ _ _ _ _ _ 3 _


output:

1 7 9 2 4 5 8 6 3
8 6 4 9 1 3 2 5 7
3 5 2 7 8 6 1 9 4
9 1 6 8 5 7 3 4 2
4 3 5 1 6 2 9 7 8
2 8 7 4 3 9 6 1 5
6 4 3 5 9 8 7 2 1
7 9 1 3 2 4 5 8 6
5 2 8 6 7 1 4 3 9

Total solutions: 1
Difficulty level: hard
```
