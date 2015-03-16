package drum

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
)

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	version string
	tempo   float32
	tracks  []Track
}

// Track is the high level representation of the
// track contained in the drum pattern contained in a .splice file.
type Track struct {
	id    int32
	name  []byte
	steps [16]byte
}

var fileIdentifier = []byte("SPLICE")

const tempoLen = 4

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	p := &Pattern{}
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return p, err
	}

	content, err, p = DecodeHeader(content)
	if err != nil {
		return p, err
	}

	buf := bytes.NewReader(content)
	for err == nil {
		track := Track{}
		track, err = DecodeTrack(buf)

		if err != nil {
			return p, nil
		}
		p.tracks = append(p.tracks, track)
	}

	return p, nil
}

// Decodes the header containing the file type, version and tempo
func DecodeHeader(contents []byte) ([]byte, error, *Pattern) {
	p := &Pattern{}
	gap := map[string]int{
		"0.708-alpha": 21,
		"0.808-alpha": 21,
		"0.909":       27,
	}

	rest := contents
	// Every pattern file always starts with the identifier SPLICE
	if bytes.HasPrefix(contents, fileIdentifier) {
		rest = rest[len(fileIdentifier):len(contents)]
	} else {
		return nil, fmt.Errorf("Can't recognize format"), p
	}

	var version string
	version, rest = GetVersion(rest)

	var err error
	rest, p.tempo, err = GetTempo(gap[string(version)], rest)
	if err != nil {
		return nil, fmt.Errorf("Can't recognize format"), p
	}
	p.version = version

	return rest, nil, p
}

// Read the tempo of the track by splicing the byte array.
// Returns the byte array without the bytes containing the tempo.
func GetTempo(gap int, contents []byte) ([]byte, float32, error) {
	var tempo float32
	rest := contents[gap : gap+tempoLen]
	buf := bytes.NewReader(rest)
	err := binary.Read(buf, binary.LittleEndian, &tempo)
	contents = contents[gap+tempoLen:]
	return contents, tempo, err
}

// Decoding the track with id, name and steps
func DecodeTrack(buf io.Reader) (track Track, err error) {
	err = binary.Read(buf, binary.LittleEndian, &track.id)

	if err != nil {
		return track, err
	}

	var nameLen uint8
	err = binary.Read(buf, binary.LittleEndian, &nameLen)
	if err != nil {
		return track, err
	}

	track.name = make([]byte, nameLen)
	err = binary.Read(buf, binary.LittleEndian, &track.name)
	if err != nil {
		return track, err
	}

	err = binary.Read(buf, binary.LittleEndian, &track.steps)
	return track, err
}

// Returns the version of used saving this pattern
func GetVersion(contents []byte) (version string, rest []byte) {
	start := false
	i := 0
	vStart := 0
	for _, b := range contents {
		if b == 0 {
			if start == true {
				version = string(contents[vStart:i])
				rest = contents[i:]
				break
			}
		} else {
			if start == false {
				vStart = i + 1
			}
			start = true
		}
		i++
	}

	return version, rest
}

// String representation of the pattern
func (p Pattern) String() string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("Saved with HW Version: %s\nTempo: %v\n", p.version, p.tempo))
	for _, track := range p.tracks {
		buf.WriteString(track.String())
	}
	return buf.String()
}

// String representation of the track
func (track Track) String() string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("(%d) %s\t", track.id, string(track.name)))
	for i := 0; i < len(track.steps); i++ {
		if i%4 == 0 {
			buf.WriteString("|")
		}
		if track.steps[i] == 0 {
			buf.WriteString("-")
		} else {
			buf.WriteString("x")
		}
	}
	buf.WriteString("|\n")
	return buf.String()
}
