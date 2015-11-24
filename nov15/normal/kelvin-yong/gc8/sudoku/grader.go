package sudoku

// The grader grades the level of difficulty of a sudoku puzzle by first
// solving it with human strategies. Based on statistics it generates from
// the solving attempt, a difficult level is obtained
import (
	"fmt"
	"reflect"
)

// Indicates the difficulty level of a puzzle
const (
	LevelAny = iota
	LevelEasy
	LevelMedium
	LevelHard
	LevelEvil
)

// Prepare a sudoku puzzle for solving via human-like strategies
func (s *Sudoku) prepareForHuman() {
	s.remaining = 81
	s.hb = &humanBoard{grid: make([]square, 81), newlySolved: nil}

	for i, n := range s.puzzle {
		if n >= 1 && n <= 9 {
			s.hb.grid[i].num = n
			s.hb.grid[i].candidates = nil
			s.hb.newlySolved = append(s.hb.newlySolved, i)

			s.remaining--
		} else {
			s.hb.grid[i].num = n
			s.hb.grid[i].candidates = make(map[byte]bool)
			for k := 1; k <= 9; k++ {
				s.hb.grid[i].candidates[byte(k)] = true
			}
		}
	}
}

// statistics records how many times each human strategy is
// successfully applied, how many strategies were used.
// Also the number of empty squares at the start and number
// of empty squares at the end
type statistics struct {
	Applied              []int
	Used                 int
	EmptyStart, EmptyEnd int
}

// SolveHuman attempts to solve puzzle using 4 human-like strategies.
// The strategies are Naked Singles, Hidden Singles, Locked Types
// and Naked Pairs. Returns the estimated difficulty level.
func (s *Sudoku) SolveHuman() int {
	all := []strategy{stNakedSingle, stHiddenSingle, stLockedType, stNakedPair}
	stats := s.solveWith(all)
	return gradeDifficulty(stats)
}

// solveWith will attempt to solve puzzle using the strategies indicated
//
// For example, given 4 strategies (A, B, C, D), (which the first should always
// be nakedSingle), use the first one.
//
// If it succeed, go back to the first. If it did failed, go to the next.
// If you are at the last strategy, you are out of luck.
/*
	         (f)        (f)        (f)        (f)
	[start] -----> A -------> B -------> C -------> D -------> [stop]
	               ↑    |(s)       |(s)       |(s)       |(s)
	               |    ↓          ↓          ↓          ↓
	          <----------<----------<----------<----------
*/
// Returns the statistics of solving
func (s *Sudoku) solveWith(strategies []strategy) statistics {
	s.prepareForHuman()
	maxCount := len(strategies)

	stats := statistics{}
	stats.Applied = make([]int, maxCount)
	stats.EmptyStart = s.remaining
	stIndex := 0

	for {
		st := strategies[stIndex]
		removed, solved := st(s.hb)

		s.remaining -= solved
		if s.remaining == 0 {
			break
		}

		strategyWorked := removed != 0 || solved != 0
		if strategyWorked {
			stats.Applied[stIndex]++
			// always go back to the simplest strategy
			stIndex = 0
			continue
		}

		if stIndex == maxCount-1 {
			// out of luck
			break
		} else {
			// go the next strategy
			stIndex++
		}
	}

	stats.EmptyEnd = s.remaining

	if s.remaining == 0 {
		for i := 0; i < 81; i++ {
			s.solution[i] = s.hb.grid[i].num
		}
	}

	for _, a := range stats.Applied {
		if a != 0 {
			stats.Used++
		}
	}
	return stats
}

