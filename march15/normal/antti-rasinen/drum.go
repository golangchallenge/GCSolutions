// Package drum is supposed to implement the decoding of .splice drum machine files.
// See golang-challenge.com/go-challenge1/ for more information
package drum

import (
	"bytes"
	"fmt"
	"strings"
)

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Header
	tracks []*Track
}

type Header struct {
	Version [32]byte
	Tempo   float32
}

type Track struct {
	Id    int32
	Name  string
	Steps [16]byte
}

func (p Pattern) String() string {
	ver := bytes.Trim(p.Version[:], "\x00")

	s := `Saved with HW Version: %s
Tempo: %v
%s
`
	ts := []string{}
	for _, t := range p.tracks {
		ts = append(ts, t.String())
	}

	allTracks := strings.Join(ts, "\n")

	return fmt.Sprintf(s, ver, p.Tempo, allTracks)

}

func (t Track) String() string {
	return fmt.Sprintf(
		"(%d) %s\t|%s|%s|%s|%s|",
		t.Id,
		t.Name,
		formatBeat(t.Steps[0:4]),
		formatBeat(t.Steps[4:8]),
		formatBeat(t.Steps[8:12]),
		formatBeat(t.Steps[12:16]))
}

func formatBeat(b []byte) string {
	buf := make([]byte, 4)
	for i := 0; i < 4; i++ {
		if b[i] == 1 {
			buf[i] = 'x'
		} else {
			buf[i] = '-'
		}
	}
	return string(buf)
}
