package drum

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"os"
)

const (
	spliceTypePattern = "SPLICE"
	typeHeaderLength  = uint8(len(spliceTypePattern))
)

// ErrUnsupportedFileFormat is returned when the file to decode does not match
// the expected format.
var ErrUnsupportedFileFormat = errors.New("unsupported file format")

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return decode(bufio.NewReader(file))
}

func decode(r io.Reader) (*Pattern, error) {
	p, err := newPayloadReader(r)
	if err != nil {
		return nil, err
	}
	return decodePattern(p)
}

func newPayloadReader(r io.Reader) (payloadReader io.Reader, error error) {
	typeHeader, err := readBytes(r, typeHeaderLength)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(typeHeader, []byte(spliceTypePattern)) {
		return nil, ErrUnsupportedFileFormat
	}
	var payloadSize int64
	if err := binary.Read(r, binary.BigEndian, &payloadSize); err != nil {
		return nil, err
	}

	return io.LimitReader(r, payloadSize), nil
}

func decodePattern(r io.Reader) (*Pattern, error) {
	v, err := readBytes(r, maxVersionLength)
	if err != nil {
		return nil, err
	}
	version := cropToString(v)
	var tempo float32
	if err := binary.Read(r, binary.LittleEndian, &tempo); err != nil {
		return nil, err
	}
	tracks, err := decodeTracks(r)
	if err != nil {
		return nil, err
	}
	return &Pattern{version, tempo, tracks}, nil
}

func decodeTracks(r io.Reader) ([]*Track, error) {
	var tracks []*Track
	for {
		tr, err := decodeSingleTrack(r)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		tracks = append(tracks, tr)
	}
	return tracks, nil
}

func decodeSingleTrack(r io.Reader) (*Track, error) {
	var trackID uint32
	if err := binary.Read(r, binary.LittleEndian, &trackID); err != nil {
		return nil, err
	}
	name, err := decodeTrackName(r)
	if err != nil {
		return nil, err
	}
	steps, err := decodeSteps(r)
	if err != nil {
		return nil, err
	}
	return &Track{trackID, name, steps}, nil
}

func decodeTrackName(r io.Reader) (string, error) {
	var lenName uint8
	if err := binary.Read(r, binary.LittleEndian, &lenName); err != nil {
		return "", err
	}
	b, err := readBytes(r, lenName)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func decodeSteps(r io.Reader) (Steps, error) {
	var steps Steps
	stepsAsBytes, err := readBytes(r, stepsLength)
	if err != nil {
		return steps, err
	}
	for i, v := range stepsAsBytes {
		steps[i], err = byteToBool(v)
		if err != nil {
			return steps, err
		}
	}
	return steps, nil
}

// readBytes reads exactly n bytes from r into a new slice
// The error is EOF only if no bytes were read.
// If an EOF happens after reading some but not all the bytes,
// ReadFull returns ErrUnexpectedEOF.
func readBytes(r io.Reader, n uint8) ([]byte, error) {
	buf := make([]byte, n)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}
	return buf, nil
}
