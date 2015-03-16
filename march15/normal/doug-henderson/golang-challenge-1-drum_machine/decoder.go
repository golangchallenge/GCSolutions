// Go Challenge 1 deadline 2015.03.15
// http://golang-challenge.com/go-challenge1/
// for details
// go test -v decoder.go decoder_test.go

package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	//fmt.Print("DecodeFile: path =",path,"\n")
	p := &Pattern{}
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return p, err
	}
	reader := bytes.NewReader(buf)

	err = p.Head.Decode(reader)
	if err != nil {
		return p, err
	}

	for {
		if reader.Len() == 0 {
			break
		}
		var t Track
		err = t.Decode(reader)
		if err != nil {
			return p, err
		}
		if t.Spliced {
			prev := len(p.Tracks) - 1
			t.Ident = p.Tracks[prev].Ident
			p.Tracks[prev] = t
		} else {
			p.Tracks = append(p.Tracks, t)
		}
	}

	return p, nil
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Head   Header
	Tracks []Track
}

func (p Pattern) String() string {
	var buf bytes.Buffer
	s := p.Head.String()
	buf.WriteString(s)
	for _, value := range p.Tracks {
		s = value.String()
		buf.WriteString(s)
	}
	return buf.String()
}

// Header is the .splice file header
type Header struct {
	Sentinal      [12]uint8
	ContentLength uint16
	Version       [32]uint8
	Tempo         float32
}

func (p Header) String() string {
	versionLen := bytes.Index(p.Version[:], []byte{0})
	version := string(p.Version[:versionLen])
	s := fmt.Sprint("Saved with HW Version: ", version, "\nTempo: ", p.Tempo, "\n")
	return s
}

// Decode parses the header from a .splice file
func (p *Header) Decode(reader io.Reader) (err error) {
	err = binary.Read(reader, binary.LittleEndian, &p.Sentinal)
	if err != nil {
		return err
	}
	goodSentinal := [12]uint8{'S', 'P', 'L', 'I', 'C', 'E', 0}
	if p.Sentinal != goodSentinal {
		return errors.New("Invalid Splice file")
	}
	err = binary.Read(reader, binary.BigEndian, &p.ContentLength)
	if err != nil {
		return err
	}
	err = binary.Read(reader, binary.LittleEndian, &p.Version)
	if err != nil {
		return err
	}
	err = binary.Read(reader, binary.LittleEndian, &p.Tempo)
	if err != nil {
		return err
	}
	return nil
}

const Steps = 16

// Track represents a single track
type Track struct {
	Spliced    bool
	Ident      uint32
	NameLength uint8
	Name       [12]uint8
	Measure    [Steps]uint8
}

func (t Track) String() string {
	name := string(t.Name[:t.NameLength])
	var sd [21]byte
	j := 0
	for step := 0; step < Steps; step++ {
		if step%4 == 0 {
			sd[j] = '|'
			j++
		}
		if t.Measure[step] != 0 {
			sd[j] = 'x'
		} else {
			sd[j] = '-'
		}
		j++
	}
	sd[j] = '|'
	s := fmt.Sprintf("(%d) %s\t%s\n", t.Ident, name, string(sd[:21]))
	return s
}

// Decode parses a track from a .splice file.
func (t *Track) Decode(reader io.Reader) (err error) {
	err = binary.Read(reader, binary.LittleEndian, &t.Ident)
	if err != nil {
		return err
	}
	if t.Ident == ('S' + ('P'+('L'+('I')<<8)<<8)<<8) {
		var junk [5]uint8
		err = binary.Read(reader, binary.LittleEndian, &junk)
		if err != nil {
			return err
		}
		t.Spliced = true
	}
	err = binary.Read(reader, binary.LittleEndian, &t.NameLength)
	if err != nil {
		return err
	}
	err = binary.Read(reader, binary.LittleEndian, t.Name[:t.NameLength])
	if err != nil {
		return err
	}
	err = binary.Read(reader, binary.LittleEndian, &t.Measure)
	if err != nil {
		return err
	}
	return nil
}
