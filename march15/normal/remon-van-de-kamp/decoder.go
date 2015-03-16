package drum

import (
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	fp, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	p := new(Pattern)

	// Read header: File length, Version, Tempo
	binary.Read(fp, binary.LittleEndian, &p.Header)

	bytesRead := 36 // bytes for Version and Tempo in header
	for bytesRead < int(p.Header.Length) {

		// Read instrument header: ID and length of instrument name
		var iHead instrumentHeader
		if err := binary.Read(fp, binary.LittleEndian, &iHead); err != nil {
			return nil, err
		}
		bytesRead += binary.Size(iHead)

		// Read instrument name
		name := make([]byte, iHead.NameLength)
		if err := binary.Read(fp, binary.LittleEndian, name); err != nil {
			return nil, err
		}
		bytesRead += binary.Size(name)

		// Read hits
		var hits patternHits
		if err := binary.Read(fp, binary.LittleEndian, &hits); err != nil {
			return nil, err
		}
		bytesRead += binary.Size(hits)

		// Add track to pattern
		p.Tracks = append(p.Tracks, patternTrack{
			Id:   iHead.Id,
			Name: string(name),
			Hits: hits,
		})
	}

	return p, nil
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Header patternHeader
	Tracks []patternTrack
}

// patternTrack is the representation of a single
// instrument track. Includes instrument id, name
// and track info
type patternTrack struct {
	Id   uint32
	Name string
	Hits patternHits
}

// patternHeader represents the header of a .splice
// file. Specifically the length of the file,
// version info, and tempo.
type patternHeader struct {
	_       [13]byte // 13 bytes
	Length  uint8    // 1 byte
	Version [32]byte // 32 bytes
	Tempo   float32  // 4 bytes
}

// instrumentHeader represents the header for an
// instrument. Specifically the instrument id and
// the length of the instrument name.
type instrumentHeader struct {
	Id         uint32 // 4 bytes
	NameLength uint8  // 1 byte
}

// patternHits represents the hits in a single
// music track.
type patternHits [16]uint8 // 16 bytes

// String for a patternHits struct is used
// to display patternHits in a fashion like
//      |---x|---x|---x|---x|
func (h patternHits) String() string {
	hits := "|"
	for i, hit := range h {
		if hit == 1 {
			hits += "x"
		} else {
			hits += "-"
		}
		if (i-3)%4 == 0 {
			hits += "|"
		}
	}
	return hits
}

// String for a Pattern struct is used
// to visually represent a Pattern.
func (p Pattern) String() string {
	// format tempo as string so we can format it
	tempo := strconv.FormatFloat(float64(p.Header.Tempo), 'f', 5, 32)

	// stip tailing zeroes
	tempo = strings.TrimRight(tempo, "0")

	// if there were only zeroes after the decimal point
	// the decimal point itself can be stripped too
	tempo = strings.TrimRight(tempo, ".")

	result := fmt.Sprintf(
		"Saved with HW Version: %s\nTempo: %s\n",
		strings.TrimRight(string(p.Header.Version[:]), "\x00"),
		tempo,
	)

	for _, track := range p.Tracks {
		result += fmt.Sprintf("(%d) %s\t%s\n", track.Id, track.Name, track.Hits)
	}

	return result
}
