package drum

import (
	"bytes"
	"encoding/binary"
	"io"
)

func readHeader(file io.Reader) (string, error) {
	buf := make([]byte, 6)
	_, err := file.Read(buf)

	if err != nil {
		return "", err
	}

	return string(buf), nil
}

func readContentSize(file io.Reader) (int64, error) {
	var size int64
	err := binary.Read(file, binary.BigEndian, &size)

	if err != nil {
		return 0, err
	}

	return size, nil
}

func readVersion(file io.Reader) (string, error) {
	buf := make([]byte, 32)
	_, err := file.Read(buf)

	if err != nil {
		return "", err
	}

	return string(bytes.Trim(buf, "\x00")), nil
}

func readTempo(file io.Reader) (float32, error) {
	var tempo float32
	err := binary.Read(file, binary.LittleEndian, &tempo)

	if err != nil {
		return 0, err
	}

	return tempo, nil
}

func readTrack(file io.Reader, size *int64) (*Track, error) {
	track := new(Track)

	var id int32
	err := binary.Read(file, binary.LittleEndian, &id)
	if err != nil {
		return nil, err
	}
	track.ID = int(id)
	*size -= 4

	var nameLength int8
	err = binary.Read(file, binary.LittleEndian, &nameLength)
	if err != nil {
		return nil, err
	}
	*size--

	buf := make([]byte, nameLength)
	file.Read(buf)
	track.Name = string(buf)
	*size -= int64(nameLength)

	var steps [16]bool
	for i := 0; i < 16; i++ {
		var buf int8
		err = binary.Read(file, binary.LittleEndian, &buf)
		if err != nil {
			return nil, err
		}

		steps[i] = (buf > 0)
	}
	*size -= 16
	track.Steps = steps

	return track, nil
}
