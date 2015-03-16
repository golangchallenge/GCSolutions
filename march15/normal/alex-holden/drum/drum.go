// The author disclaims copyright to this source code.  In place of
// a legal notice, here is a blessing:
//
//    May you do good and not evil.
//    May you find forgiveness for yourself and forgive others.
//    May you share freely, never taking more than you give.

// Package drum imelements the decoding of .splice drum machine files.
// It also contains a drum machine implementation that support the playback,
// creation, and modification of the .splice files in real time
package drum

import (
	"fmt"
	"io/ioutil"
	"path"
	"strconv"
	"strings"

	"code.google.com/p/portaudio-go/portaudio"
)

const (
	defaultSampleRate = 44100

	// (defaultSampleRate * beat) / (minute * sixteenth)
	// once divide by a bpm this gives samples/sixteenth
	sampleFactor = defaultSampleRate * 15
)

// Machine contains all the information necessary to load, play, and edit
// multiple Sequences, each containing multiple Sections.
type Machine struct {
	Sequences          []*Sequence
	Curr               *Sequence
	Samples            map[string]Sample
	On                 bool
	sequenceDir        string
	sampleDir          string
	sampleIdx          int
	beatIdx            int
	stream             *portaudio.Stream
	beatChangeCallback func(int)
}

// NewMachine creates the Machine struct and loads it with all
// available samples and sequences. it is expected that there will both a
// 'samples' directory and a 'sequences' directory within the package
func NewMachine(sequenceDir, sampleDir string) (*Machine, error) {
	m := Machine{sequenceDir: sequenceDir, sampleDir: sampleDir}
	m.Samples = make(map[string]Sample)

	// load samples
	dirs, err := ioutil.ReadDir(m.sampleDir)
	if err != nil {
		return nil, err
	}
	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}
		files, err := ioutil.ReadDir(path.Join(m.sampleDir, d.Name()))
		if err != nil {
			return nil, err
		}
		m.Samples[d.Name()] = Sample{
			Name:   d.Name(),
			sounds: make(map[soundKey]*audioData)}
		for _, f := range files {
			if strings.ToLower(path.Ext(f.Name())) != ".wav" {
				continue
			}
			ss := strings.SplitN(f.Name(), " ", 2)
			id, err := strconv.Atoi(ss[0])
			if err != nil || len(ss) < 2 {
				continue
			}
			nm := strings.SplitN(ss[1], ".", 2)
			if len(nm) < 2 {
				continue
			}
			s, err := decodeAudio(path.Join(m.sampleDir, d.Name(), f.Name()))
			if err != nil {
				return nil, err
			}
			m.Samples[d.Name()].sounds[soundKey{ID: uint16(id), Name: nm[0]}] = s
		}
	}

	// load sequences
	files, err := ioutil.ReadDir(m.sequenceDir)
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		if strings.ToLower(path.Ext(f.Name())) != ".splice" {
			continue
		}
		fp := path.Join(m.sequenceDir, f.Name())
		p, err := DecodeFile(fp)
		if err != nil {
			return nil, err
		}
		nm := strings.SplitN(f.Name(), ".", 2)
		s := Sequence{Name: nm[0], Version: p.Version, Tempo: p.Tempo}

		// create sections specified by pattern
		for _, t := range p.Tracks {
			if _, ok := m.Samples[p.Version].sounds[soundKey{ID: t.ID, Name: t.Name}]; !ok {
				return nil, fmt.Errorf("failed to load sample sound: %v %v", t.ID, t.Name)
			}
			sec := Section{ID: t.ID, Name: t.Name, Beats: t.Beats, Enabled: true}
			s.Sections = append(s.Sections, &sec)
		}

		// create sections for remaining audio samples
		for sk := range m.Samples[p.Version].sounds {
			add := true
			for _, t := range p.Tracks {
				if sk.ID == t.ID && sk.Name == t.Name {
					add = false
					break
				}
			}
			if add {
				sec := Section{ID: sk.ID, Name: sk.Name}
				s.Sections = append(s.Sections, &sec)
			}
		}
		m.Sequences = append(m.Sequences, &s)
	}

	if err := portaudio.Initialize(); err != nil {
		return nil, err
	}
	stream, err := portaudio.OpenDefaultStream(0, 1, defaultSampleRate, 0, m.audioCallback)
	if err != nil {
		return nil, err
	}
	m.stream = stream
	if err := stream.Start(); err != nil {
		return nil, err
	}
	return &m, nil
}

// SetBeatChangeCB set the function that will be called by the Machine
// at the beginning of every sixteenth note. The notes are zero based (0-15).
func (m *Machine) SetBeatChangeCB(cb func(int)) {
	m.beatChangeCallback = cb
}

// EnableSection turns on a section of a sequence, allowing the audio samples
// (if any) to be heard.
func (m *Machine) EnableSection(row int) {
	if row >= 0 && row < len(m.Curr.Sections) {
		m.Curr.Sections[row].Enabled = !m.Curr.Sections[row].Enabled
	}
}

