# Sudoku solver

You've found a Sudoku solver written in pure Go. Based on the [Golang Challenge #8](http://golang-challenge.com/go-challenge8/).

## How to get started?

The solver expects the Sudoku to be formatted like below:

##### input.txt

    1 _ 3 _ _ 6 _ 8 _
    _ 5 _ _ 8 _ 1 2 _
    7 _ 9 1 _ 3 _ 5 6
    _ 3 _ _ 6 7 _ 9 _
    5 _ 7 8 _ _ _ 3 _
    8 _ 1 _ 3 _ 5 _ 7
    _ 4 _ _ 7 8 _ 1 _
    6 _ 8 _ _ 2 _ 4 _
    _ 1 2 _ 4 5 _ 7 8 

It uses the incomming input from `stdin` in the terminal to process the Sudoku. Empty cells can be replaced by a `_`. Assuming you've saved the Sudoku in the same folder as `main.go` you can run:

    go run main.go < input.txt

Now you should see the solved Sudoku (if one exists):

    The Sudoku was solved successfully:
    
    1 2 3 4 5 6 7 8 9
    4 5 6 7 8 9 1 2 3
    7 8 9 1 2 3 4 5 6
    2 3 4 5 6 7 8 9 1
    5 6 7 8 9 1 2 3 4
    8 9 1 2 3 4 5 6 7
    3 4 5 6 7 8 9 1 2
    6 7 8 9 1 2 3 4 5
    9 1 2 3 4 5 6 7 8


## How does it work?

The solver uses [backtracking](https://en.wikipedia.org/wiki/Backtracking#Examples) and recursion to find a solution for your Sudoku. 

*Summarized:*

The solver tries to find the first empty cell (those with a `_`) and fills it with a valid digit that doesn't already occurs in the corresponding row, column or 3x3 section. In the following step we move to the next empty cell and insert another valid digit and so on. If we get stuck and tried all possible values for the current cell then we move back to the previous one (which is called as backtracking). Now we try there our luck with the next valid digit in this cell and move on. The board is finally solved if the programm was able to fill all cells with a valid digit.

Let's visualize this operations:

![Solving Sudokus using backtracking](https://upload.wikimedia.org/wikipedia/commons/8/8c/Sudoku_solved_by_bactracking.gif)

The visualization above is licensed under the [Creative Commons Attribution-Share Alike 3.0 Unported](https://creativecommons.org/licenses/by-sa/3.0/deed.en). Source: [Wikimedia](https://commons.wikimedia.org/wiki/File:Sudoku_solved_by_bactracking.gif)


## How fast the the Sudoku solver?

Thanks to Go, the solver runs very fast. But try it yourself by executing some benchmarks. Run:

    cd solver
    go test -bench=.


## License

The code is released under the GNU GPL V3. [Show me the license.](https://github.com/digitalcraftsman/sudoku/blob/master/LICENSE.md)