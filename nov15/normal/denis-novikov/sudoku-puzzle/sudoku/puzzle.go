package sudoku

import (
	"errors"
	"io"
)

var (
	// ErrInvalidPuzzle is returned when a sudoku puzzle cannot be solved.
	ErrInvalidPuzzle = errors.New("Invalid Puzzle")
	// ErrMultipleSolutions is returned when a sudoku puzzle has more than
	// one solutions.
	ErrMultipleSolutions = errors.New("Multiple Solutions Found")
	errSolutionNotFound  = errors.New("Solution Not Found")
)

// Puzzle is a struct that contains a sudoku puzzle internals. It should be
// created either with the help of function `NewPuzzleFromReader` or with
// the help of function `Generate`.
type Puzzle struct {
	rowsAvail  [9]uint16
	blockAvail [9]uint16
	colsAvail  [9]uint16
	rows       [9]puzzleRow
	emptyCells byte
	lbRow      byte
	lbCol      byte
	lbBlock    byte
	loopCount  uint
	err        error
}

// Difficulty returns a difficulty level for a solved puzzle.
func (p *Puzzle) Difficulty() DifficultyLevel {
	return p.getDifficulty()
}

// NewPuzzleFromReader returns a puzzle from an io.Reader.
func NewPuzzleFromReader(r io.Reader) (*Puzzle, error) {
	p := Puzzle{lbRow: 9, lbCol: 9, lbBlock: 9}
	for i := range p.rows {
		p.rowsAvail[i] = allPossible
		p.blockAvail[i] = allPossible
		p.colsAvail[i] = allPossible
	}
	buf := make([]byte, dataLen)
	n, err := r.Read(buf)
	if err != nil {
		return nil, err
	}

	if n != len(buf) {
		return nil, io.EOF
	}

	for j := range p.rows {
		row := &p.rows[j]

		j18 := 18 * j
		j3 := 3 * (j / 3)
		for i := range row.cells {
			row.cells[i] = buf[j18+(i<<1)]
			if row.cells[i] >= '1' && row.cells[i] <= '9' {
				pos := uint16(1 << (row.cells[i] - '1'))
				p.rowsAvail[j] ^= pos
				p.colsAvail[i] ^= pos
				p.blockAvail[j3+i/3] ^= pos
			} else if row.cells[i] != emptyCell {
				return nil, ErrInvalidPuzzle
			} else {
				p.emptyCells++
			}
		}
	}

	if !p.IsValid() {
		return nil, ErrInvalidPuzzle
	}

	p.calcAvails()

	return &p, nil
}

// String converts puzzle to a string
func (p *Puzzle) String() string {
	buf := make([]byte, dataLen+1)
	for j := range p.rows {
		row := &p.rows[j]
		for i := range row.cells {
			n := 2 * (i + (9 * j))
			buf[n], buf[n+1] = row.cells[i], ' '
		}
		buf[2*(8+9*j)+1] = '\n'
	}
	// trim last new line symbol
	return string(buf[:len(buf)-1])
}

func (r *puzzleRow) check() bool {
	nums := 0
	for _, cell := range r.cells {
		if cell == emptyCell {
			continue
		}
		if cell < '1' || cell > '9' {
			return false
		}
		flag := 1 << (cell - '1')
		if nums&flag != 0 {
			return false
		}
		nums |= flag
	}
	return true
}

func (p *Puzzle) blockAsRow(n int, r *puzzleRow) *puzzleRow {
	di, dj := 3*(n%3), 3*(n/3)
	for i := range p.rows {
		cell := &r.cells[i]
		*cell = p.rows[dj+i/3].cells[di+i%3]
	}
	return r
}

func (p *Puzzle) columnAsRow(n int, r *puzzleRow) *puzzleRow {
	for i := range p.rows {
		cell := &r.cells[i]
		*cell = p.rows[i].cells[n]
	}
	return r
}

// IsValid returns true when a puzzle passes validation.
func (p *Puzzle) IsValid() bool {
	r := &puzzleRow{}
	for i := range p.rows {
		if !p.rows[i].check() || !p.columnAsRow(i, r).check() {
			p.err = ErrInvalidPuzzle
			return false
		}

		if !p.blockAsRow(i, r).check() {
			p.err = ErrInvalidPuzzle
			return false
		}
	}

	return true
}

// Solve returns a pointer to a solution. The original puzzle p is not modified.
func (p Puzzle) Solve() (*Puzzle, error) {
	if p.err != nil {
		return nil, p.err
	}

	p.solve()
	return &p, p.err
}

// IsSolved checks if a sudoku puzzle is solved.
func (p *Puzzle) IsSolved() bool {
	if p.err != nil {
		return false
	}
	p.err = nil
	for i := range p.rows {
		if p.rowsAvail[i]+p.colsAvail[i]+p.blockAvail[i] != 0 {
			return false
		}
	}
	return true
}
