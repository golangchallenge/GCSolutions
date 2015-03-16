package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"
	"strings"
)

// ErrNotSliceFile is returned when input is not a valid .splice file
var ErrNotSliceFile = errors.New("Input is not a valid .splice file")

// ErrBadFormat is returned when input is not in the expected format
var ErrBadFormat = errors.New("Input is not in the expected format")

// header is a low level representation of the header structure in a .splice file.
type header struct {
	SpliceDecl [6]byte
	_          [8]byte
	Version    [11]byte
	_          [21]byte
	Tempo      [4]byte
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	// Version information from the .splice file
	Version string
	// Tempo in bpm (beats per minute)
	Tempo float32
	// All the tracks in the given .splice file
	Tracks []Track
}

// Track is a high level representation of a single track of
// a drum pattern contained in a .splice file.
type Track struct {
	// ID of the track
	ID uint8
	// Name of the track
	Name string
	// All the steps of the given track - If a given step is true,
	// it means a sound should be triggered for that step.
	Steps [16]Step
}

// Step represents a single step in a track of a drum pattern contained
// in a .splice file.
type Step bool

func (s Step) String() string {
	if s {
		return "x"
	}
	return "-"
}

func (t Track) String() string {
	// should end up in the format:
	// |x---|----|xxxx|x-x-x|
	// where 'x' reprents a sound and '-' represents silence
	var s string
	for i, step := range t.Steps {
		// Start each 4 step measure with a '|'
		if i%4 == 0 {
			s += "|"
		}
		s += step.String()
	}
	// End the track with a '|'
	s += "|\n"
	s = fmt.Sprintf("(%d) %s\t%s", t.ID, t.Name, s)
	return s
}

func (p Pattern) String() string {
	// Get the tempo as a string, only displaying it as an integer if the
	// decimal part is 0.
	tempo := strconv.FormatFloat(float64(p.Tempo), 'f', -1, 32)

	// Build the output
	s := fmt.Sprintf("Saved with HW Version: %s\nTempo: %s\n", p.Version, tempo)

	for _, t := range p.Tracks {
		s += t.String()
	}

	return s
}

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	r := bytes.NewReader(b)
	h, err := decodeHeader(r)
	if err != nil {
		return nil, err
	}

	tempo := decodeTempo(h.Tempo)

	p := &Pattern{
		Version: strings.TrimRight(string(h.Version[:]), "\x00"),
		Tempo:   tempo,
	}
	r = bytes.NewReader(b[50:])
	p.Tracks = decodeTracks(r)

	return p, nil
}

// decodHeader decodes the header into a header struct
// or returns an error if it is not a valid .splice file.
func decodeHeader(r io.Reader) (*header, error) {
	h := &header{}
	binary.Read(r, binary.BigEndian, h)
	if string(h.SpliceDecl[:]) != "SPLICE" {
		return nil, ErrNotSliceFile
	}
	return h, nil
}

// decodeTempo decodes the tempo from a 4 byte sequence.
func decodeTempo(b [4]byte) float32 {
	var tempo float32
	r := bytes.NewReader(b[:])
	binary.Read(r, binary.LittleEndian, &tempo)
	return tempo
}

// decodTracks decodes tracks from a reader
// where the reader starts at the beginning of the first track.
func decodeTracks(r io.Reader) []Track {
	var tracks []Track
	for {
		t := decodeTrack(r)
		if t == nil {
			break
		}
		tracks = append(tracks, *t)
	}
	return tracks
}

// decodeTrack decodes a single track from a reader, where the reader begins at
// the beginning of that track returns nil if there is any error decoding the track
func decodeTrack(r io.Reader) *Track {
	t := &Track{}
	// ID is the first byte of a track
	var id uint8
	err := binary.Read(r, binary.BigEndian, &id)
	if err != nil {
		return nil
	}
	t.ID = id

	// The next 4 bytes tell us the length of the name of the track
	var nameLength int32
	err = binary.Read(r, binary.BigEndian, &nameLength)
	if err != nil {
		return nil
	}

	// Decode the name
	var name = make([]byte, nameLength)
	err = binary.Read(r, binary.BigEndian, &name)
	if err != nil {
		return nil
	}
	t.Name = string(name)

	// Decode the steps
	var steps [16]byte
	err = binary.Read(r, binary.BigEndian, &steps)
	if err != nil {
		return nil
	}
	for i := 0; i < 16; i++ {
		t.Steps[i] = steps[i] == 1
	}
	return t
}
