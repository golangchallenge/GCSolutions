package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	p := &Pattern{}

	splice := make([]byte, 6)
	err = binary.Read(file, binary.LittleEndian, &splice)
	if err != nil {
		return nil, err
	}
	if string(splice) != "SPLICE" {
		return nil, errors.New("Unrecognized file format")
	}

	var length int64
	err = binary.Read(file, binary.BigEndian, &length)
	if err != nil {
		return nil, err
	}

	contents := make([]byte, length)
	_, err = file.Read(contents)
	if err != nil {
		return nil, err
	}

	buf := bytes.NewReader(contents)

	version := make([]byte, 32)
	err = binary.Read(buf, binary.LittleEndian, &version)
	if err != nil {
		return nil, err
	}
	p.version = string(version[:bytes.Index(version, []byte{0})])

	err = binary.Read(buf, binary.LittleEndian, &p.tempo)
	if err != nil {
		return nil, err
	}

	for buf.Len() > 0 {
		var newTrack track
		err = binary.Read(buf, binary.LittleEndian, &newTrack.id)
		if err != nil {
			return nil, err
		}

		var nameSize byte
		err = binary.Read(buf, binary.LittleEndian, &nameSize)
		if err != nil {
			return nil, err
		}
		name := make([]byte, nameSize)
		err = binary.Read(buf, binary.LittleEndian, &name)
		if err != nil {
			return nil, err
		}
		newTrack.name = string(name)

		newTrack.beats = make([]byte, 16)
		err = binary.Read(buf, binary.LittleEndian, &newTrack.beats)
		if err != nil {
			return nil, err
		}

		p.tracks = append(p.tracks, newTrack)
	}

	return p, nil
}

type track struct {
	id    int32
	name  string
	beats []byte
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	version string
	tempo   float32
	tracks  []track
}

func (p *Pattern) String() string {
	var buffer bytes.Buffer
	buffer.WriteString("Saved with HW Version: ")
	buffer.WriteString(p.version)
	buffer.WriteString("\n")
	fmt.Fprintf(&buffer, "Tempo: %v\n", p.tempo)
	for _, track := range p.tracks {
		fmt.Fprintf(&buffer, "(%d) %s\t", track.id, track.name)
		for i, beat := range track.beats {
			if i%4 == 0 {
				buffer.WriteString("|")
			}
			if beat == 1 {
				buffer.WriteString("x")
			} else {
				buffer.WriteString("-")
			}
		}
		buffer.WriteString("|\n")
	}
	return buffer.String()
}
