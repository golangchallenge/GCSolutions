package main

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"os"
	"bufio"
)


type Solver struct {
	done      bool
	elim      Matrix
	solutions []Matrix
}

/* Define Matrix struct and it's methods */
type Matrix struct {
	rows, cols int
	mtrx       [9][9]int
}
type ParsedMatrix struct {
	rows, cols, blocks *[9]string
}
type IterateFn func(r, c int, a *Matrix, d []interface{})
type IterateFnBlocks func(r, c int, a *Matrix, b int, d []interface{})
type IterateFnGet func(r, c int) int
type IterateFnSet func(r, c, v int) int

func NewMatrix(toM string) Matrix {
	const r = 9
	const c = 9
	return Matrix{
		rows: r,
		cols: c,
		mtrx: ParseInput(toM),
	}
}
func (a *Matrix) At(r, c int) int {
	return a.mtrx[r][c]
}
func (a *Matrix) Set(r, c, value int) {
	a.mtrx[r][c] = value
}
func (m *Matrix) IterateMtrxByRows(fn IterateFn, d ...interface{}) {
	// fmt.Println("Start IterateMtrxByRows")
	for r := 0; r < 9; r++ {
		for c := 0; c < 9; c++ {
			fn(r, c, m, d)
		}
	}
}
func (m *Matrix) IterateMtrxByCols(fn IterateFn, d ...interface{}) {
	for c := 0; c < 9; c++ {
		for r := 0; r < 9; r++ {
			fn(r, c, m, d)
		}
	}
}
func (m *Matrix) IterateMtrxRows(fn IterateFn, d ...interface{}) {
	for r := 0; r < 9; r++ {
		fn(r, 0, m, d)
	}
}
func (m *Matrix) IterateMtrxCols(fn IterateFn, d ...interface{}) {
	for c := 0; c < 9; c++ {
		fn(0, c, m, d)
	}
}
func (m *Matrix) IterateMtrxByBlocks(fn IterateFnBlocks, d ...interface{}) {
	for rgs := 0; rgs < 9; rgs++ {
		for r := int(rgs/3) * 3; r < int(rgs/3)*3+3; r++ {
			for c := (rgs + 3) % 3 * 3; c < ((rgs+3)%3*3)+3; c++ {
				fn(r, c, m, rgs, d)
				// fmt.Println("Row:", r, ";", "Col:", c)
			}
		}
	}
}
func (m *Matrix) printMatrix() {
	for i := 0; i < m.rows; i++ {
		fmt.Println(m.mtrx[i])
	}
}

/* Blocks definition */
var blocks = [9][9]int{
	{0, 0, 0, 1, 1, 1, 2, 2, 2},
	{0, 0, 0, 1, 1, 1, 2, 2, 2},
	{0, 0, 0, 1, 1, 1, 2, 2, 2},
	{3, 3, 3, 4, 4, 4, 5, 5, 5},
	{3, 3, 3, 4, 4, 4, 5, 5, 5},
	{3, 3, 3, 4, 4, 4, 5, 5, 5},
	{6, 6, 6, 7, 7, 7, 8, 8, 8},
	{6, 6, 6, 7, 7, 7, 8, 8, 8},
	{6, 6, 6, 7, 7, 7, 8, 8, 8},
}

func isdigit(b byte) bool {
	if '0' <= b && b <= '9' {
		return true
	} else {
		return false
	}
}

func ParseInput(s string) [9][9]int {
	var arr_out [9][9]int

	ns := strings.Replace(s, "_", "0", -1)
	ns = strings.Replace(ns, " ", "", -1)
	ns = strings.Replace(ns, "\n", "", -1)
	if len(ns) != 81 {panic("Invalid input, wrong number of cells")}
	for i,_:= range arr_out {
		for j,_:= range arr_out[i] {
			if !isdigit(ns[i*9+j]) {
				fmt.Println("ns[i*9+j]-'0':", ns[i*9+j]-'0')
				panic("Invalid input, found not digit")
			}
			arr_out[i][j] = int(ns[i*9+j]-'0')
		}
	}
	return arr_out
}

