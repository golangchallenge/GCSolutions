package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"os"
)

const (
	SPLICE = "SPLICE"
)

var (
	ErrSmallBinary   = errors.New("the binary is to small to be drum machine pattern")
	ErrInValidHeader = errors.New("a drum machine pattern binary must be start by a \"SPLICE\" string")
	ErrInValidStep   = errors.New("invalid step value")
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(pathname string) (*Pattern, error) {
	in, err := os.Open(pathname)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := in.Close(); err != nil {
			panic(err)
		}
	}()

	p := new(Pattern)
	if err := NewDecoder(in).Decode(p); err != nil {
		return nil, err
	}

	return p, nil
}

type Decoder struct {
	r io.Reader

	// body's length
	len int
	err error
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r}
}

func (dec *Decoder) Decode(p *Pattern) error {
	p.Version, p.Tempo = dec.parseHeader()
	for dec.err == nil && dec.len > 0 {
		p.Tracks = append(p.Tracks, dec.parseTrack())
	}

	if dec.err != nil && dec.len > 1 {
		dec.err = io.ErrUnexpectedEOF
	}

	return dec.err
}

func (dec *Decoder) parseHeader() (version string, tempo float32) {
	if dec.err != nil {
		return
	}

	buffer := make([]byte, 50)
	n, err := dec.r.Read(buffer)

	if err != nil {
		dec.err = err
		return
	}

	if n < 50 {
		dec.err = ErrSmallBinary
		return
	}

	// Validate the binary header
	for i := 0; i < len(SPLICE); i++ {
		if SPLICE[i] != buffer[i] {
			dec.err = ErrInValidHeader
			return
		}
	}

	dec.len = int(binary.BigEndian.Uint32(buffer[10:14]))
	version = string(bytes.Trim(buffer[14:46], "\x00"))
	tempo = math.Float32frombits(binary.LittleEndian.Uint32(buffer[46:50]))
	dec.len = dec.len - 36
	return
}

func (dec *Decoder) parseTrack() *Track {
	if dec.err != nil {
		return nil
	}

	track := new(Track)
	buffer := make([]byte, 5)

	if _, err := dec.r.Read(buffer); err != nil {
		dec.err = err
		return nil
	}

	track.Id = binary.LittleEndian.Uint16(buffer[:2])
	tlname := buffer[4]
	dec.len -= len(buffer)

	buffer = make([]byte, tlname)
	if _, err := dec.r.Read(buffer); err != nil {
		dec.err = err
		return nil
	}
	track.Name = string(buffer)
	dec.len -= len(buffer)

	buffer = make([]byte, 16)
	if _, err := dec.r.Read(buffer); err != nil {
		dec.err = err
		return nil
	}

	steps := Steps(make([]byte, 16))
	for i := 0; i < 16; i++ {
		step := buffer[i]
		if step != 0 && step != 1 {
			dec.err = ErrInValidStep
			return nil
		}
		steps[i] = step
	}

	track.Steps = steps
	dec.len -= len(buffer)
	return track
}
