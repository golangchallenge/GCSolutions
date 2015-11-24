package main

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

type FakeSolve struct {
	FastSolve
}

func (s *FakeSolve) OneGuess(c Cell, t *testing.T) []Cell {
	err := s.SetCell(c)
	t.Logf("OneGuess(%p): cell: %v, err: %v", &s, c, err)
	if err != nil {
		return nil
	}
	cellSet, _ := s.Solve(t)
	t.Logf("solve: cell: %v, total count: %d", cellSet, s.cnt)
	if s.IsSolved() {
		cellSet = append(cellSet, c)
		return cellSet
	} else if nil != cellSet {
		//this is wrong way, so cleare all
		t.Logf("clear all cell: %v", cellSet)
		for _, cc := range cellSet {
			s.ClearCell(cc)
		}
	}
	t.Logf("guess wrong, clear cell: %v", c)
	s.ClearCell(c)
	return nil
}

func (s *FakeSolve) Bruteforce(t *testing.T) []Cell {
	if s.IsSolved() {
		return nil
	}
	boxID := s.BoxMinFreeCell(s.fb)
	//find candidates for first cell
	zeroMask := s.InvertBites(s.fb[boxID])
	t.Logf("zeroMask: %b", zeroMask)
	for _, i := range s.ConvBit2Slice(zeroMask) {
		x, y := s.b.XYBox(boxID, i)
		guessSet := s.CandidatSet(x, y, boxID)
		t.Logf("candidate to pos: %v (%d,%d)\n%s", guessSet, x, y, s.b.String())
		//launch goroutine with process this guess
		for _, gcell := range guessSet {
			cellSet := s.OneGuess(gcell, t)
			if nil != cellSet {
				t.Logf("bruteforce result: %v", cellSet)
				return cellSet
			}
		}
	}
	return nil
}

func (s *FakeSolve) Solve(t *testing.T) ([]Cell, error) {
	cellSet := []Cell{}
	var err error
	if s.IsSolved() {
		return nil, nil
	}
	round := 0
	for _, strategy := range []string{"box", "row", "column", "bruteforce"} {
		gotNew := true
		for gotNew {
			t.Logf("solver: %p, round: %d, strategy: %s", &s, round, strategy)
			cellS, errS := s.Round(strategy, t)
			t.Log(cellS)
			t.Log(errS)
			round++
			gotNew = false
			if nil != errS && len(errS) > 0 {
				err = fmt.Errorf("find new error(s) while round with strategy '%s': %v", strategy, errS)
				for _, errCur := range errS {
					t.Log(errCur.Error())
				}
			}
			if nil != cellS && len(errS) == 0 && len(cellS) > 0 {
				cellSet = append(cellSet, cellS...)
				s.SetCellSet(cellS)
				gotNew = true && !s.IsSolved()
				t.Logf("defined cells count: %d, new cells: %v, gotNew: %t", s.cnt, cellS, gotNew)
			}
		}
	}
	if !s.IsSolved() {
		err = fmt.Errorf("%s\npuzzles that cannot be solved, end with %d defined cells", err, s.cnt)
	}
	return cellSet, err
}

//Round use one of the strategy for each number [1..9], return finded cells is new cell was defined
func (s *FakeSolve) Round(strategy string, t *testing.T) ([]Cell, []error) {
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
			cellSet = append(cellSet, s.Bruteforce(t)...)
		}
		cSet, errSet := s.SetCellSet(cellSet)
		t.Logf("solver: %p, get: %v, set: %v\nerror: %v", &s, cellSet, cSet, errSet)
		cellSet = cSet
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
			t.Logf("solver: %p, number: %d, get: %v, set: %v\nerror: %v", &s, n, cellS, cSet, eSet)
			cellSet = append(cellSet, cSet...)
			errSet = append(errSet, eSet...)
		}
	}
	return cellSet, errSet
}

var s = &FakeSolve{FastSolve{b: &Board{}}}