func parseRows(m *Matrix) *[9]string {
	result := [9]string{}
	template := "123456789"
	for c := 0; c < 9; c++ {
		columnParsed := template
		for r := 0; r < 9; r++ {
			if m.mtrx[r][c] != 0 {
				t := strconv.FormatInt(int64(m.mtrx[r][c]), 10)
				columnParsed = strings.Replace(columnParsed, t, "", -1)
			}
		}
		result[c] = columnParsed
	}
	return &result
}

func parseCols(m *Matrix) *[9]string {
	result := [9]string{}
	template := "123456789"

	for r := 0; r < 9; r++ {
		columnParsed := template
		for c := 0; c < 9; c++ {
			if m.mtrx[r][c] != 0 {
				t := strconv.FormatInt(int64(m.mtrx[r][c]), 10)
				columnParsed = strings.Replace(columnParsed, t, "", -1)
			}
		}
		result[r] = columnParsed
	}
	return &result
}

func parseToBlocks(m *Matrix) *[9]string {
	result := [9]string{}
	three := "123456789"
	rep := 0
	for rgs := 0; rgs < 9; rgs++ {
		threeParsed := three
		for r := int(rgs/3) * 3; r < int(rgs/3)*3+3; r++ {
			for c := (rgs + 3) % 3 * 3; c < ((rgs+3)%3*3)+3; c++ {
				t := strconv.FormatInt(int64(m.mtrx[r][c]), 10)
				threeParsed = strings.Replace(threeParsed, t, "", -1)
				rep++
			}
		}
		result[rgs] = threeParsed
	}
	return &result
}

/* Define Puzzle utils */
func reduceStrings(row string, col string, tri string) *string {
	result := ""
	alphai := [9]string{"1", "2", "3", "4", "5", "6", "7", "8", "9"}
	for i := 0; i < 9; i++ {
		if strings.Contains(row, alphai[i]) &&
			strings.Contains(col, alphai[i]) &&
			strings.Contains(tri, alphai[i]) {
			result += alphai[i]
		}
	}
	return &result
}

type trySolveStruct struct {
	numZeros                             *int
	solvedBool                           *bool
	parsedRows, parsedCols, parsedBlocks *[9]string
}

/* Define funcs to solve puzzle */
/*
@r: 	current row
@c: 	current column
@a: 	Pointer to Matrix struct
@b int: 	current puzzle block in iteration
*/
func trySolver(r, c int, a *Matrix, b int, d []interface{}) {
	trystruct := d[0].(*trySolveStruct)
	if a.mtrx[r][c] == 0 {
		ts := reduceStrings(trystruct.parsedRows[r], trystruct.parsedCols[c],
			trystruct.parsedBlocks[b])
		// fmt.Printf("%s",ts)
		if len(*ts) == 1 {
			to, _ := strconv.ParseInt(*ts, 10, 0)
			a.mtrx[r][c] = int(to)
			*trystruct.numZeros--
		} else {
			*trystruct.solvedBool = false
		}
	} else {
		*trystruct.numZeros--
	}
}

/*
Solve puzzle by cells with known numbers
*/
func trySolve(m *Matrix) (bool, int) {
	numZeros := 81
	solvedBool := true
	m.IterateMtrxByBlocks(trySolver, &trySolveStruct{
		numZeros:     &numZeros,
		solvedBool:   &solvedBool,
		parsedRows:   parseCols(m),
		parsedCols:   parseRows(m),
		parsedBlocks: parseToBlocks(m),
	})
	return solvedBool, numZeros
}

type parseCellsStruct struct {
	parsedRows, parsedCols, parsedBlocks *[9]string
	bigArray                             *[9][9]string
}

