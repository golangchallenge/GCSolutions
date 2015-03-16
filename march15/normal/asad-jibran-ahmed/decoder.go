package drum

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
)

type KnownLengthSpliceFileStruct struct {
	TypeName      [6]byte
	_             [7]byte
	ContentLength uint8
}

type TrackData struct {
	TrackId         uint16
	TrackNameLength uint8
	TrackName       []byte
	TrackSteps      [4][4]byte
}

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}

	fileStruct := KnownLengthSpliceFileStruct{}

	if err = binary.Read(file, binary.LittleEndian, &fileStruct); err != nil {
		return "", err
	}

	// Read the hardware version string, and trim it to remove zero value bytes
	var hwVersion []byte = make([]byte, 11)
	if err = binary.Read(file, binary.LittleEndian, &hwVersion); err != nil {
		return "", err
	}
	firstZeroByteIndex := bytes.IndexByte(hwVersion, 0x00)
	if firstZeroByteIndex > -1 {
		hwVersion = hwVersion[:firstZeroByteIndex]
	}
	// Discard the next 21 bytes
	var discardedBytes [21]byte
	if err = binary.Read(file, binary.LittleEndian, &discardedBytes); err != nil {
		return "", err
	}

	// Read the tempo data
	var tempo float32
	if err = binary.Read(file, binary.LittleEndian, &tempo); err != nil {
		return "", err
	}

	// Read the tracks data
	tracks := make([]*TrackData, 0)
	// Subtract the length of the fields until after the tempo field from the content length
	remainingContentLength := fileStruct.ContentLength - (11 + 21 + 4)
	for remainingContentLength > 0 {
		trackData, n, err := readTrackData(file)
		if err != nil {
			return "", err
		}
		tracks = append(tracks, trackData)
		remainingContentLength -= n
	}

	var outString string
	outString += fmt.Sprintf("Saved with HW Version: %s\n", hwVersion)
	// Since we need to print the tempo without the leading zeros, we first determine if there are any leading zeros
	_, fractionalPart := math.Modf(float64(tempo))
	if fractionalPart == 0.0 {
		outString += fmt.Sprintf("Tempo: %d\n", int(tempo))
	} else {
		outString += fmt.Sprintf("Tempo: %.1f\n", tempo)
	}
	for _, trackData := range tracks {
		var trackString string
		trackString += fmt.Sprintf("(%d) %s\t|", trackData.TrackId, string(trackData.TrackName))

		for _, quarterNotes := range trackData.TrackSteps {
			for _, note := range quarterNotes {
				if note == 1 {
					trackString += "x"
				} else {
					trackString += "-"
				}
			}
			trackString += "|"
		}

		trackString += "\n"
		outString += trackString
	}

	return outString, nil
}

func readTrackData(data io.Reader) (*TrackData, uint8, error) {
	var trackID uint16
	var discardedBytes uint16
	var trackNameLength uint8

	if err := binary.Read(data, binary.LittleEndian, &trackID); err != nil {
		return nil, 0, err
	}

	if err := binary.Read(data, binary.LittleEndian, &discardedBytes); err != nil {
		return nil, 0, err
	}

	if err := binary.Read(data, binary.LittleEndian, &trackNameLength); err != nil {
		return nil, 0, err
	}

	var trackName []byte = make([]byte, trackNameLength)
	if err := binary.Read(data, binary.LittleEndian, &trackName); err != nil {
		return nil, 0, err
	}

	var trackSteps [4][4]byte
	if err := binary.Read(data, binary.LittleEndian, &trackSteps); err != nil {
		return nil, 0, err
	}

	totalBytesRead := 2 + 2 + 1 + trackNameLength + 16
	return &TrackData{trackID, trackNameLength, trackName, trackSteps}, totalBytesRead, nil
}
