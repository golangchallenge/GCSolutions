package drum

import (
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatalf("Error[%s] while reading drum machine file[%s]\n", err.Error(), path)
	}
	p := &Pattern{}
	p.Decode(content)

	return p, nil
}

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Header  string
	Version string
	Tempo   float32
	Tracks  []*Track
}

type Track struct {
	Id         int32
	nameLength int
	Name       string
	Steps      string
}

// method for type Pattern
func (this Pattern) String() string {
	retValue := "Saved with HW Version: " + this.Version + "\nTempo: " + strconv.FormatFloat(float64(this.Tempo), 'f', -1, 32) + "\n"
	for _, t := range this.Tracks {
		retValue = retValue + t.String()
	}
	return retValue
}

// method for type Track
func (this Track) String() string {
	return "(" + fmt.Sprintf("%d", this.Id) + ") " + this.Name + "\t" + this.Steps + "\n"
}