func parseCells(r, c int, a *Matrix, b int, d []interface{}) {
	parsestruct := d[0].(*parseCellsStruct)
	if a.mtrx[r][c] == 0 {
		ts := reduceStrings(parsestruct.parsedRows[r], parsestruct.parsedCols[c],
			parsestruct.parsedBlocks[b])
		parsestruct.bigArray[r][c] = *ts
	}
}
func solveCells(m *Matrix) *[9][9]string {
	// parsedCols := parseRows(m)
	// parsedRows := parseCols(m)
	// parsedBlocks := parseToBlocks(m)
	bigArray := [9][9]string{}
	m.IterateMtrxByBlocks(parseCells, &parseCellsStruct{
		parsedRows:   parseCols(m),
		parsedCols:   parseRows(m),
		parsedBlocks: parseToBlocks(m),
		bigArray:     &bigArray,
	})
	for i := range bigArray {
		for j := range bigArray[i] {
			fmt.Printf("\t%s\t", bigArray[i][j])
			if j == 8 {
				fmt.Println("")
			} else {
				fmt.Printf(" | ")
			}
		}
		if i == 2 || i == 5 || i == 8 {
			fmt.Println("")
		}
	}
	return &bigArray
}

func parseStringToIntArray(s string) [2]int {
	result := [2]int{}
	for i, _ := range s {
		iout, err := strconv.ParseInt(string(s[i]), 10, 0)
		if err == nil {
			result[i] = int(iout)
		} else {
			panic(fmt.Sprintf("Error in parseStringToIntArray()"))
		}
	}
	return result
}

type consistentStruct struct {
	test      *map[int]struct{}
	consFlag  *bool
	direction bool
}

/*
@r: 	current row
@c: 	current column
@a: 	Pointer to Matrix struct
*/
func isConsistent(r, c int, a *Matrix, d []interface{}) {
	// fmt.Println("Start isConsistent")
	value := 0
	consStruct := d[0].(*consistentStruct)
	// if a.mtrx[r][c] > 0 {
	// 	value = a.mtrx[r][c] - 1
	// } else {
	value = a.mtrx[r][c]
	// }
	_, ok := (*consStruct.test)[value]
	if ok && value != 0 {
		// fmt.Println("Repeated in row:", r,c)
		// fmt.Println("Index:",value)
		// fmt.Println("test[value]:",(d[0].(*[9]int))[value]+1)
		*consStruct.consFlag = false
	}
	(*consStruct.test)[value] = struct{}{}
	if c == 8 && consStruct.direction == false {
		*consStruct.test = make(map[int]struct{})
	} else if r == 8 && consStruct.direction == true {
		*consStruct.test = make(map[int]struct{})
	}
	// fmt.Println("End isConsistent",*(d[0].(*[9]int)))
}
func testForConsistency(m *Matrix) bool {
	// fmt.Println("Start testForConsistent")
	test := make(map[int]struct{})
	testConsistent := true
	direction := false
	//resBool := true
	m.IterateMtrxByRows(isConsistent, &consistentStruct{
		test:      &test,
		consFlag:  &testConsistent,
		direction: direction,
	} /*&test, &testConsistent, direction*/)
	test = make(map[int]struct{})
	direction = true
	m.IterateMtrxByCols(isConsistent, &consistentStruct{
		test:      &test,
		consFlag:  &testConsistent,
		direction: direction,
	} /*&test, &testConsistent, direction*/)
	return testConsistent
}

func testForRepeats(m *Matrix, a *[]Matrix, coords *[][2]int) bool {
	aM := *a
	crds := *coords
	for i := 1; i < len(aM); i++ {
		repeats := 0
		for j := 0; j < len(crds); j++ {
			if m.mtrx[crds[j][0]][crds[j][1]] == aM[i].mtrx[crds[j][0]][crds[j][1]] {
				repeats += 1
			}
		}
		if repeats == len(crds) {

			//fmt.Println("Found repeated decision. Go further")
			return false
		}
	}
	return true
}

