package main

import "time"

// Algorithm is an interface for a tile value elimination algorithm.
type Algorithm interface {
	// Name returns the algorithm's display name.
	Name() string
	// EvaluateChanges is called when there are new changes for the algorithm to
	// evaluate.
	// The function is provided with the board, and a list of tile indices which
	// have changed since the last time the function was called.
	EvaluateChanges(*Board, []uint8) bool
	// Stats returns a pointer to an AlgorithmStats object to be used for tracking
	// statistics of the algorithm.
	Stats() *AlgorithmStats
}

// AlgorithmStats tracks statistics about the performance of an algorithm.
type AlgorithmStats struct {
	// Calls is how many times the EvaluateChanges() was called.
	Calls uint
	// Changes is how many tiles were changed by the algorithm.
	Changes uint
	// Duration is the time spent within EvaluateChanges().
	Duration time.Duration
}

// algoKnownValueElimination looks for tiles which have a known value. If any
// are found, remove that value as a possibility from its neighbors.
type algoKnownValueElimination struct {
	AlgoStats AlgorithmStats
}

func (a algoKnownValueElimination) Name() string { return "algoKnownValueElimination" }

func (a *algoKnownValueElimination) Stats() *AlgorithmStats { return &a.AlgoStats }

func (a algoKnownValueElimination) EvaluateChanges(b *Board, changes []uint8) bool {
	for _, ti := range changes {
		t := b.Tiles[ti]
		if !t.isKnown() {
			continue
		}

		x, y := indexToXY(ti)
		rgnIdx := tileIndexToRegionIndex(ti)
		rgnIndices := RegionIndices[rgnIdx][:]
		rowIndices := RowIndices[y][:]
		colIndices := ColumnIndices[x][:]

		// iterate over the region
		for _, nti := range rgnIndices {
			if nti == ti {
				// skip ourself
				continue
			}
			if !b.set(nti, ^t) {
				// invalid board configuration
				return false
			}
		}

		// iterate over the row
		for _, nti := range rowIndices {
			if nti == ti {
				// skip ourself
				continue
			}
			if !b.set(nti, ^t) {
				// invalid board configuration
				return false
			}
		}

		// iterate over the column
		for _, nti := range colIndices {
			if nti == ti {
				// skip ourself
				continue
			}
			if !b.set(nti, ^t) {
				// invalid board configuration
				return false
			}
		}
	}

	return true
}

// algoOnePossibleTile scans each set of neighbors for any values which have
// only one possible tile.
type algoOnePossibleTile struct {
	AlgoStats AlgorithmStats
}

func (a algoOnePossibleTile) Name() string { return "algoOnePossibleTile" }

func (a *algoOnePossibleTile) Stats() *AlgorithmStats { return &a.AlgoStats }

func (a algoOnePossibleTile) EvaluateChanges(b *Board, changes []uint8) bool {
	var regionsSeen uint16
	var rowsSeen uint16
	var columnsSeen uint16

	for _, ti := range changes {
		x, y := indexToXY(ti)
		rgnIdx := tileIndexToRegionIndex(ti)

		// Iterate over the region.
		// But first, see if we've already done so for this specific region.
		regionMask := uint16(1 << rgnIdx)
		if regionsSeen&regionMask == 0 {
			regionsSeen |= regionMask

			if !a.evaluateChangesNS(b, RegionIndices[rgnIdx][:]) {
				return false
			}
		}

		// Now iterate over the row.
		// Again, seeing if we've already done so.
		rowMask := uint16(1 << y)
		if rowsSeen&rowMask == 0 {
			rowsSeen |= rowMask

			if !a.evaluateChangesNS(b, RowIndices[y][:]) {
				return false
			}
		}

		// And now the column.
		columnMask := uint16(1 << x)
		if columnsSeen&columnMask == 0 {
			columnsSeen |= columnMask

			if !a.evaluateChangesNS(b, ColumnIndices[x][:]) {
				return false
			}
		}
	}

	return true
}

// evaluateChangesNS evaluates the algorithm for the given neighbor set.
func (a algoOnePossibleTile) evaluateChangesNS(b *Board, idcs []uint8) bool {
ValueLoop:
	for v := Tile(1); v < tAny; v = v << 1 {
		//TODO this feels like there should be an optimized way to find which bits are set in only one of a set of numbers
		ti := uint8(255)
		for _, nti := range idcs {
			nt := b.Tiles[nti]
			if nt == v {
				// this value already has been found
				continue ValueLoop
			}
			if nt&v == 0 {
				// not a possible tile
				continue
			}
			// is a candidate
			if ti != 255 {
				// this is the second candidate
				continue ValueLoop
			}
			ti = nti
		}
		if ti == 255 {
			// no possible tiles for this value
			return false
		}
		if !b.set(ti, v) {
			// invalid board configuration
			return false
		}
	}
	return true
}

