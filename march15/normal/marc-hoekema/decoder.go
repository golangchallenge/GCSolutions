package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	p := &Pattern{}

	f, err := os.Open(path) // Open the path file for read access.
	check(err)

	// Read the contents of the file into memory.
	fileData, err := ioutil.ReadAll(f)
	check(err)

	// Verify the header of the file.
	err = verifyHeader(fileData)
	check(err)

	// Get the version number.
	p.version, err = getVersionString(fileData)
	check(err)

	// Get the tempo.
	p.tempo = getTempo(fileData)

	// Get the track information.
	p.tracks = getTracks(fileData)

	return p, nil
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	version string
	tempo   float32
	tracks  embeddedTracks
}

/*
 * General layout of a .splice track:
 *  4  byte   id
 *  1  byte   length of the following track name field
 *  x  bytes  track name
 * 16 bytes  steps
 */

// trackData is the representation of an individual track
// pattern contained in one of the tracks from a .splice file.
type trackData struct {
	id    int
	name  string
	steps string
}

// embeddedTracks is a slice of the trackData elements which make up the drum pattern
// contained in a .splice file
type embeddedTracks []trackData

// String provides a pretty-printable representation of the data from a Pattern struct.
func (pattern Pattern) String() string {
	returnString := fmt.Sprintf("Saved with HW Version: %s\nTempo: %g\n", pattern.version, pattern.tempo)
	for _, element := range pattern.tracks {
		returnString = returnString + fmt.Sprintf("(%d) %s\t%s\n", element.id, element.name, element.steps)
	}
	return returnString
}

// check evaluates passed in error conditions, calling panic() when an error has been
func check(e error) error {
	if e != nil {
		return e
	}
	return nil
}

// verifyHeader validates that the passed in fileData[] contains the valid .splice header value ("SPLICE").
func verifyHeader(fileData []byte) error {
	var validHeaderData = "SPLICE" // Valid .splice header value

	fileHeaderValue := string(fileData[0:len(validHeaderData)])

	if fileHeaderValue != validHeaderData {
		err := errors.New("verifyHeader: Not a valid .splice file.")
		return err
	}
	return nil

}

// getVersionString extracts the version string from the passed in fileData[].
func getVersionString(fileData []byte) (string, error) {
	var err error
	versionStringStartingOffset := 14
	versionStringEndingOffset := 48

	fileVersionBytes := fileData[versionStringStartingOffset:versionStringEndingOffset]
	fileVersionString := string(fileVersionBytes[:clen(fileVersionBytes)]) //TODO: Check what error values clen() returns
	if len(fileVersionString) <= 0 {                                       // Sanity check
		err = errors.New("getVersionString: No version string found in .splice file.")
	}
	return fileVersionString, err
}

// getTempo extracts the tempo form the the passed in fileData[].
func getTempo(fileData []byte) float32 {
	var tempo float32
	tempoStartingOffset := 46
	tempoEndingOffset := 50

	buf := bytes.NewReader(fileData[tempoStartingOffset:tempoEndingOffset])
	err := binary.Read(buf, binary.LittleEndian, &tempo)
	check(err)
	return tempo
}

// getTracks extracts the individual track details from the splice data passed in fileData[].
func getTracks(fileData []byte) embeddedTracks {
	var tracks embeddedTracks
	var track trackData
	var trackData []byte
	var lengthOfTrackName int

	trackDataOffset := 50        // Offset from the beginning of a .splice file where the track data begins.
	trackStepsLength := 16       // The 16 bytes that hold the step information for a given track.
	nextTrackOffset := 0         // Holds the offset to the beginning of the next track struct.
	trackNameLengthOffset := 4   // Offset from the beginning of a track entry, to the byte holding the length of the track name.
	lengthOfTrackNameLength := 1 // Length of the field holding the lenght of the track name (1 byte).

	trackData = fileData[trackDataOffset:]
	for {
		track.id = int(trackData[nextTrackOffset])
		lengthOfTrackName = int(trackData[nextTrackOffset+trackNameLengthOffset])

		// If the track name length would not leave enough room for step data,
		// this track will be considered invalid data, and will not be returned.
		if len(fileData)-(nextTrackOffset+trackNameLengthOffset+lengthOfTrackNameLength+lengthOfTrackName) < trackStepsLength {
			break
		}

		// Extract the name of the track
		track.name = string(trackData[nextTrackOffset+trackNameLengthOffset+lengthOfTrackNameLength : (nextTrackOffset + trackNameLengthOffset + lengthOfTrackNameLength + lengthOfTrackName)])

		// Get the string representation of the step information of this track
		track.steps = getSteps(trackData[(nextTrackOffset + trackNameLengthOffset + lengthOfTrackNameLength + lengthOfTrackName):(nextTrackOffset + trackNameLengthOffset + lengthOfTrackNameLength + lengthOfTrackName + trackStepsLength)])

		nextTrackOffset = (nextTrackOffset + trackNameLengthOffset + lengthOfTrackNameLength + lengthOfTrackName + trackStepsLength)

		// Add the extracted track information to the slice of embeddedTracks.
		tracks = append(tracks, track)

		// If the next track offset would be at or past the end of the input file we return here.
		if nextTrackOffset >= len(trackData) {
			break
		}
	}
	return tracks
}

// getSteps returns a string representation of the binary step data for the given stepData[] slice.
func getSteps(stepData []byte) string {
	var steps string
	for index, element := range stepData {
		if index%4 == 0 {
			steps = steps + "|"
		}
		if element == 1 {
			steps = steps + "x"
		} else {
			steps = steps + "-"
		}
	}
	steps = steps + "|"
	return steps
}

// clen returns the length of the data contained in n to the first occurance of NUL(\0).
func clen(n []byte) int {
	for i := 0; i < len(n); i++ {
		if n[i] == 0 {
			return i
		}
	}
	return len(n)
}
