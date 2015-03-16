package drum

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

const (
	headerBytesCount  int64 = 13
	versionBytesCount int   = 32
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	return NewDecoder(f).Unmarshal()
}

// NewDecoder creates a *Decoder instance over an io.Reader instance.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		reader: r,
	}
}

// Decoder is intended for use in unmarshalling binary formatted pattern data.
type Decoder struct {
	reader io.Reader
}

// Unmarshal will return a Pattern instance or an error if there is a problem
// reading the binary data handed to the decoder.
func (d *Decoder) Unmarshal() (*Pattern, error) {
	p := &Pattern{}

	// Discard the first bytes to get the version
	if n, err := io.CopyN(ioutil.Discard, d.reader, headerBytesCount); err != nil {
		return p, err
	} else if n != headerBytesCount {
		return p, fmt.Errorf("partial header read; got %v bytes instead of %v bytes", n, headerBytesCount)
	}

	// Get the total size of the stream
	var size uint8
	var bytesRead uint8
	err := binary.Read(d.reader, binary.LittleEndian, &size)
	if err != nil {
		return p, err
	}

	// Read the version string into our buffer, being sure we read the expected amount
	buf := make([]byte, versionBytesCount)
	n, err := io.ReadFull(d.reader, buf)
	if err != nil {
		return p, err
	}

	bytesRead += uint8(n)

	// Don't hand any 0 bytes to the version string
	index := bytes.IndexByte(buf, 0x00)
	if index == -1 {
		index = versionBytesCount
	}

	p.SetVersion(string(buf[0:index]))

	var tempo float32
	err = binary.Read(d.reader, binary.LittleEndian, &tempo)
	if err != nil {
		return p, err
	}

	bytesRead += 4
	p.SetTempo(tempo)

	// Next we read tracks until we reach our expected size.
	for bytesRead < size {
		track := NewTrack()
		n, err = d.unmarshalTrack(track)

		if err != nil {
			return p, err
		}

		p.AddTrack(track)
		bytesRead += uint8(n)
	}

	return p, nil
}

// unmarshalTrack will read in the id, name and steps for a particular part of the pattern.
// The method assumes it is called at the beginning of a tracks binary data.
func (d *Decoder) unmarshalTrack(t *Track) (int, error) {

	bytesRd := 0
	var id uint32
	if err := binary.Read(d.reader, binary.LittleEndian, &id); err != nil {
		return bytesRd, err
	}

	t.SetID(id)
	bytesRd += 4

	var length uint8
	err := binary.Read(d.reader, binary.LittleEndian, &length)
	if err != nil {
		return bytesRd, err
	}

	bytesRd++

	name := make([]byte, length)
	n, err := io.ReadFull(d.reader, name)
	if err != nil {
		return bytesRd, err
	}

	bytesRd += n
	t.SetName(string(name))

	n, err = d.unmarshalSteps(t)
	if err != nil {
		return bytesRd, err
	}

	bytesRd += n
	return bytesRd, nil
}

// unamrshalSteps will perform the steps parsing part of the unmarshalTrack method. This method
// also assumes it is at the beginning of the binary data for the steps and also that there is
// only 16 of them.
func (d *Decoder) unmarshalSteps(t *Track) (int, error) {

	bytesRd := 0
	for i := 0; i < 16; i++ {
		var s uint8
		if err := binary.Read(d.reader, binary.LittleEndian, &s); err != nil {
			return bytesRd, err
		}

		bytesRd++
		t.SetStep(i, s > 0)
	}

	return bytesRd, nil
}
