// Package drum implements an interface for reading, writing and presenting the drum machine patterns.
package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"strings"
)

const (
	patternHeader = "SPLICE"

	stepsSize     = 16
	hwVersionSize = 32

	byteSize   = 1
	uint32Size = 4
	uint64Size = 8
)

var (
	errIncorrectData = errors.New("incorrect binary data format")
)

// readChunk reads the data chunk of the given length from the byte slice.
// Error == io.ErrUnexpectedEOF if the requested length is greater than
// the slice size
func readChunk(data []byte, chunkLength uint64) ([]byte, []byte, error) {
	if uint64(len(data)) < chunkLength {
		return nil, nil, io.ErrUnexpectedEOF
	}
	return data[chunkLength:], data[:chunkLength], nil
}

// readSizedChunk reads the data chunk prepended with its length from
// the byte slice.
// longChunk == true if the length is encoded as a 8-byte uint, big endian;
// otherwise it's a 4-byte uint, big endian.
// Error == io.ErrUnexpectedEOF if the requested (length+chunk) size
// is greater than the slice size
func readSizedChunk(data []byte, longChunk bool) ([]byte, []byte, error) {
	var chunkSize uint64 = uint32Size
	if longChunk {
		chunkSize = uint64Size
	}

	data1, chunk, err := readChunk(data, chunkSize)
	if err != nil {
		return nil, nil, err
	}

	if longChunk {
		return readChunk(data1, binary.BigEndian.Uint64(chunk))
	}
	return readChunk(data1, uint64(binary.BigEndian.Uint32(chunk)))
}

// Track is the high level representation of the single track from
// the drum pattern
type Track struct {
	ID    byte
	Name  string
	Steps [stepsSize]bool
}

// BinarySize returns the size of the binary-encoded track
func (t *Track) BinarySize() uint64 {
	return uint64(len(t.Name)) + byteSize + stepsSize + uint32Size
}

// MarshalBinary encodes the track into its binary form.
// Binary format:
//
//	ID: 1 byte
//	Name: len(Name)(uint32, big endian = 4 bytes) + Name(len(Name) bytes)
//	Steps: 16 bytes
//
func (t *Track) MarshalBinary() ([]byte, error) {

	b := bytes.NewBuffer(nil)
	b.WriteByte(t.ID)

	// len(Name) + Name
	binary.Write(b, binary.BigEndian, uint32(len(t.Name)))
	b.WriteString(t.Name)

	// Steps
	for i := 0; i < stepsSize; i++ {
		if t.Steps[i] {
			b.WriteByte(1)
		} else {
			b.WriteByte(0)
		}
	}
	return b.Bytes(), nil
}

// UnmarshalBinary decodes the track from its binary form
func (t *Track) UnmarshalBinary(data []byte) error {

	// ID
	data1, chunk, err := readChunk(data, byteSize)
	if err != nil {
		return err
	}
	t.ID = chunk[0]

	// len(Name) + Name
	data1, chunk, err = readSizedChunk(data1, false)
	if err != nil {
		return err
	}
	t.Name = string(chunk)

	// Steps
	_, chunk, err = readChunk(data1, stepsSize)
	if err != nil {
		return err
	}
	for i, step := range chunk {
		t.Steps[i] = step == 1
	}
	return nil
}

// String returns textual track representation
func (t *Track) String() string {
	steps := make([]byte, stepsSize+5)
	for i, cur := 0, 0; i < stepsSize; i, cur = i+1, cur+1 {
		if i%4 == 0 {
			steps[cur] = '|'
			cur++
		}
		if t.Steps[i] {
			steps[cur] = 'x'
		} else {
			steps[cur] = '-'
		}
	}
	steps[len(steps)-1] = '|'

	return fmt.Sprintf("(%v) %v\t%s", t.ID, t.Name, steps)
}

// Pattern is the high level representation of the drum pattern
// contained in a .splice file
type Pattern struct {
	HWVersion string
	Tempo     float32
	Tracks    []*Track
}

