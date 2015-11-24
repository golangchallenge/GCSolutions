# go-sudoku

`go-sudoku` is an entry for [Go Challenge 8](http://golang-challenge.com/go-challenge8/).

## Run

`go-sudoku < input.txt`

or, more interestingly:

`go-sudoku -steps < input.txt`

[See example](https://gist.githubusercontent.com/judwhite/792d3336dd4398c50186/raw/4a301869d7b6699e678931cef80f4e8c79137791/sudoku-steps.txt)

## Board input format

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

## Additional arguments

Run `go-sudoku -help` for a full list of optional arguments.

| Argument                | Description
|-------------------------|-------------
| `-time`                 | print time to solve
| `-steps`                | print steps an explanations to solve and eliminate candidates
| `-generate`             | generate a sudoku puzzle (unfortunately no difficulty selection)
| `-profile`              | enable CPU and memory profiling
| `-file`                 | run a set of Sudoku puzzles from a file
| `-max-puzzles`          | use with `-file` to limit the number of puzzles executed

## How it works

`go-sudoku` first attempts human strategy and ultimately falls back on a SAT solver.

The SAT solver takes advantage of some Sudoku characteristics to shorten execution time. It's rather good at determining unsolvable boards.

## Resources

- http://www.sudokuwiki.org/Strategy_Families
- https://en.wikipedia.org/wiki/Simplex_algorithm
- http://www.nature.com/articles/srep00725#f3
- http://arxiv.org/ftp/arxiv/papers/0805/0805.0697.pdf
- Satisfiability Solvers: http://www.cs.cornell.edu/gomes/papers/satsolvers-kr-handbook.pdf
- http://ocw.mit.edu/courses/electrical-engineering-and-computer-science/6-005-elements-of-software-construction-fall-2011/assignments/MIT6_005F11_ps4.pdf
- https://en.wikipedia.org/wiki/Conjunctive_normal_form
- https://en.wikipedia.org/wiki/DPLL_algorithm
- https://en.wikipedia.org/wiki/Backtracking
- https://en.wikipedia.org/wiki/Unit_propagation
- http://ocw.mit.edu/courses/electrical-engineering-and-computer-science/6-005-elements-of-software-construction-fall-2011/lecture-notes/
- [Puzzle Generation](http://zhangroup.aporc.org/images/files/Paper_3485.pdf)
- http://www.sudokuwiki.org/sudoku_creation_and_grading.pdf
- http://www.websudoku.com/faqs.php
- http://planetsudoku.com/how-to/sudoku-squirmbag.html
- http://www.sudoku-solutions.com/index.php?page=background
- https://gophers.slack.com/files/mem/F0DHMJBML/top95.txt
- http://www.websudoku.com/
- https://en.wikipedia.org/wiki/Exact_cover#Sudoku
