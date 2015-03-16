package drum

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
)

// .splice files have header + tracks content
//
// header
// ------
// 6 bytes  - SPLICE string
// 7 bytes  - unknown bytes - 0 bytes
// 1 byte   - content length - number of bytes containing tracks
// 32 bytes - place for version string
// 4 bytes  - tempo (float32)
//
// tracks
// -------
// 4 bytes  - id uint32
// 1 byte   - length of name
// {len} bytes - name
// 16 bytes  - quarters

const (
	// 6 bytes - SPLICE string
	SPLICE_HEADER_FILETYPE = 6
	// 7 bytes - unknown
	SPLICE_HEADER_UNKNOWN = 7
	// 1 byte is marking tracks length
	SPLICE_HEADER_CONTENT = 1
	// 32 bytes - file version + padding
	SPLICE_HEADER_NAME = 32
	// Total header size
	SPLICE_HEADER_SIZE = (SPLICE_HEADER_FILETYPE + SPLICE_HEADER_UNKNOWN + SPLICE_HEADER_CONTENT + SPLICE_HEADER_NAME)
)

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Version string
	Tempo   float32
	Tracks  []*Track
}

// Track is represented by id, name and quarters
type Track struct {
	Id       int32
	Name     string
	Quarters []*Quarter
}

// Each quarter has 4 sixteenths
type Quarter struct {
	Sixteenths []*Sixteenth
}

type Sixteenth byte

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	p := &Pattern{}

	// Open file for reading
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Reusable reader variable
	var reader io.Reader

	// Use buffered reader from now
	reader = bufio.NewReader(file)

	// Read header at once
	header := make([]byte, SPLICE_HEADER_SIZE)
	_, err = reader.Read(header)
	if err != nil {
		return nil, err
	}

	// Make sure we're dealing with correct file type
	splice := header[0:SPLICE_HEADER_FILETYPE]
	if string(splice) != "SPLICE" {
		return nil, errors.New("Not a valid splice file: " + path)
	}

	// Determine content length
	contentLength := uint8(header[SPLICE_HEADER_FILETYPE+SPLICE_HEADER_UNKNOWN : SPLICE_HEADER_FILETYPE+SPLICE_HEADER_UNKNOWN+1][0])

	// Get version string by reading full full
	name := header[SPLICE_HEADER_FILETYPE+SPLICE_HEADER_UNKNOWN+SPLICE_HEADER_CONTENT:]
	p.Version = string(bytes.Trim(name, string([]byte{0})))

	// Read tempo
	binary.Read(reader, binary.LittleEndian, &p.Tempo)

	contentBuffer := make([]byte, contentLength-SPLICE_HEADER_NAME)
	reader.Read(contentBuffer)
	reader = bytes.NewReader(contentBuffer)

	// Read each track
	for {
		track, err := ReadTrack(reader)
		if err != nil {
			break
		}
		p.Tracks = append(p.Tracks, track)
	}

	return p, nil
}

// Read single track from reader
// id : 4 bytes | len : 1 byte | name : {len} bytes | 4 x quarter
func ReadTrack(reader io.Reader) (*Track, error) {
	track := &Track{}

	var err error

	// Load track id
	err = binary.Read(reader, binary.LittleEndian, &track.Id)
	if err != nil {
		return nil, err
	}

	// Get length of track name
	var nameLength uint8
	err = binary.Read(reader, binary.LittleEndian, &nameLength)
	if err != nil {
		return nil, err
	}

	// Read the name of track
	name := make([]byte, nameLength)
	_, err = reader.Read(name)
	if err != nil {
		return nil, err
	}
	track.Name = string(name)

	// Quarter buffer
	data := make([]byte, 4)

	// Read 4 quarters
	for i := 0; i < 4; i++ {
		_, err = reader.Read(data)
		quarter, err := NewQuarter(data)
		if err != nil {
			return nil, err
		}
		track.Quarters = append(track.Quarters, quarter)
	}

	return track, nil
}

// Create new quarter from sixteenths
func NewQuarter(data []byte) (*Quarter, error) {
	if len(data) != 4 {
		return nil, errors.New("Quarter must contain 4 sixteenths")
	}

	quarter := &Quarter{}
	for i := 0; i < 4; i++ {
		sixteenth := Sixteenth(data[i])
		quarter.Sixteenths = append(quarter.Sixteenths, &sixteenth)
	}

	return quarter, nil
}

func (p *Pattern) String() string {
	out := fmt.Sprintf("Saved with HW Version: %v\n", p.Version)
	out += fmt.Sprintf("Tempo: %v\n", p.Tempo)
	for t := range p.Tracks {
		out += p.Tracks[t].String()
	}
	return out
}

func (t *Track) String() string {
	out := fmt.Sprintf("(%v) %v\t|", t.Id, t.Name)
	for i := range t.Quarters {
		out += fmt.Sprintf("%v|", t.Quarters[i])
	}
	out += "\n"
	return out
}

func (q *Quarter) String() string {
	var out string
	for i := range q.Sixteenths {
		out += fmt.Sprintf("%v", q.Sixteenths[i])
	}
	return out
}

func (s *Sixteenth) String() string {
	if byte(*s) == byte(1) {
		return "x"
	} else {
		return "-"
	}
}
