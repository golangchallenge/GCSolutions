package main

import "fmt"

var ErrUnableToSolve = fmt.Errorf("I am stupid")

type solver struct {
	g Grid

	stage    int
	maxStage int
	msg      string

	err    error
	solved bool
}

func newSolver(g Grid) *solver {
	s := solver{g: g, stage: 1}
	s.solved = g.IsSolved()
	return &s
}

func (s *solver) level() int {
	return s.maxStage
}

// step returns true if there is nothing else to do
func (s *solver) step() bool {
	if s.solved {
		return true
	}

	var nextG Grid
	switch s.stage {
	case 1:
		nextG = applyGroupRule(s.g, singles)
		s.msg = "singles"
	case 2:
		nextG = applyGroupRule(s.g, hiddenSingles)
		s.msg = "hidden singles"
	case 3:
		nextG = lockedCandidates1(s.g)
		s.msg = "locked candidates 1"
	case 4:
		nextG = lockedCandidates2(s.g)
		s.msg = "locked candidates 2"
	case 5:
		nextG = applyGroupRule(s.g, nakedPairs)
		s.msg = "naked pairs"
	case 6:
		nextG = applyGroupRule(s.g, hiddenPairs)
		s.msg = "hidden pairs"
	case 7:
		nextG = xWing(s.g)
		s.msg = "X-Wing"
	default:
		s.msg = ""
		s.err = ErrUnableToSolve
		return true
	}
	if s.stage > s.maxStage {
		s.maxStage = s.stage
	}
	// TODO: this is only for debugging, remove later
	if err := nextG.Validate(); err != nil {
		panic(fmt.Sprintf("invalid grid: %s:\n%s", err, s.g))
	}

	if nextG == s.g {
		s.stage++
		return s.step()
	} else {
		s.stage = 1
	}
	s.g = nextG
	s.solved = s.g.IsSolved()

	return s.solved
}

func (s *solver) solve() {
	for !s.step() {
	}
}

type groupRuleFunc func(g group)

func applyGroupRule(g Grid, rule groupRuleFunc) Grid {
	// rows
	for r := 0; r < GridSize; r++ {
		rule(g.row(r))
	}

	// columns
	for c := 0; c < GridSize; c++ {
		rule(g.column(c))
	}

	// regions
	for r := 0; r < GridSize; r++ {
		rule(g.region(r))
	}

	return g
}

// groupSingles analyzes each unresolved cell in a group and removes all
// candidates that have been already assigned to other cells.
func singles(g group) {
	for i := 0; i < GridSize; i++ {
		s := g.cell(i)
		for j := 0; j < GridSize; j++ {
			if i != j {
				if v := g.cell(j).value(); v != 0 {
					s.remove(v)
				}
			}
		}
	}
}

// hiddenSingles finds any hidden among other candidates singles.
//
// E.g. if non-single candidates for a group are
// 1. 1 9
// 2. 1 5 9
// 3. 1 5 6 9
// 4. 1 5 9
//
// then 6 can only be assigned to cell #3
func hiddenSingles(g group) {
	for d := byte(1); d <= GridSize; d++ {
		var cnt int
		var ii int
		for i := 0; i < GridSize; i++ {
			if g.cell(i).contains(d) {
				cnt += 1
				ii = i
			}
		}
		if cnt == 1 {
			g.cell(ii).resolve(d)
		}
	}
}

// lockedRowColCandidates finds candidates within a region that are restricted to
// one row or column. Since one of these cells must contain that specific
// candidate, the candidate can safely be excluded from the remaining cells in
// that row or column outside of the region.
func regionLockedCandidates1(g *Grid, rn int) {
	rgn := g.region(rn)
	for d := byte(1); d <= GridSize; d++ {
		var cnt int
		rr, cc := -1, -1
		for i := 0; i < GridSize; i++ {
			if rgn.cell(i).contains(d) {
				r, c := rgn.coord(i)
				cnt++
				if rr == -1 {
					// d is seen for the first time in row r
					rr = r
				} else if r != rr {
					// d is seen for the second/third/etc time but row is different
					rr = -2
				}
				if cc == -1 {
					// d is seen for the first time in column c
					cc = c
				} else if c != cc {
					// d is seen for the second/third/etc time but column is different
					cc = -2
				}
			}
		}
		if cnt == 2 || cnt == 3 {
			if rr >= 0 {
				// this means that all region d's are in the row rr
				// remove d from row(rr) if not in region r
				row := g.row(rr)
				for c := 0; c < GridSize; c++ {
					if !rgn.hasCell(rr, c) {
						row.cell(c).remove(d)
					}
				}
			}
			if cc >= 0 {
				// this means that all region d's are in the column cc
				// remove d from column(cc) if not in region r
				col := g.column(cc)
				for r := 0; r < GridSize; r++ {
					if !rgn.hasCell(r, cc) {
						col.cell(r).remove(d)
					}
				}
			}
		}
	}
}

