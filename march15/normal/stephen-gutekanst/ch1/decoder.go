package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"os"
)

// ErrExpectedSpliceMagic is returned when file data is corrupt (i.e. if the
// magic SPLICE prefix is missing).
var ErrExpectedSpliceMagic = errors.New("drum: expected SPLICE magic file prefix")

// decoder decodes splice files, and is stream-friendly. It reads data
// progressively from the underlying reader.
type decoder struct {
	r   io.Reader
	buf []byte

	dataLength uint64
	hwVersion  string
	tempo      float32
}

// smallRead performs a small read for < 128 bytes. It reads the bytes into
// d.buf and returns any errors that occur.
//
// If adv is true, d.dataLength is advanced by n.
func (d *decoder) smallRead(n int, adv bool) error {
	// Expand/slice buffer as needed.
	if cap(d.buf) < n {
		d.buf = make([]byte, n)
	} else {
		d.buf = d.buf[:n]
	}

	// Fill the buffer
	_, err := io.ReadFull(d.r, d.buf)

	// Advance if desired.
	if adv {
		d.dataLength -= uint64(n)
	}
	return err
}

// decodeHeader decodes the header of the splice file:
//
//  | Bytes | Description                    |
//  |-------|--------------------------------|
//  | 0-5   | "SPLICE" magic prefix.         |
//  | 5-13  | Big-endian uint64 data length. |
//  | 13-45 | 32-byte HW version string.     |
//  | 45-53 | Little-endian float32 tempo.   |
//
func (d *decoder) decodeHeader() error {
	// Decode the "SPLICE" magic prefix.
	if err := d.smallRead(6, false); err != nil {
		return err
	}
	if string(d.buf) != "SPLICE" {
		return ErrExpectedSpliceMagic
	}

	// Grab the data length.
	if err := d.smallRead(8, false); err != nil {
		return err
	}
	d.dataLength = binary.BigEndian.Uint64(d.buf)

	// Decode the 32-byte HW version string and trim the NULL termination.
	if err := d.smallRead(32, true); err != nil {
		return err
	}
	d.hwVersion = string(bytes.TrimRight(d.buf, "\x00"))

	// Grab the tempo.
	if err := d.smallRead(4, true); err != nil {
		return err
	}
	d.tempo = math.Float32frombits(binary.LittleEndian.Uint32(d.buf))
	return nil
}

// next decodes the next track and returns it or nil and a error. It returns
// err=io.EOF in the event that all tracks have been read.
//
//  | Bytes    | Description                              |
//  |----------|------------------------------------------|
//  | 0-1      | Track ID (unrelated to decoding).        |
//  | 1-5      | N: Big-endian uint32 name string length. |
//  | 5-N      | Name string.                             |
//  | N-(N+16) | Steps 0-16 (zero is off, one is on).     |
//
func (d *decoder) next() (Track, error) {
	var (
		err error
		t   = Track{}
	)

	// Handle the case where we've decoded all tracks described by the header.
	if d.dataLength <= 0 {
		return t, io.EOF
	}

	// Decode the track's ID, and name string length.
	if err = d.smallRead(5, true); err != nil {
		return t, err
	}
	t.ID = int(d.buf[0])
	nameLength := int(binary.BigEndian.Uint32(d.buf[1:]))

	// Decode track name string.
	if err = d.smallRead(nameLength, true); err != nil {
		return t, err
	}
	t.Name = string(d.buf)

	// Decode track steps (0 for on, 1 for off), 16 bytes total.
	if err = d.smallRead(16, true); err != nil {
		return t, err
	}
	for i, s := range d.buf {
		t.Steps[i] = s == 1
	}
	return t, nil
}

// newDecoder returns a new decoder for the given reader. It decodes the file
// header immedietely and returns all errors, if any.
func newDecoder(r io.Reader) (*decoder, error) {
	d := &decoder{
		r:   r,
		buf: make([]byte, 32),
	}
	return d, d.decodeHeader()
}

// decode is just like DecodeFile but it operates on an io.Reader instead of a
// filepath.
func decode(f io.Reader) (*Pattern, error) {
	// Create a new decoder for the file.
	dec, err := newDecoder(f)
	if err != nil {
		return nil, err
	}

	// Initialize the pattern.
	p := &Pattern{
		HWVersion: dec.hwVersion,
		Tempo:     dec.tempo,
	}

	// Decode all the tracks, progressively.
	for {
		// Decode the next track.
		t, err := dec.next()
		if err == io.EOF {
			// Done decoding tracks.
			break
		}
		if err != nil {
			// Something went wrong.
			return nil, err
		}

		// Add the track to the pattern.
		p.Tracks = append(p.Tracks, t)
	}
	return p, nil
}

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
//
// If any error occurs, it is returned along with a nil *Pattern.
func DecodeFile(path string) (*Pattern, error) {
	// Open the file.
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return decode(f)
}
