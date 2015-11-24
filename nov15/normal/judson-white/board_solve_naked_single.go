package main

import (
	"fmt"
)

func (b *board) SolveNakedSingle() error {
	// Naked Single - only hint left
	const technique = "NAKED-SINGLE"

	doLoop := true
	for doLoop {
		doLoop = false
		for i := 0; i < 81; i++ {
			if b.solved[i] != 0 {
				continue
			}

			blit := b.blits[i]
			if !HasSingleBit(blit) {
				continue
			}

			num := GetSingleBitValue(blit)

			coords := getCoords(i)
			logFormat := fmt.Sprintf("%s(%s) has single hint %d, changed to solution.",
				coords, GetBitsString(b.blits[i]), num)

			if err := b.SolvePositionWithLog(technique, logFormat, i, num); err != nil {
				return err
			}
			doLoop = true
		}
	}

	return nil
}
