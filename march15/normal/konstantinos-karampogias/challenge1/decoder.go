package drum

// Konstantinos Karampogias
// karampok@gmail.com

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	tb, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	b := bytes.Split(tb, tb[:6])[1]
	p := &Pattern{}
	p.init(b)
	return p, nil
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	HWversion string
	Tempo     float32
	Tracks    []track
}

func (p *Pattern) init(b []uint8) {
	p.HWversion = string(bytes.Split(b[8:19], []byte{0x00})[0]) //Trim 0x00
	binary.Read(bytes.NewReader(b[40:44]), binary.LittleEndian, &p.Tempo)
	p.Tracks = readTracks(b[44:])
}

type track struct {
	ID    int
	Name  string
	Steps string
}

func (t *track) init(b []uint8) int {
	t.ID = int(b[0])
	s := int(b[4])
	t.Name = string(b[5 : 5+s])
	if len(b) < 21+s {
		return -1
	}
	tmp := bytes.Replace(b[5+s:21+s], []byte{0x1}, []byte{0x78}, -1) //01 -> X
	tmp = bytes.Replace(tmp, []byte{0x00}, []byte{0x2d}, -1)         //00 -> -
	tmp = bytes.Join([][]byte{[]byte{}, tmp[0:4], tmp[4:8],          //add   |
		tmp[8:12], tmp[12:16], []byte{}}, []byte{0x7c})
	t.Steps = string(tmp)
	return int(21 + s) //21 = 5 + 16
}

func readTracks(b []uint8) (ret []track) {
	if len(b) >= 5 { //only if there is space
		t := track{}
		size := t.init(b)
		if size <= 0 {
			panic(errors.New("Error in reading Tracks"))
			//not good practise to panic but otherwise breaks the recursion
		}
		return append([]track{t}, readTracks(b[size:])...)
	}
	return
}

func (p Pattern) String() (ret string) {
	ret += fmt.Sprintf("Saved with HW Version: %s\n", p.HWversion)
	ret += fmt.Sprintf("Tempo: %g\n", p.Tempo)
	for _, t := range p.Tracks {
		ret += fmt.Sprintf("(%d) %s\t%s\n", t.ID, t.Name, t.Steps)
	}
	return
}
