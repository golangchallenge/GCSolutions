package sudoku

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

// ParseReader parses Sudoku from Reader and returns it in 81 character string format.
func ParseReader(r io.Reader) (string, error) {
	bs := make([]byte, 9*9)

	s := bufio.NewScanner(r)
	for i, l := 0, 1; l <= 9; l++ {
		if !s.Scan() {
			return "", fmt.Errorf("line %d: unexpected end of input", l)
		}

		line := s.Bytes()
		lr := bytes.NewReader(line)
		ls := bufio.NewScanner(lr)
		ls.Split(bufio.ScanWords)

		for c := 1; c <= 9; c++ {
			if !ls.Scan() {
				return "", fmt.Errorf("line %d: unexpected end of input", l)
			}
			t := ls.Bytes()
			tc := t[0]
			if len(t) > 1 || ((tc != '_') && ('0' > tc || tc > '9')) {
				return "", fmt.Errorf("line %d, column %d: expected digit or underscore symbol", l, c)
			}
			bs[i] = tc
			i++
		}
		if ls.Scan() {
			return "", fmt.Errorf("line %d: too many symbols", l)
		}
	}

	return string(bs), nil
}
