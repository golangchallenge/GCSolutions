package drum

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"strings"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error decoding file: %v", err)
	}

	p := &Pattern{}
	if err := p.UnmarshalBinary(contents); err != nil {
		return nil, fmt.Errorf("error unmarshalling binary: %v", err)
	}

	return p, nil
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Header struct {
		Title  [6]byte
		_      [7]byte
		Length uint8
	}
	Version string
	Tempo   float32
	Tracks  []*Track
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface.
func (p *Pattern) UnmarshalBinary(data []byte) error {
	src := bytes.NewReader(data)

	// Unmarshal header of pattern
	if err := binary.Read(src, binary.LittleEndian, &p.Header); err != nil {
		return fmt.Errorf("error unmarshalling length of pattern: %v", err)
	}

	// Unmarshal version of pattern
	const metaLength = 36
	versionBuf := make([]byte, metaLength-4)
	if _, err := src.Read(versionBuf); err != nil {
		return fmt.Errorf("error unmarshalling version of pattern: %v", err)
	}
	nullIndex := strings.IndexByte(string(versionBuf), byte(0))
	p.Version = string(versionBuf)[:nullIndex]

	// Unmarshal tempo of pattern
	if err := binary.Read(src, binary.LittleEndian, &p.Tempo); err != nil {
		return fmt.Errorf("error unmarshalling tempo of pattern: %v", err)
	}

	// Unmarshal tracks of pattern
	for read := 0; read < int(p.Header.Length-metaLength); {
		track := &Track{}

		lenBefore := src.Len()
		if err := track.UnmarshalBinary(src); err != nil {
			return fmt.Errorf("error unmarshalling track: %v", err)
		}
		read += lenBefore - src.Len()

		p.Tracks = append(p.Tracks, track)
	}

	return nil
}

// String returns a formatted Pattern with both its metadata and tracks.
func (p *Pattern) String() string {
	version := fmt.Sprintf("Saved with HW Version: %v", p.Version)
	tempo := fmt.Sprintf("Tempo: %v", p.Tempo)

	output := []string{version, tempo}
	for _, track := range p.Tracks {
		output = append(output, track.String())
	}

	return strings.Join(output, "\n") + "\n"
}
