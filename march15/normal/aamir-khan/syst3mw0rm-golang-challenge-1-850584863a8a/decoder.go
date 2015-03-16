package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"log"
	"os"
)

// Errors it can emit
var ErrHeader = errors.New("Invalid / Corrupt drum file")

// A Decoder reads and decodes DRUM binary files from input stream
type Decoder struct {
	r           io.Reader
	MagicHeader [10]byte
	readBytes   int32
	bytesToRead int32
}

// NewDecoder returns a new decoder that reads from r.
//
// Keep track of number of bytes read so far and total bytes to read
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r, readBytes: 0, bytesToRead: 0}
}

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	p := &Pattern{}
	file, err := os.Open(path)
	defer file.Close()

	if err != nil {
		log.Fatal(err)
	}

	dec := NewDecoder(file)
	if err := dec.Decode(p); err != nil {
		return nil, err
	}

	return p, nil
}

// Decode decodes the given decoder and translates to pattern object.
//
// Assumes the data is stores on BigEndian ordering - older machine!
func (dec *Decoder) Decode(p *Pattern) error {
	if err := binary.Read(dec.r, binary.BigEndian, &dec.MagicHeader); err != nil {
		return err
	}
	if string(dec.MagicHeader[:6]) != "SPLICE" {
		return ErrHeader
	}

	if err := binary.Read(dec.r, binary.BigEndian, &dec.bytesToRead); err != nil {
		return err
	}

	version := make([]byte, 32)
	if err := binary.Read(dec.r, binary.BigEndian, &version); err != nil {
		return err
	}
	p.Version = bytes.Trim(version, "\x00")
	dec.readBytes += 32

	if err := binary.Read(dec.r, binary.LittleEndian, &p.Tempo); err != nil {
		return err
	}
	dec.readBytes += 4

	// Now starts the track information, at 55th byte precisely.
	var t Track
	var len int32

	for {
		// that's it. no need to read more information now.
		if dec.readBytes == dec.bytesToRead {
			break
		}

		if err := binary.Read(dec.r, binary.BigEndian, &t.Id); err != nil {
			return err
		}
		dec.readBytes += 1

		if err := binary.Read(dec.r, binary.BigEndian, &len); err != nil {
			return err
		}
		dec.readBytes += 4

		t.Name = make([]byte, len)
		if _, err := io.ReadFull(dec.r, t.Name); err != nil {
			return err
		}
		dec.readBytes += len

		if err := binary.Read(dec.r, binary.BigEndian, &t.Steps); err != nil {
			return err
		}
		dec.readBytes += 16

		p.Tracks = append(p.Tracks, t)
	}

	return nil
}
