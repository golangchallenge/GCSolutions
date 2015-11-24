package main

import (
	"fmt"
	"sync"
)

func process() {
	fileLogger.Println("inside process")
	for {
		possibilities()
		fileLogger.Println("applying")
		count := applyPossibilities()
		ok := validateGrid()
		if !ok {
			return
		}
		//allCells = make(map[int]map[string]struct{})
		if count == 0 {
			break
		}
	}
	possibilities()
	if len(allCells) != 0 {
		fileLogger.Println("need to Guess")
		//printGrid()
		guessingGame()
	}
	fileLogger.Println("returning from process")
}

func possibilities() {
	fileLogger.Println("inside possibilities")
	allCells = make(map[int]map[string]struct{})
	var wg sync.WaitGroup
	var mutex sync.Mutex
	for i := 0; i < 9; i++ {
		for j := 0; j < 9; j++ {
			if grid[i][j] == "_" {
				wg.Add(1)
				go func(i int, j int) {
					defer wg.Done()
					a := make(map[string]struct{})
					a["1"] = struct{}{}
					a["2"] = struct{}{}
					a["3"] = struct{}{}
					a["4"] = struct{}{}
					a["5"] = struct{}{}
					a["6"] = struct{}{}
					a["7"] = struct{}{}
					a["8"] = struct{}{}
					a["9"] = struct{}{}
					b := rows[i]
					for key := range b {
						delete(a, key)
					}
					b = cols[j]
					for key := range b {
						delete(a, key)
					}

					switch i {
					case 0, 1, 2:
						switch j {
						case 0, 1, 2:
							for key := range blks[0] {
								delete(a, key)
							}
						case 3, 4, 5:
							for key := range blks[1] {
								delete(a, key)
							}
						case 6, 7, 8:
							for key := range blks[2] {
								delete(a, key)
							}
						}
					case 3, 4, 5:
						switch j {
						case 0, 1, 2:
							for key := range blks[3] {
								delete(a, key)
							}
						case 3, 4, 5:
							for key := range blks[4] {
								delete(a, key)
							}
						case 6, 7, 8:
							for key := range blks[5] {
								delete(a, key)
							}
						}
					case 6, 7, 8:
						switch j {
						case 0, 1, 2:
							for key := range blks[6] {
								delete(a, key)
							}
						case 3, 4, 5:
							for key := range blks[7] {
								delete(a, key)
							}
						case 6, 7, 8:
							for key := range blks[8] {
								delete(a, key)
							}
						}
					}
					mutex.Lock()
					allCells[9*i+j] = a
					mutex.Unlock()
				}(i, j)
			}

		} //j loop
	} //i loop
	wg.Wait()
	fileLogger.Println(allCells)
	fileLogger.Println("returning from possibilities")
}

func guessingGame() {
	fileLogger.Println("inside guessingGame")
	difficulty++
	var gridCopy [9][9]string
	gridCopy = grid
	minLen := 9
	var cell int
	guessed := make(map[string]struct{})
	lengths := make(map[int]int)
	for key, value := range allCells {
		currLen := len(value)
		lengths[key] = currLen
		if currLen < minLen {
			minLen = currLen
		}
	}
	for key, length := range lengths {
		if length == minLen {
			cell = key
			break
		}
	}
	a := allCells[cell]
	fileLogger.Println("cell::", cell)
	for k := range a {
		_, found := guessed[k]
		if found {
			continue
		}
		guessed[k] = struct{}{}
		setGuessedValue(cell, k)
		//printGrid()
		process()
		//printGrid()
		ok := validateGrid()
		fileLogger.Println("fffs", ok)
		if ok {
			ok = gridHasSpace()
			if ok {
				grid = gridCopy
				back++
				fileLogger.Println("assigning back grid 2")
				resetMaps()
			} else {
				break
			}
		} else {
			grid = gridCopy
			back++
			fileLogger.Println("assigning back grid")
			resetMaps()
		}
	}
	fileLogger.Println("returning from guessingGame")
}

func level() {
	fileLogger.Println("inside level")
	fileLogger.Println("difficulty::", difficulty)
	fileLogger.Println("back::", back)
	if difficulty < 3 {
		fmt.Println("Level:Easy")
	} else if difficulty < 8 {
		fmt.Println("Level:Medium")
	} else if difficulty < 12 {
		fmt.Println("Level:Hard")
	} else {
		fmt.Println("Level:Evil")
	}
	fileLogger.Println("returning from level")
}
