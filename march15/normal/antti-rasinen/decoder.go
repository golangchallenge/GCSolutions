package drum

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"os"
)

var MAGIC = []byte("SPLICE\x00\x00\x00\x00\x00\x00\x00")

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.

func DecodeFile(path string) (*Pattern, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	r := bufio.NewReader(f)

	if err := readMagic(r); err != nil {
		return nil, err
	}

	var bufSize byte

	if err := binary.Read(r, binary.LittleEndian, &bufSize); err != nil {
		return nil, err
	}

	buf, err := readPattern(r, bufSize)
	if err != nil {
		return nil, err
	}

	p, err := parsePattern(buf)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func readMagic(r io.Reader) error {
	magicbuf := make([]byte, len(MAGIC))

	n, err := r.Read(magicbuf)
	if err != nil {
		return err
	}

	if n < len(MAGIC) {
		return errors.New("Could not read magic")
	}

	if !bytes.Equal(magicbuf, MAGIC) {
		return errors.New("Bad magic at start of file")
	}

	return nil
}

func readPattern(r io.Reader, sz byte) (*bytes.Reader, error) {
	buf := make([]byte, sz)
	n, err := r.Read(buf)

	if err != nil {
		return nil, err
	}

	if n < int(sz) {
		return nil, errors.New("Didn't read whole pattern")
	}
	return bytes.NewReader(buf), nil
}

func parsePattern(r io.Reader) (*Pattern, error) {
	p := &Pattern{}

	if err := binary.Read(r, binary.LittleEndian, &p.Header); err != nil {
		return nil, err
	}

	for {
		t, err := parseTrack(r)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		p.tracks = append(p.tracks, t)
	}

	return p, nil
}

func parseTrack(r io.Reader) (*Track, error) {
	var nameLen byte

	t := &Track{}

	if err := binary.Read(r, binary.LittleEndian, &t.Id); err != nil {
		return nil, err
	}

	if err := binary.Read(r, binary.LittleEndian, &nameLen); err != nil {
		return nil, err
	}

	name := make([]byte, nameLen)
	_, err := r.Read(name)
	if err != nil {
		return nil, err
	}
	t.Name = string(name)

	if err := binary.Read(r, binary.LittleEndian, &t.Steps); err != nil {
		return nil, err
	}
	return t, nil
}
