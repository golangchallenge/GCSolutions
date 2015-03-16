package drum

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {

	body, err := getBody(path)
	if err != nil {
		return nil, err
	}

	p := new(Pattern)

	// Read the version (32 byte string right padded with the null char).
	var version [32]byte
	if _, err := io.ReadFull(body, version[:]); err != nil {
		return nil, err
	}
	p.Version = strings.TrimRight(string(version[:]), "\000")

	// Read the tempo (4 byte single precision float).
	var tempo float32
	if err := binary.Read(body, binary.LittleEndian, &tempo); err != nil {
		return nil, err
	}
	p.Tempo = tempo

	// Read tracks while there is more body to read.
	for body.Len() > 0 {
		track, err := decodeTrack(body)
		if err != nil {
			return nil, err
		}
		p.Tracks = append(p.Tracks, track)
	}

	return p, nil
}

// getBody gets the body of the drum machine file found at path. It returns the
// body as a bytes.Reader.
func getBody(path string) (*bytes.Reader, error) {

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Check for the presence of the "magic number".
	const magicNumber = "SPLICE"
	var magic [len(magicNumber)]byte
	if _, err := io.ReadFull(file, magic[:]); err != nil {
		return nil, err
	}
	if string(magic[:]) != magicNumber {
		return nil, fmt.Errorf("didn't find %s as first 6 bytes", magicNumber)
	}

	// Read in the body size (8 byte integer).
	var size uint64
	if err := binary.Read(file, binary.BigEndian, &size); err != nil {
		return nil, err
	}

	// Read the rest of the file. This is the file body, and must be at least
	// size bytes. Ignore any trailing data.
	buf, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	if len(buf) < int(size) {
		return nil, fmt.Errorf("body length only %d bytes but expected at least %d bytes",
			len(buf), size)
	}
	return bytes.NewReader(buf[:size]), err
}

func decodeTrack(r io.Reader) (Track, error) {

	var t Track

	// Read track index (4 bytes).
	if err := binary.Read(r, binary.LittleEndian, &t.Index); err != nil {
		return t, err
	}

	// Read the Name - 1 byte indicating size, then that many bytes as a string.
	var size uint8
	if err := binary.Read(r, binary.LittleEndian, &size); err != nil {
		return t, err
	}
	name := make([]byte, size)
	if _, err := io.ReadFull(r, name); err != nil {
		return t, err
	}
	t.Name = string(name)

	// Read the triggers (1 byte for each trigger).
	for i := range t.Triggers {
		var b byte
		if err := binary.Read(r, binary.BigEndian, &b); err != nil {
			return t, err
		}
		// 0x00 doesn't trigger a sound. Anything else does.
		t.Triggers[i] = b != 0x00
	}
	return t, nil
}
