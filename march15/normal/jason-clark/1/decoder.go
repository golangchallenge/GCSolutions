package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"os"
)

const InitialTrackCapacity = 10

type patternReadPartial func(io.Reader, *Pattern) error
type trackReadPartial func(io.Reader, *Track) error

var FileError = errors.New("Input file is not a splice file")
var patternDecoders = []patternReadPartial{readVersion, readTempo, readTracks}
var trackDecoders = []trackReadPartial{readTrackId, readTrackName, readTrackSteps}

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
// TODO: implement
func DecodeFile(path string) (*Pattern, error) {
	inputFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer inputFile.Close() // Close when function exits
	return Decode(inputFile)
}

func Decode(input io.Reader) (*Pattern, error) {
	reader, err := contentsReader(input)
	if err != nil {
		return nil, err
	}
	var p Pattern
	for _, decoder := range patternDecoders {
		err = decoder(reader, &p)
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			return nil, err
		}
	}
	return &p, nil
}

func contentsReader(input io.Reader) (io.Reader, error) {
	var header struct {
		Marker [6]byte
		_      uint32
		Length uint32
	}
	var spliceHeader = [6]byte{0x53, 0x50, 0x4c, 0x49, 0x43, 0x45} // SPLICE as bytes
	if err := readValue(input, &header); err != nil {
		return nil, err
	}
	if header.Marker != spliceHeader {
		return nil, FileError
	}
	return io.LimitReader(input, int64(header.Length)), nil
}

func readVersion(input io.Reader, pattern *Pattern) error {
	var version [32]byte
	if err := readValue(input, &version); err != nil {
		return err
	}
	//trim trailing 0s because string is zero terminated
	pattern.Version = string(bytes.TrimRight(version[:], "\u0000"))
	return nil
}

func readTempo(input io.Reader, pattern *Pattern) error {
	var tempo float32
	if err := readValue(input, &tempo); err != nil {
		return err
	}
	pattern.Tempo = float64(tempo)
	return nil
}

func readTracks(input io.Reader, pattern *Pattern) error {
	var err error
	for err == nil {
		var track Track
		for _, decoder := range trackDecoders {
			if err = decoder(input, &track); err != nil {
				return err
			}
		}
		pattern.Tracks = append(pattern.Tracks, track)
	}
	return err
}

func readTrackId(input io.Reader, track *Track) error {
	return readValue(input, &track.Id)
}

func readTrackName(input io.Reader, track *Track) error {
	var length byte
	if err := readValue(input, &length); err != nil {
		return err
	}
	nameBytes := make([]byte, length)
	if err := readValue(input, &nameBytes); err != nil {
		return err
	}
	track.Name = string(nameBytes[:])
	return nil
}

func readTrackSteps(input io.Reader, track *Track) error {
	var notes [16]byte
	if err := readValue(input, &notes); err != nil {
		return err
	}
	for i, note := range notes {
		track.Steps[i] = note != 0
	}
	return nil
}

func readValue(input io.Reader, data interface{}) error {
	return binary.Read(input, binary.LittleEndian, data)
}
