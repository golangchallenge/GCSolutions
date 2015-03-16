package drum

import (
	"bytes"
	"encoding/binary"
	"io/ioutil"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	p := &Pattern{}
	// Load the file
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	// Decode the Header
	r := bytes.NewReader(b[:headerLen])
	if err = binary.Read(r, binary.LittleEndian, &p.Header); err != nil {
		return nil, err
	}
	// Decode the tracks
	decoded := 0
	for decoded+headerLen < int(p.Len)+magicLen {
		t, n := DecodeTrack(b[decoded+headerLen:])
		p.Tracks = append(p.Tracks, t)
		decoded += n
	}
	return p, nil
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Header
	Tracks []Track
}
