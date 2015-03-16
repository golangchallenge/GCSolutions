package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io/ioutil"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	p := new(Pattern)
	r, err := ioutil.ReadFile(path)
	if err != nil {
		return p, err
	}
	// The first 6 bytes of the file should be SPLICE
	if string(r[:6]) != "SPLICE" {
		return p, errors.New("File specified isn't a .splice file.")
	}
	fileSize := binary.BigEndian.Uint64(r[6:14])
	buf := bytes.NewReader(r[46:54])
	err = binary.Read(buf, binary.LittleEndian, &p.Tempo)
	if err != nil {
		return p, err
	}
	for i := 14; i < 45; i++ {
		if r[i] > 0 {
			p.Version += string(r[i])
		}
	}
	index := 50
	for uint64(index) < fileSize {
		var t Track
		t.ID = r[index]
		index++
		// The number of charaters in the name of the track is stored as an unsigned int 32.
		// It is converted to an int since it is only used as an int and to limit the number
		// of conversions
		nameLen := int(binary.BigEndian.Uint32(r[index : index+4]))
		index += 4
		for i := 0; i < nameLen; i++ {
			t.Name += string(r[index+i])
		}
		index += nameLen
		for i := 0; i < 16; i++ {
			if r[index+i] == 1 {
				t.Play[i] = true
			}
		}
		index += 16
		p.Tracks = append(p.Tracks, t)
	}
	return p, nil
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Version string
	Tempo   float32
	Tracks  []Track
}

// Track is the drum sound.
type Track struct {
	ID   byte
	Name string
	Play [16]bool
}
