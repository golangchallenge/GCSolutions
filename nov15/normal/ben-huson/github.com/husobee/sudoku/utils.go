package sudoku

import "bufio"

const (
	// byte values for the allowed tokens, for use with custom
	// scanner split function
	underscore byte = 0x5f
	space           = 0x20
	zero            = 0x30
	one             = 0x31
	nine            = 0x39
	newline         = 0x0a
	// this is the maximum length of a row in bytes
	maxInputRowLength int = 17
)

// isSpace - validate this char is a space
func isSpace(c byte) bool {
	return c == space
}

// isBlank - validate this char is an underscore
func isBlank(c byte) bool {
	return c == underscore
}

// isNumber - validate this char is a number
func isNumber(c byte) bool {
	if _, err := asciiToNumber(c); err != nil {
		return false
	}
	return true
}

// asciiToNumber - convert from ascii representation to uint8
func asciiToNumber(c byte) (uint8, error) {
	if c < one || c > nine {
		return uint8(c), ErrParseInvalidNumber
	}
	return uint8(c) - zero, nil
}

// isEvenNumber - validate this is an even number
func isEvenNumber(i int) bool {
	return i%2 == 0
}

// puzzleScanSplit - This is the customer scanner split function used to both
// parse and validate the stdin representation of the puzzle
func puzzleScanSplit(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// based on scanlines, we will validate each line at a time
	advance, token, err = bufio.ScanLines(data, atEOF)
	if err == nil && token != nil {
		if len(token) != maxInputRowLength {
			// line length is incorrect, error
			err = ErrParseInvalidLineLength
			return
		}
		// check that each line is correct format
		for i, b := range token {
			if isEvenNumber(i) {
				// even, should be either a Number or Blank
				if !isNumber(b) && !isBlank(b) {
					//error
					err = ErrParseInvalidCharacter
					return
				}
			} else {
				// odd, should be space
				if !isSpace(b) {
					err = ErrParseInvalidCharacter
					return
				}
			}
		}
	}
	return
}
