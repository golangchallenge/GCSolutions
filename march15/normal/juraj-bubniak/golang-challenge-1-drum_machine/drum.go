// Package drum is supposed to implement the decoding of .splice drum machine files.
// See golang-challenge.com/go-challenge1/ for more information
package drum

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"reflect"
	"unicode"
)

const (
	// spliceHeaderVal represents header value.
	spliceHeaderVal = "SPLICE"

	// spliceHeaderBytes is the byte length of the SPLICE header.
	spliceHeaderBytes = 6

	// spliceHeaderStartByte is the starting byte of the SPLICE header inside tracks.
	spliceHeaderStartByte = 83
)

// Error codes returned by failures to decode pattern.
var (
	ErrInvalidHeader      = errors.New("invalid SPLICE header")
	ErrInvalidInterface   = errors.New("interface must be a pointer to struct")
	ErrInvalidID          = errors.New("invalid track ID")
	ErrInvalidStep        = errors.New("invalid track step")
	ErrInvalidDecodeField = errors.New("invalid decode field")
)

// NewDecoder returns a new decoder that reads from r.
func NewDecoder(r io.Reader) *Decoder {
	br := bufio.NewReader(r)
	return &Decoder{br: br}
}

// Decoder decodes patters from an input stream.
type Decoder struct {
	br *bufio.Reader
}

// Decode reads the next pattern value from its
// input and stores it in the value pointed to by v.
func (dec *Decoder) Decode(v interface{}) error {
	version, tempo, tracks, err := dec.decode()
	if err != nil {
		return err
	}

	return dec.set(v, version, tempo, tracks)
}

func (dec *Decoder) set(v interface{}, version string, tempo float64, tracks []*track) error {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return ErrInvalidInterface
	}

	val = val.Elem()
	fversion := val.FieldByName("Version")
	if !fversion.IsValid() && !fversion.CanSet() {
		return ErrInvalidDecodeField
	}

	ftempo := val.FieldByName("Tempo")
	if !ftempo.IsValid() || !fversion.CanSet() {
		return ErrInvalidDecodeField
	}

	ftracks := val.FieldByName("Tracks")
	if !ftracks.IsValid() && !ftracks.CanSet() {
		return ErrInvalidDecodeField
	}

	fversion.SetString(version)
	ftempo.SetFloat(tempo)

	for _, track := range tracks {
		ftracks.Set(reflect.Append(ftracks, reflect.ValueOf(track)))
	}

	return nil
}

func (dec *Decoder) decode() (string, float64, []*track, error) {
	err := dec.readHeader()
	if err != nil {
		return "", 0.0, nil, err
	}

	err = dec.readBlank(8)
	if err != nil {
		return "", 0.0, nil, err
	}

	version, err := dec.readVersion()
	if err != nil {
		return "", 0.0, nil, err
	}

	err = dec.readBlank(18)
	if err != nil {
		return "", 0.0, nil, err
	}

	tempo, err := dec.readTempo()
	if err != nil {
		return "", 0.0, nil, err
	}

	tracks, err := dec.readTracks()
	if err != nil {
		return "", 0.0, nil, err
	}

	return version, tempo, tracks, nil
}

func (dec *Decoder) read() (byte, error) {
	b, err := dec.br.ReadByte()
	if err != nil {
		return 0, err
	}
	return b, nil
}

func (dec *Decoder) unread() error {
	return dec.br.UnreadByte()
}

func (dec *Decoder) readHeader() error {
	var buf bytes.Buffer

	for i := 1; i <= spliceHeaderBytes; i++ {
		b, err := dec.read()
		if err != nil {
			return err
		}
		buf.WriteString(string(b))
	}

	if buf.String() != spliceHeaderVal {
		return ErrInvalidHeader
	}

	return nil
}