var bruteBoard = `1 2 3 4 5 6 7 8 9
4 5 6 7 8 9 1 2 3
7 8 9 1 2 3 4 5 6
2 3 4 5 6 7 8 9 1
5 6 7 8 9 1 2 3 4
8 9 1 2 3 4 5 6 7
_ 4 5 _ 7 8 _ 1 2
6 7 8 _ 1 2 _ 4 5
_ 1 2 _ 4 5 _ 7 8`

func ReInit(s *FakeSolve) {
	s.b.Rest()
	s.b.Read(strings.NewReader(sBoard))
	s.init()
}

func TestNewSolver(t *testing.T) {
	ReInit(s)
	/*0 1 2   3 4 5   6 7 8
	0 1 0 3 | 0 0 6 | 0 8 0
	1 0 5 0 | 0 8 0 | 1 2 0
	2 7 0 9 | 1 0 3 | 0 5 6
	  ------+-------+-------
	3 0 3 0 | 0 6 7 | 0 9 0
	4 5 0 7 | 8 0 0 | 0 3 0
	5 8 0 1 | 0 3 0 | 5 0 7
	  ------+-------+-------
	6 0 4 0 | 0 7 8 | 0 1 0
	7 6 0 8 | 0 0 2 | 0 4 0
	8 0 1 2 | 0 4 5 | 0 7 8
	*/
	nm := [9]Number{
		Number{cnt: 6, r: 359, c: 207, b: 335}, //(6) rows: 359 (101100111), columns: 207 (11001111); boxes: 335 (101001111)
		Number{cnt: 3, r: 386, c: 164, b: 196}, //(3) rows: 386 (110000010), columns: 164 (10100100); boxes: 196 (11000100)
		Number{cnt: 5, r: 61, c: 182, b: 59},   //(5) rows: 61 (111101), columns: 182 (10110110); boxes: 59 (111011)
		Number{cnt: 3, r: 448, c: 146, b: 448}, //(3) rows: 448 (111000000), columns: 146 (10010010); boxes: 448 (111000000)
		Number{cnt: 5, r: 310, c: 227, b: 173}, //(5) rows: 310 (100110110), columns: 227 (11100011); boxes: 173 (10101101)
		Number{cnt: 4, r: 141, c: 305, b: 86},  //(4) rows: 141 (10001101), columns: 305 (100110001); boxes: 86 (1010110)
		Number{cnt: 6, r: 380, c: 437, b: 441}, //(6) rows: 380 (101111100), columns: 437 (110110101); boxes: 441 (110111001)
		Number{cnt: 7, r: 499, c: 445, b: 478}, //(7) rows: 499 (111110011), columns: 445 (110111101); boxes: 478 (111011110)
		Number{cnt: 2, r: 12, c: 132, b: 33},   //(2) rows: 12 (1100), columns: 132 (10000100); boxes: 33 (100001)
	}

	checkS := &FakeSolve{FastSolve{
		b:   &b,
		n:   nm,
		cnt: 41,
		fr:  [9]int{165, 210, 429, 178, 141, 341, 178, 165, 438},
		fc:  [9]int{181, 330, 437, 20, 362, 461, 34, 479, 292},
		fb:  [9]int{341, 340, 410, 362, 142, 338, 426, 422, 402},
	}}
	if !reflect.DeepEqual(s.b, checkS.b) {
		t.Errorf("expext: \n%s\nget: \n%s\n", checkS.b, s.b)
	}
	for i := 0; i < 9; i++ {
		if !reflect.DeepEqual(s.n[i], checkS.n[i]) {
			t.Errorf("for number: %d\nexpext: \n%s\nget: \n%s\n", i+1, checkS.n[i], s.n[i])
		}
		if s.fr[i] != checkS.fr[i] {
			t.Errorf("for row: %d expext: %d (%b) get: %d (%b)\n", i, checkS.fr[i], checkS.fr[i], s.fr[i], s.fr[i])
		}
		if s.fc[i] != checkS.fc[i] {
			t.Errorf("for column: %d expext: %d (%b) get: %d (%b)\n", i, checkS.fc[i], checkS.fc[i], s.fc[i], s.fc[i])
		}
		if s.fb[i] != checkS.fb[i] {
			t.Errorf("for box: %d expext: %d (%b) get: %d (%b)\n", i, checkS.fb[i], checkS.fb[i], s.fb[i], s.fb[i])
		}
	}
}

