package drum

import (
	"bytes"
	"encoding/binary"
	"io"
)

// Encode a pattern into binary data
func (pattern *Pattern) Encode() io.Reader {
	buf := new(bytes.Buffer)
	buf.WriteString("SPLICE")

	contentbuf := new(bytes.Buffer)

	// Write version
	contentbuf.WriteString(pattern.Version)
	contentbuf.Write(make([]byte, 32-len(pattern.Version)))

	// Write tempo
	binary.Write(contentbuf, binary.LittleEndian, pattern.Tempo)

	// Write tracks
	for _, track := range pattern.Tracks {
		binary.Write(contentbuf, binary.LittleEndian, int32(track.ID))
		binary.Write(contentbuf, binary.LittleEndian, int8(len(track.Name)))
		contentbuf.WriteString(track.Name)
		for _, step := range track.Steps {
			if step {
				binary.Write(contentbuf, binary.LittleEndian, int8(1))
			} else {
				binary.Write(contentbuf, binary.LittleEndian, int8(0))
			}
		}
	}

	// Write contentlength and content to buffer
	binary.Write(buf, binary.BigEndian, int64(contentbuf.Len()))
	contentbuf.WriteTo(buf)

	return buf
}
