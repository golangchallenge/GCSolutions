// Credits to Sonia Keys
// Modified from https://github.com/soniakeys/dlx-sudoku

// Package dlx implements Dancing Links. Dancing links, also known as DLX,
// is the technique suggested by Donald Knuth to efficiently implement
// his Algorithm X.
//
// Algorithm X is a recursive, nondeterministic, depth-first, backtracking
// algorithm that finds all solutions to the exact cover problem.
// Some of the better-known exact cover problems include tiling,
// the n queens problem, and Sudoku.
package dlx

import (
	"math/rand"
	"time"
)

// Knuth's data object
type x struct {
	c          *y
	u, d, l, r *x
	// except x0 is not Knuth's.  it's pointer to first constraint in row,
	// so that the sudoku string can be constructed from the dlx solution.
	x0 *x
	// An additional field row id is used to which row makes up the dlx solution
	rID int
}

// Knuth's column object
type y struct {
	x
	s int // size
	n int // name
}

// DLX represents a generic Dancing Link structure
type DLX struct {
	ch []y  // all column headers
	h  *y   // ch[0], the root node
	o  []*x // working solution

	Solutions [][]int // Each solution is a list of row IDs
}

// NewDLX returns a new DLX with the number of columns (constraints)
func NewDLX(nCols int) *DLX {
	ch := make([]y, nCols+1)
	h := &ch[0]
	d := &DLX{ch, h, nil, [][]int{}}
	h.c = h
	h.l = &ch[nCols].x
	ch[nCols].r = &h.x
	nh := ch[1:]
	for i := range ch[1:] {
		hi := &nh[i]
		ix := &hi.x
		hi.n = i
		hi.c = hi
		hi.u = ix
		hi.d = ix
		hi.l = &h.x
		h.r = ix
		h = hi
	}

	return d
}

// AddRow adds a row to the to DLX. A row is specified with a row ID
// and a slice of int values to represent the columns (constraints)
// which this row satisfies
func (d *DLX) AddRow(rID int, nr []int) {
	if len(nr) == 0 {
		return
	}
	r := make([]x, len(nr))
	x0 := &r[0]

	for x, j := range nr {
		ch := &d.ch[j+1]
		ch.s++
		np := &r[x]
		np.c = ch
		np.u = ch.u
		np.d = &ch.x
		np.l = &r[(x+len(r)-1)%len(r)]
		np.r = &r[(x+1)%len(r)]
		np.u.d, np.d.u, np.l.r, np.r.l = np, np, np, np
		np.x0 = x0
		np.rID = rID
	}
}

// ValidRows returns the row IDs of the rows that makes up the solution
func (d *DLX) ValidRows() []int {
	vr := make([]int, len(d.o))
	for i, r := range d.o {
		vr[i] = r.rID
	}
	return vr
}

// Solve starts to search for solutions, stopping when the required number
// of solution(s) is found or when there is nothing  more to search.
// You should always pass isRandom = false, unless you believe that there are
// multiple solutions and you want a random solution.
func (d *DLX) Solve(maxSolution int, isRandom bool) {
	if maxSolution < 1 {
		return
	}
	d.search(maxSolution, isRandom)
}

// Search starts the dancing steps of DLX, and attempts to find a solution
func (d *DLX) search(maxSolution int, isRandom bool) bool {
	h := d.h
	j := h.r.c
	if j == h {
		d.Solutions = append(d.Solutions, d.ValidRows())
		return len(d.Solutions) == maxSolution
	}

	// S heuristic (choosing the column with the fewest rows)
	c, minS := j, j.s

	for {
		if minS == 1 {
			break
		}
		j = j.r.c
		if j == h {
			break
		}
		if j.s < minS {
			c, minS = j, j.s
		}
	}

	// if you want a truly random column, find out which columns has the
	// same minS, then pick from one of these columns randomly
	if isRandom {
		minCols := []*y{}
		for {
			j = j.r.c
			if j == h {
				break
			}
			if j.s == minS {
				minCols = append(minCols, j)
			}
		}
		if len(minCols) == 0 {
			minCols = append(minCols, c)
		}
		c = minCols[rand.Intn(len(minCols))]
	}

	cover(c)
	k := len(d.o)
	d.o = append(d.o, nil)
	for r := c.d; r != &c.x; r = r.d {
		d.o[k] = r
		for j := r.r; j != r; j = j.r {
			cover(j.c)
		}
		if d.search(maxSolution, isRandom) {
			return true
		}
		r = d.o[k]
		c = r.c
		for j := r.l; j != r; j = j.l {
			uncover(j.c)
		}
	}
	d.o = d.o[:len(d.o)-1]
	uncover(c)
	return false
}

func cover(c *y) {
	c.r.l, c.l.r = c.l, c.r
	for i := c.d; i != &c.x; i = i.d {
		for j := i.r; j != i; j = j.r {
			j.d.u, j.u.d = j.u, j.d
			j.c.s--
		}
	}
}

func uncover(c *y) {
	for i := c.u; i != &c.x; i = i.u {
		for j := i.l; j != i; j = j.l {
			j.c.s++
			j.d.u, j.u.d = j, j
		}
	}
	c.r.l, c.l.r = &c.x, &c.x
}

func init() {
	rand.Seed(time.Now().UnixNano())
}
