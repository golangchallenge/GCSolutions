package drum

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	data, err := beginDecode(path)

	if err != nil {
		return nil, err
	}

	p := &Pattern{}
	parseErr := parse(data, p)

	if parseErr != nil {
		return nil, parseErr
	}

	return p, nil
}

func beginDecode(path string) ([]byte, error) {
	file, err := os.Open(path)
	defer file.Close()

	if err != nil {
		return []byte{}, err
	}

	spliceBytes := make([]byte, 6)
	_, notSpliceFileErr := file.Read(spliceBytes)
	spliceHeader := string(spliceBytes)
	if notSpliceFileErr != nil || spliceHeader != "SPLICE" {
		return []byte{}, errors.New("Not a valid .splice file.")
	}

	var encodedDataSize int64
	fileSizeReadError := binary.Read(file, binary.BigEndian, &encodedDataSize)

	if fileSizeReadError != nil {
		return []byte{}, errors.New("Could not determined the encoded data size.")
	}

	data := make([]byte, encodedDataSize)
	actualDataRead, readErr := file.Read(data)

	if actualDataRead != int(encodedDataSize) || readErr != nil {
		return []byte{}, errors.New("The encoded data size is not valid.")
	}

	return data, nil
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Version string
	Tempo   float32
	Tracks  []Track
}

func (p *Pattern) String() string {
	output := fmt.Sprintf("Saved with HW Version: %s\nTempo: %g\n", p.Version, p.Tempo)

	for _, Track := range p.Tracks {
		output += fmt.Sprintln(Track.String())
	}

	return output
}

// Track represents the steps and metadata about a
// particular music track in a Pattern
type Track struct {
	ID    int32
	Name  string
	Steps []uint32
}

func (c *Track) printPattern() string {
	output := "|"

	// for each int32
	for _, val := range c.Steps {
		mask := uint32(0x000000FF)

		// for each byte
		for idx := uint32(0); idx < 4; idx++ {
			shiftedMask := mask << (idx * 8)
			if val&shiftedMask == 0x00 {
				output += "-"
			} else {
				output += "x"
			}
		}

		output += "|"
	}

	return output
}

func (c *Track) String() string {
	return fmt.Sprintf("(%d) %s\t%s", c.ID, c.Name, c.printPattern())
}
