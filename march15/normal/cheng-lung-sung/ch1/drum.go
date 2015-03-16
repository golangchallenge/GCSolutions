// Package drum is supposed to implement the decoding of .splice drum machine files.
// See golang-challenge.com/go-challenge1/ for more information
package drum

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"
)

const cacheSize = 8

// SPLICE header
var header = []byte("SPLICE")

type decoder struct {
	tmp        [cacheSize]byte
	dataLength uint32
	r          io.Reader
	p          *Pattern
}

// A FormatError reports that the input is not a valid SPLICE data.
type FormatError string

func (e FormatError) Error() string { return "drum: invalid format: " + string(e) }

// checkHeader will examine if the very first bytes matches
func (d *decoder) checkHeader() error {
	_, err := io.ReadFull(d.r, d.tmp[:len(header)])
	if err != nil {
		return err
	}
	if !bytes.Equal(d.tmp[:len(header)], header) {
		return FormatError("not a SPLICE file")
	}
	return nil
}

// parseChunk read the whole drum data by the length
func (d *decoder) parseChunk() error {
	_, err := io.ReadFull(d.r, d.tmp[:cacheSize])
	if err != nil {
		return err
	}
	d.dataLength = binary.BigEndian.Uint32(d.tmp[4:8])
	return d.parseDrum()
}

// parseVersion returns the Hardware version
func (d *decoder) parseVersion() error {
	bi := 0
	for j := 0; j < 4; j++ {
		n, err := io.ReadFull(d.r, d.tmp[:cacheSize])
		if err != nil {
			return err
		}
		d.dataLength -= uint32(n)
		for i := 0; i < len(d.tmp); i++ {
			if d.tmp[i] == uint8(0) {
				break
			}
			bi += copy(d.p.Version[bi:], d.tmp[i:i+1])
		}
	}
	return nil
}

// parseTempo returns the drum tempo
func (d *decoder) parseTempo() error {
	n, err := io.ReadFull(d.r, d.tmp[:4])
	if err != nil {
		return err
	}
	d.p.Tempo = math.Float32frombits(binary.LittleEndian.Uint32(d.tmp[:4]))
	d.dataLength -= uint32(n)
	return nil
}

// parseTrackName returns the track id
func (d *decoder) parseID(t *Track) error {
	n, err := io.ReadFull(d.r, d.tmp[:4])
	if err != nil {
		return err
	}
	t.ID = binary.LittleEndian.Uint32(d.tmp[:4])
	d.dataLength -= uint32(n)
	return nil
}

// parseTractName returns the track name
func (d *decoder) parseTrackName(t *Track) error {
	n, err := io.ReadFull(d.r, d.tmp[:1])
	if err != nil {
		return err
	}
	d.dataLength -= uint32(n)
	if d.tmp[0] > 255 {
		return FormatError("invalid track format")
	}
	nameLen := uint32(d.tmp[0])
	t.Name = make([]byte, nameLen)
	bi := 0
	for nameLen > 0 {
		n, err = io.ReadFull(d.r, d.tmp[:min(int(nameLen), cacheSize)])
		if err != nil {
			return err
		}
		d.dataLength -= uint32(n)
		nameLen -= uint32(n)
		bi += copy(t.Name[bi:], d.tmp[:n])
	}
	return nil
}

// parseStep use two uint8 to store the 16 steps
func (d *decoder) parseStep(t *Track) error {
	for j := 0; j < 2; j++ {
		n, err := io.ReadFull(d.r, d.tmp[:8])
		if err != nil {
			return err
		}
		d.dataLength -= uint32(n)
		for i := uint(0); i < 8; i++ {
			t.Steps[j] |= (d.tmp[i] << i)
		}
	}
	return nil
}

// parseDrum will extract the drum pattern from the binary format
func (d *decoder) parseDrum() (err error) {
	d.p = &Pattern{}
	err = d.parseVersion()
	if err != nil {
		return err
	}
	err = d.parseTempo()
	if err != nil {
		return err
	}
	for int(d.dataLength) > 0 {
		var t Track
		d.parseID(&t)
		if err != nil {
			return err
		}
		d.parseTrackName(&t)
		if err != nil {
			return err
		}
		d.parseStep(&t)
		if err != nil {
			return err
		}
		d.p.Tracks = append(d.p.Tracks, t)
	}
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Decode extract the pattern from the io.Reader
func Decode(r io.Reader) (*Pattern, error) {
	d := &decoder{
		r: r,
	}
	if err := d.checkHeader(); err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return nil, err
	}
	if err := d.parseChunk(); err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return nil, err
	}

	return d.p, nil
}
