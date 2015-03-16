package drum

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"reflect"
	"testing"
)

func TestTrackEncodeDecode(t *testing.T) {
	track := &Track{
		ID:   123,
		Name: "snare",
		Steps: []byte{
			0, 1, 0, 1,
			0, 1, 0, 1,
			0, 1, 0, 1,
			0, 1, 0, 1,
		},
	}
	buf := bytes.NewBuffer([]byte{})

	nw, err := track.WriteTo(buf)
	if err != nil {
		t.Fatal(err)
	}

	nr, tr, err := readTrack(buf.Bytes())
	if int64(nr) != nw {
		t.Errorf("Read different # of bytes than we wrote %d != %d", nr, nw)
	}
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(tr, t) {
		fmt.Errorf("Not equal\n%v\n%v", tr, t)
	}

}

func TestPattern(t *testing.T) {
	fbuf, err := ioutil.ReadFile("fixtures/pattern_1.splice")
	if err != nil {
		t.Fatal(err)
	}

	buf := bytes.NewBuffer(fbuf)
	p, err := DecodeReader(buf)
	if err != nil {
		t.Fatal(err)
	}

	wbuf := bytes.NewBuffer([]byte{})
	_, err = p.WriteTo(wbuf)
	if err != nil {
		t.Fatal(err)
	}

	if d := bytes.Compare(fbuf, wbuf.Bytes()); d != 0 {
		t.Errorf("Difference between what was read and what was written: %d", d)
		t.Log("original\n", hex.Dump(fbuf))
		t.Log("new\n", hex.Dump(wbuf.Bytes()))

	}

}
