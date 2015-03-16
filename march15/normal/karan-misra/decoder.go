package drum

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"os"
	"text/template"
	"unsafe"
)

// Notes
//
// The overall design is dictated by the current limitations in Go's escape
// analysis algorithm. A lot of things escape while in reality they don't. This
// requires for things like allocating and sharing a single buffer. If most buffer
// could be allocated on the stack, the code could have a little cleaner.
//
// Changing the io.ReadFull to io.ReadAtLeast improves performance by around 4%
// but we are leaving it be for cleaner code.
//
// binary.LittleEndian.Uint32() is more efficient than using reflection via
// binary.Read(). So while it leads to slightly more verbose code, the tradeoff
// is worthwhile considering the gains in perf/mem.

// File format
//
// The splice file format for describing a drum pattern is quite simple.
// It consists of a header, followed by atmost 9 tracks of instruments, with each
// track specifying 16 steps of the corresponding instrument. The individual sections
// are described in more detail near their corresponding unmarshal() methods.

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	br := bufio.NewReader(f)

	var p Pattern
	if err := p.unmarshal(br); err != nil {
		return nil, err
	}
	return &p, nil
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Version string
	Tempo   float32

	Tracks Tracks
}

// String returns the pattern formatted using the template:
//
//	Saved with HW Version: {{.Version}}
//	Tempo: {{.Tempo}}
//	{{range .Tracks}}({{.ID}}) {{printf "%s\t" .Instrument}}{{.Steps}}
//	{{end}}
func (p *Pattern) String() string {
	var buf bytes.Buffer

	if err := patternTemplate.Execute(&buf, p); err != nil {
		panic(err)
	}

	return buf.String()
}

const format = `Saved with HW Version: {{.Version}}
Tempo: {{.Tempo}}
{{range .Tracks}}({{.ID}}) {{printf "%s\t" .Instrument}}{{.Steps}}
{{end}}`

var patternTemplate = template.Must(template.New("pattern").Parse(format))

// The pattern file has a header which is laid out like so:
//
//	6   bytes (SPLICE; format string)
//	7   bytes (N/A)
//	1   byte  (size of the rest of the file)
//	32  bytes (version string)
//	4   bytes (IEEE-754 encoded tempo)
//	rest      (tracks data, upto a max of 219 bytes)
func (p *Pattern) unmarshal(r io.Reader) error {
	// Allocate a buf before hand to work around all the
	// Go escape analysis shortfalls.
	// 8 (max buffer space we need) + 32 (see below) = 40
	buf := make([]byte, 40)

	// We are taking away 32 bytes from allocated buffer
	// to server as the buffer when reading the version.
	// This space cannot be reused as we directly use it
	// as a string (see stringView below). Hence the need
	// to reslice buf.
	vbuf := buf[:32]
	buf = buf[32:]

	if _, err := io.ReadFull(r, buf[:headerLen]); err != nil {
		return err
	}

	if bytes.Compare(buf[:headerLen], header) != 0 {
		return errHeaderMissing
	}

	if _, err := io.ReadFull(r, buf[:8]); err != nil {
		return err
	}
	size := int64(buf[7])

	// We only care about size bytes from this point on.
	// Everything in the file beyond that is ignored.
	lr := io.LimitReader(r, int64(size))

	// We are using vbuf which was allocated as part of the main buffer.
	if _, err := io.ReadFull(lr, vbuf); err != nil {
		return err
	}
	// The version string could be a full 32 bytes.
	// If it is not, then terminate it at the full 0 byte.
	i := bytes.IndexByte(vbuf, 0)
	if i < 0 {
		i = len(vbuf)
	}
	p.Version = stringView(vbuf[:i])

	// Calculate the tempo of the pattern.
	if _, err := io.ReadFull(lr, buf[:4]); err != nil {
		return err
	}
	tempoBits := binary.LittleEndian.Uint32(buf[:4])
	p.Tempo = math.Float32frombits(tempoBits)

	if err := p.Tracks.unmarshal(lr, buf); err != nil {
		return err
	}

	return nil
}

var (
	header    = []byte("SPLICE")
	headerLen = len(header)

	errHeaderMissing = errors.New("drum: header missing from file")
)

