package main

import (
	"fmt"
	"strings"
)

func (b *board) SolveHiddenN(n int) error {
	if n < 2 || n > 5 {
		return fmt.Errorf("n must be between [2,5], actual=%d", n)
	}
	// If there are N unique hints in N cells within one container,
	// then no other hints could be valid within those cells.
	// See: http://planetsudoku.com/how-to/sudoku-hidden-triple.html
	// Triple example (from URL above):
	// - X = 4,5,6,7
	// - Y = 1,4,5,6,8
	// - Z = 3,4,7,8
	// Algorithm:
	// - X | Y | Z = 1,3,4,5,6,7,8
	// - bits.GetNumberOfSetBits(X | Y | Z) >= 3 (can continue)
	// - Other cells represented by "O1, O2, O3, ..."
	// - let sum = O1 | O2 | O3 ... = 1,2,3,4,5,9 (combined hints of other cells)
	// - ^(sum) = 6,7,8 (hints not present in other cells)
	// - X|Y|Z & ^sum = 011111101 & 011100000 = 011100000 = 6,7,8 (unique hints in considered cells)
	// - bits.GetNumberOfSetBits((X | Y | Z) & ^sum) == 3 (can continue)
	// - X,Y,Z can &= ^(sum), removing 1,3,4,5 as hints from any of the considered cells
	for i := 0; i < 81; i++ {
		if b.solved[i] != 0 {
			continue
		}

		var pickList []int

		storePickList := func(target int, source int) error {
			if target == source || b.solved[target] != 0 {
				return nil
			}
			pickList = append(pickList, target)
			return nil
		}

		ops := []containerOperator{
			b.operateOnRow,
			b.operateOnColumn,
			b.operateOnBox,
		}

		for _, op := range ops {
			pickList = make([]int, 0)
			if err := op(i, storePickList); err != nil {
				return err
			}
			lists := getPermutations(n, pickList, []int{i})
			if err := b.checkHiddenPermutations(n, i, op, lists); err != nil {
				return err
			}
		}
	}
	return nil
}

func (b *board) checkHiddenPermutations(n int, source int, op containerOperator, lists [][]int) error {
	const techniqueFormat = "HIDDEN-%s"

	var err error
	var logEntry *updateLog

	for _, list := range lists {
		var sumBits uint
		for _, pos := range list {
			sumBits |= b.blits[pos]
		}
		if GetNumberOfSetBits(sumBits) < uint(n) {
			continue
		}

		sumOthers := uint(0)
		sumTheOthers := func(target int, source int) error {
			if b.solved[target] != 0 {
				return nil
			}
			for _, v := range list {
				if v == target {
					return nil
				}
			}
			sumOthers |= b.blits[target]
			return nil
		}

		if err = op(source, sumTheOthers); err != nil {
			return err
		}

		if sumOthers == 0 {
			continue
		}

		leftOver := (sumBits ^ sumOthers) & sumBits

		if GetNumberOfSetBits(leftOver) == uint(n) {
			technique := fmt.Sprintf(techniqueFormat, numberToTechniqueWord(n))

			for _, pos := range list {
				if logEntry, err = b.updateCandidates(pos, leftOver); err != nil {
					return err
				}

				if logEntry != nil {
					var args []interface{}
					for _, logPos := range list {
						args = append(args, logPos)
					}
					args = append(args, leftOver)
					b.AddLog(technique, logEntry, strings.Repeat("%v ", n)+"have unique hints %s", args...)
				}
			}
		}
	}

	return nil
}

func numberToTechniqueWord(n int) string {
	switch n {
	case 1:
		return "SINGLE"
	case 2:
		return "PAIR"
	case 3:
		return "TRIPLE"
	case 4:
		return "QUAD"
	case 5:
		return "QUINT"
	}

	return "???"
}
