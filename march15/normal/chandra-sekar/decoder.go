package drum

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

const (
	// Size of the header consisting of version (32 bytes) and tempo (8 bytes).
	headerSize int64 = 32 + 4

	// Size of fixed track info consisting of id (1 byte), length of name (4 bytes)
	// and steps (16 bytes).
	trackFixedSize int64 = 1 + 4 + 16
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	// Discard "SPLICE"
	_, err = io.CopyN(ioutil.Discard, f, 6)
	if err != nil {
		return nil, err
	}

	var lenPat int64
	err = binary.Read(f, binary.BigEndian, &lenPat)
	if err != nil {
		return nil, err
	}

	p := &Pattern{}

	p.Version, err = decodeVersion(f)
	if err != nil {
		return nil, err
	}

	err = binary.Read(f, binary.LittleEndian, &p.Tempo)
	if err != nil {
		return nil, err
	}

	for pos := headerSize; pos < lenPat; {
		t, l, err := decodeTrack(f)
		if err != nil {
			return nil, err
		}

		p.Tracks = append(p.Tracks, t)
		pos += l
	}

	return p, nil
}

// decodeVersion decodes the 32-byte version string. The version string
// is expected to be terminated with '\x00'
func decodeVersion(r io.Reader) (string, error) {
	vbuf := make([]byte, 32)
	_, err := io.ReadFull(r, vbuf)
	if err != nil {
		return "", err
	}

	end := bytes.IndexByte(vbuf, 0)
	return string(vbuf[:end]), nil
}

func decodeTrack(r io.Reader) (t Track, length int64, err error) {
	id := make([]byte, 1)
	_, err = io.ReadFull(r, id)
	if err != nil {
		return t, 0, err
	}

	t.ID = id[0]

	var lenName int32
	err = binary.Read(r, binary.BigEndian, &lenName)
	if err != nil {
		return t, 0, err
	}

	name := make([]byte, lenName)
	_, err = io.ReadFull(r, name)
	if err != nil {
		return t, 0, err
	}

	t.Name = string(name)
	_, err = io.ReadFull(r, t.Steps[:])
	if err != nil {
		return t, 0, err
	}

	return t, trackFixedSize + int64(lenName), nil
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Version string
	Tempo   float32

	Tracks []Track
}

// Track represents a track within a Pattern.
type Track struct {
	ID    byte
	Name  string
	Steps [16]byte
}

func (p *Pattern) String() string {
	var buf bytes.Buffer
	buf.WriteString("Saved with HW Version: ")
	buf.WriteString(p.Version)
	buf.WriteString(fmt.Sprintf("\nTempo: %v\n", p.Tempo))

	for _, t := range p.Tracks {
		buf.WriteString(fmt.Sprintf("(%d) %s\t", t.ID, t.Name))
		for i, s := range t.Steps {
			if i%4 == 0 {
				buf.WriteByte('|')
			}

			if s == 1 {
				buf.WriteByte('x')
			} else {
				buf.WriteByte('-')
			}
		}
		buf.WriteString("|\n")
	}

	return buf.String()
}
