package sudoku

import "fmt"

func PrintReadableGrid(sol string) {
	for row := 0; row < GridSize; row++ {
		for col := 0; col < GridSize; col++ {
			fmt.Printf("%c ", sol[row*GridSize+col])
		}
		fmt.Println()
	}
}

func asciiToInts(numbers string) []int {
	toReturn := []int{}
	for _, c := range numbers {
		toReturn = append(toReturn, int(c)-int('0'))
	}
	return toReturn
}

func intToAscii(input int) byte {
	return byte(input) + byte('0')
}
