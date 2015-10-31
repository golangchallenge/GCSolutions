package snd

import "math"

const DefaultNotesLen = 128

// Notes is a collection of note function evaluations.
type Notes []float64

// Eval evaluates fn over the length of ns. If ns is nil, ns will be allocated
// with length DefaultNotesLen.
func (ns *Notes) Eval(tones int, freq float64, pos int, fn NotesFunc) {
	if *ns == nil {
		*ns = make([]float64, DefaultNotesLen)
	}
	fn(*ns, tones, freq, pos)
}

// NotesFunc defines a note function to be evaluated over the length of ns.
type NotesFunc func(ns Notes, tones int, freq float64, pos int)

// EqualTempermantFunc evaluates notes as an octave containing n tones at an
// equal distance on a logarithmic scale. The reference freq and pos is used
// to find all other values.
func EqualTempermantFunc(ns Notes, tones int, freq float64, pos int) {
	for i := range ns {
		ns[i] = freq * math.Pow(math.Pow(2, 1/float64(tones)), float64(i-pos))
	}
}

// EqualTempermant is a helper function for returning an evaluated Notes.
func EqualTempermant(tones int, freq float64, pos int) Notes {
	var ns Notes
	ns.Eval(tones, freq, pos, EqualTempermantFunc)
	return ns
}
