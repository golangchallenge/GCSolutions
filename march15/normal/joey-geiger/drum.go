// Package drum is supposed to implement the decoding of .splice drum machine files.
// See golang-challenge.com/go-challenge1/ for more information
package drum

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// Header for binary drum machine file
// FileHeader: describes the file type
// ContentLength: number of bytes that make up the rest of the file
type Header struct {
	FileHeader    [6]byte
	ContentLength int64
}

func (h Header) parse(r io.Reader) ([]byte, error) {
	err := binary.Read(r, binary.BigEndian, &h)
	if err != nil {
		fmt.Println("read Header failed:", err)
		return nil, err
	}

	var data = make([]byte, h.ContentLength)
	binary.Read(r, binary.BigEndian, &data)
	if err != nil {
		fmt.Println("read ContentLength parse failed:", err)
		return nil, err
	}

	return data, nil
}

// Track for Pattern
// ID: identifier for the Track
// Name: name of the track instrument
// Steps: 16 steps that make up the track
type Track struct {
	ID    uint8
	Name  string
	Steps [16]byte
}

func (t *Track) parse(b *bytes.Reader) error {
	err := binary.Read(b, binary.BigEndian, &t.ID)
	if err != nil {
		fmt.Println("read Track ID failed:", err)
		return err
	}

	t.Name, err = t.parseName(b)
	if err != nil {
		fmt.Println("parse Track name failed:", err)
		return err
	}

	err = binary.Read(b, binary.BigEndian, &t.Steps)
	if err != nil {
		fmt.Println("read Steps failed:", err)
		return err
	}

	return nil
}

func (t *Track) parseName(b *bytes.Reader) (string, error) {
	var strLength uint32
	err := binary.Read(b, binary.BigEndian, &strLength)
	if err != nil {
		fmt.Println("read Track name length failed:", err)
		return "", err
	}

	trackName := make([]byte, strLength)
	err = binary.Read(b, binary.BigEndian, &trackName)
	if err != nil {
		fmt.Println("read Track name bytes failed:", err)
		return "", err
	}

	return string(trackName[:]), nil
}

func (t Track) String() string {
	return fmt.Sprintf("(%v) %s\t%v\n", t.ID, t.Name, t.stepsToString())
}

func (t Track) stepsToString() string {
	s := ""
	for i, step := range t.Steps {
		if i%4 == 0 {
			s += "|"
		}
		if step == 1 {
			s += "x"
		} else {
			s += "-"
		}
	}
	s += "|"

	return s
}

// Pattern include information to describe the pattern
// Version: version used to create the pattern
// Tempo: tempo used to play the pattern
// Tracks: tracks included in the pattern
type Pattern struct {
	Version string
	Tempo   float32
	Tracks  []Track
}

func (p *Pattern) parseInfo(b *bytes.Reader) error {
	var version [32]byte
	var tempo float32

	err := binary.Read(b, binary.BigEndian, &version)
	if err != nil {
		fmt.Println("read Pattern Version failed:", err)
		return err
	}

	err = binary.Read(b, binary.LittleEndian, &tempo)
	if err != nil {
		fmt.Println("read Pattern Tempo failed:", err)
		return err
	}

	p.Version = string(bytes.Trim(version[:], "\x00"))
	p.Tempo = tempo

	return nil
}

func (p *Pattern) parseTracks(r *bytes.Reader) error {
	for r.Len() > 0 {
		t := &Track{}
		err := t.parse(r)
		if err != nil {
			fmt.Println("parse Track failed:", err)
			return err
		}
		p.Tracks = append(p.Tracks, *t)
	}

	return nil
}

func (p *Pattern) parse(data []byte) error {
	b := p.getBuffer(data)
	err := p.parseInfo(b)
	if err != nil {
		fmt.Println("parse Pattern failed:", err)
		return err
	}

	err = p.parseTracks(b)
	if err != nil {
		fmt.Println("parse Tracks failed:", err)
		return err
	}

	return nil
}

func (p Pattern) getBuffer(data []byte) *bytes.Reader {
	return bytes.NewReader(data)
}

func (p Pattern) String() string {
	return fmt.Sprintf("Saved with HW Version: %s\nTempo: %v\n%v", p.Version, p.Tempo, p.tracksToString())
}

func (p Pattern) tracksToString() string {
	s := ""
	for _, track := range p.Tracks {
		s += track.String()
	}

	return s
}
