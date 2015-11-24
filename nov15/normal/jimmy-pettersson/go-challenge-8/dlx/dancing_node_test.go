package dlx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVerticals(t *testing.T) {
	dn1 := newDancingNode(nil)
	dn2 := newDancingNode(nil)
	dn3 := newDancingNode(nil)

	/*
		dn1
		 |
		dn2
		 |
		dn3
	*/
	dn1.hookDown(dn2)
	dn2.hookDown(dn3)
	assert.True(t, dn1.down == dn2)
	assert.True(t, dn2.down == dn3)

	/*
		dn1
		 |
		dn3
	*/
	dn2.unlinkUpDown()
	assert.True(t, dn1.down == dn3)

	/*
		dn1
		 |
		dn2
		 |
		dn3
	*/
	dn2.relinkUpDown()
	assert.True(t, dn1.down == dn2)
	assert.True(t, dn2.down == dn3)
}

func TestHorizontals(t *testing.T) {
	dn1 := newDancingNode(nil)
	dn2 := newDancingNode(nil)
	dn3 := newDancingNode(nil)

	// dn1 - dn2 - dn3
	dn1.hookRight(dn2)
	dn2.hookRight(dn3)
	assert.True(t, dn1.right == dn2)
	assert.True(t, dn2.right == dn3)

	// dn1 - dn3
	dn2.unlinkLeftRight()
	assert.True(t, dn1.right == dn3)

	// dn1 - dn2 - dn3
	dn2.relinkLeftRight()
	assert.True(t, dn1.right == dn2)
	assert.True(t, dn2.right == dn3)
}
