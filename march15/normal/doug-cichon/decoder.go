package drum

import (
	"encoding/binary"
	"os"
	"strings"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	p := &Pattern{}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Read header, but discard it.
	if _, err = file.Seek(6, 0); err != nil {
		return nil, err
	}

	// Read number of total bytes left in file
	var bytesRemaining uint64
	if err = binary.Read(file, binary.BigEndian, &bytesRemaining); err != nil {
		return nil, err
	}

	// Read the version number
	version := make([]byte, 32)
	if n, err := file.Read(version); err != nil || n < 32 {
		return nil, err
	}
	p.Version = strings.TrimRight(string(version), "\x00")
	bytesRemaining -= 32

	// Read the tempo
	if err = binary.Read(file, binary.LittleEndian, &p.Tempo); err != nil {
		return nil, err
	}
	bytesRemaining -= 4

	for bytesRemaining > 0 {
		t, n, err := decodeTrack(file)

		if err != nil {
			return nil, err
		}

		p.Tracks = append(p.Tracks, t)
		bytesRemaining -= n
	}

	return p, nil
}

func decodeTrack(file *os.File) (*Track, uint64, error) {
	t := &Track{}
	var bytesRead uint64
	// Read id
	if err := binary.Read(file, binary.LittleEndian, &t.ID); err != nil {
		return nil, 0, err
	}
	bytesRead += 4

	// Read description length
	var instrumentLength byte
	if err := binary.Read(file, binary.BigEndian, &instrumentLength); err != nil {
		return nil, 0, err
	}
	bytesRead++

	// Read description
	instrument := make([]byte, instrumentLength)
	if _, err := file.Read(instrument); err != nil {
		return nil, 0, err
	}
	t.Instrument = string(instrument)
	bytesRead += uint64(instrumentLength)

	// Read beats
	var beatBytes [16]byte
	if err := binary.Read(file, binary.BigEndian, &beatBytes); err != nil {
		return nil, 0, err
	}
	bytesRead += 16

	// Convert from bytes to booleans
	for i := 0; i < 16; i++ {
		t.Beats[i] = beatBytes[i] == 1
	}

	return t, bytesRead, nil
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Version string
	Tempo   float32
	Tracks  []*Track
}

// Track represents a single instrument's portion of a Pattern.
type Track struct {
	ID         uint32
	Instrument string
	Beats      [16]bool
}
