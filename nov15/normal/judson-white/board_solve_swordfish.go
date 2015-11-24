package main

import (
	"strings"
)

type swordfishOperation struct {
	blockType                    string
	op                           containerOperator
	opInverted                   containerOperator
	nextContainer                func(int) int
	doesOverlap                  func(x int, y int) bool
	isInSameDimension            func(x int, y int) bool
	maxMisses                    func(overlapSet []int) int
	getMaxPosFromOverlapSet      func(overlapSet []int) int
	getInvertedDimensionPosition func(pos int) int
}

func (b *board) SolveSwordFish() error {
	// http://www.sudokuwiki.org/Sword_Fish_Strategy
	// find a 3x3 which share the same candidate
	// hint cannot be repeated on container (row or col depending on orientation)
	// 333,332,322,222 = all valid.

	// create dimensions for looking for SwordFish in the row and column dimension
	dims := []swordfishOperation{
		{
			blockType:  "row",
			op:         b.operateOnRow,
			opInverted: b.operateOnColumn,
			nextContainer: func(cur int) int {
				// next start row index
				next := cur + 9
				if next >= 81 {
					return -1
				}
				return next
			},
			doesOverlap: func(x int, y int) bool {
				coords1 := getCoords(x)
				coords2 := getCoords(y)
				return coords1.col == coords2.col
			},
			isInSameDimension: func(x int, y int) bool {
				coords1 := getCoords(x)
				coords2 := getCoords(y)
				return coords1.row == coords2.row
			},
			maxMisses: func(overlapSet []int) int {
				// count distinct cols
				cols := make(map[int]interface{})
				for _, pos := range overlapSet {
					cols[getCoords(pos).col] = struct{}{}
				}
				return 3 - len(cols)
			},
			getMaxPosFromOverlapSet: func(overlapSet []int) int {
				maxRow := 0
				for _, pos := range overlapSet {
					coords := getCoords(pos)
					if coords.row > maxRow {
						maxRow = coords.row
					}
				}
				return maxRow * 9
			},
			getInvertedDimensionPosition: func(pos int) int {
				return getCoords(pos).col
			},
		},
		{
			blockType:  "column",
			op:         b.operateOnColumn,
			opInverted: b.operateOnRow,
			nextContainer: func(cur int) int {
				// next start column index
				next := cur + 1
				if next >= 9 {
					return -1
				}
				return next
			},
			doesOverlap: func(x int, y int) bool {
				coords1 := getCoords(x)
				coords2 := getCoords(y)
				return coords1.row == coords2.row
			},
			isInSameDimension: func(x int, y int) bool {
				coords1 := getCoords(x)
				coords2 := getCoords(y)
				return coords1.col == coords2.col
			},
			maxMisses: func(overlapSet []int) int {
				// count distinct rows
				rows := make(map[int]interface{})
				for _, pos := range overlapSet {
					rows[getCoords(pos).row] = struct{}{}
				}
				return 3 - len(rows)
			},
			getMaxPosFromOverlapSet: func(overlapSet []int) int {
				maxCol := 0
				for _, pos := range overlapSet {
					coords := getCoords(pos)
					if coords.col > maxCol {
						maxCol = coords.col
					}
				}
				return maxCol
			},
			getInvertedDimensionPosition: func(pos int) int {
				return getCoords(pos).row
			},
		},
	}

	for _, dim := range dims {

		for i := 0; i != -1; i = dim.nextContainer(i) {
			// i = start of row/column

			// hint, list of positions
			initialCandidates := make(map[uint][]int)

			// operate on all cells in container
			// get list of candidate cells: map(hint, []pos)
			extractHints := func(target int, source int) error {
				bitList := GetBitList(b.blits[target])
				if len(bitList) < 2 {
					return nil
				}

				for _, bit := range bitList {
					list, ok := initialCandidates[bit]
					if !ok {
						initialCandidates[bit] = []int{target}
					} else {
						initialCandidates[bit] = append(list, target)
					}
				}
				return nil
			}

			if err := dim.op(i, extractHints); err != nil {
				return err
			}

			// get a list of hints containing permutations of 2 or 3 cells
			candidatePerms := swordfishGetCandidatePermutations(initialCandidates)
			if len(candidatePerms) == 0 {
				continue
			}

			for hint, v := range candidatePerms {
				for _, x := range v {
					sfPerms, err := b.swordfishGetPermutations(dim, x, hint, i)
					if err != nil {
						return err
					}

					for _, y := range sfPerms {
						// combine first and second level to get new 'one must overlap' set
						// (in this case, two must overlap...)
						var overlapSet []int
						overlapSet = append(overlapSet, x...)
						overlapSet = append(overlapSet, y...)

						nextPos := dim.getMaxPosFromOverlapSet(overlapSet)
						sfPerms2, err := b.swordfishGetPermutations(dim, overlapSet, hint, nextPos)
						if err != nil {
							return err
						}

						if len(sfPerms2) != 0 {
							for _, z := range sfPerms2 {
								// Here we go...
								// if we pulled sets from columns, given a hint, remove that hint
								// from all cells in the superset of ROWS covered by our swordfish
								// set. same concept applies for rows. still confused? me too.
								// http://www.sudokuwiki.org/Sword_Fish_Strategy

								err := b.swordfishApply(dim, hint, x, y, z)
								if err != nil {
									return err
								}
								if b.changed {
									// try simpler techniques before re-trying swordfish
									return nil
								}
							}
						}
					}
				}
			}
		}
	}

	return nil
}

