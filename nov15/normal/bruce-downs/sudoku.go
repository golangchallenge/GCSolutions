/*

This cli app provides a solver of sudoku puzzles
as defined at https://en.wikipedia.org/wiki/Sudoku.

Written by Bruce Downs <bruceadowns at gmail dot com>.

This is in response to the challenge posted
at http://golang-challenge.com/go-challenge8.

Usage:

$ sudoku --help
Usage of sudoku:
  -categorize
    	categorize the puzzle
  -dry
    	do not compute solution
  -verbose
    	emit verbose information

Examples for Linux and OSX:
$ cat input/official.txt | sudoku
$ cat input/official.txt | sudoku -categorize
$ cat input/official.txt | sudoku -categorize -verbose

Examples for Windows:
C:\>type input\official.txt | sudoku.exe
C:\>type input\official.txt | sudoku.exe -categorize
C:\>type input\official.txt | sudoku.exe -categorize -verbose

Where official.txt is:

1 _ 3 _ _ 6 _ 8 _
_ 5 _ _ 8 _ 1 2 _
7 _ 9 1 _ 3 _ 5 6
_ 3 _ _ 6 7 _ 9 _
5 _ 7 8 _ _ _ 3 _
8 _ 1 _ 3 _ 5 _ 7
_ 4 _ _ 7 8 _ 1 _
6 _ 8 _ _ 2 _ 4 _
_ 1 2 _ 4 5 _ 7 8

*/

package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"text/tabwriter"
)

const (
	// size is size of the square puzzle
	size = 9

	// boxSize is size of the inner box
	boxSize = size / 3

	// blank indicates an unknown space in the input
	blank = "_"
)

/*

Nomenclature:

* element - single element with 9 potential values 1-9
* row - row of elements
* column - column of elements
* box - 3x3 square of elements
* peers - generalized row or column or box
* peer group - an element's set of row, column, and box
* puzzle - 9x9 square of 81 elements, 9 rows, 9 columns, 9 boxes

*/

type element struct {
	elim  [size]bool
	value int
}

type puzzle [size][size]element

// init reads in 9 lines from reader and populates the 9x9 puzzle
func (puzz *puzzle) init(in io.Reader) error {
	// read all lines from reader
	var lines []string
	reader := bufio.NewReader(in)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		lines = append(lines, line)
	}
	if len(lines) != size {
		return fmt.Errorf("expect %d lines - actual %d", size, len(lines))
	}

	// populate puzzle from input lines
	for lineidx, line := range lines {
		fields := strings.Fields(line)
		if len(fields) != size {
			return fmt.Errorf("expect %d fields - actual %d [%s]", size, len(fields), line[:len(line)-1])
		}

		for fieldidx, field := range fields {
			switch field {
			//case "_", ".", "0":
			case "_":
				puzz[lineidx][fieldidx].value = 0
			case "1", "2", "3", "4", "5", "6", "7", "8", "9":
				i, _ := strconv.Atoi(field)
				puzz[lineidx][fieldidx].value = i
			default:
				return fmt.Errorf("expect '_123456789' - actual %s", field)
			}
		}
	}

	return nil
}

// count returns the element's potential value
// if all else are eliminated
func (e *element) potentialValue() (value int, ok bool) {
	var count int
	for k, v := range e.elim {
		if !v {
			count++
			value = k + 1
		}
	}
	if count == 1 {
		ok = true
	}

	return
}

// setEliminated eliminates an element's potential value
// and sets its value if only one potential value is left
func (puzz *puzzle) setEliminated(x, y, pos int) (value int, ok bool) {
	if puzz[x][y].elim[pos] {
		log.Fatalf("element value is already eliminated [%d,%d,%d]", x, y, pos)
	}

	puzz[x][y].elim[pos] = true

	if value, ok = puzz[x][y].potentialValue(); ok {
		puzz.setValue(x, y, value)
	}

	return
}

