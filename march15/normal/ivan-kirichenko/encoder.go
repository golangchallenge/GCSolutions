package drum

import (
	"bytes"
	"fmt"
	"strconv"
)

// PatternEncoder encodes a pattern to string
type PatternEncoder interface {
	Encode(p *Pattern) string
}

// TrackEncoder encodes a track to string
type TrackEncoder interface {
	Encode(t *Track) string
}

// PatternEncoderFunc is adapter to allow the use of ordinary functions as PatternEncoder
type PatternEncoderFunc func(p *Pattern) string

// Encode encodes a pattern to string
func (f PatternEncoderFunc) Encode(p *Pattern) string { return f(p) }

// TrackEncoderFunc is adapter to allow the use of ordinary functions as TrackEncoder
type TrackEncoderFunc func(t *Track) string

// Encode encodes a track to string
func (f TrackEncoderFunc) Encode(t *Track) string { return f(t) }

// DefaultPatternEncoder default implementation of PatternEncoder which can be reset
var DefaultPatternEncoder PatternEncoder = PatternEncoderFunc(TextPatternEncoder)

// DefaultTrackEncoder default implementation of TrackEncoder which can be reset
var DefaultTrackEncoder TrackEncoder = TrackEncoderFunc(TextTrackEncoder)

// TextPatternEncoder encodes a pattern to text
// Example:
//
// Saved with HW Version: 0.708-alpha
// Tempo: 999
// (1) Kick	|x---|----|x---|----|
// (2) HiHat	|x-x-|x-x-|x-x-|x-x-|
func TextPatternEncoder(p *Pattern) string {
	var buf bytes.Buffer
	// version
	buf.WriteString("Saved with HW Version: ")
	buf.WriteString(p.Version)
	buf.WriteByte('\n')

	// tempo
	buf.WriteString("Tempo: ")
	buf.WriteString(fmt.Sprintf("%g", p.Tempo))
	buf.WriteByte('\n')

	// tracks
	for i := range p.tracks {
		buf.WriteString(TextTrackEncoder(&p.tracks[i]))
	}

	return buf.String()
}

// TextTrackEncoder encodes a track to text
// Example:
//
// (2) HiHat	|x-x-|x-x-|x-x-|x-x-|
func TextTrackEncoder(t *Track) string {
	var buf bytes.Buffer

	buf.WriteByte('(')
	buf.WriteString(strconv.FormatUint(uint64(t.ID), 10))
	buf.WriteByte(')')
	buf.WriteByte(' ')

	buf.WriteString(t.Name)
	buf.WriteByte('\t')

	steps := t.Steps
	for i := uint8(0); i < trackSteps; i++ {
		if i%4 == 0 {
			buf.WriteByte('|')
		}

		if steps&1 == 1 {
			buf.WriteByte('x')
		} else {
			buf.WriteByte('-')
		}

		steps = steps >> 1
	}

	buf.WriteByte('|')
	buf.WriteByte('\n')

	return buf.String()
}
