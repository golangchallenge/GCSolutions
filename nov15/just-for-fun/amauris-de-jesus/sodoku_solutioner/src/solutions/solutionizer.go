package solutions;

import (
	//"fmt"
	"math"
	sodoku "sodoku"
)

type Solutionizer struct {
	possibilities int
}

type oppertunity struct {
	I int
	J int
	Entries []int
}

//main entry function that calls the right
//functions for getting sodoku answer
func (inst *Solutionizer) GetSodokuSolution(board *sodoku.Board) string {

	inst.possibilities = 0;

	pass := inst.SetIndicesWithLeastPossibleChoices(board)

	if !pass {
		panic("Can't be solved!")
		return ""
	}

	return board.GetStringFormat()
}

//retrieves the family with the least
//free options to choose from. For example 
//if a particular row only has one blank
//indices then we know we can fill that entry
//with 100% certainty(the unused number will go there)
//returns true if solution was found
func (inst *Solutionizer) SetIndicesWithLeastPossibleChoices(board *sodoku.Board) bool {

	toBeFilled := board.GetEmptyIndices()
	
	if len(toBeFilled)<=0 && board.IsBoardComplete() {
		return true
	}

	oppertunityFound := false
	oppertunities := [9][]oppertunity{}
	boardChange := false

	for  _, v := range(toBeFilled) {
		
		i, j := v[0], v[1]

		numAvailable := inst.getPossibilities(i, j, board)
		length := len(numAvailable)
		
		//if no oppertunities or certainties were found then
		//we are dealing with a faulty/broken board
		if(length<=0) {
			return false
		//if we only have one choice to choose from then we know 100% we can set it
		} else if(length==1) {
			//set indices for relatives
			board.SetEntry(i, j, numAvailable[0])
			boardChange = true
		//other wise we track available oppurtunities thats ordered
		//based on amount of numbers available
		} else {
			oppertunityFound = true
			oppertunities[length] = append(oppertunities[length], oppertunity{i, j, numAvailable})
		}
	}

	//if board was changed we call recursion on updated board
	if boardChange {
		return inst.SetIndicesWithLeastPossibleChoices(board)
	//else if at least one oppurtunity was found
	//we insert oppurtunity and recompute recursion
	//notice how oppertunities are traverse based on order
	//of minimum possibilities. This gives it a much higher chance
	//of success
	} else if oppertunityFound {

		//make copy of current entries before any alterations
		originalEntry := inst.copy(board.Entries)
		inst.possibilities += 1

		for _, ops := range(oppertunities) {

			for _, op := range(ops) {

				if len(op.Entries)<=0 {
					continue
				}

				for _, v := range(op.Entries) {
					board.SetEntry(op.I, op.J, v)
					pass := inst.SetIndicesWithLeastPossibleChoices(board)
					//if recursion returns true
					if pass {
						return true
					//otherwise if this oppertunity wasnt the best choice
					//we set it back to 0 and try next oppertunity
					} else {
						board.SetEntries(originalEntry)
					}
				}
				
			}
			
		}
	}

	return false
}

//creates a copy of multideimensional array
func (inst *Solutionizer) copy(values [][]int) [][]int {

	a := make([][]int, 9)

	// manual deep copy
	for i := range(values) {
	    a[i] = make([]int, len(values[i]))
	    copy(a[i], values[i])
	}
	return a
}

//for a particular empty index, returns the 
//numbers available that can poltentially be
//inserted in that index
func (inst *Solutionizer) getPossibilities(i, j int, board *sodoku.Board) []int {

	families := board.GetFamilies(i, j)
	availableNumbers := 987654321

	for _, family := range(families) {
		availableNumbers = inst.availableNumbers(availableNumbers, family)
	}

	numsAvailable := inst.getPossibilitiesFromAvailableNumbers(availableNumbers)
	
	return numsAvailable
}

//return a number in which 0's represent numbers taken. For, example, from 
//080600320, the numbers 8, 6, 3, and 2 are available to take 
func (inst *Solutionizer) availableNumbers(availableNumbers int, family []int) int {

	for _, v := range(family) {

		if(v==0 || inst.isNumberTaken(availableNumbers, v)) {
			continue
		}

		reducer := v*int(math.Pow(10, float64(v-1)))
		availableNumbers -= reducer
	}

	return availableNumbers
}

//based on 987654321 number format output from availableNumbers,
//getPossibilitiesFromAvailableNumbers converts it to an array of numbers
//for example 080600320, will return []int{8, 6, 3, 2}
func (inst *Solutionizer) getPossibilitiesFromAvailableNumbers(availableNumbers int) []int {

	possibilities := []int{}
	//first get the maxinum nth number available
	max := int(math.Log10(float64(availableNumbers))) + 1

	for max > 0 {

		if(!inst.isNumberTaken(availableNumbers, max)) {
			possibilities = append(possibilities, max)
		}

		max -= 1
	}

	return possibilities
}

//determines wether a number is present in the 987654321 number format.
//for example isNumberTaken(080600320, 2) will return true because 2
//is present
func (inst *Solutionizer) isNumberTaken(availableNumbers, nth int) bool {

	//assumes nth start from 0
	//check if nth position of 987654321 is zero
	//by dividing 987654321 by 10^nth and retrieving remainder
	//then with remainder you divide by 10^(n-1) in order to retrieve
	//nth digit
	nthFloat := float64(nth)
	remainder := availableNumbers%int(math.Pow(10, nthFloat))

	denominator := int(math.Pow(10, nthFloat-1))

	if(denominator==0) {
		return false
	}

	nthNumber := remainder/denominator
	
	isTaken := (nth!=nthNumber)

	return isTaken
}

//difficulty is corrolated to number of possibilities
//found while solving the solution
//the more possibilities, the less chances of success
//you have at choosing the right possibility
func (inst *Solutionizer) Difficulty() string {

	difficulty := "Easy"

	switch {
		case inst.possibilities <= 0:
		    break
		case inst.possibilities <= 5:
		    difficulty = "Medium"
		case inst.possibilities <= 14:
		    difficulty = "Hard"
		case inst.possibilities > 14:
		    difficulty = "Evil"
	}

	return difficulty
}