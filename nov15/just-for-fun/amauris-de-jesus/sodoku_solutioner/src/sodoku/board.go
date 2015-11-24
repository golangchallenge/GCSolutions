package sodoku;

import (
	//"fmt"
	"strings"
	"strconv"
	"math"
)

type Board struct {
	cursor int
	Entries [][]int
	preDefinedLayout string
	dimensions int
	rowFamilyCache map[int][]int
	columnFamilyCache map[int][]int
	quadrantFamilyCache map[int][]int
}

func GetCleanBoard(dimensions int) *Board {

	newSodokuBoard := createBoard("", dimensions)
	//instantiate board
	newSodokuBoard.initBoard()
	
	//cache that makes it easy for 
	//accessing rows, columns, quadrants
	//newSodokuBoard.SetFamilyCache()

	return newSodokuBoard
}

func GetPreDefinedBoard(layout string, dimensions int) *Board {

	newPreDefinedSodokuBoard := createBoard(layout, dimensions)

	//instantiate board
	newPreDefinedSodokuBoard.initBoard()
	//fill board with the right values
	newPreDefinedSodokuBoard.fillBoard()
	
	//cache that makes it easy for 
	//accessing rows, columns, quadrants
	//newPreDefinedSodokuBoard.SetFamilyCache()

    return newPreDefinedSodokuBoard
}

func createBoard(layout string, dimensions int) *Board {

	newBoard := &Board{1, [][]int{}, layout, dimensions, make(map[int][]int, dimensions), make(map[int][]int, dimensions), make(map[int][]int, dimensions)}

	return newBoard
}

func (inst *Board) initBoard() {

	entries := make([][]int, inst.dimensions)

    for i, _ := range(entries) {
            entries[i] = make([]int, inst.dimensions)
    }

    inst.Entries = entries
}

//returns the current cursor position
//as well as the entry value
func (inst *Board) GetNextEntry() (i, j, v int) {

	if inst.cursor>(inst.dimensions*inst.dimensions) {

		return -1, -1, -1
	}

	i = int((inst.cursor-1)/inst.dimensions)
	j = (inst.cursor-1)%inst.dimensions
	v = inst.Entries[i][j]

	inst.cursor += 1

	return i, j, v
}

//returns the current cursor position
//as well as the entry value
func (inst *Board) HasNextEntry() bool {

	if inst.cursor>(inst.dimensions*inst.dimensions) {

		return false
	}

	return true
}

func (inst *Board) ResetCursor() {

	inst.cursor = 1
}

func (inst *Board) SetCursor(i, j int) {

	j += 1

	inst.cursor = (i*inst.dimensions) + j
}

func (inst *Board) fillBoard() {

	rows := strings.Split(inst.preDefinedLayout, "\n")
	entries := make([][]int, inst.dimensions)

	i := 0
    for _, row := range(rows) {

    	row := strings.TrimSpace(row)

    	if(len(row)<=0) {
    		continue;
    	}

		columns := strings.Split(row, " ")

		if i>=inst.dimensions {
			break
		}

		entries[i] = make([]int, inst.dimensions)
		j := 0
		for _, entry := range(columns) {

			entry = strings.TrimSpace(entry)

			if(len(entry)<=0) {
    			continue;
    		}

			if j>=inst.dimensions {
				break
			}

			if entry=="_" {
				entries[i][j] = 0
			} else {
				entryInt, err := strconv.Atoi(entry)

				if err!=nil {
					entries[i][j] = 0
				} else {
					entries[i][j] = entryInt
				}
			}

			j += 1
		}

		i += 1
    }

	inst.Entries = entries
}

//traverse through all family(rows, columns, and quadrants)
//and insert to private family type cache
func (inst *Board) SetFamilyCache() {

	//set columns cache
	//theres n(count of dimensinos) of columns
	for i, row := range(inst.Entries) {

		column := make([]int, inst.dimensions)
		for j, _ := range(row) {
			column[j] = inst.Entries[j][i]

			//set quadrant family
			quadrantX, quadrantY := inst.getQuadrantIndex(i, j)
			quadrantIndexHash := inst.getQuadrantIndexHash(quadrantX, quadrantY)
			
			if _, ok := inst.quadrantFamilyCache[quadrantIndexHash]; !ok {
				inst.quadrantFamilyCache[quadrantIndexHash] = inst.GetQuadrant(i, j)
			} 
		}

		inst.columnFamilyCache[i] = column
		inst.rowFamilyCache[i] = row
	}
	//for every 
}

//Gets all empty indices from the board
func (inst *Board) SetEntries(newEntries [][]int) {

	inst.Entries = newEntries
}

