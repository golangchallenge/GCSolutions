package sudoku

type Node struct {
	Left  *Node
	Right *Node
	Up    *Node
	Down  *Node

	Head *Node

	Col int
	Row int
}

func nodeInit(row int, col int) *Node {
	n := Node{}
	n.Left = &n
	n.Right = &n
	n.Up = &n
	n.Down = &n
	n.Row = row
	n.Col = col
	return &n
}

func NodeRegular(row int, col int) *Node {
	return nodeInit(row, col)
}

func NodeHeader(col int) *Node {
	return nodeInit(0, col)
}

func (n *Node) Cover() {
	if n.Head != nil {
		n.Up.Down = n.Down
		n.Down.Up = n.Up
		return
	}
	n.Left.Right = n.Right
	n.Right.Left = n.Left
}
