package drum

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Version string
	Tempo   float32
	Tracks  []*Track
}

// NewPattern creates a new Pattern instance.
func NewPattern() *Pattern {
	return &Pattern{Tracks: []*Track{}}
}

func (p Pattern) String() string {
	buf := bytes.NewBufferString(fmt.Sprintf(
		"Saved with HW Version: %s\nTempo: %v\n",
		p.Version,
		p.Tempo,
	))

	// Output each track.
	for _, t := range p.Tracks {
		buf.WriteString(fmt.Sprintf("%s\t%s\n", t, t.Steps.String()))
	}

	return buf.String()
}

// Encode encodes the pattern in binary to the Splice format.
func (p Pattern) Encode(w io.Writer) error {
	body := bytes.NewBuffer([]byte{})
	if err := p.EncodeMeta(body); err != nil {
		return fmt.Errorf("unable to write meta to body, %v", err)
	}

	if err := p.EncodeTracks(body); err != nil {
		return fmt.Errorf("unable to encode tracks, %v", err)
	}

	if err := EncodeHeader(w, int32(body.Len())); err != nil {
		return fmt.Errorf("unable to encode header, %v", err)
	}

	if _, err := io.Copy(w, body); err != nil {
		return fmt.Errorf("unable to copy body to writer, %v", err)
	}
	return nil
}

// EncodeMeta writes the version and tempo to the writer.
func (p Pattern) EncodeMeta(w io.Writer) error {
	ver := []byte(p.Version)
	verLen := len(ver)
	if verLen > widthVersion {
		ver = ver[:widthVersion]
		verLen = widthVersion
	}

	if _, err := w.Write(ver); err != nil {
		return fmt.Errorf("unable to write version text, %v", err)
	}
	if _, err := w.Write(bytes.Repeat([]byte{0}, widthVersion-verLen)); err != nil {
		return fmt.Errorf("unable to write version padding, %v", err)
	}

	if err := binary.Write(w, binary.LittleEndian, p.Tempo); err != nil {
		return fmt.Errorf("unable to write tempo, %v", err)
	}

	return nil
}

// EncodeTracks encodes the tracks in binary format to the writer.
func (p Pattern) EncodeTracks(w io.Writer) error {
	for _, t := range p.Tracks {
		if err := t.Encode(w); err != nil {
			return fmt.Errorf("unable to write track, %v", err)
		}
	}
	return nil
}

// EncodeHeader encodes the identifier and body size to the writer.
func EncodeHeader(w io.Writer, dataLen int32) error {
	if _, err := w.Write(FileIdentifier); err != nil {
		return fmt.Errorf("unable to write identifier, %v", err)
	}
	if err := binary.Write(w, binary.BigEndian, dataLen); err != nil {
		return fmt.Errorf("unable to write the body size, %v", err)
	}
	return nil
}