func TestInRow(t *testing.T) {
	testSet := []struct {
		n, y int
		res  bool
	}{
		{7, 2, true},
		{7, 4, true},
		{7, 5, true},
		{7, 0, false},
		{7, 7, false},
		{5, 2, true},
		{5, 3, false},
		{3, 5, true},
		{3, 1, false},
		{1, 2, true},
		{1, 3, false},
	}
	for _, test := range testSet {
		if s.IsInRow(test.n, test.y) != test.res {
			t.Errorf("is in row (%d) for number (%d), expect: %t get: %t", test.y, test.n, test.res, s.IsInRow(test.n, test.y))
		}
	}
}

func TestInColumn(t *testing.T) {
	testSet := []struct {
		n, x int
		res  bool
	}{
		{7, 2, true},
		{7, 1, false},
		{5, 1, true},
		{5, 3, false},
		{3, 5, true},
		{3, 3, false},
		{1, 2, true},
		{1, 4, false},
	}
	for _, test := range testSet {
		if s.IsInColumn(test.n, test.x) != test.res {
			t.Errorf("is in column (%d) for number (%d), expect: %t get: %t", test.x, test.n, test.res, s.IsInColumn(test.n, test.x))
		}
	}
}

func TestInBox(t *testing.T) {
	testSet := []struct {
		n, b int
		res  bool
	}{
		{7, 3, true},
		{7, 1, false},
		{5, 2, true},
		{5, 4, false},
		{3, 5, true},
		{3, 2, false},
		{1, 2, true},
		{1, 4, false},
	}
	for _, test := range testSet {
		if s.IsInBox(test.n, test.b) != test.res {
			t.Errorf("is in box (%d) for number (%d), expect: %t get: %t", test.b, test.n, test.res, s.IsInBox(test.n, test.b))
		}
	}
}

func TestLastInBox(t *testing.T) {
	testSet := []struct {
		boxID, n, mask int
	}{
		{0, 8, 128},
		{5, 8, 1},
		{2, 7, 1},
		{6, 7, 16},
	}
	for _, test := range testSet {
		boxFreeCellMask := s.LastInBox(test.boxID, test.n)
		if boxFreeCellMask != test.mask {
			t.Errorf("search for free cell in box (%d) for number (%d), expect: %b (%d) get: %b (%d)", test.boxID, test.n, test.mask, test.mask, boxFreeCellMask, boxFreeCellMask)
		}
	}
}

func TestLastInUnit(t *testing.T) {
	testSet := []struct {
		uc, ufm, i, n, mask int
	}{
		{0, 429, 2, 8, 2},
		{0, 210, 1, 7, 8},
		{0, 141, 4, 4, 352},
		{0, 178, 6, 2, 320},
		{1, 330, 1, 6, 48},
		{1, 20, 3, 5, 9},
		{1, 34, 6, 3, 448},
		{1, 292, 8, 1, 24},
	}
	for _, test := range testSet {
		maskF := s.LastInUnit(test.ufm, test.uc, test.i, test.n)
		if maskF != test.mask {
			t.Errorf("search for free cell in unit %d(%b) for number (%d), expect: %b (%d) get: %b (%d)", test.i, test.ufm, test.n, test.mask, test.mask, maskF, maskF)
		}
	}
}

func TestSSetCell(t *testing.T) {
	testSet := []Cell{
		Cell{1, 2, 8},
		Cell{6, 3, 8},
		Cell{1, 7, 7},
		Cell{6, 0, 7},
		Cell{3, 1, 7},
		Cell{4, 0, 1},
	}
	for _, test := range testSet {
		err := s.SetCell(test)
		if test.x != 4 && err != nil {
			t.Errorf("error: %s, while set %d at (%d,%d)\n", err.Error(), test.v, test.x, test.y)
		}
		if test.x == 4 && err == nil {
			t.Errorf("expect error: The value (1) is duplicated with one in row #0, while set %d at (%d,%d)\n", test.v, test.x, test.y)
		}
		if nil == err {
			s.ClearCell(test)
		}
	}
}

