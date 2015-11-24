package main

import (
	"errors"
	"fmt"
	"io"
	"time"
)

// xyToIndex converts board x,y coordinates into an index.
func xyToIndex(x, y uint8) (idx uint8) {
	return y*9 + x
}

// indexToXY converts a board index into x,y coordinates.
func indexToXY(idx uint8) (x, y uint8) {
	return idx % 9, idx / 9
}

// tileIndexToRegionIndex converts the index of a tile within a board, to the index
// of the region within the board.
func tileIndexToRegionIndex(idx uint8) uint8 {
	return idx/3%3 + idx/(9*3)*3
}

// RegionIndices is a pre-calculated lookup table for obtaining the tile
// indices within a board for the given region index.
var RegionIndices [9][9]uint8 = func() (idcs [9][9]uint8) {
	for ri := range idcs {
		idx0 := uint8((ri / 3 * 27) + (ri % 3 * 3))
		idcs[ri] = [9]uint8{
			idx0 + 9*0 + 0,
			idx0 + 9*0 + 1,
			idx0 + 9*0 + 2,
			idx0 + 9*1 + 0,
			idx0 + 9*1 + 1,
			idx0 + 9*1 + 2,
			idx0 + 9*2 + 0,
			idx0 + 9*2 + 1,
			idx0 + 9*2 + 2,
		}
	}
	return
}()

// RowIndices is a pre-calculated lookup table for obtaining the tile
// indices within a board for the given row index.
var RowIndices [9][9]uint8 = func() (idcs [9][9]uint8) {
	for y := range idcs {
		idx0 := uint8(y * 9)
		idcs[y] = [9]uint8{
			idx0 + 0,
			idx0 + 1,
			idx0 + 2,
			idx0 + 3,
			idx0 + 4,
			idx0 + 5,
			idx0 + 6,
			idx0 + 7,
			idx0 + 8,
		}
	}
	return
}()

// ColumnIndices is a pre-calculated lookup table for obtaining the tile
// indices within a board for the given column index.
var ColumnIndices [9][9]uint8 = func() (idcs [9][9]uint8) {
	for x := range idcs {
		idcs[x] = [9]uint8{
			uint8(x) + 9*0,
			uint8(x) + 9*1,
			uint8(x) + 9*2,
			uint8(x) + 9*3,
			uint8(x) + 9*4,
			uint8(x) + 9*5,
			uint8(x) + 9*6,
			uint8(x) + 9*7,
			uint8(x) + 9*8,
		}
	}
	return
}()

// MaskBits is a pre-calculated lookup table for converting a uint16
// (values 0-511) into a slice indicating which bits are set.
// E.G. `MaskBits[0b001000101] == []uint8{0,2,6}`
// The main use case is to know which values are possiblities within a Tile.
var MaskBits [512][]uint8 = func() (mbs [512][]uint8) {
	for i := uint16(0); i < 512; i++ {
		for j := uint8(0); j < 9; j++ {
			if i&(1<<j) != 0 {
				mbs[i] = append(mbs[i], j)
			}
		}
	}
	return
}()

// Tile represents a sudoku tile, and the possible values it may hold.
// The value is a 9-bit mask (uint16 with 7 bits unused), with bit 0 indicating
// whether the tile can hold the digit 1, through bit 8 indicating whether the
// tile can hold the digit 9.
type Tile uint16

// tAny is a tile which holds any possible value (1-9).
const tAny = Tile((1 << 9) - 1)

// byteToTileMap is a mapping of ASCII characters to their Tile value.
var byteToTileMap = map[byte]Tile{
	'1': 1 << 0, // 0b000000001
	'2': 1 << 1, // 0b000000010
	'3': 1 << 2, // 0b000000100
	'4': 1 << 3, // 0b000001000
	'5': 1 << 4, // 0b000010000
	'6': 1 << 5, // 0b000100000
	'7': 1 << 6, // 0b001000000
	'8': 1 << 7, // 0b010000000
	'9': 1 << 8, // 0b100000000
	'_': tAny,   // 0b111111111
}