// setValue sets an element's value
// and eliminates the value from its peer group
func (puzz *puzzle) setValue(x, y, v int) {
	if puzz[x][y].value != 0 {
		log.Fatalf("element value is already set [%d,%d,%d]", x, y, v)
	}

	puzz[x][y].value = v

	// eliminate value from peer row
	for peer := 0; peer < size; peer++ {
		if puzz[peer][y].value == 0 {
			if !puzz[peer][y].elim[v-1] {
				puzz[peer][y].elim[v-1] = true
			}
		}
	}

	// eliminate value from peer column
	for peer := 0; peer < size; peer++ {
		if puzz[x][peer].value == 0 {
			if !puzz[x][peer].elim[v-1] {
				puzz[x][peer].elim[v-1] = true
			}
		}
	}

	// eliminate value from peer box
	topBoxRow := x - x%boxSize
	topBoxColumn := y - y%boxSize
	for peerRow := 0; peerRow < boxSize; peerRow++ {
		for peerColumn := 0; peerColumn < boxSize; peerColumn++ {
			if puzz[topBoxRow+peerRow][topBoxColumn+peerColumn].value == 0 {
				if !puzz[topBoxRow+peerRow][topBoxColumn+peerColumn].elim[v-1] {
					puzz[topBoxRow+peerRow][topBoxColumn+peerColumn].elim[v-1] = true
				}
			}
		}
	}

	// recursively set values for peer row
	for peer := 0; peer < size; peer++ {
		if puzz[peer][y].value == 0 {
			if value, ok := puzz[peer][y].potentialValue(); ok {
				puzz.setValue(peer, y, value)
			}
		}
	}

	// recursively set values for peer column
	for peer := 0; peer < size; peer++ {
		if puzz[x][peer].value == 0 {
			if value, ok := puzz[x][peer].potentialValue(); ok {
				puzz.setValue(x, peer, value)
			}
		}
	}

	// recursively set values for peer box
	for peerRow := 0; peerRow < boxSize; peerRow++ {
		for peerColumn := 0; peerColumn < boxSize; peerColumn++ {
			if puzz[topBoxRow+peerRow][topBoxColumn+peerColumn].value == 0 {
				if value, ok := puzz[topBoxRow+peerRow][topBoxColumn+peerColumn].potentialValue(); ok {
					puzz.setValue(topBoxRow+peerRow, topBoxColumn+peerColumn, value)
				}
			}
		}
	}
}

// solved returns whether we have found
// a valid solution for the puzzle
func (puzz *puzzle) solved() error {
	// check that all elements have values
	// short-circuit return on first unset value
	for row := 0; row < size; row++ {
		for column := 0; column < size; column++ {
			if puzz[row][column].value == 0 {
				return fmt.Errorf("value not set [%d,%d]", row, column)
			}
		}
	}

	var res []string

	// check that each row has unique values
	for row := 0; row < size; row++ {
		var visited [size]bool
		for peer := 0; peer < size; peer++ {
			if visited[puzz[row][peer].value-1] {
				res = append(res, fmt.Sprintf("already visited [%d,%d,%d] in row",
					row, peer, puzz[row][peer].value))
			}
			visited[puzz[row][peer].value-1] = true
		}
	}

	// check that each column has unique values
	for column := 0; column < size; column++ {
		var visited [size]bool
		for peer := 0; peer < size; peer++ {
			if visited[puzz[peer][column].value-1] {
				res = append(res, fmt.Sprintf("already visited [%d,%d,%d] in column",
					peer, column, puzz[peer][column].value))
			}
			visited[puzz[peer][column].value-1] = true
		}
	}

	// check that each box has unique values
	for topBoxRow := 0; topBoxRow < size; topBoxRow += boxSize {
		for topBoxColumn := 0; topBoxColumn < size; topBoxColumn += boxSize {
			var visited [size]bool
			for peerRow := 0; peerRow < boxSize; peerRow++ {
				for peerColumn := 0; peerColumn < boxSize; peerColumn++ {
					if visited[puzz[topBoxRow+peerRow][topBoxColumn+peerColumn].value-1] {
						res = append(res, fmt.Sprintf("already visited [%d,%d,%d] in box",
							topBoxRow+peerRow, topBoxColumn+peerColumn, puzz[topBoxRow+peerRow][topBoxColumn+peerColumn].value))
					}
					visited[puzz[topBoxRow+peerRow][topBoxColumn+peerColumn].value-1] = true
				}
			}
		}
	}

	if len(res) > 0 {
		return fmt.Errorf(strings.Join(res, "; "))
	}

	return nil
}

