package main

import "fmt"

func (b *board) Validate() error {
	for pos := 0; pos < 81; pos++ {
		var blit uint

		// validate row
		rowVals := make([]uint, 9)
		startRow := (pos / 9) * 9
		blit = 0
		for r := startRow; r < startRow+9; r++ {
			rowVals[r-startRow] = b.solved[r]
			blit |= b.blits[r]
		}
		if err := validate(rowVals); err != nil {
			return NewErrUnsolvable("row error %#2v %s", getCoords(pos), err.Error())
		}
		if !b.loading && blit != 0x1FF {
			return NewErrUnsolvable("row missing hint %#2v %09b", getCoords(pos), blit)
		}

		// validate column
		colVals := make([]uint, 9)
		colIndex := 0
		blit = 0
		for c := pos % 9; c < 81; c += 9 {
			colVals[colIndex] = b.solved[c]
			colIndex++
			blit |= b.blits[c]
		}
		if err := validate(colVals); err != nil {
			return NewErrUnsolvable("col error %#2v %s", getCoords(pos), err.Error())
		}
		if !b.loading && blit != 0x1FF {
			return NewErrUnsolvable("col missing hint %#2v %09b", getCoords(pos), blit)
		}

		// validate box
		startRow = ((pos / 9) / 3) * 3
		startCol := ((pos % 9) / 3) * 3
		boxVals := make([]uint, 9)
		boxIndex := 0
		blit = 0
		for r := startRow; r < startRow+3; r++ {
			for c := startCol; c < startCol+3; c++ {
				boxVals[boxIndex] = b.solved[r*9+c]
				boxIndex++
				blit |= b.blits[r*9+c]
			}
		}
		if err := validate(boxVals); err != nil {
			return NewErrUnsolvable("box error %#2v %s", getCoords(pos), err.Error())
		}
		if !b.loading && blit != 0x1FF {
			return NewErrUnsolvable("box missing hint %#2v %09b", getCoords(pos), blit)
		}
	}

	return nil
	//	return b.ValidateKnownAnswer()
}

/*func (b *board) ValidateKnownAnswer() error {
	if b.knownAnswer == nil {
		return nil
	}

	for pos := 0; pos < 81; pos++ {
		ka := uint(b.knownAnswer[pos])
		if b.solved[pos] != 0 {
			if b.solved[pos] != ka {
				return fmt.Errorf("pos %#2v solved with %d, known answer is %d", getCoords(pos), b.solved[pos], ka)
			}
		} else {
			blit := b.blits[pos]
			kaMask := uint(1 << (ka - 1))

			if blit&kaMask != kaMask {
				return fmt.Errorf("pos %#2v missing known answer %d as hint", getCoords(pos), ka)
			}
		}
	}

	return nil
}*/

func validate(vals []uint) error {
	if len(vals) != 9 {
		return fmt.Errorf("len(vals), expected: 9, actual = %d", len(vals))
	}

	avail := 0x1FF
	for _, v := range vals {
		if v == 0 {
			continue
		}

		mask := 1 << (v - 1)

		if avail&mask != mask {
			return NewErrUnsolvable("val %d repeated", v)
		}
		avail &= ^mask
	}
	return nil
}
