package drum

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
)

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	version string
	tempo   float32
	tracks  []track
}

func (p *Pattern) readPattern(b []byte) {
	b = b[0 : lHeader+int(b[pRead])] // Read only these many bytes
	p.version = fmt.Sprintf("%s", bytes.Trim(b[lHeader:pVerEnd], "\x00"))
	p.tempo =
		math.Float32frombits(binary.LittleEndian.Uint32(b[pTempo:pTrackStart]))

	for pos := pTrackStart; pos < len(b); {
		var t track
		pos = t.readTrack(b, pos)
		p.tracks = append(p.tracks, t)
	}
}

func (p *Pattern) String() string {
	s := fmt.Sprintf("Saved with HW Version: %s\nTempo: %g\n", p.version, p.tempo)
	for _, t := range p.tracks {
		s += fmt.Sprintf("%s\n", &t)
	}
	return s
}
