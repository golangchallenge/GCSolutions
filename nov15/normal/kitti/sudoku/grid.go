package main

import (
	"bufio"
	"fmt"
	"os"
)

func applyPossibilities() int {
	fileLogger.Println("inside applyPossibilities")
	count := 0
	for key, value := range allCells {
		if len(value) == 1 {
			var num string
			for vkey := range value {
				num = vkey
			}
			j := key % 9
			i := key / 9
			grid[i][j] = num
			a := rows[i]
			if a == nil {
				a = make(map[string]struct{})
			}
			a[num] = struct{}{}
			rows[i] = a

			a = cols[j]
			if a == nil {
				a = make(map[string]struct{})
			}
			a[num] = struct{}{}
			cols[j] = a

			switch i {
			case 0, 1, 2:
				switch j {
				case 0, 1, 2:
					a := blks[0]
					if a == nil {
						a = make(map[string]struct{})
					}
					a[num] = struct{}{}
					blks[0] = a
				case 3, 4, 5:
					a := blks[1]
					if a == nil {
						a = make(map[string]struct{})
					}
					a[num] = struct{}{}
					blks[1] = a
				case 6, 7, 8:
					a := blks[2]
					if a == nil {
						a = make(map[string]struct{})
					}
					a[num] = struct{}{}
					blks[2] = a
				}
			case 3, 4, 5:
				switch j {
				case 0, 1, 2:
					a := blks[3]
					if a == nil {
						a = make(map[string]struct{})
					}
					a[num] = struct{}{}
					blks[3] = a
				case 3, 4, 5:
					a := blks[4]
					if a == nil {
						a = make(map[string]struct{})
					}
					a[num] = struct{}{}
					blks[4] = a
				case 6, 7, 8:
					a := blks[5]
					if a == nil {
						a = make(map[string]struct{})
					}
					a[num] = struct{}{}
					blks[5] = a
				}
			case 6, 7, 8:
				switch j {
				case 0, 1, 2:
					a := blks[6]
					if a == nil {
						a = make(map[string]struct{})
					}
					a[num] = struct{}{}
					blks[6] = a
				case 3, 4, 5:
					a := blks[7]
					if a == nil {
						a = make(map[string]struct{})
					}
					a[num] = struct{}{}
					blks[7] = a
				case 6, 7, 8:
					a := blks[8]
					if a == nil {
						a = make(map[string]struct{})
					}
					a[num] = struct{}{}
					blks[8] = a
				}

			}

			count++
		}
	}
	fileLogger.Println("count::", count)
	fileLogger.Println("returning from applyPossibilities")
	return count
}

func setGuessedValue(cell int, value string) {
	fileLogger.Println("ïnside setGuessValue", cell, value)
	j := cell % 9
	i := cell / 9
	grid[i][j] = value
	addToRow(i, value)
	addToCol(j, value)
	addToBlk(i, j, value)
	possibilities()
	fileLogger.Println("returning from setGuessedValue")
}

func getGrid() {
	fileLogger.Println("inside getGrid")
	var input string
	scanner := bufio.NewScanner(os.Stdin)
	i := 0
	for {
		fmt.Println("Enter row no:", i)
		scanner.Scan()
		input = scanner.Text()
		//fmt.Println(input)
		added := false
		cells, ok := validateInputRow(input)
		if ok {
			ok = validateInputCol(cells)
			if ok {
				ok = validateInputBlock(i, cells)
				if ok {
					for j := 0; j < 9; j++ {
						grid[i][j] = cells[j]
						addToRow(i, cells[j])
						addToCol(j, cells[j])
						addToBlk(i, j, cells[j])
						added = true
					}
				}
			}

		}
		if !added {
			fmt.Println("invalid input.")
			continue
		}
		if i < 8 {
			i++
		} else {
			break
		}
	}
	fileLogger.Println("returning from getGrid")
}

func printGrid() {
	for i := 0; i < 9; i++ {
		for j := 0; j < 9; j++ {
			fmt.Print(grid[i][j])
			if j != 8 {
				fmt.Print(" ")
			}
		}
		fmt.Print("\n")
	}
}
