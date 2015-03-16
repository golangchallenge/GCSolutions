package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"os"
)

// SPLICE file format:
//
// File header
//       6 bytes - file signature ("SPLICE")
//       8 bytes - length of the pattern header + track data (uint64 big-endian)
//
// Pattern header
//      32 bytes - hardware version string (null-terminated) & padding
//       4 bytes - tempo (float32 little-endian)
//
// Track data (repeated)
//       4 bytes - track ID (uint32 little-endian)
//       1 byte  - track name length (uint8)
//   0-255 bytes - track name
//      16 bytes - track steps (1 = play sound, 0 = silence)

const fileSignature = "SPLICE"

// DecodeFile will read a file and return the Pattern
func DecodeFile(path string) (*Pattern, error) {
	// open file
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// decode the file's contents
	return Decode(f)
}

// Decode will read from a reader and return the Pattern
func Decode(reader io.Reader) (*Pattern, error) {
	return new(fileDecoder).decode(reader)
}

type fileDecoder struct {
	fh fileHeader
	ph patternHeader
	td trackDecoder
}

func (d *fileDecoder) decode(reader io.Reader) (*Pattern, error) {
	// Decode file header
	if err := d.fh.decode(reader); err != nil {
		return nil, err
	}

	// Create a limited reader to read more more than fh.Length bytes
	dataReader := io.LimitReader(reader, int64(d.fh.Length))

	// Decode pattern header
	pattern, err := d.ph.decode(dataReader)
	if err != nil {
		return nil, err
	}

	// Decode all the Tracks
	for {
		track, err := d.td.decode(dataReader)
		if err == io.EOF {
			// no more tracks, exit the loop
			break
		}
		if err != nil {
			return nil, err
		}
		pattern.AddTrack(*track)
	}

	return pattern, nil
}

type fileHeader struct {
	Signature [6]byte
	Length    uint64
}

func (fh *fileHeader) decode(reader io.Reader) error {
	if err := binary.Read(reader, binary.BigEndian, fh); err != nil {
		return err
	}

	// Ensure that signature matches the expected value
	if string(fh.Signature[:len(fileSignature)]) != fileSignature {
		return errors.New("SPLICE file signature not found")
	}

	return nil
}

type patternHeader struct {
	HwVers [32]byte
	Tempo  float32
}

func (ph *patternHeader) decode(reader io.Reader) (*Pattern, error) {
	if err := binary.Read(reader, binary.LittleEndian, ph); err != nil {
		return nil, err
	}

	p := &Pattern{
		Version: string(ph.HwVers[:bytes.IndexByte(ph.HwVers[:], 0)]),
		Tempo:   ph.Tempo,
	}
	return p, nil
}

type trackDecoder struct {
	len   uint8
	tmp   [256]byte // big enough for the maximum name length
	track Track
}

func (t *trackDecoder) decode(reader io.Reader) (*Track, error) {
	// ID: 4 bytes
	if err := binary.Read(reader, binary.LittleEndian, &t.track.ID); err != nil {
		return nil, err
	}

	// nameLength: 1 byte
	if err := binary.Read(reader, binary.LittleEndian, &t.len); err != nil {
		return nil, err
	}

	// Name: nameLength bytes
	nameBytes := t.tmp[0:t.len]
	if _, err := io.ReadFull(reader, nameBytes); err != nil {
		return nil, err
	}
	t.track.Name = string(nameBytes)

	// steps: 16 bytes
	steps := t.tmp[0:16]
	if _, err := io.ReadFull(reader, steps); err != nil {
		return nil, err
	}

	// convert step bytes into booleans
	for i, val := range steps {
		t.track.Steps[i] = (val > 0)
	}

	return &(t.track), nil
}
