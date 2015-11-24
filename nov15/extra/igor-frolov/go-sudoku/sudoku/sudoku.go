package sudoku

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"unicode"
)

type SudokuField [81]byte

func (sudoku SudokuField) Print() {
	for i := 0; i < 81; i++ {
		if sudoku[i] != 0 {
			fmt.Print(sudoku[i], " ")
		} else {
			fmt.Print("_ ")
		}
		if i%9 == 8 {
			fmt.Println()
		}
	}
}

func (sudoku *SudokuField) Input() {
	for i := 0; i < 9; i++ {
		reader := bufio.NewReader(os.Stdin)
		line, _ := reader.ReadString('\n')
		//fmt.Println(line, "!", sudoku)
		for j, element := range strings.Fields(line) {
			if unicode.IsDigit(rune(element[0])) {
				sudoku[9*i+j] = byte(element[0] - '0')
			} else if element == "_" {
				sudoku[9*i+j] = byte(0)
			} else {
				fmt.Println("Wrong element")
				fmt.Println("Digit or \"_\" needed")
				fmt.Println("Try again")
				i = 0
			}
		}
	}
}

func (in *SudokuField) Clone() (out SudokuField) {
	for i := 0; i < 81; i++ {
		out[i] = in[i]
	}
	return
}

func (sudoku *SudokuField) Copy(source SudokuField) {
	for i := 0; i < 81; i++ {
		sudoku[i] = source[i]
	}
}

type set map[byte]bool

func (num1 set) Intersection(num2 set) set {
	res := make(set)
	for n1, _ := range num1 {
		if _, thereIs := num2[n1]; thereIs {
			res[n1] = true
		}
	}
	return res
}

func (num set) Inverse() set {
	res := make(set)
	for i := 1; i < 10; i++ {
		if _, thereIs := num[byte(i)]; !thereIs {
			res[byte(i)] = true
		}
	}
	return res
}

func getSquare(index int) byte {
	//get square 3x3 by cell index
	return byte((index/27)*3 + (index/3)%3)
}

func (sudoku *SudokuField) possibleNumbersInRow(index int) set {
	inRow := set{}
	start := (index / 9) * 9
	for i := start; i < start+9; i++ {
		if sudoku[i] != 0 {
			inRow[sudoku[i]] = true
		}
	}
	return inRow.Inverse()
}

func (sudoku *SudokuField) possibleNumbersInColumn(index int) set {
	inCol := set{}
	start := index % 9
	for i := start; i <= start+72; i += 9 {
		if sudoku[i] != 0 {
			inCol[sudoku[i]] = true
		}
	}
	return inCol.Inverse()
}

func (sudoku *SudokuField) possibleNumbersInSquare(index int) set {
	inSq := set{}
	square := getSquare(index)
	start := (square/3)*27 + (square%3)*3
	for i := start; i < start+3; i++ {
		if sudoku[i] != 0 {
			inSq[sudoku[i]] = true
		}
		if sudoku[i+9] != 0 {
			inSq[sudoku[i+9]] = true
		}
		if sudoku[i+18] != 0 {
			inSq[sudoku[i+18]] = true
		}
	}
	return inSq.Inverse()
}

func (sudoku *SudokuField) Evidents() bool {
	//Look all empty cells
	//If cell has evident value set this value to cell
	//Return true if there are "evident cells"
	//Else return false

	possible := false
	for i := 0; i < 81; i++ {
		if sudoku[i] == 0 {
			row := sudoku.possibleNumbersInRow(i)
			col := sudoku.possibleNumbersInColumn(i)
			possibleNumbers := row.Intersection(col)
			if len(possibleNumbers) == 1 {
				possible = true
				for key, _ := range possibleNumbers {
					sudoku[i] = key
				}
			} else {
				sqr := sudoku.possibleNumbersInSquare(i)
				possibleNumbers = possibleNumbers.Intersection(sqr)
				if len(possibleNumbers) == 1 {
					possible = true
					for key, _ := range possibleNumbers {
						sudoku[i] = key
					}
				}
			}
		}
	}
	return possible
}

