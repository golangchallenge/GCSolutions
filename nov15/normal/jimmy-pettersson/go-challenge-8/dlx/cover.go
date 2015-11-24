package dlx

import (
	"strconv"

	"github.com/slimmy/go-challenge-8/puzzle"
)

const (
	size = 9
	side = 3
)

// dlxBoard builds the matrix of nodes given the cover to solve for.
// Returns the starting header node
func dlxBoard(cover [][]int) *columnNode {
	cols := len(cover[0])
	rows := len(cover)

	headerNode := newColumnNode("header")
	columnNodes := []*columnNode{}

	for i := 0; i < cols; i++ {
		n := newColumnNode(strconv.Itoa(i))
		columnNodes = append(columnNodes, n)
		headerNode = headerNode.hookRight(n.dancingNode).columnNode
	}

	headerNode = headerNode.right.columnNode

	for i := 0; i < rows; i++ {
		var prev *dancingNode

		for j := 0; j < cols; j++ {
			if cover[i][j] == 1 {
				cn := columnNodes[j]
				dn := newDancingNode(cn)
				if prev == nil {
					prev = dn
				}

				cn.up.hookDown(dn)
				prev = prev.hookRight(dn)
				cn.size++
			}
		}
	}

	headerNode.size = cols

	return headerNode
}

// exactCover takes the basic cover and updates it with extra
// contraints for the numbers we have been given through the board
func exactCover(board *puzzle.Board) [][]int {
	base := baseCover()

	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			n := board.ValueAt(i, j)
			if n != 0 {
				for v := 0; v < size; v++ {
					if v != n-1 {
						arr := base[index(i, j, v)]
						for k := range arr {
							arr[k] = 0
						}
					}
				}
			}
		}
	}

	return base
}

// baseCover builds the basic exact cover grid for a Sudoku
func baseCover() [][]int {
	grid := make([][]int, size*size*size)
	for i := range grid {
		grid[i] = make([]int, size*size*4)
	}

	var hb int

	// row-col constraints
	for r := 0; r < size; r++ {
		for c := 0; c < size; c, hb = c+1, hb+1 {
			for n := 0; n < size; n++ {
				grid[index(r, c, n)][hb] = 1
			}
		}
	}

	// row-num constraints
	for r := 0; r < size; r++ {
		for n := 0; n < size; n, hb = n+1, hb+1 {
			for c := 0; c < size; c++ {
				grid[index(r, c, n)][hb] = 1
			}
		}
	}

	// col-num constraints
	for c := 0; c < size; c++ {
		for n := 0; n < size; n, hb = n+1, hb+1 {
			for r := 0; r < size; r++ {
				grid[index(r, c, n)][hb] = 1
			}
		}
	}

	// zone-num constraints
	for zr := 0; zr < size; zr += side {
		for zc := 0; zc < size; zc += side {
			for n := 0; n < size; n, hb = n+1, hb+1 {
				for rd := 0; rd < side; rd++ {
					for cd := 0; cd < side; cd++ {
						grid[index(zr+rd, zc+cd, n)][hb] = 1
					}
				}
			}
		}
	}

	return grid
}

func index(row, col, val int) int {
	return row*size*size + col*size + val
}
