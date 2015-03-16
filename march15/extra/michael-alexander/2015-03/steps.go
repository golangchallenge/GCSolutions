package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const (
	stepSep  byte = '|'
	stepPlay byte = 'x'
	stepRest byte = '-'
)

// Steps is a definition of a 16 step drum loop, usually for a track.
type Steps [16]bool

func (s Steps) String() string {
	buf := bytes.NewBuffer([]byte{stepSep})
	for i, step := range s {
		c := stepRest
		if step {
			c = stepPlay
		}
		buf.WriteByte(c)
		if (i+1)%4 == 0 {
			buf.WriteByte(stepSep)
		}
	}
	return buf.String()
}

// Encode writes the steps as 0/1 bytes to the writer.
func (s Steps) Encode(w io.Writer) error {
	for _, st := range s {
		val := byte(0)
		if st {
			val = 1
		}
		if err := binary.Write(w, binary.LittleEndian, val); err != nil {
			return fmt.Errorf("unable to write step, %v", err)
		}
	}
	return nil
}

// DecodeSteps decodes 16 step bytes from a reader.
func DecodeSteps(r io.Reader) (b Steps, err error) {
	b = Steps{}
	p := make([]byte, len(b))
	if _, err = r.Read(p); err != nil {
		if err != io.EOF {
			err = fmt.Errorf("unable to read steps, %v", err)
		}
	}

	for i, v := range p {
		switch v {
		case 0:
			b[i] = false
		case 1:
			b[i] = true
		default:
			err = errors.New("decoding steps failed, found a non 0 or 1 byte")
			return
		}
	}

	return
}
