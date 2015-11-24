package sudoku

import (
	"math/rand"
	"time"
)

// IntGenerator interface for random integer generation.
type IntGenerator interface {
	// Intn returns a non-negative integer in the range [0;n).
	Intn(int) int
	// Int31n returns a non-negative 32-bit integer in the range [0;n).
	Int31n(int32) int32
}

// Generate creates a sudoku puzzle with a difficulty level d.
func Generate(d DifficultyLevel) *Puzzle {
	p := Puzzle{lbRow: 9, lbCol: 9, lbBlock: 9}
	src := rand.NewSource(time.Now().UnixNano())
	p.generate(rand.New(src), d)
	return &p
}

func (p *Puzzle) generate(rnd IntGenerator, d DifficultyLevel) {
	for i := range p.rows {
		p.rowsAvail[i] = allPossible
		p.blockAvail[i] = allPossible
		p.colsAvail[i] = allPossible
	}

	for j := range p.rows {
		row := p.rows[j].cells[:]
		for i := range row {
			row[i] = emptyCell
		}
	}

	for n := 0; n < 9; n++ {
		i := rnd.Intn(9)
		for foundAvail1 := false; !foundAvail1; {
			if foundAvail1 = p.getAvail(n, i)&1 != 0; !foundAvail1 {
				i++
				i %= 9
			}
		}
		p.setValue(n, i, 1)

		for k := n + 1; k < 9; k++ {
			p.setValue(k, i, byte(1+k-n))
		}

		for k := 0; k < n; k++ {
			p.setValue(k, i, byte(10+k-n))
		}
	}

	swaps := [...]swap{
		p.swapRows, p.swapColumns, p.swapBlockRows, p.swapBlockColumns}
	p.roll(rnd.Int31n(4))

	swapNum := 200 + rnd.Int31n(200)
	for ; swapNum != 0; swapNum-- {
		swaps[rnd.Intn(len(swaps))](rnd.Int31n(3))
	}

	p.err = nil
	p.emptyCells = 0
	p.lbRow = 9
	p.lbCol = 9
	p.lbRow = 9

	sx := new(Puzzle)
	// Unique
	removeLimit := 29
	for remove := 0; remove < removeLimit; {
		var i int
		var j int
		for foundNonEmpty := false; !foundNonEmpty; {
			j, i := int(rnd.Intn(5)), int(rnd.Intn(5))
			foundNonEmpty = p.rows[j].cells[i] != emptyCell
		}

		*sx = *p
		sx.symmetricUnset(j, i)

		if sx.solve(); sx.err == nil {
			count := p.symmetricUnset(j, i)
			remove += count
			p.emptyCells += byte(count)
		}
	}

	p.loopCount = 0
	p.calcAvails()

	if d < DLMedium {
		return
	}

	var unsettable [81]bool
	for i := range unsettable {
		unsettable[i] = p.rows[i/9].cells[i%9] == emptyCell
	}

	*sx = *p
	sx.solve()
	//SetDifficulty:
	for sx.getDifficulty() < d {
		foundNonUnsettable := false
		i := rnd.Intn(81)
		for ; i < 81; i++ {
			foundNonUnsettable = !unsettable[i]
			if foundNonUnsettable {
				break
			}
		}

		if !foundNonUnsettable {
			for i = 0; i < 81; i++ {
				foundNonUnsettable = !unsettable[i]
				if foundNonUnsettable {
					break
				}
			}
		}
		if !foundNonUnsettable {
			p.loopCount = 0
			p.calcAvails()
			return
		}

		*sx = *p
		sx.unsetValue(i/9, i%9)
		unsettable[i] = true
		sx.solve()
		if sx.err == nil {
			p.unsetValue(i/9, i%9)
		}
	}

	p.loopCount = 0
	p.calcAvails()
}

func (p *Puzzle) calcAvails() {
	for j := range p.rows {
		lb := 9 - popCount9Bit(p.rowsAvail[j])
		if lb < p.lbRow {
			p.lbRow = lb
		}
		lb = 9 - popCount9Bit(p.colsAvail[j])
		if lb < p.lbCol {
			p.lbCol = lb
		}
		lb = 9 - popCount9Bit(p.blockAvail[j])
		if lb < p.lbBlock {
			p.lbBlock = lb
		}
	}

}

