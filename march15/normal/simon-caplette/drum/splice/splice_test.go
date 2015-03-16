package splice

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path"
	"testing"
)

func TestDecodeHeader(t *testing.T) {
	content, _ := ioutil.ReadFile(path.Join("..", "fixtures", "pattern_1.splice"))
	reader := NewReader(content)
	header, err := reader.GetHeader()
	if err != nil {
		fmt.Println("decoding header failed", err)
	}

	m := header.MagicNumber
	if !bytes.Equal(m[:], []byte(MagicNumber)) {
		t.Fatalf("header magic number: expecting %v", m)
	}
	v := header.Vers
	expv := []byte{'0', '.', '8', '0', '8', '-', 'a', 'l', 'p', 'h', 'a', '\x00'}
	if !bytes.Equal(v[:], expv) {
		t.Fatalf("header version: expecting %v", v)
	}
	vstring := header.Version()
	if vstring != "0.808-alpha" {
		t.Fatalf("header version string: expecting %q", vstring)
	}
	tempo := header.Tempo
	if tempo != 120 {
		t.Fatalf("header tempo: expecting %v", tempo)
	}
}

func TestDecodeTracks(t *testing.T) {
	content, _ := ioutil.ReadFile(path.Join("..", "fixtures", "pattern_2.splice"))
	reader := NewReader(content)
	tracks, err := reader.GetTracks()
	if err != nil {
		fmt.Println("decoding tracks failed", err)
	}

	if len(tracks) != 4 {
		t.Fatalf("tracks count: expecting %v", len(tracks))
	}

	track := tracks[2]

	id := track.Id
	if id != 3 {
		t.Fatalf("track id: expecting %v", id)
	}

	name := string(track.Name)
	if name != "hh-open" {
		t.Fatalf("track name: expecting %v", name)
	}

	s := fmt.Sprint(track.Steps)
	if s != "0010001010100010" {
		t.Fatalf("track steps: expecting %v", s)
	}
}
