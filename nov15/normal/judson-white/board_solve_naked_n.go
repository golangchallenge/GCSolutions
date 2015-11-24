package main

import (
	"fmt"
	"strings"
)

func (b *board) SolveNakedN(n int) error {
	// When a cell has N candidates and (N-1) other cells have combined
	// hints equal to the N candidates, then all N candidates can be removed
	// from the rest of the cells in common.
	// http://planetsudoku.com/how-to/sudoku-naked-triple.html
	// http://planetsudoku.com/how-to/sudoku-naked-quad.html
	const techniqueFormat = "NAKED-%s"

	if n < 2 || n > 5 {
		return fmt.Errorf("n must be between [2,5], actual=%d", n)
	}

	for i := 0; i < 81; i++ {
		if b.solved[i] != 0 {
			continue
		}
		if GetNumberOfSetBits(b.blits[i]) > uint(n) {
			continue
		}

		ops := []containerOperator{
			b.operateOnRow,
			b.operateOnColumn,
			b.operateOnBox,
		}

		var err error

		for _, op := range ops {
			var pickList []int
			addToPickList := func(target int, source int) error {
				if target == source || b.solved[target] != 0 {
					return nil
				}
				pickList = append(pickList, target)
				return nil
			}

			if err = op(i, addToPickList); err != nil {
				return err
			}

			if len(pickList) <= n {
				continue
			}

			perms := getPermutations(n, pickList, []int{i})
			for _, list := range perms {
				var blit uint
				for _, item := range list {
					blit |= b.blits[item]
				}

				if GetNumberOfSetBits(blit) != uint(n) {
					continue
				}

				removeHints := func(target int, source int) error {
					for _, item := range list {
						if item == target {
							return nil
						}
					}

					var logEntry *updateLog
					if logEntry, err = b.updateCandidates(target, ^blit); err != nil {
						return err
					}

					if logEntry != nil {
						technique := fmt.Sprintf(techniqueFormat, numberToTechniqueWord(n))
						var args []interface{}
						for _, item := range list {
							args = append(args, item)
						}
						args = append(args, blit)
						b.AddLog(technique, logEntry, strings.Repeat("%v ", len(list))+"have linked hints %v", args...)
					}

					return nil
				}

				if err = op(i, removeHints); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
