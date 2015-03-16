package drum

import (
	"encoding/binary"
	"os"
)

// EncodePattern creates a new splice file at the passed in path location.
// Returns an error if there is an issue writing the file to disk.
func EncodePattern(pattern *Pattern, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}

	err = binary.Write(file, binary.BigEndian, []byte(SpliceFileHeader))
	if err != nil {
		return err
	}

	var size uint64
	size = uint64(VersionSize + TempoSize)
	for _, track := range pattern.Tracks {
		size += uint64(TrackIDSize + 4 + len(track.Name) + StepSequenceSize)
	}

	err = binary.Write(file, binary.BigEndian, &size)
	if err != nil {
		return err
	}

	version := make([]byte, VersionSize)
	tmp := []byte(pattern.Version)
	for i := range tmp {
		version[i] = tmp[i]
	}

	err = binary.Write(file, binary.BigEndian, version)
	if err != nil {
		return err
	}

	err = binary.Write(file, binary.LittleEndian, &pattern.Tempo)
	if err != nil {
		return err
	}

	for _, track := range pattern.Tracks {
		err = binary.Write(file, binary.BigEndian, &track.ID)
		if err != nil {
			return err
		}

		var trackNameLen uint32
		trackNameLen = uint32(len(track.Name))
		err = binary.Write(file, binary.BigEndian, &trackNameLen)
		if err != nil {
			return err
		}

		trackName := []byte(track.Name)
		err = binary.Write(file, binary.BigEndian, trackName)
		if err != nil {
			return err
		}

		for _, step := range track.StepSequence.Steps {
			if step == byte(0) {
				err = binary.Write(file, binary.BigEndian, byte(0))
			} else if step == byte(1) {
				err = binary.Write(file, binary.BigEndian, byte(1))
			}

			if err != nil {
				return err
			}
		}

	}

	return nil
}
