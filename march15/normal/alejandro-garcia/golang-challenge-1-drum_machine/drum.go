// Package drum is supposed to implement the decoding of .splice drum machine files.
// See golang-challenge.com/go-challenge1/ for more information
package drum

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
)

const (
	headerTrackIndexLimit = 36
	trackMaxLenth         = 16
)

// In the binary data the tempo has three possible encodings (66,67,68)
type tempoFunc func(b []byte) string

// The craziest thin ever
func tempSixSixFunc(b []byte) string {
	result := b[0] / 2
	if len(b) > 1 {
		decimal := float32(b[1]%10.0) / 10.0
		res := float32(result) + decimal
		return fmt.Sprintf("%v", res)
	}
	return fmt.Sprintf("%v", result)
}

// Almost crazy
func tempSixSevenFunc(b []byte) string {
	return fmt.Sprintf("%v", (b[0]*2)+16)
}

// Probably a not defined tempo
func tempSixEightFunc(b []byte) string {
	return fmt.Sprintf("%v", 999)
}

var (
	// map of tempo encoding and tempo function that caluculates it value
	tempoFuncMap = map[byte]tempoFunc{66: tempSixSixFunc, 67: tempSixSevenFunc, 68: tempSixEightFunc}
)

// Reads a binary pattern into a byte array
// It also discards the SPLICE useless header
// Returns an error when reading the file went wrong
func readFile(path string) (*[]byte, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	// trim header, we dont care about that SPLICE header
	data = data[14:]
	return &data, err
}

// Returns a slice containing the pattern headers
func header(data *[]byte) ([]byte, error) {
	if len(*data) < headerTrackIndexLimit {
		return []byte{}, errors.New("There is an error with data, too smal to hold all header information.")
	}
	return (*data)[:headerTrackIndexLimit], nil
}

// Returns a slice containing all tracks information
func tracks(data *[]byte) []byte {
	return (*data)[headerTrackIndexLimit:]
}

// Returns the Pattern version name
func patterVersion(data *[]byte) (string, error) {
	h, err := header(data)
	if err != nil {
		return "", err
	}
	result := []byte{}
	bytesToFollow := bytes.IndexByte(h, 0)
	for i := 0; i < bytesToFollow; i++ {
		result = append(result, h[i])
	}
	return string(result), nil
}

// Returns the pattern tempo value
func patternTempo(data *[]byte) string {
	// if we get here we dont care about the header error
	h, _ := header(data)
	values := []byte{}
	code := h[len(h)-1]
	for i := len(h) - 2; h[i] != 0; i-- {
		values = append(values, h[i])
	}
	return tempoFuncMap[code](values)
}

// Returns an array of Track information
func patternTracks(data *[]byte) []Track {
	result := []Track{}
	ts := tracks(data)
	tsSize := len(ts)
	for i := 0; i < tsSize; i = i + trackMaxLenth {
		// read id
		uid := int(ts[i])

		i = i + 4
		bytesToFollow := int(ts[i])
		//if we have a bad name or no space left after name for the steps then break
		if i+bytesToFollow >= tsSize || (tsSize-1)-(i+bytesToFollow) < trackMaxLenth {
			break
		}

		// read name
		name := []byte{}
		for j := 1; j <= bytesToFollow; j++ {
			name = append(name, ts[i+j])
		}

		i = i + bytesToFollow + 1
		steps := []string{"-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-"}
		// Read steps
		for j := 0; j < trackMaxLenth; j++ {
			if ts[i+j] != 0 {
				steps[j] = "x"
			}
		}
		tr := Track{id: strconv.Itoa(uid), name: string(name), steps: steps}
		result = append(result, tr)
	}
	return result
}
