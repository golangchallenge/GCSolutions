package sequencer

import (
	portaudio "code.google.com/p/portaudio-go/portaudio"
	"fmt"
	drum "github.com/kellydunn/go-challenge-1"
	"github.com/mkb218/gosndfile/sndfile"
)

// Sequencer describes the mechanism that
// Triggers and synchronizses a Pattern for audio playback.
type Sequencer struct {
	Timer   *Timer
	Bar     int
	Beat    int
	Pattern *drum.Pattern
	Stream  *portaudio.Stream
}

// NewSequencer creates and returns a pointer to a New Sequencer.
// Returns an error if there is one encountered
// During initializing portaudio, or the default stream
func NewSequencer() (*Sequencer, error) {
	err := portaudio.Initialize()
	if err != nil {
		return nil, err
	}

	s := &Sequencer{
		Timer: NewTimer(),
		Bar:   0,
		Beat:  0,
	}

	stream, err := portaudio.OpenDefaultStream(
		InputChannels,
		OutputChannels,
		float64(SampleRate),
		portaudio.FramesPerBufferUnspecified,
		s.ProcessAudio,
	)

	if err != nil {
		return nil, err
	}

	s.Stream = stream

	return s, nil
}

// Start starts the sequencer.
// Starts counting the Pulses Per Quarter note for the given BPM.
// Triggers samples based on each 16th note that is triggered.
func (s *Sequencer) Start() {
	go func() {
		ppqnCount := 0

		for {
			select {
			case <-s.Timer.Pulses:
				ppqnCount++

				// TODO add in time signatures
				if ppqnCount%(int(Ppqn)/4) == 0 {
					index := (s.Bar * 4) + s.Beat
					go s.PlayTrigger(index)

					s.Beat++
					s.Beat = s.Beat % 4
				}

				// TODO Add in time signatures
				if ppqnCount%int(Ppqn) == 0 {
					s.Bar++
					s.Bar = s.Bar % 4
				}

				// 4 bars of quarter notes
				if ppqnCount == (int(Ppqn) * 4) {
					ppqnCount = 0
				}

			}
		}
	}()

	s.Timer.Start()
	s.Stream.Start()
}

// ProcessAudio is the callback function for the portaudio stream
// Attached the the current Sequencer.
// Writes all active Track Samples to the output buffer
// At the playhead for each track.
func (s *Sequencer) ProcessAudio(out []float32) {
	for i := range out {
		var data float32

		for _, track := range s.Pattern.Tracks {
			if track.Playhead < len(track.Buffer) {
				data += track.Buffer[track.Playhead]
				track.Playhead++
			}
		}

		if data > 1.0 {
			data = 1.0
		}

		out[i] = data
	}
}

// PlayTrigger triggers a playback for any track that is active for the passed in index.
// Triggers a playback by resetting the playhead for the matching tracks.
func (s *Sequencer) PlayTrigger(index int) {
	for _, track := range s.Pattern.Tracks {
		if track.StepSequence.Steps[index] == byte(1) {
			track.Playhead = 0
		}
	}
}

// LoadSample loads an audio sample from the passed in filename
// Into memory and returns the buffer.
// Returns an error if there was one in audio processing.
func LoadSample(filename string) ([]float32, error) {
	var info sndfile.Info
	soundFile, err := sndfile.Open(filename, sndfile.Read, &info)
	if err != nil {
		fmt.Printf("Could not open file: %s\n", filename)
		return nil, err
	}

	buffer := make([]float32, 10*info.Samplerate*info.Channels)
	numRead, err := soundFile.ReadItems(buffer)
	if err != nil {
		fmt.Printf("Error reading data from file: %s\n", filename)
		return nil, err
	}

	defer soundFile.Close()

	return buffer[:numRead], nil
}
