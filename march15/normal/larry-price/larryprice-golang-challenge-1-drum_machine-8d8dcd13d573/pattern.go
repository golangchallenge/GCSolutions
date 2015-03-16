package drum

import "fmt"

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	version Version
	tempo   Tempo
	tracks  Tracks
}

func (p Pattern) String() string {
	return fmt.Sprintf("%s\n%s\n%s", p.version.String(), p.tempo.String(), p.tracks.String())
}

// A Version describes the version of the drum machine used to generate the .splice file.
type Version string

func (v Version) String() string {
	return fmt.Sprintf("Saved with HW Version: %s", string(v))
}

// A Tempo is the number of beats-per-minute used when playing tracks from a given pattern.
type Tempo float64

func (t Tempo) String() string {
	return fmt.Sprintf("Tempo: %.3g", t)
}

// Tracks contains a reference to tracks for any instrument used in creating the file.
type Tracks []Track

func (t Tracks) String() string {
	pretty := ""
	for _, track := range t {
		pretty = fmt.Sprintf("%s%s\n", pretty, track.String())
	}
	return pretty
}

// A Track is a single instrument's part of the music.
type Track struct {
	ID    int
	Name  string
	Steps Steps
}

func (t Track) String() string {
	return fmt.Sprintf("(%d) %s\t%s", t.ID, t.Name, t.Steps.String())
}

// Steps refers to a collection of bars combined to create a track.
type Steps struct {
	Bars [4]Bar
}

func (s Steps) String() string {
	pretty := "|"
	for _, bar := range s.Bars {
		pretty = pretty + bar.String() + "|"
	}
	return pretty
}

// A Bar is a collection of notes.
type Bar [4]Note

func (bar Bar) String() string {
	pretty := ""
	for _, note := range bar {
		pretty = pretty + note.String()
	}
	return pretty
}

// A Note is a single unit of musical composition.
type Note bool

func (n Note) String() string {
	if n {
		return "x"
	}
	return "-"
}
