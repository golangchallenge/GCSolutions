# Sudoku
This project is in response to the 8th golang challenge posted at http://golang-challenge.com/go-challenge8/. It is a cli application that takes a sudoku puzzle in via standard input and emits the solution to standard output.

It was written in golang exclusively by Bruce Downs <bruceadowns at gmail dot com>.

## Requirements of the challenge
* Read puzzle from standard input
* Write the solution to standard output
* Reject malformed or invalid inputs
* Recognize and report puzzles that cannot be solved

### Puzzle is well-formed if
* it has 9 rows
* each row has 9 columns
* each element is an underscore, or 1-9

### Bonus features
* Print a rating of the puzzle's difficulty (easy, medium, hard) - implemented
* Implement a puzzle generator that produces a puzzle of the given difficulty rating - not implemented
* Maximize the efficiency of your program (benchmark) - implemented
* Write test cases, and use the cover tool to make sure your tests are thorough - implemented
* Use a non-obvious technique, like Knuth's "Dancing Links" or something of your own invention - implemented via restricted brute force

### Hints
* try using an array (not a slice)
* consider using recursion to simplify implementation

## CLI Usage
```
$ sudoku --help
Usage of sudoku:
  -categorize
    	categorize the puzzle
  -dry
    	do not compute solution
  -verbose
    	emit verbose information
```

All options default to false.

Examples for Linux and OSX:

* `$ cat input/official.txt | sudoku`
* `$ cat input/official.txt | sudoku -categorize`
* `$ cat input/official.txt | sudoku -categorize -verbose`

Examples for Windows:

* `C:\>type input\official.txt | sudoku.exe`
* `C:\>type input\official.txt | sudoku.exe -categorize`
* `C:\>type input\official.txt | sudoku.exe -categorize -verbose`

## References
* http://golang-challenge.com/go-challenge8/
* https://twitter.com/golangchallenge
* https://en.wikipedia.org/wiki/Sudoku

## Journal
* learn the basics for how to play sudoku.
* establish a private gitlab project.
* read stdin and write stdout based on specific requirements.
* empty indicator implemented using _ and . and 0, but backed off per underscore requirement.
* added pass 1 of solution - sweep and eliminate potential values per row, column,box.
* learn how to solve medium puzzles.
* experimented with pluggable solvers.
* a pluggable solver registers itself at startup and a cli flag specifies it.
* relented due to inability to specify local packages relative to root.
* choose a static solver and a single source for cleaner handoff.
* downloaded puzzles as large text files with single, 81 character lines.
* wrote app to transform single line puzzle(s) to well-formed stdin.
* added dry-run option to baseline stdin and stdout operation.
* test coverage 86.7% (go test -cover). (ref https://blog.golang.org/cover)
* view test coverage using `go test -coverprofile cover.out` and `go tool cover -html=cover.out`.
* created tests and benchmarks (go test -bench).
* added verbose option to display ongoing solution.
* leverage io.Writer interface to control output (ioutil.Discard, os.Stdout)
* added tabwriter output to print incremental solution.
* added pass 2 of solution - propagate unique potential values.
* added recursion to simplify implementation.
* refined sweep and uniquify algorithms.
* strictly follow requirements
* considered ideas around a dynamically sized square
* completed pass 2 solution - this seems to solve easy and medium.
* learn how to solve medium puzzles.
* reviewed pass 1 and 2 and came up with ideas for pass 3.
* pass 3 will involve choosing the element with the least potential values and attempting to solve.
* implemented pass 3 - guess and solve.
* reworked return values to handle recursion for guessed puzzles.
* changed solve functions to accept a puzzle object and return a puzzle object.
* changed from having solve functions modify pointer reference.
* the app now solves easy,medium,hard. some have multiple solutions.
* took out runMain as it was unnecessary.
* added test/benchmark to solve every puzzle initially downloaded. (~30,000 in ~40sec on mbp)
* need to think where to take the app to next. something interesting wrt golang.
* after a clean sweep review of the code, I whittled the time to solve a hard puzzle down to ~millisecond (mbp).
* the code is complete.
* found a particularly egregious bug with guess/backtrack recursion affecting extreme puzzles.
* added proper verification of puzzle to ensure solution correctness.
* found some extremely difficult puzzles with lots of guessing run for > 1s on mbp.
* found sizable inefficiency in propagate unique algorithm.
* unless I can think of anything else interesting, I will submit the code as is.
