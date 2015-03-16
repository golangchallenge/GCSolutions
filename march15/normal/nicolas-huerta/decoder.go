package drum

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"strconv"
)

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Header      Header  // Header for this file
	Tracks      []Track // Track data slice
	StringValue string  // This is for convenience for embedding the Pattern in text format required by decoder_test.go.
}

// A Header represents the header for a splice file.
type Header struct {
	FormatName string  // File format name: SPLICE, max string length 10
	DataSize   int32   // Size of next data chunk containing valid data
	Version    string  // HW Version that created the file, max string length 32
	Tempo      float32 // Tempo value
}

// A Track contains data for a single track contained in a splice file
type Track struct {
	Id         byte   // Track id 0-255
	NameLength int32  // Length of the track name string
	Name       string // Track name string
	PatternStr string // Track pattern as a string: |x-x-|x-x-|x-x-|x-x-| extracted from  a 16 byte slice
}

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	spliceFilePattern, err := DecodeSpliceFileToPattern(data)
	if err != nil {
		return nil, err
	}

	// StringValue contains the string required by the decoder_test
	spliceFilePattern.StringValue = PatternToString(spliceFilePattern)

	return spliceFilePattern, nil
}

// DecodeHeader decodes data in rawBytes and returns a new Header that
// represents the decoded data file header.
func DecodeHeader(rawBytes []byte) (header *Header, err error) {
	header = &Header{}

	header.FormatName = ParseCString(rawBytes[:kMaxDataFormatLength])
	header.DataSize, err = ParseInt32BE(rawBytes[kDataSizeOffset:])
	if err != nil {
		return nil, err
	}

	header.Version = ParseCString(rawBytes[kVersionOffset:])
	header.Tempo, err = ParseFloat32LE(rawBytes[kTempoOffset:])
	if err != nil {
		return nil, err
	}

	return header, err
}

// DecodeTracks uses the data in header and the input slice rawBytes to genereate the tracks slice.
func DecodeTracks(header *Header, rawBytes []byte) (tracks []Track, err error) {

	// currTrackDataOffset is a counter that points to the current piece of data being parsed
	// it is initialized to the starting point of the track sets data
	currTrackDataOffset := kTrackDataOffset

	// Offset of the end of the track data sets
	trackDataSize := kDataSizeOffset + int(header.DataSize)

	// currTrackDataOffset is incremented every time we read data and the for loop iterates over each track data set
	// until we reach trackDataSize
	for currTrackDataOffset < trackDataSize {
		track := &Track{}

		track.Id, err = ParseUInt8BE(rawBytes[currTrackDataOffset:])
		if err != nil {
			return nil, err
		}
		currTrackDataOffset += binary.Size(track.Id)

		track.NameLength, err = ParseInt32BE(rawBytes[currTrackDataOffset:])
		if err != nil {
			return nil, err
		}
		currTrackDataOffset += binary.Size(track.NameLength)

		track.Name = ParseString(rawBytes[currTrackDataOffset:], int(track.NameLength))
		currTrackDataOffset += int(track.NameLength)

		track.PatternStr = ParseTrackPatternAsString(rawBytes[currTrackDataOffset:])
		tracks = append(tracks, *track)

		currTrackDataOffset += kTrackPatternLength
	}

	return tracks, nil
}

// PatternToString returns a string representing the Pattern trackPattern in the format
// required by decoder_test.go
func PatternToString(trackPattern *Pattern) string {
	outputStr := "Saved with HW Version: "
	outputStr += (trackPattern.Header.Version + "\nTempo: ")
	var tempoStr string = fmt.Sprintf("%g", trackPattern.Header.Tempo)
	outputStr += (tempoStr + "\n")
	trackStr := ""
	for i := 0; i < len(trackPattern.Tracks); i++ {
		trackStr += "("
		trackStr += strconv.Itoa(int(trackPattern.Tracks[i].Id))
		trackStr += ") "
		trackStr += trackPattern.Tracks[i].Name
		trackStr += "\t"
		trackStr += trackPattern.Tracks[i].PatternStr
	}
	outputStr += trackStr
	return outputStr
}

// DecodeSpliceFileToPattern decodes the header and track data for a file and returns patternData, which contains
// the decoded data contained in a splice file represented by rawBytes.
func DecodeSpliceFileToPattern(rawBytes []byte) (patternData *Pattern, err error) {
	patternData = &Pattern{}

	header, err := DecodeHeader(rawBytes)

	if err != nil {
		return nil, err
	}

	patternData.Header = *header
	patternData.Tracks, err = DecodeTracks(&patternData.Header, rawBytes)
	if err != nil {
		return nil, err
	}

	return patternData, nil
}