// algoOnlyRow checks if there is only a single row or column within a region
// which can hold a value. If so, it eliminates the value from the
// possibilities within the same row/column of neighboring regions.
type algoOnlyRow struct {
	AlgoStats AlgorithmStats
}

func (a algoOnlyRow) Name() string { return "algoOnlyRow" }

func (a *algoOnlyRow) Stats() *AlgorithmStats { return &a.AlgoStats }

func (a algoOnlyRow) EvaluateChanges(b *Board, changes []uint8) bool {
	var regionsSeen uint16
	for _, ti := range changes {
		// skip any regions we've already seen this round
		rgnIdx := tileIndexToRegionIndex(ti)
		regionMask := uint16(1 << rgnIdx)
		if regionsSeen&regionMask != 0 {
			continue
		}
		regionsSeen |= regionMask

		rgnIndices := RegionIndices[rgnIdx][:]

		// row first
	OnePossibleRowLoop:
		for v := Tile(1); v < tAny; v = v << 1 {
			tcRow := uint8(255)
			for _, nti := range rgnIndices {
				nt := b.Tiles[nti]
				if nt == v {
					// this value has already been found
					continue OnePossibleRowLoop
				}
				if nt&v == 0 {
					// not a possible tile
					continue
				}
				_, y := indexToXY(nti)
				if tcRow == y {
					// row already a candidate
					continue
				}
				if tcRow != 255 {
					// multiple candidate rows
					continue OnePossibleRowLoop
				}
				tcRow = y
			}
			if tcRow == 255 {
				// no candidate rows. Wat?
				return false
			}

			// iterate over the candidate row, excluding the value from tiles in other regions
			for _, nti := range RowIndices[tcRow][:] {
				if tileIndexToRegionIndex(nti) == rgnIdx {
					// skip our region
					continue
				}
				if !b.set(nti, ^v) {
					// invalid board configuration
					return false
				}
			}
		}

	OnePossibleColumnLoop:
		for v := Tile(1); v < tAny; v = v << 1 {
			tcCol := uint8(255)
			for _, nti := range rgnIndices {
				nt := b.Tiles[nti]
				if nt == v {
					// this value has already been found
					continue OnePossibleColumnLoop
				}
				if nt&v == 0 {
					// not a possible tile
					continue
				}
				x, _ := indexToXY(nti)
				if tcCol == x {
					// column already a candidate
					continue
				}
				if tcCol != 255 {
					// multiple candidate columns
					continue OnePossibleColumnLoop
				}
				tcCol = x
			}
			if tcCol == 255 {
				// no candidate columns. Wat?
				return false
			}

			for _, nti := range ColumnIndices[tcCol][:] {
				if tileIndexToRegionIndex(nti) == rgnIdx {
					// skip our region
					continue
				}
				if !b.set(nti, ^v) {
					// invalid board configuration
					return false
				}
			}
		}
	}

	return true
}

// algoNakedSubset finds any tiles within a neighbor set for which the number of
// possible values within the tile is the same as the number of tiles with the
// same possible values, and eliminates the values in those tiles from all other
// tiles within the set.
//
// https://www.kristanix.com/sudokuepic/sudoku-solving-techniques.php "Naked Subset"
//
type algoNakedSubset struct {
	AlgoStats AlgorithmStats
}

func (a algoNakedSubset) Name() string { return "algoNakedSubset" }

func (a *algoNakedSubset) Stats() *AlgorithmStats { return &a.AlgoStats }

func (a algoNakedSubset) EvaluateChanges(b *Board, changes []uint8) bool {
	var regionsSeen uint16
	var rowsSeen uint16
	var columnsSeen uint16

	for _, ti := range changes {
		rgnIdx := tileIndexToRegionIndex(ti)
		x, y := indexToXY(ti)

		// first scan the region
		regionMask := uint16(1 << rgnIdx)
		if regionsSeen&regionMask == 0 {
			regionsSeen |= regionMask

			if !a.evaluateChangesNS(b, RegionIndices[rgnIdx][:]) {
				return false
			}
		}

		// now scan the row
		rowMask := uint16(1 << y)
		if rowsSeen&rowMask == 0 {
			rowsSeen |= rowMask

			if !a.evaluateChangesNS(b, RowIndices[y][:]) {
				return false
			}
		}

		// now scan the column
		columnMask := uint16(1 << x)
		if columnsSeen&columnMask == 0 {
			columnsSeen |= columnMask

			if !a.evaluateChangesNS(b, ColumnIndices[x][:]) {
				return false
			}
		}
	}

	return true
}

