package drum

import (
	"io"
	"os"
)

// Encoder writes binary-encoded Pattern to output stream.
type Encoder struct {
	w io.Writer
}

// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

// Encode writes the binary encoding of pattern to the stream.
func (e *Encoder) Encode(pattern *Pattern) error {
	b, err := pattern.MarshalBinary()
	if err != nil {
		return err
	}

	if _, err = e.w.Write(b); err != nil {
		return err
	}

	return nil
}

// EncodeFile writes the binary encoding of pattern to a file named filename.
func EncodeFile(filename string, pattern *Pattern) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}

	return NewEncoder(f).Encode(pattern)
}
