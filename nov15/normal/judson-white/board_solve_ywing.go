package main

import "fmt"

func (b *board) SolveYWing() error {
	// http://www.sudokuwiki.org/Y_Wing_Strategy
	// start simple..
	// look for three cells which each have two hints and can 'see' each other,
	// like 3 corners of an xwing.
	// the three cells have hints AB,BC,CB.. any C in the 4th corner can be removed.

	// define a 'hinge' cell and the two 'wing' cells
	// the wings can be 'seen' by the hinge
	// the value NOT in the hinge ('C' for example) can be taken off any
	// cells which can be seen by both wing cells
	const technique = "Y-WING"

	for i := 0; i < 81; i++ {
		if b.solved[i] != 0 {
			continue
		}

		blit := b.blits[i]
		if GetNumberOfSetBits(blit) != 2 {
			continue
		}

		visibleToHinge := b.getVisibleCells(i)

		// filter visible list to only those which have two set bits
		// and have ONLY one in common with the hinge
		var candidates []int
		for _, item := range visibleToHinge {
			itemBlit := b.blits[item]
			if GetNumberOfSetBits(itemBlit) != 2 {
				continue
			}
			if !HasSingleBit(blit & itemBlit) {
				continue
			}
			candidates = append(candidates, item)
		}

		// get all permutations of the candidates
		perms := getPermutations(2, candidates, []int{})

		// filter permutations where the candidates share only one hint
		var wingsList [][]int
		for _, list := range perms {
			if len(list) != 2 {
				fmt.Println("len(list) != 2 ???")
				continue
			}
			wingBlit1 := b.blits[list[0]]
			wingBlit2 := b.blits[list[1]]

			if HasSingleBit(wingBlit1&wingBlit2) &&
				GetNumberOfSetBits(blit|wingBlit1|wingBlit2) == 3 {
				wingsList = append(wingsList, list)
			}
		}

		if len(wingsList) == 0 {
			continue
		}

		for _, wings := range wingsList {
			if len(wings) != 2 {
				// NOTE: len(wings) should always be 2, being defensive
				continue
			}

			sum := b.blits[wings[0]] | b.blits[wings[1]]
			targets := b.getVisibleCells(wings[0])
			targets = intersect(targets, b.getVisibleCells(wings[1]))

			if len(targets) == 0 {
				continue
			}

			removeHint := sum & ^blit

			updated := false
			for _, target := range targets {
				if target == i || target == wings[0] || target == wings[1] {
					continue
				}

				if b.blits[target]&removeHint == removeHint {
					logEntry, err := b.updateCandidates(target, ^removeHint)
					if err != nil {
						return err
					}

					if logEntry != nil {
						b.AddLog(technique, logEntry, "hinge=%v wing1=%v wing2=%v", i, wings[0], wings[1])
					}

					updated = b.changed
				}
			}

			if updated {
				// let simpler techniques take over
				return nil
			}
		}
	}

	return nil
}
