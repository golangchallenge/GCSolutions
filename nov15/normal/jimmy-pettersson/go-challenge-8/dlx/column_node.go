package dlx

type columnNode struct {
	*dancingNode
	size int
	name string
}

func newColumnNode(name string) *columnNode {
	cn := &columnNode{
		size: 0, // nbr of 1s in the column
		name: name,
	}
	cn.dancingNode = newDancingNode(cn)

	return cn
}

// cover removes a column (cn) from the matrix of nodes.
// It also removes all rows in the column from other columns the are part of
func (cn *columnNode) cover() {
	cn.unlinkLeftRight()

	for i := cn.down; i != cn.dancingNode; i = i.down {
		for j := i.right; j != i; j = j.right {
			j.unlinkUpDown()
			j.size--
		}
	}
}

// uncover performs the opposite operation of cover() to allow backtracking
func (cn *columnNode) uncover() {
	for i := cn.up; i != cn.dancingNode; i = i.up {
		for j := i.left; j != i; j = j.left {
			j.size++
			j.relinkUpDown()
		}
	}

	cn.relinkLeftRight()
}
