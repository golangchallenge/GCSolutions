package drum

import (
	"bytes"
	bin "encoding/binary"
	"errors"
	"io"
	"os"
)

// IncorrectMagicBytes is returned when the first 6 bytes of a given splice stream is not equal to SPLICE
var IncorrectMagicBytes = errors.New("Wrong Magic bytes (expected SPLICE).")

// EmptyPatternErr is returned when you managed to have a splice with null/zero filesize.
var EmptyPatternErr = errors.New("What kind of sick joke is this, this splice is empty!")

// I'm dead sure Go has something for checking the length of null terminated strings.
// I'm just really bad at googling the solution, so HERE:
func cstringlen(chars []byte) int {
	l := bytes.IndexByte(chars, 0)
	if l >= 0 {
		return l
	}
	return len(chars)
}

type binReader struct{}

func (d *binReader) ReadUint32(f io.Reader) (uint32, error) {
	var i uint32
	if err := bin.Read(f, bin.LittleEndian, &i); err != nil {
		return 0, err
	}
	return i, nil
}

func (d *binReader) ReadFloat32(f io.Reader) (float32, error) {
	var flt float32
	if err := bin.Read(f, bin.LittleEndian, &flt); err != nil {
		return 0.0, err
	}
	return flt, nil
}

func (d *binReader) ReadBytes(f io.Reader, length int) ([]byte, error) {
	i := make([]byte, length)
	if _, err := f.Read(i); err != nil {
		return nil, err
	}
	return i, nil
}

func (d *binReader) ReadByte(f io.Reader) (byte, error) {
	var i byte
	if err := bin.Read(f, bin.LittleEndian, &i); err != nil {
		return 0, err
	}
	return i, nil
}

func (d *binReader) ReadPascalString(f io.Reader) (int, string, error) {
	ln, err := d.ReadByte(f)
	if err != nil {
		return 0, "", err
	}
	str, err := d.ReadBytes(f, int(ln))
	if err != nil {
		return 0, "", err
	}
	return int(ln), string(str), nil
}

func (d *binReader) Seek(f io.Reader, length int) error {
	if _, err := d.ReadBytes(f, length); err != nil {
		return err
	}
	return nil
}

type TrackDecoder struct {
	r binReader
}

func (d *TrackDecoder) Decode(f io.Reader) (*Track, error) {
	trk := &Track{}
	// The first 4 bytes (encoded as a uint32 or just int) is the ID of the track.
	id, err := d.r.ReadUint32(f)
	if err != nil {
		return nil, err
	}
	// The next variable number of bytes is a PascalString representing the name
	_, name, err := d.r.ReadPascalString(f)
	if err != nil {
		return nil, err
	}
	// The next 16 bytes (the length of the trk.Steps) are the steps.
	stps, err := d.r.ReadBytes(f, len(trk.Steps))
	if err != nil {
		return nil, err
	}
	trk.ID = int32(id)
	trk.Name = string(name)
	copy(trk.Steps[:], stps)
	return trk, nil
}

func NewTrackDecoder() *TrackDecoder {
	return &TrackDecoder{}
}

type PatternDecoder struct {
	r binReader
}

func (d *PatternDecoder) readVersionString(f io.Reader) (string, error) {
	// Version string
	exprt, err := d.r.ReadBytes(f, 32)
	if err != nil {
		return "", err
	}
	// one thing, this is a null terminated string, so, we need to trim it up by finding the length.
	return string(exprt[0:cstringlen(exprt)]), nil
}

func (d *PatternDecoder) readTempo(f io.Reader) (float32, error) {
	return d.r.ReadFloat32(f)
}

func (d *PatternDecoder) readTrack(f io.Reader) (*Track, error) {
	return NewTrackDecoder().Decode(f)
}

func (d *PatternDecoder) Decode(f io.Reader) (*Pattern, error) {
	p := &Pattern{}
	// Version string
	version, err := d.readVersionString(f)
	if err != nil {
		return nil, err
	}
	p.Version = version
	tempo, err := d.readTempo(f)
	if err != nil {
		return nil, err
	}
	p.Tempo = tempo
	// just keep nabbing data until the end of the file.
	for {
		trk, err := d.readTrack(f)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return nil, err
			}
		}
		p.Tracks = append(p.Tracks, trk)
	}
	return p, nil
}

func NewPatternDecoder() *PatternDecoder {
	return &PatternDecoder{}
}

type SpliceDecoder struct {
	r binReader
}

func (d *SpliceDecoder) readMagicBytes(f io.Reader) (string, error) {
	// now, we expect the magic to be SPLICE (6 bytes)
	magicBytes, err := d.r.ReadBytes(f, 6)
	if err != nil {
		return "", err
	}
	magic := string(magicBytes)
	// validate that the magic is SPLICE, and not something else
	if magic != "SPLICE" {
		// hey you, don't judge my error messages, I did/do ruby everyday.
		return "", IncorrectMagicBytes
	}
	return magic, nil
}

func (d *SpliceDecoder) readFilesize(f io.Reader) (int, error) {
	// now we nab the filesize, which, is just an unit8 (well you guessed it from the line below),
	// so from that we can say that contents can be no larger than 256 bytes, that or someone encoded this as a BigEndian!
	// . Assuming filesize is 1 byte
	fsz, err := d.r.ReadByte(f)
	if err != nil {
		return 0, err
	}
	filesize := int(fsz)
	if filesize <= 0 {
		return 0, EmptyPatternErr
	}
	return filesize, nil
}

func (d *SpliceDecoder) readPattern(f io.Reader) (*Pattern, error) {
	return NewPatternDecoder().Decode(f)
}

func (d *SpliceDecoder) Decode(f io.Reader) (*Pattern, error) {
	if _, err := d.readMagicBytes(f); err != nil {
		return nil, err
	}
	d.r.Seek(f, 7) // skip padding
	if _, err := d.readFilesize(f); err != nil {
		return nil, err
	}
	return d.readPattern(f)
}

func NewSpliceDecoder() *SpliceDecoder {
	return &SpliceDecoder{}
}

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	d := NewSpliceDecoder()
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return d.Decode(f)
}
