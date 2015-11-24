# soodohkoo
soodohkoo is a sudoku puzzle generator & solver.

## Flags

* `-mode=` - Controls the operational mode of the program.  
  * `solve` - Solves a single board provided over STDIN.
  * `solveStream` - Solves multiple boards provided over STDIN. Program exits with non-zero on the first invalid board.  
  * `generate` - Creates a new board.

* `-difficulty=` - Used with `--mode=generate` to control the difficulty of the generated board. Difficulty is judged by the number of unknown tiles.
  * `1`-`64` - How many tiles to set unknown.
  * `easy` - Synonym for `45`.
  * `medium` - Synonym for `50`.
  * `hard` - Synonym for `55`.
  * `insane` - Synonym for `60`.

  *Note:* the actual number of unknown tiles might be less than the value provided if during the generation process the program can remove no further tiles.

* `-stats` - Used with `--mode=solve` to show algorithm statistics after solving the puzzle.

## Solver input format

When using `-mode=solve` and `-mode=solveStream`, the board must be provided in the format of:

    1 _ 3 _ _ 6 _ 8 _
    _ 5 _ _ 8 _ 1 2 _
    7 _ 9 1 _ 3 _ 5 6
    _ 3 _ _ 6 7 _ 9 _
    5 _ 7 8 _ _ _ 3 _
    8 _ 1 _ 3 _ 5 _ 7
    _ 4 _ _ 7 8 _ 1 _
    6 _ 8 _ _ 2 _ 4 _
    _ 1 2 _ 4 5 _ 7 8

or

    1 _ 3 _ _ 6 _ 8 _ _ 5 _ _ 8 _ 1 2 _ 7 _ 9 1 _ 3 _ 5 6 _ 3 _ _ 6 7 _ 9 _ 5 _ 7 8 _ _ _ 3 _ 8 _ 1 _ 3 _ 5 _ 7 _ 4 _ _ 7 8 _ 1 _ 6 _ 8 _ _ 2 _ 4 _ _ 1 2 _ 4 5 _ 7 8
