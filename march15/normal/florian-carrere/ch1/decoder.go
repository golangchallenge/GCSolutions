package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
)

var (
	// SplicePrefix is the header prefix for a .splice file
	SplicePrefix = "SPLICE"

	// ErrWrongFormat is the error returned by DecodeFile
	//  if the provided file is not well-formed
	ErrWrongFormat = errors.New("Invalid file format!")
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	fileStat, err := file.Stat()
	if err != nil {
		return nil, err
	}
	data := make([]byte, fileStat.Size())
	err = binary.Read(file, binary.LittleEndian, &data)
	if err != nil {
		return nil, err
	}
	if !bytes.HasPrefix(data, []byte(SplicePrefix)) {
		return nil, ErrWrongFormat
	}

	// Parse headers
	headers := data[0:50]
	version, bpm := ParseHeaders(headers)

	// Parse tracks
	tracks := ParseTracks(data[50:])

	p := &Pattern{
		Version: version,
		Tempo:   bpm,
		Tracks:  tracks,
	}

	return p, nil
}

// ParseHeaders decodes version and tempo from headers of the binary file format
func ParseHeaders(headers []byte) (string, float32) {
	headers = bytes.TrimPrefix(headers, []byte(SplicePrefix))[8:]
	// Extract version
	version := string(bytes.Trim(headers[:32], "\x00"))
	// Extract BPM
	bpmFromBits := binary.LittleEndian.Uint32(headers[32:])
	bpm := math.Float32frombits(bpmFromBits)

	return version, bpm
}

// ParseTracks decodes tracks
func ParseTracks(tracksBits []byte) []Track {
	if bytes.Contains(tracksBits, []byte(SplicePrefix)) {
		tracksBits = tracksBits[0:bytes.Index(tracksBits, []byte(SplicePrefix))]
	}
	tracks := []Track{}
	if len(tracksBits) == 0 {
		return tracks
	}

	for len(tracksBits) > 0 {
		track := Track{
			ID:    int(tracksBits[0]),
			Name:  "",
			Steps: []byte{},
		}
		namelength := int(tracksBits[4])
		track.Name = string(tracksBits[5 : 4+namelength+1])
		track.Steps = tracksBits[4+namelength+1 : 4+namelength+1+16]
		tracks = append(tracks, track)
		tracksBits = tracksBits[4+namelength+1+16:]
	}

	return tracks
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Version string
	Tempo   float32
	Tracks  []Track
}

// Pattern implements Stringer interface
func (p Pattern) String() string {
	s := "Saved with HW Version: " + p.Version + "\nTempo: " + fmt.Sprintf("%g", p.Tempo) + "\n"
	for _, track := range p.Tracks {
		s += track.String()
	}
	return s
}

// Track representation included in a drum pattern
type Track struct {
	ID    int
	Name  string
	Steps []byte
}

// Track implements Stringer interface
func (tr Track) String() string {
	s := "(" + strconv.Itoa(tr.ID) + ") " + tr.Name + "\t"
	for k, b := range tr.Steps {
		if k%4 == 0 {
			s += "|"
		}
		if b == 1 {
			s += "x"
		} else {
			s += "-"
		}
	}
	s += "|\n"
	return s
}
