package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io/ioutil"
)

// Reads one Track.
func readTrack(rd *bytes.Reader) (*Track, error) {
	var err error
	var id int32
	var nameLen byte
	var name []byte
	var nameString string
	var steps []byte = make([]byte, 16)
	var stepsBool []bool = make([]bool, 16)

	// The first 4 bytes is the ID of the track.
	err = binary.Read(rd, binary.LittleEndian, &id)
	if err != nil {
		return nil, err
	}

	// The next byte specifies the length of the track's name.
	err = binary.Read(rd, binary.LittleEndian, &nameLen)
	if err != nil {
		return nil, err
	}

	// Read `nameLen` bytes and convert to string.
	name = make([]byte, nameLen)
	err = binary.Read(rd, binary.LittleEndian, &name)
	if err != nil {
		return nil, err
	}
	nameString = string(name)

	// The next 16 bytes is the steps of the track.
	err = binary.Read(rd, binary.LittleEndian, &steps)
	if err != nil {
		return nil, err
	}

	// The internal representation uses boolean values for steps, so convert them.
	for i, s := range steps {
		stepsBool[i] = s == 1
	}

	t := &Track{
		Id:    id,
		Name:  nameString,
		Steps: stepsBool,
	}

	return t, nil
}

// Reads a Pattern.
func readPattern(rd *bytes.Reader) (*Pattern, error) {
	var err error
	var magic []byte = make([]byte, 6)
	var patLen byte
	var version []byte = make([]byte, 32)
	var versionString string
	var tempo float32
	var tracks []*Track

	// Read and verify magic number.
	err = binary.Read(rd, binary.LittleEndian, &magic)
	if err != nil {
		return nil, err
	}
	if string(magic) != "SPLICE" {
		return nil, errors.New("Not a valid SPLICE file.")
	}

	// Skip 7 null bytes.
	rd.Seek(7, 1)

	// Read pattern length. Note that length is from current offset (14).
	err = binary.Read(rd, binary.LittleEndian, &patLen)
	if err != nil {
		return nil, err
	}

	// Read the version and convert to string.
	err = binary.Read(rd, binary.LittleEndian, &version)
	if err != nil {
		return nil, err
	}
	versionString = string(bytes.Trim(version, "\x00"))

	// Read the tempo.
	err = binary.Read(rd, binary.LittleEndian, &tempo)
	if err != nil {
		return nil, err
	}

	// Read the tracks of this pattern.
	for pos, _ := rd.Seek(0, 1); pos < int64(patLen)+14; pos, _ = rd.Seek(0, 1) {
		t, err := readTrack(rd)
		if err != nil {
			return nil, err
		}
		tracks = append(tracks, t)
	}

	p := &Pattern{
		Version: versionString,
		Tempo:   tempo,
		Tracks:  tracks,
	}

	return p, nil
}

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	rd := bytes.NewReader(data)

	p, err := readPattern(rd)
	if err != nil {
		return nil, err
	}

	return p, nil
}
