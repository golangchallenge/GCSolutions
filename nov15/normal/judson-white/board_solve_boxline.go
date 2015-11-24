package main

func (b *board) SolveBoxLine() error {
	// Two cells in a BOX that share a hint which isn't anywhere else on
	// the ROW or COLUMN they share can be removed as hints from other cells
	// in the same BOX.
	// http://planetsudoku.com/how-to/sudoku-box-line.html
	const technique = "BOX-LINE"
	const logFormat = "%s %s are in a shared box and are the only cell(s) with hint %s in their %s outside their box"

	for i := 0; i < 81; i++ {
		if b.solved[i] != 0 {
			continue
		}

		blit := b.blits[i]
		coords := getCoords(i)

		var colPickList []int
		var rowPickList []int

		savePickLists := func(target int, source int) error {
			if target == source || b.solved[target] != 0 {
				return nil
			}
			if b.blits[target]&blit == 0 {
				// nothing shared
				return nil
			}

			targetCoords := getCoords(target)
			if targetCoords.row == coords.row {
				rowPickList = append(rowPickList, target)
			} else if targetCoords.col == coords.col {
				colPickList = append(colPickList, target)
			}
			return nil
		}

		if err := b.operateOnBox(i, savePickLists); err != nil {
			return err
		}

		dims := []struct {
			pickList  []int
			op        containerOperator
			canRemove func(int) bool
			name      string
		}{
			// rows
			{
				pickList:  rowPickList,
				op:        b.operateOnRow,
				canRemove: func(target int) bool { return coords.row != getCoords(target).row },
				name:      "row",
			},
			// columns
			{
				pickList:  colPickList,
				op:        b.operateOnColumn,
				canRemove: func(target int) bool { return coords.col != getCoords(target).col },
				name:      "column",
			},
		}

		for _, dim := range dims {
			perms := getPermutations(1, dim.pickList, []int{})
			for _, list := range perms {
				for _, item := range list {
					sharedHints := blit & b.blits[item]
					hintList := GetBitList(sharedHints)

					for _, hint := range hintList {
						safeToRemove := true
						checkLine := func(target int, source int) error {
							if target == source {
								return nil
							}
							if getCoords(target).box == coords.box {
								return nil
							}
							// the shared hints must be the only ones in the shared row/col
							if b.blits[target]&hint != 0 {
								safeToRemove = false
							}
							return nil
						}

						if err := dim.op(i, checkLine); err != nil {
							return err
						}

						if safeToRemove {
							removeBoxLineHint := func(target int, source int) error {
								if !dim.canRemove(target) {
									return nil
								}

								logEntry, err := b.updateCandidates(target, ^hint)
								if err != nil {
									return err
								}

								if logEntry != nil {
									b.AddLog(technique, logEntry, logFormat, i, item, hint, dim.name)
								}
								return nil
							}

							if err := b.operateOnBox(i, removeBoxLineHint); err != nil {
								return err
							}

							if b.changed {
								// let simpler techniques take over
								return nil
							}
						}
					}
				}
			}
		}
	}
	return nil
}
