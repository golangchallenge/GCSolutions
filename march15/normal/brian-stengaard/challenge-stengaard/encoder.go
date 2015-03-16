package drum

import (
	"bytes"
	"encoding/binary"
	"io"
)

// WriteTo writes p serialized on to w.
func (p *Pattern) WriteTo(w io.Writer) (n int64, err error) {

	write := func(b []byte) {
		if err == nil {
			var wn int
			wn, err = w.Write(b)
			n += int64(wn)
		}

	}

	tbuf := bytes.NewBuffer([]byte{})
	for _, track := range p.Tracks {
		tn, terr := track.WriteTo(tbuf)
		if terr != nil {
			return 0, terr
		}
		n += tn
	}

	write(magic[:])
	write([]byte{byte(versionLen + tempoLen + tbuf.Len())})

	write([]byte(p.Version))

	// Padding version string
	write(make([]byte, versionLen-len(p.Version)))

	tempoBuf := bytes.NewBuffer([]byte{})
	binary.Write(tempoBuf, binary.LittleEndian, p.Tempo)
	write(tempoBuf.Bytes())

	write(tbuf.Bytes())

	return n, err

}

// WriteTo writes a serialised version of t onto w
func (t *Track) WriteTo(w io.Writer) (n int64, err error) {
	data := struct {
		TrackID uint32
		NameLen byte
	}{
		TrackID: t.ID,
		NameLen: byte(len(t.Name)),
	}
	buf := bytes.NewBuffer([]byte{})

	err = binary.Write(buf, binary.LittleEndian, data)
	if err != nil {
		return 0, err
	}

	buf.Write([]byte(t.Name))
	buf.Write(t.Steps)

	return buf.WriteTo(w)

}