func lockedCandidates1(g Grid) Grid {
	for i := 0; i < GridSize; i++ {
		regionLockedCandidates1(&g, i)
	}
	return g
}

// rowColLockedCandidates2 finds candidates within a row or column that are
// restricted to one region. Since one of these cells must contain that specific
// candidate, the candidate can safely be excluded from the remaining cells in
// the region.
func rowColLockedCandidates2(g *Grid, rc group) {
	for d := byte(1); d <= GridSize; d++ {
		var cnt int
		ii := -1
		for i := 0; i < GridSize; i++ {
			if rc.cell(i).contains(d) {
				cnt++
				if ii == -1 {
					// d is seen for the first time in group rc
					ii = i
				} else if i/3 != ii/3 {
					// d is seen for the second/third/etc time but region is different
					ii = -2
				}
			}
		}

		if ii >= 0 && cnt > 1 {
			r, c := rc.coord(ii)
			rgn := g.region(3*(r/3) + c/3)
			// this means that all rc's d's are in region rgn
			// so all d's from rgn not in rc can be removed
			for j := 0; j < GridSize; j++ {
				rr, cc := rgn.coord(j)
				if !rc.hasCell(rr, cc) {
					rgn.cell(j).remove(d)
				}
			}
		}
	}
}

func lockedCandidates2(g Grid) Grid {
	// rows
	for r := 0; r < GridSize; r++ {
		rowColLockedCandidates2(&g, g.row(r))
	}

	// columns
	for c := 0; c < GridSize; c++ {
		rowColLockedCandidates2(&g, g.column(c))
	}
	return g
}

// nakedPairs tries to find two cells in a group that contain an identical pair
// of candidates and only those two candidates, then no other cells in that
// group could be those values.
func nakedPairs(g group) {
	find := func(d1 byte, d2 byte, start int) int {
		for i := start; i < GridSize; i++ {
			if g.cell(i).containsOnly(d1, d2) {
				return i
			}
		}
		return -1
	}

	for d1 := byte(1); d1 <= GridSize; d1++ {
		for d2 := d1 + 1; d2 <= GridSize; d2++ {
			i1 := find(d1, d2, 0)
			i2 := find(d1, d2, i1+1)

			if i1 >= 0 && i2 >= 0 {
				// found naked pair
				for i := 0; i < GridSize; i++ {
					if i != i1 && i != i2 {
						g.cell(i).remove(d1)
						g.cell(i).remove(d2)
					}
				}
			}
		}
	}
}

// nakedPairs tries to find two cells in a group that contain a pair of
// candidates (hidden amongst other candidates) that are not found in any other
// cells in that group. The other candidates in those two cells can be excluded
// safely.
func hiddenPairs(g group) {
	type pair struct {
		d1 byte
		d2 byte
	}

	pos := [GridSize + 1][]int{}

	for i := 0; i < GridSize; i++ {
		for d := byte(1); d <= GridSize; d++ {
			if g.cell(i).contains(d) {
				pos[d] = append(pos[d], i)
			}
		}
	}

	for d1 := byte(1); d1 <= GridSize; d1++ {
		for d2 := d1 + 1; d2 <= GridSize; d2++ {
			p1 := pos[d1]
			p2 := pos[d2]
			if len(p1) == 2 && len(p2) == 2 && p1[0] == p2[0] && p1[1] == p2[1] {
				// we found a hidden pair d1, d2
				g.cell(p1[0]).set([]byte{d1, d2})
				g.cell(p1[1]).set([]byte{d1, d2})
			}
		}
	}
}

func positionsOf(g group, d byte) []int {
	var p []int
	for i := 0; i < GridSize; i++ {
		if g.cell(i).contains(d) {
			p = append(p, i)
		}
	}
	return p
}

// X-Wing rule (on rows only)
func xWing(g Grid) Grid {
	for d := byte(1); d <= GridSize; d++ {
		var cols []int
		r1 := -1
		r2 := -1
		for i := 0; i < GridSize; i++ {
			p := positionsOf(g.row(i), d)
			if len(p) == 2 {
				// d occurs only 2 times in row i
				if cols == nil {
					cols = p
					r1 = i
				} else {
					// this is the second time where d occurs only twice
					if p[0] == cols[0] && p[1] == cols[1] {
						// found x-wing
						r2 = i
						break
					}
				}
			}
		}
		if r1 >= 0 && r2 >= 0 {
			col1 := g.column(cols[0])
			col2 := g.column(cols[1])
			// remove d from col1 and col2
			for r := 0; r < GridSize; r++ {
				if r != r1 && r != r2 {
					col1.cell(r).remove(d)
					col2.cell(r).remove(d)
				}
			}
		}
	}

	return g
}
