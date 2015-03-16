package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	// The first 6 bytes of every valid .splice file.
	fileHeader = "SPLICE"

	// Number of bytes allowed for specifying the pattern version.
	versionMaxLen = 32

	// Tempo occupies 4 bytes.
	tempoSize = 4
)

var fileHeaderLen = len(fileHeader)

// A Decoder reads and decodes a drum pattern
// from a binary input stream.
//
// The binary format for a pattern is as follows:
//
// - The first 32 bytes are allocated for the version, a string.
// - The next 4 bytes are the tempo, a float32, little endian byte order.
// - All bytes after this describe pattern's tracks.
//
// The binary format for a track is as follows:
//
// - The first 4 bytes are the track ID, a uint32, little endian byte order.
// - The next byte is the length of the track name, a uint8.
// - The next `name length` bytes are the track name, a string.
// - The following 16 bytes are the track's steps, where a value of `1` means
//   trigger the track on a given step, and a value of  `0` means don't.
type Decoder struct {
	r   io.Reader
	off int
	err error

	// bc is byte count of the pattern data
	bc uint64
}

// NewDecoder returns a new pattern decoder that reads
// from r.
//
// As per the format of .splice files, the decoder only
// reads the first 14 bytes, which contain the header (6 bytes)
// and byte count (8 bytes), a uint64, and then a further `byte count`
// bytes, which should contain the pattern in the correct binary
// format.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r, off: 0}
}

// Decode decodes a pattern from the binary data in its input
// and stores the result in the Pattern struct pointed to by p.
func (d *Decoder) Decode(p *Pattern) error {
	if err := d.stripHeader(); err != nil {
		return fmt.Errorf("couldn't strip file header: %s", err)
	}

	if err := d.parseByteCount(); err != nil {
		return fmt.Errorf("couldn't parse byte count: %s", err)
	}

	ver, err := d.parseVersion()
	if err != nil {
		return fmt.Errorf("couldn't parse version: %s", err)
	}
	p.Version = ver

	tempo, err := d.parseTempo()
	if err != nil {
		return fmt.Errorf("couldn't parse tempo: %s", err)
	}
	p.Tempo = tempo

	for d.off < int(d.bc) {
		t, err := d.parseTrack()
		if err != nil {
			return fmt.Errorf("failed to decode track: %s", err)
		}

		p.Tracks = append(p.Tracks, t)
	}

	return nil
}

func (d *Decoder) stripHeader() error {
	buf := make([]byte, fileHeaderLen)
	n, err := d.r.Read(buf)

	switch {
	case err != nil:
		return err
	case n != fileHeaderLen:
		return io.ErrUnexpectedEOF
	case !bytes.Equal([]byte(fileHeader), buf):
		return errors.New("invalid file format")
	default:
		return err
	}
}

func (d *Decoder) parseByteCount() error {
	return binary.Read(d.r, binary.BigEndian, &d.bc)
}

func (d *Decoder) parseVersion() (string, error) {
	buf := make([]byte, versionMaxLen)
	n, err := d.r.Read(buf)

	switch {
	case err != nil:
		return "", err
	case n != versionMaxLen:
		return "", io.ErrUnexpectedEOF
	default:
		d.off += versionMaxLen
		// Only trimming right allows null characters at start and in
		// the middle of the string. This feels correct, although not
		// sure why anyone would have null characters before/within a string.
		return strings.TrimRight(string(buf), "\x00"), nil
	}
}

func (d *Decoder) parseTempo() (float32, error) {
	var tempo float32
	if err := binary.Read(d.r, binary.LittleEndian, &tempo); err != nil {
		return tempo, err
	}

	d.off += tempoSize

	return tempo, nil
}

func (d *Decoder) parseTrack() (*Track, error) {
	t := &Track{}

	td := newTrackDecoder(d.r)
	if err := td.decode(t); err != nil {
		return nil, err
	}

	d.off += td.off

	return t, nil
}

////////////

// Track ID occupies 4 bytes.
const trackIDSize = 4

// A trackDecoder reads and decodes a drum track from
// a binary input stream.
//
// It's unexported because decoding a track should only
// be done when decoding a pattern; put another way,
// all tracks should be part of a pattern. Therefore,
// only the pattern decoder and methods relating to it
// are exported.
type trackDecoder struct {
	r   io.Reader
	off int
	err error
}

func newTrackDecoder(r io.Reader) *trackDecoder {
	return &trackDecoder{r: r, off: 0}
}

func (td *trackDecoder) decode(t *Track) error {
	ID, err := td.parseID()
	if err != nil {
		return fmt.Errorf("couldn't parse track id: %s", err)
	}
	t.ID = ID

	name, err := td.parseName()
	if err != nil {
		return fmt.Errorf("couldn't parse track name: %s", err)
	}
	t.Name = name

	steps, err := td.parseSteps()
	if err != nil {
		return fmt.Errorf("couldn't parse track steps: %s", err)
	}
	t.Steps = steps

	return nil
}

func (td *trackDecoder) parseID() (uint32, error) {
	var ID uint32
	if err := binary.Read(td.r, binary.LittleEndian, &ID); err != nil {
		return ID, err
	}

	td.off += trackIDSize

	return ID, nil
}

func (td *trackDecoder) parseName() (string, error) {
	// This byte tells us how many
	// bytes the track name occupies.
	buf := make([]byte, 1)
	if _, err := td.r.Read(buf); err != nil {
		return "", err
	}

	td.off++

	nameLen := int(buf[0])
	buf = make([]byte, nameLen)
	n, err := td.r.Read(buf)

	switch {
	case err != nil:
		return "", err
	case n != nameLen:
		return "", io.ErrUnexpectedEOF
	default:
		td.off += nameLen

		return string(buf), nil
	}
}

func (td *trackDecoder) parseSteps() ([trackSteps]bool, error) {
	var steps [trackSteps]bool

	for i := range steps {
		buf := make([]byte, 1)
		if _, err := td.r.Read(buf); err != nil {
			return steps, err
		}

		td.off++

		trigger := buf[0]

		if trigger == 0 {
			steps[i] = false
		} else {
			steps[i] = true
		}
	}

	return steps, nil
}

////////////////

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	p := new(Pattern)

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("couldn't open file: %s", err)
	}

	dec := NewDecoder(f)
	if err := dec.Decode(p); err != nil {
		return nil, fmt.Errorf("failed to decode pattern: %s", err)
	}

	return p, nil
}