// isKnown indicates whether the tile only has a single possible value.
func (t Tile) isKnown() bool {
	// http://graphics.stanford.edu/~seander/bithacks.html#DetermineIfPowerOf2
	return (t & (t - 1)) == 0
}

// Num returns the number held by a tile. If the tile is not known (holds
// multiple possible values), 0 is returned.
func (t Tile) Num() uint8 {
	if !t.isKnown() {
		return 0
	}
	// http://graphics.stanford.edu/~seander/bithacks.html#IntegerLogDeBruijn
	lookupTable := [32]uint8{
		0, 1, 28, 2, 29, 14, 24, 3, 30, 22, 20, 15, 25, 17, 4, 8,
		31, 27, 13, 23, 21, 19, 16, 7, 26, 12, 18, 6, 11, 5, 10, 9,
	}
	//TODO adjust the table to remove the <<1
	// the table is also larger than we need
	// it should also be global so it's not redeclared
	return lookupTable[(uint32(t<<1)*0x077CB531)>>27]
}

// Board represents a sudoku board.
type Board struct {
	// Tiles holds a 9x9 grid of the tiles on the board.
	// The tiles are stored serially by row. Index 0 is x=0,y=0, index 9 is
	// x=0,y=1, index 19 is x=1,y=2, and so forth.
	Tiles [9 * 9]Tile

	// Algorithms is a list of algorithms to use when solving the board.
	Algorithms []Algorithm
	// activeAlgorithmStats is a pointer the the AlgorithmStats for the algorithm
	// which is currently running.
	activeAlgorithmStats *AlgorithmStats
	// guessStats tracks the AlgorithmStats for the guesser.
	guessStats *AlgorithmStats

	// changeSet is a bit mask representing which tiles have changed.
	// Each row of regions is a uint32 (27 tiles per region-row, so 5 bytes
	// unused).
	changeSet [3]uint32

	// changesBase is an array used as the backing store for the slice returned by
	// changes(). This is to reduce heap allocations.
	changesBase [9 * 9]uint8
}

// NewBoard creates a new board with all tiles unknown.
func NewBoard() Board {
	return Board{
		Tiles: [81]Tile{
			tAny, tAny, tAny, tAny, tAny, tAny, tAny, tAny, tAny,
			tAny, tAny, tAny, tAny, tAny, tAny, tAny, tAny, tAny,
			tAny, tAny, tAny, tAny, tAny, tAny, tAny, tAny, tAny,
			tAny, tAny, tAny, tAny, tAny, tAny, tAny, tAny, tAny,
			tAny, tAny, tAny, tAny, tAny, tAny, tAny, tAny, tAny,
			tAny, tAny, tAny, tAny, tAny, tAny, tAny, tAny, tAny,
			tAny, tAny, tAny, tAny, tAny, tAny, tAny, tAny, tAny,
			tAny, tAny, tAny, tAny, tAny, tAny, tAny, tAny, tAny,
			tAny, tAny, tAny, tAny, tAny, tAny, tAny, tAny, tAny,
		},
		Algorithms: []Algorithm{
			&algoKnownValueElimination{},
			&algoOnePossibleTile{},
			&algoOnlyRow{},
			&algoNakedSubset{},
			&algoHiddenSubset{},
		},
		guessStats: &AlgorithmStats{},
	}
}

// Set tries to set the given index to the given Tile value, and then evaluates
// all the algorithms to make any eliminations possible with the new change.
// The value set on the board might be different than the one provided if
// possiblities can be eliminated.
// Returns whether the operation was successful or not. The operation will be
// unsuccessful if the value results in an invalid board.
func (b *Board) Set(ti uint8, t Tile) bool {
	b0 := *b

	if !b.set(ti, t) {
		return false
	}
	for b.hasChanges() {
		if !b.evaluateAlgorithms() {
			*b = b0
			return false
		}
	}
	return true
}