func TestSSetCellSet(t *testing.T) {
	testSet := []Cell{
		Cell{1, 2, 8},
		Cell{6, 3, 8},
		Cell{1, 7, 7},
		Cell{6, 0, 7},
		Cell{3, 1, 7},
		Cell{4, 0, 1},
	}
	setCell, errS := s.SetCellSet(testSet)
	if !reflect.DeepEqual(setCell, testSet[:5]) {
		t.Errorf("expect set 5 and seted %d", len(setCell))
	}
	if 1 != len(errS) {
		t.Errorf("expect 1 error: The value (1) is duplicated with one in row #0... but get: %b errors", len(errS))
	}
}

func TestClearCell(t *testing.T) {
	testSet := []Cell{
		Cell{1, 2, 8},
		Cell{6, 3, 8},
		Cell{1, 7, 7},
		Cell{6, 0, 7},
		Cell{3, 1, 7},
		Cell{4, 0, 1},
	}
	cnt := s.cnt - 5 //will be cleared only 5 cell
	for _, c := range testSet {
		s.ClearCell(c)
	}
	if cnt != s.cnt {
		t.Errorf("expect %d, but left %b", cnt, s.cnt)
	}
}

func TestStrategyLastInUnit(t *testing.T) {
	cellSet := s.StrategyLastInUnit(8, s.n[7].b, "box")

	for _, cell := range cellSet {
		if !(cell.x != 1 && cell.y != 2) && !(cell.x != 6 && cell.y != 3) {
			t.Errorf("expect coordinates: (1,2) or (6,3) get: (%d,%d)\n", cell.x, cell.y)
		}
	}
}

func TestCandidateSet(t *testing.T) {
	testSet := []struct {
		x, y, b int
		cs      []Cell
	}{
		{1, 0, 0, []Cell{Cell{1, 0, 2}}},
		{2, 1, 0, []Cell{Cell{2, 1, 4}, Cell{2, 1, 6}}},
		{6, 2, 2, []Cell{Cell{6, 2, 4}}},
		{1, 5, 3, []Cell{Cell{1, 5, 2}, Cell{1, 5, 6}, Cell{1, 5, 9}}},
		{4, 7, 7, []Cell{Cell{4, 7, 1}, Cell{4, 7, 9}}},
		{6, 7, 8, []Cell{Cell{6, 7, 3}, Cell{6, 7, 9}}},
	}
	for _, test := range testSet {
		cs := s.CandidatSet(test.x, test.y, test.b)
		if !reflect.DeepEqual(cs, test.cs) {
			t.Errorf("expect: %v; get: %v", test.cs, cs)
		}
	}
}

func TestBruteforce(t *testing.T) {
	s.b.Rest()
	s.b.Read(strings.NewReader(bruteBoard))
	s.init()
	cellSet := s.Bruteforce(t)
	testSet := []Cell{
		Cell{0, 8, 9},
		Cell{6, 8, 6},
		Cell{3, 7, 9},
		Cell{6, 6, 9},
		Cell{3, 8, 3},
		Cell{6, 7, 3},
		Cell{3, 6, 6},
		Cell{0, 6, 3},
	}
	if !reflect.DeepEqual(testSet, cellSet) {
		t.Errorf("bruteforce, expect: \n%v\nget: \n%v\n", testSet, cellSet)
	}
}

func TestOneGuess(t *testing.T) {
	s.b.Rest()
	s.b.Read(strings.NewReader(bruteBoard))
	s.init()
	testSet := []Cell{Cell{0, 8, 9}, Cell{6, 8, 6}, Cell{3, 7, 9}, Cell{6, 6, 9}, Cell{3, 8, 3}, Cell{6, 7, 3}, Cell{3, 6, 6}, Cell{0, 6, 3}, Cell{4, 7, 1}}
	cellSet := s.OneGuess(Cell{4, 7, 1}, t)
	if !reflect.DeepEqual(testSet, cellSet) {
		t.Errorf("oneguess, expect: \n%v\nget: \n%v\n", testSet, cellSet)
	}
}

