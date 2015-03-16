package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

const (
	headerLength     = 6
	trackSteps       = 16
	versionMaxLength = 32

	spliceHeader = "SPLICE"
)

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Version string
	Tempo   float32
	Tracks  []Track

	lastErr error
	buffer  io.ReadSeeker
}

// Track is the representation of a single track in the pattern.
type Track struct {
	ID    byte
	Name  string
	Steps [trackSteps]byte
}

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	p := &Pattern{}
	err = p.UnmarshalBinary(data)
	if err != nil {
		return nil, err
	}

	return p, nil
}

// UnmarshalBinary loads pattern attributes from data.
func (p *Pattern) UnmarshalBinary(data []byte) error {
	p.buffer = bytes.NewReader(data)

	p.checkHeader()

	length := p.readLength()
	maxOffset := p.currentOffset() + length

	p.readVersion()
	p.readTempo()

	for p.currentOffset() < maxOffset {
		p.readTrack()
	}

	return p.lastErr
}

// currentOffset returns current offset of internal buffer.
func (p *Pattern) currentOffset() uint64 {
	if p.lastErr != nil {
		return 0
	}

	offset, err := p.buffer.Seek(0, os.SEEK_CUR)
	if err != nil {
		p.lastErr = err
	}

	return uint64(offset)
}

// read reads binary data from internal buffer into data.
func (p *Pattern) read(data interface{}) {
	var order binary.ByteOrder

	switch data.(type) {
	case *float32, *float64, *[]float32, *[]float64:
		order = binary.LittleEndian
	default:
		order = binary.BigEndian
	}

	err := binary.Read(p.buffer, order, data)
	if err != nil {
		p.lastErr = err
	}
}

// checkHeader reads header from internal buffer and checks if it is correct.
func (p *Pattern) checkHeader() {
	if p.lastErr != nil {
		return
	}

	header := make([]byte, headerLength)
	p.read(header)

	if !bytes.Equal(header, []byte(spliceHeader)) {
		p.lastErr = errors.New("invalid header")
	}
}

// readLength reads content's length from internal buffer and returns it.
func (p *Pattern) readLength() uint64 {
	if p.lastErr != nil {
		return 0
	}

	var length uint64
	p.read(&length)

	return length
}

// readVersion reads pattern version from internal buffer.
func (p *Pattern) readVersion() {
	if p.lastErr != nil {
		return
	}

	version := make([]byte, versionMaxLength)
	p.read(version)

	// Save version up to null byte
	n := bytes.Index(version, []byte{0})
	p.Version = string(version[:n])
}

// readTempo reads pattern tempo from internal buffer.
func (p *Pattern) readTempo() {
	if p.lastErr != nil {
		return
	}

	p.read(&p.Tempo)
}

// readTrack reads single track from internal buffer.
func (p *Pattern) readTrack() {
	if p.lastErr != nil {
		return
	}

	track := Track{}

	p.read(&track.ID)

	// Name's length
	var length uint32
	p.read(&length)

	// Track's name
	name := make([]byte, length)
	p.read(name)
	track.Name = string(name)

	// Track's steps
	steps := make([]byte, trackSteps)
	p.read(steps)
	copy(track.Steps[:], steps)

	p.Tracks = append(p.Tracks, track)
}

func (p *Pattern) String() string {
	var buffer bytes.Buffer

	buffer.WriteString(fmt.Sprintf("Saved with HW Version: %s\n", p.Version))
	buffer.WriteString(fmt.Sprintf("Tempo: %v\n", p.Tempo))

	for _, track := range p.Tracks {
		buffer.WriteString(fmt.Sprintf("(%d) %s\t", track.ID, track.Name))

		for i, step := range track.Steps {
			if i%4 == 0 {
				buffer.WriteString("|")
			}

			if step == 1 {
				buffer.WriteString("x")
			} else {
				buffer.WriteString("-")
			}
		}

		buffer.WriteString("|\n")
	}

	return buffer.String()
}
