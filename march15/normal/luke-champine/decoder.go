package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"os"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	// open file
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// decode contents
	p := new(Pattern)
	err = NewDecoder(file).Decode(p)
	if err != nil {
		return nil, err
	}

	return p, nil
}

// A Decoder reads and decodes a Pattern from an input stream. See the package
// docstring for the Pattern encoding specification.
type Decoder struct {
	r io.Reader
}

// NewDecoder returns a Pattern decoder that reads from r. The .splice header
// specifies the Pattern size in bytes; the decoder will not read past this
// limit.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r}
}

// Decode reads Pattern data from its input and stores the result in p.
func (d *Decoder) Decode(p *Pattern) error {
	// decode header data
	err := d.readHeader(p)
	if err != nil {
		return err
	}
	// decode tracks until EOF is reached
	for {
		done, err := d.readTrack(p)
		if done {
			break
		} else if err != nil {
			return err
		}
	}

	return nil
}

// readHeader reads header data from d and stores it in p.
func (d *Decoder) readHeader(p *Pattern) error {
	// decode big-endian portion of header
	var header struct {
		SPLICE  [6]byte
		Size    int64
		Version [32]byte
	}
	err := binary.Read(d.r, binary.BigEndian, &header)
	if err != nil {
		return errors.New("could not decode header: " + err.Error())
	}
	// decode tempo separately, since it's little-endian
	var tempo float32
	err = binary.Read(d.r, binary.LittleEndian, &tempo)
	if err != nil {
		return errors.New("could not decode header: " + err.Error())
	}

	// check SPLICE bytes
	if string(header.SPLICE[:]) != "SPLICE" {
		return errors.New("header missing SPLICE bytes")
	}

	// Modify d.r to be an io.LimitReader, so we're guaranteed to read at most
	// 'size' bytes. If 'size' is too small, we'll get an early EOF. If it's
	// too big, we'll get an error trying to read extra tracks.
	//
	// NOTE: if 'size' cuts off a track exactly, decoding will simply end
	// early (without returning an error), and the returned Pattern will not
	// contain the extra tracks. While unfortunate, this is the only
	// reasonable approach given that the decoder does not know the full size
	// of the reader.
	header.Size -= 36 // we already read the version and tempo
	d.r = io.LimitReader(d.r, header.Size)

	// trim '0's to get the actual version string
	p.version = string(bytes.TrimRight(header.Version[:], string(0)))

	// set the tempo
	p.tempo = tempo
	return nil
}

// readTrack reads a track from d and adds it to p's track list.
func (d *Decoder) readTrack(p *Pattern) (bool, error) {
	// decode track header
	var trackHeader struct {
		ID      byte
		NameLen uint32
	}
	err := binary.Read(d.r, binary.BigEndian, &trackHeader)
	if err == io.EOF {
		// there are no more tracks to decode
		return true, nil
	} else if err != nil {
		return false, errors.New("could not decode track header: " + err.Error())
	}

	// read track name
	name := make([]byte, trackHeader.NameLen)
	_, err = io.ReadFull(d.r, name)
	if err != nil {
		return false, errors.New("could not read track name")
	}

	// read steps
	var steps [stepsPerTrack]bool
	step := make([]byte, 1)
	for i := range steps {
		_, err = d.r.Read(step) // ReadByte would be nicer, but not worth the trouble
		if err != nil {
			return false, errors.New("could not read track steps")
		}
		steps[i] = (step[0] == 1)
	}

	// add track to Pattern
	p.tracks = append(p.tracks, track{
		id:    trackHeader.ID,
		name:  string(name),
		steps: steps,
	})
	return false, nil
}