func TestOneCandidat(t *testing.T) {
	ReInit(s)
	test := []Cell{
		Cell{1, 0, 2},
		Cell{4, 1, 9},
		Cell{5, 1, 4},
		Cell{1, 2, 8},
		Cell{3, 2, 2},
		Cell{6, 2, 4},
		Cell{5, 3, 1},
		Cell{4, 4, 2},
		Cell{7, 4, 6},
		Cell{5, 5, 9},
		Cell{1, 8, 9},
	}
	cellSet := s.OneCandidat()
	if !reflect.DeepEqual(test, cellSet) {
		t.Errorf("one candidat: expect \n %v\nget \n%v\n", test, cellSet)
	}
	for _, c := range test {
		s.ClearCell(c)
	}
}

func TestRound(t *testing.T) {
	testSet := []struct {
		strategy string
		cs       []Cell
		es       []error
	}{
		{"one", []Cell{
			Cell{1, 0, 2},
			Cell{4, 1, 9},
			Cell{5, 1, 4},
			Cell{1, 2, 8},
			Cell{3, 2, 2},
			Cell{6, 2, 4},
			Cell{5, 3, 1},
			Cell{4, 4, 2},
			Cell{7, 4, 6},
			Cell{5, 5, 9},
			Cell{1, 8, 9},
		}, []error{}},
		{"box", []Cell{
			Cell{4, 7, 1},
			Cell{8, 1, 3},
			Cell{3, 3, 5},
			Cell{2, 6, 5},
			Cell{8, 7, 5},
			Cell{2, 1, 6},
			Cell{6, 0, 7},
			Cell{1, 7, 7},
			Cell{1, 2, 8},
			Cell{6, 3, 8},
			Cell{8, 0, 9},
		}, []error{}},
		{"row", []Cell{
			Cell{4, 7, 1},
			Cell{8, 6, 2},
			Cell{6, 1, 3},
			Cell{4, 0, 5},
			Cell{8, 6, 5},
			Cell{1, 1, 6},
			Cell{7, 4, 6},
			Cell{3, 6, 6},
			Cell{6, 0, 7},
			Cell{1, 2, 8},
			Cell{6, 3, 8},
		}, []error{}},
		{"column", []Cell{
			Cell{4, 3, 1},
			Cell{0, 6, 3},
			Cell{6, 1, 3},
			Cell{2, 7, 5},
			Cell{3, 0, 5},
			Cell{1, 7, 7},
			Cell{3, 0, 7},
		}, []error{}},
	}
	for _, test := range testSet {
		ReInit(s)
		cellSet, errSet := s.Round(test.strategy, t)
		if !reflect.DeepEqual(test.cs, cellSet) || !reflect.DeepEqual(test.es, errSet) {
			t.Errorf("strategy: %s\nexpect \ncells: %v\nerr: %v\nget \ncells: %v\nerr: %v\n", test.strategy, test.cs, test.es, cellSet, errSet)
		}
		for _, c := range test.cs {
			s.ClearCell(c)
		}
	}

}

/*0 1 2   3 4 5   6 7 8
0 1 0 3 | 0 0 6 | 0 8 0
1 0 5 0 | 0 8 0 | 1 2 0
2 7 0 9 | 1 0 3 | 0 5 6
  ------+-------+-------
3 0 3 0 | 0 6 7 | 0 9 0
4 5 0 7 | 8 0 0 | 0 3 0
5 8 0 1 | 0 3 0 | 5 0 7
  ------+-------+-------
6 0 4 0 | 0 7 8 | 0 1 0
7 6 0 8 | 0 0 2 | 0 4 0
8 0 1 2 | 0 4 5 | 0 7 8
*/

func TestSolveBase(t *testing.T) {
	ReInit(s)
	_, err := s.Solve(t)
	if nil != err {
		t.Log(err.Error())
	}
	if !s.IsSolved() {
		t.Errorf("puzzle not solved!\n%s", s.b.String())
	}
}

