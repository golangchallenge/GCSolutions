package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

const (
	// spliceHeader is the file header for .splice files
	spliceHeader = "SPLICE\x00\x00\x00\x00\x00\x00"

	// maxTrackNameLen is the maximum name of a track name we currently
	// support.
	// TODO: find out if there's an actual size limit.
	maxTrackNameLen = 255
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("drum: could not open file: %s\n", err)
	}
	defer f.Close()

	p := &Pattern{}
	err = NewDecoder(f).Decode(p)
	return p, err
}

// Decoder reads and decodes Patterns from an input stream.
type Decoder struct {
	r   io.Reader
	buf *bytes.Reader
	err error
}

// NewDecoder returns a new decoder that reads from r.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r}
}

// Decode reads the next pattern from its input and stores it in p.
func (d *Decoder) Decode(p *Pattern) error {
	// 12 byte header
	header := make([]byte, 12)
	if err := binary.Read(d.r, binary.BigEndian, &header); err != nil {
		return fmt.Errorf("drum: read error: %s", d.err)
	} else if string(header) != spliceHeader {
		return errors.New("drum: unrecognized .splice file format")
	}

	// 2-byte length of the data
	var size uint16
	if err := binary.Read(d.r, binary.BigEndian, &size); err != nil {
		return d.err
	}

	// read the rest of the data into a buffer
	buf := make([]byte, size)
	if n, err := io.ReadFull(d.r, buf); err != nil {
		return fmt.Errorf("drum: read error: %s", err)
	} else if n != int(size) {
		return errors.New("drum: unexpected end of file")
	}
	d.buf = bytes.NewReader(buf)

	// 32-byte version
	version := make([]byte, 32)
	d.read(binary.BigEndian, &version)
	p.Version = string(bytes.Trim(version, "\x00"))

	// 4-byte tempo (little-endian float)
	d.read(binary.LittleEndian, &p.Tempo)

	// keep reading tracks until we reach the end of the buffer
	for {
		var t Track

		// 1-byte track identifier
		d.read(binary.BigEndian, &t.ID)

		// the previous read may have failed with an EOF; this is the only
		// point where that is expected.
		if d.err == io.EOF {
			break
		}

		// 4-byte track name length
		var nlen uint32
		d.read(binary.BigEndian, &nlen)
		if nlen > maxTrackNameLen {
			return fmt.Errorf("drum: unexpected track name length: %d\n", nlen)
		}

		name := make([]byte, nlen)
		d.read(binary.BigEndian, &name)
		t.Name = string(name)

		// 16 1-byte steps
		var step byte
		for i := 0; i < 16; i++ {
			d.read(binary.BigEndian, &step)
			t.Steps[i] = (step == 1)
		}

		if d.err != nil {
			return fmt.Errorf("drum: incomplete .splice file: %s", d.err)
		}
		p.Tracks = append(p.Tracks, t)
	}

	// if there's any data left in the reader, read it into the raw field so
	// it's preserved when we encode this pattern again
	var err error
	p.raw, err = ioutil.ReadAll(d.r)
	return err
}

// read is a convenience method for binary.Read from the internal buffer so we
// reduce the number of error checks. It only reads if it hasn't encountered an
// error so far.
func (d *Decoder) read(order binary.ByteOrder, data interface{}) {
	if d.err != nil {
		return
	}
	d.err = binary.Read(d.buf, order, data)
}
