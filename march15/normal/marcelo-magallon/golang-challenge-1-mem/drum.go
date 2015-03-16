package drum

import "fmt"

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Version string
	Tempo   float32
	Tracks  []Track
}

func (p *Pattern) String() string {
	s := fmt.Sprintf(
		"Saved with HW Version: %s\nTempo: %g\n",
		p.Version,
		p.Tempo)
	for _, track := range p.Tracks {
		s += track.String() + "\n"
	}
	return s
}

type Track struct {
	Id   int
	Name string
	Data Steps
}

func (t Track) String() string {
	return fmt.Sprintf("(%d) %s\t%s", t.Id, t.Name, t.Data)
}

type Steps [16]byte

func (s Steps) String() string {
	str := ""
	for i, b := range s {
		if i%4 == 0 {
			str += "|"
		}
		switch b {
		case 1:
			str += "x"
		case 0:
			str += "-"
		}
	}
	str += "|"
	return str
}
