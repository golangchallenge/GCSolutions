Sudoku solver implemented for the 8th Go challenge - http://golang-challenge.com/go-challenge8/

I took a stab at implementing Knuths' Dancing Links algorithm (https://en.wikipedia.org/wiki/Dancing_Links) for this challenge. I can't say that I fully understand the entire inner workings of the algorithm but I do understand enough of it to iron out an implementation. It was a fun learning experience!

Tested on go version go1.5.1 linux/amd64

- Output from ./<binary> -help

      -generate string
            Generate a Sudoku board, accepts inputs: 'easy', 'medium' 'hard'
      -print-difficulty
            Prints the difficulty of the input board


- Running the solver
    
    Normal solving: cat /path/to/input | ./<binary>
    Print difficulty: cat /path/to/input | ./<binary> -print-difficulty
    Generate a puzzle: ./<binary> -generate {easy, medium, hard}
    Generate and solve: ./<binary> -generate {easy, medium, hard} | ./<binary>


- Solvers

    Two algorithms have been implemented for solving Sudoku puzzles
        - Backtracking (used as a reference implementation and in generating puzzles)
        - Knuths' Dancing Links algorithm (inspiration from https://www.ocf.berkeley.edu/~jchu/publicportal/sudoku/sudoku.paper.html). This is the default solver when running the program.


- On difficulty

    The difficulies when generating puzzles (and evaluating a boards difficulty) have been estimated by manually sampling different puzzles at websudoku.com.


- External dependencies

    github.com/stretchr/testify/assert - used for testing
