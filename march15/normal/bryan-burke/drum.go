// Package drum implements the decoding of .splice drum machine files.
// See golang-challenge.com/go-challenge1/ for more information
//
// Implementation by Bryan Burke <btburke@fastmail.com>
package drum

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"strings"
)

// headerLength is the fixed size splice file header
const headerLength int = 14

// Read parses the byte representation of a splice file returning a Pattern
// with data from the Header, Version, Tempo and the Tracks
func Read(b []byte) (*Pattern, error) {

	buf := bytes.NewReader(b)

	p := &Pattern{}
	err := p.Header.Read(buf)
	if err != nil {
		return p, err
	}
	// truncate record length if necessary and re-read header
	if int(p.Header.RecordLength)+headerLength < len(b) {
		b = b[:int(p.Header.RecordLength)+headerLength]
		buf = bytes.NewReader(b)
		if err := p.Header.Read(buf); err != nil {
			return p, err
		}
	}

	if err := p.Version.Read(buf); err != nil {
		return p, err
	}

	if err := p.Tempo.Read(buf); err != nil {
		return p, err
	}

	for buf.Len() > 0 {
		t := &Track{}
		if err := t.Read(buf); err != nil {
			return p, err
		}
		p.Tracks = append(p.Tracks, t)
	}
	return p, nil

}

// Header represents the header of a splice file containing the identifier
// and the length of the exported record
type Header struct {
	Identifier   [6]byte
	RecordLength uint64
}

func (h *Header) Read(buf io.Reader) error {
	err := binary.Read(buf, binary.BigEndian, h)
	if err != nil {
		return err
	}
	return nil
}

// Version represents the version number of the splice HW
type Version struct {
	Version [32]byte
}

func (v *Version) Read(buf io.Reader) error {
	err := binary.Read(buf, binary.BigEndian, v)
	if err != nil {
		return err
	}
	return nil
}

func (v *Version) String() string { return strings.TrimRight(fmt.Sprintf("%s", v.Version), "\x00") }

// Tempo represents the tempo for the audio tracks
type Tempo struct {
	Tempo float32
}

func (t *Tempo) Read(buf io.Reader) error {
	err := binary.Read(buf, binary.LittleEndian, t)
	if err != nil {
		return err
	}
	return nil
}

func (t *Tempo) String() string {
	if math.Remainder(float64(t.Tempo), float64(int(t.Tempo))) > 0.1 {
		return fmt.Sprintf("%.1f", t.Tempo)
	}
	return fmt.Sprintf("%d", int(t.Tempo))
}

// Track header contains the Instrument ID and the variable
// length of the instrument name
type trackHeader struct {
	InstrumentID uint8
	_            [3]uint8
	NameLength   uint8
}

func (t *trackHeader) Read(buf io.Reader) error {
	err := binary.Read(buf, binary.LittleEndian, t)
	if err != nil {
		return err
	}
	return nil
}

// Track contains the InstrumentID from the header, its name, and a
// representation of the measure
type Track struct {
	InstrumentID uint8
	Name         []byte
	Measure      []byte
}

func (t *Track) Read(buf io.Reader) error {
	header := &trackHeader{}
	if err := header.Read(buf); err != nil {
		return err
	}

	t.InstrumentID = header.InstrumentID
	t.Name = make([]byte, header.NameLength)
	t.Measure = make([]byte, 16)

	if err := binary.Read(buf, binary.BigEndian, &t.Name); err != nil {
		return err
	}
	if err := binary.Read(buf, binary.BigEndian, &t.Measure); err != nil {
		return err
	}

	return nil
}
