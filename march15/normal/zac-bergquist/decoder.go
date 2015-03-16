package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"os"
)

var (
	// ErrMissingHeader is the error returned if the SPLICE
	// header can't be read from a pattern file.
	ErrMissingHeader = errors.New("drum: missing SPLICE header")
)

var magic = []byte{'S', 'P', 'L', 'I', 'C', 'E'}

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return DecodeReader(file)
}

// DecodeReader decodes a drum machine pattern from the specified reader
// and returns a pointer to the parsed pattern which is the entry point
// to the rest of the data.
func DecodeReader(r io.Reader) (*Pattern, error) {
	start := make([]byte, 6)
	_, err := io.ReadFull(r, start)
	if err != nil || bytes.Compare(start, magic) != 0 {
		return nil, ErrMissingHeader
	}

	var length uint64
	if err := binary.Read(r, binary.BigEndian, &length); err != nil {
		return nil, err
	}

	version := make([]byte, 32)
	_, err = io.ReadFull(r, version)
	if err != nil {
		return nil, err
	}
	// find where the version string ends
	endv := bytes.IndexByte(version, 0)
	if endv == -1 {
		endv = len(version)
	}

	var tempo float32
	if err := binary.Read(r, binary.LittleEndian, &tempo); err != nil {
		return nil, err
	}

	p := Pattern{
		Version: string(version[:endv]),
		Tempo:   tempo,
	}

	err = nil
	for track, err := readTrack(r); err == nil; {
		p.AddTrack(track)
		track, err = readTrack(r)
	}

	// EOF is expected (not an error), but any other errors
	// should propagate to the caller
	if err == io.EOF {
		err = nil
	}
	return &p, err
}

func readTrack(r io.Reader) (Track, error) {
	t := Track{}

	var id uint32
	if err := binary.Read(r, binary.LittleEndian, &id); err != nil {
		return t, err
	}
	t.ID = id

	var nameLen [1]byte
	_, err := r.Read(nameLen[:])
	if err != nil {
		return t, err
	}
	name := make([]byte, nameLen[0])
	_, err = io.ReadFull(r, name)
	if err != nil {
		return t, err
	}
	t.Name = string(name)

	measure := make([]byte, 16)
	_, err = io.ReadFull(r, measure)
	if err != nil {
		return t, err
	}
	for i := uint16(0); i < 16; i++ {
		t.ToggleStep(i, measure[i] != 0)
	}

	return t, nil
}