// gradeDifficulty assigns a difficulty level based on statistics generated
// when solving using human strategies. The values are somewhat based on
// 25 puzzles x 4 levels from websudoku.com
func gradeDifficulty(stats statistics) int {
	if stats.EmptyEnd != 0 {
		// can't solve. must be real tough
		return LevelEvil
	}

	if stats.Used == 1 {
		if stats.Applied[0] <= 13 {
			// used only naked single and for 13 rounds or less
			return LevelEasy
		}
		// used only naked single and for more than 13 rounds
		return LevelMedium
	}

	if stats.Used == 2 {
		// used 2 strategies, but the second strategy was used less than
		// 10 rounds and the board has less than 52 unknowns initially
		if stats.Applied[1] <= 10 && stats.EmptyStart <= 52 {
			return LevelMedium
		}
		return LevelHard
	}

	if stats.Used == 3 {
		// required 3 strategies... must be hard
		return LevelHard
	}

	// stats.Used == 4
	return LevelEvil
}

// ========================
// Strategies
// ========================
type strategy func(hb *humanBoard) (removed, solved int)

// Naked Single Strategy: For each newly solved square with a number n,
// eliminate n as a potential candidate from all the square's peers
var stNakedSingle = func(hb *humanBoard) (removed, solved int) {
	nowSolved := []int{}

	for _, id := range hb.newlySolved {
		n := hb.grid[id].num
		row, col, box := rcbForIndex(id)
		peerIDs := make([]int, 0, 27)
		peerIDs = append(peerIDs, peersForRow[row][0:]...)
		peerIDs = append(peerIDs, peersForCol[col][0:]...)
		peerIDs = append(peerIDs, peersForBox[box][0:]...)

		for _, pid := range peerIDs {
			if hb.grid[pid].num != 0 {
				continue
			}
			r, s := hb.eliminate(n, pid)
			removed, solved = removed+r, solved+s
			if s == 1 {
				nowSolved = append(nowSolved, pid)
			}
		}
	}
	hb.newlySolved = nowSolved

	return
}

var peerTypes = [3]*peersArray{&peersForRow, &peersForCol, &peersForBox}

// Hidden Single Strategy: 	Loop thru each square, check if each individual
// candidate of the square is unique for a group (i.e. is not a candidate of the
// peers in of the same row, or same col, or same box.)
// If so, assign that candidate to be the square's number. Stop processing and
// eliminate that candidate from all other peers (using nakedSingle)
var stHiddenSingle = func(hb *humanBoard) (removed, solved int) {
	nowSolved := []int{}

OUTLOOP:
	for i := 0; i < 81; i++ {
		if hb.grid[i].num != 0 {
			continue
		}

		row, col, box := rcbForIndex(i)

		var checkUnique = func(c byte, myID int, peerIDs []int) bool {
			u := true
			for _, pID := range peerIDs {
				if myID == pID {
					continue
				}
				if _, found := hb.grid[pID].candidates[c]; found {
					u = false
					break
				}
			}
			return u
		}

		var solveSquare = func(id int, n byte) {
			hb.grid[id].num = n
			hb.grid[id].candidates = nil
			nowSolved = append(nowSolved, id)
			solved++
		}

		for c := range hb.grid[i].candidates {
			var isUnique bool
			isUnique = checkUnique(c, i, peersForRow[row][0:])
			if isUnique {
				solveSquare(i, c)
				break OUTLOOP
			}

			isUnique = checkUnique(c, i, peersForCol[col][0:])
			if isUnique {
				solveSquare(i, c)
				break OUTLOOP
			}

			isUnique = checkUnique(c, i, peersForBox[box][0:])
			if isUnique {
				solveSquare(i, c)
				break OUTLOOP
			}
		}
	}

	hb.newlySolved = nowSolved
	return
}

var stNakedPair = func(hb *humanBoard) (removed, solved int) {
	nowSolved := []int{}

	var eliminateNakedPair = func(peers []int, aid, bid int) {
		for _, id := range peers {
			if id == aid || id == bid {
				continue
			}
			for n := range hb.grid[aid].candidates {
				r, s := hb.eliminate(n, id)
				removed, solved = removed+r, solved+s
				if s == 1 {
					nowSolved = append(nowSolved, id)
				}
			}
		}
	}

	var findNakedPair = func(peers []int) {
		for i, aid := range peers {
			if len(hb.grid[aid].candidates) != 2 {
				continue
			}
			for _, bid := range peers[i+1:] {
				if reflect.DeepEqual(hb.grid[aid].candidates, hb.grid[bid].candidates) {
					eliminateNakedPair(peers, aid, bid)
					break
				}
			}
		}
	}

	for r := 0; r < 3; r++ {
		pt := peerTypes[r]

		for i := 0; i < 9; i++ {
			findNakedPair(pt[i])
		}
	}

	hb.newlySolved = nowSolved
	return
}

