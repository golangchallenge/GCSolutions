package main

import "fmt"

//Solver tool for chacking sudoku quiz
type Solver interface {
	String() string
	IsInRow(n, y int) bool
	IsInColumn(n, x int) bool
	IsInBox(n, b int) bool
	IsSolved() bool
	LastInBox(b, n int) int
	LastInUnit(m, c, i, n int) int
	StringBoard() string
	Solve() ([]Cell, error)
}

//FastSolve agrigate all elements for finding optimal solution for the task
type FastSolve struct {
	Box
	Binary
	b          *Board
	n          [9]Number
	fr, fc, fb [9]int //filled units
	cnt        int
}

//NewSolver return instance of Solver with initials parameters
func NewSolver(b *Board) Solver {
	s := &FastSolve{b: b}
	s.init()
	return s
}

//init - initialize statistics data for solving puzzle
func (s *FastSolve) init() {
	s.n = [9]Number{}
	s.fr, s.fc, s.fb = [9]int{}, [9]int{}, [9]int{}
	s.cnt = 0
	for i := range s.b.cell {
		for j := range s.b.cell[i] {
			if s.b.cell[i][j] == 0 {
				continue
			}
			s.AddStat(j, i, s.b.cell[i][j])
		}
	}
}

//AddStat process statistics data when added new cell
func (s *FastSolve) AddStat(x, y, v int) {
	s.cnt++
	boxID := s.b.BoxID(x, y)
	s.n[v-1].r |= (1 << uint(y))
	s.n[v-1].c |= (1 << uint(x))
	s.n[v-1].b |= (1 << uint(boxID))
	s.n[v-1].cnt++
	s.fr[y] |= (1 << uint(x))
	s.fc[x] |= (1 << uint(y))
	s.fb[boxID] |= (1 << uint((y%3)*3+(x%3)))
}

//RallbackStat remove data about number from stat
func (s *FastSolve) RallbackStat(x, y, n int) {
	s.cnt--
	boxID := s.b.BoxID(x, y)
	s.n[n-1].r &^= (1 << uint(y))
	s.n[n-1].c &^= (1 << uint(x))
	s.n[n-1].b &^= (1 << uint(boxID))
	s.n[n-1].cnt--
	s.fr[y] &^= (1 << uint(x))
	s.fc[x] &^= (1 << uint(y))
	s.fb[boxID] &^= (1 << uint((y%3)*3+(x%3)))
}

//String convert all solver struct to a string
func (s FastSolve) String() (st string) {
	st = s.b.String()
	for i, nStat := range s.n {
		st += fmt.Sprintf("%d - %s", i+1, nStat)
	}
	st += fmt.Sprintf("Known number count: %d\n", s.cnt)
	return
}

//StringBoard display only board, as required at the task
func (s FastSolve) StringBoard() string {
	return s.b.String()
}

//IsInRow check is n exists in row y (use binary mask)
func (s FastSolve) IsInRow(n, y int) bool {
	return s.n[n-1].r>>uint(y)&0x01 == 1
}

//IsInColumn check is n exists in column x (use binary mask)
func (s FastSolve) IsInColumn(n, x int) bool {
	return s.n[n-1].c>>uint(x)&0x01 == 1
}

//IsInBox check is n exists in box x (use binary mask)
func (s FastSolve) IsInBox(n, b int) bool {
	return s.n[n-1].b>>uint(b)&0x01 == 1
}

//IsSolved check is the values of all cells were defined
func (s FastSolve) IsSolved() bool {
	return 81 == s.cnt
}

//LastInBox define which cell free and may contains our number result as binary mask
func (s FastSolve) LastInBox(boxID, n int) (boxFreeCellMask int) {
	x, y := s.b.FirstXYBox(boxID)
	boxFreeCellMask = (s.fb[boxID] ^ BinaryMask9)
	for i := y; i < y+3; i++ {
		mask := BinaryMask3
		if s.IsInRow(n, i) {
			boxFreeCellMask &^= (mask << uint((i-y)*3))
		}
	}
	for j := x; j < x+3; j++ {
		mask := BinaryMask12
		if s.IsInColumn(n, j) {
			boxFreeCellMask &^= (mask << uint(j%3))
		}
	}
	return
}

