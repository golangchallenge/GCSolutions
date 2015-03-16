package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Version string
	Tempo   float32
	Tracks  []Track
}

// Track represents a single track inside the .splice file
type Track struct {
	ID    byte
	Name  string
	Steps [16]bool
}

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	// it would make more sense to pass in a io.Reader
	p := &Pattern{}

	file, err := os.Open(path)
	if err != nil {
		return nil, nil
	}
	defer file.Close()

	// we create an internal buffer to handle all reads to avoid multiple allocations
	buf := make([]byte, 100)

	_, err = io.ReadFull(file, buf[:6])
	if err != nil {
		return nil, fmt.Errorf("error reading SPLICE signature: %s", err)
	}

	if !bytes.Equal(buf[:6], spliceMagic) {
		return nil, errors.New("invalid SPLICE signature, is not equal to \"SPLICE\"")
	}

	_, err = io.ReadFull(file, buf[:8])
	if err != nil {
		return nil, fmt.Errorf("error reading splice file length: %s", err)
	}

	length := int64(binary.BigEndian.Uint64(buf[:8]))

	// check for file len vs tracks_len+8+6 ?

	// ver
	_, err = io.ReadFull(file, buf[:32])
	if err != nil {
		return nil, fmt.Errorf("error reading splice file version string: %s", err)
	}
	p.Version = strings.TrimRight(string(buf[:32]), "\x00")

	// bpm
	err = binary.Read(file, binary.LittleEndian, &p.Tempo)
	if err != nil {
		return nil, fmt.Errorf("error reading splice file tempo: %s", err)
	}

	// reduce read length by version string and bpm
	length -= 36

	// lets create an array with cap calculated from remaining length and tracks
	// with a name length of 5, to avoid doing any reallocs later
	p.Tracks = make([]Track, 0, length/(16+1+4+5))

	for length > 0 {
		t := Track{}
		n, err := io.ReadFull(file, buf[:5])
		if err != nil {
			return nil, fmt.Errorf("error reading track id and name length: %s", err)
		}
		length -= int64(n)
		t.ID = buf[0]
		nameLen := int(binary.BigEndian.Uint32(buf[1:5]))
		if nameLen > cap(buf) {
			return nil, fmt.Errorf("internal buffer to small for track name (%d>%d)", nameLen, cap(buf))
		}
		n, err = io.ReadFull(file, buf[:nameLen])
		if err != nil {
			return nil, fmt.Errorf("error reading track name: %s", err)
		}
		t.Name = string(buf[:nameLen])
		length -= int64(n)
		n, err = io.ReadFull(file, buf[:16])
		if err != nil {
			return nil, fmt.Errorf("error reading track steps: %s", err)
		}
		length -= int64(n)
		for k, s := range buf[:16] {
			t.Steps[k] = (s == 1)
		}

		p.Tracks = append(p.Tracks, t)
	}

	return p, nil
}
