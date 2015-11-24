package sudoku

const (
	GridSize        = 9
	RegionSize      = 3
	ConstraintTypes = 4
	NumOfCells      = GridSize * GridSize
	Constraints     = NumOfCells * ConstraintTypes
	Possibilities   = GridSize * GridSize * GridSize
)

func generateDlxMatrix() *Node {
	rows, cols := generateUnassociatedNodes()

	header := headerRow()
	associateVertical(header.Right, cols)
	associateHorizontal(rows)
	return header
}

func associateHorizontal(rows [Possibilities][ConstraintTypes]*Node) {
	for _, nodes := range rows {
		n := len(nodes)
		for i, node := range nodes {
			node.Left = nodes[(i-1+n)%n]
			node.Right = nodes[(i+1)%n]
		}
	}
}

func associateVertical(headers *Node, cols [Constraints][GridSize]*Node) {
	header := headers
	for _, nodes := range cols {
		n := len(nodes)
		for i, node := range nodes {
			//We need pointer to col header from every cell (for efficiency purposes)
			node.Head = header
			node.Up = nodes[(i-1+n)%n]
			node.Down = nodes[(i+1)%n]
		}
		attachHeader(header, nodes[0])
		header = header.Right
	}
}

func attachHeader(header *Node, to *Node) {
	header.Up = to.Up
	header.Down = to
	to.Up.Down = header
	to.Up = header
}

func generateUnassociatedNodes() ([Possibilities][ConstraintTypes]*Node, [Constraints][GridSize]*Node) {
	cols := [Constraints][GridSize]*Node{}
	rows := [Possibilities][ConstraintTypes]*Node{}

	rowCnt := [Possibilities]int{}
	colCnt := [Constraints]int{}

	for row := 0; row < GridSize; row++ {
		for col := 0; col < GridSize; col++ {
			for val := 0; val < GridSize; val++ {
				pos := encodePossibility(row, col, val)
				for _, con := range constraintPositions(row, col, val) {
					n := NodeRegular(pos, con)
					rows[pos][rowCnt[pos]] = n
					cols[con][colCnt[con]] = n
					rowCnt[pos]++
					colCnt[con]++
				}
			}
		}
	}
	return rows, cols
}

func constraintPositions(row int, col int, val int) []int {
	box := getBoxNumber(row, col)
	return []int{
		row*GridSize + col,
		NumOfCells + row*GridSize + val,
		NumOfCells*2 + col*GridSize + val,
		NumOfCells*3 + box*GridSize + val}
}

func headerRow() *Node {
	root := NodeHeader(-1)
	last := header(root, 0)
	root.Left = last
	last.Right = root
	return root
}

func header(node *Node, n int) *Node {
	if n == Constraints {
		return node
	}
	node.Right = NodeHeader(n)
	node.Right.Left = node
	return header(node.Right, n+1)
}

func encodePossibility(row int, col int, val int) int {
	return NumOfCells*row + col*GridSize + val
}

func decodePosibility(encoded int) (int, int, int) {
	row := encoded / NumOfCells
	col := (encoded % NumOfCells) / GridSize
	val := (encoded % NumOfCells) % GridSize
	return row, col, val
}

func getBoxNumber(row int, col int) int {
	return (row/RegionSize)*RegionSize + col/RegionSize
}

func getColumnHeader(start *Node, index int) *Node {
	for ptr := start.Right; ptr != start; ptr = ptr.Right {
		if ptr.Col == index {
			return ptr
		}
	}
	return nil
}
