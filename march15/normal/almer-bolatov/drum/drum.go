// Package drum is supposed to implement the decoding of .splice drum machine files.
// See golang-challenge.com/go-challenge1/ for more information
package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"
)

// Number of bytes used in decoding
const (
	i8 = 1 << iota
	i16
	i32
	i64
	i128
	i256
)

// decode parses the encoded data and stores the result in the value
// pointed by p.
func decode(b []byte, p *Pattern) error {
	// Skip header 'SPLICE'
	var offset = 6

	// Total number of important bytes
	var size int64
	err := decodeNumeric(b, binary.BigEndian, &size, offset)
	if err != nil {
		return err
	}
	offset += i64

	// Pattern.Version
	version, err := decodeString(b, offset, i256)
	if err != nil {
		return err
	}
	offset += i256

	// Pattern.Tempo
	var tempo float32
	err = decodeNumeric(b, binary.LittleEndian, &tempo, offset)
	if err != nil {
		return err
	}
	offset += i32

	//
	// Pattern.Tracks
	//
	var ts []Track
	for int64(offset) < size {
		// Track.ID
		var ID uint8
		err = decodeNumeric(b, binary.BigEndian, &ID, offset)
		if err != nil {
			return err
		}
		offset += i8

		// Track name size
		var tsz uint32
		err = decodeNumeric(b, binary.BigEndian, &tsz, offset)
		if err != nil {
			return err
		}
		offset += i32

		// Track.Name
		name, err := decodeString(b, offset, (int)(tsz))
		if err != nil {
			return err
		}
		offset += (int)(tsz)

		//
		// Track.Steps
		//
		is := make([]uint8, 16)
		err = decodeNumeric(b, binary.BigEndian, is, offset)
		if err != nil {
			return err
		}
		offset += len(is)

		// Move uint steps to bool steps
		var bs [16]bool
		for i, v := range is {
			if v != 0 {
				bs[i] = true
			}
		}

		ts = append(ts, Track{ID, name, bs})
	}

	// No error occurred. Return decoded pattern
	*p = Pattern{version, tempo, ts}
	return nil
}

// decodeNumeric decodes structured binary data from b into data. Wrapper for
// binary.Read(). Offset argument specifies the offset position in b slice.
func decodeNumeric(b []byte, order binary.ByteOrder, data interface{}, offset int) error {
	if offset >= len(b) {
		return errors.New("array index out of bounds")
	}

	buf := bytes.NewReader(b[offset:])
	return binary.Read(buf, order, data)
}

// decodeString decodes a string from b in the range specified by offset and
// length of the expected string. If the string is shorter than the specified
// range, the non-ascii suffix is ignored.
func decodeString(b []byte, offset int, length int) (string, error) {
	if offset >= len(b) || offset+length >= len(b) {
		return "", errors.New("array index out of bounds")
	}
	bb := b[offset : offset+length]
	n := bytes.Index(bb, []byte{0})
	if n == -1 {
		n = len(bb)
	}
	return string(bb[:n]), nil
}

// String implements a String() interface for Pattern
func (p Pattern) String() string {
	var tracks []string
	for _, t := range p.Tracks {
		tracks = append(tracks, fmt.Sprint(t))
	}
	return fmt.Sprintf("Saved with HW Version: %v\nTempo: %v\n%v\n",
		p.Version, p.Tempo, strings.Join(tracks, "\n"))
}

// String implements a String() interface for Track
func (t Track) String() string {
	var steps string
	for i, is := range t.Steps {
		if i%4 == 0 {
			steps += "|"
		}
		if is {
			steps += "x"
		} else {
			steps += "-"
		}

		if i == len(t.Steps)-1 {
			steps += "|"
		}
	}
	return fmt.Sprintf("(%v) %v\t%v", t.ID, t.Name, steps)
}
