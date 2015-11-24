package sudoku

const (
	dataLen = 9*9*2 - 1

	allPossible = 0x01FF

	emptyCell = '_'
)

type puzzleRow struct {
	cells [9]byte
}

func (p *Puzzle) setValue(row, col int, val byte) {
	p.rows[row].cells[col] = 0x30 | val
	disable := uint16(1 << uint8(val-1))
	p.rowsAvail[row] ^= disable
	p.colsAvail[col] ^= disable
	p.blockAvail[3*(row/3)+col/3] ^= disable
}

func (p *Puzzle) setAndCheck(j, i int, val byte) (updated bool) {

	j3, id3 := 3*(j/3), i/3
	di, dj := 3*((j3+id3)%3), 3*((j3+id3)/3)

	p.setValue(j, i, val)

	row := p.rows[j].cells[:]

	p.loopCount += 27
	for n := range row {
		nd3 := n / 3
		if row[n] == emptyCell {
			block := j3 + nd3
			avail := p.rowsAvail[j] &
				p.colsAvail[n] &
				p.blockAvail[block]

			if popCount9Bit(avail) == 1 {
				row[n] = ctz16Bits(avail<<1) |
					0x30
				p.rowsAvail[j] ^= avail
				p.colsAvail[n] ^= avail
				p.blockAvail[block] ^= avail
				updated = true
			}
		}

		if p.rows[n].cells[i] == emptyCell {
			block := 3*nd3 + id3
			avail := p.rowsAvail[n] &
				p.colsAvail[i] &
				p.blockAvail[block]

			if popCount9Bit(avail) == 1 {
				p.rows[n].cells[i] =
					ctz16Bits(avail<<1) |
						0x30
				p.rowsAvail[n] ^= avail
				p.colsAvail[i] ^= avail
				p.blockAvail[block] ^= avail
				updated = true
			}
		}

		nr, nc := dj+nd3, di+n%3
		if p.rows[nr].cells[nc] == emptyCell {
			avail := p.rowsAvail[nr] &
				p.colsAvail[nc] &
				p.blockAvail[j3+id3]
			if popCount9Bit(avail) == 1 {
				p.rows[nr].cells[nc] =
					ctz16Bits(avail<<1) |
						0x30
				p.rowsAvail[nr] ^= avail
				p.colsAvail[nc] ^= avail
				p.blockAvail[j3+id3] ^= avail
				updated = true
			}
		}
	}

	return
}

func (p *Puzzle) suggestSolution(j, i int, val byte) {
	if p.setAndCheck(j, i, val) {
		p.solve()
	} else {
		p.tryBruteForce()
	}
}

func (p *Puzzle) tryBruteForce() {
	mcount := byte(10)
	var mavail uint16
	var mx int
	var my int

FindMinAvail:
	for j := range p.rows {
		var ravail uint16
		if ravail = p.rowsAvail[j]; ravail == 0 {
			continue
		}

		row := p.rows[j].cells[:]
		j3 := 3 * (j / 3)

		for i := range row {
			p.loopCount++
			if row[i] != emptyCell {
				continue
			}
			avail := ravail & p.colsAvail[i] & p.blockAvail[j3+i/3]

			if avail == 0 {
				p.err = ErrInvalidPuzzle
				return
			}

			pc := popCount9Bit(avail)
			if pc < mcount {
				mcount = pc
				mavail = avail
				my, mx = j, i
				if pc < 3 {
					break FindMinAvail
				}
			}
		}
	}

	var result *Puzzle
	ctz := ctz16Bits(mavail)
	mavail >>= ctz
	ctz++

	// suggest solutions
	for result1 := new(Puzzle); mavail != 0; mavail, ctz = mavail>>1, ctz+1 {
		if mavail&1 != 1 {
			continue
		}

		*result1 = *p
		result1.suggestSolution(my, mx, ctz)
		if result1.err == ErrMultipleSolutions {
			*p = *result1
			return
		} else if result1.err == nil &&
			result != nil {
			*p = *result
			p.err = ErrMultipleSolutions
			return
		} else if result1.err == nil {
			result = result1
			result1 = new(Puzzle)
		}
	}

	if result != nil {
		result.loopCount += p.loopCount
		*p = *result
	}
}

func (p *Puzzle) solve() {
	p.simpleSolve()

	if p.err == nil || (p.err != nil && p.err != errSolutionNotFound) {
		return
	}
	p.tryBruteForce()
}

var (
	pc4Bits = []byte{0, 1, 1, 2, 1, 2, 2, 3, 1, 2, 2, 3, 2, 3, 3, 4}
)

func popCount9Bit(x uint16) byte {
	//return byte((uint64(x) * 0x200040008001 & 0x111111111111111) % 0xF)
	return pc4Bits[x&0xF] + pc4Bits[(x>>4)&0xF] + byte(x>>8)
}

func ctz16Bits(x uint16) (ctz byte) {
	if x&1 == 0 {
		ctz = 1
		if x&0xff == 0 {
			x >>= 8
			ctz = 9

		} else {
			ctz = 1
		}
		if x&0xf == 0 {
			x >>= 4
			ctz += 4
		}

		if x&0x3 == 0 {
			x >>= 2
			ctz += 2
		}
		ctz -= byte(x & 1)
	}
	return
}

func (p *Puzzle) simpleSolve() {

	for {
		p.loopCount += 81
		updated := false

		for j := range p.rows {
			if p.rowsAvail[j] == 0 {
				continue
			}

			row := p.rows[j].cells[:]
			j3 := 3 * (j / 3)
			ravail := p.rowsAvail[j]

			for i3, i := 0, 0; i3 < 3; i3++ {
				block := j3 + i3
				avail0 := ravail & p.blockAvail[block]

				for n := 0; n < 3; n, i = n+1, i+1 {
					if row[i] != emptyCell {
						continue
					}
					avail := avail0 & p.colsAvail[i]
					p.err = errSolutionNotFound

					if pc := popCount9Bit(avail); pc == 1 {
						p.setValue(j, i, ctz16Bits(avail<<1))
						ravail = p.rowsAvail[j]
						avail0 = ravail & p.blockAvail[block]
						updated = true
					}
				}
			}
		}
		if !updated {
			return
		}
		p.err = nil
	}
}
