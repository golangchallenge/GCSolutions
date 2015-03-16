package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

// A Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Tempo   float32 // Tempo in beats per minute (bpm)
	Version string  // Software version that saved the track
	Tracks  []Track // Will contain individual tracks
}

// This will be returned, if the magic value "SPLICE" isn't found in the beginning
// of a pattern file.
var ErrMagicCode = errors.New("magic code for splice does not match")

// This error is returned if the file is too small to be decoded.
// It is likely the file has been truncated.
var ErrFileTooSmall = errors.New("file is too small to contain tracks")

// ErrPayloadTooBig will be returned if the payload size of the file is considered
// too big. The value is set to 2^24 bytes, which allows for patterns up to 16MB.
var ErrPayloadTooBig = errors.New("payload size too big")

// Set a fixed maximum payload size (1 << 24 = 16MB)
// This is set to prevent excessive memory allocations on corrupt files.
const maxPayloadSize = 1 << 24

// String will return a printable string
// representation of the Pattern
func (p Pattern) String() string {
	tempo := strconv.FormatFloat(float64(p.Tempo), 'f', -1, 32)
	var tracks = make([]string, 0, len(p.Tracks))
	for _, track := range p.Tracks {
		tracks = append(tracks, track.String())
	}
	return fmt.Sprintf("Saved with HW Version: %s\nTempo: %s\n%s\n",
		p.Version, tempo, strings.Join(tracks, "\n"))
}

// Decode will decode a pattern into p.
// On succcessful decode, p will contain the information from the file,
// and the reader will be placed after the last byte of the payload.
// If an error occurs it will be returned, and p may contain partial data
// that has been read.
func (p *Pattern) Decode(r io.Reader) error {
	// Decode the header, so we can grab the payload.
	header := newSpliceHeader()
	err := header.Decode(r)
	if err != nil {
		return err
	}

	// Check that payload size is reasonable. Set in constant 'maxPayloadSize'.
	if header.Payload > maxPayloadSize {
		return ErrPayloadTooBig
	}

	// Read the payload into a new buffer.
	// We want to be sure we don't read beyond the payload size.
	payload := make([]byte, header.Payload)
	err = binary.Read(r, binary.BigEndian, &payload)
	if err != nil {
		return err
	}
	b := bytes.NewBuffer(payload)

	// Decode the pattern header.
	ph := newPatternHeader()
	err = ph.Decode(b)
	if err != nil {
		return err
	}
	p.Tempo = ph.Tempo
	p.Version = strings.Trim(string(ph.Version[:]), "\000")

	// Read tracks until we have no more content in the buffer.
	for {
		track, err := decodeTrack(b)
		// If we got an error that isn't an indication of EOF, return it.
		if err != nil && err != io.EOF {
			return err
		}

		// If we get EOF we reached the end of the payload.
		// This is an expected error, and indicates that we
		// have read all tracks.
		if err == io.EOF {
			break
		}

		// We got a track.
		if track != nil {
			p.Tracks = append(p.Tracks, *track)
		}
	}
	return nil
}

// Splice Header, used for internal decoding.
// Must match the binary format.
type spliceHeader struct {
	Magic   [6]byte // Magic value, identifying the splice format. Should be "SPLICE"
	Padding uint32  // Most likely padding, set to 0.
	Payload uint32  // Size of the pattern data.
}

// Returns a new Splice header, with fields unset.
func newSpliceHeader() *spliceHeader {
	return &spliceHeader{}
}

// Decode the content of the reader into the spliceHeader
// struct. The function will check if the magic word matches.
// The reader will be forwarded to the end of the splice header.
// If a read error occurs, or the magic header bytes mismatch an error is returned.
func (s *spliceHeader) Decode(r io.Reader) error {
	// Read binary into s
	err := binary.Read(r, binary.BigEndian, s)
	if err != nil {
		return err
	}
	// Check if we get the magic we expected.
	if string(s.Magic[:]) != "SPLICE" {
		return ErrMagicCode
	}
	return nil
}

// Header for a pattern.
// This is the part of the header that is first part of the payload.
// Must match the binary format.
type patternHeader struct {
	Version [32]byte
	Tempo   float32
}

// Returns a new Splice header, with fields unset.
func newPatternHeader() *patternHeader {
	return &patternHeader{}
}

// Decode the pattern header after the payload size has been given.
func (p *patternHeader) Decode(r io.Reader) error {
	return binary.Read(r, binary.LittleEndian, p)
}

// PlayAt returns the instruments to play at beat 'x'.
// Individual tracks will loop once they have finished
// playing. If x is negative nothing will be returned.
func (p Pattern) PlayAt(x int) []Instrument {
	if x < 0 {
		return nil
	}
	var ins = make([]Instrument, 0, len(p.Tracks))
	for _, track := range p.Tracks {
		if track.At(x) != StepNothing {
			ins = append(ins, track.Instrument)
		}
	}
	return ins
}

// Duration returns the duration of the pattern.
// If the length of the individual tracks diverge, the longest
// duration is returned.
func (p Pattern) Duration() time.Duration {
	return time.Duration(p.LongestTrack()) * p.BeatDuration()
}

// BeatDuration returns the duration of a single beat.
func (p Pattern) BeatDuration() time.Duration {
	return time.Duration(int64(float64(time.Minute) / float64(p.Tempo)))
}

// LongestTrack returns the number of beats of the longest track.
// If no tracks are present, 0 is returned.
func (p Pattern) LongestTrack() int {
	var maxBeats int
	for _, track := range p.Tracks {
		if track.Len() > maxBeats {
			maxBeats = track.Len()
		}
	}
	return maxBeats
}
