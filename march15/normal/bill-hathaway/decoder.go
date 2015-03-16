package drum

// This file contains the main logic to decode a .splice file
//
// Each .splice file contains the following components
// Header
//   Magic token of 'SPLICE'
//   Hardware version
//   Tempo
//
//  Zero or more tracks of:
//    Index (0-255)
//    Length of instrument name
//    Name of instrument
//    Steps where the sound occured

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
)

// ErrInvalid is returned when we can detect the file is invalid due to missing the Magic value.
var ErrInvalid = errors.New("invalid data")

// Magic is the pattern we expect in the beginning of the file.
var Magic = [6]byte{'S', 'P', 'L', 'I', 'C', 'E'}

type (
	// Pattern consists of the Header and zero or more Tracks.
	Pattern struct {
		Header Header
		Tracks []Track
	}

	// Header contains the magic identifier, hardware version and tempo.
	Header struct {
		Magic     [6]byte
		_         [8]byte
		HWVersion [18]byte
		_         [14]byte
		Tempo     float32
	}

	// Track represents a drum track and contains an index, name, and which steps trigger sound.
	Track struct {
		Index   int32
		NameLen byte
		Name    []byte
		Steps   [16]byte
	}
)

// readTrack reads an individual track.
// If the index is over 255, we have hit corrupt data and just return EOF to allow any previous tracks to be considered valid.
func readTrack(r io.Reader) (Track, error) {
	i := Track{}
	err := binary.Read(r, binary.LittleEndian, &i.Index)
	if err != nil {
		return i, err
	}
	if i.Index > 255 {
		return i, io.EOF
	}
	err = binary.Read(r, binary.LittleEndian, &i.NameLen)
	if err != nil {
		return i, err
	}
	i.Name = make([]byte, i.NameLen)
	err = binary.Read(r, binary.LittleEndian, &i.Name)
	if err != nil {
		return i, err
	}
	err = binary.Read(r, binary.LittleEndian, &i.Steps)
	return i, err
}

// DecodeFile decodes the drum machine file found at the provided path.
// First the header is read, then all tracks until EOF is found.
func DecodeFile(path string) (*Pattern, error) {
	p := Pattern{}
	fh, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fh.Close()
	err = binary.Read(fh, binary.LittleEndian, &p.Header)
	if p.Header.Magic != Magic {
		return nil, ErrInvalid
	}
	for err == nil {
		track, err := readTrack(fh)
		if err != nil {
			if err == io.EOF {
				return &p, nil
			}
			return nil, err
		}
		p.Tracks = append(p.Tracks, track)
	}
	return &p, err
}

// String is the text representation of the drum pattern file.
func (p Pattern) String() string {
	var buf bytes.Buffer
	firstNil := bytes.IndexByte(p.Header.HWVersion[:], 0)
	buf.WriteString(fmt.Sprintf("Saved with HW Version: %s\nTempo: %v\n", string(p.Header.HWVersion[:firstNil]), p.Header.Tempo))
	for _, track := range p.Tracks {
		buf.WriteString(track.String())
	}
	return buf.String()
}

// String is the text representation of an individual drum track.
func (t Track) String() string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("(%d) %s\t", t.Index, string(t.Name[:t.NameLen])))
	for i := 0; i < len(t.Steps); i++ {
		if i%4 == 0 {
			buf.WriteRune('|')
		}
		if t.Steps[i] == 0 {
			buf.WriteRune('-')
		} else {
			buf.WriteRune('x')
		}
	}
	buf.WriteRune('|')
	buf.WriteRune('\n')
	return buf.String()
}
