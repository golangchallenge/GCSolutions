package sodoku;

import (
	"testing"
)

//predefined variable used through script
var (
	DIMENSIONS int = 9

	SIMPLE_TABLE_STRING string =  `1 _ 3 _ _ 6 _ 8 _
		_ 5 _ _ 8 _ 1 2 _
		7 _ 9 1 _ 3 _ 5 6
		_ 3 _ _ 6 7 _ 9 _
		5 _ 7 8 _ _ _ 3 _
		8 _ 1 _ 3 _ 5 _ 7
		_ 4 _ _ 7 8 _ 1 _
		6 _ 8 _ _ 2 _ 4 _
		_ 1 2 _ 4 5 _ 7 8`

	SIMPLE_TABLE_ARRAY [][]int =  [][]int{
		[]int{1, 0, 0, 3, 0, 0, 6, 0, 8}, 
		[]int{0, 5, 0, 0, 8, 0, 1, 2, 0},
		[]int{7, 0, 9, 1, 0, 3, 0, 5, 6},
		[]int{0, 3, 0, 0, 6, 7, 0, 9, 0},
		[]int{5, 0, 7, 8, 0, 0, 0, 3, 0},
		[]int{8, 0, 1, 0, 3, 0, 5, 0, 7},
		[]int{0, 4, 0, 0, 7, 8, 0, 1, 0},
		[]int{6, 0, 8, 0, 0, 2, 0, 4, 0},
		[]int{0, 1, 2, 0, 4, 5, 0, 7, 8},
	}

	HARD_TABLE_STRING string = `5 3 _ _ 7 _ _ _ _
		6 _ _ 1 9 5 _ _ _
		_ 9 8 _ _ _ _ 6 _
		8 _ _ _ 6 _ _ _ 3
		4 _ _ 8 _ 3 _ _ 1
		7 _ _ _ 2 _ _ _ 6
		_ 6 _ _ _ _ 2 8 _
		_ _ _ 4 1 9 _ _ 5
		_ _ _ _ 8 _ _ 7 9`

	HARD_TABLE_ARRAY [][]int = [][]int{
		[]int{5, 3, 0, 0, 7, 0, 0, 0, 0},
		[]int{6, 0, 0, 1, 9, 5, 0, 0, 0},
		[]int{0, 9, 8, 0, 0, 0, 0, 6, 0},
		[]int{8, 0, 0, 0, 6, 0, 0, 0, 3},
		[]int{4, 0, 0, 8, 0, 3, 0, 0, 1},
		[]int{7, 0, 0, 0, 2, 0, 0, 0, 6},
		[]int{0, 6, 0, 0, 0, 0, 2, 8, 0},
		[]int{0, 0, 0, 4, 1, 9, 0, 0, 5},
		[]int{0, 0, 0, 0, 8, 0, 0, 7, 9},
	}
)

//Helper functions

//return true if both array contani same values
//else return false
func compareArray(array1, array2 []int) bool {
	
	if len(array1)!=len(array2) {
		return false
	}

	for i, _ := range(array1) {
		if array1[i]!=array2[i] {
			return false
		}
	}

	return true
}

func buildSodokuBoard() *Board {
	
	sodoku := GetPreDefinedBoard(HARD_TABLE_STRING, DIMENSIONS)

	return sodoku
}

//End of helper functions

func TestGetCleanBoard(t *testing.T) {

	sodoku := GetCleanBoard(DIMENSIONS)
	testArray := []int{0, 0, 0, 0, 0, 0, 0, 0, 0}

	if len(sodoku.Entries) <= 0 {
		t.Error("Expected ", testArray, " Got ", sodoku.Entries)
	}

	for _, row := range(sodoku.Entries) {
		if !compareArray(testArray, row) {
			t.Error("Expected ", testArray, " Got ", row)
		}
	}

}

func TestGetPreDefinedBoard(t *testing.T) {

	sodoku := GetPreDefinedBoard(HARD_TABLE_STRING, DIMENSIONS)

	if len(sodoku.Entries) <= 0 {
		t.Error("Expected ", HARD_TABLE_ARRAY, " Got ", sodoku.Entries)
	}

	for i, row := range(sodoku.Entries) {
		if !compareArray(HARD_TABLE_ARRAY[i], row) {
			t.Error("Expected ", HARD_TABLE_ARRAY[i], " Got ", row)
		}
	}
}

func TestInitBoard(t *testing.T) {

	sodoku := buildSodokuBoard()

	//if rows is not equal to dimensions
	if len(sodoku.Entries)!=DIMENSIONS {
		t.Error("Length should be ", DIMENSIONS)
	}

	for _, row := range(sodoku.Entries) {
		//if rows is not equal to dimensions
		if len(row)!=DIMENSIONS {
			t.Error("Length of row should be ", DIMENSIONS)
		}
	}
}

func TestGetNextEntry(t *testing.T) {

	sodoku := buildSodokuBoard()

	for _, row := range(HARD_TABLE_ARRAY) {

		for _, entry := range(row) {

			_, _, v := sodoku.GetNextEntry()

			if v != entry {
				t.Error("Expected entry ", entry, " got ", v)
			}
		}
	}

}

//traverse through all family(rows, columns, and quadrants)
//and insert to private family type cache
func TestSetFamilyCache(t *testing.T) {

	//set columns cache
	//for every 
}

func TestGetRow(t *testing.T) {

	sodoku := buildSodokuBoard()

	for i, _ := range(sodoku.Entries) {
		row := sodoku.GetRow(i)
		if !compareArray(HARD_TABLE_ARRAY[i], row) {
			t.Error("Expected ", HARD_TABLE_ARRAY[i], " Got ", row)
		}
	}
}

func TestSetEntry(t *testing.T) {

}

func TestGetColumn(t *testing.T) {

	sodoku := buildSodokuBoard()
	column := sodoku.GetColumn(0)

	for i, v := range(column) {
		if v!=HARD_TABLE_ARRAY[i][0] {
			t.Error("Expected ", HARD_TABLE_ARRAY[i][0], " Got ", v)
		}
	}
}

func TestGetQuadrant(t *testing.T) {

	sodoku := buildSodokuBoard()
	quadrant := sodoku.GetQuadrant(0, 0)

	i := 0
	for i<=2 {
		row := quadrant[(i*3):(i*3)+2]
		if !compareArray(HARD_TABLE_ARRAY[i][0:2], row) {
			t.Error("Expected ", HARD_TABLE_ARRAY[i][0:2], " Got ", row)
		}

		i += 1
	}
}

//map i, j to hash so you can retrieve
//from appropriate columnFamilyCache index
func TestgetColumnFamily(t *testing.T) {

}

//map i, j to hash so you can retrieve
//from appropriate quadrantFamilyCache index
func TestgetQuadrantFamily(t *testing.T) {

}


//make sure for every entry, its corresponding
//family is unique
func TestIsBoardComplete(t *testing.T) {


}
