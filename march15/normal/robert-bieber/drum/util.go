package drum

import (
	"io"
)

func readCString(fin io.Reader, maxLength int) string {
	bin := make([]byte, 1)
	bout := make([]byte, 0, 8)

	for len(bout) < maxLength {
		n, err := fin.Read(bin)
		if err == nil && n == 1 && bin[0] != 0 {
			bout = append(bout, bin[0])
		} else {
			break
		}
	}

	return string(bout)
}