// set tries to set the given index to the given Tile value.
// set is different from Set in that no algorithms are run after the tile is
// changed.
func (b *Board) set(ti uint8, t Tile) bool {
	t0 := b.Tiles[ti]

	// discard possible values based on the current tile mask
	t &= t0
	if t == 0 {
		// not possible captain
		return false
	}

	if t == t0 {
		// no change
		return true
	}

	if b.activeAlgorithmStats != nil {
		b.activeAlgorithmStats.Changes++
	}

	b.Tiles[ti] = t
	b.changeSet[ti/27] |= 1 << (ti % 27)

	return true
}

// hasChanges indicates whether any tiles have been changed since the last call
// to evaluateAlgorithms.
func (b *Board) hasChanges() bool {
	return b.changeSet[0] != 0 || b.changeSet[1] != 0 || b.changeSet[2] != 0
}

// clearChanges resets the change list used by hasChanges and changes.
func (b *Board) clearChanges() {
	b.changeSet[0] = 0
	b.changeSet[1] = 0
	b.changeSet[2] = 0
}

// changes returns a slice of tile indices for all the tiles which have changed
// since the last call to evaluateAlgorithms.
// Note: as an optimization, all returned values share the same underlying
// storage. This means that each call to changes invalidates the previous return
// value.
func (b *Board) changes() []uint8 {
	changes := b.changesBase[:0]
	for rri, rrm := range b.changeSet {
		for i := uint8(0); i < 27; i++ {
			if rrm&1 != 0 {
				changes = append(changes, uint8(rri)*27+i)
			}
			rrm = rrm >> 1
		}
	}
	return changes
}

// evaluateAlgorithms evalutes all the algorithms against all the tiles that
// have changed since the last time evaluateAlgorithms was called.
// The algorithms are evaluated in a loop until none of them make a change.
func (b *Board) evaluateAlgorithms() bool {
	// back this up in case we're recursing
	defer func(s *AlgorithmStats) { b.activeAlgorithmStats = s }(b.activeAlgorithmStats)

	// This is designed such that any time an algorithms makes a change, we go back
	// to the first algorithm in the list. This is so that we let the cheap
	// algorithms do as much as they can, and we call the expensive ones as little
	// as possible.
	// This is the reason for backing up the changeSet. Each time an algo is
	// evaluated, it needs the changes since the last time it was run. But since we
	// can restart the loop, algo-1 might run a dozen times before algo-2 is run
	// once. Thus we clear the changeSet before algo evaluation, and then restore
	// the changeSet when we advance to the next algorithm. This way, if algo-2
	// makes a change, and we restart back at algo-1, algo-1 only sees the changes
	// by algo-2, but when algo-3 finally gets a turn, it will see them all.
	cs := b.changeSet
AlgorithmsLoop:
	for b.hasChanges() {
		for _, a := range b.Algorithms {
			changes := b.changes()
			b.clearChanges()

			b.activeAlgorithmStats = a.Stats()
			a.Stats().Calls++
			tStart := time.Now()
			ok := a.EvaluateChanges(b, changes)
			a.Stats().Duration += time.Now().Sub(tStart)
			b.activeAlgorithmStats = nil
			if !ok {
				return false
			}

			if b.hasChanges() {
				// add any changes just made to the backed-up changeset since the next algo
				// hasn't seen them yet.
				cs[0] |= b.changeSet[0]
				cs[1] |= b.changeSet[1]
				cs[2] |= b.changeSet[2]

				// restart from the first algorithm
				continue AlgorithmsLoop
			}
			// no changes, restore the change set for the next algo
			b.changeSet = cs
		}
		// we made it through all the algorithms with no changes
		b.clearChanges()
		break
	}

	return true
}

