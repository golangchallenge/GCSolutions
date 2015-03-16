package drum

import (
	"bytes"
	"fmt"
)

const endOfString = 0x00

func byteToBool(b byte) (bool, error) {
	switch b {
	case 0:
		return false, nil
	case 1:
		return true, nil
	}
	return false, fmt.Errorf("can not convert %x to bool", b)
}

func cropToString(b []byte) string {
	n := bytes.Index(b, []byte{endOfString})
	if n < 0 {
		n = len(b)
	}
	return string(b[:n])
}
