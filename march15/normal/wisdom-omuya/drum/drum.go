// Package drum implements functions for processing .splice drum machine files.
package drum

import (
	"fmt"
	"io"
)

const (
	bigEndian = iota
	littleEndian
)

const (
	// size ofa PatternHeader
	hdrSize = 50
	// null byte
	null = '\x00'
)

// endian specifies the word byte-order convention used.
type endian byte

// Pattern is the high level representation of the drum pattern contained
// in a .splice file.
type Pattern struct {
	// metadata on pattern
	Header *PatternHeader
	// tracks forming this pattern
	Tracks []*Track
}

// PatternHeader holds metadata for a Pattern.
type PatternHeader struct {
	// drum machine hardware company name
	company string
	// size of entire encoded message following
	size int32
	// hardware version used in encoding
	Version string
	// beats per minute used for playback
	Tempo float32
}

// Track stores information on independently programmed tracks for a pattern.
type Track struct {
	Header *TrackHeader
	// track name
	Name string
	// steps for the track
	Steps []byte
	// length of entire track struct
	Size int32
}

// TrackHeader holds metadata for a Track.
type TrackHeader struct {
	// Id of the track
	Id byte
	// length of entire track struct
	Size int32
}

// advanceReader takes an io.Reader and reads up to c bytes - one at a time to
// minimize memory footprint - and returns the number of bytes read and any
// error it hits.
func advanceReader(r io.Reader, c int) (int, error) {
	// placeholder byte slice
	d := make([]byte, 1, 1)
	t := 0
	for i := 0; i < c; i++ {
		n, err := r.Read(d)
		if err != nil {
			return 0, err
		}
		t += n
	}

	// means we didn't read enough padding
	if t != c {
		return t, fmt.Errorf("expected: %vB got: %vB", c, t)
	}

	return t, nil
}

// computePadding checks how many padding bytes are in the PatternHeader -
// as determined by max bytes, m, and read string, r.
func computePadding(m int, r string) int {
	return m /* max bytes available */ - len(r) - 1 /* null byte */
}

// readInt32 reads an int32 from r with the specified endian.
func readInt32(r io.Reader, e endian) (int32, error) {
	b, err := readBytes(r, 4)
	if err != nil {
		return 0, err
	}

	return toInt32(b, e)
}

// readBytes reads s bytes from r into a buffer that's returned.
func readBytes(r io.Reader, s int32) ([]byte, error) {
	b := make([]byte, s, s)
	n, err := r.Read(b)
	if err != nil {
		return nil, err
	}

	// since Read can return less than len(b)
	if int32(n) != s {
		return nil, fmt.Errorf("expected: %vB got: %vB", s, n)
	}

	return b, nil
}

// toInt32 converts b to an int32 integer using specified endian.
func toInt32(b []byte, e endian) (int32, error) {
	if len(b) != 4 {
		return int32(0), fmt.Errorf("got %vB for int32", len(b))
	}

	val := int32(0)

	for i := 0; i < len(b); i++ {
		x := i
		// big-endian convention stores a word's MSB in the smallest address.
		if e == bigEndian {
			x = len(b) - i - 1
		}
		val |= int32(b[x]) << uint(i*8)
	}

	return val, nil
}

// readString reads bytes from r until it hits a null character. Returns the
// string representation of the bytes read with error set if read fails.
func readString(r io.Reader) (string, error) {
	buf := make([]byte, 1)
	acc := []byte{}

	for {
		n, err := r.Read(buf)
		if err != nil {
			return "", err
		}
		if n != 1 {
			return "", fmt.Errorf("no bytes read")
		}
		if buf[0] == null {
			break
		}
		acc = append(acc, buf[0])
	}

	return string(acc), nil
}

// String representation of a Pattern.
func (p *Pattern) String() string {
	str := fmt.Sprintln("Saved with HW Version:", p.Header.Version)
	str += fmt.Sprintln("Tempo:", p.Header.Tempo)
	for _, track := range p.Tracks {
		str += fmt.Sprintf("(%v) %v\t|", track.Header.Id, track.Name)
		for j, step := range track.Steps {
			if step == null {
				str += fmt.Sprintf("-")
			} else {
				// assuming any non-null byte is valid
				str += fmt.Sprintf("x")
			}
			// separate each quarter note with a pipe symbol
			if (j+1)%4 == 0 {
				str += fmt.Sprintf("|")
			}
		}
		str += fmt.Sprintln()
	}

	return str
}