// BinarySize returns the size of the binary-encoded pattern
func (p *Pattern) BinarySize() uint64 {
	ret := uint64(hwVersionSize) + uint32Size +
		uint64(len(patternHeader)) + uint64Size
	for i := range p.Tracks {
		ret += p.Tracks[i].BinarySize()
	}
	return ret
}

// MarshalBinary encodes the pattern into its binary form.
// Binary format:
//
//	"SPLICE": 6 bytes
//	tail size: uint64 big endian (8 bytes)
//	hw version: 32 bytes
//	tempo: float32 little endian (4 bytes)
//	tracks
//
func (p *Pattern) MarshalBinary() ([]byte, error) {
	b := bytes.NewBuffer(nil)

	// "SPLICE"
	if _, err := b.WriteString(patternHeader); err != nil {
		return nil, err
	}

	// tail size
	if err := binary.Write(b, binary.BigEndian,
		p.BinarySize()-uint64(len(patternHeader))-uint64Size); err != nil {
		return nil, err
	}

	// hw version
	if _, err := b.WriteString(p.HWVersion +
		strings.Repeat("\x00", hwVersionSize-len(p.HWVersion))); err != nil {
		return nil, err
	}

	// tempo
	if err := binary.Write(b, binary.LittleEndian, p.Tempo); err != nil {
		return nil, err
	}

	// tracks
	for _, track := range p.Tracks {
		chunk, err := track.MarshalBinary()
		if err != nil {
			return nil, err
		}
		if _, err := b.Write(chunk); err != nil {
			return nil, err
		}
	}

	return b.Bytes(), nil
}

// UnmarshalBinary decodes the pattern from its binary form
func (p *Pattern) UnmarshalBinary(data []byte) error {

	// "SPLICE"
	data1, chunk, err := readChunk(data, uint64(len(patternHeader)))
	if string(chunk) != patternHeader {
		return errIncorrectData
	}

	// tail
	_, data1, err = readSizedChunk(data1, true)
	if err != nil {
		return err
	}

	// hw version
	data1, chunk, err = readChunk(data1, hwVersionSize)
	if err != nil {
		return err
	}
	p.HWVersion = string(bytes.TrimRight(chunk, "\x00"))

	// tempo
	data1, chunk, err = readChunk(data1, uint32Size)
	if err != nil {
		return err
	}
	p.Tempo = math.Float32frombits(binary.LittleEndian.Uint32(chunk))

	// tracks
	for len(data1) > 0 {
		var t Track
		if err := t.UnmarshalBinary(data1); err != nil {
			return err
		}
		p.Tracks = append(p.Tracks, &t)

		data1 = data1[t.BinarySize():]
	}
	return nil
}

// String returns textual pattern representation
func (p *Pattern) String() string {
	b := bytes.NewBuffer(nil)
	fmt.Fprintf(b, "Saved with HW Version: %v\n", p.HWVersion)
	fmt.Fprintf(b, "Tempo: %v\n", p.Tempo)
	for _, track := range p.Tracks {
		fmt.Fprintln(b, track)
	}
	return b.String()
}

// Decode decodes the buffer with the drum machine backup data
// and returns a pointer to a parsed pattern,
// which is the entry point to the rest of the data
func Decode(data []byte) (*Pattern, error) {

	data1 := data
	for i := bytes.Index(data1, []byte(patternHeader)); i != -1; {
		// look for the next "SPLICE" header in the buffer; if found, try to decode
		// ToDo: consider using KMP algorithm here
		var p Pattern
		if err := p.UnmarshalBinary(data1[i:]); err == nil {
			return &p, nil
		}

		data1 = data1[i+len(patternHeader):]
		i = bytes.Index(data1, []byte(patternHeader))
	}

	return nil, errIncorrectData
}

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern,
// which is the entry point to the rest of the data
func DecodeFile(path string) (*Pattern, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Decode(data)
}
