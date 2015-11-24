// author: Jayant Ameta
// https://github.com/wittyameta

package main

import (
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/wittyameta/sudoku-solver/datatypes"
)

const max int = 9
const easy, medium, hard = "easy", "medium", "hard"
const rowIdentifier, colIdentifier, blockIdentifier = "r", "c", "b"

var initIdentifiers = map[string]bool{rowIdentifier: true, colIdentifier: true, blockIdentifier: true}
var numSolutions int

func main() {
	numSolutions = 0
	// create the grid.
	grid := *datatypes.InitGrid()
	count := 0
	inputValues := make(map[int]bool)
	// read input into the grid.
	for i := 0; i < max; i++ {
		count += readRow(&grid, i, inputValues)
	}
	// At least 17 values, and 8 distinct values are required for a unique solution. (Necessary condition, but not sufficient).
	if count < 17 || len(inputValues) < 8 {
		handleError("Too few input values given. At least 17 values, and 8 distinct values must be given.", nil)
	}
	// solve using given inputs without making any guess.
	positions := solve(&grid, count)
	// make a guess for a position and start solving; backtrack if there is any conflict.
	solveByGuessing(&grid, positions, 0)
	fmt.Println("Total solutions:", numSolutions)
	// Set difficulty as easy if solved without making any guess, medium if number of guesses is less than 9, and hard otherwise.
	difficultyLevel := easy
	if len(positions) > 0 {
		if len(positions) < max {
			difficultyLevel = medium
		} else {
			difficultyLevel = hard
		}
	}
	fmt.Println("Difficulty level:", difficultyLevel)
}

// readRow scans the input, verifies it, and sets the value in the grid.
// For each integer in the input grid, the corresponding cell is updated with the value of the input.
// Returns the number of integers in the input.
func readRow(grid *datatypes.Grid, rownum int, inputValues map[int]bool) (count int) {
	var row [max]string
	format := ""
	for i := 0; i < max; i++ {
		format += "%s"
	}
	format += "\n"
	n, err := fmt.Scanf(format, &row[0], &row[1], &row[2], &row[3], &row[4], &row[5], &row[6], &row[7], &row[8])
	if n != max || err != nil {
		handleError("", err)
	}
	for i, elem := range row {
		val := grid[rownum][i].IterationValues
		input := verifyElement(elem)
		if input > 0 {
			if val[0].Possible[input] {
				val[0] = *datatypes.SetValue(input)
				*grid[rownum][i].Val = input
				count++
				inputValues[input] = true
			} else {
				handleError("no solution possible", nil)
			}
		}
	}
	return
}

// verifyElement verifies that the input is either "_" or an integer from 1 to 9.
func verifyElement(elem string) int {
	if "_" == elem {
		return 0
	}
	n, err := strconv.Atoi(elem)
	if err != nil {
		handleError("", err)
	}
	if n < 1 || n > max {
		handleError("number should be from 1 to "+strconv.Itoa(max)+".", nil)
	}
	return n
}

// solve solves the grid. For each value which is set, a goroutine is started to update the grid.
// returns a map with entries for positions which are still not set.
func solve(grid *datatypes.Grid, count int) map[datatypes.Position]bool {
	wg := sync.WaitGroup{}
	wg.Add(count)
	verificationCount := 0
	for i := 0; i < max; i++ {
		for j := 0; j < max; j++ {
			val := *grid[i][j].IterationValues[0].Val
			if val > 0 {
				if verificationCount < count {
					verificationCount++
				} else {
					wg.Add(1)
				}
				go initialElimination(grid, i, j, val, &wg)
			}
		}
	}
	potentialCountDiff := count - verificationCount
	if potentialCountDiff > 0 {
		wg.Add(potentialCountDiff)
	}
	wg.Wait()
	return initPositions(grid)
}

// initialElimination starts solving the grid using val at position{row, column}.
// Iteration count is set to 0 for initial elimination.
func initialElimination(grid *datatypes.Grid, row int, column int, val int, wg *sync.WaitGroup) {
	defer wg.Done()
	if eliminateUsingGivenValues(grid, 0, row, column, val) {
		handleError("No solution", nil)
	}
}

