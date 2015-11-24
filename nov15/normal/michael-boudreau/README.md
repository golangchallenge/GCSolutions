# Sudoku Solver
## Go Challenge #8

This program has some fun to it. It has two different approaches (aka finders) to solve sudoku puzzles and it also will colorize the solution as it is solving it, including showing backtracking. 

Works best with a black terminal.

## Getting the Program

    go get github.com/mkboudreau/sudoku
    make
    #optional
    #make install

## Usage 

### Two Modes of Operation
1. Provide a board (3 different ways)
    - `./sudoku < sample`
    - `cat sample | ./sudoku`
    - `./sudoku <paste in board>`

2. Generate a board
    - `./sudoku -g <difficulty>`

### Options
    `-h` Highlights the solution
    `-p` Shows the game actually being solved. Very Fun! Enlarge your terminal. :)
    `-t <time duration>` Default is "1ms". Slow it down to watch the solution play out, including backtracking. (try 250ms)
    `-f <finder>` There are two different next available coordinate finders in the code. The default is "closest" which is faster for any given coordinate, but can be super slow when backtracking a lot. The other one that is implemented is "rank" which will recompute the best available coordinate among remaining available coordinates. This appears to be faster on harder boards.
    `-d` Displays the difficulty of the board. Default is true
    `-a` Turns on autosolving, so the program will not prompt you to be ready.
    `-g <difficulty>` Create new puzzle with diffulty between 1-5 with 5 being the most difficult. Default is 3 (default -1)
    `-l <loglevel>` Sets the log level. [off,error,debug,trace] (default "error")
  
## Examples

### Solves the sample file and displays the difficulty
- Option 1: `./sudoku < sample`

- Option 2: `cat sample | ./sudoku`

### Solves the sample file, highlights the solution
    cat sample | ./sudoku -h

### Solves the sample file, showing the solution at a very fast speed
    cat sample | ./sudoku -p

### Solves the sample file, showing the solution at a slower speed and using the rank finder
    cat sample | ./sudoku -p -t 500ms -f rank

### Creates a hard puzzle, shows the puzzle, and prompts user to solve
    ./sudoku -g 5 

### Creates a hard puzzle, auto-solves using the rank finder and shows it solve with a 250ms delay
    ./sudoku -f rank -p -t 250ms -g 5 -a