//LastInUnit check for possible to set number n to any place in unit (row or column)
//return the mask (byte with value 1 is mean the possible position)
func (s FastSolve) LastInUnit(unitFullMask, unitCode, unitID, n int) (unitFreeMask int) {
	if unitCode == 0 && s.IsInRow(n, unitID) || unitCode == 1 && s.IsInColumn(n, unitID) {
		return 0
	}
	unitFreeMask = (unitFullMask ^ BinaryMask9)
	for i := 0; i < 9; i++ {
		if unitCode == 0 && s.IsInColumn(n, i) || unitCode == 1 && s.IsInRow(n, i) {
			unitFreeMask &^= (1 << uint(i))
		}
	}
	boxID := -1
	for i := 0; i < 3; i++ {
		if unitCode == 0 {
			boxID = (unitID/3)*3 + i
		} else if unitCode == 1 {
			boxID = i*3 + unitID/3
		}
		mask := BinaryMask3
		if s.IsInBox(n, boxID) {
			unitFreeMask &^= (mask << uint(i*3))
		}
	}
	return
}

//SetCell use with binary mask checks and set value deractly to the board
func (s *FastSolve) SetCell(c Cell) error {
	switch {
	case c.v < 0 || c.v > 9:
		return fmt.Errorf("The value out of range [1..9] or _ separated by whitespace: %d", c.v)
	case c.v > 0 && c.v == s.b.cell[c.y][c.x]:
		return nil
	case c.v > 0 && s.IsInRow(c.v, c.y):
		return fmt.Errorf("The value (%d) is duplicated with one in row #%d", c.v, c.y)
	case c.v > 0 && s.IsInColumn(c.v, c.x):
		return fmt.Errorf("The value (%d) is duplicated with one in column #%d", c.v, c.x)
	case c.v > 0 && s.IsInBox(c.v, s.b.BoxID(c.x, c.y)):
		return fmt.Errorf("The value (%d) is duplicated with one in box #%d", c.v, s.b.BoxID(c.x, c.y))
	default:
		s.b.cell[c.y][c.x] = c.v
		s.AddStat(c.x, c.y, c.v)
		return nil
	}
}

//SetCellSet call SetCell for each element
func (s *FastSolve) SetCellSet(cSet []Cell) ([]Cell, []error) {
	errSet := []error{}
	setedCellSet := []Cell{}
	for _, c := range cSet {
		err := s.SetCell(c)
		if nil == err {
			setedCellSet = append(setedCellSet, c)
		} else {
			errSet = append(errSet, err)
		}
	}
	return setedCellSet, errSet
}

//ClearCell - delete number from sell and statistics
func (s *FastSolve) ClearCell(c Cell) {
	if 0 == s.b.cell[c.y][c.x] {
		return
	}
	s.b.cell[c.y][c.x] = 0
	s.RallbackStat(c.x, c.y, c.v)
}

//StrategyLastInUnit  - return slice of number n is any row, column, box
//that not defined at the board
func (s *FastSolve) StrategyLastInUnit(n, unitNumMas int, unitAlias string) (cellSet []Cell) {
	//define which units has not number n
	unitFreeSet := s.InvertBites(unitNumMas)

	for _, unitID := range s.ConvBit2Slice(unitFreeSet) {
		unitFreeCellMask := 0
		getXY := func(uID, pos int) (int, int) {
			return uID, pos //as for column
		}
		switch unitAlias {
		case "box":
			unitFreeCellMask = s.LastInBox(unitID, n)
			getXY = func(uID, pos int) (int, int) {
				x, y := s.FirstXYBox(uID)
				x, y = x+pos%3, y+pos/3
				return x, y
			}
		case "row":
			unitFreeCellMask = s.LastInUnit(s.fr[n-1], 0, unitID, n)
			getXY = func(uID, pos int) (int, int) {
				return pos, uID
			}
		case "column":
			unitFreeCellMask = s.LastInUnit(s.fc[n-1], 1, unitID, n)
			getXY = func(uID, pos int) (int, int) {
				return uID, pos
			}
		}
		unitFreeCellSet := s.ConvBit2Slice(unitFreeCellMask)
		if len(unitFreeCellSet) == 1 {
			x, y := getXY(unitID, unitFreeCellSet[0])
			c := Cell{x, y, n}
			err := s.SetCell(c)
			if err == nil {
				cellSet = append(cellSet, Cell{x, y, n})
			}
		}
	}
	return
}

