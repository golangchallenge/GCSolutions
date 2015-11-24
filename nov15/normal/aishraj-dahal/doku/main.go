package main

import (
	"bufio"
	"fmt"
	"github.com/aishraj/doku/parser"
	"log"
	"os"
	//"strings"
)

type cell struct {
	x, y, v        int
	possibleValues [9]bool //1 thorugh 9
}

func (c *cell) eliminate(val int, grid *[9][9]cell) {
	if c.v != 0 {
		return
	}
	c.possibleValues[val-1] = false
	cnt := 0
	retValue := -1
	for _, v := range c.possibleValues {
		if v {
			cnt++
		}
	}

	if cnt == 1 {
		for i, v := range c.possibleValues {
			if v {
				retValue = i + 1
				break
			}
		}
		setValue(c, retValue, grid)
	}

}

func setValue(r *cell, v int, grid *[9][9]cell) {

	for i := 0; i < 9; i++ {
		r.possibleValues[i] = false
	}

	r.possibleValues[v-1] = true

	grid[r.y][r.x].v = v
	r.v = v

	for x := 0; x < 9; x++ {
		if x != r.x && grid[r.y][r.x].v != 0 {
			grid[r.y][x].eliminate(v, grid)
		}
	}

	//eliminate all row values based on column values ;eliminate from 2nd unit
	for y := 0; y < 9; y++ {
		if y != r.y && grid[r.y][r.x].v != 0 {
			grid[y][r.x].eliminate(r.v, grid)
		}
	}

	//now eleiminate from all other peers
	xStart, yStart := r.x/3*3, r.y/3*3
	for y := yStart; y < yStart+3; y++ {
		for x := xStart; x < xStart+3; x++ {
			if x != r.x && y != r.y && grid[r.y][r.x].v != 0 {
				grid[y][x].eliminate(r.v, grid)
			}
		}
	}

}

func startGame(game [9][9]int) [9][9]cell {
	grid := initializeGrid()
	for i := 0; i < 9; i++ {
		for j := 0; j < 9; j++ {
			currentCell := &grid[i][j]
			if game[i][j] != 0 {
				setValue(currentCell, game[i][j], &grid)
			}
		}
	}

	return grid
}

func initializeGrid() [9][9]cell {
	var grid [9][9]cell
	for i := 0; i < 9; i++ {
		for j := 0; j < 9; j++ {
			var possible [9]bool
			for k := 0; k < 9; k++ {
				possible[k] = true
			}
			grid[i][j] = cell{possibleValues: possible, x: j, y: i}
		}
	}
	return grid
}

func solve(game [9][9]int) (bool, [9][9]cell) {
	grid := startGame(game)
	for i := 0; i < 9; i++ {
		for j := 0; j < 9; j++ {
			cnt := 0
			x := grid[i][j].possibleValues
			for _, flag := range x {
				if flag {
					cnt++
				}
			}
			if cnt != 1 {
				return false, grid
			}
		}
	}

	return true, grid
}

func main() {
	// var puzzle = [9][9]int{{0, 3, 0, 9, 0, 0, 0, 8, 0}, {0, 0, 6, 2, 0, 3, 7, 9, 0}, {0, 0, 0, 1, 0, 0, 0, 0, 0},
	// 	{0, 2, 0, 3, 0, 0, 0, 7, 0}, {0, 0, 0, 0, 7, 0, 0, 6, 4}, {1, 0, 0, 0, 0, 0, 0, 0, 0},
	// 	{0, 5, 0, 0, 0, 4, 9, 0, 0}, {0, 7, 2, 0, 0, 0, 0, 0, 0}, {0, 9, 0, 0, 5, 0, 8, 3, 0}}

	//var puzzle2 = [9][9]int{{1, 0, 3, 0, 0, 6, 0, 8, 0}, {0, 5, 0, 0, 8, 0, 1, 2, 0}, {7, 0, 9, 1, 0, 3, 0, 5, 6}, {0, 3, 0, 0, 6, 7, 0, 9, 0}, {5, 0, 7, 8, 0, 0, 0, 3, 0}, {8, 0, 1, 0, 3, 0, 5, 0, 7}, {0, 4, 0, 0, 7, 8, 0, 1, 0}, {6, 0, 8, 0, 0, 2, 0, 4, 0}, {0, 1, 2, 0, 4, 5, 0, 7, 8}}
	// rawInputString := "1 _ 3 _ _ 6 _ 8 _" + "\n" +
	// 	"_ 5 _ _ 8 _ 1 2 _" + "\n" +
	// 	"7 _ 9 1 _ 3 _ 5 6" + "\n" +
	// 	"_ 3 _ _ 6 7 _ 9 _" + "\n" +
	// 	"5 _ 7 8 _ _ _ 3 _" + "\n" +
	// 	"8 _ 1 _ 3 _ 5 _ 7" + "\n" +
	// 	"_ 4 _ _ 7 8 _ 1 _" + "\n" +
	// 	"6 _ 8 _ _ 2 _ 4 _" + "\n" +
	// 	"_ 1 2 _ 4 5 _ 7 8" + "\n"

	//stringReader := strings.NewReader(rawInputString)
	ioReader := bufio.NewReader(os.Stdin)
	outGrid, err := parser.GetInput(ioReader)
	if err != nil {
		log.Panic("ERROR PARSING")
	}
	done, solution := solve(outGrid)
	if !done {
		recurseSolution, board := recurseSolve(&solution)
		if !recurseSolution {
			fmt.Println("No solution exits")
		} else {
			//fmt.Println("Solution exists but recursive one")
			for i := 0; i < 9; i++ {
				for j := 0; j < 9; j++ {
					fmt.Printf("%d ", board[i][j].v)
				}
				fmt.Println("")
			}
		}
	} else {
		//fmt.Println("simple solution exits.")
		for i := 0; i < 9; i++ {
			for j := 0; j < 9; j++ {
				fmt.Printf("%d ", solution[i][j].v)
			}
			fmt.Println("")
		}
	}
}

func findUnassignedLocation(grid *[9][9]cell) (int, int, bool) {
	for y := 0; y < 9; y++ {
		for x := 0; x < 9; x++ {
			if grid[y][x].v == 0 {
				return x, y, true
			}
		} //inner for
	} // outer for
	return -1, -1, false
}

func recurseSolve(grid *[9][9]cell) (bool, *[9][9]cell) {
	x, y, found := findUnassignedLocation(grid)
	if !found {
		return true, grid
	}
	for num := 1; num <= 9; num++ {
		if isSafe(grid, x, y, num) {
			setValue(&grid[y][x], num, grid)
			if flag, cells := recurseSolve(grid); flag {
				return true, cells
			}
			setValue(&grid[y][x], 0, grid)
		}
	}
	var dummy [9][9]cell
	return false, &dummy
}

func isSafe(grid *[9][9]cell, x1, y1, v int) bool {
	for y := 0; y < 9; y++ {
		if grid[y][x1].v == v {
			return false
		}
	}
	for x := 0; x < 9; x++ {
		if grid[y1][x].v == v {
			return false
		}
	}

	xStart, yStart := x1/3*3, y1/3*3

	for y := yStart; y < yStart+3; y++ {
		for x := xStart; x < xStart+3; x++ {
			if grid[y][x].v == v {
				return false
			}
		}
	}

	return true
}
