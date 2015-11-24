package sudoku

import "errors"

var (
	InputTooShortError  = errors.New("Input is too short!")
	InputMalformedError = errors.New("Input is malformed!")
)
