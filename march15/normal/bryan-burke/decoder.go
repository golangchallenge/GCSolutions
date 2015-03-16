package drum

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"text/template"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {

	rawBytes, err := ioutil.ReadFile(path)

	if err != nil {
		return &Pattern{}, fmt.Errorf("Error: Could not read file %s", path)
	}

	p, err := Read(rawBytes)
	if err != nil {
		return p, err
	}

	return p, nil
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Header  Header
	Version Version
	Tempo   Tempo
	Tracks  []*Track
}

func (p *Pattern) String() string {

	funcMap := template.FuncMap{
		"ToString": func(x []byte) string { return string(x) },
	}

	const pat = `Saved with HW Version: {{.Version}}
Tempo: {{.Tempo}}
{{range $track := .Tracks}}({{$track.InstrumentID}}) {{$track.Name | ToString}}	|{{range $index, $elem := $track.Measure}}{{if eq $index 4 8 12}}|{{end}}{{if $elem}}x{{else}}-{{end}}{{end}}|
{{end}}`

	t := template.Must(template.New("pat").Funcs(funcMap).Parse(pat))
	var buf bytes.Buffer
	err := t.Execute(&buf, p)
	if err != nil {
		fmt.Print(err)
	}
	return buf.String()

}
