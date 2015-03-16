package drum

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"math"
	"strings"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Couldn't read file at %s", path)
	}

	// We will use this cursor to keep track of our progress
	i := 0

	// Read and check file type before decoding
	filetype := string(data[i : i+13])
	filetype = strings.TrimRight(filetype, string(0x00))
	if filetype != "SPLICE" {
		return nil, fmt.Errorf("Unexpected file type: Was '%s' expected 'SPLICE'", filetype)
	}
	i += 13

	// Obtain length of pattern data
	length := int(data[i])
	i += 1

	// Make sure length of pattern data corresponds to given pattern length
	if len(data)-i < length {
		return nil, fmt.Errorf("File shorter than expected: Was %d expected %d", len(data), length)
	}
	if len(data)-i > length {
		fmt.Printf("File longer than expected. Only %d of %d bytes will be considered.\n", length, len(data))
	}

	// The cursor is now at the start of the pattern
	// The first 32 bytes contain a string specifying the version
	version := string(data[i : i+32])
	version = strings.TrimRight(version, string(0x00))
	i += 32

	// The following 4 bytes contain the tempo of the pattern encoded as a float
	tempoBits := binary.LittleEndian.Uint32(data[i : i+4])
	tempo := math.Float32frombits(tempoBits)
	i += 4

	// The rest of the file contain the tracks of the pattern
	// Decode the tracks one by one until reaching given pattern length
	tracks := make([]Track, 0)
	for i < length {
		id := data[i]
		i += 4
		nameLength := int(data[i])
		i++
		name := string(data[i : i+nameLength])
		i += nameLength
		steps := make([]bool, 16)
		for j, b := range data[i : i+16] {
			switch b {
			case 0x00:
				steps[j] = false
			case 0x01:
				steps[j] = true
			default:
				return nil, fmt.Errorf("Invalid step byte %x", b)
			}
		}
		i += 16
		tracks = append(tracks, Track{id, name, steps})
	}

	p := &Pattern{version, tempo, tracks}
	return p, nil
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	version string
	tempo   float32
	tracks  []Track
}

// Each pattern can have one or more tracks
type Track struct {
	id    byte
	name  string
	steps []bool
}

// Builds a string representation of a pattern in the desired format
func (p Pattern) String() string {
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("Saved with HW Version: %s\n", p.version))
	buffer.WriteString(fmt.Sprintf("Tempo: %g\n", p.tempo))
	for _, t := range p.tracks {
		buffer.WriteString(fmt.Sprintf("(%d) %s\t", t.id, t.name))
		for i, step := range t.steps {
			if i%4 == 0 {
				buffer.WriteString("|")
			}
			if step {
				buffer.WriteString("x")
			} else {
				buffer.WriteString("-")
			}
		}
		buffer.WriteString(fmt.Sprintf("|\n"))
	}
	return buffer.String()
}
