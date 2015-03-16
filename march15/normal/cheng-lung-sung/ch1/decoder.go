package drum

import (
	"bytes"
	"fmt"
	"io/ioutil"
)

var byteStep = []byte("-x")

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	binData, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	p, err := Decode(bytes.NewReader(binData))
	if err != nil {
		return nil, err
	}
	//	p.FileName = filepath.Base(path)
	return p, err
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Version [32]byte
	Tempo   float32
	Tracks  []Track
}

func (p *Pattern) String() string {
	var dst bytes.Buffer

	dst.WriteString("Saved with HW Version: ")
	for _, b := range p.Version {
		if b == uint8(0) {
			break
		}
		dst.WriteByte(b)
	}
	dst.WriteString(fmt.Sprintf("\nTempo: %.3g\n", p.Tempo))
	for _, t := range p.Tracks {
		dst.WriteString(t.String())
	}
	return dst.String()
}

// Track keeps the information of the specific drum track
type Track struct {
	ID    uint32
	Name  []byte
	Steps [2]uint8
}

func (t *Track) String() string {
	var dst bytes.Buffer
	for i := 0; i < len(t.Steps); i++ {
		for j := uint(0); j < 4; j++ {
			dst.WriteByte(byteStep[(t.Steps[i]>>j)%2])
		}
		dst.WriteByte('|')
		for j := uint(0); j < 4; j++ {
			dst.WriteByte(byteStep[(t.Steps[i]>>(4+j))%2])
		}
		dst.WriteByte('|')
	}
	return fmt.Sprintf("(%d) %s\t|%s\n", t.ID, t.Name, dst.String())
}
