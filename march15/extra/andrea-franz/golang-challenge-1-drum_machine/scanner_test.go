package drum

import (
	"bytes"
	"testing"
)

func TestScanner_scanSignature(t *testing.T) {
	p := NewPattern()
	r := bytes.NewBuffer([]byte{'S', 'P', 'L', 'I', 'C', 'E', 0, 0, 0, 0, 0, 0, 0})
	s := newScanner(r, p)

	_, err := s.scanSignature()
	if err != nil {
		t.Errorf("scanSignature returned an unexpected error: %v", err)
	}

	expected := "SPLICE"
	sig := p.Header.Signature
	if sig != expected {
		t.Errorf("scanSignature hasn't set the right Signature to Pattern's Header.\nExpected:\n`%s`\nGot:\n`%s`", expected, sig)
	}
}

func TestScanner_scanVersion(t *testing.T) {
	p := NewPattern()
	r := bytes.NewBuffer([]byte{
		'2', 'b', 'e', 't', 'a', 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		0, 0,
	})
	s := newScanner(r, p)

	_, err := s.scanVersion()
	if err != nil {
		t.Errorf("scanVersion returned an unexpected error: %v", err)
	}

	expected := "2beta"
	v := p.Header.Version
	if v != expected {
		t.Errorf("scanVersion hasn't set the right Version to Pattern's Header.\nExpected:\n%s\nGot:\n%s", expected, v)
	}
}

func TestScanner_scanTempo(t *testing.T) {
	p := NewPattern()
	r := bytes.NewBuffer([]byte{0, 0, 0xF0, 'B'})
	s := newScanner(r, p)

	_, err := s.scanTempo()
	if err != nil {
		t.Errorf("scanTempo returned an unexpected error: %v", err)
	}

	var expected float32 = 120.0
	tp := p.Header.Tempo
	if tp != expected {
		t.Errorf("scanTempo hasn't set the right Tempo to Pattern's Header.\nExpected:\n%v\nGot:\n%v", expected, tp)
	}
}

func TestTrackScanner_scanID(t *testing.T) {
	tr := &Track{}
	r := bytes.NewBuffer([]byte{1, 0, 0, 0})
	s := newTrackScanner(r, tr)

	_, err := s.scanID()
	if err != nil {
		t.Errorf("scanID returned an unexpected error: %v", err)
	}

	var expected int32 = 1
	if tr.ID != expected {
		t.Errorf("scanID hasn't set the right ID to Track.\nExpected:\n%v\nGot:\n%v", expected, tr.ID)
	}
}

func TestTrackScanner_scanID_invalidID(t *testing.T) {
	tr := &Track{}
	r := bytes.NewBuffer([]byte{1, 1, 1, 1})
	s := newTrackScanner(r, tr)

	_, err := s.scanID()
	if err == nil {
		t.Errorf("scanID was expected to return an InvalidTrackIDError")
	}
}

func TestTrackScanner_scanInstrument(t *testing.T) {
	tr := &Track{}
	r := bytes.NewBuffer([]byte{6, 'g', 'u', 'i', 't', 'a', 'r', 'z'})
	s := newTrackScanner(r, tr)

	_, err := s.scanInstrument()
	if err != nil {
		t.Errorf("scanInstrument returned an unexpected error: %v", err)
	}

	expected := "guitar"
	if tr.Instrument != expected {
		t.Errorf("scanInstrument hasn't set the right Instrument to Track.\nExpected:\n%v\nGot:\n%v", expected, tr.Instrument)
	}
}

func TestTrackScanner_scanSteps(t *testing.T) {
	tr := &Track{}
	b := []byte{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1}
	r := bytes.NewBuffer(b)
	s := newTrackScanner(r, tr)

	_, err := s.scanSteps()
	if err != nil {
		t.Errorf("scanSteps returned an unexpected error: %v", err)
	}

	expected := [16]byte{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1}
	if tr.Steps != expected {
		t.Errorf("scanSteps hasn't set the right Steps to Track.\nExpected:\n%v\nGot:\n%v", expected, tr.Steps)
	}
}
