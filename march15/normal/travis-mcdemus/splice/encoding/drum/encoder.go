package drum

import (
	"encoding/binary"
	"io"
	"os"
)

// EncodeFile a text file backup in a custom binary data format.
func EncodeFile(path string, pattern Pattern) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	e := NewEncoder(f)
	err = e.Encode(pattern)
	return err
}

// An Encoder represents a parser to binary from a drum pattern.
type Encoder struct {
	w io.Writer
}

// NewEncoder creates a new drum pattern encoder writing to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w}
}

// Encode writes to the encoder's output stream to serialize a drum pattern.
func (e Encoder) Encode(p Pattern) error {
	if err := binary.Write(e.w, binary.LittleEndian, p.header()); err != nil {
		return err
	}
	for _, t := range p.Tracks {
		if err := binary.Write(e.w, binary.LittleEndian, t.encode()); err != nil {
			return err
		}
	}
	return nil
}
