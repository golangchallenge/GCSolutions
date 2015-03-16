package drum

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"os"
)

// ErrorInvalidData is for error occured during decoding of .splice file
var ErrorInvalidData error = errors.New("Invalid Data")

// ErrorDuplicateHeader is returned when a Duplicate of header `SPLICE`
// is encountered during decoding of .splice file
var ErrorDuplicateHeader error = errors.New("Duplicate Header")
var Header string = "SPLICE"

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	p := &Pattern{}
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	// Discard the useless start bytes
	reader := bufio.NewReader(f)
	if n, err := reader.Read(make([]byte, 14)); n != 14 || err != nil {
		return nil, err
	}
	// Read till next null byte to retrieve version
	version, _ := reader.ReadBytes(byte(0))
	// Store version and exclude the last null byte
	p.Version = string(version[:len(version)-1])

	// Skip all null bytes till tempo
	for b, err := reader.Peek(1); ; b, err = reader.Peek(1) {
		if err != nil {
			return nil, ErrorInvalidData
		}
		// If null byte skip and continue
		if b[0] == byte(0) {
			_, err = reader.ReadByte()
			if err != nil {
				return nil, ErrorInvalidData
			}
		} else {
			break
		}
	}
	// Read till next null byte to retrieve tempo
	tempo, err := reader.ReadBytes(byte(0))
	if err != nil {
		return nil, err
	}

	// Check if we have 0 id and stopped at null byte for id.
	// Store id and note with idAvail if reader has passed position
	// of id of first track
	b, err := reader.Peek(3)
	idAvail := false
	var trackId byte
	if b[2] != 0 {
		idAvail = true
		trackId = tempo[len(tempo)-2]
		tempo = tempo[:len(tempo)-2]
		// Skip to name of track
		if n, err := reader.Read(make([]byte, 3)); n != 3 || err != nil {
			return nil, ErrorInvalidData
		}
	} else {
		tempo = tempo[:len(tempo)-1]
		if err = reader.UnreadByte(); err != nil {
			return nil, err
		}
	}

	// Prepend tempo with required 0 bytes to enable successful binary read
	t := make([]byte, 4-len(tempo))
	for i := 0; i < len(t); i++ {
		t[i] = byte(0)
	}
	tempo = append(t, tempo...)
	buf := bytes.NewBuffer(tempo)
	err = binary.Read(buf, binary.LittleEndian, &p.Tempo)
	if err != nil {
		return nil, err
	}

	// If reader has skipped id pos but id has been stored,
	// read first track
	if idAvail {
		// retrieve first track
		track, err := extractTrack(reader, true, trackId)
		if err != nil {
			return nil, err
		}
		p.Tracks = append(p.Tracks, track)
	}

	//Read all tracks
	for {
		track, err := extractTrack(reader, false, 0)
		if err != nil {
			if len(p.Tracks) == 0 {
				return nil, err
			}
			break
		}
		p.Tracks = append(p.Tracks, track)
	}

	return p, nil
}

// extractTrack decodes a Track from the reader. It returns a usable Track and error if an error occured.
func extractTrack(reader *bufio.Reader, useId bool, id byte) (Track, error) {
	track := Track{}
	// Check for duplicate header
	if b, _ := reader.Peek(6); string(b) == Header {
		return track, ErrorDuplicateHeader
	}
	if useId {
		track.Id = id
	} else {
		// If id is not supplied, read id
		id, err := reader.ReadByte()
		if err != nil {
			return track, err
		}
		track.Id = id
		// Skip till the name of the track
		if n, err := reader.Read(make([]byte, 4)); n != 4 || err != nil {
			return track, ErrorInvalidData
		}
	}

	// Extract track name.
	// Peek and append next byte until a 1 or 0 is encountered.
	var name []byte
	for b, err := reader.Peek(1); ; b, err = reader.Peek(1) {
		if err != nil {
			return track, err
		}
		if len(b) < 1 {
			return track, ErrorInvalidData
		}
		if b[0] == byte(0) || b[0] == byte(1) {
			break
		} else {
			c, err := reader.ReadByte()
			if err != nil {
				return track, err
			}
			name = append(name, c)
		}
	}
	track.Name = string(name)
	// read the 16 steps
	for i := 0; i < 16; i++ {
		c, err := reader.ReadByte()
		if err != nil {
			return track, err
		}
		track.Steps[i] = c > byte(0)
	}
	return track, nil
}
