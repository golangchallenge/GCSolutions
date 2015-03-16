package drum

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	fi, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	var head header
	binary.Read(fi, binary.LittleEndian, &head)

	var tracks []Track
	for {
		var id [4]byte
		if _, err := fi.Read(id[:]); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		var nameStringLength [1]byte
		if _, err := fi.Read(nameStringLength[:]); err != nil {
			return nil, err
		}

		name := make([]byte, nameStringLength[0])
		if _, err := fi.Read(name[:]); err != nil {
			return nil, err
		}

		var measure [16]byte
		if _, err := fi.Read(measure[:]); err != nil {
			// the following 3 lines probably shouldn't be here since reading an EOF
			// here would make more sense to be an error, but pattern_5.splice has
			// some weird input. We'll just ignore this track, carry on and create our
			// pattern
			if err == io.EOF {
				break
			}

			return nil, err
		}
		tracks = append(tracks, Track{
			Id:      binary.LittleEndian.Uint32(id[:]),
			Name:    string(name),
			Measure: measure,
		})
	}

	return &Pattern{
		Tempo:   math.Float32frombits(binary.LittleEndian.Uint32(head.Tempo[:])),
		Version: strings.TrimRight(string(head.Version[:]), "\x00"),
		Tracks:  tracks,
	}, nil
}

// Represents the binary structure of the header
type header struct {
	_       [14]byte // string "SPLICE........"
	Version [32]byte // string
	Tempo   [4]byte  // 32bit float
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Tempo   float32
	Version string
	Tracks  []Track
}

func (p *Pattern) String() string {

	writer := new(bytes.Buffer)

	writer.WriteString("Saved with HW Version: ")
	writer.WriteString(p.Version)
	writer.WriteRune('\n')

	writer.WriteString("Tempo: ")

	// Write out the tempo with "-1" precision
	// "The special precision -1 uses the smallest number of digits necessary
	// such that ParseFloat will return f exactly"
	// - http://golang.org/pkg/strconv/#FormatFloat
	writer.WriteString(strconv.FormatFloat(float64(p.Tempo), 'f', -1, 32))
	writer.WriteRune('\n')

	for _, tracks := range p.Tracks {
		writer.WriteString(tracks.String())
	}

	return writer.String()
}

type Track struct {
	Id   uint32
	Name string

	// Might make more sense to be [16]bool so that it's makes sense semantically.
	// It'll easier to write back the same format if we leave it as [16]byte,
	// though
	Measure [16]byte
}

// converts array of bytes into something pretty
// zero values are outputted as -
// any non-zero value is outputted as x
// ex: 0x0101 => -x-x
//     0x1001 => x--x
func bytesToPrettyBar(bar []byte) string {
	writer := new(bytes.Buffer)
	for _, p := range bar {
		if p != 0 {
			writer.WriteRune('x')
		} else {
			writer.WriteRune('-')
		}
	}
	return writer.String()
}

func (i *Track) String() string {
	return fmt.Sprintf("(%d) %s\t|%v|%v|%v|%v|\n",
		i.Id,
		i.Name,
		bytesToPrettyBar(i.Measure[0:4]),
		bytesToPrettyBar(i.Measure[4:8]),
		bytesToPrettyBar(i.Measure[8:12]),
		bytesToPrettyBar(i.Measure[12:16]),
	)
}
