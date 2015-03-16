package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io/ioutil"
	"math"
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	r := &countReader{r: bytes.NewReader(buf)}
	p := &Pattern{}
	b := make([]byte, 32)

	if err = decodeHeader(r, b[:6]); err != nil {
		return nil, err
	}
	length, err := decodeLength(r, b[:8])
	if err != nil {
		return nil, err
	}
	if p.Version, err = decodeVersion(r, b[:32]); err != nil {
		return nil, err
	}
	if p.Tempo, err = decodeTempo(r, b[:4]); err != nil {
		return nil, err
	}
	for r.count < int(length) {
		track, err := decodeTrack(r, b)
		if err != nil {
			return nil, err
		}
		p.Tracks = append(p.Tracks, track)
	}
	return p, nil
}

func decodeHeader(r *countReader, b []byte) error {
	if err := r.read(b); err != nil {
		return err
	}
	if string(b) != "SPLICE" {
		return errors.New("drum: expected SPLICE header")
	}
	return nil
}

func decodeLength(r *countReader, b []byte) (uint64, error) {
	if err := r.read(b); err != nil {
		return 0, err
	}
	l := binary.BigEndian.Uint64(b)
	return l, nil
}

func decodeVersion(r *countReader, b []byte) (string, error) {
	if err := r.read(b); err != nil {
		return "", err
	}
	end := bytes.IndexByte(b, 0)
	v := string(b[:end])
	return v, nil
}

func decodeTempo(r *countReader, b []byte) (float32, error) {
	if err := r.read(b); err != nil {
		return 0, err
	}
	bits := binary.LittleEndian.Uint32(b)
	t := math.Float32frombits(bits)
	return t, nil
}

func decodeTrack(r *countReader, b []byte) (track Track, err error) {
	if track.ID, err = r.readByte(); err != nil {
		return track, err
	}
	track.Name, err = decodeName(r, b[:4])
	if err != nil {
		return track, err
	}
	for i := range track.Steps {
		if track.Steps[i], err = decodeStep(r); err != nil {
			return track, err
		}
	}
	return track, nil
}

func decodeName(r *countReader, b []byte) (string, error) {
	if err := r.read(b); err != nil {
		return "", err
	}
	length := binary.BigEndian.Uint32(b)
	b = b[:length]
	if err := r.read(b); err != nil {
		return "", err
	}
	return string(b), nil
}

func decodeStep(r *countReader) (bool, error) {
	b, err := r.readByte()
	if err != nil {
		return false, err
	}
	return b == 1, nil
}

// countReader counts the read bytes.
type countReader struct {
	count int
	r     *bytes.Reader
}

func (r *countReader) read(p []byte) error {
	n, err := r.r.Read(p)
	r.count += n
	return err
}

func (r *countReader) readByte() (byte, error) {
	b, err := r.r.ReadByte()
	r.count++
	return b, err
}
