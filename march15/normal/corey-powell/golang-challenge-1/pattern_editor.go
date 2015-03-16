package drum

import (
	"errors"
)

type TrackEditor struct {
	*Track
}

// IndexOutOfRange is returned when a provided index is out of range (duh).
var IndexOutOfRange = errors.New("Index is out of range.")

func (m *TrackEditor) checkStepIndex(idx int) error {
	if idx < 0 {
		return IndexOutOfRange
	} else if idx >= len(m.Track.Steps) {
		return IndexOutOfRange
	}
	return nil
}

// FillFor will start setting the steps from strt for a lenth of run.
// NOTE: run will be wrapped if it is longer than the Steps.
func (m *TrackEditor) FillFor(val byte, strt int, run int) {
	en := strt + run
	ln := len(m.Track.Steps)
	for i := strt; i < en; i++ {
		m.Track.Steps[i%ln] = val
	}
}

// Fill will change all the steps to the provided val
func (m *TrackEditor) Fill(val byte) {
	m.FillFor(val, 0, len(m.Track.Steps))
}

// Clear will Fill all the steps with 0 (OFF)
func (m *TrackEditor) Clear() {
	m.Fill(0)
}

// FillEvery will fill every other count, setting off if the step is not on the count, otherwise on
func (m *TrackEditor) FillEvery(count int, off byte, on byte) {
	if count <= 0 {
		return
	}
	for i := range m.Track.Steps {
		v := off
		if i%count == 0 {
			v = on
		}
		m.Track.Steps[i] = v
	}
}

// Rotate the Track steps by amt
func (m *TrackEditor) Rotate(amt int) {
	ln := len(m.Track.Steps)
	amt = amt % ln
	if amt > 0 {
		bg := m.Track.Steps[0:amt]
		en := m.Track.Steps[amt:ln]
		copy(m.Track.Steps[:], append(en, bg...))
	}
}

// Rotate the Track steps to the left, same as Rotate(-amt)
func (m *TrackEditor) RotateLeft(amt int) {
	m.Rotate(-amt)
}

// Rotate the Track steps to the right, same as Rotate(amt)
func (m *TrackEditor) RotateRight(amt int) {
	m.Rotate(amt)
}

func (m *TrackEditor) Set(idx int, value byte) error {
	if err := m.checkStepIndex(idx); err != nil {
		return err
	}
	m.Track.Steps[idx] = value
	return nil
}

func (m *TrackEditor) Get(idx int) (byte, error) {
	if err := m.checkStepIndex(idx); err != nil {
		return 0, err
	}
	return m.Track.Steps[idx], nil
}

func (m *TrackEditor) Toggle(idx int) error {
	v, err := m.Get(idx)
	if err != nil {
		return err
	}
	if v == 0 {
		return m.Set(idx, 1)
	}
	return m.Set(idx, 0)
}

func NewTrackEditor(track *Track) *TrackEditor {
	return &TrackEditor{Track: track}
}

type PatternEditor struct {
	*Pattern
}

func (m *PatternEditor) makeTrack(name string) *Track {
	return &Track{Name: name}
}

func (m *PatternEditor) addTrack(track *Track) {
	m.Pattern.Tracks = append(m.Pattern.Tracks, track)
}

// Adds a new Track to the Pattern and returns its index, use CreateTrack instead
// if you want a TrackEditor instead of the Index
func (m *PatternEditor) NewTrack(name string) int {
	track := m.makeTrack(name)
	idx := len(m.Pattern.Tracks)
	track.ID = int32(idx)
	m.addTrack(track)
	return idx
}

// Provided a track index, EditTrack will return a TrackEditor for that track,
// if the index is out of range, a IndexOutOfRange error is returned
func (m *PatternEditor) EditTrack(idx int) (*TrackEditor, error) {
	if idx < 0 || idx >= len(m.Pattern.Tracks) {
		return nil, IndexOutOfRange
	}
	return NewTrackEditor(m.Pattern.Tracks[idx]), nil
}

// Same as EditTrack(NewTrack(name))
func (m *PatternEditor) CreateTrack(name string) (*TrackEditor, error) {
	return m.EditTrack(m.NewTrack(name))
}

func NewPatternEditor(p *Pattern) *PatternEditor {
	return &PatternEditor{Pattern: p}
}
