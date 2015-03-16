package drum

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return newDecoder(f).decode()
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	HWVersion string
	Tempo     float32
	Tracks    []Track
}

// Track is the high level representation of an indivdual
// track within a drum pattern
type Track struct {
	ID      uint32
	Name    string
	Measure [16]byte
}

func (p *Pattern) String() string {
	var buf bytes.Buffer
	format := `Saved with HW Version: %[1]s
Tempo: %.[3]*[2]f
`
	tPrec := 0
	if float32(int(p.Tempo)) != p.Tempo {
		tPrec = 1
	}
	buf.WriteString(fmt.Sprintf(format, p.HWVersion, p.Tempo, tPrec))
	for _, t := range p.Tracks {
		if t.Name == "" {
			continue
		}
		buf.WriteString(fmt.Sprintf("(%[1]d) %[2]s\t|", t.ID, t.Name))
		for q := 0; q < 4; q++ { // 4 quarters in a measure
			for s := 0; s < 4; s++ { // 4 sixteenths in a quarter
				if t.Measure[(q*4)+s] == 1 {
					buf.WriteByte('x')
				} else {
					buf.WriteByte('-')
				}
			}
			buf.WriteByte('|')
		}
		buf.WriteByte('\n')
	}
	return buf.String()
}

// A decoder holds the decode state for a particular drum machine file being decoded.
// it uses a deferred failure mechanism, where the first failure will be recorded
// and subsequent decode actions will be no-ops. The first recorded error will be returned
// from the close() method.
type decoder struct {
	p    Pattern
	err  error
	buf  []byte // a re-usable temporary decoding buffer
	r    *bufio.Reader
	pos  byte // current position in the data stream
	rlen byte // length of record
}

func newDecoder(r io.Reader) *decoder {
	return &decoder{buf: make([]byte, 255), r: bufio.NewReader(r)}
}
func (d *decoder) decode() (*Pattern, error) {
	d.decodeHeader()
	for !d.done() {
		d.decodeTrack()
	}
	return &d.p, d.close()
}

var boilerplate = []byte{'S', 'P', 'L', 'I', 'C', 'E', 0, 0, 0, 0, 0, 0, 0}

const hwVersionLen = 32

func (d *decoder) decodeHeader() {
	buf := d.buf[:len(boilerplate)]
	n, err := d.r.Read(buf)
	if n != len(buf) || err != nil {
		d.err = fmt.Errorf("couldn't read SPLICE header")
		return
	}
	if bytes.Compare(buf, boilerplate) != 0 {
		d.err = fmt.Errorf("boilerplate header didn't match. Read: %v, expected: %v", buf, boilerplate)
		return
	}
	d.rlen, err = d.r.ReadByte()
	if err != nil {
		d.err = fmt.Errorf("couldn't read record length, err: %s", err)
		return
	}
	buf = d.buf[:hwVersionLen]
	n, err = d.r.Read(buf)
	if n != hwVersionLen || err != nil {
		d.err = fmt.Errorf("couldn't read hwVersion string. Read %d bytes, expected %d, err: %s", n, hwVersionLen, err)
		return
	}
	if i := bytes.IndexByte(buf, '\x00'); i != -1 {
		buf = buf[:i]
	}
	d.p.HWVersion = string(buf)
	d.pos += hwVersionLen
	err = binary.Read(d.r, binary.LittleEndian, &d.p.Tempo)
	if err != nil {
		d.err = fmt.Errorf("couldn't read hw version, err: %s", err)
		return
	}
	d.pos += 4
}

func (d *decoder) decodeTrack() {
	if d.err != nil {
		return
	}
	t := Track{}
	err := binary.Read(d.r, binary.LittleEndian, &t.ID)
	if err != nil {
		d.err = fmt.Errorf("couldn't read track ID for track %d, err: %s", len(d.p.Tracks)+1, err)
		return
	}
	d.pos += 4
	nameLen, err := d.r.ReadByte()
	if err != nil {
		d.err = fmt.Errorf("couldn't read track name length for track %d, err: %s", len(d.p.Tracks)+1, err)
		return
	}
	if int(nameLen) > cap(d.buf) {
		d.err = fmt.Errorf("track name too long for track %d. %d found, max %d", len(d.p.Tracks)+1, nameLen, cap(d.buf))
		return
	}
	d.pos++
	buf := d.buf[:nameLen]
	n, err := d.r.Read(buf)
	if n != int(nameLen) || err != nil {
		d.err = fmt.Errorf("couldn't read track name for track %d. Read %d bytes, expected %d, err: %s", len(d.p.Tracks)+1, n, nameLen, err)
		return
	}
	t.Name = string(buf)
	d.pos += nameLen
	n, err = d.r.Read(t.Measure[:])
	if n != len(t.Measure) || err != nil {
		d.err = fmt.Errorf("couldn't read measure for track %d. Read %d bytes, expected %d, err: %s", len(d.p.Tracks)+1, n, len(t.Measure), err)
		return
	}
	d.pos += 16
	d.p.Tracks = append(d.p.Tracks, t)
}

func (d *decoder) done() bool {
	return d.err != nil || d.pos >= d.rlen
}

func (d *decoder) close() error {
	return d.err
}
