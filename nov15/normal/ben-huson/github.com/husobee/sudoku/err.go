package sudoku

import "errors"

var (
	// ErrParseInvalidNumber - invalid numerical in input
	ErrParseInvalidNumber = errors.New("invalid number")
	// ErrParseInvalidCharacter - invalid character in input
	ErrParseInvalidCharacter = errors.New("invalid character")
	// ErrParseInvalidLineLength - invalid line length in input
	ErrParseInvalidLineLength = errors.New("invalid line length")
	// ErrParseInvalidRowCount - invalid number of rows in input
	ErrParseInvalidRowCount = errors.New("invalid number of rows")
	// ErrSolveStuck - cannot solve this puzzle
	ErrSolveStuck = errors.New("stuck in backtrack")
	// ErrSolveNoSolution - no solution to puzzle
	ErrSolveNoSolution = errors.New("no solution")
	// ErrSolveExceedRecursionDepth - no solution to puzzle, exceeded recursion depth
	ErrSolveExceedRecursionDepth = errors.New("exceeded recursion depth")
)