func TestSolveEasy(t *testing.T) {
	s.b.Rest()
	//http://www.websudoku.com/?level=1&set_id=9645055255
	s.b.Read(strings.NewReader(`_ _ _ 4 _ 5 2 _ 3
_ 2 5 _ _ 8 _ 4 1
4 _ _ _ 7 _ _ _ _
_ _ 6 _ 3 4 _ _ 7
_ _ 1 5 _ 6 4 _ _
8 _ _ 1 9 _ 5 _ _
_ _ _ _ 5 _ _ _ 4
5 9 _ 3 _ _ 8 2 _
1 _ 4 6 _ 2 _ _ _`))
	s.init()
	_, err := s.Solve(t)
	if nil != err {
		t.Log(err.Error())
	}
	if !s.IsSolved() {
		t.Errorf("puzzle not solved!\n%s", s.b.String())
	}
}

func BenchmarkSolveBase(b *testing.B) {
	board := &Board{}
	board.Read(strings.NewReader(`1 _ 3 _ _ 3 _ 8 _
_ 5 _ _ 8 _ 1 2 _
7 _ 9 1 _ 3 _ 5 6
_ 3 _ _ 6 7 _ 9 _
5 _ 7 8 _ _ _ 3 _
8 _ 1 _ 3 _ 5 _ 7
_ 4 _ _ 7 8 _ 1 _
6 _ 8 _ _ 2 _ 4 _
_ 1 2 _ 4 5 _ 7 8`))
	for n := 0; n < b.N; n++ {
		s := &FastSolve{b: board}
		s.init()
		s.Solve()
	}
}

func BenchmarkSolveEasy(b *testing.B) {
	board := &Board{}
	board.Read(strings.NewReader(`_ _ _ 4 _ 5 2 _ 3
_ 2 5 _ _ 8 _ 4 1
4 _ _ _ 7 _ _ _ _
_ _ 6 _ 3 4 _ _ 7
_ _ 1 5 _ 6 4 _ _
8 _ _ 1 9 _ 5 _ _
_ _ _ _ 5 _ _ _ 4
5 9 _ 3 _ _ 8 2 _
1 _ 4 6 _ 2 _ _ _`))
	for n := 0; n < b.N; n++ {
		s := &FastSolve{b: board}
		s.init()
		s.Solve()
	}
}

/*
//algoritm need to be improved for solving this levels
func TestSolveMedium(t *testing.T) {
	s.b.Rest()
	//http://www.websudoku.com/?level=2&set_id=4505023030
	s.b.Read(strings.NewReader(`_ _ _ 2 _ _ 6 9 4
_ _ _ 9 _ 4 7 _ _
6 _ _ _ _ 5 _ 3 _
_ _ _ 3 2 _ 5 _ _
3 4 _ _ _ _ _ 6 7
_ _ 6 _ 9 7 _ _ _
_ 1 _ 7 _ _ _ _ 6
_ _ 2 1 _ 9 _ _ _
4 5 7 _ _ 2 _ _ _`))
	s.init()
	_, err := s.Solve(t)
	if nil != err {
		t.Log(err.Error())
	}
	if !s.IsSolved() {
		t.Errorf("puzzle not solved!\n%s", s.b.String())
	}
}

func TestSolveHard(t *testing.T) {
	s.b.Rest()
	//http://www.websudoku.com/?level=3&set_id=5771150657
	s.b.Read(strings.NewReader(`5 _ _ 7 _ _ _ _ _
_ 1 _ 9 2 _ _ _ _
_ _ 3 _ 6 _ _ _ 9
_ 6 1 5 _ _ _ _ _
_ 7 9 _ _ _ 6 4 _
_ _ _ _ _ 4 1 2 _
8 _ _ _ 1 _ 9 _ _
_ _ _ _ 3 9 _ 5 _
_ _ _ _ _ 7 _ _ 4`))
	s.init()
	_, err := s.Solve(t)
	if nil != err {
		t.Log(err.Error())
	}
	if !s.IsSolved() {
		t.Errorf("puzzle not solved!\n%s", s.b.String())
	}
}
*/