func findZeros(m *Matrix) int {
	zeros := 0
	for i, _ := range m.mtrx {
		for j, _ := range m.mtrx[i] {
			if m.mtrx[i][j] == 0 {
				zeros += 1
			}
		}
	}
	return zeros
}

func resolveCell(depthWrong, numToSearch int, coords [2]int,
	m *Matrix, results *Cache, wg *sync.WaitGroup) {
	// fmt.Println("Start resolver wrong inside")
	zM := *m
	// fmt.Println("Resolve wrong at:",(coords), "numToSearch:", numToSearch, "depth:", depthWrong)
	
	zM.mtrx[coords[0]][coords[1]] = numToSearch
	
	if !results.AtomicCache(&zM) || !testForConsistency(&zM)  {
		// fmt.Println("Test inconsistent or cached test. Return")
		return
	}
	// fmt.Println("Test consistent and not cached yet. Continue")
	
	backires := findZeros(&zM)
	for il := 0; il < depthWrong; il++ {
		// bres, ires := trySolve(&zM)
		_, ires := trySolve(&zM)
		// fmt.Println("bres, ires:",bres, ires, "\n")

		if ires == backires && ires != 0 {
			if testForConsistency(&zM) || depthWrong > 0 {
				wg.Add(1)
				go func(depthWrong int, zM Matrix, results *Cache, wg *sync.WaitGroup) {
					defer wg.Done()
					// fmt.Println("Start recursive resolveWrong()")
					resolveWrong(depthWrong, &zM, results, wg)
				}(depthWrong-1, zM, results, wg)
				break
			} else {
				// fmt.Println("Error: !testForConsistency(&zM) || depthWrong == 0")
				break
			}
		} else if ires == 0 {
			// fmt.Println("Found with no 0s")
			if testForConsistency(&zM) {
				// fmt.Println("and it is consistent")
				// fmt.Println("foundResult:")
				// zM.printMatrix()
				results.AtomicSet(zM)
				//solvedBool = true
			}
			// zM.printMatrix()
			// fmt.Println("It is not consistent")
			break
		} else {

			backires = ires
		}
	}
}

type findCellStruct struct {
	stop                                 bool
	coords                               [2]int
	cellVars                             [2]int
	minBlockLenNum                       int
	parsedCols, parsedRows, parsedBlocks *[9]string
}

func findCell(r, c int, a *Matrix, b int, d []interface{}) {
	result := d[0].(*findCellStruct)
	if result.stop {
		return
	}
	if b == result.minBlockLenNum && a.mtrx[r][c] == 0 {
		str := reduceStrings(result.parsedRows[r], result.parsedCols[c],
			result.parsedBlocks[b])
		if len(*str) == 2 {
			result.coords = [2]int{r, c}
			result.cellVars = parseStringToIntArray(*str)
			result.stop = true
		}
	}
}
func resolveWrong(depthWrong int, m *Matrix, results *Cache, wg *sync.WaitGroup) {
	// fmt.Println("Start resolve wrong")
	// m.printMatrix()
	// fmt.Println("")

	parsedCols := parseRows(m)
	parsedRows := parseCols(m)
	parsedBlocks := parseToBlocks(m)
	minBlockLen := 9
	minBlockLenNum := 0
	for bi, _ := range parsedBlocks {
		if len(parsedBlocks[bi]) < minBlockLen && len(parsedBlocks[bi]) > 0 {
			minBlockLen = len(parsedBlocks[bi])
			minBlockLenNum = bi
		}
	}

	findCStruct := &findCellStruct{
		stop:           false,
		coords:         [2]int{},
		cellVars:       [2]int{},
		minBlockLenNum: minBlockLenNum,
		parsedCols:     parsedCols,
		parsedRows:     parsedRows,
		parsedBlocks:   parsedBlocks,
	}
	m.IterateMtrxByBlocks(findCell, findCStruct)
	// fmt.Println("FindCell coords and numsTS:", findCStruct.coords, findCStruct.cellVars)

	for i:=0;i<2;i++ {
		wg.Add(1)
		go func(depthWrong int, numToSearch int, coords [2]int,
			m Matrix, results *Cache, wg *sync.WaitGroup) {
			defer wg.Done()
			resolveCell(depthWrong, numToSearch, coords,
				&m, results, wg)
		}(depthWrong, findCStruct.cellVars[i], findCStruct.coords, *m, results, wg)
	}

	return
}

