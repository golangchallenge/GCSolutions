package puzzle

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	// Valid input
	testInput1 = `1 _ 3 _ _ 6 _ 8 _
_ 5 _ _ 8 _ 1 2 _
7 _ 9 1 _ 3 _ 5 6
_ 3 _ _ 6 7 _ 9 _
5 _ 7 8 _ _ _ 3 _
8 _ 1 _ 3 _ 5 _ 7
_ 4 _ _ 7 8 _ 1 _
6 _ 8 _ _ 2 _ 4 _
_ 1 2 _ 4 5 _ 7 8
`
	// Invalid "A" at row 0, col 1
	testInput2 = `1 A 3 _ _ 6 _ 8 _
_ 5 _ _ 8 _ 1 2 _
7 _ 9 1 _ 3 _ 5 6
_ 3 _ _ 6 7 _ 9 _
5 _ 7 8 _ _ _ 3 _
8 _ 1 _ 3 _ 5 _ 7
_ 4 _ _ 7 8 _ 1 _
6 _ 8 _ _ 2 _ 4 _
_ 1 2 _ 4 5 _ 7 8
`
	// Double 1s on row 0
	testInput3 = `1 1 3 _ _ 6 _ 8 _
_ 5 _ _ 8 _ 1 2 _
7 _ 9 1 _ 3 _ 5 6
_ 3 _ _ 6 7 _ 9 _
5 _ 7 8 _ _ _ 3 _
8 _ 1 _ 3 _ 5 _ 7
_ 4 _ _ 7 8 _ 1 _
6 _ 8 _ _ 2 _ 4 _
_ 1 2 _ 4 5 _ 7 8
`
	// Solved
	testInput4 = `1 2 3 4 5 6 7 8 9
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

func TestNew(t *testing.T) {
	b, err := New([]byte(testInput1))
	assert.NoError(t, err)
	assert.NotNil(t, b)

	b, err = New([]byte(testInput2))
	assert.Error(t, err)
	assert.Nil(t, b)

	b, err = New([]byte(testInput3))
	assert.Error(t, err)
	assert.Nil(t, b)
}

func TestValueAt(t *testing.T) {
	b, err := New([]byte(testInput1))
	assert.NoError(t, err)
	assert.NotNil(t, b)

	assert.Equal(t, 1, b.ValueAt(0, 0))
	assert.Equal(t, 0, b.ValueAt(0, 1))
	assert.Equal(t, 8, b.ValueAt(8, 8))
}

func TestAllowed(t *testing.T) {
	b, err := New([]byte(testInput1))
	assert.NoError(t, err)
	assert.NotNil(t, b)

	assert.True(t, b.Allowed(0, 1, 2))
	assert.True(t, b.Allowed(8, 6, 3))
	assert.True(t, b.Allowed(8, 6, 6))
	assert.True(t, b.Allowed(8, 6, 9))

	assert.False(t, b.Allowed(0, 0, 1))
	assert.False(t, b.Allowed(0, 1, 5))
	assert.False(t, b.Allowed(8, 6, 8))
}

func TestSolved(t *testing.T) {
	b, err := New([]byte(testInput1))
	assert.NoError(t, err)
	assert.NotNil(t, b)
	assert.False(t, b.Solved())

	b, err = New([]byte(testInput4))
	assert.NoError(t, err)
	assert.NotNil(t, b)
	assert.True(t, b.Solved())
}

func TestSetAndCopy(t *testing.T) {
	b, err := New([]byte(testInput1))
	assert.NoError(t, err)
	assert.NotNil(t, b)

	bb := b.SetAndCopy(0, 1, 2)
	assert.Equal(t, 0, b.ValueAt(0, 1))
	assert.Equal(t, 2, bb.ValueAt(0, 1))
}

func TestRowColZone(t *testing.T) {
	b, err := New([]byte(testInput4))
	assert.NoError(t, err)
	assert.NotNil(t, b)

	assert.Equal(t, []int{1, 2, 3, 4, 5, 6, 7, 8, 9}, b.row(0))
	assert.Equal(t, []int{5, 8, 2, 6, 9, 3, 7, 1, 4}, b.col(4))
	assert.Equal(t, []int{9, 1, 2, 3, 4, 5, 6, 7, 8}, b.zone(8))
}