func (sudoku_prev *SudokuField) checkAll(possibleNumbers set, index int,
	complexity int) (bool, int) {
	var sudoku_next SudokuField
	res := false
	complexity += len(possibleNumbers) - 1
	saveCompl := complexity
	for key, _ := range possibleNumbers {
		sudoku_next = sudoku_prev.Clone()
		sudoku_next[index] = key
		res, complexity = sudoku_next.Solve(saveCompl)
		if res {
			sudoku_prev.Copy(sudoku_next)
			return true, complexity
		}
	}
	return false, complexity
}

func (sudoku *SudokuField) Solved() bool {
	//Sudoku solved if there is no empty cells

	for i := 0; i < 81; i++ {
		if sudoku[i] == 0 {
			return false
		}
	}
	return true
}

func (sudoku *SudokuField) Solve(complexity int) (bool, int) {
	//Solve sudoku
	//Return true if sudoku has a solution
	//Else return false

	thereIsEvidentCeil := true
	for thereIsEvidentCeil {
		thereIsEvidentCeil = sudoku.Evidents()
	}
	if sudoku.Solved() {
		return true, complexity
	} else {

		//find empty cell, which has minimal count of possible numbers
		minPossible := set{1: true, 2: true, 3: true,
			4: true, 5: true, 6: true,
			7: true, 8: true, 9: true}
		indexMinPossible := 0
		for i := 0; i < 81; i++ {
			if sudoku[i] == 0 {
				row := sudoku.possibleNumbersInRow(i)
				col := sudoku.possibleNumbersInColumn(i)
				sqr := sudoku.possibleNumbersInSquare(i)
				possibleNumbers := row.Intersection(col.Intersection(sqr))
				if len(possibleNumbers) < len(minPossible) {
					minPossible = possibleNumbers
					indexMinPossible = i
				}
			}
		}
		//empty cell with minimal count of possible numbers finded

		//If len(minPossible) > 0 (2 and more possible numbers),
		//sudoku can have a solution.
		//We will check all possible ways.
		//If len(minPossible) equal 0, then sudoku has no solution
		var solved bool
		if len(minPossible) != 0 {
			solved, complexity = sudoku.checkAll(minPossible,
				indexMinPossible, complexity)
			return solved, complexity
		}
	}
	return false, complexity
}

func (sudoku SudokuField) lookCells(n int) set {
	var sudokuCopy SudokuField
	cells := set{}
	for i := 0; i < 81; i++ {
		sudokuCopy = sudoku.Clone()
		sudokuCopy[i] = 0
		row := sudokuCopy.possibleNumbersInRow(i)
		col := sudokuCopy.possibleNumbersInColumn(i)
		sqr := sudokuCopy.possibleNumbersInSquare(i)
		possibleNumbers := row.Intersection(col.Intersection(sqr))
		if sudoku[i] != 0 && len(possibleNumbers) == n {
			cells[byte(i)] = true
		}
	}
	return cells
}

func (sudoku SudokuField) lookFilledCells() set {
	filledCells := set{}
	for i := 0; i < 81; i++ {
		if sudoku[i] != 0 {
			filledCells[byte(i)] = true
		}
	}
	return filledCells
}

func SudokuNew(complexity int) SudokuField {
	var sudoku SudokuField
	sudoku.Solve(0)
	completed := false
	for !completed {
		filledCells := sudoku.lookFilledCells()
		for key, _ := range filledCells {
			sudoku[key] = 0
			break
		}
		sudokuCopy := sudoku.Clone()
		_, compl := sudokuCopy.Solve(0)
		completed = (compl >= complexity)
	}
	evidentCells := sudoku.lookCells(1)
	for len(evidentCells) > 0 {
		for key, _ := range evidentCells {
			sudoku[key] = 0
			break
		}
		evidentCells = sudoku.lookCells(1)
	}
	return sudoku
}
