package drum

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

const (
	// HeaderSignature is the signature by which a .splice file is identified
	HeaderSignature = "SPLICE"
	// VersionSize is size of the version in a .splice file in bytes
	VersionSize = 32
	// TempoSize is the size of the tempo in a .splice file in bytes
	TempoSize = 4
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	sp, err := NewSplice(path)
	if err != nil {
		return nil, err
	}
	p := &Pattern{}
	p.Version, err = sp.Version()
	if err != nil {
		return nil, err
	}
	p.Tempo, err = sp.Tempo()
	if err != nil {
		return nil, err
	}
	p.Tracks, err = sp.Tracks()
	if err != nil {
		return nil, err
	}
	return p, nil
}

// ErrInvalidSignature denotes an error while checking for the .splice signature
//
// The error will hold the actual len(signature) bytes present
type ErrInvalidSignature struct {
	sig []byte
}

// Error to implement the error interface
func (e ErrInvalidSignature) Error() string {
	return fmt.Sprintf("invalid file signature: %v", hex.Dump(e.sig))
}

// Splice represents a .splice file for further processing
type Splice struct {
	BodyLength int64
	body       []byte
}

// NewSplice creates a new .splice file representation for the given file path
func NewSplice(path string) (*Splice, error) {
	fd, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fd.Close()
	buf := bufio.NewReader(fd)

	sp := &Splice{}
	err = sp.readHeader(buf)
	if err != nil {
		return nil, err
	}

	sp.body = make([]byte, int(sp.BodyLength))
	read, err := buf.Read(sp.body)
	if err != nil {
		return nil, err
	}
	if read != int(sp.BodyLength) {
		return nil, fmt.Errorf("invalid body. expected length %d, read %d", sp.BodyLength, read)
	}
	return sp, nil
}

// readHeader checks for a correct header signature and sets the body length
func (s *Splice) readHeader(r io.Reader) error {
	var err error
	if err = s.checkSignature(r); err != nil {
		return err
	}
	err = binary.Read(r, binary.BigEndian, &s.BodyLength)
	if err != nil {
		return err
	}
	return nil
}

func (s *Splice) checkSignature(r io.Reader) error {
	sig := make([]byte, len(HeaderSignature))
	_, err := r.Read(sig)
	if err != nil {
		return err
	}
	if string(sig) != HeaderSignature {
		return ErrInvalidSignature{sig}
	}
	return nil
}

// Body returns the .splice file body as a byte buffer
func (s *Splice) Body() *bytes.Buffer {
	return bytes.NewBuffer(s.body)
}

// Version returns the version of the .splice file
func (s *Splice) Version() (string, error) {
	buf := make([]byte, VersionSize)
	_, err := s.Body().Read(buf)
	if err != nil {
		return "", err
	}
	return string(bytes.TrimRight(buf, "\x00")), nil
}

// Tempo returns the tempo of the drum pattern from the .splice file
func (s *Splice) Tempo() (float32, error) {
	var f float32
	buf := s.Body()
	// discard version
	_ = buf.Next(VersionSize)
	err := binary.Read(buf, binary.LittleEndian, &f)
	return f, err
}

// track is the internal representation of a .splice track
//
// It implements the Tracker for further processing.
type track struct {
	id    uint8
	name  string
	track []byte
}

// ID returns the track ID
func (t track) ID() uint8 {
	return t.id
}

// Name returns the track Name
func (t track) Name() string {
	return t.name
}

// Track returns the byte representation of the track
func (t track) Track() []byte {
	return t.track
}

// Beats returns the number of the beats in a track
//
// .splice file currently supports only 4 beats
func (t track) Beats() int {
	return 4
}

// StepsPerBeat returns the number of the steps per beat
//
// .splice currently supports only 4 steps per beat for a 4/4 beat = 16 steps
func (t track) StepsPerBeat() int {
	return 4
}

// newTrack creates a new track from the .splice file body buffer
//
// It expects the cursor of the buffer to be advanced to the correct position.
// The cursor will be advanced. If no track could be read from that position,
// newTrack() will return an error.
func newTrack(buf io.Reader) (*track, error) {
	t := &track{}
	err := binary.Read(buf, binary.BigEndian, &t.id)
	if err != nil {
		return nil, err
	}
	var nameLen int32
	err = binary.Read(buf, binary.BigEndian, &nameLen)
	if err != nil {
		return nil, err
	}
	nm := make([]byte, int(nameLen))
	_, err = buf.Read(nm)
	if err != nil {
		return nil, err
	}
	t.name = string(nm)
	t.track = make([]byte, 16)
	n, err := buf.Read(t.track)
	if n != 16 {
		return nil, fmt.Errorf("invalid pattern. expect len %d, got %d", 16, n)
	}
	return t, err
}

// Tracks returns the track representation
func (s *Splice) Tracks() ([]Tracker, error) {
	var tr *track
	var err error
	tracks := make([]Tracker, 0)
	buf := s.Body()
	// discard version, tempo
	_ = buf.Next(VersionSize + TempoSize)
	for {
		tr, err = newTrack(buf)
		if err != nil && err != io.EOF {
			return nil, err
		}
		if err == io.EOF {
			return tracks, nil
		}
		tracks = append(tracks, tr)
	}
}
