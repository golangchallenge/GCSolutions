package drum

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
	"strings"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	p := &Pattern{}
	headerLen := 36 // 50 bytes - 14 initial bytes which are not considered in the length field
	inFile, err := os.Open(path)
	defer inFile.Close()
	err = binary.Read(inFile, binary.LittleEndian, &p.Header)
	if err != nil {
		fmt.Println("There was an error reading the Pattern Header: ", err)
	}
	codedTracks := make([]byte, int(p.Header.Len)-headerLen)
	err = binary.Read(inFile, binary.LittleEndian, &codedTracks)
	if err != nil {
		fmt.Println("Error reading Tracks: ", err)
		return p, err
	}
	p.Tracks, err = DecodeTracks(codedTracks)
	return p, err
}

// DecodeTracks decodes a byte array containing the tracks of a
// drum pattern. Returns a slice of Tracks.
func DecodeTracks(cTracks []byte) ([]Track, error) {
	var tracks []Track
	var track Track
	var pad [3]byte
	var lName int8
	var err error
	trackBuf := bytes.NewReader(cTracks)
	for {
		if err = binary.Read(trackBuf, binary.LittleEndian, &track.ID); err != nil {
			break
		}
		if err = binary.Read(trackBuf, binary.LittleEndian, &pad); err != nil {
			break
		}
		if err = binary.Read(trackBuf, binary.LittleEndian, &lName); err != nil {
			break
		}
		tName := make([]byte, lName)
		if err = binary.Read(trackBuf, binary.LittleEndian, &tName); err != nil {
			break
		}
		track.Name = string(tName)
		if err = binary.Read(trackBuf, binary.LittleEndian, &track.Steps); err != nil {
			break
		}
		tracks = append(tracks, track)
	}
	if err == io.EOF && tracks != nil {
		err = nil
	}
	return tracks, err
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Encoding  [6]byte
	pad       [7]byte
	Len       uint8
	HWVersion [32]byte
	Tempo     float32
	Header    PHeader
	Tracks    []Track
}

type PHeader struct {
	Encoding  [6]byte
	_         [7]byte
	Len       uint8
	HWVersion [32]byte
	Tempo     float32
}

// Track is the high level representation of a drum
// track contained in a .splice file.
type Track struct {
	ID    uint8
	Name  string
	Steps TrackSteps
}

// TrackSteps is the high level representation of the
// track steps in each track.
type TrackSteps [16]byte

// String method implementation for TrackSteps type.
// Used to print tracks using the expected format.
func (steps TrackSteps) String() string {
	var s string
	s = "|"
	for i, b := range steps {
		if b == 0 {
			s += "-"
		} else {
			s += "x"
		}
		if math.Mod(float64(i)+1.0, 4) == 0 {
			s += "|"
		}
	}
	return fmt.Sprintf("%s\n", s)
}

// String method implementation for Pattern struct.
// Used to print a drum patter with the expected format.
func (p Pattern) String() string {
	var tracks string
	var d byte
	hwBuf := bytes.NewBuffer(p.Header.HWVersion[:])
	hwV, err := hwBuf.ReadString(d)
	if err != nil {
		fmt.Println("There was an error Printing p.HWVersion: ", err)
	}
	hwV = strings.TrimSuffix(hwV, "\x00")
	for _, track := range p.Tracks {
		tracks += fmt.Sprintf("%s", track)
	}
	return fmt.Sprintf("Saved with HW Version: %s\nTempo: %v\n%s", hwV, p.Header.Tempo, tracks)
}

// String method for Track struct. Used to
// print tracks with the expected format.
func (t Track) String() string {
	return fmt.Sprintf("(%d) %s\t%s", t.ID, t.Name, t.Steps)
}
