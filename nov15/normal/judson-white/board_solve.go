package main

import (
	"fmt"
	"runtime"
)

func (b *board) SolveWithSolversList(solvers []solver) error {
	// first iteration naked single
	b.loading = true // turn off logging, this run is boring

	if err := b.SolveNakedSingle(); err != nil {
		return err
	}

	b.loading = false

	if err := b.runSolvers(solvers); err != nil {
		return err
	}

	return nil
}

func (b *board) Solve() error {
	return b.SolveWithSolversList(b.getSolvers())
}

type solver struct {
	run        func() error
	name       string
	difficulty int
}

func (b *board) getSolvers() []solver {
	solvers := []solver{
		{name: "NAKED SINGLE", difficulty: 0, run: b.SolveNakedSingle},
		{name: "HIDDEN SINGLE", difficulty: 1, run: b.SolveHiddenSingle},
		{name: "NAKED PAIR", difficulty: 1, run: b.getSolverN(b.SolveNakedN, 2)},
		{name: "NAKED TRIPLE", difficulty: 3, run: b.getSolverN(b.SolveNakedN, 3)},
		{name: "NAKED QUAD", difficulty: 6, run: b.getSolverN(b.SolveNakedN, 4)},
		{name: "NAKED QUINT", difficulty: 6, run: b.getSolverN(b.SolveNakedN, 5)},
		{name: "HIDDEN PAIR", difficulty: 1, run: b.getSolverN(b.SolveHiddenN, 2)},
		{name: "HIDDEN TRIPLE", difficulty: 4, run: b.getSolverN(b.SolveHiddenN, 3)},
		{name: "HIDDEN QUAD", difficulty: 8, run: b.getSolverN(b.SolveHiddenN, 4)},
		{name: "HIDDEN QUINT", difficulty: 8, run: b.getSolverN(b.SolveHiddenN, 5)},
		{name: "POINTING PAIR AND TRIPLE REDUCTION", difficulty: 10, run: b.SolvePointingPairAndTripleReduction},
		{name: "BOX LINE", difficulty: 10, run: b.SolveBoxLine},
		{name: "X-WING", difficulty: 10, run: b.SolveXWing},
		{name: "SIMPLE-COLORING", difficulty: 10, run: b.SolveSimpleColoring},
		{name: "Y-WING", difficulty: 10, run: b.SolveYWing},
		{name: "SWORDFISH", difficulty: 10, run: b.SolveSwordFish},
		{name: "XY-CHAIN", difficulty: 10, run: b.SolveXYChain},
		{name: "EMPTY RECTANGLES", difficulty: 10, run: b.SolveEmptyRectangles},
		{name: "SAT", difficulty: 100, run: b.SolveSAT},
	}

	return solvers
}

func (b *board) getSolverN(solver func(int) error, n int) func() error {
	return func() error {
		if err := solver(n); err != nil {
			return err
		}
		return nil
	}
}

func (b *board) runSolvers(solvers []solver) error {
mainLoop:
	for {
		b.changed = false
		for _, solver := range solvers {
			if err := solver.run(); err != nil {
				return NewErrUnsolvable(err.Error())
			}
			if b.isSolved() {
				b.updateDifficulty(solver.difficulty)
				return nil
			}
			if b.changed {
				b.updateDifficulty(solver.difficulty)
				continue mainLoop
			}
		}
		// for tests and generation we may intentionally do an incomplete solve
		break
	}
	return nil
}

func getCandidates(n int, store map[int]map[uint]int) map[int][]uint {
	list := make(map[int][]uint)
	for pos, hintEliminations := range store {
		for hint, eliminations := range hintEliminations {
			if eliminations == n {
				hintList := list[pos]
				list[pos] = append(hintList, hint)
			}
		}
	}
	return list
}

func (b *board) SolvePositionNoValidate(pos int, val uint) {
	b.solved[pos] = val
	b.blits[pos] = 1 << (val - 1)
}

func (b *board) SolvePosition(pos int, val uint) error {
	mask := uint(^(1 << (val - 1)))
	removeCandidates := func(target int, source int) error {
		if _, opErr := b.updateCandidates(target, mask); opErr != nil {
			return opErr
		}
		return nil
	}

	if err := b.solvePositionWithRemover(pos, val, removeCandidates); err != nil {
		return NewErrUnsolvable("%#v val:%d - %s", getCoords(pos), val, err)
	}

	logLastBoardWithHints = b.GetTextBoardWithHints()
	return nil
}

func (b *board) SolvePositionWithLog(technique, logFormat string, pos int, val uint) error {
	mask := uint(^(1 << (val - 1)))

	logFormat += fmt.Sprintf(" solved: %d/81", b.numSolved()+1)
	removeCandidates := func(target int, source int) error {
		logEntry, opErr := b.updateCandidates(target, mask)
		if opErr != nil {
			return opErr
		}

		if logEntry != nil {
			b.AddLog(technique, logEntry, logFormat)
		}
		return nil
	}

	// writes header
	b.AddLog(technique, nil, logFormat)

	if err := b.solvePositionWithRemover(pos, val, removeCandidates); err != nil {
		return NewErrUnsolvable("%#v val:%d - %s", getCoords(pos), val, err)
	}

	return nil
}

func (b *board) solvePositionWithRemover(pos int, val uint, candidateRemover inspector) error {
	if b.solved[pos] != 0 && (!b.loading || b.solved[pos] != val) {
		return NewErrUnsolvable("pos %d has value %d, tried to set with %d", pos, b.solved[pos], val)
	}
	b.solved[pos] = val
	b.blits[pos] = 1 << (val - 1)

	if err := b.Validate(); err != nil {
		return NewErrUnsolvable("%#v val:%d - %s", getCoords(pos), val, err)
	}

	if err := b.operateOnRCB(pos, candidateRemover); err != nil {
		return NewErrUnsolvable("%#v val:%d - %s", getCoords(pos), val, err)
	}

	return nil
}

func (b *board) updateCandidates(target int, mask uint) (*updateLog, error) {
	if b.solved[target] != 0 {
		return nil, nil
	}
	oldBlit := b.blits[target]
	newBlit := oldBlit & mask
	if newBlit != oldBlit {
		b.changed = true

		if newBlit == 0 {
			_, file, line, _ := runtime.Caller(1)
			return nil, NewErrUnsolvable("tried to remove last candidate from %#2v hints:%s mask:%s file:%s:%d",
				getCoords(target), GetBitsString(oldBlit), GetBitsString(mask), file, line)
		}

		b.blits[target] = newBlit

		logEntry := updateLog{pos: target, oldHints: oldBlit, newHints: newBlit}
		return &logEntry, b.Validate()
	}
	return nil, nil
}

func (b *board) updateDifficulty(difficulty int) {
	b.difficulty += difficulty
}
