package drum

import (
	"errors"
	"strings"
	"time"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/sdl_mixer"
)

const (
	SoundPath string = "sounds/"
)

// Helper function to load a WAV file as a SDL Chunk. It expects a track name and converts
// it to match the filename of the WAVs. E.g. "Kick" -> "kick.wav", "Hi Hat" -> "hi-hat.wav", etc.
func loadWav(n string) (*mix.Chunk, error) {
	wavName := strings.Replace(strings.ToLower(n), " ", "-", -1)
	path := SoundPath + wavName + ".wav"
	wav := mix.LoadWAV(path)
	if wav == nil {
		return nil, errors.New("Unable to load WAV: " + path)
	}
	return wav, nil
}

// A Player plays back a Pattern using the provided WAV files and SDL2.
type Player struct {
	pattern       *Pattern
	stepCallbacks []func(s int)
	isPlaying     bool
	Sounds        []*mix.Chunk
	Shuffle       float64
}

func PlayerNew(p *Pattern) (*Player, error) {

	// Try to initialize the SDL2 mixer.
	if !mix.OpenAudio(44100, sdl.AUDIO_U16, mix.DEFAULT_CHANNELS, 512) {
		return nil, errors.New("Unable to open audio.")
	}

	// We (possibly) need to be able to play many chunks on the same time.
	// Thus we have to ask the SDL-mixer to allocate some more channels for us.
	mix.AllocateChannels(64)

	pl := &Player{pattern: p, Shuffle: 0}

	// Iterate over the tracks and load the corresponding WAVs.
	for _, t := range p.Tracks {
		wav, err := loadWav(t.Name)
		if err != nil {
			return nil, err
		}
		pl.Sounds = append(pl.Sounds, wav)
	}

	return pl, nil
}

// Frees all the chunks (loaded WAVs) and closes the SDL-mixer.
func (pl *Player) Free() {
	for _, s := range pl.Sounds {
		s.Free()
	}
	mix.CloseAudio()
}

// Add a callback function to be called on each step.
func (pl *Player) AddCallback(f func(s int)) {
	pl.stepCallbacks = append(pl.stepCallbacks, f)
}

// Get the number of milliseconds to sleep between steps based on the Pattern's tempo.
func (pl *Player) getSleepMs() time.Duration {
	return (time.Duration(60000.0/pl.pattern.Tempo) / 4) * time.Millisecond
}

// The loop that plays the pattern.
func (pl *Player) playLoop() {
	var sleep time.Duration

loop:
	for {

		// Update sleep time every bar (in case tempo has changed).
		sleep = pl.getSleepMs()

		// Loop through each step in the pattern.
		for s := 0; s < 16; s++ {

			// Should we stop playing?
			if !pl.isPlaying {
				break loop
			}

			// Call step callbacks.
			for _, f := range pl.stepCallbacks {
				f(s)
			}

			// Loop through each track, and if the current step is «on» in the track, play the track's sound.
			for i, t := range pl.pattern.Tracks {
				if t.Steps[s] {
					pl.Sounds[i].PlayChannel(-1, 0)
				}
			}

			// A simple step shuffle effect is achieved by sleeping a little longer/shorter each step.
			if s%2 == 0 {
				time.Sleep(time.Duration(float64(sleep) * (1 + pl.Shuffle)))
			} else {
				time.Sleep(time.Duration(float64(sleep) * (1 - pl.Shuffle)))
			}

		}
	}
}

// Starts playing the pattern (if not already playing).
func (pl *Player) Play() {
	if !pl.isPlaying {
		pl.isPlaying = true
		go pl.playLoop()
	}
}

// Stops playing.
func (pl *Player) Stop() {
	pl.isPlaying = false
}

// Starts playing if not playing, and stops playing if playing.
func (pl *Player) TogglePlay() {
	if pl.isPlaying {
		pl.Stop()
	} else {
		pl.Play()
	}
}
