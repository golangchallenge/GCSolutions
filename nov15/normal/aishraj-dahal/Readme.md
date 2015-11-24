## Doku ##

Doku is a simple sudoku solver. It uses constraint propagation for the first phase and then uses backtracking in the second phase (if the 1st phase didn't work out). Its inspired by Peter Norvig's sudoku solver.

### Usage ###

Doku reads the input from the command line in the given form:

```
1 _ 3 _ _ 6 _ 8 _
_ 5 _ _ 8 _ 1 2 _
7 _ 9 1 _ 3 _ 5 6
_ 3 _ _ 6 7 _ 9 _
5 _ 7 8 _ _ _ 3 _
8 _ 1 _ 3 _ 5 _ 7
_ 4 _ _ 7 8 _ 1 _
6 _ 8 _ _ 2 _ 4 _
_ 1 2 _ 4 5 _ 7 8
```

Note that the last character should be EOF (end of file).
