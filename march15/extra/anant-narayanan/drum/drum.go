// Package drum implements the decoding of .splice drum machine files.
// See golang-challenge.com/go-challenge1/ for more information
package drum

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Version string   // The version that generated the pattern.
	Tempo   float32  // The tempo that the pattern must be played at.
	Tracks  []*Track // A list of tracks contained by this pattern.
}

// Track is the high level representation of an indivual
// track in a drum pattern.
type Track struct {
	ID    byte   // The track ID.
	Name  string // The track name.
	Steps []byte // A 16-byte array indicating steps at which this track is played.
}
