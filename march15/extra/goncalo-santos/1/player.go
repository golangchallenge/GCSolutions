package drum

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"

	"code.google.com/p/portaudio-go/portaudio"
)

const sampleRate = 44100

var toneMap map[byte]*tone

type tone struct {
	step, phase float64
	playing     bool
}

func newTone(freq, sampleRate float64) (*tone, error) {
	t := &tone{freq / sampleRate, 0, false}
	return t, nil
}

// Get sine wave amplitude
func (t *tone) next() float32 {
	_, t.phase = math.Modf(t.phase + t.step)

	if !t.playing {
		return 0
	}

	retVal := float32(math.Sin(2 * math.Pi * t.phase))

	return retVal
}

// Mix every sound wave
func processAudio(out []float32) {
	for i := range out {
		out[i] = 0
		for _, t := range toneMap {
			out[i] += t.next()
		}
	}
}

// Play will play the given pattern.
// It will generate different sounds for each instrument
// every time it is called. For simplification, those sounds
// are just sine waves.
func (p *Pattern) Play(playingTime time.Duration) error {
	// Initialize portaudio
	portaudio.Initialize()
	defer portaudio.Terminate()

	stream, err := portaudio.OpenDefaultStream(0, 1, sampleRate, 0, processAudio)
	if err != nil {
		return errors.New("could not open default stream")
	}
	defer stream.Close()

	// Create random tone map
	toneMap = make(map[byte]*tone)
	rand.Seed(time.Now().Unix())
	for i := range p.instruments {
		var err error
		toneMap[i], err = newTone(rand.Float64()*600+300, sampleRate)
		if err != nil {
			return fmt.Errorf("could not create tone for instrument %v", i)
		}
	}

	stream.Start()
	defer stream.Stop()

	// Signal for stopping
	timeOut := time.After(playingTime)

	timePerStep := time.Duration(60/p.header.BPM*1000) * time.Millisecond
	ticker := time.NewTicker(timePerStep)
	currentStep := 0

	// Play!
	for _ = range ticker.C {
		for i, instrument := range p.instruments {
			if instrument.Pattern[currentStep] == 0 {
				toneMap[i].playing = false
			} else {
				toneMap[i].playing = true
			}
		}

		currentStep++
		if currentStep > 15 {
			currentStep = 0
		}

		select {
		case <-timeOut:
			ticker.Stop()
			return nil
		default:
		}
	}

	return nil
}
