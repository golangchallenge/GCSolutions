package drum

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	p := &Pattern{}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	err = binary.Read(f, binary.LittleEndian, &p.Header)
	if err != nil {
		return nil, err
	}
	for {
		track, err := readTrack(newTrackReader(f))
		if err != nil {
			break
		}
		p.Tracks = append(p.Tracks, track)
	}
	return p, nil
}

type trackReader struct {
	reader io.Reader
	err    error
}

// newTrackReader helper structure to simplify error handling for track parsing
func newTrackReader(reader io.Reader) *trackReader {
	return &trackReader{reader: reader}
}

func (t *trackReader) read(data interface{}) {
	if t.err != nil {
		return
	}
	t.err = binary.Read(t.reader, binary.LittleEndian, data)
}

// readTrack reads one track from the splice file
func readTrack(reader *trackReader) (*Track, error) {
	track := &Track{}
	reader.read(&track.Id)
	reader.read(&track.Name.Length)
	track.Name.Bytes = make([]byte, track.Name.Length)
	reader.read(&track.Name.Bytes)
	reader.read(&track.Measure)
	return track, reader.err
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Header SpliceHeader
	Tracks []*Track
}

// SpliceHeader represents the header of the splice file
type SpliceHeader struct {
	_       [14]byte
	Version [32]byte
	Tempo   float32
}

// Track represents a track from the splice file
type Track struct {
	Id      int32
	Name    TrackName
	Measure [4]QuarterNote
}

type TrackName struct {
	Length byte
	Bytes  []byte
}

type QuarterNote struct {
	Steps [4]byte
}

func (p *Pattern) String() string {
	buffer := bytes.NewBuffer([]byte{})
	p.Header.WriteString(buffer)
	for _, t := range p.Tracks {
		t.WriteString(buffer)
	}
	return buffer.String()
}

func (h *SpliceHeader) WriteString(buffer *bytes.Buffer) {
	headerLength := bytes.IndexByte(h.Version[:], 0)
	buffer.WriteString(fmt.Sprintf("Saved with HW Version: %s\nTempo: %g\n", h.Version[:headerLength], h.Tempo))
}

func (t *Track) WriteString(buffer *bytes.Buffer) {
	buffer.WriteString(fmt.Sprintf("(%v) %s\t", t.Id, t.Name.Bytes))
	for _, m := range t.Measure {
		buffer.WriteString("|")
		m.WriteString(buffer)
	}
	buffer.WriteString("|\n")
}

func (q QuarterNote) WriteString(buffer *bytes.Buffer) {
	for i := 0; i < 4; i++ {
		if q.Steps[i] > 0 {
			buffer.WriteString("x")
		} else {
			buffer.WriteString("-")
		}
	}
}
