package puzzle

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateBoard(t *testing.T) {
	var tests = []struct {
		diffuculty string
		shouldErr  bool
	}{
		{"easy", false},
		{"MeDiUm", false},
		{"HARD", false},
		{"bad", true},
		{"", true},
	}

	for _, test := range tests {
		b, err := Generate(test.diffuculty)

		if test.shouldErr {
			assert.Nil(t, b)
			assert.Error(t, err)
		} else {
			assert.NotNil(t, b)
			assert.NoError(t, err)
			assert.Equal(t, strings.ToLower(test.diffuculty), b.Difficulty())
		}
	}
}

func TestCellsToFill(t *testing.T) {
	for i := 0; i < 100; i++ {
		e := cellsToFill(Easy)
		assert.True(t, e >= 32)
		assert.True(t, e <= 36)

		m := cellsToFill(Medium)
		assert.True(t, m >= 28)
		assert.True(t, m <= 31)

		h := cellsToFill(Hard)
		assert.True(t, h >= 26)
		assert.True(t, h <= 27)

		// should default to easy
		d := cellsToFill(-1)
		assert.True(t, d >= 32)
		assert.True(t, d <= 36)
	}
}
