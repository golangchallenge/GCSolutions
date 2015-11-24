package sudoku

import (
	"bufio"
	"io"
)

var (
	maxRecursionDepth = -1
)

// SetRecursionDepth - set the maxium recursion depth for this solver
// defaults to unlimited
func SetRecursionDepth(d int) {
	maxRecursionDepth = d
}

// Puzzle - a sudoku puzzle structure
type Puzzle struct {
	p              [9][9]uint8
	recursionDepth int
}

// Dump - Dump the current state of the puzzle to a writer
func (p *Puzzle) Dump(writer io.Writer) {
	for _, v := range p.p {
		line := []byte{}
		for i, vv := range v {
			if i != 0 {
				line = append(line, space)
			}
			if vv == 0 {
				line = append(line, underscore)
				continue
			}
			line = append(line, vv+zero)
		}
		// write the line
		line = append(line, newline)
		writer.Write(line)
	}
}

// ParsePuzzle - take an io.Reader and deserialize into a Puzzle
func ParsePuzzle(reader io.Reader) (Puzzle, error) {
	p := Puzzle{}
	// use a scanner to validate, and parse input
	scanner := bufio.NewScanner(reader)
	// use a custom splitter, to break tokens into lines, and validate each line
	// for correctness
	scanner.Split(puzzleScanSplit)
	rowCount := 0
	// scan one row at a tim
	for scanner.Scan() {
		if rowCount > 8 {
			// we have exceeded the allowable number of rows, report invalid
			// row count
			return p, ErrParseInvalidRowCount
		}
		// grab the token bytes
		token := scanner.Bytes()
		posCount := 0
		for i := 0; i < len(token); i += 2 {
			// since we have already validated the correctness
			// of the puzzle input, we will skip to every other
			// value from the line
			var value uint8
			if token[i] != underscore {
				// if the value is not an underscore, set to
				// the number value of the ascii token
				value, _ = asciiToNumber(token[i])
			}
			// populate the value in the matrix
			p.p[rowCount][posCount] = value
			posCount++
		}
		rowCount++
	}

	if err := scanner.Err(); err != nil {
		// if there are errors, return the errors
		return p, err
	}

	if rowCount < 8 {
		// we have exceeded the allowable number of rows, report invalid
		// row count
		return p, ErrParseInvalidRowCount
	}

	return p, nil
}

// checkRow - Check that k isnt duplicated in the row i
func (p *Puzzle) checkRow(i int, k uint8) bool {
	for x := 0; x < 9; x++ {
		if p.p[i][x] == k {
			return false
		}
	}
	return true
}

// checkCol - Check that k isnt duplicated in the column j
func (p *Puzzle) checkCol(j int, k uint8) bool {
	for x := 0; x < 9; x++ {
		if p.p[x][j] == k {
			return false
		}
	}
	return true

}

// checkBox - given the 3x3 unit square i, j is a member of,
// check if k is allowed within this unit square
func (p *Puzzle) checkBox(i, j int, k uint8) bool {
	minX := 3 * int((i)/3)
	minY := 3 * int((j)/3)
	for x := minX; x < minX+3; x++ {
		for y := minY; y < minY+3; y++ {
			if p.p[x][y] == k {
				return false
			}
		}
	}
	return true
}

// allowed - check that k is allowed as a potential solution, check the row,
// column, and unit box for duplicates
func (p *Puzzle) allowed(i, j int, k uint8) bool {
	return p.checkRow(i, k) && p.checkCol(j, k) && p.checkBox(i, j, k)
}

// isSolved - validate there are no 0s left in the puzzle, that means
// everything is populated
func (p *Puzzle) isSolved() bool {
	for _, v := range p.p {
		for _, vv := range v {
			if vv == 0 {
				return false
			}
		}
	}
	return true
}

// BacktrackSolve - solve using backtrack algorithm
func (p *Puzzle) BacktrackSolve() error {
	if maxRecursionDepth != -1 && maxRecursionDepth < p.recursionDepth {
		return ErrSolveExceedRecursionDepth
	}
	p.recursionDepth++
	// iterating over the rows
	for i := range p.p {
		// iterating over the columns
		for j := range p.p[i] {
			// if this position is blank
			if p.p[i][j] == 0 {
				// to be solved, start at k=1 to k=9
				var k uint8 = 1
				for ; k < 10; k++ {
					// if k is allowed in this position, by checking row, column
					// and unit box for representation already.
					if p.allowed(i, j, k) {
						// copy the puzzle value to a tmp puzzle
						tmp := *p
						// insert the value into the tmp puzzle loacation
						tmp.p[i][j] = k
						// recursively call backtrack with tmp puzzle,
						// if solved, or nil error this is our solution
						err := tmp.BacktrackSolve()
						if tmp.isSolved() || err == nil {
							// overwrite p with tmp, and pop stack
							*p = tmp
							return nil
						}
						if err == ErrSolveExceedRecursionDepth {
							return err
						}
					}
				}
				// unfortunately nothing fit in this position
				return ErrSolveNoSolution
			}
		}
	}
	// should never get here.
	return ErrSolveNoSolution
}