// LoadSequence changes the Machine's currently playing sequence. If a
// beat changed callback has been registered, it will be called with beat 0
func (m *Machine) LoadSequence(row int) {
	m.On = false
	m.Curr = m.Sequences[row]
	m.beatIdx = 0
	m.sampleIdx = 0
	if m.beatChangeCallback != nil {
		m.beatChangeCallback(0)
	}
}

// TogglePlayPause stops/start the audio playback.
func (m *Machine) TogglePlayPause() {
	m.On = !m.On
}

// ChangeTempo modifies the tempo of the currently playing sequence.
func (m *Machine) ChangeTempo(tempo float32) {
	if tempo >= 1.0 && tempo < 1000.0 {
		m.Curr.Tempo = tempo
	}
}

// ToggleBeat flips the current beat value for a given section/beat combination.
func (m *Machine) ToggleBeat(row int, beat int) {
	if row >= 0 && row < len(m.Curr.Sections) && beat >= 0 && beat < 16 {
		m.Curr.Sections[row].Beats[beat] = !m.Curr.Sections[row].Beats[beat]
	}
}

// ClearBeats sets all beats for a given section of the current sequence to false.
func (m *Machine) ClearBeats(row int) {
	if row < len(m.Curr.Sections) && row >= 0 {
		for i := 0; i < 16; i++ {
			m.Curr.Sections[row].Beats[i] = false
		}
	}
}

// Close stops playback of the Machine and closes the associated audio stream.
func (m *Machine) Close() {
	m.On = false
	m.stream.Stop()
	m.stream.Close()
}

// SaveCurrentSequence writes the current machine sequence to a file in
// binary format. The file will be located in the sequence directory specified
// on Machine creation have the filename <name>.splice
func (m *Machine) SaveCurrentSequence(name string) {
	if m.Curr != nil {
		p := Pattern{Version: m.Curr.Version, Tempo: m.Curr.Tempo}
		newSeq := Sequence{Name: name, Version: m.Curr.Version, Tempo: m.Curr.Tempo}
		for _, s := range m.Curr.Sections {
			if s.Enabled {
				t := Track{ID: s.ID, Name: s.Name, Beats: s.Beats}
				p.Tracks = append(p.Tracks, &t)
			}
			sec := Section{ID: s.ID, Name: s.Name, Beats: s.Beats, Enabled: s.Enabled}
			newSeq.Sections = append(newSeq.Sections, &sec)
		}
		p.EncodeFile(path.Join(m.sequenceDir, name+".splice"))
		if name != m.Curr.Name {
			m.Sequences = append(m.Sequences, &newSeq)
		}
	}
}

// NewSequence will create a new sequence in the machine with the given name,
// based on the specified version (sample). Note that this will not create
// a splice file for the new sequence.
func (m *Machine) NewSequence(name, version string) {
	p := Pattern{Version: version, Tempo: 60.0}
	newSeq := Sequence{Name: name, Version: version, Tempo: 60.0}
	for sk := range m.Samples[version].sounds {
		sec := Section{ID: sk.ID, Name: sk.Name, Beats: [16]bool{}}
		newSeq.Sections = append(newSeq.Sections, &sec)
	}
	p.EncodeFile(path.Join(m.sequenceDir, name+".splice"))
	m.Sequences = append(m.Sequences, &newSeq)
	m.Curr = &newSeq
}

func (m *Machine) audioCallback(
	out []int32,
	timeInfo portaudio.StreamCallbackTimeInfo,
	flags portaudio.StreamCallbackFlags) {

	for i := 0; i < len(out); i, m.sampleIdx = i+1, m.sampleIdx+1 {
		out[i] = 0
		if !m.On || m.Curr == nil {
			continue
		}
		if m.sampleIdx >= (sampleFactor / int(m.Curr.Tempo)) {
			m.sampleIdx = 0
			m.beatIdx = (m.beatIdx + 1) % 16
			if m.beatChangeCallback != nil {
				m.beatChangeCallback(m.beatIdx)
			}
		}
		var tmp int64
		var enc int64
		for _, t := range m.Curr.Sections {
			if t.Enabled {
				enc++
				if t.Beats[m.beatIdx] {

					d := m.Samples[m.Curr.Version].sounds[soundKey{t.ID, t.Name}].Data
					if m.sampleIdx < len(d) {
						tmp += int64(d[m.sampleIdx])
					}
				}
			}
		}
		if enc > 0 {
			out[i] += int32(tmp / enc)
		}
	}
}

// Sequence represent a set of sounds (or 'Sections') and the associated playback
// information such as tempo, and the sample it is connected to (Version).
type Sequence struct {
	Name     string
	Version  string
	Tempo    float32
	Sections []*Section
}

// Section represents an audio sample, beat pattern, and misc. associated data.
type Section struct {
	ID      uint16
	Name    string
	Beats   [16]bool
	Enabled bool
	Volume  float32
}

// Sample represents mutiple audio signal grouped under a common name
type Sample struct {
	Name   string
	sounds map[soundKey]*audioData
}

type soundKey struct {
	ID   uint16
	Name string
}
