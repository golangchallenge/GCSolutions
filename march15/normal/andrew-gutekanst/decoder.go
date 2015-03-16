package drum

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"
	"os"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	// Open the given file
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return decodeReader(f)
}

// getValidByteReader returns a of the reader of valid bytes (ignores junk data appended to the end)
// given a io.Reader of a .splice file
func getValidByteReader(f io.Reader) (*bytes.Reader, error) {
	// Read past the file magic ("SPLICE")
	_, err := io.ReadFull(f, make([]byte, 6))
	if err != nil {
		return nil, err
	}

	// Read in the amount of bytes left until the EOF
	bytesLeftBuf := make([]byte, 8)
	_, err = io.ReadFull(f, bytesLeftBuf)
	if err != nil {
		return nil, err
	}
	bytesLeft := binary.BigEndian.Uint64(bytesLeftBuf)

	// Make a buffer the size of the remaining bytes in the file as specified in the file header(bytesLeft)
	buf := make([]byte, bytesLeft)
	_, err = io.ReadFull(f, buf)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(buf), nil
}

func decodeReader(r io.Reader) (*Pattern, error) {
	p := &Pattern{}

	// Get a reader
	f, err := getValidByteReader(r)
	if err != nil {
		return nil, err
	}

	// Read in the version string
	var version [32]byte
	_, err = io.ReadFull(f, version[:])
	if err != nil {
		return nil, err
	}
	p.Version = string(bytes.Trim(version[:], "\x00"))

	// Read in the tempo
	tempoBuf := make([]byte, 4)
	_, err = io.ReadFull(f, tempoBuf)
	if err != nil {
		return nil, err
	}
	p.Tempo = math.Float32frombits(binary.LittleEndian.Uint32(tempoBuf))

	// Start reading in the tracks
	for {
		t := &Track{}

		// Get track ID
		idBuf := make([]byte, 1)
		_, err := io.ReadFull(f, idBuf)
		if err == io.EOF {
			// EOF reached, no more tracks
			break
		} else if err != nil {
			return nil, err
		}
		t.ID = int(idBuf[0])

		// Get the length of the track name string
		nameLenBuf := make([]byte, 4)
		_, err = io.ReadFull(f, nameLenBuf)
		if err != nil {
			return nil, err
		}
		nameLength := binary.BigEndian.Uint32(nameLenBuf)

		// Read the bytes of the track name
		tmpName := make([]byte, nameLength)
		_, err = io.ReadFull(f, tmpName)
		if err != nil {
			return nil, err
		}
		t.Name = string(tmpName[:])

		// Get the track beats
		var beats [16]byte
		_, err = io.ReadFull(f, beats[:])
		if err != nil {
			return nil, err
		}
		t.Beats = beats

		// Append the current track
		p.Tracks = append(p.Tracks, t)
	}

	return p, nil
}
