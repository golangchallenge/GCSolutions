package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"os"
)

const (
	lenHeader       = 12
	lenLength       = 2
	lenVersion      = 32
	lenTempo        = 4
	lenTrackId      = 4
	lenTrackNameLen = 1
	lenSteps        = 16
)

var (
	spliceHeader = [lenHeader]byte{'S', 'P', 'L', 'I', 'C', 'E'}
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
// TODO: implement
func DecodeFile(path string) (*Pattern, error) {
	p := &Pattern{}

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if header, err := readBytes(f, lenHeader); err != nil {
		return nil, err
	} else if !bytes.Equal(header, spliceHeader[:]) {
		return nil, errors.New("Invalid file")
	}

	var length int
	if lenBytes, err := readBytes(f, lenLength); err != nil {
		return nil, err
	} else {
		length = int(binary.BigEndian.Uint16(lenBytes))
	}

	if version, err := readBytes(f, lenVersion); err != nil {
		return nil, err
	} else {
		length -= lenVersion
		p.Version = string(bytes.TrimRight(version, "\u0000"))
	}

	if tempo, err := readBytes(f, lenTempo); err != nil {
		return nil, err
	} else {
		length -= lenTempo
		p.Tempo = math.Float32frombits(binary.LittleEndian.Uint32(tempo))
	}

	for {
		if length == 0 {
			return p, nil
		}

		t, n, err := readTrack(f)
		if err != nil {
			return nil, err
		}

		length -= n
		p.Tracks = append(p.Tracks, *t)
	}

	return p, nil
}

func readTrack(f io.Reader) (*Track, int, error) {
	track := Track{}
	bytesRead := 0

	if trackId, err := readBytes(f, lenTrackId); err != nil {
		return nil, 0, err
	} else {
		bytesRead += lenTrackId
		track.Id = binary.LittleEndian.Uint32(trackId)
	}

	if nameLen, err := readBytes(f, lenTrackNameLen); err != nil {
		return nil, 0, err
	} else {
		bytesRead += lenTrackNameLen

		if name, err := readBytes(f, int(nameLen[0])); err != nil {
			return nil, 0, err
		} else {
			bytesRead += len(name)
			track.Name = string(name)
		}
	}

	if steps, err := readBytes(f, lenSteps); err != nil {
		return nil, 0, err
	} else {
		bytesRead += lenSteps
		track.Steps = steps
	}

	return &track, bytesRead, nil
}

func readBytes(r io.Reader, n int) ([]byte, error) {
	bytes := make([]byte, n)
	if _, err := io.ReadFull(r, bytes); err != nil {
		return nil, err
	}

	return bytes, nil
}
