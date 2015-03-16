package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
)

var header = []byte{83, 80, 76, 73, 67, 69, 0, 0, 0, 0, 0, 0, 0}

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	p := &Pattern{}

	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return p, errors.New("open: error reading file")
	}

	fileBuf := bytes.NewBuffer(raw)

	if bytes.Equal(fileBuf.Next(13), header) { // Check file for a valid header
		var msgLength uint8

		binary.Read(bytes.NewBuffer(fileBuf.Next(1)), binary.LittleEndian, &msgLength) // Get length of message (to pass testcase 5)
		if err != nil {
			return p, errors.New("decode: unable to decode message length")
		}

		msgBuf := bytes.NewBuffer(fileBuf.Next(int(msgLength))) // Create new buffer with the message

		fileBuf.Reset() // Clear the buffer that we're not using anymore

		p.Version = string(bytes.Trim(msgBuf.Next(32), "\x00"))
		p.Tempo = btof(msgBuf.Next(4))

		for msgBuf.Len() > 0 {

			id, err := btoi(msgBuf.Next(1))
			if err != nil {
				return p, errors.New("decode: unable to decode id")
			}

			nameLength, err := btoi(bytes.Trim(msgBuf.Next(4), "\x00"))
			if err != nil {
				return p, errors.New("decode: unable to decode name length")
			}

			name := string(msgBuf.Next(nameLength))
			rhythm := msgBuf.Next(16)

			t := Track{id, name, rhythm}

			p.Tracks = append(p.Tracks, t)

		}
	} else {
		return p, errors.New("decode: missing 'SPLICE' header")
	}

	return p, nil
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Header  [14]byte
	Version string
	Tempo   float32
	Tracks  []Track
}

// Pattern.String converts the Pattern struct to the splice format
// Example:
// Saved with HW Version: 0.808-alpha
// Tempo: 120
// (0) kick	|x---|x---|x---|x---|
// (1) snare	|----|x---|----|x---|
// (2) clap	|----|x-x-|----|----|
// (3) hh-open	|--x-|--x-|x-x-|--x-|
// (4) hh-close	|x---|x---|----|x--x|
// (5) cowbell	|----|----|--x-|----|
func (p *Pattern) String() string {
	s := fmt.Sprintf("Saved with HW Version: %s\nTempo: %v\n", p.Version, p.Tempo)
	for _, t := range p.Tracks {
		s += t.String()
	}
	return s
}

// Track is the representation of the track parts belonging to a
// pattern in a .splice file.
type Track struct {
	ID     int
	Name   string
	Rhythm []byte
}

// Track.String converts the Track struct to the splice format
// Example: (7) hithat |---x|-x-x|---x|-x-x|
func (t *Track) String() string {
	rhythmString := "|"
	for i, b := range t.Rhythm {
		if b == 0 {
			rhythmString += "-"
		} else {
			rhythmString += "x"
		}

		if (i+1)%4 == 0 {
			rhythmString += "|"
		}
	}

	s := fmt.Sprintf("(%d) %s\t%s\n", t.ID, t.Name, rhythmString)
	return s
}

// btof converts a byte array to a float32
func btof(b []byte) float32 {
	var f float32

	buf := bytes.NewBuffer(b)
	binary.Read(buf, binary.LittleEndian, &f)

	return f
}

// btoi converts a byte array to an int
// internally it first converts to an uint8 and finally returns an int
func btoi(b []byte) (int, error) {
	var i uint8

	buf := bytes.NewBuffer(b)
	err := binary.Read(buf, binary.LittleEndian, &i)
	if err != nil {
		return int(i), err
	}

	return int(i), nil
}
