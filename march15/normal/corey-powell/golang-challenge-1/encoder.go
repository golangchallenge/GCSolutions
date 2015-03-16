package drum

import (
	bin "encoding/binary"
	"errors"
	"io"
	"os"
)

// PascalStringTooLong is returned when a provided string is too long to be encoded as a PascalString.
// Pascal strings can be no longer than 256 characters.
var PascalStringTooLong = errors.New("String is too long to be enocded as a PascalString.")

// PaddedStringTooSmall is returned when the provided padding for a PaddedString is too small to hold the
// provided string.
var PaddedStringTooSmall = errors.New("Provided padding is too small for PaddedString.")

// PatternTooLarge is returned when a pattern is larger than 256 bytes, this does not include
// the Magic bytes (SPLICE) and the filesize encoding, only the contents of the Pattern.
var PatternTooLarge = errors.New("Pattern size is too large for splice format.")

type binWriter struct{}

func (e *binWriter) WriteByte(w io.Writer, b byte) error {
	lna := make([]byte, 1)
	lna[0] = b
	if _, err := w.Write(lna); err != nil {
		return err
	}
	return nil
}

func (e *binWriter) WriteInt32(w io.Writer, i int32) error {
	return bin.Write(w, bin.LittleEndian, i)
}

func (e *binWriter) WriteFloat32(w io.Writer, f float32) error {
	return bin.Write(w, bin.LittleEndian, f)
}

func (e *binWriter) WriteString(w io.Writer, s string) error {
	if _, err := w.Write([]byte(s)); err != nil {
		return err
	}
	return nil
}

func (e *binWriter) WritePaddedString(w io.Writer, length int, s string) error {
	pad := length - len(s)
	if pad < 0 {
		return PaddedStringTooSmall
	}
	if err := e.WriteString(w, s); err != nil {
		return err
	}
	if pad > 0 {
		if _, err := w.Write(make([]byte, pad)); err != nil {
			return err
		}
	}
	return nil
}

func (e *binWriter) WritePascalString(w io.Writer, s string) error {
	ln := len(s)
	if ln >= 256 {
		return PascalStringTooLong
	}
	if err := e.WriteByte(w, byte(ln)); err != nil {
		return err
	}
	return e.WriteString(w, s)
}

type Encoder struct {
	w binWriter
}

func (e *Encoder) WriteTrack(w io.Writer, trk *Track) error {
	if err := e.w.WriteInt32(w, trk.ID); err != nil {
		return err
	}
	if err := e.w.WritePascalString(w, trk.Name); err != nil {
		return err
	}
	for _, stp := range trk.Steps {
		if err := e.w.WriteByte(w, stp); err != nil {
			return err
		}
	}
	return nil
}

func (e *Encoder) WritePattern(w io.Writer, p *Pattern) error {
	if err := e.w.WritePaddedString(w, 32, p.Version); err != nil {
		return err
	}
	if err := e.w.WriteFloat32(w, p.Tempo); err != nil {
		return err
	}
	for _, trk := range p.Tracks {
		if err := e.WriteTrack(w, trk); err != nil {
			return err
		}
	}
	return nil
}

func (e *Encoder) CalculatePatternSize(p *Pattern) int {
	// initialize tracksize to 0
	tracksize := 0
	for _, trk := range p.Tracks {
		// sizeof(ID)
		tracksize += 4
		// sizeof(Name)
		tracksize += len(trk.Name)
		// sizeof(Steps) // "normally" 16 bytes
		tracksize += len(trk.Steps)
	}
	// Version String + sizeof(float32) + tracksize
	return 32 + 4 + tracksize
}

func (e *Encoder) WriteSplice(w io.Writer, p *Pattern) error {
	// The magic!
	if err := e.w.WritePaddedString(w, 13, "SPLICE"); err != nil {
		return err
	}
	filesize := e.CalculatePatternSize(p)
	if filesize >= 256 {
		return PatternTooLarge
	}
	if err := e.w.WriteByte(w, byte(filesize)); err != nil {
		return err
	}
	return e.WritePattern(w, p)
}

func NewEncoder() *Encoder {
	return &Encoder{}
}

func EncodeFile(path string, p *Pattern) error {
	e := NewEncoder()
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	return e.WriteSplice(f, p)
}