// evaluateChangesNS evaluates the algorithm for the given neighbor set.
func (a algoNakedSubset) evaluateChangesNS(b *Board, idcs []uint8) bool {
	setCounts := map[Tile]uint8{}
	for _, nti := range idcs {
		nt := b.Tiles[nti]
		if nt.isKnown() {
			// Technically this is one such case. One possible value within the tile,
			// and one tile with this possible set. But this is already handled by
			// algoOnePossibleTile.
			continue
		}

		setCounts[nt]++
	}
	for t, setCount := range setCounts {
		possibilityCount := uint8(len(MaskBits[t]))
		if possibilityCount != setCount {
			continue
		}
		// if we're here, then we have a combination of N tiles with N possibilities.
		for _, nti := range idcs {
			nt := b.Tiles[nti]
			if nt == t {
				// this is one of the N tiles
				continue
			}
			if nt&t != 0 {
				// this tile has some of the possibilities, remove them
				if !b.set(nti, ^t) {
					return false
				}
			}
		}
	}

	return true
}

// algoHiddenSubset finds all subsets which have only the same number of
// possible tiles as the number of possible values within the set.
// Think of 2 tiles with possiblities [1,4,7] and [1,4,9], where these are the
// only to tiles to contain possibilities for 1 & 4. Because of that, we can
// exempt 7 & 9 from the possibilities of these 2 tiles.
// Likewise for [1,4,7,9],[1,3,4,7],[1,2,4,7], if no other tile has 1, 4, or 7,
// we can set all 3 tiles to [1,4,7].
type algoHiddenSubset struct {
	AlgoStats AlgorithmStats
}

func (a algoHiddenSubset) Name() string { return "algoHiddenSubset" }

func (a *algoHiddenSubset) Stats() *AlgorithmStats { return &a.AlgoStats }

func (a algoHiddenSubset) EvaluateChanges(b *Board, changes []uint8) bool {
	// the algorithm works like this:
	// 1. Iterate over the values 1-9
	// 1.1. Find each tile which can hold that value.
	// 2. Group the values together which have the same candidate tiles.
	// 2.1 If the number of grouped values is the same as the number of candidate
	//     tiles, that is a hidden subset.
	// 2.2 Remove all other possible values from the candidate tiles.

	var regionsSeen uint16
	var rowsSeen uint16
	var columnsSeen uint16

	for _, ti := range changes {
		x, y := indexToXY(ti)
		rgnIdx := tileIndexToRegionIndex(ti)

		// iterate over the region
		regionMask := uint16(1 << rgnIdx)
		if regionsSeen&regionMask == 0 {
			regionsSeen |= regionMask

			if !a.evaluateChangesNS(b, RegionIndices[rgnIdx][:]) {
				return false
			}
		}

		// iterate over the row
		rowMask := uint16(1 << y)
		if rowsSeen&rowMask == 0 {
			rowsSeen |= rowMask

			if !a.evaluateChangesNS(b, RowIndices[y][:]) {
				return false
			}
		}

		// iterate over the column
		colMask := uint16(1 << x)
		if columnsSeen&colMask == 0 {
			columnsSeen |= colMask

			if !a.evaluateChangesNS(b, ColumnIndices[x][:]) {
				return false
			}
		}
	}
	return true
}

// evaluateChangesNS evaluates the algorithm for the given neighbor set.
func (a algoHiddenSubset) evaluateChangesNS(b *Board, idcs []uint8) bool {
	// valueTileIndices is a list of values to a bit mask of tile indices which hold that value.
	// E.G. `3 => 0b001000010` means that the value 3 is a possibility for tiles 2 & 7.
	valueTileIndices := [9]uint16{}
	// 1. Iterate over the values 1-9
	for v := uint8(0); v < 9; v++ { // v is one less than the actual number we're dealing with
		// 1.1. Find each tile which can hold that value.
		for i, nti := range idcs {
			nt := b.Tiles[nti]
			if nt&(1<<v) == 0 {
				continue
			}
			valueTileIndices[v] |= 1 << uint16(i)
		}
	}

	// 2. Group the values together which have the same candidate tiles.
	// We basically reverse the valueTileIndices list.
	// sets is a map of a set of indices (as a bit mask) to a bit mask of the
	// values in that set.
	// E.G. `0b001000010 => 0b000000101` means that tiles 2 & 7 are both the only
	// candidates for values 1 & 3.
	sets := make(map[uint16]Tile, 9)
	for v, stiMask := range valueTileIndices[:] {
		sets[stiMask] |= 1 << uint8(v)
	}

	// 2.1 If the number of grouped values is the same as the number of candidate
	// tiles, that is a hidden subset.
	for stiMask, valuesMask := range sets {
		// break the tile indicies bitmask out into separate indicies
		tileIndices := MaskBits[stiMask]

		valuesCount := len(MaskBits[valuesMask])
		if valuesCount != len(tileIndices) {
			// not a hidden subset
			continue
		}

		// 2.2 Remove all other possible values from the candidate tiles.
		for _, sti := range tileIndices {
			if !b.set(idcs[sti], valuesMask) {
				return false
			}
		}
	}

	return true
}
