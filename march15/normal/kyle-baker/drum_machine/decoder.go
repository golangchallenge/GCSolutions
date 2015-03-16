package drum

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

// Pattern is the high level representation of the drum pattern contained
// in a .splice file.
type Pattern struct {
	Version []byte
	Size    uint8
	Tempo   float32
	Tracks  []Track
}

// String representation of a Pattern
func (p *Pattern) String() string {
	str := fmt.Sprintf("Saved with HW Version: %s\n", p.Version)
	str += fmt.Sprintf("Tempo: %g\n", p.Tempo)
	for _, track := range p.Tracks {
		str += fmt.Sprintf("%s", &track)
	}
	return str
}

// Track holds data about a single track
type Track struct {
	Name  []byte
	Id    uint8
	Beats [16]bool
}

// String representation of a Track
func (t *Track) String() string {
	str := fmt.Sprintf("(%d) %s\t", t.Id, t.Name)
	for i, beat := range t.Beats {
		if i%4 == 0 {
			str += "|"
		}
		if beat {
			str += "x"
		} else {
			str += "-"
		}
	}
	str += "|\n"
	return str
}

var (
	InvalidFormat      = errors.New("Invalid or corrupted splice file")
	UnsupportedVersion = errors.New("Splice version not supported")
	TempoError         = errors.New("There was an error reading the tempo")
	TrackError         = errors.New("There was an error reading track data")
	SupportedVersions  = map[string]struct{}{
		"0.808-alpha": struct{}{},
		"0.708-alpha": struct{}{},
		"0.909":       struct{}{},
	}
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {

	r, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	// Create a buffered reader for reading our header data
	hr := bufio.NewReader(r)

	// Read the header string and make sure its SPLICE
	headerString := make([]byte, 6)
	_, err = hr.Read(headerString)
	if err != nil || !bytes.Equal(headerString, []byte("SPLICE")) {
		return nil, InvalidFormat
	}

	// Skip 7 null padding bytes
	skipBytes(hr, 7)

	// Read the size field
	var spliceSize uint8
	err = binary.Read(hr, binary.BigEndian, &spliceSize)
	if err != nil {
		return nil, InvalidFormat
	}

	// Create a new buffered reader from the remaining bytes up until the
	// max size of spliceSize. aka, discard any bytes past the header whose
	// position is greater than spliceSize
	remaining := make([]byte, spliceSize)
	_, err = hr.Read(remaining)
	if err != nil {
		return nil, InvalidFormat
	}
	br := bufio.NewReader(bytes.NewReader(remaining))

	// Read the version and check that its supported
	version, _ := br.ReadBytes(byte(0))
	version = stripLast(version)
	if _, exists := SupportedVersions[string(version)]; !exists {
		return nil, UnsupportedVersion
	}

	// Skip null padding bytes depending on version
	// 808-alpha: 21
	// 909: 27
	// 708-alpha: 21
	if bytes.Equal(version, []byte("0.909")) {
		skipBytes(br, 26)
	} else {
		skipBytes(br, 20)
	}

	// Get the tempo value
	var tempo float32
	err = binary.Read(br, binary.LittleEndian, &tempo)
	if err != nil {
		return nil, TempoError
	}

	tracks := make([]Track, 0)

	// Get each one of our tracks, Reading until either \x0a is reached,
	// or we reach the end of the buffer defined by spliceSize
	for {

		// Get track ID
		trackId, err := br.ReadByte()
		if trackId == byte(0x0a) || err == io.EOF {
			break
		}
		if err != nil {
			return nil, TrackError
		}

		// Get track name
		skipBytes(br, 3)
		var trackNameLen uint8
		err = binary.Read(br, binary.BigEndian, &trackNameLen)
		if err != nil {
			return nil, TrackError
		}
		trackName := make([]byte, trackNameLen)
		_, err = br.Read(trackName)
		if err != nil {
			return nil, TrackError
		}

		// Get track beats
		beatsAsBytes := make([]byte, 16)
		_, err = br.Read(beatsAsBytes)
		if err != nil {
			return nil, TrackError
		}
		var beats [16]bool
		for i, beat := range beatsAsBytes {
			switch beat {
			case byte(0x00):
				beats[i] = false
			case byte(0x01):
				beats[i] = true
			default:
				e := fmt.Sprintf("Invalid beat value on track id %d\n", trackId)
				return nil, errors.New(e)
			}
		}

		track := Track{Id: trackId, Name: trackName, Beats: beats}
		tracks = append(tracks, track)
	}

	p := &Pattern{
		Version: version,
		Size:    spliceSize,
		Tempo:   tempo,
		Tracks:  tracks,
	}

	return p, nil
}

// stripLast returns a byte slice with the last byte of b removed
func stripLast(b []byte) []byte {
	return b[:len(b)-1]
}

// skipBytes will advance the reader, r, n number of bytes
func skipBytes(r *bufio.Reader, n int64) {
	io.CopyN(ioutil.Discard, r, n)
}
