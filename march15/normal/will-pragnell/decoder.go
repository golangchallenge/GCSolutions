package drum

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"strings"
)

const spliceHeader = "SPLICE"

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	// ok to ignore error here since os.Open opens the file in read-only mode
	defer file.Close()

	reader := &errBinaryReader{reader: file}

	// header is only the first six bytes, but the next seven after that are
	// just padding, so we read them all at once to save an additional read call
	header := reader.readBytes(13)
	if string(header[:6]) != spliceHeader {
		if reader.err != nil {
			return nil, reader.err
		}
		return nil, errors.New("unrecognised file type or format")
	}

	bytesRemaining := reader.readBytes(1)[0]

	version := trimZeroBytesFromString(string(reader.readBytes(32)))
	tempo := float32fromLittleEndianBytes(reader.readBytes(4))
	bytesRemaining -= 36

	var tracks []track
	for bytesRemaining > 0 && reader.err == nil {
		track, bytesRead := readTrack(reader)
		tracks = append(tracks, track)
		bytesRemaining -= bytesRead
	}

	if reader.err != nil {
		return nil, fmt.Errorf("%v: unable to decode\nerror: %v", path, reader.err)
	}
	return &Pattern{
		version: version,
		tempo:   tempo,
		tracks:  tracks,
	}, nil
}

func trimZeroBytesFromString(s string) string {
	return strings.Trim(s, string([]byte{0}))
}

func float32fromLittleEndianBytes(bytes []byte) float32 {
	return math.Float32frombits(binary.LittleEndian.Uint32(bytes))
}

func readTrack(reader *errBinaryReader) (track, uint8) {
	trackHeader := reader.readBytes(5)
	trackNameLength := uint8(trackHeader[4])
	name := string(reader.readBytes(trackNameLength))

	steps := make([]bool, 16)
	for i, b := range reader.readBytes(16) {
		steps[i] = b > 0
	}

	return track{
		id:    uint8(trackHeader[0]),
		name:  name,
		steps: steps,
	}, 5 + trackNameLength + 16
}

// errBinaryReader is a convinence wrapper around binary.Read and an io.Reader
// which has a "sticky" error, preventing the need for repeated error checking
// at call sites
type errBinaryReader struct {
	reader io.Reader
	err    error
}

func (ebr *errBinaryReader) readBytes(n uint8) []byte {
	data := make([]byte, n)
	if ebr.err != nil {
		// return empty data in err cases so that callers don't have
		// to check whether the returned value is nil every time
		return data
	}
	ebr.err = binary.Read(ebr.reader, binary.BigEndian, &data)
	return data
}