func PuzzeSolve(puzzle *Solver, results *Cache, wg *sync.WaitGroup) {
	defer wg.Done()
	var depthWrong int = findZeros(&puzzle.elim)
	// fmt.Println("Start Matrix:")
	// fmt.Println("Start zeros number:", findZeros(&puzzle.elim))
	// puzzle.elim.printMatrix()
	// fmt.Println("parseCols:", *parseRows(&puzzle.elim))
	// fmt.Println("parseRows:", *parseCols(&puzzle.elim))
	// fmt.Println("parseToBlocks:", *parseToBlocks(&puzzle.elim))
	backires := depthWrong
	for i := 0; i < depthWrong; i++ {
		_, ires := trySolve(&(puzzle.solutions[0]))
		//fmt.Println(bres, ires, "\n")
		if ires == backires && ires != 0 {
			break
		} else {
			backires = ires
		}
	}
	if !testForConsistency(&puzzle.solutions[0]) {
		fmt.Println("Puzzle inconsistent. Stop working")
		return
	}
	wg.Add(1)
	go func(depthWrong int, m *Matrix, results *Cache, wg *sync.WaitGroup) {
		defer wg.Done()
		resolveWrong(depthWrong, m, results, wg)
	}(depthWrong, &puzzle.solutions[0], results, wg)
}

func StartSolve() {
	results := NewCache()
	var wg sync.WaitGroup
	searchPuzzle:=""
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		tempStr:=scanner.Text()
		if tempStr == "" {break}
		searchPuzzle+=tempStr
	}
	if err := scanner.Err(); err != nil {
		panic("Input error")
	}

	res := NewMatrix(searchPuzzle)
	sols := []Matrix{}
	sols = append(sols, res)
	solve := Solver{done: false,
		elim:      res,
		solutions: sols}
	wg.Add(1)
	go PuzzeSolve(&solve, results, &wg)
	wg.Wait()
	solutions := results.AtomicGet()
	fmt.Println("\nFound", len(solutions), "decisions:")

	for k, _ := range solutions {
		k.printMatrix()
		consistent := testForConsistency(&k)
		fmt.Println("decision consistent?:", consistent, "\n")
	}
	fmt.Println("\n", len(solutions), "decisions:")
}

func main() {
	StartSolve()
}


type DataCache struct {
	sync.Mutex
	data map[Matrix]struct{}
}

type Cache struct {
	matrices DataCache
	cache    DataCache
}

func NewCache() *Cache {
	return &Cache{
		matrices: DataCache{data: make(map[Matrix]struct{})},
		cache:    DataCache{data: make(map[Matrix]struct{})},
	}
}

func (c *Cache) AtomicSet(cin Matrix) {
	c.matrices.Lock()
	c.matrices.data[cin] = struct{}{}
	c.matrices.Unlock()
}
func (c *Cache) AtomicCache(cin *Matrix) bool {
	c.cache.Lock()
	_, ok := c.cache.data[*cin]
	if !ok {
		c.cache.data[*cin] = struct{}{}
	}
	c.cache.Unlock()
	return !ok
}

func (c *Cache) AtomicGet() map[Matrix]struct{} {
	c.matrices.Lock()
	defer c.matrices.Unlock()
	return c.matrices.data
}