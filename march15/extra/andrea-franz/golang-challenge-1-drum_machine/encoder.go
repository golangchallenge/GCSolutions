package drum

import (
	"encoding/binary"
	"fmt"
	"io"
)

type encodeHeaderFn func(*Pattern) error
type encodeTrackFn func(*Track) error

// An Encoder writes a Pattern objects to an output stream.
type Encoder struct {
	w io.Writer
}

// Encode writes the Pattern p to the output stream
func (e *Encoder) Encode(p *Pattern) error {
	err := e.encodeHeader(p)
	if err != nil {
		return err
	}

	for _, t := range p.Tracks {
		err = e.encodeTrack(t)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *Encoder) encodeHeader(p *Pattern) error {
	funcs := []encodeHeaderFn{
		e.encodeSignature,
		e.encodeX,
		e.encodeVersion,
		e.encodeTempo,
	}

	for _, f := range funcs {
		err := f(p)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *Encoder) encodeSignature(p *Pattern) error {
	if len(validSignature) > signatureLength {
		return fmt.Errorf("signature `%s` is too long", validSignature)
	}
	_, err := e.w.Write([]byte(validSignature))
	if err != nil {
		return err
	}

	if len(validSignature) < signatureLength {
		empty := make([]byte, signatureLength-len(validSignature))
		_, err = e.w.Write(empty)
	}

	return err
}

func (e *Encoder) encodeX(p *Pattern) error {
	err := binary.Write(e.w, binary.LittleEndian, p.Header.X)
	return err
}

func (e *Encoder) encodeVersion(p *Pattern) error {
	if len(p.Header.Version) > versionLength {
		return fmt.Errorf("version `%s` is too long", p.Header.Version)
	}

	err := binary.Write(e.w, binary.LittleEndian, []byte(p.Header.Version))
	if err != nil {
		return err
	}

	if len(p.Header.Version) < versionLength {
		empty := make([]byte, versionLength-len(p.Header.Version))
		_, err = e.w.Write(empty)
	}

	return err
}

func (e *Encoder) encodeTempo(p *Pattern) error {
	err := binary.Write(e.w, binary.LittleEndian, p.Header.Tempo)
	return err
}

func (e *Encoder) encodeTrack(t *Track) error {
	funcs := []encodeTrackFn{
		e.encodeTrackID,
		e.encodeTrackInstrument,
		e.encodeTrackSteps,
	}

	for _, f := range funcs {
		err := f(t)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *Encoder) encodeTrackID(t *Track) error {
	return binary.Write(e.w, binary.LittleEndian, t.ID)
}

func (e *Encoder) encodeTrackInstrument(t *Track) error {
	length := uint8(len(t.Instrument))
	err := binary.Write(e.w, binary.LittleEndian, length)
	if err != nil {
		return err
	}

	return binary.Write(e.w, binary.LittleEndian, []byte(t.Instrument))
}

func (e *Encoder) encodeTrackSteps(t *Track) error {
	return binary.Write(e.w, binary.LittleEndian, t.Steps)
}

// NewEncoder returns a new Encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w}
}