// Tracks represents a collection of tracks.
//
// We are deliberately not using pointers to Track
// as this allows us to directly allocate all memory
// required for the tracks in one fell swoop.
type Tracks []Track

// Tracks are laid out consecutively.
func (ts *Tracks) unmarshal(r io.Reader, buf []byte) error {
	// The min space a track can take is 22 bytes. Total space
	// available for the tracks is 219 bytes. Therefore preallocating
	// a max of 9 tracks will help avoid doing multiple allocations.
	*ts = make(Tracks, 9)

	var i int

	for {
		// t is a pointer to the current track (at pos i).
		t := &(*ts)[i]
		if err := t.unmarshal(r, buf); err != nil {
			// We use errNoMoreTracks as a sentinel to indicate
			// when we have finished reading all the tracks.
			if err == errNoMoreTracks {
				// We reslice so that we only return the
				// exact num of tracks.
				*ts = (*ts)[:i]
				return nil
			}
			return err
		}

		i++
	}
}

var errNoMoreTracks = errors.New("drum: no more tracks")

// Track represents a single instrument track.
type Track struct {
	ID uint32

	Instrument string
	Steps      Steps

	// This gets allocated with the struct, thus saving
	// us precious allocations PER track. We cannot afford
	// to embed all the space which could potentially be
	// required for a name, hence choosing 16 as a sweet
	// compromise. If the instrument name is > 16 bytes,
	// the unmarshaler below will allocate dynamically.
	nbuf [16]byte
}

// Each track is laid out like so:
//
//	4  bytes (track ID)
//	1  byte  (size of the instrument name)
//	n  bytes (instrument name, where n == val(prev byte))
//	16 bytes (steps for this track)
func (t *Track) unmarshal(r io.Reader, buf []byte) error {
	// Read the track id.
	if _, err := io.ReadFull(r, buf[:4]); err != nil {
		// We check for io.EOF when starting to read a new track
		// as that indicates that we have run out of tracks to read.
		if err == io.EOF {
			return errNoMoreTracks
		}
		return err
	}
	t.ID = binary.LittleEndian.Uint32(buf[:4])

	// Calculate the size of the instrument name.
	if _, err := io.ReadFull(r, buf[:1]); err != nil {
		return err
	}
	size := int(buf[0])

	// If we have enough space in the preallocated
	// array, use that as the buffer for reading the
	// instrument name. Otherwise, allocate a slice
	// with the exact required size.
	var nbuf []byte
	if size > len(t.nbuf) {
		nbuf = make([]byte, size)
	} else {
		nbuf = t.nbuf[:size]
	}

	// Read the instrument name.
	if _, err := io.ReadFull(r, nbuf); err != nil {
		return err
	}
	t.Instrument = stringView(nbuf)

	if err := t.Steps.unmarshal(r); err != nil {
		return err
	}

	return nil
}

// Steps represents the steps for a single instrument track.
type Steps [16]byte

// String returns the steps formatted with the following rules:
//
// * x represents a step where the instrument is active.
// * - represents a step where the instrument is inactive.
//
// Note: | is used as a separator.
func (s Steps) String() string {
	var buf bytes.Buffer

	for i := 0; i < 16; i++ {
		if i%4 == 0 {
			buf.WriteByte(separator)
		}

		switch s[i] {
		case 0:
			buf.WriteByte(inactive)
		case 1:
			buf.WriteByte(active)
		}
	}
	buf.WriteByte(separator)

	return buf.String()
}

const (
	separator = byte('|')
	active    = byte('x')
	inactive  = byte('-')
)

// Steps are represented by 16 consecutive bytes. A 0 value indicates
// a step where the corresponding instrument is inactive. A 1 indicates
// the opposite.
func (s *Steps) unmarshal(r io.Reader) error {
	_, err := io.ReadFull(r, s[:])
	return err
}

type unsafeString struct {
	Data uintptr
	Len  int
}

// stringView returns a view of the []byte as a string.
// In unsafe mode, it doesn't incur allocation and copying caused by conversion.
// In regular safe mode, it is an allocation and copy.
func stringView(v []byte) string {
	x := unsafeString{uintptr(unsafe.Pointer(&v[0])), len(v)}
	return *(*string)(unsafe.Pointer(&x))
}
