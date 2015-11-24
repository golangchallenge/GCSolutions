package dlx

import (
	"strconv"

	"github.com/slimmy/go-challenge-8/puzzle"
)

// Solver implementes the SudokuSolver interface
type Solver struct {
	board    *puzzle.Board
	header   *columnNode
	solution []*dancingNode
	solved   bool
}

// NewSolver returns a sudoku solver that implements the Dancing Links algorithm
func NewSolver(board *puzzle.Board) *Solver {
	return &Solver{
		board:    board,
		header:   dlxBoard(exactCover(board)),
		solution: []*dancingNode{},
	}
}

// Solve tries to solve the sudoku using Knuths' Dancing Links algorithm
func (s *Solver) Solve() *puzzle.Board {
	s.dlx()
	return s.board
}

// dlx tries to find a solution to the exact cover
func (s *Solver) dlx() {
	if s.header.right == s.header.dancingNode {
		s.buildSolvedBoard()
		s.solved = true
		return
	}

	c := s.selectColumnNode()
	c.cover()

	for r := c.down; r != c.dancingNode; r = r.down {
		s.solution = append(s.solution, r)

		for j := r.right; j != r; j = j.right {
			j.cover()
		}

		s.dlx()

		// dlx finds all solutions but we're only interested
		// in the first one
		if s.solved {
			return
		}

		i := len(s.solution) - 1
		r = s.solution[i]
		// delete last element from s.solution
		s.solution, s.solution[i] = append(s.solution[:i], s.solution[i+1:]...), nil
		c = r.columnNode

		for j := r.left; j != r; j = j.left {
			j.uncover()
		}
	}

	c.uncover()
}

// Returns the column with the least number of nodes
func (s *Solver) selectColumnNode() *columnNode {
	var min = int(^uint(0) >> 1) // Max int
	var ret *columnNode

	for c := s.header.right.columnNode; c != s.header; c = c.right.columnNode {
		if c.size < min {
			ret, min = c, c.size
		}
	}

	return ret
}

// buildSolvedBoard builds the solved *puzzle.Board from the solution slice
func (s *Solver) buildSolvedBoard() {
	board := &puzzle.Board{}

	for _, n := range s.solution {
		rcNode := n
		min, _ := strconv.Atoi(rcNode.name)

		for tmp := n.right; tmp != n; tmp = tmp.right {
			val, _ := strconv.Atoi(tmp.name)
			if val < min {
				rcNode, min = tmp, val
			}
		}

		pos, _ := strconv.Atoi(rcNode.name)
		num, _ := strconv.Atoi(rcNode.right.name)
		row, col, val := pos/size, pos%size, (num%size)+1

		board.SetValue(row, col, val)
	}

	s.board = board
}
