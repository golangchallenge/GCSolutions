## Preamble

This code is my submission for the [Go Challenge - 8](http://golang-challenge.com/go-challenge8/). The code has hopefully fulfilled the requirements and bonus features of the challenge.

* Name: Kelvin Yong
* Country: Singapore
* GitHub id: kelvin-yong
* Category: Just participating

A [summary](#summary) of the challenge is found at the end of this document.

## Running the code
### Directory
	.
	├── README.md
	├── main.go
	├── dlx
	│   ├── dlx.go
	│   └── dlx_test.go
	└── sudoku
	    ├── generate.go
	    ├── grader.go
	    ├── sudoku.go
	    ├── sudoku_test.go
	    └── testdata
	        ├── ...
	        └── ...

### Building
`go build main.go`

### Solving a sudoku puzzle
Either  
`.main`

Or  
`./main < sudoku/testdata/input_valid1.txt`

The puzzle will be solved and the estimated level of difficulty is shown.

### Generating a puzzle

`./main -generate=x` to generate a puzzle, where x can be  

* 1: easy
* 2: medium
* 3: hard
* 4: evil

Both the puzzle and the solution will be printed.

## Main Features

### Puzzle Solving

The solver uses Knuth's "Dancing Links" ([DLX](http://arxiv.org/pdf/cs/0011047v1.pdf)) to solve the suduko puzzles and it is very fast. On my machine MacBook Pro (Retina, 13-inch, Early 2015, 2.7 GHz Intel Core i5), the time to solve the puzzles:

* Simple: 0.166ms
* Hard: 2.844ms
* Top 95 puzzles: 37.75ms, or an average of 0.4ms per puzzle.


### Difficulty Grading

DLX is a brute force method and cannot be used to grade puzzles.

To grade puzzles, a set of 4 human strategies is coded:

* [Naked Singles](http://www.sadmansoftware.com/sudoku/nakedsingle.php)
* [Hidden Singles](http://www.sudoku129.com/puzzles/tips_1.php)
* [Naked Pair](http://www.sudoku129.com/puzzles/tips_3.php)
* [Locked Candidates](http://www.sudoku129.com/puzzles/tips_2.php)

Basic idea is that an easy puzzle can be solved by Naked Singles alone. The Naked Singles strategy is the core of solving any puzzle - if you have solved any single square of the puzzle, you would have used Naked Singles!

A more complicated (medium or hard)puzzle would require 2 or 3 strategies to solve.

An extremely tough (evil) puzzle would take all 4 strategies to solve or remains unsolved after applying the 4 strategies.

The grader also takes into consideration how many initial unknown squares are in the puzzle and how many rounds each strategy was applied.

Using 100 puzzles (Easy, Medium, Hard, Evil x 25 each) from [Web Sudoku](http://www.websudoku.com/), the grader grades

* 88 puzzles correctly
* 12 puzzles by a grade easier or harder
* 0 puzzles by more than a grade difference.

5 puzzles out of 100 remains unsolvable after applying the 4 strategies.

### Puzzle Generation

A sudoku puzzle should have only 1 solution.

The generator fills the the first row of a 9x9 grid with values 1 through 9. Subsequently it uses DLX to solve the puzzle randomly. The grid now has all 81 cells filled up and satisfies all constraints of a Sudoku puzzle.

Then we randomly choose a cell to remove the number. We check if the puzzle now has multiple solutions using DLX. If no, we pick another cell to remove. Otherwise, we put back the cell and we are done.

> This generates an “irreducible puzzle”: a grid from which no more cells can be removed without leaving multiple solutions. It’s a sudoku puzzle, but at this stage we don’t know anything about how hard it is!
>
> -- <cite>[Zendoku puzzle generation](http://garethrees.org/2007/06/10/zendoku-generation/#section-4)</cite>

We then grade our new puzzle. If this puzzle matches our desired difficulty level, we are done. If not, we will generate another puzzle.


## Other Features
### Commented code

	┌───────────────────────┬──────────┬────────┬─────────┬───────┐
	│ Path                  │ Physical │ Source │ Comment │ Empty │
	├───────────────────────┼──────────┼────────┼─────────┼───────┤
	│ main.go               │ 88       │ 78     │ 0       │ 10    │
	├───────────────────────┼──────────┼────────┼─────────┼───────┤
	│ sudoku/sudoku.go      │ 248      │ 162    │ 52      │ 36    │
	├───────────────────────┼──────────┼────────┼─────────┼───────┤
	│ sudoku/generate.go    │ 163      │ 106    │ 34      │ 23    │
	├───────────────────────┼──────────┼────────┼─────────┼───────┤
	│ sudoku/grader.go      │ 431      │ 310    │ 65      │ 56    │
	├───────────────────────┼──────────┼────────┼─────────┼───────┤
	│ dlx/dlx.go            │ 203      │ 154    │ 35      │ 20    │
	└───────────────────────┴──────────┴────────┴─────────┴───────┘

*Tip*: godoc is helpful too

### Coverage
`go tool cover -html=dlx/coverage.out`  
`go tool cover -html=sudoku/coverage.out`  

* Package dlx: 100.0% of statements
* Package sudoku: 95.8% of statements (the remaining 4.2% mostly deal with pretty print of the grid )

## Test Data and Benchmark
### Test Data
The following test data were used.

#### Inputs for main
* input_valid1.txt: a valid puzzle
* input_valid2.txt: another valid puzzle
* input_duplicate.txt: an invalid puzzle that does not meet Sudoku rules
* input_multiple_solution.txt: a puzzle that has multiple solution
* input_tooshort.txt: an invalid puzzle as the input has less than 81 squares
* input_unsolvable.txt: an invalid puzzle as the puzzle has no solution

#### Data for testing and benchmarking
* `p_general.txt`: a set of 2 puzzles/solutions
* `p_invalid.txt`: a set of 2 invalid puzzles
* `p_level_easy.txt`: a set of 25 easy puzzles/solutions from Web Sudoku
* `p_level_medium.txt`: a set of 25 medium puzzles/solutions from Web Sudoku
* `p_level_hard.txt`: a set of 25 hard puzzles/solutions from Web Sudoku
* `p_level_evil.txt`: a set of 25 evil puzzles/solutions from Web Sudoku
* `p_method_nakedsingle.txt`: a set of 2 puzzles/solutions that can be solved using naked single strategy
* `p_method_hiddensingle.txt`: a set of 2 puzzles/solutions that can be solved using hidden single strategy
* `p_method_lockedtype.txt`: a set of 2 puzzles/solutions that can be solved using locked type
* `p_method_nakedpair.txt`: a set of 2 puzzles/solutions that can be solved using naked pair strategy
* `top95.txt`: a set of [95 hardest puzzles](http://magictour.free.fr/top95)

Puzzles for `p_method_xxx` are obtained from [Sadman Suduku](http://www.sadmansoftware.com/sudoku/solvingtechniques.php).

### Benchmark for solving via DLX
Solving via DLX takes between 0.166ms to 2.844ms.  

Solving an unsolvable puzzle takes 6.7ms to know it is unsolvable.

A set of 95 hardest puzzles takes 37ms to solve, or an average of 0.4ms per puzzle.

	BenchmarkDLX_Simple-4           	   10000	    165988 ns/op
	BenchmarkDLX_Hard-4             	     500	   2844345 ns/op
	BenchmarkDLX_NakedSingle-4      	   10000	    215284 ns/op
	BenchmarkDLX_HiddenSingle-4     	   10000	    214345 ns/op
	BenchmarkDLX_NakedPair-4        	   10000	    202101 ns/op
	BenchmarkDLX_LockedType-4       	   10000	    225713 ns/op
	BenchmarkDLX_Unsolvable-4       	     200	   6669019 ns/op
	BenchmarkDLX_AllTop95-4         	      50	  37753880 ns/op

### Benchmark for solving using human strategies
Human strategies aren't too shabby either. If a puzzle can be solved by just Naked Singles or Naked and Hidden Singles, it is faster than using DLX.

	BenchmarkHuman_NakedSingle-4    	   10000	    161050 ns/op
	BenchmarkHuman_HiddenSingle-4   	   10000	    157596 ns/op
	BenchmarkHuman_NakedPair-4      	    3000	    429871 ns/op
	BenchmarkHuman_LockedType-4     	    3000	    546697 ns/op


### Benchmark for solving generating puzzles
An easy puzzle takes about 1.68ms to generate while an evil puzzle takes 123ms to generate. In 1000 runs of random generation, only 4 puzzles were of "evil" difficulty.

	BenchmarkGeneratePuzzle_Easy-4  	    1000	   1681468 ns/op
	BenchmarkGeneratePuzzle_Medium-4	     100	  19214934 ns/op
	BenchmarkGeneratePuzzle_Hard-4  	      10	 107276052 ns/op
	BenchmarkGeneratePuzzle_Evil-4  	      20	 123118755 ns/op



## <a name="summary"></a>Summary of Go Challenge 8
The challenge is summarised here. For full details, please see [Go Challenge - 8](http://golang-challenge.com/go-challenge8/).

### The Goal of the challenge

The goal of this challenge is to implement a Sudoku solver.

### Requirements of the challenge

Your program should read a puzzle of this form from standard input:


	1 _ 3 _ _ 6 _ 8 _
	_ 5 _ _ 8 _ 1 2 _
	7 _ 9 1 _ 3 _ 5 6
	_ 3 _ _ 6 7 _ 9 _
	5 _ 7 8 _ _ _ 3 _
	8 _ 1 _ 3 _ 5 _ 7
	_ 4 _ _ 7 8 _ 1 _
	6 _ 8 _ _ 2 _ 4 _
	_ 1 2 _ 4 5 _ 7 8

And it should write the solution to standard output:


	4 5 6 7 8 9 1 2 3
	7 8 9 1 2 3 4 5 6
	2 3 4 5 6 7 8 9 1
	5 6 7 8 9 1 2 3 4
	8 9 1 2 3 4 5 6 7
	3 4 5 6 7 8 9 1 2
	6 7 8 9 1 2 3 4 5
	9 1 2 3 4 5 6 7 8

### Bonus features:

* Print a rating of the puzzle's difficulty (easy, medium, hard). This rating should roughly coincide with the ratings of shown by sites like Web Sudoku.
* Implement a puzzle generator that produces a puzzle of the given difficulty rating.
* Maximize the efficiency of your program. (Write a benchmark.)
* Write test cases, and use the cover tool to make sure your tests are thorough. (I found a bug in both my implementation and a test case when I checked the coverage.)
* Use a non-obvious technique, like Knuth's "Dancing Links" or something of your own invention.
