package main

import "fmt"

func (b *board) SolveHiddenSingle() error {
	// Hidden Single - a given cell contains a candidate which is only
	// present in this cell and not in the rest of the row/column/box
	const technique = "HIDDEN-SINGLE"

	for i := 0; i < 81; i++ {
		if b.solved[i] != 0 {
			continue
		}
		blit := b.blits[i]

		var sumBlits uint
		sumHints := func(target int, source int) error {
			if target == source {
				return nil
			}
			sumBlits |= b.blits[target]

			return nil
		}

		dims := []struct {
			op   containerOperator
			name string
		}{
			{name: "row", op: b.operateOnRow},
			{name: "column", op: b.operateOnColumn},
			{name: "box", op: b.operateOnBox},
		}

		for _, dim := range dims {
			sumBlits = 0
			if err := dim.op(i, sumHints); err != nil {
				return err
			}
			leftOver := blit & ^sumBlits

			if HasSingleBit(leftOver) {
				val := GetSingleBitValue(leftOver)

				coords := getCoords(i)
				logFormat := fmt.Sprintf("%s(%s) is the only cell with %d in its %s, changed to solution.",
					coords, GetBitsString(b.blits[i]), val, dim.name)

				if err := b.SolvePositionWithLog(technique, logFormat, i, val); err != nil {
					return err
				}
				return nil
			}
		}
	}

	return nil
}
