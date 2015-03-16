// Paul Schuster
// 01MAR15
// encoding.go
package drum

import (
	"bytes"
	"encoding/binary"
	"math"
	"os"
)

type errWriter struct {
	buf bytes.Buffer
	err error
}

func (ew *errWriter) write(b []byte) {
	if ew.err != nil {
		return
	}
	_, ew.err = ew.buf.Write(b)
}

func (ew *errWriter) writePad(b []byte, c int) {
	if ew.err != nil {
		return
	}
	t := make([]byte, c)
	for k, v := range b {
		t[k] = v
	}
	for i := len(b); i < c; i++ {
		t[i] = '\x00'
	}
	ew.write(t)
}

// EncodeFile creates a drum pattern file from a parsed pattern. It returns an
// error if there was a problem writing to the file.
func EncodeFile(p *Pattern, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	var o bytes.Buffer
	ew := &errWriter{buf: o}
	ew.writePad([]byte("SPLICE"), 13)

	var tmp bytes.Buffer
	tw := &errWriter{buf: tmp}
	tw.writePad([]byte(p.Version), 32)

	tempo := make([]byte, 4)
	binary.LittleEndian.PutUint32(tempo, math.Float32bits(p.Tempo))
	tw.write(tempo)

	for _, v := range p.Tracks {
		tw.writePad([]byte{byte(v.ID)}, 4)
		tw.write([]byte{byte(len(v.Name))})
		tw.write([]byte(v.Name))
		tw.write(v.Pattern)
	}
	if tw.err != nil {
		return tw.err
	}

	ew.write([]byte{byte(tw.buf.Len())})
	ew.write(tw.buf.Bytes())
	if ew.err != nil {
		return ew.err
	}
	_, err = f.Write(ew.buf.Bytes())
	return err
}

// MoreCowbell is a destructive toy function to add more cowbell, to a drum
// pattern.
func MoreCowbell(p *Pattern) {
	for _, v := range p.Tracks {
		if v.Name == "cowbell" {
			v.Pattern[0] = '\x01'
			v.Pattern[4] = '\x01'
			v.Pattern[6] = '\x01'
			v.Pattern[12] = '\x01'
			v.Pattern[14] = '\x01'
		}
	}
}
