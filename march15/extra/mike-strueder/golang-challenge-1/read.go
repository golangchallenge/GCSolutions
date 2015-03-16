package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

func readHeader(r io.Reader) (string, error) {
	b := make([]byte, 12)
	_, err := r.Read(b)

	b = bytes.TrimRight(b, "\x00")
	return string(b), err
}

func readSize(r io.Reader) (uint16, error) {
	var s uint16
	err := binary.Read(r, binary.BigEndian, &s)
	return s, err
}

func readVersion(r io.Reader) (string, error) {
	b := make([]byte, 32)
	_, err := r.Read(b)

	b = bytes.TrimRight(b, "\x00")
	return string(b), err
}

func readTempo(r io.Reader) (float32, error) {
	var s float32
	err := binary.Read(r, binary.LittleEndian, &s)
	return s, err
}

func readTrackID(r io.Reader) (byte, error) {
	var s byte
	err := binary.Read(r, binary.BigEndian, &s)
	return s, err
}

func readTrackNameLength(r io.Reader) (byte, error) {
	var buf = make([]byte, 4)

	_, err := r.Read(buf)
	if err != nil {
		return 0, err
	}
	return buf[3], err
}

func readTrackName(r io.Reader, c byte) (string, error) {
	s := make([]byte, c)
	_, err := r.Read(s)

	return string(s), err
}

func readTrackSteps(r io.Reader) ([]bool, error) {
	data := make([]byte, 16)
	_, err := r.Read(data)
	if err != nil {
		return nil, err
	}

	res := make([]bool, 16)
	for i, s := range data {

		switch {
		case s == 0x01:
			res[i] = true
		case s == 0x00:
			res[i] = false
		default:
			return nil, errors.New("invalid step encountered")
		}

	}
	return res, err
}

func readTracks(r io.Reader, size uint16) ([]Track, error) {
	var tracks []Track
	for {
		track, err := readTrack(r, size)
		size = size - 21 - uint16(len(track.Name))

		switch err {
		case nil:
			tracks = append(tracks, track)
			if size == 0 {
				return tracks, nil
			}

		case io.EOF:
			return tracks, nil

		default:
			return nil, err
		}
	}

}

func readTrack(r io.Reader, size uint16) (Track, error) {
	id, err := readTrackID(r)
	length, err := readTrackNameLength(r)

	if size-21-uint16(length) < 0 {
		return Track{}, errors.New("pattern size mismatch")
	}

	name, err := readTrackName(r, length)
	steps, err := readTrackSteps(r)

	return Track{
		ID:    id,
		Name:  name,
		Steps: steps,
	}, err

}
