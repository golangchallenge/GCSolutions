package drum

// Author: Jaime Pillora
// Twitter: @jpillora
// Github: jpillora
// Date 2/2/2015

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
)

const (
	preamble = "SPLICE\x00\x00\x00\x00\x00\x00"
	numNotes = 4
	numBars  = 4
	numSteps = numNotes * numBars
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	p := &Pattern{}
	// Read file!
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Offset initially preamble length
	o := len(preamble)

	// Confirm its a valid splice file
	if string(b[:o]) != preamble {
		return nil, errors.New("Invalid splice file")
	}

	// length from this point on
	length := int(binary.BigEndian.Uint16(b[o : o+2]))
	o += 2

	// Length sanity check
	if o+length > len(b) {
		return nil, errors.New("Invalid file length")
	}

	// Version (ascii - trim from null, 32bytes)
	v := b[o : o+32]
	p.Version = string(v[:bytes.IndexByte(v, 0x00)])
	o += 32

	// Tempo (float32, 4bytes)
	err = binary.Read(bytes.NewReader(b[o:o+4]), binary.LittleEndian, &p.Tempo)
	if err != nil {
		fmt.Println("Tempo read failed:", err)
		return nil, err
	}
	o += 4

	// Tracks
	for o < length {
		t := &Track{}
		p.Tracks = append(p.Tracks, t)
		// ID
		t.ID = binary.LittleEndian.Uint32(b[o : o+4])
		o += 4
		// Name
		l := int(b[o])
		o++
		t.Name = string(b[o : o+l])
		o += l
		// Steps
		for i := 0; i < numSteps; i++ {
			t.Steps[i] = b[o+i] == 1
		}
		o += numSteps
	}

	return p, nil
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Version string
	Tempo   float32
	Tracks  []*Track
}

func (p *Pattern) String() string {
	// Convert []fmt.Stringer into []string
	tracks := make([]string, len(p.Tracks))
	for i, t := range p.Tracks {
		tracks[i] = t.String()
	}
	// Format tempo as float and drop trailing .0
	tempo := strings.TrimSuffix(fmt.Sprintf("%.1f", p.Tempo), ".0")
	// Pattern format
	return "Saved with HW Version: " + p.Version +
		"\nTempo: " + tempo + "\n" +
		strings.Join(tracks, "")
}

// Track is a high level representation of a
// drum track contained in a Pattern
type Track struct {
	ID    uint32
	Name  string
	Steps [numSteps]bool
}

func (t *Track) String() string {
	//String build steps (instead of concat)
	var steps bytes.Buffer
	for i, s := range t.Steps {
		if i%numBars == 0 {
			steps.WriteString("|")
		}
		if s {
			steps.WriteString("x")
		} else {
			steps.WriteString("-")
		}
	}
	steps.WriteString("|")
	// Track format
	return fmt.Sprintf("(%d) %s\t%s\n", t.ID, t.Name, steps.String())
}
