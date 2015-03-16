package drum

import (
	"bytes"
	"encoding/binary"
	"io"
	"os"
)

// decodeHeader reads a drum machine file's header and returns
// the decoded header and an error, if any.
func decodeHeader(r io.Reader) (*header, error) {
	var h header

	err := binary.Read(r, binary.LittleEndian, &h)
	if err != nil {
		return nil, err
	}

	if h.Magic != spliceMagic {
		return nil, ErrNotSplice
	}
	return &h, nil
}

// Decode reads a Splice drum machine file from r.
// It returns a pointer to a drum kit pattern as a Pattern and an error, if any.
func Decode(r io.Reader) (*Pattern, error) {
	h, err := decodeHeader(r)
	if err != nil {
		return nil, err
	}

	p := Pattern{
		Version: bytes.TrimRight(h.Version[0:], "\x00"), // strip null
		Tempo:   float64(h.Tempo),
	}

	// Limit reads to the file size in the header
	lr := io.LimitReader(r, int64(h.Size))

	for {
		t, err := ReadTrack(lr)
		if err != nil {
			break
		}
		p.Tracks = append(p.Tracks, *t)
	}
	if err != nil && err != io.EOF {
		return nil, err
	}

	return &p, nil
}

// DecodeFile decodes the Splice drum machine file found at the provided path.
// It returns a pointer to a drum kit pattern as a Pattern and an error, if any.
func DecodeFile(path string) (*Pattern, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return Decode(f)
}
