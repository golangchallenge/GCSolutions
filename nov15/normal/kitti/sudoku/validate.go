package main

import (
	"fmt"
	"strconv"
	"strings"
)

func validateInputRow(row string) ([]string, bool) {
	fileLogger.Println("inside validateInputRow")
	cells := strings.Split(row, " ")
	if len(cells) != 9 {
		fmt.Println("Please enter 9 cells per line")
		return nil, false
	}
	fileLogger.Println(cells)
	set := make(map[int]bool)
	for _, cell := range cells {
		v, err := strconv.Atoi(cell)
		if err == nil {
			if v > 0 && v < 10 {
				_, found := set[v]
				if found {
					fmt.Println("duplicate numbers")
					return nil, false
				}
				set[v] = true
			} else {
				fmt.Println("number not in range 1-9")
				return nil, false
			}
		} else {
			if !(cell == "_") {
				fmt.Println("only 1 to 9 and _ are allowed")
				return nil, false
			}
		}
	}
	fileLogger.Println("cells::", cells)
	fileLogger.Println("returning from validateInputRow")
	return cells, true
}

func validateInputCol(cells []string) bool {
	fileLogger.Println("inside validateInputCol")
	for i, cell := range cells {
		a := cols[i]
		_, found := a[cell]

		if found {
			fmt.Println("duplicate column numbers found.")
			fileLogger.Println("returning false from validateInputCol")
			return false
		}
	}
	fileLogger.Println("returning from validateInputCol")
	return true
}

func validateInputBlock(row int, cells []string) bool {
	fileLogger.Println("inside validateInputBlock")
	found := false
	switch row {
	case 0, 1, 2:
		for i, cell := range cells {
			switch i {
			case 0, 1, 2:
				a := blks[0]
				_, found = a[cell]
			case 3, 4, 5:
				a := blks[1]
				_, found = a[cell]
			case 6, 7, 8:
				a := blks[2]
				_, found = a[cell]
			}
			if found {
				fmt.Println("duplicate numbers found in block.")
				return false
			}

		}
	case 3, 4, 5:
		for i, cell := range cells {
			switch i {
			case 0, 1, 2:
				a := blks[3]
				_, found = a[cell]
			case 3, 4, 5:
				a := blks[4]
				_, found = a[cell]
			case 6, 7, 8:
				a := blks[5]
				_, found = a[cell]
			}
			if found {
				fmt.Println("duplicate numbers found in block.")
				return false
			}

		}
	case 6, 7, 8:
		for i, cell := range cells {
			switch i {
			case 0, 1, 2:
				a := blks[6]
				_, found = a[cell]
			case 3, 4, 5:
				a := blks[7]
				_, found = a[cell]
			case 6, 7, 8:
				a := blks[8]
				_, found = a[cell]
			}
			if found {
				fmt.Println("duplicate numbers found in block.")
				return false
			}

		}
	}
	fileLogger.Println("returning from validateInputBlock")
	return true
}

func validateGrid() bool {
	fileLogger.Println("inside validateGrid")
	fileLogger.Println("will validate rows")
	ok := validateRows()
	if ok {
		fileLogger.Println("will validate columns")
		ok = validateColumns()
		if ok {
			fileLogger.Println("will validate blocks")
			ok = validateBlocks()
		}
	}
	if ok {
		fileLogger.Println("validation successful")
		fileLogger.Println("returning true from validateGrid")
		return true
	}
	fileLogger.Println("returning false from validateGrid")
	return false
}

func validateRows() bool {
	//newMap := make(map[string]struct{})
	for i := 0; i < 9; i++ {
		newMap := make(map[string]struct{})
		for j := 0; j < 9; j++ {
			a := grid[i][j]
			_, found := newMap[a]
			if found {
				return false
			}
			if !(a == "_") {
				newMap[a] = struct{}{}
			}
		}
	}
	return true
}

func validateColumns() bool {
	//newMap := make(map[string]struct{})
	for i := 0; i < 9; i++ {
		newMap := make(map[string]struct{})
		for j := 0; j < 9; j++ {
			a := grid[j][i]
			_, found := newMap[a]
			if found {
				return false
			}
			if !(a == "_") {
				newMap[a] = struct{}{}
			}
		}
	}
	return true
}

func validateBlocks() bool {
	for i := 0; i < 9; {
		newMap := make(map[string]struct{})
		for j := 0; j < 9; j++ {
			if j == 3 || j == 6 {
				newMap = make(map[string]struct{}) //resetting map for next block
			}
			k := i + 3
			for i < k {
				a := grid[i][j]
				_, found := newMap[a]
				if found {
					return false
				}
				if !(a == "_") {
					newMap[a] = struct{}{}
				}
				i++
			}
			i -= 3
		}
		i += 3
	}
	return true
}

func gridHasSpace() bool {
	for i := 0; i < 9; i++ {
		for j := 0; j < 9; j++ {
			a := grid[i][j]
			if a == "_" {
				return true
			}
		}
	}
	return false
}
