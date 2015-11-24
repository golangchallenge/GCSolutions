This respository is my submission to go-challenge8.

It is efficient sudoku solver based on Dancing Links Algorithm by Donald Knuth.


Running application
-------------------
```
go build
go-challenge8 < input/sample
```
To run your own sudoku please follow the same format as in sample file.
```
_ 3 6 _ _ _ _ _ 8
2 _ 7 5 _ _ 4 9 _
_ _ 4 _ _ 8 3 6 _
_ 2 _ 6 _ _ _ _ 3
4 9 3 _ _ _ 6 8 5
6 _ _ _ _ 5 _ 2 _
_ 7 9 4 _ _ 1 _ _
_ 6 2 _ _ 1 8 _ 9
3 _ _ _ _ _ 2 5 _
```
Benchmarking solver
-------------------
```
cd sudoku/
go test -test.bench=BenchmarkSolve -test.benchtime 10s -cpuprofile cpu.out
```

For my Lenovo Yoga (Intel i5-5300U 2.30GHz) results are as follows:
```
PASS
BenchmarkSolve-2	   30000	    574000 ns/op
ok  	bitbucket.org/jrozansk/go-challenge8/sudoku	23.043s
```