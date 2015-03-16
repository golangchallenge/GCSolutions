package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"strings"
)

var (
	// ErrMagic indicates that the decoded stream lacked the appropriate header
	// magic.
	ErrMagic = errors.New("drum: illegal magic")
	// ErrVersion indicates that the encountered version heading is unknown.
	ErrVersion = errors.New("drum: unknown version")
)

// populateData populates the provided data with the bytes that it reads from
// stream per the endianness assumptions of the .splice format, returning
// whatever error it encounters.  If insufficient bytes are available for data,
// it returns io.ErrUnexpectedEOF.
func populateData(r io.Reader, data interface{}) error {
	// The use of binary.Read as a general-purpose decoder is convenient in terms
	// of boilerplate reduction, but it is semi-time and -memory inefficient due
	// to its use of typeswitching (cf. https://github.com/grpc/grpc-go/pull/87).
	// Were we to use this library in a performance-sensitive context, we would
	// hand decode the stream with io.Reader.
	return binary.Read(r, binary.LittleEndian, data)
}

// byteString converts the null-terminated byte array into an appropriate Go
// string.
func byteString(b []byte) string {
	return strings.TrimRight(string(b), string(0))
}

var magic = []byte("SPLICE")

// decodeHeader decodes the .drum file header from the stream and returns the
// byte size of all the tracks with associated errors.
func decodeHeader(r io.Reader) (length int64, err error) {
	// header models the preface of a .drum file.  It is defined as an anonymous
	// struct because it is neither exported nor desirable to pollute the package
	// namespace and is only used in this singular deserialization routine.
	var header struct {
		Magic   [6]byte // special binary sequence for .drum files.
		Padding [7]byte // apparent padding
		Length  uint8   // length of the patterns after the descriptor is read.
	}
	if err := populateData(r, &header); err != nil {
		return 0, err
	}
	if !bytes.Equal(header.Magic[:], magic) {
		return 0, ErrMagic
	}
	return int64(header.Length), nil
}

// decodeVersion converts the version string to a constant.
func decodeVersion(s string) Version {
	switch Version(s) {
	case V0708Alpha:
		return V0708Alpha
	case V0808Alpha:
		return V0808Alpha
	case V0909:
		return V0909
	default:
		return Unknown
	}
}

// decodePattern extracts a Pattern from the stream.
func decodePattern(r io.Reader) (*Pattern, error) {
	var descriptor struct {
		Version [32]byte
		Tempo   float32
	}
	if err := populateData(r, &descriptor); err != nil {
		return nil, err
	}
	version := byteString(descriptor.Version[:])
	switch version := decodeVersion(version); version {
	case V0708Alpha, V0808Alpha, V0909:
		tracks, err := decodeTracks(r)
		if err != nil {
			return nil, err
		}
		return &Pattern{Tempo: descriptor.Tempo, Version: version, Tracks: tracks}, nil
	case Unknown:
		return nil, ErrVersion
	default:
		panic("unreachable")
	}
}

// decodeTracks consumes the stream until it reaches io.EOF or other error,
// accumulating and emitting all found Track entities.
func decodeTracks(r io.Reader) (ts []Track, _ error) {
	for {
		t, err := decodeTrack(r)
		switch err {
		case nil:
			ts = append(ts, t)
		case io.EOF:
			return
		default:
			return nil, err
		}
	}
}

// decodeTrack incrementally consumes the buffer and emits whatever track
// data can be detected.
func decodeTrack(r io.Reader) (track Track, err error) {
	// trackDescriptor models the a single audio track in a .drum file.  It is
	// defined as an anonymous struct because it is neither exported nor
	// desirable to pollute the package namespace and only used in this
	// deserialization routine.
	var trackDescriptor struct {
		ID         uint8   // track ID
		Padding    [3]byte // apparent padding
		NameLength uint8   // the length of the byte array containing the track name.
	}
	if err = populateData(r, &trackDescriptor); err != nil {
		return track, err
	}
	track.ID = int(trackDescriptor.ID)
	name := make([]byte, trackDescriptor.NameLength)
	if _, err = io.ReadFull(r, name); err != nil {
		return track, err
	}
	track.Name = string(name)
	type (
		// 0 for off; 1 for on; whether or not instrument is active for channel at
		// each step.
		step uint8
		// a string of steps.
		measure [4]step
		// a run of measures.
		trackSequence [4]measure
	)
	const (
		off step = 0
		on       = 1
	)
	var measures trackSequence
	if err = populateData(r, &measures); err != nil {
		return track, err
	}
	for i, m := range measures {
		for j, s := range m {
			switch s {
			case on:
				track.Measures[i][j] = On
			case off:
				track.Measures[i][j] = Off
			default:
				panic("unreachable")
			}
		}
	}
	return track, err
}

// Decode decodes the drum machine stream and returns a pointer to a parsed
// pattern which is the entry point to the rest of the data.
func Decode(r io.Reader) (_ *Pattern, err error) {
	n, err := decodeHeader(r)
	if err != nil {
		return nil, err
	}
	// The .drum file header defines the byte size of all tracks.  We want to
	// limit our scanning of the track data, especially because the legacy file
	// format appears to include garbage or supplemental metadata (unused) at the
	// file's foot (0.708-alpha).  Read what is advertised---no more.
	lr := &io.LimitedReader{R: r, N: n}
	defer func(lr *io.LimitedReader) {
		if err == nil && lr.N != 0 {
			// Explicitly call out the case that not all of the track data was read!
			err = io.ErrUnexpectedEOF
		}
	}(lr)
	return decodePattern(lr)
}

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return Decode(file)
}
