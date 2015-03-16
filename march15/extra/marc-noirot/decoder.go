package drum

import (
	"bytes"
	"encoding/binary"
	"io"
	"os"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	in, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer in.Close()

	p, err := DecodePattern(in)
	if err != nil {
		return nil, err
	}

	return p, nil
}

// DecodePattern returns a pattern read from a binary reader.
func DecodePattern(in io.Reader) (*Pattern, error) {
	// read signature, 13 bytes
	var sig [13]byte
	if _, err := in.Read(sig[:]); err != nil {
		return nil, err
	}
	// check signature
	if !bytes.Equal(sig[:], []byte(magicHeader)) {
		return nil, ErrBadSignature
	}

	// read data size, one unsigned byte
	var size [1]uint8
	if _, err := in.Read(size[:]); err != nil {
		return nil, err
	}

	// switch to a LimitReader so we don't read more than the advertised size
	in = io.LimitReader(in, int64(size[0]))

	// read software version, null-terminated string
	// within a fixed space of 32 bytes
	var hwVer [32]byte
	if _, err := in.Read(hwVer[:]); err != nil {
		return nil, err
	}

	p := &Pattern{}
	p.HWVersion = string(bytes.Trim(hwVer[:], "\x00"))

	// read tempo, little endian 32-bit float
	if err := binary.Read(in, binary.LittleEndian, &p.Tempo); err != nil {
		return nil, err
	}

	// read list of tracks
	p.Tracks = make([]Track, 0, 10)
	for {
		var track Track

		// read track ID, little endian 32-bit integer
		if err := binary.Read(in, binary.LittleEndian, &track.ID); err != nil {
			if err == io.EOF {
				break // end of data, this is expected, break the loop
			} else {
				return nil, err
			}
		}

		// read track name length, one unsigned byte
		var nameLen [1]uint8
		if _, err := in.Read(nameLen[:]); err != nil {
			return nil, err
		}

		// read track name
		name := make([]byte, nameLen[0])
		if _, err := in.Read(name); err != nil {
			return nil, err
		}
		track.Name = string(name)

		// read all steps
		if _, err := in.Read(track.Steps[:]); err != nil {
			return nil, err
		}

		p.Tracks = append(p.Tracks, track)
	}

	return p, nil
}
