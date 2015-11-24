package main

import (
	"math/rand"
	"sort"
	"time"
)

// algoGenerateShuffle isn't a real algorithm. It's instead used to shuffle
// MaskBits.
// The reason is that guess() picks the first value from a set in MaskBits. This
// would result in a very non-random board. So we shuffle the MaskBits around
// each loop through evaluateAlgorithms.
type algoGenerateShuffle struct {
	rand *rand.Rand
}

func (a algoGenerateShuffle) Name() string { return "algoGenerateShuffle" }

func (a algoGenerateShuffle) Stats() *AlgorithmStats { return &AlgorithmStats{} }

func (a algoGenerateShuffle) EvaluateChanges(b *Board, changes []uint8) bool {
	// This is a little heavy as we're shuffling the entire MaskBits, when we
	// really just need to shuffle a single set (the one guess() is going to use).
	// But this won't be called often, so we go with simplicity.
	for i, nums := range MaskBits[:] {
		// copy the slice so we don't mutate it in case a function up the stack is
		// ranging across it.
		numsCopy := make([]uint8, len(nums))
		copy(numsCopy, nums)
		// Fisher–Yates shuffle.
		for i := len(numsCopy) - 1; i > 0; i-- {
			j := a.rand.Intn(i + 1)
			numsCopy[i], numsCopy[j] = numsCopy[j], numsCopy[i]
		}
		MaskBits[i] = numsCopy
	}

	return true
}

// NewRandomBoard generates a new Board of the given difficulty. Difficulty is
// the number of tiles to set as unknown.
// Note that the actual number of unknown tiles may be less than the number
// requested if the algorithm can remove no further tiles.
func NewRandomBoard(difficulty int) Board {
	b := NewBoard()

	// The board is deterministic by seed.
	// Meaning the same seed always generates the same board.
	seed := time.Now().UnixNano()
	rng := rand.New(rand.NewSource(seed))

	defer func(algos []Algorithm) { b.Algorithms = algos }(b.Algorithms)
	b.Algorithms = append(b.Algorithms, &algoGenerateShuffle{rng})

	b.Set(0, 1<<uint(rng.Intn(9)))
	b.guess()

	// We now have a fully filled out board.
	// Drop some values.
	for i := 0; i < difficulty; i++ {
		if !b.dropRandomTile(rng) {
			break
		}
	}

	return b
}

// dropCandidate is a candidate tile for dropping from the board.
type dropCandidate struct {
	ti    uint8
	score int
}
type dropCandidates []dropCandidate

func (dcs dropCandidates) Len() int           { return len(dcs) }
func (dcs dropCandidates) Less(i, j int) bool { return dcs[i].score < dcs[j].score }
func (dcs dropCandidates) Swap(i, j int)      { dcs[i], dcs[j] = dcs[j], dcs[i] }
func (dcs dropCandidates) Sort()              { sort.Sort(sort.Reverse(dcs)) }
func (dcs *dropCandidates) Remove(i int)      { *dcs = append((*dcs)[:i], (*dcs)[i+1:]...) }
func (dcs dropCandidates) Shuffle(rng *rand.Rand) {
	for i := len(dcs) - 1; i > 0; i-- {
		j := rng.Intn(i + 1)
		dcs[i], dcs[j] = dcs[j], dcs[i]
	}
}

// dropRandomTile drops a random tile from the board.
// If no further tiles can be dropped without resulting in a board with multiple
// solutions, it returns false;
func (b *Board) dropRandomTile(rng *rand.Rand) bool {
	dcs := dropCandidates{}
	for ti, t := range b.Tiles {
		if !t.isKnown() {
			continue
		}
		ti := uint8(ti)

		// score is the preference to removing this tile
		score := 0

		// favor tiles with lots of known neighbors
		rgnIdx := tileIndexToRegionIndex(ti)
		for _, nti := range RegionIndices[rgnIdx][:] {
			if b.Tiles[nti].isKnown() {
				score++
			}
		}
		colIdx, rowIdx := indexToXY(ti)
		for _, nti := range RowIndices[rowIdx][:] {
			if b.Tiles[nti].isKnown() {
				score++
			}
		}
		for _, nti := range ColumnIndices[colIdx][:] {
			if b.Tiles[nti].isKnown() {
				score++
			}
		}

		// encourage diagonal symmetry
		x, y := indexToXY(ti)
		sti := xyToIndex(y, x) // flip over the up-left axis
		if !b.Tiles[sti].isKnown() {
			score += 3
		}
		sti = xyToIndex(8-y, 8-x) // flip over the up-right axis
		if !b.Tiles[sti].isKnown() {
			score += 3
		}

		dcs = append(dcs, dropCandidate{ti: ti, score: score})
	}

	dcs.Shuffle(rng)
	dcs.Sort()

	for dcs.Len() > 0 {
		dc := dcs[0]
		dcs.Remove(0)
		ti := dc.ti

		// Try to solve the board with the current value excluded as a possibility.
		// If we have a solution, then clearing this tile would result in a board with
		// multiple solutions. So retry with a different tile.
		bTest := *b
		bTest.Tiles[ti] = (^bTest.Tiles[ti]) & tAny
		bTest.changeSet[ti/27] |= 1 << (ti % 27)
		if bTest.Solve() {
			// Have multiple solutions. Try again
			continue
		}
		// Still just a single solution, so we're good to remove this tile.
		b.Tiles[ti] = tAny
		return true
	}
	return false
}
