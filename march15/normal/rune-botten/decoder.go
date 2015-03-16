package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// Reads and cleans up a string from binary data in an io.Reader
func readString(r io.Reader, length int) (n int, res string, err error) {
	buf := make([]byte, length)
	n, err = r.Read(buf)
	if err == nil {
		// Remove any zero bytes at right of string
		res = strings.TrimRight(string(buf[:]), "\x00")
	}

	return n, res, err
}

// DecodeFile implemented by Rune Botten <rbotten@gmail.com>
//
// I suspect this is not the "Go way" of doing this, and that there probably is
// some kind of interface that can be implemented to do this cleaner, but this
// is my first ever Go program so please be gentle!
func DecodeFile(path string) (*Pattern, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	// Magic "SPLICE" header (6 bytes)
	_, magic, err := readString(file, 6)
	if err != nil {
		return nil, err
	}
	if magic != "SPLICE" {
		return nil, errors.New("Invalid file")
	}

	var pattern = Pattern{}
	var bytesRead, bytesLeft int

	// File length (8 bytes)
	// Number of remaining bytes in file after these 8
	var length int64
	if err = binary.Read(file, binary.BigEndian, &length); err != nil {
		return nil, err
	}
	bytesLeft = int(length)

	// Pattern.Version (32 bytes)
	bytesRead, str, err := readString(file, 32)
	if err != nil {
		return nil, err
	}
	bytesLeft -= bytesRead
	pattern.Version = str

	// Pattern.Tempo (4 bytes)
	if err = binary.Read(file, binary.LittleEndian, &pattern.Tempo); err != nil {
		return nil, err
	}
	bytesLeft -= 4

	// Pattern.Tracks
	var tracks []Instrument
	for bytesLeft > 0 {

		var instrument = Instrument{}

		// Instrument.Id (1 byte)
		if err = binary.Read(file, binary.BigEndian, &instrument.ID); err != nil {
			return nil, err
		}
		bytesLeft -= 1 // No, Golint, bytesLeft-- isn't better in this case

		// Length of name (4 bytes)
		var nameLen int32
		if err = binary.Read(file, binary.BigEndian, &nameLen); err != nil {
			return nil, err
		}
		bytesLeft -= 4

		// Instrument.Name (variable length)
		if bytesRead, str, err = readString(file, int(nameLen)); err != nil {
			return nil, err
		}
		bytesLeft -= bytesRead
		instrument.Name = str

		// Instrument.Beats (16 bytes)
		if err = binary.Read(file, binary.BigEndian, &instrument.Beats); err != nil {
			return nil, err
		}
		bytesLeft -= len(instrument.Beats)

		tracks = append(tracks, instrument)
	}

	_ = file.Close()
	pattern.Tracks = &tracks
	return &pattern, nil
}

// Instrument represents one track in the Pattern
type Instrument struct {
	ID    uint8
	Name  string
	Beats [16]byte
}

// BeatsPerMinute represents the Pattern tempo
type BeatsPerMinute float32

// Pattern represents a composition
type Pattern struct {
	Version string
	Tempo   BeatsPerMinute
	Tracks  *[]Instrument
}

func (p *Pattern) String() string {
	var buffer bytes.Buffer

	buffer.WriteString(fmt.Sprintf("Saved with HW Version: %s\n", p.Version))

	bpmStr := strconv.FormatFloat(float64(p.Tempo), 'f', -1, 32)
	buffer.WriteString(fmt.Sprintf("Tempo: %s\n", bpmStr))

	for _, val := range *p.Tracks {
		buffer.WriteString(val.String())
	}
	return buffer.String()
}

var beatMarks = []string{"-", "x"}

func (i *Instrument) String() string {
	var buffer bytes.Buffer

	for i, val := range i.Beats {
		if i%4 == 0 {
			buffer.WriteString("|")
		}
		buffer.WriteString(beatMarks[val])
	}
	buffer.WriteString("|")
	return fmt.Sprintf("(%d) %s\t%s\n", i.ID, i.Name, buffer.String())
}
