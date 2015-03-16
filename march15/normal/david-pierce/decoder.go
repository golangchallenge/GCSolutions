package drum

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"os"
)

const (
	tOffset  int = 50 // Bytes offset where track data starts.
	tPadding int = 3  // Number of empty bytes after the track sample ID.
	pLen     int = 16 // Number of bytes used to store the track pattern.
	hwLen    int = 32 // Number of bytes used to store the hardware version
	tLenPos  int = 13 // Position of byte representing tracks byte length
)

// Pattern is the high level representation of the drum pattern contained
// in a .splice file.
type Pattern struct {
	HWVersion string
	BPM       float32
	Tracks    []Track
}

// String returns multiple lines describing the pattern's
// metadata and the beat pattern for each of its tracks.
func (p Pattern) String() string {
	buffer := bytes.NewBufferString("")

	fmt.Fprint(buffer, fmt.Sprintf("Saved with HW Version: %s\n", p.HWVersion))
	fmt.Fprint(buffer, fmt.Sprintf("Tempo: %v\n", p.BPM))
	for _, track := range p.Tracks {
		fmt.Fprint(buffer, fmt.Sprintf("%v\n", track))
	}

	return buffer.String()
}

// Track is a line representing the sample and which beats the sample
// should be played on.
type Track struct {
	SampleID   int
	SampleName string
	Pattern    []byte
}

// String returns the track ID, name, and beat pattern.
func (t Track) String() string {
	prelude := fmt.Sprintf("(%d) %s\t", t.SampleID, t.SampleName)

	bytesToSteps := func(steps []byte) string {
		var buf bytes.Buffer
		for _, b := range steps {
			if b == 0x00 {
				buf.WriteString("-")
			} else {
				buf.WriteString("x")
			}
		}

		return buf.String()
	}

	track := fmt.Sprintf("|%s|%s|%s|%s|",
		bytesToSteps(t.Pattern[0:4]),
		bytesToSteps(t.Pattern[4:8]),
		bytesToSteps(t.Pattern[8:12]),
		bytesToSteps(t.Pattern[12:16]),
	)

	return prelude + track
}

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	p := &Pattern{}

	f, err := os.Open(path)
	if err != nil {
		return p, err
	}

	scn := bufio.NewScanner(f)
	scn.Scan()
	b := scn.Bytes()

	err = scn.Err()
	if err != nil {
		return p, err
	}

	hw, err := readHwVersion(b[14:46])
	if err != nil {
		return p, err
	}
	p.HWVersion = hw

	bpm, err := readBPM(b[46:50])
	if err != nil {
		return p, err
	}
	p.BPM = bpm

	// Derive track length by finding number of bytes to read stored in
	// unsigned int at position `tLenPos`, and then remove bytes describing
	// pattern to determine start and end of track bytes
	tEnd := tOffset + int(uint16(b[tLenPos])) - 36
	tracks, err := readTracks(b[tOffset:tEnd])
	if err != nil {
		return p, err
	}
	p.Tracks = tracks

	return p, nil
}

// readHwVersion extracts the HWVersion from the input data.
func readHwVersion(data []byte) (string, error) {
	if len(data) != hwLen {
		return "", fmt.Errorf("wrong amount of data to parse hardware version")
	}

	var str []byte
	for _, b := range data {
		if b != 0x00 {
			str = append(str, b)
		}
	}

	return string(str), nil
}

// readBPM extracts the BPM bytes and converts them to a float32 integer.
func readBPM(data []byte) (float32, error) {
	if len(data) != 4 {
		return -1, fmt.Errorf("wrong amount of data to parse bpm")
	}

	bits := binary.LittleEndian.Uint32(data)
	return math.Float32frombits(bits), nil
}

// readTracks takes in the byte array representing the data
// with the leading pattern data removed. The first track starts
// at byte 0 of the input byte slice.
//
// Since each track is variable length, it loops its way through the
// data and converts each section to track it represents as it goes.
//
// Extracted tracks are added to a results array and the values returned.
func readTracks(data []byte) ([]Track, error) {
	var tracks []Track
	pos := 0

	for {
		tData := data[pos:]
		if len(tData) == 0 {
			break
		} else {
			t, read, err := readTrack(data[pos:])
			pos += read
			if err != nil {
				// Track read error. Bail early
				return tracks, err
			}
			tracks = append(tracks, t)
		}
	}

	return tracks, nil
}

// readTrack takes in a byte array and seeks through to extract
// one Track object. Note, the track extracted may represent a leading
// subset of the entire input.
//
// It returns the number of bytes used to store the Track so that
// the following Track position can be calculated if one exists.
func readTrack(data []byte) (Track, int, error) {
	t := Track{}
	read := 0
	dl := len(data)

	if dl < 2+tPadding {
		return t, read, fmt.Errorf("insufficient data to read track metadata")
	}

	// Get Sample ID
	t.SampleID = int(data[read])
	read++

	// Skip over padding
	read += tPadding

	// Get Sample Name
	nLen := int(data[read])
	read++
	var name []byte

	if dl < read+nLen {
		return t, read, fmt.Errorf("insufficient data to read track name")
	}
	for _, b := range data[read : read+nLen] {
		name = append(name, b)
	}
	t.SampleName = string(name)
	read += nLen

	// Get Track Data
	if dl < read+pLen {
		return t, read, fmt.Errorf("insufficient data to read track pattern")
	}
	t.Pattern = data[read : read+pLen]
	read += pLen

	return t, read, nil
}
