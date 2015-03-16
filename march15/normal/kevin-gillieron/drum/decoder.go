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

// Decode reads a drum machine file from r and returns a pointer to a parsed
// pattern which is the entry point to the rest of the data.
func Decode(r io.Reader) (*Pattern, error) {
	dec := newDecoder(r)
	return dec.decode()
}

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

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Version string
	Tempo   float32
	Tracks  []Track
}

func (p Pattern) String() string {
	buf := &bytes.Buffer{}
	buf.WriteString("Saved with HW Version: " + p.Version + "\n")
	buf.WriteString(fmt.Sprintf("Tempo: %g\n", p.Tempo))

	for _, t := range p.Tracks {
		buf.WriteString(t.String() + "\n")
	}

	return buf.String()
}

// A Track represents a sound.
type Track struct {
	Header TrackHeader
	Name   string
	Steps  [16]bool
}

// A TrackHeader represents the header of a Track.
type TrackHeader struct {
	ID      uint32 // ID of the track.
	NameLen uint8  // Length (in bytes) of the track name.
}

func (t Track) String() string {
	buf := bytes.NewBufferString(fmt.Sprintf("(%d) %s\t", t.Header.ID, t.Name))

	for i, s := range t.Steps {
		if i%4 == 0 {
			buf.WriteString("|")
		}

		if s {
			buf.WriteString("x")
		} else {
			buf.WriteString("-")
		}
	}

	buf.WriteString("|")

	return buf.String()
}

var (
	headerLen        = 14 // size (in bytes) of the header
	versionLen       = 32 // size (in bytes) of the version number
	stepsLen   uint8 = 16 // size (in bytes) required for the 16 steps
)

// A decoder reads and decodes a drum machine file.
type decoder struct {
	r   io.Reader
	pat *Pattern
}

// newDecoder returns a new decoder that reads from r.
func newDecoder(r io.Reader) *decoder {
	return &decoder{r: r, pat: &Pattern{}}
}

func (dec *decoder) decode() (*Pattern, error) {
	if err := dec.checkHeader(); err != nil {
		return nil, err
	}

	if err := dec.parseVersion(); err != nil {
		return nil, err
	}

	if err := dec.parseTempo(); err != nil {
		return nil, err
	}

	for {
		if err := dec.parseTrack(); err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
	}

	return dec.pat, nil
}

func (dec *decoder) checkHeader() error {
	buf := make([]byte, headerLen, headerLen)
	if _, err := dec.r.Read(buf); err != nil {
		return err
	}

	if string(buf[:6]) != "SPLICE" {
		return errors.New("not a splice file")
	}

	return nil
}

func (dec *decoder) parseVersion() error {
	buf := make([]byte, versionLen, versionLen)
	if _, err := dec.r.Read(buf); err != nil {
		return err
	}

	dec.pat.Version = strings.TrimRight(string(buf), "\x00")

	return nil
}

func (dec *decoder) parseTempo() error {
	return binary.Read(dec.r, binary.LittleEndian, &dec.pat.Tempo)
}

func (dec *decoder) parseTrack() error {
	t := Track{}
	if err := binary.Read(dec.r, binary.LittleEndian, &t.Header); err != nil {
		return err
	}

	buf := make([]byte, t.Header.NameLen+stepsLen)
	if n, err := dec.r.Read(buf); err != nil {
		return err
	} else if n != int(t.Header.NameLen+stepsLen) {
		return io.EOF
	}

	t.Name = string(buf[:t.Header.NameLen])

	for i, b := range buf[t.Header.NameLen:] {
		t.Steps[i] = b == 1
	}

	dec.pat.Tracks = append(dec.pat.Tracks, t)

	return nil
}