// eliminateUsingGivenValues starts solving the grid using val at position{row, column} for given iteration.
// returns true if there is a conflict while solving for this iteration.
func eliminateUsingGivenValues(grid *datatypes.Grid, iteration int, row int, column int, val int) bool {
	if eliminatePossibilities(grid, iteration, row, column, val, initIdentifiers) {
		return true
	}
	for i := 1; i <= max; i++ {
		if i != val {
			if checkIfUniqueAndEliminate(grid, iteration, row, column, i, initIdentifiers) {
				return true
			}
		}
	}
	return false
}

// eliminatePossibilities is called when a number is set in a cell.
// Eliminates the possibilities from the grid given a number at a row,col
// When eliminating possibilities from a cell in the row, recursive check is done for column and block only.
// When eliminating possibilities from a cell in the column, recursive check is done for row and block only.
// When eliminating possibilities from a cell in the block, recursive check is done for row and column only.
// returns true if there is a conflict while solving for this iteration.
func eliminatePossibilities(grid *datatypes.Grid, iteration int, row int, column int, val int, identifiers map[string]bool) bool {
	for identifier := range identifiers {
		defaultIdentifiers := map[string]bool{rowIdentifier: true, colIdentifier: true, blockIdentifier: true}
		delete(defaultIdentifiers, identifier)
		minPosition, maxPosition := getMinMaxPositions(identifier, datatypes.Position{X: row, Y: column})
		for i := minPosition.X; i <= maxPosition.X; i++ {
			for j := minPosition.Y; j <= maxPosition.Y; j++ {
				if i == row && j == column {
					continue
				}
				if eliminatePossibilitiesForPosition(grid, iteration, i, j, val, defaultIdentifiers) {
					return true
				}
			}
		}
	}
	return false
}

// eliminatePossibilitiesForPosition is called when a number is set in a cell.
// Eliminates the possibility of having val from the Position{i, j} in the grid for the given iteration.
// If the updated cell now has only 1 possibility, then that value is set, and elimination from that cell is called.
// If the cell is updated, then it is checked if the eliminated value now occurs only once
// in the corresponding row/column/block from this position. If so, then the value is set and elimination called.
// returns true if there is a conflict while solving for this iteration.
func eliminatePossibilitiesForPosition(grid *datatypes.Grid, iteration int, i int, j int, val int, identifiers map[string]bool) bool {
	cell := &grid[i][j]
	cell.Mutex.Lock()
	setValue, updated, backtrack := updateCell(cell, iteration, val)
	cell.Mutex.Unlock()
	if backtrack {
		return true
	}
	if setValue > 0 {
		if eliminatePossibilities(grid, iteration, i, j, setValue, identifiers) {
			return true
		}
	}
	if updated {
		if checkIfUniqueAndEliminate(grid, iteration, i, j, val, identifiers) {
			return true
		}
	}
	return false
}

// setValueForCell updates the Value in the cell for given iteration.
// Possibilities are updated so that it only contains the value to be set.
// returns the values removed from the map of possibilities, and a boolean to specify if the value was set.
func setValueForCell(cell *datatypes.Cell, iteration int, setValue int) (eliminatedValues []int, isValueSet bool) {
	existingValue := cell.IterationValues[iteration]
	if (existingValue.Possible[setValue] && *cell.Val == 0) || *cell.Val == setValue {
		*existingValue.Val = setValue
		*cell.Val = setValue
		for key := range existingValue.Possible {
			if key != setValue {
				delete(existingValue.Possible, key)
				eliminatedValues = append(eliminatedValues, key)
			}
		}
		isValueSet = true
		return
	}
	isValueSet = false
	return
}

