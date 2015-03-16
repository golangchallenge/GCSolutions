package drum

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io/ioutil"
)

// Serialization format
// Each splice file begins with "SPLICE"
// On index versionInd = 14, the version of the software will be present, terminated by 0. Max len = 32 bytes
// On index tempoInd = 46, there will be 4 LittleEndian bytes representing the tempo of the piece in a float32
// On index trackInd = 55, the tracks begin.
// Each track is represented by 4 pieces of data. For a track starting at index i
// i-5 contains the track ID. Next 3 indices are zeros
// i-1 contains the length of the name (namelen).
// i to i+namelen contains the name
// i+namelen to i+namelen+16 contains the beats of the track

const (
	versionInd = 14 // index at which the version starts
	tempoInd   = 46 // index at which the tempo starts
	trackInd   = 55 // index at which the tracks start
	trackLen   = 16 // length of the sequence of beats on a track
	trackSep   = 5  // number of bytes between tracks
)

var (
	errVersion   = errors.New("Version not found")
	errTracks    = errors.New("Found no tracks")
	errTrackName = errors.New("Track name empty")
	errTrackSeq  = errors.New("Found unexpected number in track sequence")
)

// DecodeFile decodes the drum machine file found at the provided path
// and returns a pointer to a parsed pattern which is the entry point to the
// rest of the data.
func DecodeFile(path string) (*Pattern, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return &Pattern{}, err
	}

	version, err := getVersion(b)
	if err != nil {
		return &Pattern{}, err
	}

	tempo, err := getTempo(b)
	if err != nil {
		return &Pattern{}, err
	}

	tracks, err := getTracks(b)
	if err != nil {
		return &Pattern{}, err
	}

	p := &Pattern{version, tempo, tracks}
	return p, nil
}

func getVersion(b []byte) (string, error) {
	version := bytes.Trim(b[versionInd:versionInd+32], "\x00")
	if len(version) == 0 {
		return "", errVersion
	}
	return string(version), nil
}

func getTempo(b []byte) (float32, error) {
	var f float32
	buf := bytes.NewReader(b[tempoInd : tempoInd+4])
	err := binary.Read(buf, binary.LittleEndian, &f)
	if err != nil {
		return 0, err
	}
	return f, nil
}

func getTracks(b []byte) ([]Track, error) {
	tracks := make([]Track, 0, 10)
	for i := trackInd; i < len(b); {
		if isValidTrack(b, i) {
			t, err := getTrack(b, i)
			if err != nil {
				return tracks, err
			}
			tracks = append(tracks, t)
			i += len(t.Name) + trackSep + trackLen
		} else {
			break
		}
	}
	if len(tracks) == 0 {
		return tracks, errTracks
	}
	return tracks, nil
}

func isValidTrack(b []byte, i int) bool {
	if b[i-1] > 0 && b[i-2] == 0 && b[i-3] == 0 && b[i-4] == 0 {
		return true
	}
	return false
}

func getTrack(b []byte, i int) (Track, error) {
	id := int(b[i-5])
	namelen := int(b[i-1])

	name, err := getName(b, i, i+namelen)
	if err != nil {
		return Track{}, err
	}

	beats, err := getSeq(b, i+namelen)
	if err != nil {
		return Track{}, err
	}

	return Track{name, beats, id}, nil
}

func getName(b []byte, start, end int) (string, error) {
	name := bytes.Trim(b[start:end], "\x00")
	if len(name) == 0 {
		return "", errTrackName
	}
	return string(name), nil
}

func getSeq(b []byte, i int) ([]int, error) {
	r := make([]int, trackLen)
	for j := 0; j < trackLen; j++ {
		r[j] = int(b[i])
		if r[j] != 1 && r[j] != 0 {
			return r, errTrackSeq
		}
		i++
	}
	return r, nil
}
