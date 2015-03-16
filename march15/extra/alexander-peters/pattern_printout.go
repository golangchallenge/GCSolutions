package drum

import (
	"bytes"
	"fmt"
)

const (
	blockSize          = 4
	blockSeparator     = '|'
	symbolStepEnabled  = 'x'
	symbolStepDisabled = '-'
)

// String returns the Pattern in the printout format as a string.
func (p Pattern) String() string {
	w := new(bytes.Buffer)
	w.WriteString(fmt.Sprintf("Saved with HW Version: %s\n", p.version))
	w.WriteString(fmt.Sprintf("Tempo: %v\n", p.tempo))
	for _, v := range p.tracks {
		appendTrack(w, *v)
	}
	return w.String()
}

func appendTrack(w *bytes.Buffer, t Track) {
	w.WriteString(fmt.Sprintf("(%v) %v\t", t.id, t.name))
	appendSteps(w, t.steps)
	w.WriteString("\n")
}

func appendSteps(w *bytes.Buffer, s Steps) {
	for i, enabled := range s {
		if i%blockSize == 0 {
			w.WriteRune(blockSeparator)
		}
		if enabled {
			w.WriteRune(symbolStepEnabled)
		} else {
			w.WriteRune(symbolStepDisabled)
		}
	}
	w.WriteRune(blockSeparator)
}
