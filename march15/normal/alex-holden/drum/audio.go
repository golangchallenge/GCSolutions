// The author disclaims copyright to this source code.  In place of
// a legal notice, here is a blessing:
//
//    May you do good and not evil.
//    May you find forgiveness for yourself and forgive others.
//    May you share freely, never taking more than you give.

package drum

import (
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

type audioData struct {
	Channels         uint16
	Samples          uint32
	SamplesPerSecond uint32
	BytesPerSecond   uint32
	BytesPerBlock    uint16
	BitsPerSample    uint16
	Data             []int32
}

func decodeAudio(p string) (*audioData, error) {
	switch strings.ToLower(path.Ext(p)) {
	case ".wav":
		return decodeWAV(p)
	default:
		return nil, fmt.Errorf("unsupported audio format")
	}
}

func decodeWAV(path string) (*audioData, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	if string(b[0:4]) != "RIFF" || string(b[8:16]) != "WAVEfmt " {
		return nil, fmt.Errorf("bad file format")
	}

	a := audioData{
		Channels:         binary.LittleEndian.Uint16(b[22:24]),
		SamplesPerSecond: binary.LittleEndian.Uint32(b[24:28]),
		BytesPerSecond:   binary.LittleEndian.Uint32(b[28:32]),
		BytesPerBlock:    binary.LittleEndian.Uint16(b[32:34]),
		BitsPerSample:    binary.LittleEndian.Uint16(b[34:36]),
	}
	if string(b[36:40]) != "data" {
		return nil, fmt.Errorf("bad file format")
	}
	sampleDataSize := binary.LittleEndian.Uint32(b[40:44])

	b = b[44 : 44+sampleDataSize]
	a.Data = make([]int32, len(b)/4)

	switch a.BytesPerBlock {
	case 3:
		ts := make([]byte, 4)
		for i := 0; i < len(a.Data); i++ {
			copy(ts[1:4], b[i*3:(i+1)*3])
			a.Data[i] = int32(binary.LittleEndian.Uint32(ts))
		}
	default:
		for i := 0; i < len(a.Data); i++ {
			a.Data[i] = int32(binary.LittleEndian.Uint32(b[i*4 : (i+1)*4]))
		}
	}

	return &a, nil
}
