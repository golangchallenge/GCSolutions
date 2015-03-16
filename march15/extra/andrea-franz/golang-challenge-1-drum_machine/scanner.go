package drum

import (
	"encoding/binary"
	"fmt"
	"io"
	"strings"
)

const (
	validSignature  = "SPLICE"
	signatureLength = 13
	versionLength   = 32
	stepsLength     = 16
	maxValidTrackID = 9999
)

// InvalidSignatureError is returned when the stream decoded starts with a signature
// different from "SPLICE"
type InvalidSignatureError struct {
	signature string
}

func (e InvalidSignatureError) Error() string {
	return fmt.Sprintf("invalid signature `%s`", e.signature)
}

// InvalidTrackIDError is returned if a Track ID is greater than 9999
type InvalidTrackIDError struct {
	ID int32
}

func (e InvalidTrackIDError) Error() string {
	return fmt.Sprintf("invalid track ID `%d`", e.ID)
}

type scanFn func() (scanFn, error)

type scanner struct {
	r     io.Reader
	p     *Pattern
	state scanFn
}

func newScanner(r io.Reader, p *Pattern) *scanner {
	return &scanner{r: r, p: p}
}

func (s *scanner) scanSignature() (scanFn, error) {
	var b [signatureLength]byte
	err := binary.Read(s.r, binary.LittleEndian, &b)
	if err != nil {
		return nil, err
	}

	bs := fmt.Sprintf("%s", b)
	signature := strings.Trim(string(bs), "\x00")
	if signature != validSignature {
		return nil, InvalidSignatureError{signature: signature}
	}

	s.p.Header.Signature = signature

	return s.scanX, nil
}

func (s *scanner) scanX() (scanFn, error) {
	var b byte
	err := binary.Read(s.r, binary.LittleEndian, &b)
	if err != nil {
		return nil, err
	}

	s.p.Header.X = b

	return s.scanVersion, nil
}

func (s *scanner) scanVersion() (scanFn, error) {
	var b [versionLength]byte
	err := binary.Read(s.r, binary.LittleEndian, &b)
	if err != nil {
		return nil, err
	}

	version := fmt.Sprintf("%s", b)
	version = strings.Trim(version, "\x00")

	s.p.Header.Version = version

	return s.scanTempo, nil
}

func (s *scanner) scanTempo() (scanFn, error) {
	var tempo float32
	err := binary.Read(s.r, binary.LittleEndian, &tempo)
	if err != nil {
		return nil, err
	}

	s.p.Header.Tempo = tempo

	return s.scanTrack, nil
}

func (s *scanner) scanTrack() (scanFn, error) {
	for {
		t := &Track{}
		ts := newTrackScanner(s.r, t)
		err := ts.unmarshal(t)

		switch err.(type) {
		case InvalidTrackIDError:
			return nil, nil
		case nil:
			s.p.AddTrack(t)
		default:
			return nil, err
		}
	}
}

func (s *scanner) unmarshal(p *Pattern) error {
	for s.state = s.scanSignature; s.state != nil; {
		next, err := s.state()
		if err == io.EOF {
			return nil
		}

		if err != nil {
			return err
		}

		s.state = next
	}

	return nil
}

type trackscanner struct {
	r     io.Reader
	t     *Track
	state scanFn
}

func newTrackScanner(r io.Reader, t *Track) *trackscanner {
	return &trackscanner{r: r, t: t}
}

func (ts *trackscanner) unmarshal(t *Track) error {
	for ts.state = ts.scanID; ts.state != nil; {
		next, err := ts.state()
		if err != nil {
			return err
		}

		ts.state = next
	}

	return nil
}

func (ts *trackscanner) scanID() (scanFn, error) {
	var ID int32
	err := binary.Read(ts.r, binary.LittleEndian, &ID)
	if err != nil {
		return nil, err
	}

	if ID > maxValidTrackID {
		return nil, InvalidTrackIDError{ID: ID}
	}

	ts.t.ID = ID

	return ts.scanInstrument, nil
}

func (ts *trackscanner) scanInstrument() (scanFn, error) {
	var length byte
	err := binary.Read(ts.r, binary.LittleEndian, &length)
	if err != nil {
		return nil, err
	}

	instrument := make([]byte, length)
	err = binary.Read(ts.r, binary.LittleEndian, &instrument)
	if err != nil {
		return nil, err
	}

	ts.t.Instrument = string(instrument)

	return ts.scanSteps, nil
}

func (ts *trackscanner) scanSteps() (scanFn, error) {
	var steps [stepsLength]byte
	err := binary.Read(ts.r, binary.LittleEndian, &steps)
	if err != nil {
		return nil, err
	}

	ts.t.Steps = steps

	return nil, nil
}
