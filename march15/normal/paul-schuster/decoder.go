// Paul Schuster
// 01MAR15
// decoder.go - definitions for DecodeFile as provided by JoshSoftware
package drum

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"strings"
)

type errReader struct {
	file *os.File
	err  error
}

func (er *errReader) read(c int) []byte {
	if er.err != nil {
		return []byte{}
	}
	t := make([]byte, c)
	_, er.err = er.file.Read(t)
	return t
}

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	p := &Pattern{}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	er := &errReader{file: f}
	_ = er.read(13)
	size := er.read(1)

	version := er.read(32)
	p.Version = strings.Trim(string(version), "\x00")

	tempo := er.read(4)
	p.Tempo = math.Float32frombits(binary.LittleEndian.Uint32(tempo))

	tracks := er.read(int(size[0]) - 36)
	if er.err != nil {
		return nil, err
	}

	i := 0
	for i < len(tracks)-1 {
		var t Track
		t.ID = int(tracks[i])
		l := int(tracks[i+4])
		t.Name = string(tracks[i+5 : i+5+l])
		t.Pattern = []byte(tracks[i+l+5 : i+l+21])
		i += l + 21
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

// Track is the representation of .splice file track.
type Track struct {
	ID      int
	Name    string
	Pattern []byte
}

func (p *Pattern) String() string {
	var a string
	a += fmt.Sprintf("Saved with HW Version: %s\n", p.Version)
	a += fmt.Sprintf("Tempo: %g\n", p.Tempo)
	for _, v := range p.Tracks {
		a += v.String()
	}
	return a
}

func (t *Track) String() string {
	return fmt.Sprintf("(%d) %s\t%s\n", t.ID, t.Name, notation(t.Pattern))
}

func notation(m []byte) string {
	if len(m) != 16 {
		return ""
	}
	var s string
	for k, v := range m {
		if k%4 == 0 {
			s += "|"
		}
		if v == 0 {
			s += "-"
		} else {
			s += "x"
		}
	}
	s += "|"
	return s
}
