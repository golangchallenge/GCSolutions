package drum

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"os"
)

// magic is the file format specifier at the beginning of each drum file
var magic = []byte("SPLICE")

// zero is a single \x00 used for searching without allocation.
var zero = []byte{0}

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
//
// The decoder makes no judgement about the names, tempo, ids, or tracks. For
// example: a negative tempo is allowed; null characters and un-unicode
// characters are allowed in names; and the music could be bad. That's not the
// job of the decoder to diagnose.
//
// Extra data is ignored past the end of the tracks.
//
// The returned pattern is nil iff the returned error != nil
func DecodeFile(path string) (*Pattern, error) {
	file, err := os.Open(path)
	if file != nil {
		defer file.Close()
	}

	if err != nil {
		return nil, err
	}

	return parse(bufio.NewReaderSize(file, 512))
}

// parse drives parsing a drum file. See the parser type for details.
func parse(src io.Reader) (*Pattern, error) {
	pattern := new(Pattern)

	// The size will be adjusted after reading the preamble. To avoid
	// corner-cases around EOF, it is initially set much larger than the
	// preamble.
	parse := &parser{src: io.LimitedReader{R: src, N: 32}}

	parse.src.N = parse.preamble()
	pattern.HardwareVersion = parse.hardwareVersion()
	pattern.Tempo = parse.tempo()

	var tracks []*Track
	for parse.src.N > 0 && parse.err == nil {
		tracks = append(tracks, parse.track())
	}
	pattern.Tracks = tracks

	// Readers are allowed to return EOF on their last read. To account for
	// this, we ignore EOF unless we needed more data, in which case it is
	// converted to ErrUnexpectedEOF
	if parse.err != nil && parse.err != io.EOF {
		return nil, parse.err
	}

	return pattern, nil
}

// parser for the drum format
//
// The parser keeps the first error, which should be checked in the end. See
// https://blog.golang.org/errors-are-values. Practically, in the event of an
// error a little of the parser state is filled with garbage, but is then
// discarded.
//
// The parser is deliberately loose. Specifically, it ignores spaces in the
// file where anything can go; and converts ANY non-zero beat to a full-on
// beat. This could be useful for extensions to the drum file format.
type parser struct {
	src io.LimitedReader
	err error

	// This buffer is used for all parsing to reduce allocations. Basically,
	// the only extraneous allocations are used to convert names from bytes to
	// strings. Names can be up to 256 bytes long, mostly they're in the 4-16
	// range. Everything else is much less than this; 4 bytes for nubmers and
	// 16 bytes for tracks.
	buf [256]byte
}

// preamble gets the magic word in the file and returns the length of the rest
// of the input.
func (p *parser) preamble() int64 {
	p.fillBuffer(p.buf[:14])
	p.confirm(bytes.Equal(p.buf[:6], magic), "unknown message format")

	// the bytes in buf[6:13] are ignored.

	sz := int64(p.buf[13])
	p.confirm(sz >= 36, "need enough bytes for hardware version and tempo")

	return sz
}

// fillBuffer reads into the buffer unless an error has already occurred.
// If the read fails, the error is stored, essentially stopping the parse
func (p *parser) fillBuffer(buf []byte) {
	if p.err == nil {
		_, p.err = io.ReadFull(&p.src, buf)
	} else if p.err == io.EOF {
		p.err = io.ErrUnexpectedEOF
	}
}

// confirm moves the parser to the error state if the condition doesn't hold.
// The error uses the given message.
func (p *parser) confirm(ok bool, msg string) {
	if !ok && p.err == nil {
		p.err = errors.New(msg)
	}
}

// hardwareVersion parses the said name.
//
// A hardware version is a zero-padded name. An empty hardware version is
// allowed. The parser deliberately ignores anything after the initial
// null-terminator (if there is one).
func (p *parser) hardwareVersion() string {
	buf := p.buf[:32]
	p.fillBuffer(buf)

	sz := bytes.Index(buf, zero)
	if sz < 0 {
		sz = 32
	}

	return string(buf[:sz])
}

// tempo reads the floating-point tempo.
//
// It's possible this gets a negative tempo. If that's what's stored on disk, its
// not our problem: maybe it has special meaning for the play-back?
func (p *parser) tempo() float32 {
	p.fillBuffer(p.buf[:4])
	return parseFloat32(p.buf[:4])
}

// track parses and returns an entire track.
func (p *parser) track() *Track {
	t := new(Track)

	p.fillBuffer(p.buf[:5])
	t.ID = parseUint32(p.buf[:4])

	// The name is a length-prefixed string. the length is one byte, leading to
	// a maximum of 256 characters. This parser deliberately ignores the
	// contents of the name because the exact format is unknown.
	sz := uint8(p.buf[4])
	p.fillBuffer(p.buf[:sz])
	t.Name = string(p.buf[:sz])

	// The steps are bytes. Future encoders may put fanciness in there, this
	// decoder only cares about on and off
	p.fillBuffer(p.buf[:16])
	for i := 0; i < 16; i++ {
		if p.buf[i] != 0 {
			t.Steps[i] = true
		}
	}

	return t
}

// parseUint32 gets a little-endian uint32 from the beginning of the buffer.
func parseUint32(buf []byte) uint32 {
	return binary.LittleEndian.Uint32(buf)
}

// parseFloat32 gets a little-endian float32 from the beginning of the buffer.
func parseFloat32(buf []byte) float32 {
	// It would be *easy* to read it via binary.Read(...), but that uses
	// reflection. Boo. It's cheaper to do it ourselves, and practically the
	// same amount of code.
	return math.Float32frombits(parseUint32(buf))
}