func (b *board) swordfishGetPermutations(dim swordfishOperation, mustOverlap []int, hint uint, pos int) ([][]int, error) {
	var perms [][]int

	// make sure we're still on the board
	nextPos := dim.nextContainer(pos)
	if nextPos == -1 {
		return perms, nil
	}

	var candidates []int
	getCandidates := func(target int, source int) error {
		// cell must contain the hint we're looking for
		if b.blits[target]&hint != hint {
			return nil
		}

		candidates = append(candidates, target)
		return nil
	}

	if err := dim.op(nextPos, getCandidates); err != nil {
		return perms, err
	}

	tmpPerms := swordfishGetTwosAndThrees(candidates)

	maxMisses := dim.maxMisses(mustOverlap)
	if maxMisses < 0 {
		return perms, nil
	}

	for _, perm := range tmpPerms {
		misses := 0
		for _, item := range perm {
			doesOverlap := false
			for _, existingSetItem := range mustOverlap {
				if dim.doesOverlap(item, existingSetItem) {
					doesOverlap = true
					break
				}
			}
			if !doesOverlap {
				misses++
				if misses > maxMisses {
					break
				}
			}
		}
		if misses <= maxMisses {
			perms = append(perms, perm)
		}
	}

	permsN, err := b.swordfishGetPermutations(dim, mustOverlap, hint, nextPos)
	if err != nil {
		return perms, err
	}

	perms = append(perms, permsN...)

	return perms, nil
}

func swordfishGetCandidatePermutations(orig map[uint][]int) map[uint][][]int {
	filtered := make(map[uint][][]int)
	for k, v := range orig {
		list := swordfishGetTwosAndThrees(v)
		if len(list) != 0 {
			filtered[k] = list
		}
	}
	return filtered
}

func swordfishGetTwosAndThrees(v []int) [][]int {
	var emptyList [][]int

	fours := getPermutations(4, v, []int{})
	if len(fours) != 0 {
		return emptyList
	}

	threes := getPermutations(3, v, []int{})
	if len(threes) != 0 {
		return threes
	}

	twos := getPermutations(2, v, []int{})
	if len(twos) != 0 {
		return twos
	}

	return emptyList
}

func (b *board) swordfishApply(sf swordfishOperation, hint uint, set1 []int, set2 []int, set3 []int) error {
	const technique = "SWORDFISH"

	var overlap []int
	overlap = append(overlap, set1...)
	overlap = append(overlap, set2...)
	overlap = append(overlap, set3...)

	dupeCheck := make(map[int]interface{})
	for _, item := range overlap {
		key := sf.getInvertedDimensionPosition(item)
		_, found := dupeCheck[key]
		if found {
			continue
		}
		dupeCheck[key] = struct{}{}

		removeHint := func(target int, source int) error {
			for _, pos := range overlap {
				if sf.isInSameDimension(target, pos) {
					return nil
				}
			}

			logEntry, err := b.updateCandidates(target, ^hint)
			if err != nil {
				return err
			}

			if logEntry != nil {
				var args []interface{}
				for _, sfSigCells := range overlap {
					args = append(args, sfSigCells)
				}
				args = append(args, hint)
				b.AddLog(technique, logEntry, strings.Repeat("%v ", len(overlap))+"hint %v", args...)
			}

			return nil
		}

		if err := sf.opInverted(item, removeHint); err != nil {
			return err
		}
	}

	return nil
}
