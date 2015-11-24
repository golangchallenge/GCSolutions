package dlx

type dancingNode struct {
	*columnNode
	left  *dancingNode
	right *dancingNode
	up    *dancingNode
	down  *dancingNode
}

func newDancingNode(cn *columnNode) *dancingNode {
	dn := &dancingNode{}
	dn.columnNode = cn
	dn.left = dn
	dn.right = dn
	dn.up = dn
	dn.down = dn

	return dn
}

// "inserts" n below dn
func (dn *dancingNode) hookDown(n *dancingNode) *dancingNode {
	n.down = dn.down
	n.down.up = n
	n.up = dn
	dn.down = n

	return n
}

// "inserts" n to the right of dn
func (dn *dancingNode) hookRight(n *dancingNode) *dancingNode {
	n.right = dn.right
	n.right.left = n
	n.left = dn
	dn.right = n

	return n
}

// removes dn from the row
func (dn *dancingNode) unlinkLeftRight() {
	dn.left.right = dn.right
	dn.right.left = dn.left
}

// inserts dn into the row
func (dn *dancingNode) relinkLeftRight() {
	dn.left.right = dn
	dn.right.left = dn
}

// removes dn from the col
func (dn *dancingNode) unlinkUpDown() {
	dn.up.down = dn.down
	dn.down.up = dn.up
}

// inserts dn into the col
func (dn *dancingNode) relinkUpDown() {
	dn.up.down = dn
	dn.down.up = dn
}