func (p *Puzzle) symmetricUnset(j, i int) (count int) {
	p.unsetValue(j, i)
	count++

	if i != j {
		p.unsetValue(i, j)
		count++

		if i != 4 {
			p.unsetValue(j, 8-i)
			p.unsetValue(8-i, j)
			count += 2

			if j != 4 {
				p.unsetValue(8-j, 8-i)
				p.unsetValue(8-i, 8-j)
				count += 2
			}
		}

		if j != 4 {
			p.unsetValue(8-j, i)
			p.unsetValue(i, 8-j)
			count += 2
		}
	} else if i != 4 {
		p.unsetValue(8-i, 8-i)
		p.unsetValue(i, 8-i)
		p.unsetValue(8-i, i)
		count += 3
	}
	return
}

type swap func(int32)

func (p *Puzzle) swapRows(n int32) {
	r1 := n >> 1
	r2 := 1 + ((1 + n) >> 1)
	p.rows[r1], p.rows[r2] = p.rows[r2], p.rows[r1]
	p.rowsAvail[r1], p.rowsAvail[r2] = p.rowsAvail[r2], p.rowsAvail[r1]
}

func (p *Puzzle) swapColumns(n int32) {
	c1 := n >> 1
	c2 := 1 + ((1 + n) >> 1)

	for j := range p.rows {
		p.rows[j].cells[c1], p.rows[j].cells[c2] =
			p.rows[j].cells[c2], p.rows[j].cells[c1]
	}
	p.colsAvail[c1], p.colsAvail[c2] = p.colsAvail[c2], p.colsAvail[c1]
}

func (p *Puzzle) swapBlockRows(n int32) {
	r1 := 3 * (n >> 1)
	r2 := 3 * (1 + ((1 + n) >> 1))

	for j := 0; j < 3; j++ {
		p.rows[r1], p.rows[r2] = p.rows[r2], p.rows[r1]
		p.rowsAvail[r1], p.rowsAvail[r2] =
			p.rowsAvail[r2], p.rowsAvail[r1]
		r1++
		r2++
	}
}

func (p *Puzzle) swapBlockColumns(n int32) {
	c1 := 3 * (n >> 1)
	c2 := 3 * (1 + ((1 + n) >> 1))

	for i := 0; i < 3; i++ {
		for j := range p.rows {
			p.rows[j].cells[c1], p.rows[j].cells[c2] =
				p.rows[j].cells[c2], p.rows[j].cells[c1]
		}
		p.colsAvail[c1], p.colsAvail[c2] =
			p.colsAvail[c2], p.colsAvail[c1]
		c1++
		c2++
	}
}

func (p *Puzzle) roll(n int32) {
	for n++; n > 0; n-- {
		for j := range p.rows {
			row1 := p.rows[j].cells[:]
			row2 := p.rows[8-j].cells[:]

			for i := range row1 {
				row1[i], row2[i], row1[8-i], row2[8-i] =
					row2[i], row1[8-i], row2[8-i], row1[i]
			}

			p.rowsAvail[j], p.colsAvail[8-j] =
				p.colsAvail[8-j], p.rowsAvail[j]
		}
	}
}

func (p *Puzzle) getAvail(row, col int) uint16 {
	return p.rowsAvail[row] & p.colsAvail[col] &
		p.blockAvail[3*(row/3)+col/3]
}

func (p *Puzzle) unsetValue(row, col int) {
	val := uint8(0xF & p.rows[row].cells[col])
	p.rows[row].cells[col] = emptyCell
	block := 3*(row/3) + col/3

	enable := uint16(1 << (val - 1))
	p.rowsAvail[row] |= enable
	p.colsAvail[col] |= enable
	p.blockAvail[block] |= enable

	pc := popCount9Bit(p.rowsAvail[row])
	if pc < p.lbRow {
		p.lbRow = pc
	}
	if pc = popCount9Bit(p.colsAvail[col]); pc < p.lbCol {
		p.lbCol = pc
	}
	if pc = popCount9Bit(p.blockAvail[block]); pc < p.lbBlock {
		p.lbBlock = pc
	}
}

func init() {
	seed := time.Now().UnixNano()
	rand.Seed(seed)
}
