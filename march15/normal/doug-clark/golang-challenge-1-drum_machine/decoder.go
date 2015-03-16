package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
	"math"
)

// patternHeader is the valid header of the splice file format
var patternHeader = []byte{0x53, 0x50, 0x4c, 0x49, 0x43, 0x45, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	// read the file -- decode data
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return Decode(b)
}

// Decode decodes a byte slice into the drum Pattern
func Decode(b []byte) (*Pattern, error) {
	buf := bufferHelper{Buffer: bytes.NewBuffer(b), Error: nil}

	// confirm SPLICE header [0-12], 13 bytes
	if !bytes.Equal(buf.Next(13), patternHeader) {
		return nil, errors.New("Invalid file header")
	}

	p := &Pattern{}

	// [13] file content length, 1 byte
	remainingBytes := buf.Next1AsInt("content length", nil)

	// [14-46] HWVersion, 32 bytes
	p.HWVersion = buf.NextXAsString("HWVersion", 32, &remainingBytes)

	// [47-50] Tempo, 4 bytes float32 LittleEndian
	p.Tempo = buf.Next4AsFloat("tempo", &remainingBytes)

	if buf.Error != nil {
		return nil, buf.Error
	}

	// [51+] Repeating tracks
	/////////////
	for remainingBytes > 0 {
		t := Track{}

		// [0-3] track ID 4 bytes
		t.ID = buf.Next4AsUInt32("track ID", &remainingBytes)

		// [4] len of track name 1 byte
		trackNameLen := buf.Next1AsInt("track name length", &remainingBytes)

		// [5-X] track name, var bytes
		t.Name = buf.NextXAsString("track name", trackNameLen, &remainingBytes)

		// 16 bytes, 1 per 1/4 note, 0x00 is off, 0x01 is on
		t.Steps = buf.NextX("track", 16, &remainingBytes)

		if buf.Error != nil {
			return nil, buf.Error
		}

		p.Tracks = append(p.Tracks, t)
	}

	return p, nil
}

// wrap a buffer with util methods and cached errors
type bufferHelper struct {
	*bytes.Buffer

	// Cached error from our "Next" calls
	Error error
}

// NextX returns the next X bytes.  It confirms the bytes were in the buffer and decrements the remainingBytes
func (b *bufferHelper) NextX(fieldName string, x int, remainingBytes *int) []byte {
	if b.Error != nil {
		return nil
	}

	ret := b.Next(x)
	if len(ret) != x {
		b.Error = fmt.Errorf("Invalid file format, missing or invalid %v", fieldName)
	}

	if remainingBytes != nil {
		*remainingBytes -= len(ret)
	}

	return ret
}

func (b *bufferHelper) Next1AsInt(fieldName string, remainingBytes *int) int {
	bInt := b.NextX(fieldName, 1, remainingBytes)
	if b.Error != nil {
		return 0
	}

	return int(bInt[0])
}

func (b *bufferHelper) NextXAsString(fieldName string, x int, remainingBytes *int) string {
	bString := b.NextX(fieldName, x, remainingBytes)
	if b.Error != nil {
		return ""
	}

	// trim nulls from end
	return string(bytes.TrimRightFunc(bString, func(r rune) bool {
		return r == 0x0
	}))
}

func (b *bufferHelper) Next4AsUInt32(fieldName string, remainingBytes *int) uint32 {
	bInt := b.NextX(fieldName, 4, remainingBytes)
	if b.Error != nil {
		return 0
	}

	return binary.LittleEndian.Uint32(bInt)
}

func (b *bufferHelper) Next4AsFloat(fieldName string, remainingBytes *int) float32 {
	bFloat := b.NextX(fieldName, 4, remainingBytes)
	if b.Error != nil {
		return 0
	}

	return math.Float32frombits(binary.LittleEndian.Uint32(bFloat))
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	HWVersion string
	Tempo     float32
	Tracks    []Track
}

// Track represents a single instrument track within a Pattern
type Track struct {
	ID    uint32
	Name  string
	Steps []byte
}

func (p Pattern) String() string {
	tracks := ""
	for _, c := range p.Tracks {
		tracks += c.String() + "\n"
	}

	return fmt.Sprintf("Saved with HW Version: %v\nTempo: %v\n%v", p.HWVersion, p.Tempo, tracks)
}

func (c Track) String() string {
	return fmt.Sprintf("(%v) %v\t%v", c.ID, c.Name, stepsString(c.Steps))
}

func stepsString(steps []byte) string {
	//every 4th item add a pipe
	buf := &bytes.Buffer{}
	for i := 0; i < len(steps); i++ {
		if i%4 == 0 {
			buf.WriteString("|")
		}

		if steps[i] == 0 {
			buf.WriteString("-")
		} else {
			buf.WriteString("x")
		}
	}
	// cap it off
	buf.WriteString("|")

	return buf.String()
}
