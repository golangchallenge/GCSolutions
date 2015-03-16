package drum

import (
	"encoding/binary"
	"io"
)

// pstring reads and return a pascal-style string from the reader.
// Pascal string is a string prefixed by its length value, which is
// assumed to be a uint8 value
func pstring(r io.Reader) ([]byte, error) {
	var (
		l   uint8
		err error
	)

	if err = binary.Read(r, binary.LittleEndian, &l); err != nil {
		return nil, err
	}

	b := make([]byte, l)
	if err = binary.Read(r, binary.LittleEndian, b); err != nil {
		return b, err
	}

	return b, nil
}
