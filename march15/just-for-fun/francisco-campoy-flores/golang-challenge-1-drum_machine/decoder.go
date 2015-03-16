package drum

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// spliceTag is the tag to be found at the beginning of all drum files.
var spliceTag = [6]byte{'S', 'P', 'L', 'I', 'C', 'E'}

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Decode(f)
}

// Decode decodes and returns a drum Patterh fromt the given io.Reader.
func Decode(ir io.Reader) (*Pattern, error) {
	r, err := newReader(ir)
	if err != nil {
		return nil, err
	}

	var p Pattern
	if err := readString(r, 32, &p.Version); err != nil {
		return nil, fmt.Errorf("parse version: %v", err)
	}
	if err := binary.Read(r, binary.LittleEndian, &p.Tempo); err != nil {
		return nil, fmt.Errorf("parse tempo: %v", err)
	}

	for !r.done() {
		t, err := decodeTrack(r)
		if err != nil {
			return nil, err
		}
		p.Tracks = append(p.Tracks, t)
	}
	return &p, nil
}

// decodeTrack decodes and returns a drum Track from the given reader.
func decodeTrack(r io.Reader) (*Track, error) {
	var t Track
	if err := binary.Read(r, binary.BigEndian, &t.ID); err != nil {
		return nil, fmt.Errorf("parse track id: %v", err)
	}
	var n int32
	if err := binary.Read(r, binary.BigEndian, &n); err != nil {
		return nil, fmt.Errorf("parse track name length: %v", err)
	}
	if err := readString(r, int(n), &t.Name); err != nil {
		return nil, fmt.Errorf("parse track name: %v", err)
	}
	if err := binary.Read(r, binary.BigEndian, &t.Steps); err != nil {
		return nil, fmt.Errorf("parse track %v steps: %v", t.ID, err)
	}
	return &t, nil
}

// reader is a utility type that keeps the count of remaining bytes so we can
// now when we're done.
type reader struct {
	n uint64
	r io.Reader
}

// newReader checks that the given reader contains the beginning of a drum
// Pattern binary format and returns a new reader.
func newReader(r io.Reader) (*reader, error) {
	var tag [6]byte
	if err := binary.Read(r, binary.LittleEndian, tag[:]); err != nil {
		return nil, fmt.Errorf("parse splice tag: %v", err)
	}
	if tag != spliceTag {
		return nil, fmt.Errorf("wrong splice tag: %v", tag)
	}

	var size uint64
	if err := binary.Read(r, binary.BigEndian, &size); err != nil {
		return nil, fmt.Errorf("parse pattern length: %v", err)
	}
	return &reader{size, r}, nil
}

// reader is an io.Reader.
func (r *reader) Read(p []byte) (int, error) {
	n, err := r.r.Read(p)
	if uint64(n) > r.n {
		return 0, fmt.Errorf("reading more than expected")
	}
	r.n -= uint64(n)
	return n, err
}

// done returns true when all bytes have been read.
func (r *reader) done() bool {
	return r.n == 0
}

// readString reads a string of at most n bytes and assigns it to the given
// pointer.
func readString(r io.Reader, n int, s *string) error {
	buf := make([]byte, n)
	n, err := r.Read(buf)
	if err != nil {
		return err
	}
	if i := bytes.IndexByte(buf, 0); i >= 0 {
		n = i
	}
	*s = string(buf[:n])
	return nil
}
