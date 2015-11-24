package main

// The generate/grading function is a work in progress.

import "math/rand"

func getValidBoard() (*board, error) {
	b, err := loadBoard([]byte("000000000000000000000000000000000000000000000000000000000000000000000000000000000"))
	if err != nil {
		return nil, err
	}

	for !b.isSolved() {
		n := rand.Intn(81)
		if b.solved[n] != 0 {
			continue
		}
		bitList := GetBitList(b.blits[n])
		bn := rand.Intn(len(bitList))
		val := GetSingleBitValue(bitList[bn])

		err = b.SolvePosition(n, val)
		if err != nil {
			return nil, err
		}

		err = b.SolveWithSolversList(b.getGeneratorSolvers())
		if err != nil {
			return nil, err
		}
	}
	return b, nil
}

func generatePuzzle(minDifficulty, maxDifficulty int) (*board, error) {
	for {
		var err error
		var b *board
		for b == nil || err != nil {
			b, err = getValidBoard()
		}
		b2, err := digHoles(b)
		if err != nil {
			return nil, err
		}

		b3, err := loadBoard([]byte(b2.GetCompact()))
		if err != nil {
			return nil, err
		}
		if err = b3.SolveWithSolversList(b.getGeneratorSolvers()); err != nil {
			return nil, err
		}

		return b2, nil
	}
}

func digHoles(b *board) (*board, error) {
	b2 := &board{solved: b.solved, blits: b.blits}

	step := 4
	failures := 0
	check := make(map[int]interface{})
	for len(check) != 81 && b2.numSolved() >= 27 {
		goodSolved := b2.solved
		goodBlits := b2.blits

		pos1 := rand.Intn(81)
		if step == 1 {
			if _, ok := check[pos1]; ok {
				continue
			}
			check[pos1] = struct{}{}
		}
		if b2.solved[pos1] == 0 {
			continue
		}

		coords := getCoords(pos1)
		secondRow := 8 - coords.row
		if secondRow < 0 {
			secondRow += 8
		}
		secondCol := 8 - coords.col
		if secondCol < 0 {
			secondCol += 8
		}

		if step == 4 {
			// dig out 4
			pos2 := coords.row*9 + secondCol
			pos3 := secondRow*9 + coords.col
			pos4 := secondRow*9 + secondCol

			if b2.solved[pos2] == 0 || b2.solved[pos3] == 0 || b2.solved[pos4] == 0 {
				continue
			}

			b2.solved[pos1] = 0
			b2.solved[pos2] = 0
			b2.solved[pos3] = 0
			b2.solved[pos4] = 0
		} else if step == 2 {
			// dig out 2
			pos2 := secondRow*9 + secondCol

			if b2.solved[pos2] == 0 {
				continue
			}

			b2.solved[pos1] = 0
			b2.solved[pos2] = 0
		} else {
			// dig out 1
			b2.solved[pos1] = 0
		}

		// attempt to solve using selected difficulty
		b3, err := loadBoard([]byte(b2.GetCompact()))
		if err != nil {
			return nil, err
		}

		err = b3.SolveWithSolversList(b.getGeneratorSolvers())

		if err != nil || !b3.isSolved() {
			// bad dig, doesn't fit difficulty or more than one solution
			b2.solved = goodSolved
			b2.blits = goodBlits
			failures++
			if step > 1 && failures == 2 {
				failures = 0
				step /= 2
			}
		}
	}

	return b2, nil
}

func (b *board) getHints(pos int) (uint, error) {
	check := make(map[uint]interface{})
	for i := uint(1); i <= 9; i++ {
		check[i] = struct{}{}
	}

	removeHints := func(target int, source int) error {
		if target == source {
			return nil
		}
		val := b.solved[target]
		if val == 0 {
			return nil
		}

		if _, ok := check[val]; ok {
			delete(check, val)
		}

		return nil
	}

	if err := b.operateOnRCB(pos, removeHints); err != nil {
		return 0, err
	}

	blits := uint(0)
	for k := range check {
		blits |= 1 << (k - 1)
	}
	return blits, nil
}

func (b *board) getGeneratorSolvers() []solver {
	var solvers []solver
	for _, solver := range b.getSolvers() {
		if solver.name != "SAT" {
			solvers = append(solvers, solver)
		}
	}

	return solvers
}