//CandidatSet find all possible values for cell
func (s FastSolve) CandidatSet(x, y, boxID int) (cellSet []Cell) {
	for n := 1; n < 10; n++ {
		if !s.IsInBox(n, boxID) && !s.IsInRow(n, y) && !s.IsInColumn(n, x) {
			cellSet = append(cellSet, Cell{x, y, n})
		}
	}
	return cellSet
}

//OneCandidat check each empty cell for possible values, if only one possible use it
func (s *FastSolve) OneCandidat() (cellSet []Cell) {
	for i := 0; i < 9; i++ {
		for j := 0; j < 9; j++ {
			if 0 != s.b.cell[j][i] {
				continue
			}
			cs := s.CandidatSet(j, i, s.BoxID(j, i))
			if len(cs) == 1 {
				err := s.SetCell(cs[0])
				if nil == err {
					cellSet = append(cellSet, cs[0])
				}
			}
		}
	}
	return
}

//Bruteforce - find all free cell with candidates to it, cheack each of them
//if one of it will route us to the solution - the others will be ignored
func (s *FastSolve) Bruteforce() []Cell {
	boxID := s.BoxMinFreeCell(s.fb)
	//find candidates for first cell
	zeroMask := s.InvertBites(s.fb[boxID])
	for _, i := range s.ConvBit2Slice(zeroMask) {
		x, y := s.b.XYBox(boxID, i)
		guessSet := s.CandidatSet(x, y, boxID)
		//launch goroutine with process this guess
		for _, gcell := range guessSet {
			cellSet := s.OneGuess(gcell)
			if nil != cellSet {
				return cellSet
			}
		}
	}
	return nil
}

//OneGuess choose one value and try find solution
func (s *FastSolve) OneGuess(c Cell) []Cell {
	err := s.SetCell(c)
	if err != nil {
		return nil
	}
	cellSet, _ := s.Solve()
	if s.IsSolved() {
		cellSet = append(cellSet, c)
		return cellSet
	} else if nil != cellSet {
		//this is wrong way, so cleare all
		for _, cc := range cellSet {
			s.ClearCell(cc)
		}
	}
	s.ClearCell(c)
	return nil
}

//Round use one of the strategy for each number [1..9], return finded cells is new cell was defined
func (s *FastSolve) Round(strategy string) ([]Cell, []error) {
	if s.IsSolved() {
		return nil, nil
	}
	errSet := []error{}
	cellSet := []Cell{}

	switch strategy {
	case "one", "bruteforce":
		switch strategy {
		case "one":
			cellSet = append(cellSet, s.OneCandidat()...)
		case "bruteforce":
			cellSet = append(cellSet, s.Bruteforce()...)
		}
		cellSet, errSet = s.SetCellSet(cellSet)
	case "box", "row", "column":
		for n := 1; n < 10; n++ {
			cellS := []Cell{}
			switch strategy {
			case "box":
				cellS = s.StrategyLastInUnit(n, s.n[n-1].b, strategy)
			case "row":
				cellS = s.StrategyLastInUnit(n, s.n[n-1].r, strategy)
			case "column":
				cellS = s.StrategyLastInUnit(n, s.n[n-1].c, strategy)
			}
			cSet, eSet := s.SetCellSet(cellS)
			cellSet = append(cellSet, cSet...)
			errSet = append(errSet, eSet...)
		}
	}
	return cellSet, errSet
}

//Solve run all strategy, not call SetCell, but call the method that do that
func (s *FastSolve) Solve() ([]Cell, error) {
	if s.IsSolved() {
		return nil, nil
	}
	cellSet := []Cell{}
	var err error
	for _, strategy := range []string{"box", "row", "column", "bruteforce"} {
		gotNew := true
		for gotNew {
			cellS, errS := s.Round(strategy)
			gotNew = false
			if nil != errS && len(errS) > 0 {
				err = fmt.Errorf("find new error(s) while round with strategy '%s': %v", strategy, errS)
			}
			if nil != cellS && len(errS) == 0 && len(cellS) > 0 {
				cellSet = append(cellSet, cellS...)
				gotNew = true && !s.IsSolved()
				s.SetCellSet(cellS)
			}
		}
	}
	if !s.IsSolved() {
		err = fmt.Errorf("%s\npuzzles that cannot be solved, end with %d defined cells", err, s.cnt)
	}
	return cellSet, err
}