// updateCell is called as part of eliminatePossibilities.
// Removes the value from the possibilities in the cell
// Returns an integer - if only 1 value is possible for this cell,
// a boolean to specify if the cell was updated , and a boolean to specify a conflict.
func updateCell(cell *datatypes.Cell, iteration int, valToDelete int) (int, bool, bool) {
	existingValue := cell.IterationValues[iteration]
	if *cell.Val == valToDelete {
		return 0, false, true
	}
	updated := false
	setValue := 0
	if *cell.Val == 0 && existingValue.Possible[valToDelete] {
		updated = true
		delete(existingValue.Possible, valToDelete)
		if len(existingValue.Possible) == 1 {
			for key := range existingValue.Possible {
				setValue = key
				*existingValue.Val = key
				*cell.Val = key
			}
		}
	}
	return setValue, updated, false
}

// checkIfUniqueAndEliminate checks if the eliminated value now occurs only once
// in the corresponding identifiers (row/column/block) from this position. If so, then the value is set and elimination called.
// returns true if there is a conflict while solving for this iteration.
func checkIfUniqueAndEliminate(grid *datatypes.Grid, iteration int, i int, j int, val int, identifiers map[string]bool) bool {
	uniquePositions, conflict := checkIfUnique(grid, iteration, val, datatypes.Position{X: i, Y: j}, identifiers)
	if conflict {
		return true
	}
	for _, pos := range uniquePositions {
		setCell := &grid[pos.X][pos.Y]
		setCell.Mutex.Lock()
		eliminatedValues, isValueSet := setValueForCell(setCell, iteration, val)
		setCell.Mutex.Unlock()
		if isValueSet {
			for _, eliminatedVal := range eliminatedValues {
				if checkIfUniqueAndEliminate(grid, iteration, pos.X, pos.Y, eliminatedVal, initIdentifiers) {
					return true
				}
			}
			if eliminatePossibilities(grid, iteration, pos.X, pos.Y, val, initIdentifiers) {
				return true
			}
		} else {
			return true
		}
	}
	return false
}

// checkIfUnique checks if the value deleted now exists once in the identifiers(row/block/col), then the cell is returned.
// returns uniquePositions array, and a boolean to specify if there was a conflict
func checkIfUnique(grid *datatypes.Grid, iteration int, valDeleted int, pos datatypes.Position, identifiers map[string]bool) ([]datatypes.Position, bool) {
	var uniquePositions []datatypes.Position
	var uniquePos datatypes.Position
	var foundUnique, atLeastOnce bool

	for identifier := range identifiers {
		uniquePos, foundUnique, atLeastOnce = checkIfUniqueWithIdentifier(grid, iteration, valDeleted, pos, identifier)
		if foundUnique {
			uniquePositions = append(uniquePositions, uniquePos)
		} else if !atLeastOnce {
			return uniquePositions, true
		}
	}
	return uniquePositions, false
}

// checkIfUniqueWithIdentifier checks if the value deleted now exists once in the identifier(row/block/col), then the cell is returned.
// Returns uniquePosition, a boolean to specify if unique position was found,
// and a boolean to specify if there was at least one position with this value - meaning there is no conflict.
func checkIfUniqueWithIdentifier(grid *datatypes.Grid, iteration int, valDeleted int, pos datatypes.Position, identifier string) (datatypes.Position, bool, bool) {
	minPosition, maxPosition := getMinMaxPositions(identifier, pos)
	row := pos.X
	column := pos.Y
	found := false
	for i := minPosition.X; i <= maxPosition.X; i++ {
		for j := minPosition.Y; j <= maxPosition.Y; j++ {
			val := grid[i][j].IterationValues[iteration]
			cell := grid[i][j]
			if *cell.Val == valDeleted {
				return pos, false, true
			}
			if val.Possible[valDeleted] {
				if found {
					return pos, false, true
				}
				found = true
				row = i
				column = j
			}
		}
	}
	if found {
		return datatypes.Position{X: row, Y: column}, true, true
	}
	return pos, false, false
}

