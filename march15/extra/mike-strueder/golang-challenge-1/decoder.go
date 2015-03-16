package drum

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
)

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	header  string
	size    uint16
	Version string
	Tempo   float32
	Tracks  []Track
}

func (p Pattern) String() string {
	s := fmt.Sprintf("Saved with HW Version: %s\nTempo: %g\n", p.Version, p.Tempo)
	for _, t := range p.Tracks {
		s += t.String()
	}
	return s
}

// Track represents one track of a drum pattern
type Track struct {
	ID    uint8
	Name  string
	Steps []bool
}

func (t Track) String() string {
	var s string
	for i, step := range t.Steps {
		if i%4 == 0 {
			s += "|"
		}
		if step {
			s += "x"
		} else {
			s += "-"
		}
	}
	s += "|"
	return fmt.Sprintf("(%d) %s\t%s\n", t.ID, t.Name, s)
}

// TrackByName returns the index of the track described by n or an error if
// the track is not found
func (p *Pattern) TrackByName(n string) (int, error) {
	for i, t := range p.Tracks {
		if t.Name == n {
			return i, nil
		}
	}
	return 0, errors.Errorf("track %s not found", n)
}

// SetTrackByName replaces the steps in the given track by the provided steps.
// It returns an error in case the track name can not be found.
func (p *Pattern) SetTrackByName(n string, steps []bool) error {
	i, err := p.TrackByName(n)
	if err != nil {
		return err
	}

	p.Tracks[i] = Track{
		ID:    p.Tracks[i].ID,
		Name:  p.Tracks[i].Name,
		Steps: steps,
	}

	return nil

}

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	p := &Pattern{}
	err = Unmarshal(data, p)

	return p, err
}

// EncodeFile encodes the provided pattern and writes it to a file at the provided path.
func EncodeFile(path string, p Pattern) error {

	data, err := Marshal(p)
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = ioutil.WriteFile(path, data, 0644)

	return err
}

// Marshal encodes the provided pattern into a byte slice.
func Marshal(p Pattern) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := p.writeHeader(buf)
	err = p.writeSize(buf)
	err = p.writeVersion(buf)
	err = p.writeTempo(buf)

	for _, t := range p.Tracks {
		err = t.write(buf)
	}

	return buf.Bytes(), err
}

// Unmarshal decodes the provided byte slice into the provided pointer to a
// pattern.
func Unmarshal(data []byte, p *Pattern) error {
	buf := bytes.NewReader(data)

	header, err := readHeader(buf)
	size, err := readSize(buf)
	version, err := readVersion(buf)
	tempo, err := readTempo(buf)
	tracks, err := readTracks(buf, size-32-4)

	if err != nil {
		return err
	}

	p.header = header
	p.Version = version
	p.size = size
	p.Tempo = tempo
	p.Tracks = tracks

	return nil

}
