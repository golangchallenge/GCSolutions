package dlx

import (
	"testing"

	"github.com/slimmy/go-challenge-8/puzzle"
	"github.com/stretchr/testify/assert"
)

var (
	emptyBoard = `_ _ _ _ _ _ _ _ _
_ _ _ _ _ _ _ _ _
_ _ _ _ _ _ _ _ _
_ _ _ _ _ _ _ _ _
_ _ _ _ _ _ _ _ _
_ _ _ _ _ _ _ _ _
_ _ _ _ _ _ _ _ _
_ _ _ _ _ _ _ _ _
_ _ _ _ _ _ _ _ _
`
	solvedBoard = `1 2 3 4 5 6 7 8 9
4 5 6 7 8 9 1 2 3
7 8 9 1 2 3 4 5 6
2 3 4 5 6 7 8 9 1
5 6 7 8 9 1 2 3 4
8 9 1 2 3 4 5 6 7
3 4 5 6 7 8 9 1 2
6 7 8 9 1 2 3 4 5
9 1 2 3 4 5 6 7 8
`
)

func TestCoverEmptyBoard(t *testing.T) {
	board, err := puzzle.New([]byte(emptyBoard))
	assert.NotNil(t, board)
	assert.NoError(t, err)

	baseCover := baseCover()
	cover := exactCover(board)

	// Since we haven't introduced any new constraints (empty board) except for
	// the basic ones the exact cover should equal the base cover
	for i := range baseCover {
		for j := range baseCover[i] {
			assert.Equal(t, cover[i][j], baseCover[i][j])
		}
	}
}

func TestCoverSolvedBoard(t *testing.T) {
	board, err := puzzle.New([]byte(solvedBoard))
	assert.NotNil(t, board)
	assert.NoError(t, err)

	cover := exactCover(board)

	// For each cell = N (N in [1,9]) all arrays in the cover matrix at the cells
	// row/col should be all zeroes except for the Nth array.
	for i := 0; i < size; i++ {
		for j := 0; j < size; j++ {
			n := board.ValueAt(i, j)
			for num := 0; num < size; num++ {
				if num != n-1 {
					arr := cover[index(i, j, num)]
					for _, v := range arr {
						assert.Equal(t, 0, v)
					}
				}
			}
		}
	}
}

func TestDLXBoard(t *testing.T) {
	expectedSizes := []int{2, 1, 1}
	cover := [][]int{
		{0, 1, 0},
		{1, 0, 0},
		{1, 0, 1},
	}

	header := dlxBoard(cover)
	assert.Equal(t, 3, header.size)

	for c, i := header.right, 0; c != header.dancingNode; c, i = c.right, i+1 {
		assert.Equal(t, expectedSizes[i], c.size)
	}
}