// guess tries to guess the the remaining unknown values on the board.
// If all guesses result in an invalid board, false it returned.
func (b *Board) guess() bool {
	b.guessStats.Calls++
	b.activeAlgorithmStats = b.guessStats
	tStart := time.Now()
	defer func() {
		b.guessStats.Duration += time.Now().Sub(tStart)
	}()

	// first look for the tile with the least amount of possible values
	uti := uint8(255)
	utPossibilityCount := uint8(255)
	for ti := range b.Tiles {
		t := b.Tiles[ti]
		if t.isKnown() {
			continue
		}
		pc := uint8(len(MaskBits[t]))
		if pc < utPossibilityCount {
			uti = uint8(ti)
			utPossibilityCount = pc
			if pc == 2 {
				// can't get less than 2 and still be unknown
				break
			}
		}
	}
	if uti == 255 {
		// entire board already solved
		return true
	}

	ut := b.Tiles[uti]

	b0 := *b
	// now try guessing a value
	for _, v := range MaskBits[ut] {
		t := Tile(1 << v)
		b.guessStats.Duration += time.Now().Sub(tStart) // pause timer
		if !b.Set(uti, t) {
			// this value is invalid
			if !b.Set(uti, ^t) {
				// the board is invalid
				tStart = time.Now()
				*b = b0
				return false
			}
			tStart = time.Now()
			continue
		}
		tStart = time.Now()
		if b.Solved() {
			return true
		}
		// still have other tiles to guess
		b.guessStats.Duration += time.Now().Sub(tStart) // pause timer
		if b.guess() {
			tStart = time.Now()
			return true
		}
		tStart = time.Now()
		// invalid board
		// reset and try the next possible value for this tile
		*b = b0
	}

	// all guesses failed. Invalid board.
	*b = b0
	return false
}

// Solved indicates whether all tiles have a known value.
func (b *Board) Solved() bool {
	for ti := uint8(0); ti < 9*9; ti++ {
		t := b.Tiles[ti]
		if !t.isKnown() {
			return false
		}
	}
	return true
}

// Solve tries to solve the board. If the board has no solution, false is
// returned.
func (b *Board) Solve() bool {
	if !b.evaluateAlgorithms() {
		return false
	}
	return b.guess()
}

// ReadFrom reads the board from the provided io.Reader. In addition to read
// errors, if the provided board is invalid, an error will be returned.
//
// The input format of the board is:
//  1 _ 3 _ _ 6 _ 8 _
//  _ 5 _ _ 8 _ 1 2 _
//  7 _ 9 1 _ 3 _ 5 6
//  _ 3 _ _ 6 7 _ 9 _
//  5 _ 7 8 _ _ _ 3 _
//  8 _ 1 _ 3 _ 5 _ 7
//  _ 4 _ _ 7 8 _ 1 _
//  6 _ 8 _ _ 2 _ 4 _
//  _ 1 2 _ 4 5 _ 7 8
func (b *Board) ReadFrom(r io.Reader) (int64, error) {
	var ba [9 * 9 * 2]byte
	nr, err := io.ReadFull(r, ba[:])
	if err != nil {
		if err == io.EOF && nr == len(ba)-1 {
			// The trailing newline is missing. This is acceptable
		} else {
			return int64(nr), err
		}
	}
	return int64(nr), b.Unmarshal(ba[:])
}

func (b *Board) Unmarshal(ba []byte) error {
	for i := 0; i < len(ba); i += 2 {
		x := uint8(i / 2 % 9)
		y := uint8(i / 2 / 9)
		ti := xyToIndex(x, y)
		t := byteToTileMap[ba[i]]
		if t == 0 {
			return errors.New("invalid byte")
		}
		if !b.set(ti, t) {
			return fmt.Errorf("invalid board")
		}
	}

	return nil
}

// Art generates a simple representation of the board, suitable for human
// viewing.
func (b Board) Art() []byte {
	var ba [9 * 9 * 2]byte
	for y := uint8(0); y < 9; y++ {
		rowStart := y * 9 * 2
		for x, ti := range RowIndices[y][:] {
			t := b.Tiles[ti]
			i := rowStart + uint8(x)*2
			ba[i] = '0' + t.Num()
			if ba[i] == '0' {
				ba[i] = '_'
			}
			ba[i+1] = ' '
		}
		ba[rowStart+8*2+1] = '\n'
	}
	return ba[:]
}
