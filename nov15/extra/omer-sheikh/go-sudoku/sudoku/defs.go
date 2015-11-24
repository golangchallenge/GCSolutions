package sudoku

const blockLen = 3
const gridLen = 9
const constraintTypes = 4
const numCells = gridLen * gridLen
const maxCols = constraintTypes * numCells
const maxRows = numCells * gridLen

const rowConstraintsOff = numCells
const colConstraintsOff = numCells * 2
const blkConstraintsOff = numCells * 3

const (
	unranked = iota // Do not rank Puzzles with multiple solutions
	easy
	medium
	hard
	evil
)

func rankMessage(r int) string {
	switch {
	case evil == r:
		return "Difficulty: Evil"
	case hard == r:
		return "Difficulty: Hard"
	case medium == r:
		return "Difficulty: Medium"
	case easy == r:
		return "Difficulty: Easy"
	case unranked == r:
		return "Multiple solutions; no difficulty ranking"
	}
	return "Difficulty rating not determined."
}