// printme prints the puzzle using the given writer
func (puzz *puzzle) printme(out io.Writer) {
	tw := tabwriter.NewWriter(out, 0, 0, 1, ' ', 0)

	for rowidx, row := range puzz {
		for elemidx, elem := range row {
			if elem.value == 0 {
				var count int
				for k, v := range elem.elim {
					if !v {
						fmt.Fprintf(tw, fmt.Sprintf("%d", k+1))
						count++
					}
				}
				if count == 1 {
					log.Fatalf("unexpected elim count equal to 1 [%d,%d]", rowidx, elemidx)
				}
			} else {
				fmt.Fprintf(tw, fmt.Sprintf("%d", elem.value))
			}

			if elemidx < 8 {
				fmt.Fprintf(tw, "\t")
			}
		}

		fmt.Fprintf(tw, "\n")
	}

	tw.Flush()
}

// solve handles calling each solving algorithm
func solve(puzz puzzle, vout io.Writer) (puzzRes puzzle, category string) {
	for {
		puzzRes = sweepEliminate(puzz, vout)
		if puzzRes.solved() == nil {
			category = "easy"
			break
		}
		puzzRes.printme(vout)

		puzzRes = propagateUnique(puzzRes, vout)
		if puzzRes.solved() == nil {
			category = "medium"
			break
		}
		puzzRes.printme(vout)

		puzzRes = guessBacktrack(puzzRes, vout)
		if puzzRes.solved() == nil {
			category = "hard"
			break
		}
		puzzRes.printme(vout)

		category = "unsolvable"
		break
	}

	return
}

// sweepEliminate cycles through each element
// and eliminates known values
func sweepEliminate(puzz puzzle, vout io.Writer) puzzle {
	fmt.Fprintln(vout, "sweep and eliminate")

	for row := 0; row < size; row++ {
	OUTER_LOOP:
		for column := 0; column < size; column++ {
			if puzz[row][column].value != 0 {
				continue
			}

			// propagate row peers
			for peer := 0; peer < size; peer++ {
				if peer != column {
					if puzz[row][peer].value != 0 {
						if !puzz[row][column].elim[puzz[row][peer].value-1] {
							if _, ok := puzz.setEliminated(row, column, puzz[row][peer].value-1); ok {
								continue OUTER_LOOP
							}
						}
					}
				}
			}

			// propagate column peers
			for peer := 0; peer < size; peer++ {
				if peer != row {
					if puzz[peer][column].value != 0 {
						if !puzz[row][column].elim[puzz[peer][column].value-1] {
							if _, ok := puzz.setEliminated(row, column, puzz[peer][column].value-1); ok {
								continue OUTER_LOOP
							}
						}
					}
				}
			}

			// propagate box peers
			topBoxRow := row - row%boxSize
			topBoxColumn := column - column%boxSize
			for peerRow := 0; peerRow < boxSize; peerRow++ {
				for peerColumn := 0; peerColumn < boxSize; peerColumn++ {
					if topBoxRow+peerRow != row && topBoxColumn+peerColumn != column {
						if puzz[topBoxRow+peerRow][topBoxColumn+peerColumn].value != 0 {
							if !puzz[row][column].elim[puzz[topBoxRow+peerRow][topBoxColumn+peerColumn].value-1] {
								if _, ok := puzz.setEliminated(row, column, puzz[topBoxRow+peerRow][topBoxColumn+peerColumn].value-1); ok {
									continue OUTER_LOOP
								}
							}
						}
					}
				}
			}
		}
	}

	return puzz
}