func (dec *Decoder) readVersion() (string, error) {
	var buf bytes.Buffer

	// version should take 14 bytes
	for i := 0; i < 14; i++ {
		b, err := dec.read()
		if err != nil {
			return "", err
		}

		r := rune(b)
		if isVersionChar(r) {
			buf.WriteString(string(b))
		}
	}

	return buf.String(), nil
}

func (dec *Decoder) readTempo() (float64, error) {
	data := make([]byte, 4)

	// tempo should take 4 bytes
	for i := 0; i < 4; i++ {
		b, err := dec.read()
		if err != nil {
			return 0.0, err
		}
		data[i] = b
	}

	return bytesToFloat(data), nil
}

func (dec *Decoder) readTrackStates() (steps, error) {
	steps := make(steps, 16)

	// there should be 16 steps stores in 16 bytes
	for i := 0; i < 16; i++ {
		b, err := dec.read()
		if err != nil {
			return nil, err
		}

		if b == 1 {
			steps[i] = newStep(true)
		} else if b == 0 {
			steps[i] = newStep(false)
		} else {
			return nil, ErrInvalidStep
		}
	}

	return steps, nil
}

func (dec *Decoder) readTrackID() (int, error) {
	b, err := dec.read()
	if err != nil {
		return 0, err
	}

	// not sure about this range check
	// but there are only values between 0 and 255 in tests
	id := int(b)
	if 0 <= id && id <= 255 {
		return id, nil
	}

	return 0, ErrInvalidID
}

func (dec *Decoder) readTrackName() (string, error) {
	var buf bytes.Buffer

	for {
		b, err := dec.read()
		if err != nil {
			return "", err
		}

		r := rune(b)
		if !isNameChar(r) {
			err := dec.unread()
			if err != nil {
				return "", err
			}
			break
		}

		buf.WriteString(string(b))
	}

	return buf.String(), nil
}

func (dec *Decoder) readTracks() ([]*track, error) {
	var tracks []*track

	for {
		id, err := dec.readTrackID()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		// turns out there can be SPLICE header in the tracks
		// we need to handle this case
		if id == spliceHeaderStartByte {
			err := dec.checkSplice()
			if err != nil {
				return nil, err
			}

			break
		}

		err = dec.consumeBlank()
		if err != nil {
			return nil, err
		}

		name, err := dec.readTrackName()
		if err != nil {
			return nil, err
		}

		steps, err := dec.readTrackStates()
		if err != nil {
			return nil, err
		}

		tracks = append(tracks, newTrack(id, name, steps))
	}

	return tracks, nil
}

func (dec *Decoder) checkSplice() error {
	err := dec.unread()
	if err != nil {
		return err
	}

	var buf bytes.Buffer

	for i := 1; i <= spliceHeaderBytes; i++ {
		b, err := dec.read()
		if err != nil {
			return err
		}
		buf.WriteString(string(b))
	}

	if buf.String() != spliceHeaderVal {
		return ErrInvalidHeader
	}

	return nil
}

func (dec *Decoder) readBlank(n int) error {
	for i := 0; i < n; i++ {
		_, err := dec.read()
		if err != nil {
			return err
		}
	}
	return nil
}

func (dec *Decoder) consumeBlank() error {
	for {
		b, err := dec.read()
		if err != nil {
			return err
		}
		if b != 0 {
			break
		}
	}
	return nil
}

// bytesToFloat converts byte array to 32 bit float.
func bytesToFloat(data []byte) float64 {
	bits := binary.LittleEndian.Uint32(data)
	return float64(math.Float32frombits(bits))
}

// isNameChar returns true for valid name character.
func isNameChar(r rune) bool {
	if !unicode.IsLetter(r) && !unicode.IsPunct(r) && !unicode.IsSpace(r) {
		return false
	}
	return true
}

// isVersionChar returns true for valid version character.
func isVersionChar(r rune) bool {
	if !unicode.IsDigit(r) && !unicode.IsLetter(r) && !unicode.IsPunct(r) {
		return false
	}
	return true
}