// If in a block all candidates of a certain digit are confined to a row or
// column, that digit cannot appear outside of that block in that row or column.
var stLockedType = func(hb *humanBoard) (removed, solved int) {
	nowSolved := []int{}

	checkBox := func(b int) {
		o := peersForBox[b][0]
		m := make(map[byte][]int)
		for _, id := range peersForBox[b] {
			if hb.grid[id].num != 0 {
				continue
			}
			for c := range hb.grid[id].candidates {
				m[c] = append(m[c], id)
			}
		}

		for n, where := range m {
			if len(where) > 3 {
				continue
			}
			r, c, _ := rcbForIndex(where[0])
			var peers []int
			switch {
			case subset(where, []int{o + 0, o + 1, o + 2}):
				peers = peersForRow[r]
			case subset(where, []int{o + 9, o + 10, o + 11}):
				peers = peersForRow[r]
			case subset(where, []int{o + 18, o + 19, o + 20}):
				peers = peersForRow[r]
			case subset(where, []int{o + 0, o + 9, o + 18}):
				peers = peersForCol[c]
			case subset(where, []int{o + 1, o + 10, o + 19}):
				peers = peersForCol[c]
			case subset(where, []int{o + 2, o + 11, o + 20}):
				peers = peersForCol[c]
			}

			for _, id := range peers {
				if !subset([]int{id}, where) {
					r, s := hb.eliminate(n, id)
					removed, solved = removed+r, solved+s
					if s == 1 {
						nowSolved = append(nowSolved, id)
					}
				}
			}
		}
	}

	for b := 0; b < 9; b++ {
		checkBox(b)
	}

	hb.newlySolved = nowSolved
	return
}

// ========================
// Utility methods
// ========================

// Eliminate a number, n, as a potential candidate from an identified square
// of the human board. If the square then has only 1 candidates, it is solved.
func (hb *humanBoard) eliminate(n byte, id int) (removed, solved int) {
	_, found := hb.grid[id].candidates[n]
	if found {
		removed++
		delete(hb.grid[id].candidates, n)

		// check if the peer square contains only 1 candidate.
		// if so, the peer square is solved!
		if len(hb.grid[id].candidates) == 1 {
			for k := range hb.grid[id].candidates {
				hb.grid[id].num = k
			}
			hb.grid[id].candidates = nil
			solved++
		}
	}
	return
}

// For a particular id, return the row, col and box
func rcbForIndex(i int) (row, col, box int) {
	row = i / 9
	col = i - row*9
	box = row/3*3 + col/3
	return
}

// subset returns true if the first array is completely
// contained in the second array. There must be at least
// the same number of duplicate values in second as there
// are in first.
func subset(first, second []int) bool {
	set := make(map[int]int)
	for _, value := range second {
		set[value]++
	}

	for _, value := range first {
		if count, found := set[value]; !found {
			return false
		} else if count < 1 {
			return false
		} else {
			set[value] = count - 1
		}
	}

	return true
}

// for debugging purposes
func printHumanBoard(g []square) {
	for r, i := 0, 0; r < 9; r, i = r+1, i+9 {
		fmt.Printf("%d %d %d | %d %d %d | %d %d %d\n",
			g[i].num, g[i+1].num, g[i+2].num,
			g[i+3].num, g[i+4].num, g[i+5].num,
			g[i+6].num, g[i+7].num, g[i+8].num)
		if r == 2 || r == 5 {
			fmt.Println("------+-------+------")
		}
	}
}
