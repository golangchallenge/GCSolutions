package drum

import (
	"encoding/binary"
	"io"
	"os"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {

	// // open file and return error if unable to open
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	d := NewDecoder(file)
	return d.Decode()
}

// A Decoder is the type for decode .slice file information
// see the Decode function for details.
type decoder struct {
	r   io.ReadSeeker
	p   *Pattern
	err error

	length uint8 // track name length
}

// NewDecoder return a new decoder with
// the buffer specified.
func NewDecoder(r io.ReadSeeker) *decoder {
	return &decoder{r: r, p: &Pattern{}}
}

// Decode reads the tokens from the reader
// and returns a pointer to the pattern.
func (d *decoder) Decode() (*Pattern, error) {

	d.reset()
	d.skipPadding(HEADER_PADDING)

	d.decodeSize()
	d.readBytes(d.p.Header.Version[:])
	d.decodeItem(&d.p.Header.Tempo)

	// read track informations
	d.decodeTracks()

	if d.err != nil {
		return nil, d.err
	}
	return d.p, nil
}

// Reset the decoder to begining of the input buffer.
func (d *decoder) reset() {
	_, d.err = d.r.Seek(0, 0)
}

// decodeTracks decodes track informations from the io.reader
// return error if there is any failure.
func (d *decoder) decodeTracks() {
	if d.err != nil {
		return
	}
	// calculate track size which is size minus Version size + Tempo size
	contentSize := d.p.Header.Size - 36

	// decode each track information
	var t Track

	for contentSize > 0 {
		d.decodeItem(&t.ID)
		d.skipPadding(TRACK_PADDING)
		d.decodeItem(&d.length)

		// create name byte slice with correct length
		t.Name = make([]byte, d.length)
		d.readBytes(t.Name)
		d.readBytes(t.Steps[:])

		if d.err != nil {
			return
		}

		d.p.Tracks = append(d.p.Tracks, t)
		// reset the remaining content size
		contentSize -= 21 + uint16(d.length)
	}
}

// readBytes reads upto len(data) bytes from the buffer.
func (d *decoder) readBytes(data []byte) {
	if d.err != nil {
		return
	}
	_, d.err = d.r.Read(data)
}

// decodeSize decodes file size information from buffer.
// Size can be more than one byte so using big endian.
func (d *decoder) decodeSize() {
	if d.err != nil {
		return
	}

	d.err = binary.Read(d.r, binary.BigEndian, &d.p.Header.Size)
}

// Decode numbers and other datatypes from the decoder buffer.
func (d *decoder) decodeItem(data interface{}) {
	if d.err != nil {
		return
	}

	d.err = binary.Read(d.r, binary.LittleEndian, data)
}

// Skip padding from the decoder
// padding length specified as the parameter
func (d *decoder) skipPadding(length int64) {
	if d.err != nil {
		return
	}
	_, d.err = d.r.Seek(length, 1)
}
