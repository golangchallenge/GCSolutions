// Copyright 2015 by Simon Kern.
// Use of this source code is governed by a cc-style license.

// Package drum is supposed to implement the decoding of .splice drum machine files.
// See golang-challenge.com/go-challenge1/ for more information
package drum

import (
	"bytes"
	"fmt"
)

func (p *Pattern) String() string {
	buf := bytes.NewBuffer(nil)
	buf.WriteString(fmt.Sprintf("Saved with HW Version: %v\nTempo: %v\n", p.Payload.Version, p.Payload.Tempo))

	for _, track := range p.Payload.Tracks {
		buf.WriteString(track.String())
	}

	return buf.String()
}

func (t *Track) String() string {
	buf := bytes.NewBuffer(nil)
	buf.WriteString(fmt.Sprintf("(%v) %v\t", t.ID, t.Name))

	// Output for steps
	for i, step := range t.Steps {

		if i%4 == 0 {
			buf.WriteRune('|')
		}

		switch {
		// Play
		case step == 1:
			buf.WriteRune('x')
		case step != 1:
			buf.WriteRune('-')
		}
	}

	buf.WriteString("|\n")
	return buf.String()
}