// propagateUnique sets an elements value
// if it has a unique potential value
// within a row,column,box
func propagateUnique(puzz puzzle, vout io.Writer) puzzle {
	fmt.Fprintln(vout, "propagate unique values")

	// propagate unique potential values for rows
	for row := 0; row < size; row++ {
		var rowPotentials [size]struct {
			count int
			last  int
		}
		for peer := 0; peer < size; peer++ {
			if puzz[row][peer].value == 0 {
				for k, v := range puzz[row][peer].elim {
					if puzz[row][k].value == 0 && !v {
						rowPotentials[k].count++
						rowPotentials[k].last = peer
					}
				}
			}
		}
		for k, v := range rowPotentials {
			if puzz[row][v.last].value != 0 && v.count == 1 {
				puzz.setValue(row, v.last, k+1)
			}
		}
	}

	// check that each column has unique values
	for column := 0; column < size; column++ {
		// propagate unique potential values for column
		var columnPotentials [size]struct {
			count int
			last  int
		}
		for peer := 0; peer < size; peer++ {
			if puzz[peer][column].value == 0 {
				for k, v := range puzz[peer][column].elim {
					if !v {
						columnPotentials[k].count++
						columnPotentials[k].last = peer
					}
				}
			}
		}
		for k, v := range columnPotentials {
			if puzz[v.last][column].value != 0 && v.count == 1 {
				puzz.setValue(v.last, column, k+1)
			}
		}
	}

	// check that each box has unique values
	for topBoxRow := 0; topBoxRow < size; topBoxRow += boxSize {
		for topBoxColumn := 0; topBoxColumn < size; topBoxColumn += boxSize {
			// propagate unique potential values for box
			var boxPotentials [size]struct {
				count      int
				lastRow    int
				lastColumn int
			}
			for peerRow := 0; peerRow < boxSize; peerRow++ {
				for peerColumn := 0; peerColumn < boxSize; peerColumn++ {
					if puzz[topBoxRow+peerRow][topBoxColumn+peerColumn].value == 0 {
						for k, v := range puzz[topBoxRow+peerRow][topBoxColumn+peerColumn].elim {
							if !v {
								boxPotentials[k].count++
								boxPotentials[k].lastRow = topBoxRow + peerRow
								boxPotentials[k].lastColumn = topBoxColumn + peerColumn
							}
						}
					}
				}
			}
			for k, v := range boxPotentials {
				if puzz[v.lastRow][v.lastColumn].value == 0 && v.count == 1 {
					puzz.setValue(v.lastRow, v.lastColumn, k+1)
				}
			}
		}
	}

	return puzz
}

// guessBacktrack recursively guesses a solution
func guessBacktrack(puzz puzzle, vout io.Writer) puzzle {
	fmt.Fprintln(vout, "guess and backtrack")

	if puzz.solved() == nil {
		log.Fatal("guess and backtrack already solved")
	}

	// find element with least amount of potential values
	// short-circuit if potential values equals 2
	least, guessX, guessY :=
		math.MaxInt32, math.MaxInt32, math.MaxInt32
OUTER_LOOP:
	for row := 0; row < size; row++ {
		for column := 0; column < size; column++ {
			if puzz[row][column].value == 0 {
				var count int
				for _, v := range puzz[row][column].elim {
					if !v {
						count++
					}
				}
				if count < least {
					least, guessX, guessY = count, row, column
				}
				if count == 2 {
					break OUTER_LOOP
				}
			}
		}
	}

	if least > size || guessX > size || guessY > size {
		fmt.Fprintf(vout, "invalid guess/solution [%d,%d,%d]",
			least, guessX, guessY)
		return puzz
	}

	// set its value successively and recursively attempt to solve
	for k, v := range puzz[guessX][guessY].elim {
		if !v {
			// copy puzzle and guess
			puzzGuess := puzz
			guessV := k + 1
			fmt.Fprintf(vout, "guess [%d,%d,%d]\n",
				guessX, guessY, guessV)

			puzzGuess.setValue(guessX, guessY, guessV)
			if puzzGuess.solved() == nil {
				return puzzGuess
			}
			puzzGuess.printme(vout)

			puzzGuess = propagateUnique(puzzGuess, vout)
			if puzzGuess.solved() == nil {
				return puzzGuess
			}
			puzzGuess.printme(vout)

			puzzGuess = guessBacktrack(puzzGuess, vout)
			if puzzGuess.solved() == nil {
				return puzzGuess
			}
			puzzGuess.printme(vout)

			fmt.Fprintf(vout, "guess was incorrect [%d,%d,%d]\n",
				guessX, guessY, guessV)
		}
	}

	return puzz
}

// main is the entry point to the cli application
func main() {
	var categorize, dry, verbose bool
	flag.BoolVar(&categorize, "categorize", false, "categorize the puzzle")
	flag.BoolVar(&dry, "dry", false, "do not compute solution")
	flag.BoolVar(&verbose, "verbose", false, "emit verbose information")
	flag.Parse()

	var puzz puzzle

	if err := puzz.init(os.Stdin); err != nil {
		log.Fatal(err)
	}

	if !dry {
		vout := ioutil.Discard
		if verbose {
			vout = os.Stdout
		}

		var category string
		puzz, category = solve(puzz, vout)

		if err := puzz.solved(); err != nil {
			puzz.printme(vout)
			log.Fatal(err)
		}

		if categorize {
			fmt.Fprintf(os.Stdout, "Puzzle is %s\n", category)
		}
	}

	puzz.printme(os.Stdout)
}
