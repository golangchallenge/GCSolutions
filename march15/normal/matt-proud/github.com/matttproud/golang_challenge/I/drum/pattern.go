package drum

import (
	"bytes"
	"text/template"
)

// Version expresses the known file format versions.
type Version string

const (
	// V0708Alpha models a little endian binary format that appears as follows:
	//     [Magic (ASCII): 6B][Padding: 7B][Track Data Length (UINT): 1B]
	//     [Tempo (Float 32): 4B]{Repeated Tracks [ID (UINT): 1B][Padding: 3B]
	//     [Name Length (UINT): 1B][Name (ASCII): Name Length]
	//     [Steps (UINT): 1Bx16]}[Optional Splice Bits; Perhaps Previous
	//     Revision]
	V0708Alpha Version = "0.708-alpha"
	// V0808Alpha models a little endian binary format that appears as follows:
	//     [Magic (ASCII): 6B][Padding: 7B][Track Data Length (UINT): 1B]
	//     [Tempo (Float 32): 4B]{Repeated Tracks [ID (UINT): 1B][Padding: 3B]
	//     [Name Length (UINT): 1B][Name (ASCII): Name Length]
	//     [Steps (UINT): 1Bx16]}
	V0808Alpha = "0.808-alpha"
	// V0909 models a little endian binary format that appears as follows:
	//     [Magic (ASCII): 6B][Padding: 7B][Track Data Length (UINT): 1B]
	//     [Tempo (Float 32): 4B]{Repeated Tracks [ID (UINT): 1B][Padding: 3B]
	//     [Name Length (UINT): 1B][Name (ASCII): Name Length]
	//     [Steps (UINT): 1Bx16]}
	V0909 = "0.909"
	// Unknown indicates a version that we cannot handle.
	Unknown = "<Unknown>"
)

// Step models the state of the synthesized instrument at a given point of
// time.
type Step bool

func (s Step) String() string {
	switch s {
	case On:
		return "x"
	case Off:
		return "-"
	default:
		panic("unreachable")
	}
}

const (
	// On is the flyweight that models that an instrument is active at the given
	// instant.
	On Step = true
	// Off is the flyweight that models that an instrument is inactive at the
	// given instant.
	Off = false
)

// Measure encompasses four step states dispersed across an equal intervals.
type Measure [4]Step

var measureTmpl = template.Must(template.New("measure").Parse("{{range .}}{{.}}{{end}}"))

func (m *Measure) String() string { return execTmpl(measureTmpl, m) }

// Measures captures a sequence of Measure per the drum machine's storage
// capabilities.
type Measures [4]Measure

var measuresTmpl = template.Must(template.New("measures").Parse("|{{range .}}{{.}}|{{end}}"))

func (m *Measures) String() string { return execTmpl(measuresTmpl, m) }

// Track is a single channel of a synthesized audio stream.
type Track struct {
	// ID identifies it.
	ID int
	// Name identifies the instrument type.
	Name string
	// Measures is an encoding of the instruments steps across time.
	Measures Measures
}

var trackTmpl = template.Must(template.New("track").Parse("({{.ID}}) {{.Name}}\t{{.Measures}}"))

func (t *Track) String() string { return execTmpl(trackTmpl, t) }

// Tracks is a collection of Track.
type Tracks []Track

var tracksTmpl = template.Must(template.New("tracks").Parse("{{range .}}{{.}}\n{{end}}"))

func (t Tracks) String() string { return execTmpl(tracksTmpl, t) }

// Pattern is the high-level representation of the drum pattern contained in a
// .splice file.
type Pattern struct {
	// Version conveys the format used to encode the file.
	Version Version
	// Tempo is the tempo, the speed of the music.
	Tempo float32
	// Tracks contains all the audio channels.
	Tracks Tracks
}

var patternTmpl = template.Must(template.New("pattern").Parse(`Saved with HW Version: {{.Version}}
Tempo: {{.Tempo}}
{{.Tracks}}`))

func (p *Pattern) String() string { return execTmpl(patternTmpl, p) }

// execTmpl executes the provided template on the data, emitting the string.
// It presumes that the input data is semantically correct for the template.
func execTmpl(t *template.Template, data interface{}) string {
	var b bytes.Buffer
	if err := t.Execute(&b, data); err != nil {
		panic("drum: " + err.Error())
	}
	return b.String()
}
