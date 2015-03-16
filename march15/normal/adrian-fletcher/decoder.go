package drum

import (
	"encoding/binary"
	"fmt"
	"os"
	"strings"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data
func DecodeFile(path string) (*Pattern, error) {
	p := &Pattern{}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	// See gochallenge.pdf for more details.
	// Header Details:
	// 13 bytes - are a file header
	// 1  byte  - filesize in bytes
	// 32 bytes - pattern name
	// 4  bytes - tempo in float32 format

	// Read in Header
	err = binary.Read(file, binary.LittleEndian, &p.header)
	if err != nil {
		return nil, err
	}

	// Tracks are made from the following binary representation
	//  4 Bytes - Track ID
	//  1 Byte  - Size of Track Name (in bytes)
	//  ? Bytes - Track Name (Maximum 255 characters)
	// 16 Bytes - 16 steps represented by either a 0x00 or a 0x01 (beat)

	// Temporary track for storage
	var t track

	trackSize := p.Filesize - 32 - 4   // minutes 'version' and 'tempo'
	for p.TrackBytesRead < trackSize { // filesize minus file version

		p.Read(file, &t.ID)         // Read Track ID
		p.Read(file, &t.NameLength) // Track name Length

		// Track Name
		trackName := make([]byte, t.NameLength)
		p.Read(file, &trackName)
		t.Name = string(trackName[:])

		p.Read(file, &t.Steps) // Steps - 16 bytes

		// Did we read the tracks correctly
		if p.ReadError != nil {
			return p, p.ReadError
		}

		p.Tracks = append(p.Tracks, t)
	}

	return p, nil
}

func (t track) String() string {
	return fmt.Sprintf("(%d) %s	|%s|%s|%s|%s|\n",
		t.ID,
		t.Name,
		markup(t.Steps[:4]),
		markup(t.Steps[4:8]),
		markup(t.Steps[8:12]),
		markup(t.Steps[12:16]))
}

// Convert bytes 0x0 or 0x1 in to 'x' or '-' for user presentation.
func markup(b []byte) string {
	// Formatted Bytes
	var fb [4]byte
	for i, val := range b {
		switch val {
		case 0x00:
			fb[i] = '-'
		case 0x01:
			fb[i] = 'x'
		}
	}
	return string(fb[:])
}

func (p *Pattern) String() string {
	var tracks = make([]string, 0)
	for _, val := range p.Tracks {
		tracks = append(tracks, fmt.Sprint(val))
	}

	return fmt.Sprintf("Saved with HW Version: %s\nTempo: %g\n%s", strings.TrimRight(string(p.Version[:]), string(0x00)), p.Tempo, strings.Join(tracks, ""))
}

func (p *Pattern) Read(file *os.File, destination interface{}) {
	p.TrackBytesRead += uint8(binary.Size(destination))
	err := binary.Read(file, binary.LittleEndian, destination)
	if err != nil {
		p.ReadError = err
	}
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	header
	TrackBytesRead uint8
	ReadError      error
	Tracks         []track
}

type header struct {
	Identifier [13]byte
	Filesize   uint8
	Version    [32]byte
	Tempo      float32
}

type track struct {
	ID         uint32
	NameLength uint8
	Name       string
	Steps      [16]byte
}
