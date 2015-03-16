package drum

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

const (
	headerLength         = 6
	patternStartPosition = 14
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
// The formtat of drum machine file is:
//   First 6 bytes is header.
//   Next 8 bytes is length of one pattern.
//   Next is the content of pattern.
//   There may be some unknow contents at the end of file, assume they are
//   ignorable.
// The format of Pattern:
//   First 32 bytes is pattern version.
//   Next 4 bytes is pattern tempo.
//   Next bytes contains number of tracks.
// The format of Track:
//   First 2 bytes is track id.
//   Next 4 bytes is the length of track name.
//   Next variable bytes is track name.
//   Last 16 bytes is the steps.
func DecodeFile(path string) (*Pattern, error) {
	p := &Pattern{}
	p.Tracks = make([]*Track, 0)

	// Open file from path
	f, err := os.Open(path)
	if err != nil {
		return p, err
	}
	defer f.Close()

	// Skip the header
	_, err = f.Seek(headerLength, 0)
	if err != nil {
		return p, err
	}

	// Get the length of one pattern
	var pLen int64
	err = binary.Read(f, binary.BigEndian, &pLen)
	if err != nil {
		return p, err
	}

	// Get the binary content of one pattern
	b := make([]byte, pLen)
	_, err = f.ReadAt(b, patternStartPosition)
	if err != nil {
		return p, err
	}

	err = p.dump(b)
	return p, err
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Version string
	Tempo   float32
	Tracks  []*Track
}

func (p *Pattern) dump(b []byte) error {
	buf := bytes.NewReader(b)

	// Get Version
	var version [32]byte
	err := binary.Read(buf, binary.BigEndian, &version)
	if err != nil {
		return err
	}
	p.Version = string(bytes.Trim(version[:], "\x00"))

	// Get Tempo
	err = binary.Read(buf, binary.LittleEndian, &p.Tempo)
	if err != nil {
		return err
	}

	// Get all tracks
	for {
		track := new(Track)

		// Get track ID
		err = binary.Read(buf, binary.BigEndian, &track.ID)
		if err != nil {
			break
		}

		// Get the length of track Name
		var nameLen uint32
		err = binary.Read(buf, binary.BigEndian, &nameLen)
		if err != nil {
			break
		}

		// Get track Name
		name := make([]byte, nameLen)
		err = binary.Read(buf, binary.BigEndian, &name)
		if err != nil {
			break
		}
		track.Name = string(name[:])

		// Get track Steps
		err = binary.Read(buf, binary.BigEndian, &track.Steps)
		if err != nil {
			break
		}
		p.Tracks = append(p.Tracks, track)
	}
	if err != io.EOF {
		return err
	}
	return nil

}

func (p Pattern) String() string {
	var buff bytes.Buffer
	buff.WriteString(fmt.Sprintf("Saved with HW Version: %s\n", p.Version))
	buff.WriteString(fmt.Sprintf("Tempo: %v\n", p.Tempo))
	for _, track := range p.Tracks {
		buff.WriteString(fmt.Sprintf("%s\n", track))
	}
	return buff.String()
}

// Track is the high level representation of the one track in drum pattern
// contained in a .splice file.
type Track struct {
	ID    uint8
	Name  string
	Steps [16]byte
}

func (t Track) String() string {
	var buff bytes.Buffer
	for index, step := range t.Steps {
		if index%4 == 0 {
			buff.WriteRune('|')
		}
		if step == 0x00 {
			buff.WriteRune('-')
		} else {
			buff.WriteRune('x')
		}
	}
	buff.WriteRune('|')
	return fmt.Sprintf("(%d) %s\t%s", t.ID, t.Name, buff.String())
}
