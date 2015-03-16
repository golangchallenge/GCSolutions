package drum

import (
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
)

// NewTrack creates a new track instance from its id, name and sequence
func NewTrack(id int64, name string, seq []byte) *Track {
	sequence := strings.Replace(strings.Replace(hex.EncodeToString(seq), "01", "x", 100), "00", "-", 100)
	return &Track{id, name, string(sequence)}
}

// String formats a Track as a String where steps are grouped by 4 and separated with a pipe character
func (t *Track) String() string {
	reg, _ := regexp.Compile("(....)")
	return fmt.Sprintf("(%d) %s\t%s", t.number, t.name, reg.ReplaceAllString(t.sequence, "|$1")+"|")
}

// NewPattern creates a new Pattern instance from its Splice version and tempo
func NewPattern(version string, tempo float32) *Pattern {
	return &Pattern{version, tempo, make([]Track, 0)}
}

// AddTrack adds a Track to a Pattern
func (p *Pattern) AddTrack(track *Track) {
	p.tracks = append(p.tracks, *track)
}

// String formats a Pattern as a String using the challenge specified format
func (p *Pattern) String() string {
	str := fmt.Sprintf("Saved with HW Version: %s\nTempo: %g\n", p.version, p.tempo)
	for _, track := range p.tracks {
		str += track.String() + "\n"
	}
	return str
}
