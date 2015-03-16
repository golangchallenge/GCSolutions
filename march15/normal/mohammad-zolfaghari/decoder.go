package drum

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	p := &Pattern{}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// header
	if string(data[:6]) != "SPLICE" {
		return nil, errors.New("Invalid binary file")
	}
	plen := int(data[13]) + 14

	// version
	i := 14
	for int(data[i]) != 0 {
		i++
	}
	p.version = string(data[14:i])
	log.Print(p.version)

	// tempo
	p.tempo = math.Float32frombits(binary.LittleEndian.Uint32(data[46:50]))
	log.Print(p.tempo)
	i = 50

	// tracks
	ts := make([]Track, 0)
	for i < plen {
		tr := Track{id: int(data[i])}
		log.Printf("id: %d", int(data[i]))
		i++

		for data[i] == 0 {
			i++
		}
		tlen := int(data[i])
		i++

		tr.name = string(data[i : i+tlen])
		log.Printf("name: %s", tr.name)
		i += tlen

		for j := 0; j < 16; j++ {
			if int(data[i]) == 1 {
				tr.steps[j] = true
			}
			i++
		}

		log.Printf("%#v", tr)
		ts = append(ts, tr)
	}
	p.tracks = ts

	return p, nil
}

type (
	// Pattern is the high level representation of the
	// drum pattern contained in a .splice file.
	Pattern struct {
		version string
		tempo   float32
		tracks  []Track
	}

	Track struct {
		id    int
		name  string
		steps [16]bool
	}
)

func (p *Pattern) String() string {
	res := fmt.Sprintf("Saved with HW Version: %s\nTempo: %g\n",
		p.version, p.tempo)
	for _, track := range p.tracks {
		res += fmt.Sprintf("%s\n", track.String())
	}
	return res
}

func (t *Track) String() string {
	res := fmt.Sprintf("(%d) %s\t", t.id, t.name)
	for i, step := range t.steps {
		if i%4 == 0 {
			res += "|"
		}
		if step {
			res += "x"
		} else {
			res += "-"
		}
		if i == 15 {
			res += "|"
		}
	}
	return res
}
