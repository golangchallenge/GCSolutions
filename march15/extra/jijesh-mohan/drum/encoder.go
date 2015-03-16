package drum

import (
	"encoding/binary"
	"io"
	"os"
)

// EncodeFile encode the drum pattern to a file path specified
// in the argument
func EncodeFile(pattern *Pattern, path string) error {
	// create a new file
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	e := encoder{w: file, p: pattern}

	return e.Encode()
}

// Encoder is a type to encode the drum pattern to
// a .splice filw
type encoder struct {
	w   io.Writer
	p   *Pattern
	err error
}

// Encode reads the pattern type and create .splice file.
func (e *encoder) Encode() error {

	// create header information for .splice file
	title := make([]byte, HEADER_PADDING)
	copy(title, "SPLICE")
	e.encodeNextItem(title)

	e.encodeFileSize()
	e.encodeNextItem(e.p.Header.Version)
	e.encodeNextItem(e.p.Header.Tempo)

	// encode track information
	padding := make([]byte, TRACK_PADDING)
	for _, t := range e.p.Tracks {
		e.encodeNextItem(t.ID)
		e.encodeNextItem(padding)
		e.encodeNextItem(uint8(len(t.Name)))
		e.encodeNextItem(t.Name)
		e.encodeNextItem(t.Steps)
	}
	return e.err
}

// encode next data item to the writer
func (e *encoder) encodeNextItem(data interface{}) {
	if e.err != nil {
		return
	}
	e.err = binary.Write(e.w, binary.LittleEndian, data)
}

// encodeFileSize calculate and encode the actual size
// to the writer.
func (e *encoder) encodeFileSize() {
	if e.err != nil {
		return
	}
	e.p.Header.Size = e.headerSize() + e.trackSize()
	e.err = binary.Write(e.w, binary.BigEndian, e.p.Header.Size)
}

// headerSize calculates the size of the header
// with respect the .splice file format.
func (e *encoder) headerSize() uint16 {
	return uint16(binary.Size(e.p.Header.Version) + binary.Size(e.p.Header.Tempo))
}

// trackSize calculate the total track size
// in the pattern.
func (e *encoder) trackSize() uint16 {
	var size uint16
	for _, t := range e.p.Tracks {
		size += uint16(binary.Size(t.ID))
		size += TRACK_PADDING
		size += 1 // one byte length for storing name's length
		size += uint16(len(t.Name))
		size += uint16(len(t.Steps))
	}
	return size
}
