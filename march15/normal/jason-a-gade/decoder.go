package drum

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

const (
	corruptFileText = `"%s" is not a properly formatted pattern file.`
	unknownOffset   = 0x0D // The purpose of this value is unknown at this time.
	versionOffset   = 0x0E // 32-byte ASCII string padded by zeros.
	tempoOffset     = 0x2E // float32
	trackStart      = 0x32
	// trackIDOffset      = 0x00 // uint32
	// trackNameLenOffset = 0x04 // byte
	// trackNameString    = 0x05 // ASCII string
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	corruptFileErr := fmt.Errorf(corruptFileText, path)
	patternFile, err := getPatternFile(path)

	if err != nil {
		return nil, err
	}

	ph := &PatternHead{}
	patternHeadReader := bytes.NewReader(patternFile[:trackStart])
	err = binary.Read(patternHeadReader, binary.LittleEndian, ph)

	if err != nil {
		if err == io.EOF {
			err = corruptFileErr
		}
		return nil, err
	}

	p := &Pattern{}
	p.PatternHead = *ph

	var t Track
	var nameLen byte
	patternTracksReader := bytes.NewReader(patternFile[trackStart:])
	err = binary.Read(patternTracksReader, binary.LittleEndian, &t.ID)
	if err != nil {
		if err == io.EOF {
			err = corruptFileErr
		}
		return nil, err
	}

	// Pattern 5 breaks the pattern by including "SPLICE" and then a
	// duplicate track. We will end the loop if t.ID is "SPLI"
	splice := uint32(0x494C5053) // "SPLI"
	for err != io.EOF && t.ID != splice {
		err = binary.Read(patternTracksReader, binary.LittleEndian, &nameLen)
		if err != nil {
			if err == io.EOF {
				err = corruptFileErr
			}
			break
		}

		name := make([]byte, nameLen)
		err = binary.Read(patternTracksReader, binary.LittleEndian, name)
		if err != nil {
			if err == io.EOF {
				err = corruptFileErr
			}
			break
		}

		t.Name = string(name)

		err = binary.Read(patternTracksReader, binary.LittleEndian, &t.Step)
		if err != nil {
			if err == io.EOF {
				err = corruptFileErr
			}
			break
		}

		p.Track = append(p.Track, t)

		err = binary.Read(patternTracksReader, binary.LittleEndian, &t.ID)

	}

	if err != io.EOF && t.ID != splice {
		return nil, err
	}

	return p, nil
}

// PatternHead is part of the Pattern struct. This had to be created because
// binary.Read() expects datatypes of fixed size. It should not be used
// independently; it is only exported because of the requirements of
// binary.Read().
type PatternHead struct {
	Magic   [6]byte
	_       [7]byte
	Unknown byte
	Version [32]byte
	Tempo   float32
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	PatternHead
	Track []Track
}

// Stringer interface for Pattern.
func (p *Pattern) String() string {
	var st bytes.Buffer
	_, err := st.WriteString("Saved with HW Version: ")
	if err != nil {
		return ""
	}

	version := bytes.TrimRight(p.Version[:], "\x00")
	_, err = st.WriteString(string(version) + "\nTempo: ")
	if err != nil {
		return ""
	}

	_, err = st.WriteString(fmt.Sprintf("%v", p.Tempo) + "\n")
	if err != nil {
		return ""
	}

	for _, t := range p.Track {
		_, err = st.WriteString(t.String())
		if err != nil {
			return ""
		}
	}

	return st.String()
}

// Track is a part of the Pattern structure which represents a single
// sample.
type Track struct {
	ID   uint32
	Name string
	Step step
}

// Stringer interface for Track.
func (t Track) String() string {
	var st bytes.Buffer
	var err error
	_, err = st.WriteString(fmt.Sprintf("(%v) %s\t", t.ID, t.Name))
	if err != nil {
		return ""
	}

	_, err = st.WriteString(t.Step.String())
	if err != nil {
		return ""
	}

	return st.String()
}

// step is part of the Track struct.
type step [16]byte

// Stringer interface for step.
func (s step) String() string {
	var st bytes.Buffer
	var err error
	for k, v := range s {
		if k%4 == 0 {
			_, err = st.WriteString("|")
			if err != nil {
				return ""
			}
		}
		switch v {
		case 0:
			_, err = st.WriteString("-")
		case 1:
			_, err = st.WriteString("x")
		default:
			return ""
		}
	}
	_, err = st.WriteString("|\n")
	if err != nil {
		return ""
	}

	return st.String()

}

// getPatternFile reads in a binary representation of the Pattern file.
func getPatternFile(path string) ([]byte, error) {
	corruptFileErr := fmt.Errorf(corruptFileText, path)
	empty := []byte{}

	f, err := os.Open(path)
	if err != nil {
		return empty, err
	}
	defer f.Close()

	// Check for a valid header before reading the rest of the file.
	magic := []byte("SPLICE")
	head := make([]byte, len(magic))

	_, err = f.Read(head)
	if err != nil {
		return empty, err
	}

	if !bytes.Equal(magic, head) {
		return empty, corruptFileErr
	}

	// If we get here then we can read the rest of the file.
	stat, err := f.Stat()
	if err != nil {
		return empty, err
	}

	size := stat.Size()
	body := make([]byte, size)
	_, err = f.ReadAt(body, 0)
	if err != nil {
		return empty, err
	}

	return body, nil
}
