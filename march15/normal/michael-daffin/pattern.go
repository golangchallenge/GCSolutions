package drum

import (
	"encoding/binary"
	"fmt"
	"io"
	"strings"
)

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Version string
	Tempo   float32
	Tracks  []Track
}

// Decode a binary sequence containing a drum pattern.
func (p *Pattern) Decode(r io.Reader) error {
	h := struct {
		MagicNumber [12]byte
		Length      uint16 // Length tracks the remaing length of the file
		Version     [32]byte
	}{}

	// Read the files header
	if err := binary.Read(r, binary.BigEndian, &h); err != nil {
		return err
	}
	h.Length -= 32 // Version string

	// Check the magic number to verify the file type
	if h.MagicNumber != [12]byte{'S', 'P', 'L', 'I', 'C', 'E'} {
		return fmt.Errorf("invalid file format")
	}

	p.Version = strings.Trim(string(h.Version[:32]), "\x00")

	if err := binary.Read(r, binary.LittleEndian, &p.Tempo); err != nil {
		return err
	}
	h.Length -= 4 // Tempo float

	// Rest of file represents the tracks, read until we have read h.Length bytes
	for h.Length > 0 {
		t := Track{}

		length, err := t.Decode(r)
		if err != nil {
			return err
		}

		h.Length -= uint16(length) // Track length, depends on size of title

		p.Tracks = append(p.Tracks, t)
	}

	return nil
}

// String returns the pattern in a printable string
func (p Pattern) String() string {
	tracks := make([]string, len(p.Tracks))

	for index, track := range p.Tracks {
		tracks[index] = track.String()
	}

	return fmt.Sprintf("Saved with HW Version: %s\nTempo: %v\n%s\n", p.Version, p.Tempo, strings.Join(tracks, "\n"))
}
