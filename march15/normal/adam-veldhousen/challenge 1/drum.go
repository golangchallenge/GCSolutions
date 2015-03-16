// Package drum is supposed to implement the decoding of .splice drum machine files.
// See golang-challenge.com/go-challenge1/ for more information
package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
)

func parse(data []byte, p *Pattern) error {
	dataReader := bytes.NewReader(data)

	encodedDataSize := len(data)

	hwVersion, hwVersionErr := readHardwareVersion(dataReader)
	if hwVersionErr != nil {
		return hwVersionErr
	}
	p.Version = hwVersion

	tempo, readTempErr := readTempo(dataReader)
	if readTempErr != nil {
		return readTempErr
	}
	p.Tempo = tempo

	tracks, readTracksErr := readTracks(dataReader, encodedDataSize)
	if readTracksErr != nil {
		return readTracksErr
	}
	p.Tracks = tracks

	return nil
}

func readTracks(reader *bytes.Reader, encodedDataSize int) ([]Track, error) {
	var tracks []Track

	position := encodedDataSize - reader.Len()

	for position < encodedDataSize {
		var id int32
		binary.Read(reader, binary.LittleEndian, &id)

		channelNameSize, _ := reader.ReadByte()
		channelBytes := make([]byte, channelNameSize)
		_, err := reader.Read(channelBytes)
		if err != nil {
			return []Track{}, errors.New("Could not read Track name with id " + string(id))
		}

		pattern := make([]uint32, 4)
		patternReadErr := binary.Read(reader, binary.LittleEndian, &pattern)
		if patternReadErr != nil {
			return []Track{}, errors.New("Could not read Track step with id " + string(id))
		}

		tracks = append(tracks, Track{
			id,
			string(channelBytes),
			pattern})

		position += int(21) + int(channelNameSize)
	}

	return tracks, nil
}

func readTempo(reader *bytes.Reader) (float32, error) {
	var tempo float32
	err := binary.Read(reader, binary.LittleEndian, &tempo)
	if err != nil {
		return 0.0, errors.New("Could not read tempo")
	}
	return tempo, nil
}

func readHardwareVersion(reader *bytes.Reader) (string, error) {
	versionBytes := make([]byte, 32)
	_, versionReadError := reader.Read(versionBytes)
	if versionReadError != nil {
		return "", versionReadError
	}

	versionString := string(bytes.Trim(versionBytes, "\x00"))
	if versionString == "" {
		return "", errors.New("The file version is incorrect.")
	}

	return versionString, nil
}
