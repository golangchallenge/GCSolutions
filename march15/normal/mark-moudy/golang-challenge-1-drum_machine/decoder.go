//Golang-Challenge-1 Submission by Mark Moudy
package drum

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	p := &Pattern{}
	//decode file into Pattern struct
	file, err := ioutil.ReadFile(path)
	if err != nil {
		panic("error reading file")
	}

	var lengthOfPattern uint8
	buf := bytes.NewBuffer(file)

	// read length of pattern
	buf.Next(13)
	binary.Read(buf, binary.LittleEndian, &lengthOfPattern)
	EOP := buf.Len() - int(lengthOfPattern)

	// read version
	p.Version = buf.Next(12)
	//read tempo
	buf.Next(20)
	binary.Read(buf, binary.LittleEndian, &p.Tempo)

	//read tracks ---
	for buf.Len() > EOP {
		track := &track{}
		//read id
		binary.Read(buf, binary.LittleEndian, &track.ID)
		//length of name
		buf.Next(3)
		var nameLength uint8
		binary.Read(buf, binary.LittleEndian, &nameLength)
		track.Name = buf.Next(int(nameLength))
		track.Steps = buf.Next(16)
		p.Tracks = append(p.Tracks, track)
	}

	return p, nil
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Version []uint8
	Tempo   float32
	Tracks  []*track
}

func (p *Pattern) String() string {
	var tracks string
	tmp := bytes.NewBuffer(p.Version)
	delim := uint8(0)
	version, _ := tmp.ReadBytes(delim)
	version = version[:len(version)-1]
	for _, v := range p.Tracks {
		tracks += fmt.Sprintf("%s", v)
	}
	return fmt.Sprintf("Saved with HW Version: %s\nTempo: %v\n%s", version, p.Tempo, tracks)
}

// track holds the details about each track and
// will make it easier to manipulate an individual track in the future
type track struct {
	ID    uint8
	Name  []uint8
	Steps []uint8
}

func (t *track) String() string {
	var cSteps string
	for _, v := range t.Steps {
		switch v {
		case 0:
			cSteps += "-"
		case 1:
			cSteps += "x"
		}
	}
	steps := fmt.Sprintf("|%s|%s|%s|%s|", cSteps[0:4], cSteps[4:8], cSteps[8:12], cSteps[12:])
	return fmt.Sprintf("(%d) %s\t%s\n", t.ID, t.Name, steps)
}
