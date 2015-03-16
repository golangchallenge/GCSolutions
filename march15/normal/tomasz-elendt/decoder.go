package drum

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"runtime"
)

// ErrInvalidHeader is returned by Decode when input stream contains
// invalid header.
var ErrInvalidHeader = errors.New("invalid header")

// A Decoder reads and decodes patterns from an input stream.
type Decoder struct {
	r     io.Reader
	order binary.ByteOrder
}

// NewDecoder returns a new decoder that reads from r.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		r:     r,
		order: binary.LittleEndian,
	}
}

func (dec *Decoder) read(data interface{}) {
	err := binary.Read(dec.r, dec.order, data)
	if err != nil {
		panic(err)
	}
}

func (dec *Decoder) readString() string {
	var len uint8
	dec.read(&len)
	b := make([]byte, len)
	dec.read(&b)
	return string(b)
}

type header struct {
	_     [13]byte // ignore first 13 bytes
	Len   uint8
	Ver   [32]byte // null terminated string
	Tempo float32
}

func (dec *Decoder) decodeHeader() (int, string, float32) {
	var h header
	dec.read(&h)
	// number of bytes left for tracks
	tSize := int(h.Len) - binary.Size(h.Ver) - binary.Size(h.Tempo)
	var ver string
	if i := bytes.IndexByte(h.Ver[:], '\x00'); i != -1 {
		ver = string(h.Ver[:i])
	} else {
		panic(ErrInvalidHeader)
	}
	return tSize, ver, h.Tempo
}

func (dec *Decoder) compressSteps(steps *[16]byte) uint16 {
	c := uint16(0)
	for i, step := range steps {
		if step > 0 {
			c |= uint16(1) << uint16(i)
		}
	}
	return c
}

func (dec *Decoder) decodeTrack() *Track {
	var id uint32
	dec.read(&id)
	name := dec.readString()
	var steps [16]byte
	dec.read(&steps)
	return &Track{
		ID:    id,
		Name:  name,
		Steps: dec.compressSteps(&steps)}
}

// Decode reads binary encoded pattern value from its
// input and returns it.
func (dec *Decoder) Decode() (p *Pattern, err error) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			err = r.(error)
			if err == io.EOF {
				err = io.ErrUnexpectedEOF
			}
		}
	}()

	tSize, ver, tempo := dec.decodeHeader()
	var tracks []*Track
	for tSize > 0 {
		track := dec.decodeTrack()
		tracks = append(tracks, track)
		tSize -= binary.Size(track.ID) + 16 + 1 + len(track.Name)
	}
	// TODO: fail if bLen == 0 or no EOF?
	return &Pattern{
		Version: ver,
		Tempo:   tempo,
		Tracks:  tracks,
	}, nil
}

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	dec := NewDecoder(bufio.NewReader(file))
	return dec.Decode()
}