// getMinMaxPositions gives the min and max positions for the identifier.
// The min and max give the range to check for any conflict or elimination.
// For example: if the identifier is 'rowIdentifier', then the minPos to maxPos will be the whole row ({row,0} to {row,8}).
func getMinMaxPositions(identifier string, pos datatypes.Position) (minPos datatypes.Position, maxPos datatypes.Position) {
	if identifier == rowIdentifier {
		return datatypes.Position{X: pos.X, Y: 0}, datatypes.Position{X: pos.X, Y: max - 1}
	}
	if identifier == colIdentifier {
		return datatypes.Position{X: 0, Y: pos.Y}, datatypes.Position{X: max - 1, Y: pos.Y}
	}
	if identifier == blockIdentifier {
		leftX, leftY := getBlockTopLeft(pos.X, pos.Y)
		return datatypes.Position{X: leftX, Y: leftY}, datatypes.Position{X: leftX + 2, Y: leftY + 2}
	}
	return datatypes.Position{X: 0, Y: 0}, datatypes.Position{X: max - 1, Y: max - 1}
}

// getBlockTopLeft returns the position of the top-left cell from the same block.
func getBlockTopLeft(x int, y int) (int, int) {
	return x - x%3, y - y%3
}

// remainingPositions returns the map with positions where the value is not yet set.
func remainingPositions(grid *datatypes.Grid, positions map[datatypes.Position]bool) map[datatypes.Position]bool {
	emptyPositions := make(map[datatypes.Position]bool)
	for pos := range positions {
		if *grid[pos.X][pos.Y].Val == 0 {
			emptyPositions[pos] = true
		}
	}
	return emptyPositions
}

// initPositions returns the map with positions where the value is not yet set, before solveByGuessing is called.
func initPositions(grid *datatypes.Grid) map[datatypes.Position]bool {
	positions := make(map[datatypes.Position]bool)
	index := 0
	for i := 0; i < max; i++ {
		for j := 0; j < max; j++ {
			positions[datatypes.Position{X: i, Y: j}] = true
			index++
		}
	}
	return remainingPositions(grid, positions)
}

// solveByGuessing selects the position with minimum possibilities out of the remaining empty positions.
// For each of the possible values, the grid is solved. For each conflict, the state is backtracked.
// If no conflict is there, then recursively solveByGuessing on the remaining empty positions.
// Prints the solution, if found. Also increments the number of solutions.
func solveByGuessing(grid *datatypes.Grid, positions map[datatypes.Position]bool, iteration int) {
	// if all positions have been filled, then return
	if len(positions) == 0 {
		numSolutions++
		grid.Print()
		return
	}
	// copy remaining positions to next iteration, and start guessing for the position with minimum possibilities.
	pos := copyValuesForNextIteration(grid, positions, iteration)
	existingValue := grid[pos.X][pos.Y].IterationValues[iteration]
	for val := range existingValue.Possible {
		// update the cell with val for next iteration
		nextValue := grid[pos.X][pos.Y].IterationValues[iteration+1]
		*grid[pos.X][pos.Y].Val = val
		*nextValue.Val = val
		for key := range nextValue.Possible {
			if key != val {
				delete(nextValue.Possible, key)
			}
		}
		// start solving using the set value.
		if !eliminateUsingGivenValues(grid, iteration+1, pos.X, pos.Y, val) {
			// if no conflict, then call solveByGuessing for remaining positions.
			updatedPositions := remainingPositions(grid, positions)
			solveByGuessing(grid, updatedPositions, iteration+1)
		}
		// backtrack to previous state
		copyValuesForNextIteration(grid, positions, iteration)
	}
	return
}

// copyValuesForNextIteration copies the values of the cells at given positions from current iteration to next.
// Returns the position with minimum number of possible values.
func copyValuesForNextIteration(grid *datatypes.Grid, positions map[datatypes.Position]bool, iteration int) (minPos datatypes.Position) {
	minPossibilities := max + 1
	for pos := range positions {
		cell := grid[pos.X][pos.Y]
		cell.IterationValues[iteration+1] = *datatypes.CopyValue(cell.IterationValues[iteration])
		*cell.Val = 0
		countPossibilities := len(cell.IterationValues[iteration].Possible)
		if countPossibilities < minPossibilities {
			minPossibilities = countPossibilities
			minPos = pos
		}
	}
	return
}

// handleError prints error message, and exits the program.
func handleError(msg string, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
	} else {
		fmt.Fprintf(os.Stderr, "error: %v\n", msg)
	}
	os.Exit(1)
}
