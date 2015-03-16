package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"os"
)

// End indices of header data
const (
	// ePrefix is an end index of prefix data.
	ePrefix = 6
	// eRemLen is an end index of remaining length data.
	eRemLen = 14
	// eVersion is an end index of version data.
	eVersion = 46
	// eTempo is an end index of tempo data.
	eTempo = 50
)

// Lengths of data
const (
	// lenVersion is a length of version data.
	lenVersion = eVersion - eRemLen
	// lenTempo is a length of tempo data.
	lenTempo = eTempo - eVersion
	// lenIDNameLen is a length of track id and
	// track name length data.
	lenIDNameLen = 5
	// lenSteps is a length of track steps data
	lenSteps = 16
)

// Initial capacity of the tracks
const capTracks = 8

// prefix is prefix data.
var prefix = []byte{83, 80, 76, 73, 67, 69}

// Error values
var (
	errInvalidData = errors.New("invalid data")
)

// decoder reads drum machine file data from an input stream.
type decoder struct {
	r   io.Reader
	max int
	n   int
}

// decode reads the drum machine file data from the reader,
// stores them in the pattern.
func (dec *decoder) decode(p *Pattern) error {
	// Read a header.
	b := make([]byte, eTempo)
	if err := dec.read(b); err != nil {
		return err
	}

	// Validate a prefix of the header.
	if !bytes.Equal(b[:ePrefix], prefix) {
		return errInvalidData
	}

	// Extract a remaining length from the header
	// and set it to the decoder as a maximum
	// size to read.
	dec.max = encodeInt(b[ePrefix:eRemLen]) - lenVersion - lenTempo

	// Extract a version from the header, trim it
	// and set it to the pattern.
	// Refrain from using bytes.Trim when trimming
	// to decrease memory allocation.
	e := eVersion
	for eRemLen < e {
		if b[e-1] != 0x00 {
			break
		}
		e--
	}
	p.Version = b[eRemLen:e]

	// Extract a tempo from the header
	// and set it to the pattern.
	if err := binary.Read(bytes.NewReader(b[eVersion:]), binary.LittleEndian, &p.Tempo); err != nil {
		return err
	}

	// Read track.
	// Refrain from reading the whole remaining data
	// at one time because the size of the remaining data
	// might be etremely large.
	for dec.n < dec.max {
		// Read a track id and a track name length.
		b := make([]byte, lenIDNameLen)
		if err := dec.read(b); err != nil {
			return err
		}

		// Extrack a track id.
		id := int(b[0])

		// Extrack a track name length.
		ln := encodeInt(b[1:])

		// Read a track name and steps.
		b = make([]byte, ln+lenSteps)
		if err := dec.read(b); err != nil {
			return err
		}

		// Extrack a track and steps and create a track.
		t := &Track{
			ID:    id,
			Name:  b[:ln],
			Steps: b[ln:],
		}

		// Append the track to the Pattern.
		p.Tracks = append(p.Tracks, t)
	}

	return nil
}

// read reads data from the reader.
func (dec *decoder) read(p []byte) error {
	n, err := dec.r.Read(p)
	if err != nil {
		return err
	}
	if n != len(p) {
		return errInvalidData
	}

	// Increment the n of the decoder if the maximum
	// size to read of the decoder is set.
	if dec.max > 0 {
		dec.n += n
	}

	return nil
}

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	dec := &decoder{r: f}

	p := &Pattern{
		Tracks: make([]*Track, 0, capTracks),
	}

	if err := dec.decode(p); err != nil {
		return nil, err
	}

	return p, nil
}

// encodeInt converts a byte slice to int.
// It can be used instead of binary.Read
// to decrease memory allocation when we
// need to convert a byte slice to int in
// Big Endian order.
func encodeInt(b []byte) int {
	var n int
	for i, l := 0, len(b); i < l; i++ {
		n += int(b[i]) * int(math.Pow(16, float64(l-i-1)))
	}
	return n
}
