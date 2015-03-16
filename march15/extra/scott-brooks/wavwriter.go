package drum

import (
	"encoding/binary"
	"io"
	"os"
	"path"
)

type wavHeader struct {
	ChunkID   [4]byte
	ChunkSize int32
	Format    [4]byte

	Subchunk1ID   [4]byte
	Subchunk1Size int32
	AudioFormat   int16
	NumChannels   int16
	SampleRate    int32
	ByteRate      int32
	BlockAlign    int16
	BitsPerSample int16

	Subchunk2ID   [4]byte
	Subchunk2Size int32
}

const (
	sampleRate = 44100
)

// WriteWav writes our pattern to w for a certain number of seconds
func (p Pattern) WriteWav(w io.Writer, seconds int) error {
	samples := make([]float32, seconds*sampleRate)
	shortSamples := make([]int16, seconds*sampleRate)

	dataSize := len(shortSamples) * 2

	h := wavHeader{
		ChunkID:   [4]byte{'R', 'I', 'F', 'F'},
		ChunkSize: int32(24 + dataSize),
		Format:    [4]byte{'W', 'A', 'V', 'E'},

		Subchunk1ID:   [4]byte{'f', 'm', 't', ' '},
		Subchunk1Size: 16,
		AudioFormat:   1,
		NumChannels:   1,
		SampleRate:    sampleRate,
		ByteRate:      sampleRate * 2,
		BlockAlign:    2,
		BitsPerSample: 16,

		Subchunk2ID:   [4]byte{'d', 'a', 't', 'a'},
		Subchunk2Size: int32(dataSize),
	}

	err := binary.Write(w, binary.LittleEndian, h)
	if err != nil {
		return err
	}

	drumSamples := make([][]float32, len(p.Instruments))
	for idx, i := range p.Instruments {
		drumSamples[idx], err = loadSample(string(i.Name[:]))
		if err != nil {
			return err
		}
	}
	drumSampleCounts := make([]int, len(p.Instruments))

	// Beats per second
	bps := float32(p.Header.Tempo / 60.0)
	// Samples per beat
	spb := int(sampleRate / bps)
	beatCount := 0

	for i := range samples {
		// Start of a beat
		if i%spb == 0 {
			beatCount++
			for j := range drumSampleCounts {
				drumSampleCounts[j] = 0
			}
		}
		sample := float32(0)
		samplesUsed := float32(0)
		for j := 0; j < len(p.Instruments); j++ {
			if p.Instruments[j].Steps.onBeat(beatCount % 16) {
				if drumSampleCounts[j] < len(drumSamples[j])-1 {
					sample += drumSamples[j][drumSampleCounts[j]]
					samplesUsed++
					drumSampleCounts[j]++
				}
			}
		}
		samples[i] = sample / samplesUsed
	}

	for i, s := range samples {
		shortSamples[i] = int16(s * 32767)
	}
	err = binary.Write(w, binary.LittleEndian, shortSamples)
	if err != nil {
		return err
	}
	return nil

}

func loadSample(name string) ([]float32, error) {
	f, err := os.Open(path.Join("samples", name+".wav"))
	if err != nil {
		return nil, err
	}
	defer f.Close()

	h := wavHeader{}
	err = binary.Read(f, binary.LittleEndian, &h)
	if err != nil {
		return nil, err
	}
	numSamples := h.Subchunk2Size / int32(h.BitsPerSample) / 8
	samples := make([]float32, numSamples)
	shortSamples := make([]int16, numSamples)
	err = binary.Read(f, binary.LittleEndian, &shortSamples)

	for i, s := range shortSamples {
		samples[i] = float32(s) / 32767
	}

	return samples, nil
}
