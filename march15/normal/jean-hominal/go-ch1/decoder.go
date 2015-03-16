package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (p *Pattern, err error) {
	f, err := os.Open(path)
	if err != nil {
		err = fmt.Errorf("Error opening %v: %v", path, err)
		return
	}
	defer f.Close()

	p, err = decode(f)
	if err != nil {
		err = fmt.Errorf("Error reading %v: %v", path, err)
	}

	return
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	hwVersion string
	tempo     float32
	tracks    []track
}

// track represents a single track in the drum
// pattern contained in a .splice file.
type track struct {
	id    uint32
	name  string
	beats [16]bool
}

func (p *Pattern) String() string {
	if p == nil {
		return ""
	}

	var buf bytes.Buffer

	fmt.Fprintf(&buf, "Saved with HW Version: %v\nTempo: %g\n", p.hwVersion, p.tempo)
	for _, t := range p.tracks {
		fmt.Fprintf(&buf, "(%d) %v\t", t.id, t.name)

		for j, b := range t.beats {
			if j%4 == 0 {
				buf.WriteByte('|')
			}
			if b {
				buf.WriteByte('x')
			} else {
				buf.WriteByte('-')
			}
		}
		buf.WriteByte('|')
		buf.WriteByte('\n')
	}

	return buf.String()
}

type decoder struct {
	r        io.Reader
	buf      []byte
	hasLimit bool
	limit    uint64
	err      error
}

func (d *decoder) ensureBufferCapacity(l int) {
	if cap(d.buf) < l {
		d.buf = make([]byte, l)
	}
}

func (d *decoder) Read(p []byte) (n int, err error) {
	switch {
	case d.err != nil:
		return 0, d.err
	case d.hasLimit && d.limit == 0:
		d.err = io.EOF
		return 0, d.err
	case d.hasLimit && d.limit < uint64(len(p)):
		p = p[:int(d.limit)]
	}
	n, d.err = d.r.Read(p)
	if d.hasLimit {
		d.limit -= uint64(n)
	}
	return n, d.err
}

func (d *decoder) readBytes(l int) []byte {
	if d.err != nil {
		return nil
	}
	d.ensureBufferCapacity(l)
	io.ReadFull(d, d.buf[:l])
	return d.buf[:l]
}

func (d *decoder) readString(l int) string {
	return string(bytes.TrimRight(d.readBytes(l), "\x00"))
}

func decode(r io.Reader) (p *Pattern, err error) {
	d := decoder{r: r}
	// Allocate a relatively large buffer, to reuse in the whole function
	d.ensureBufferCapacity(32)

	// First part in the file is a magic header, that should have the SPLICE value;
	magicHeader := d.readString(6)

	// Second part of the file is the size of the data splice
	var size uint64
	binary.Read(&d, binary.BigEndian, &size)

	switch {
	case d.err == io.EOF || d.err == io.ErrUnexpectedEOF:
		err = errors.New("Incomplete splice header.")
	case d.err != nil:
		err = fmt.Errorf("Error reading splice header: %v", d.err)
	case magicHeader != "SPLICE":
		err = errors.New("Input does not start with 'SPLICE' magic header.")
	case size < 36:
		err = fmt.Errorf("Declared data frame size %v is too small.", size)
	}

	if err != nil {
		return
	}

	// The frame size is known, so set the limit.
	d.limit = size - 4
	d.hasLimit = true

	// Read the version string (null-terminated ASCII)
	hwVersion := d.readString(32)

	// Now read the tempo
	var tempo float32
	binary.Read(r, binary.LittleEndian, &tempo)

	// Now decode the tracks
	var tracks []track
	for d.err == nil && d.limit > 0 {
		// First, the track ID
		var tID uint32
		binary.Read(&d, binary.LittleEndian, &tID)

		// Second, the name (length byte + name)
		var l byte
		binary.Read(&d, binary.LittleEndian, &l)
		tName := d.readString(int(l))

		// Change error code of EOF so that incomplete lines are detected.
		if d.err == io.EOF {
			d.err = io.ErrUnexpectedEOF
		}

		// Third, the beats
		beatChars := d.readBytes(16)
		var tBeats [16]bool
		for i, b := range beatChars {
			switch b {
			case 0:
				tBeats[i] = false
			case 1:
				tBeats[i] = true
			default:
				err = fmt.Errorf("Invalid track byte value: %v", b)
				return
			}
		}

		tracks = append(tracks, track{tID, tName, tBeats})
	}
	switch {
	case d.limit > 0 && (d.err == io.EOF || d.err == io.ErrUnexpectedEOF):
		err = fmt.Errorf("Incomplete file, %v bytes missing from frame.", d.limit)
	case d.limit == 0 && (d.err == io.ErrUnexpectedEOF):
		err = fmt.Errorf("Incomplete data frame, header has the wrong size.")
	case d.err != io.EOF:
		err = d.err
	}

	if err != nil {
		return
	}

	p = &Pattern{hwVersion, tempo, tracks}
	return
}
