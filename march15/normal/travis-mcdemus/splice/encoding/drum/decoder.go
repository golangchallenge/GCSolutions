package drum

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	p := NewPattern()
	f, err := os.Open(path)
	if err != nil {
		return p, err
	}
	defer f.Close()
	d := NewDecoder(f)
	err = d.Decode(p)
	return p, err
}

// A Decoder represents a drum pattern parser.
// The parser assumes that input follows an undocumented specification.
type Decoder struct {
	// TODO(aoeu): Provide a specfication and better documentation.
	r io.Reader
}

// NewDecoder creates a new drum pattern decoder reading from r.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r}
}

// Decode reads from the decoder's input stream to initialize a drum pattern.
func (d *Decoder) Decode(p *Pattern) error {
	h := header{}
	if err := binary.Read(d.r, binary.LittleEndian, &h); err != nil {
		return err
	}
	p.HardwareVersion = string(bytes.Trim(h.HardwareVersion[:], "\x00"))
	switch p.HardwareVersion {
	case "0.808-alpha":
		p.Tempo = int(h.Tempo) / 2
		if h.TempoDecimal != 0 {
			// TODO(aoeu): Is this really the correct way to determine the decimal?
			p.TempoDecimal = int(h.TempoDecimal) - 200
		}
	case "0.909":
		// TODO(aoeu): Is there no byte this value can be derived from?
		p.Tempo = 240
	case "0.708-alpha":
		// TODO(aoeu): Is there no byte this value can be derived from?
		p.Tempo = 999
	}
	var err error
	p.Tracks, err = readAllTracks(d.r)
	if err == io.ErrUnexpectedEOF {
		return nil
	}
	return err
}

func readAllTracks(r io.Reader) (Tracks, error) {
	var t Tracks
	for {
		track, err := readTrack(r)
		if err != nil {
			if err == io.EOF {
				return t, nil
			}
			return t, err
		}
		t = append(t, track)
	}
}

// Tracks is a drum Track series that comprises the pattern.
type Tracks []Track

func (t Tracks) String() string {
	var s string
	for _, drumPart := range t {
		s += fmt.Sprintf("%v\n", drumPart)
	}
	return s
}

func readTrack(r io.Reader) (Track, error) {
	t := *NewTrack()
	if err := binary.Read(r, binary.LittleEndian, &t.ID); err != nil {
		return t, err
	}
	padding := make([]byte, 3)
	if err := binary.Read(r, binary.LittleEndian, &padding); err != nil {
		return t, err
	}
	var nameLen byte
	if err := binary.Read(r, binary.LittleEndian, &nameLen); err != nil {
		return t, err
	}
	nameBytes := make([]byte, nameLen)
	if err := binary.Read(r, binary.LittleEndian, &nameBytes); err != nil {
		return t, err
	}
	t.Name = string(nameBytes)
	if err := binary.Read(r, binary.LittleEndian, &t.Sequence); err != nil {
		return t, err
	}
	return t, nil
}
