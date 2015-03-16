package drum

import (
	"bytes"
	"encoding/binary"
)

// A PatternHeaderParser is responsible for parsing a PatternHeader.
type PatternHeaderParser struct{}

// NewPatternHeaderParser creates and returns a PatternHeaderParser.
func NewPatternHeaderParser() *PatternHeaderParser {
	return &PatternHeaderParser{}
}

// Parse returns a parsed PatternHeader given a valid byte slice.
func (php *PatternHeaderParser) Parse(b []byte) (PatternHeader, error) {
	var ph PatternHeader
	buf := bytes.NewReader(b[:52])
	err := binary.Read(buf, binary.LittleEndian, &ph)
	if err != nil {
		return ph, err
	}

	return ph, nil
}
