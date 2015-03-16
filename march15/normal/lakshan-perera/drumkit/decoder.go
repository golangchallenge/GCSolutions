package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"strings"
)

const magicStr = "SPLICE"

type steps []byte

type track struct {
	ID    uint32
	Name  []byte
	Steps steps
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Version string
	Tempo   float32
	Tracks  []*track
}

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	p := &Pattern{}

	f, err := os.Open(path)
	if err != nil {
		return p, err
	}
	defer f.Close()

	// read the first 6 bytes decode splice magic string
	magic := make([]byte, 6)
	if _, err := f.Read(magic[:]); err != nil {
		return nil, err
	}
	if string(magic) != magicStr {
		return nil, errors.New("invalid file format")
	}

	// valid splice file, continue to read

	// seek next 8 bytes for version header
	if _, err := f.Seek(8, 1); err != nil {
		return nil, err
	}

	// read the version
	version := make([]byte, 24)
	if _, err := f.Read(version[:]); err != nil {
		return nil, err
	}
	// return the version string removing trailing whitespace
	p.Version = string(bytes.Trim(version, "\x00"))

	// seek next 8 bytes for tempo header
	if _, err := f.Seek(8, 1); err != nil {
		return nil, err
	}

	// read the tempo
	var tempo uint32
	if err := binary.Read(f, binary.LittleEndian, &tempo); err != nil {
		return nil, err
	}
	p.Tempo = math.Float32frombits(tempo)

	p.Tracks, err = readTracks(f)
	// we fail gracefully on EOF errors (eg. legacy versions).
	// ie. return the tracks read until EOF reached.
	if err != nil && err != io.EOF {
		return nil, err
	}

	return p, nil
}

func readTracks(r io.Reader) ([]*track, error) {
	var ts []*track

	// read tracks
	for {
		t := &track{}

		var id uint32
		if err := binary.Read(r, binary.LittleEndian, &id); err != nil {
			if err == io.EOF { // intentional EOF
				return ts, nil
			}
			return ts, err
		}
		t.ID = id

		var nameLen uint8
		if err := binary.Read(r, binary.LittleEndian, &nameLen); err != nil {
			return ts, err
		}

		name := make([]byte, nameLen)
		if _, err := r.Read(name[:]); err != nil {
			return ts, err
		}
		t.Name = name

		steps := make(steps, 16)
		if _, err := r.Read(steps[:]); err != nil {
			return ts, err
		}
		t.Steps = steps

		ts = append(ts, t)
	}
	return ts, nil
}

func (s steps) String() string {
	output := "|"

	for i, b := range []byte(s) {
		if b == 1 {
			output = output + "x"
		} else {
			output = output + "-"
		}
		if (i+1)%4 == 0 {
			output = output + "|"
		}
	}
	return output
}

func (p *Pattern) String() string {
	output := []string{}
	output = append(output, fmt.Sprintf("Saved with HW Version: %s", p.Version))
	output = append(output, fmt.Sprintf("Tempo: %v", p.Tempo))

	for _, t := range p.Tracks {
		output = append(output, fmt.Sprintf("(%v) %s\t%s", t.ID, t.Name, t.Steps))
	}
	return strings.Join(output, "\n") + "\n"
}
