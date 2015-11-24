package main

import (
	"fmt"
	"os"
	"errors"
	"io/ioutil"
	"time"
)

//var test = "8_2_______6___31_89_16____25__368_14___2_9___39_174__62____62_36_35___8_______7_5"
//var test = "1_3__6_8__5__8_12_7_91_3_56_3__67_9_5_78___3_8_1_3_5_7_4__78_1_6_8__2_4__12_45_78"
//var test = "___9421_7__93______32___469__3__95_1___473___2_71__3__691___85______59__3_8691___"
var test = "090730004000000009234080000000063700050000040007510000000050298100000000600024030"

func ratePuzzle(puz [81]int8) string {
	emptyCount := 0
	for i:=0;i<81;i++ {
		if(puz[i] == 0){
			emptyCount++
		}
	}
	if(emptyCount > 50) {
		return "Hard"
	}
	if(emptyCount > 45) {
		return "Medium"
	}
	return "Easy"
}

func main() {

	args := os.Args[1:]

	if len(args) == 0 {
		fmt.Printf("Usage: sudoku command\n\nCommands:\n  create\tCreate a puzzle to solve\n  solve  \tSolve a puzzle\n\n")
		os.Exit(0)
	}

	if args[0] == "create" {
		printPuzzle(test)
		os.Exit(0)
	}
	if args[0] == "solve" {
		var solved [81]int8
		puz,err := readPuzzle()
		if err!=nil {
			fmt.Printf("Error: %s", err)
			os.Exit(1)
		}

		rating := ratePuzzle(puz)

		fmt.Printf("Original (%s):\n", rating)
		printPuzzleInt(puz)

		t0 := time.Now()
		solved,err = solvePuzzle(puz, 0)
		t1 := time.Now()
		fmt.Printf("\nSolution (%v):\n", t1.Sub(t0))

		if err != nil {
			fmt.Printf("Unsolvable %s\n", err)
			printPuzzleInt(solved)
		} else {
			printPuzzleInt(solved)
		}
		os.Exit(0)
	}
}

func printPuzzle(puzzle string) {
	for i, d := range puzzle {
		var sep = ' '
		if i % 9 == 8 {
			sep = '\n'
		}
		if(d == '0') {
			fmt.Printf("_%c", sep)
		} else {
			fmt.Printf("%c%c", d, sep)
		}
	}
}
func printPuzzleInt(puzzle [81]int8) {
	for i, d := range puzzle {
		var sep = ' '
		if i % 9 == 8 {
			sep = '\n'
		}
		if (d == 0) {
			fmt.Printf("_%c", sep)
		} else {
			fmt.Printf("%d%c", d, sep)
		}
	}
}

func readPuzzle() ([81]int8,error) {
	var pos = 0
	var puzzle [81]int8

	buf, _ := ioutil.ReadAll(os.Stdin)

	for _, d := range buf {
		if (pos == 81) {
			return puzzle,nil
		}
		if d == 95 {
			puzzle[pos] = 0
			pos++
		}
		if d >= 49 && d <= 57 {
			puzzle[pos] = int8(d - 48)
			pos++
		}
	}
	if(pos < 81) {
		return puzzle,errors.New("There is not enough squares entered")
	}
	return puzzle,nil
}

func solvePuzzle(puz [81]int8, depth int) ([81]int8, error) {
	if (depth == 100000) {
		return puz, errors.New("Max Depth")
	}
	var available int16
	var i int;
	for i = 0; i < 81; i++ {
		if (puz[i] == 0) {
			available = availableSet(
				createSet(getBlock(puz, i)),
				createSet(getRow(puz, i)),
				createSet(getCol(puz, i)),
			)
			if available == 0 {
				return puz, errors.New("No available")
			}
			if (countSetBits(available) == 1) {
				puz[i] = setToNumber(available);
				return solvePuzzle(puz, depth + 1)
			} else {
					for j := 1; j <= 9; j++ {
						if (numberInSet(j, available)) {
							puz[i] = int8(j);
							solved, err := solvePuzzle(puz, depth + 1)
							if err == nil {
								return solved, nil;
							}
						}
					}
					return puz, errors.New("All Guessed Failed")
				}
			break;
		}
	}
	return puz, nil;
}

// 0 1 2
// 3 4 5
// 6 7 8
func getSegmentNumber(i int) int {
	segRow := int(i / 27)
	segCol := int(i % 9 / 3)
	return segRow * 3 + segCol
}

func getBlock(puz [81]int8, pos int) [9]int8 {
	var block [9]int8
	var l int = 0
	i := (int(pos / 27) * 27) + (int(pos % 9 / 3) * 3)
	for j := 0; j < 3; j++ {
		for k := 0; k < 3; k++ {
			block[l] = puz[i]
			i++
			l++
		}
		i += 6 // move to next row +9 and back 3 space -3 = +6
	}
	return block
}

func getRow(puz [81]int8, pos int) [9]int8 {
	var row [9]int8
	var j int = 0
	for i := (int(pos / 9) * 9); i < (int(pos / 9) * 9) + 9; i++ {
		row[j] = puz[i]
		j++
	}
	return row
}
func getCol(puz [81]int8, pos int) [9]int8 {
	var row [9]int8
	var j int = 0
	for i := (pos % 9); i < 81; i += 9 {
		row[j] = puz[i]
		j++
	}
	return row
}

// 16 ... 10  9  8  7  6  5  4  3  2  1
//  0 ...  0  0  0  0  0  0  0  0  0  0
//  0 ...  0  0  0  0  0  1  1  0  1  0 = 2,4,5
//  0 ...  0  1  1  1  1  1  1  1  1  1 = 1,2,3,4,5,6,7,8,9
func createSet(input [9]int8) int16 {
	var used int16 = 0
	for i := 0; i < 9; i++ {
		used += (1 << uint(input[i] - 1))
	}
	return used;
}

func countSetBits(v int16) int {
	var count int = 0;
	for i := 0; i < 9; i++ {
		count += int(v & 1)
		v = v >> 1
	}
	return count
}

func availableSet(set1 int16, set2 int16, set3 int16) int16 {
	return 511 - (set1 | set2 | set3)
}

func setToNumber(set int16) int8 {
	for i := 1; i <= 9; i++ {
		if set == 1 << uint(i - 1) {
			return int8(i)
		}
	}
	return 0
}

func numberInSet(i int, set int16) bool {
	return ((set >> uint(i - 1)) & 1) == 1;
}



