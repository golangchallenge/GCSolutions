package sudoku

type Solver struct {
	solution []int
	header   *Node
}

func InitSolver(input string) (*Solver, error) {
	if len(input) != NumOfCells {
		return nil, InputTooShortError
	}
	solver := Solver{header: generateDlxMatrix()}
	fixedCells := asciiToInts(input)
	if err := solver.addToSolution(fixedCells); err != nil {
		return nil, err
	}
	return &solver, nil
}

func (s *Solver) Solve() bool {
	col := s.header.Right
	if s.header.Right == s.header {
		return true
	}
	if col.Down == col {
		return false
	}
	cover(col)
	for row := col.Down; row != col; row = row.Down {
		for cell := row.Right; cell != row; cell = cell.Right {
			cover(cell.Head)
		}
		if s.Solve() {
			s.solution = append(s.solution, row.Row)
			return true
		}
		for cell := row.Left; cell != row; cell = cell.Left {
			uncover(cell.Head)
		}
	}
	uncover(col)
	return false
}

func (s *Solver) GetSolution() string {
	result := make([]byte, NumOfCells)
	for _, v := range s.solution {
		row, col, val := decodePosibility(v)
		result[row*GridSize+col] = intToAscii(val) + 1
	}
	return string(result)
}

func (s *Solver) addToSolution(fixedCells []int) error {
	for i, val := range fixedCells {
		if val == 0 {
			continue
		}

		row := i / GridSize
		col := i % GridSize

		constraints := constraintPositions(row, col, val-1)
		for _, columnIdx := range constraints {
			col := getColumnHeader(s.header, columnIdx)
			if col == nil {
				return InputMalformedError
			}
			cover(col)
		}
		s.solution = append(s.solution, encodePossibility(row, col, val-1))
	}
	return nil
}

func cover(col *Node) {
	col.Cover()
	coverRows(col)
}

func uncover(header *Node) {
	uncoverRows(header)
	uncoverCol(header)
}

func coverRows(header *Node) {
	for ptr := header.Down; ptr != header; ptr = ptr.Down {
		for cell := ptr.Right; cell != ptr; cell = cell.Right {
			cell.Cover()
		}
	}
}

func uncoverRows(header *Node) {
	for ptr := header.Up; ptr != header; ptr = ptr.Up {
		for cell := ptr.Left; cell != ptr; cell = cell.Left {
			uncoverRow(cell)
		}
	}
}

func uncoverCol(header *Node) {
	header.Left.Right = header
	header.Right.Left = header
}

func uncoverRow(cell *Node) {
	cell.Up.Down = cell
	cell.Down.Up = cell
}
