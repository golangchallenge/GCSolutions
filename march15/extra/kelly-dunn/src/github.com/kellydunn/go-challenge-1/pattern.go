package drum

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
)

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Version string
	Tempo   float32
	Tracks  []*Track
}

// NewPattern creates and returns a pointer to a new Pattern.
// Returns an error if one was encountered from attempting to read
// from the passed in io.Reader.  This can happen if the
// passed in reader does not comply with the binary Splice format.
//
// To learn more about the format, visit the README.md at the root level of this project.
func NewPattern(reader io.Reader) (*Pattern, error) {
	p := &Pattern{}

	// Read the "SPLICE" header from the binary file.
	spliceFile := make([]byte, SpliceFileSize)
	err := binary.Read(reader, binary.BigEndian, spliceFile)
	if err != nil {
		return nil, err
	}

	// Read the size of the file from the header.
	var size uint64
	err = binary.Read(reader, binary.BigEndian, &size)
	if err != nil {
		return nil, err
	}

	// This counter keeps track of all the bytes read after the headers
	// such that it will ignore garbage data that occurs after the designated size.
	bytesRead := 0

	// Reads the Pattern Version from the passed in binary file.
	read, err := readPatternVersion(reader, p)
	if err != nil {
		return nil, err
	}

	bytesRead += read

	// Reads the Pattern Tempo from the passed in binary file.
	read, err = readPatternTempo(reader, p)
	if err != nil {
		return nil, err
	}

	bytesRead += read

	p.Tracks = []*Track{}

	// Until we read up to the passed in size of the file
	// We will continue to consume data as Tracks.
	for bytesRead < int(size) {

		t := &Track{}

		// Reads the Track Id from the passed in binary file.
		read, err = readTrackID(reader, t)
		if err != nil {
			return nil, err
		}

		bytesRead += read

		// Reads the Track Name from the passed in binary file.
		read, err = readTrackName(reader, t)
		if err != nil {
			return nil, err
		}

		bytesRead += read

		// Reads the Track Step sequence from the passed in binary file.
		read, err = readTrackStepSequence(reader, t)
		if err != nil {
			return nil, err
		}

		bytesRead += read

		p.Tracks = append(p.Tracks, t)
	}

	return p, nil
}

// String implements the Stringer interface.
// Returns a string representation of the current Pattern.
func (p *Pattern) String() string {
	buf := bytes.NewBufferString("")
	_, err := buf.WriteString(fmt.Sprintf("Saved with HW Version: %s\n", p.Version))
	if err != nil {
		log.Printf("Error writing string to buffer: %v", err)
	}

	_, err = buf.WriteString(fmt.Sprintf("Tempo: %3v\n", p.Tempo))
	if err != nil {
		log.Printf("Error writing string to buffer: %v", err)
	}

	for _, track := range p.Tracks {
		_, err := buf.WriteString(fmt.Sprintf("(%d) %s\t%s\n", track.ID, track.Name, track.StepSequence))
		if err != nil {
			log.Printf("Error writing string to buffer: %v", err)
		}
	}

	return buf.String()
}

// Reads the pattern Version from the passed in reader and applies to the passed in pattern pointer.
// Returns the number of bytes read, or an error if there is one that is encountered.
func readPatternVersion(reader io.Reader, p *Pattern) (int, error) {
	version := make([]byte, VersionSize)
	err := binary.Read(reader, binary.BigEndian, version)
	if err != nil {
		return 0, err
	}

	p.Version = string(bytes.Trim(version, EmptyByteString))

	return VersionSize, nil
}

// Reads the pattern Tempo from the passed in reader and applies to the passed in pattern pointer.
// Returns the number of bytes read, or an error if there is one that is encountered.
func readPatternTempo(reader io.Reader, p *Pattern) (int, error) {
	var tempo float32
	err := binary.Read(reader, binary.LittleEndian, &tempo)
	if err != nil {
		return 0, err
	}

	p.Tempo = tempo

	return TempoSize, nil
}