func (inst *Board) SetEntry(i, j, value int) {

	//make sure to update cache values also

	/*
	//update row cache
	inst.rowFamilyCache[i][j] = value

	//update column cache
	inst.columnFamilyCache[j][i] = value

	//update quadrant cache
	x, y := inst.getQuadrantIndex(i, j)
	hash := inst.getQuadrantIndexHash(x, y)
	quadrant := inst.quadrantFamilyCache[hash]

	quadrant[(i%3)*3 + j%3] = value
	//fmt.Println(inst.columnFamilyCache)
	*/
	
	inst.Entries[i][j] = value	
}

func (inst *Board) getQuadrantIndexHash(i, j int) int {

	return i + j*10

}

func (inst *Board) getQuadrantIndex(i, j int) (int, int) {

	quadrantX := int(math.Floor(float64(i/3)))*3
	quadrantY := int(math.Floor(float64(j/3)))*3

	return quadrantX, quadrantY
}

func (inst *Board) GetRow(i int) []int {
	return inst.Entries[i]
}


func (inst *Board) GetColumn(j int) []int {

	row := make([]int, inst.dimensions)

	for i, r:= range(inst.Entries) {
		row[i] = r[j]
	}

	return row

}

func (inst *Board) GetQuadrant(i, j int) []int {

	quadrant := make([]int, inst.dimensions)

	quadrantX, quadrantY := inst.getQuadrantIndex(i, j)
	
	quadrantTemp := inst.Entries[quadrantX:quadrantX+3]
	for i, v := range(quadrantTemp) {
		tempSubRow := v[quadrantY:quadrantY+3]
		for j, v2 := range(tempSubRow) {
			quadrant[(i*3)+j] = v2
		}
	}

	return quadrant
}

/*
func (inst *Board) GetRowCache(i int) []int {
	
	return inst.rowFamilyCache[i]
}


func (inst *Board) GetColumnCache(j int) []int {

	return inst.columnFamilyCache[j]
}

func (inst *Board) GetQuadrantCache(i, j int) []int {

	quadrantX, quadrantY := inst.getQuadrantIndex(i, j)
	quadrantIndexHash := inst.getQuadrantIndexHash(quadrantX, quadrantY)

	return inst.quadrantFamilyCache[quadrantIndexHash]
}
*/

//map i, j to hash so you can retrieve
//from appropriate rowFamilyCache index
func (inst *Board) GetFamilies(i, j int) [][]int {

	/*rowFamily := inst.GetRowCache(i)
	columnFamily := inst.GetColumnCache(j)
	quadrantFamily := inst.GetQuadrantCache(i, j)*/

	
	rowFamily := inst.GetRow(i)
	columnFamily := inst.GetColumn(j)
	quadrantFamily := inst.GetQuadrant(i, j) 

	return [][]int{rowFamily, columnFamily, quadrantFamily}
}

//make sure for every entry, its corresponding
//family is unique
func (inst *Board) IsBoardComplete() bool {

	for i, row := range(inst.Entries) {
		for j, _ := range(row) {
			families := inst.GetFamilies(i, j)
			for _, family := range(families) {
				numsToCount := make(map[int]int, inst.dimensions)

				for _, value := range(family) {
					if numsToCount[value]==0 {
						numsToCount[value] = 1
					} else {
						//fmt.Println(numsToCount);
						return false
					}
				}
			}
		}
	}

	return true
}

//Gets indices that need to be filled(with value 0)
//along the families from related to the index 
//i, j
func (inst *Board) GetFamilyEmptyIndices(i, j int) [][]int {

	emptyIndices := [][]int{}

	for inst.HasNextEntry() {
		i, j, v := inst.GetNextEntry()

		if v==0 {
			emptyIndices = append(emptyIndices, []int{i, j})
		}
	}

	inst.ResetCursor()

	return emptyIndices
}

//Gets all empty indices from the board
func (inst *Board) GetEmptyIndices() [][]int {

	emptyIndices := [][]int{}

	for inst.HasNextEntry() {
		i, j, v := inst.GetNextEntry()

		if v==0 {
			emptyIndices = append(emptyIndices, []int{i, j})
		}
	}

	inst.ResetCursor()

	return emptyIndices
}

func (inst *Board) GetStringFormat() string {

	boardString := ""

	for _, row := range(inst.Entries) {
		for _, entry := range(row) {
			if entry==0 {
				boardString += "_"
			} else {
				boardString += strconv.Itoa(entry)
			}
			boardString += " "
		}
		boardString += "\n" 
	}

	return boardString
}