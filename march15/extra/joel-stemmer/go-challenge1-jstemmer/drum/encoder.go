package drum

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// EncodeFile writes the encoded Pattern p to the file pointed at by path.
func EncodeFile(p Pattern, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("drum: could not create file: %s\n", err)
	}
	defer f.Close()
	return NewEncoder(f).Encode(p)
}

// Encoder encodes and writes Patterns to an output stream.
type Encoder struct {
	w   io.Writer
	buf io.Writer
	err error
}

// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}

// Encode writes the encoded version of Patter p to the stream.
func (e *Encoder) Encode(p Pattern) error {
	// first encode the pattern into a buffer, we need the size
	var buf bytes.Buffer
	e.buf = &buf

	// 32-byte version
	version := make([]byte, 32)
	copy(version, []byte(p.Version))
	e.write(binary.BigEndian, version)

	// 4-byte tempo
	e.write(binary.LittleEndian, p.Tempo)

	for _, track := range p.Tracks {
		// 1-byte track identifier
		e.write(binary.BigEndian, track.ID)

		// 4-byte track name length followed by name
		nlen := uint32(len(track.Name))
		if len(track.Name) > maxTrackNameLen {
			return fmt.Errorf("drum: unexpected track name length: %d\n", nlen)
		}
		e.write(binary.BigEndian, nlen)
		e.write(binary.BigEndian, []byte(track.Name))

		// 16 1-byte steps
		for _, step := range track.Steps {
			var s byte
			if step {
				s = 1
			}
			e.write(binary.BigEndian, s)
		}
	}

	// switch the buffer for the actual output writer
	e.buf = e.w

	// 12-byte header
	e.write(binary.BigEndian, []byte(spliceHeader))

	// 2-byte length of the data followed by the data itself and any raw data
	size := uint16(buf.Len())
	e.write(binary.BigEndian, size)
	e.write(binary.BigEndian, buf.Bytes())
	e.write(binary.BigEndian, p.raw)

	if e.err != nil {
		return fmt.Errorf("drum: write error: %s", e.err)
	}
	return nil
}

// write is a convenience method to write into the internal buffer so we reduce
// the number of error checks while writing. It only writes if it hasn't
// encountered an error so far.
func (e *Encoder) write(order binary.ByteOrder, data interface{}) {
	if e.err != nil {
		return
	}
	e.err = binary.Write(e.buf, order, data)
}
