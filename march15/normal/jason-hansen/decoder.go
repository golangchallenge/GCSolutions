package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"os"
)

var (
	// The first 6 bytes of a splice file should always be "SPLICE"
	validSignature = [6]byte{'S', 'P', 'L', 'I', 'C', 'E'}

	// ErrInvalidSignature is the error returned by Decode when the first
	// six bytes of the input are not exactly "SPLICE".
	ErrInvalidSignature = errors.New("invalid signature")

	// ErrInvalidStep is the error returned by Decode when a byte
	// representing a step is neither 0 nor 1.
	ErrInvalidStep = errors.New("invalid step")

	// ErrUnexpectedEOF is the error returned by Decode if the end
	// of the file is reached when there is still data left to decode.
	ErrUnexpectedEOF = errors.New("unexpected EOF")
)

// The version string is stored as 32 bytes padded with zero bytes on the right.
const maxVersionLength = 32

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	r, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	p := &Pattern{}
	dec := NewPatternDecoder(r)
	if err = dec.Decode(p); err != nil {
		return nil, err
	}

	return p, nil
}

// A PatternDecoder reads and decodes patterns from an input stream.
type PatternDecoder struct {
	r         io.Reader // Input to decode from
	remaining uint64    // Number of bytes remaining to read from r
	err       error     // The first error encountered while decoding
}

// NewPatternDecoder returns a new decoder that reads from r.
func NewPatternDecoder(r io.Reader) *PatternDecoder {
	return &PatternDecoder{r: r}
}

// Decode reads and decodes a pattern from its input
// and stores it in the pattern pointed to by p.
func (dec *PatternDecoder) Decode(p *Pattern) error {
	// The first error encountered will be stored in dec.err
	// so there is no need to check errors after every call.
	dec.decodeHeader()
	p.version = dec.decodeVersion()
	p.tempo = dec.decodeTempo()
	p.tracks = dec.decodeTracks()

	// All EOF errors are unexpected since we're
	// given the remaining bytes to encode.
	if dec.err == io.EOF {
		dec.err = ErrUnexpectedEOF
	}

	return dec.err
}

// decodeHeader reads and checks the file signature and reads in
// the number of remaining bytes left to decode after the header.
func (dec *PatternDecoder) decodeHeader() {
	var signature [6]byte
	dec.decodeIgnoreRemaining(&signature, nil)
	if dec.err == nil && signature != validSignature {
		dec.err = ErrInvalidSignature
		return
	}
	dec.decodeIgnoreRemaining(&dec.remaining, binary.BigEndian)
}

// decodeVersion decodes the hardware version the pattern data
// was created with and returns it.
func (dec *PatternDecoder) decodeVersion() string {
	version := make([]byte, maxVersionLength)
	dec.decode(&version, nil)
	// We need to remove trailing zero bytes from the version.
	end := bytes.IndexByte(version, 0)
	if end > 0 {
		version = version[:end]
	}
	return string(version)
}

// decodePattern decodes the tempo from its input in beats per minute
// and returns it.
func (dec *PatternDecoder) decodeTempo() float32 {
	var tempo float32
	dec.decode(&tempo, binary.LittleEndian)
	return tempo
}

// decodeTracks decodes all of the tracks from its input
// and returns a slice of pointers to them.
func (dec *PatternDecoder) decodeTracks() []track {
	tracks := []track{}
	// Keep decoding tracks while there are bytes left to read
	for dec.err == nil && dec.remaining > 0 {
		t := track{}
		dec.decode(&t.id, nil)

		// Determine the length of the track name and decode it.
		var nameLen uint32
		dec.decode(&nameLen, binary.BigEndian)
		name := make([]byte, nameLen)
		dec.decode(&name, nil)
		t.name = string(name)

		// Decode the steps for the track. There is one byte per step.
		for i := 0; i < stepsPerTrack; i++ {
			dec.decode(&t.steps[i], nil)
			if t.steps[i] != stepOn && t.steps[i] != stepOff {
				if dec.err == nil {
					dec.err = ErrInvalidStep
					break
				}
			}
		}

		if dec.err == nil {
			tracks = append(tracks, t)
		}
	}
	return tracks
}

// decode reads structured binary data from its input into data.
//
// Data must be a pointer to a fixed-size value or a slice of fixed-size values.
// Bytes are decoded using the specified byte order. The first error
// encountered will be stored in dec.err and subsequent calls will have no
// effect. This allows for decode to be called multiple times without checking
// errors each time. Each successful read will also decrement dec.remaining
// by the number of bytes read.
func (dec *PatternDecoder) decode(data interface{}, order binary.ByteOrder) error {
	err := dec.decodeIgnoreRemaining(data, order)
	if err == nil {
		dec.remaining -= uint64(binary.Size(data))
	}
	return err
}

// decodeIgnoreRemaining behaves exactly the same as decode except
// dec.remaining does not get modified.
func (dec *PatternDecoder) decodeIgnoreRemaining(data interface{}, order binary.ByteOrder) error {
	if dec.err != nil {
		return dec.err
	}
	dec.err = binary.Read(dec.r, order, data)
	return dec.err
}
