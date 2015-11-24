# Sudoku

To run solver

    go run main.go sudoku.go < input/easy.in

Other puzzles that can be tested - easy.in, medium.in, hard.in, evil.in

To run tests (Test over 100 hard problems)
    
    go build
    go test

To run benchmarks
    
    go test -run=XXX -bench=.
