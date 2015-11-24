# Go Challenge - 8

http://golang-challenge.com/go-challenge8/

## The Goal of the challenge

The goal of this challenge is to implement a Sudoku solver.

## Requirements of the challenge

Your program should read a puzzle of this form from standard input:

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

And it should write the solution to standard output:

```
1 2 3 4 5 6 7 8 9
4 5 6 7 8 9 1 2 3
7 8 9 1 2 3 4 5 6
2 3 4 5 6 7 8 9 1
5 6 7 8 9 1 2 3 4
8 9 1 2 3 4 5 6 7
3 4 5 6 7 8 9 1 2
6 7 8 9 1 2 3 4 5
9 1 2 3 4 5 6 7 8
```

It should reject malformed or invalid inputs and recognize and report puzzles
that cannot be solved.

(Incidentally, the puzzle above makes a nice test case, because the solution is
easy to validate by sight.)

## Bonus features

* Print a rating of the puzzle's difficulty (easy, medium, hard). This rating
  should roughly coincide with the ratings of shown by sites like Web Sudoku.
* Implement a puzzle generator that produces a puzzle of the given difficulty
  rating.
* Maximize the efficiency of your program. (Write a benchmark.)
* Write test cases, and use the cover tool to make sure your tests are
  thorough. (I found a bug in both my implementation and a test case when I
  checked the coverage.)
* Use a non-obvious technique, like Knuth's "Dancing Links" or something of your
  own invention.

## Hints

* For an elegant and efficient representation of the puzzle, try using an array
  (not a slice).
* Recursion can dramatically simplify your implementation.
