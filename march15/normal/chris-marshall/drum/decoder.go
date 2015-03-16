package drum

import (
	"bufio"
	"encoding/binary"
	"os"
	"strings"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	file, err := os.Open(path)
	if err != nil {
		return &(Pattern{}), err
	}
	reader := bufio.NewReader(file)

	pattern := readVersionAndTempo(reader)
	for anyTracks(reader) {
		pattern.Tracks = append(pattern.Tracks, readTrack(reader))
	}

	return &pattern, nil
}

type decodedFile struct {
	Header  [13]byte
	_       [1]byte
	Version [11]byte
	_       [21]byte
	Tempo   float32
}

func readVersionAndTempo(reader *bufio.Reader) Pattern {
	var decoded decodedFile
	var pattern Pattern

	binary.Read(reader, binary.LittleEndian, &decoded)
	pattern.Version = trimVersion(decoded.Version)
	pattern.Tempo = decoded.Tempo

	return pattern
}

func readTrack(reader *bufio.Reader) Track {
	return Track{
		ID:    readTrackID(reader),
		Name:  readTrackName(reader),
		Steps: readTrackSteps(reader),
	}
}

func readTrackID(reader *bufio.Reader) int {
	var id int32

	binary.Read(reader, binary.LittleEndian, &id)
	return int(id)
}

func readTrackName(reader *bufio.Reader) string {
	var name []byte

	next, err := reader.Peek(1)
	for err == nil && next[0] != 0 && next[0] != 1 {
		charByte, _ := reader.ReadByte()
		name = append(name, charByte)
		next, err = reader.Peek(1)
	}

	return string(name[1:])
}

func readTrackSteps(reader *bufio.Reader) [16]Step {
	var decodedSteps [16]int8
	var steps [16]Step

	binary.Read(reader, binary.LittleEndian, &decodedSteps)
	for i, intStep := range decodedSteps {
		steps[i] = Step{Active: intStep != 0}
	}

	return steps
}

func anyTracks(reader *bufio.Reader) bool {
	endString, err := reader.Peek(6)
	return reader.Buffered() > 0 && err == nil && string(endString) != "SPLICE"
}

func trimVersion(rawVersion [11]byte) string {
	return strings.TrimRight(string(rawVersion[:]), string([]byte{0x00}))
}
