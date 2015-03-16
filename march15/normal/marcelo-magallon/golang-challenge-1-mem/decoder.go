package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io/ioutil"
)

// Format of the file
//
// Byte order: little endian
//
// File header:
//
//	Format: 13 bytes, 0 padded
//	Total data length: 1 byte
//	Writer version: 32 bytes, 0 padded
//	Tempo (bpm): 4 bytes, float
//	Tracks
//
// Track format:
//
//	Track id: 32 bit unsigned int
//	Track name lenght: 1 byte
//	Track name: n bytes, as indicated by previous field
//	Data: 16 bytes, 1 byte per step

const (
	formatFieldBytes     = 13
	dataLengthFieldBytes = 1
	headerBytes          = formatFieldBytes + dataLengthFieldBytes
	versionFieldBytes    = 32
	tempoFieldBytes      = 4
	tracksOffset         = headerBytes + versionFieldBytes + tempoFieldBytes
)

var (
	spliceByteOrder = binary.LittleEndian
	spliceMarker    = [formatFieldBytes]byte{'S', 'P', 'L', 'I', 'C', 'E'}
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	input, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	p, datalen, err := readHeader(input)
	if err != nil {
		return nil, err
	}

	if p.Tracks, err = readTracks(input[tracksOffset : headerBytes+datalen]); err != nil {
		return nil, err
	}

	return p, nil
}

// readHeader decodes the drum machine file's header.
//
// "input" should contain the entire file, as this function will
// validate that the data length indicated in the file does not exceed
// the file's actual size.
//
// It returns the Pattern filled with Version and Tempo, and the data
// length indicated in the header.
func readHeader(input []byte) (*Pattern, int, error) {
	buf := bytes.NewReader(input)

	header := struct {
		Format     [formatFieldBytes]byte
		DataLength uint8
		Writer     [versionFieldBytes]byte
		Tempo      float32
	}{}

	binary.Read(buf, spliceByteOrder, &header)

	if spliceMarker != header.Format {
		return nil, 0, errors.New("Bad format")
	}

	if n, expected := len(input), headerBytes+int(header.DataLength); n < expected {
		err := errors.New(fmt.Sprintf("Not enough data in file: %d vs %d", n, expected))
		return nil, 0, err
	}

	n := bytes.Index(header.Writer[:], []byte{0})
	p := &Pattern{
		Version: string(header.Writer[:n]),
		Tempo:   header.Tempo,
	}

	return p, int(header.DataLength), nil
}

// readtracks will decode the track information contained in the drum
// machine file.
//
// "input" should contain the entire track data
func readTracks(input []byte) ([]Track, error) {
	buf := bytes.NewReader(input)

	tracks := []Track{}

	for {
		header := struct {
			Id      uint32
			NameLen uint8
		}{}

		if err := binary.Read(buf, spliceByteOrder, &header); err != nil {
			if err.Error() == "EOF" {
				break
			}
			return nil, err
		}

		name := make([]byte, header.NameLen)
		if err := binary.Read(buf, spliceByteOrder, &name); err != nil {
			return nil, err
		}

		track := Track{
			Id:   int(header.Id),
			Name: string(name),
		}

		if err := binary.Read(buf, spliceByteOrder, &track.Data); err != nil {
			return nil, err
		}

		tracks = append(tracks, track)
	}

	return tracks, nil
}
