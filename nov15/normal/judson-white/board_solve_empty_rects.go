package main

import "fmt"

type rectForm struct {
	intersectOffset int
	cornerOffsets   []int
}

var checkRects []rectForm

func init() {
	// 00 01 02
	// 09 10 11
	// 18 19 20

	checkRects = []rectForm{
		{intersectOffset: 0, cornerOffsets: []int{10, 11, 19, 20}},
		{intersectOffset: 1, cornerOffsets: []int{9, 11, 18, 20}},
		{intersectOffset: 2, cornerOffsets: []int{9, 10, 18, 19}},
		{intersectOffset: 9, cornerOffsets: []int{1, 2, 19, 20}},
		{intersectOffset: 10, cornerOffsets: []int{0, 2, 18, 20}},
		{intersectOffset: 11, cornerOffsets: []int{0, 1, 18, 19}},
		{intersectOffset: 18, cornerOffsets: []int{1, 2, 10, 11}},
		{intersectOffset: 19, cornerOffsets: []int{0, 2, 9, 11}},
		{intersectOffset: 20, cornerOffsets: []int{0, 1, 9, 10}},
	}
}

func (b *board) getIntersection(pos1 int, pos2 int, excludePos int) (int, error) {
	storePos := func(list *[]int) inspector {
		return func(target int, source int) error {
			if target == source || target == excludePos {
				return nil
			}
			*list = append(*list, target)
			return nil
		}
	}

	var pos1List []int
	if err := b.operateOnRow(pos1, storePos(&pos1List)); err != nil {
		return 0, err
	}
	if err := b.operateOnColumn(pos1, storePos(&pos1List)); err != nil {
		return 0, err
	}

	var pos2List []int
	if err := b.operateOnRow(pos2, storePos(&pos2List)); err != nil {
		return 0, err
	}
	if err := b.operateOnColumn(pos2, storePos(&pos2List)); err != nil {
		return 0, err
	}

	intersection := intersect(pos1List, pos2List)
	if len(intersection) != 1 {
		return 0, fmt.Errorf("getIntersection: expected len(intersection)==1, actual: %v", intersection)
	}
	return intersection[0], nil
}

func (b *board) SolveEmptyRectangles() error {
	// http://www.sudokuwiki.org/Empty_Rectangles

	const technique = "EMPTY RECTANGLES"
	const logFormat = "ERI: %v strong link 1: %v strong link 2: %v target corner: %v hint: %v"

	// find a candidate box
	prevBox := -1
	b.changed = false // TODO: it'd be nice to get a changed notification from updateCandidates
	for i := 0; i < 81; i += 3 {
		c1 := getCoords(i)

		if b.changed {
			// let simpler techniques take over
			return nil
		}

		if c1.box <= prevBox {
			continue
		}
		prevBox = c1.box

		var boxCoords []coords
		getBoxCells := func(target int, source int) error {
			boxCoords = append(boxCoords, getCoords(target))
			return nil
		}

		var err error

		if err = b.operateOnBox(i, getBoxCells); err != nil {
			return err
		}

		// we must have at least two candidates left in the box
		// and four or more empty cells.
		emptyCount := 0
		for _, cell := range boxCoords {
			if b.solved[cell.pos] == 0 {
				emptyCount++
				if emptyCount == 4 {
					break
				}
			}
		}
		if emptyCount < 4 {
			continue
		}

		// An Empty Rectangle occurs within a box where four cells form a
		// rectangle which does NOT contain a certain candidate.

		for _, form := range checkRects {
			var hintSum uint
			for _, cellOffset := range form.cornerOffsets {
				hintSum |= b.blits[i+cellOffset]
			}

			missingHints := hintSum ^ 0x1FF
			if missingHints == 0 {
				continue
			}

			// one or more of the missing hints have to exist in all cells in the
			// the row and col based on the empty-rectangle-intersection hinge
			// a picture makes this clearer: http://www.sudokuwiki.org/PuzImages/ER5.png

			sharedHints := missingHints
			tmpNum := 0
		boxLoop:
			for _, cell := range boxCoords {
				// skip cells we're not interested in
				if form.intersectOffset+i == cell.pos {
					continue
				}
				for _, cellOffset := range form.cornerOffsets {
					if cellOffset+i == cell.pos {
						continue boxLoop
					}
				}

				tmpNum++

				sharedHints &= b.blits[cell.pos]
				if sharedHints == 0 {
					break
				}
			}

			if sharedHints == 0 {
				continue
			}

			// alright we've got our shared hint(s)
			// now look for a "strong link" in the col or row direction

			// get cells in the row/col of our hinge/eri which have the shared hint,
			// see if they're a strong link (only 2 candidates) in the inverse dimension
			// eg, if we found it by looking at the row, check the column for a strong link
			for _, hint := range GetBitList(sharedHints) {
				// get cells in the row/col of our eri with the desired hint, excluding our those in our box
				checkCands := func(cands *[]int) func(int, int) error {
					return func(target int, source int) error {
						if getCoords(target).box == c1.box {
							return nil
						}
						if b.blits[target]&hint != 0 {
							*cands = append(*cands, target)
						}
						return nil
					}
				}

				var rowCands []int
				if err = b.operateOnRow(i+form.intersectOffset, checkCands(&rowCands)); err != nil {
					return err
				}

				var colCands []int
				if err = b.operateOnColumn(i+form.intersectOffset, checkCands(&colCands)); err != nil {
					return err
				}

				// check if any of the canidates have strong links
				getCellsWithHint := func(cands *[]int) func(int, int) error {
					return func(target int, source int) error {
						if target == source {
							return nil
						}
						if b.blits[target]&hint != 0 {
							*cands = append(*cands, target)
						}
						return nil
					}
				}

				var logEntry *updateLog
				for _, rowCand := range rowCands {
					var links []int
					if err = b.operateOnColumn(rowCand, getCellsWithHint(&links)); err != nil {
						return err
					}

					if len(links) == 1 {
						target, err := b.getIntersection(links[0], i+form.intersectOffset, rowCand)
						if err != nil {
							return err
						}

						// only if the target is not in our box is it a candidate.
						if getCoords(target).box != c1.box {
							if logEntry, err = b.updateCandidates(target, ^hint); err != nil {
								return err
							}

							if logEntry != nil {
								var args []interface{}
								args = append(args, i+form.intersectOffset)
								args = append(args, rowCand)
								args = append(args, links[0])
								args = append(args, target)
								args = append(args, hint)
								b.AddLog(technique, logEntry, logFormat, args...)
							}
						}
					}
				}

				for _, colCand := range colCands {
					var links []int
					if err = b.operateOnRow(colCand, getCellsWithHint(&links)); err != nil {
						return err
					}

					if len(links) == 1 {
						target, err := b.getIntersection(links[0], i+form.intersectOffset, colCand)
						if err != nil {
							return err
						}

						// only if the target is not in our box is it a a candidate.
						if getCoords(target).box != c1.box {
							if logEntry, err = b.updateCandidates(target, ^hint); err != nil {
								return err
							}

							if logEntry != nil {
								var args []interface{}
								args = append(args, i+form.intersectOffset)
								args = append(args, colCand)
								args = append(args, links[0])
								args = append(args, target)
								args = append(args, hint)
								b.AddLog(technique, logEntry, logFormat, args...)
							}
						}
					}
				}
			}
		}
	}

	return nil
}
