package drum

import (
	"container/list"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"code.google.com/p/portaudio-go/portaudio"
	"github.com/mkb218/gosndfile/sndfile"
)

// Some sane default values. All wav files are expected to be 44100Hz stereo.
const numChannels = 2
const sampleRate = 44100
const framesPerBuffer = 64
const bufferSize = numChannels * framesPerBuffer

// A buffer is a collection of float values of size bufferSize.
type buffer []float32

// A sound is a collection of buffers. The length of the buffer array
// will depend on the total length of the .wav file.
type sound []buffer

// A queueItem encapsulates a sound that is queued to be played. 'contents' is
// a reference to the actual sound buffers, while 'played' keeps track of how
// many of these buffers have already been played.
//
// A queueItem will remove itself from the playQueue when all buffers in 'contents'
// have been played (i.e. played == len(contents)).
type queueItem struct {
	contents sound
	played   int
}

// Initialize an empty map of sound files.
var tracks = map[string]sound{}

// Preload all tracks into memory as "sounds" so they are ready for use by
// the portaudio callback.
func init() {
	// Walk all files inside the tracks directory.
	filepath.Walk("tracks", func(loc string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		// Open .wav file with libsndfile.
		var i sndfile.Info
		snd, err := sndfile.Open(loc, sndfile.Read, &i)
		if err != nil {
			fmt.Printf("could not read %s: %v\n", loc, err)
			os.Exit(-1)
		}

		// Now, load all the contents into memory.
		var cur sound
		for {
			set := make(buffer, bufferSize)
			n, err := snd.ReadItems(set)
			if err != nil {
				fmt.Printf("could not read items from file %s: %v\n", loc, err)
				os.Exit(-1)
			}
			if int(n) < len(set) {
				// We reached the end of the file, exit loop.
				break
			}
			cur = append(cur, set)
		}

		tracks[filepath.Base(loc)] = cur
		return nil
	})
}

// Play attempts to play the specified pattern for 'n' seconds on the
// default audio output device on the current system. An error is
// returned if the audio could not be played correctly.
func Play(pt *Pattern, n int) error {
	// Validate that all tracks present in our pattern have sound files loaded.
	for _, tr := range pt.Tracks {
		_, ok := tracks[strings.ToLower(tr.Name)+".wav"]
		if !ok {
			return fmt.Errorf("track %s could not be found, required to play pattern", tr.Name)
		}
	}

	// Let's do some math first. At a tempo of 120 (bpm), we play 2 beats per second.
	// That means each beat takes 0.5 seconds. A beat is comprised of a quarter note,
	// which in turn is comprised of 4 steps. This means, each step takes 0.125 seconds.
	// The time taken by a step is what we are interested in, as that is what we will
	// use to schedule the playing of sounds.
	beatTime := 1.0 / (pt.Tempo / 60.0)
	stepTime := beatTime / 4.0

	// Initialize a list that represents our play queue. This is a list of
	// sounds that are queued up for playback.
	var queue = list.New()

	// We need to keep track of how many steps have been played so far.
	// This is always a value between 0 and 15 (since our tracks have a fixed
	// 16 steps). We also track the last time a step was played, so we know
	// when we can schedule the next.
	lastStep := 0
	lastStepPlayed := time.Duration(0)

	// Initialize portaudio.
	err := portaudio.Initialize()
	if err != nil {
		return err
	}
	defer portaudio.Terminate()

	// Let's open a stream up and configure our callback.
	stream, err := portaudio.OpenDefaultStream(0, numChannels, sampleRate, framesPerBuffer, func(output buffer, info portaudio.StreamCallbackTimeInfo) {
		// If sufficient time has passed since we last played a step, queue up the next.
		if info.OutputBufferDacTime-lastStepPlayed >= time.Duration(stepTime*float32(time.Second)) {
			for _, tr := range pt.Tracks {
				// Is this track scheduled to be played at this step?
				if tr.Steps[lastStep] == 0x1 {
					queue.PushBack(&queueItem{tracks[strings.ToLower(tr.Name)+".wav"], 0})
				}
			}

			lastStepPlayed = info.OutputBufferDacTime

			lastStep++
			if lastStep > 15 {
				lastStep = 0
			}
		}

		// Multiplex all queued sounds and play them.
		multiplexSounds(output, queue)
	})
	if err != nil {
		return err
	}

	// Let's play some drums!
	err = stream.Start()
	if err != nil {
		return err
	}

	// For 'n' seconds....
	time.Sleep(time.Duration(n) * time.Second)

	// Stop and cleanup.
	err = stream.Stop()
	if err != nil {
		return err
	}
	err = stream.Close()
	if err != nil {
		return err
	}

	return nil
}

// Go through the current buffer for all sounds in the play queue, and sum
// up their values. Multiple sounds playing at the same time mathematically
// (and in the real world) works like this, but in software, we have a minimum/maximum
// value of -1.0/1.0. Thus, if the sum is outside those bounds, "clipping" will occur.
// Let's live with it for v1.
func multiplexSounds(output buffer, queue *list.List) {
	var finished []*list.Element

	multiplexed := make(buffer, bufferSize)
	for e := queue.Front(); e != nil; e = e.Next() {
		// Cast list element to queueItem.
		item := e.Value.(*queueItem)

		// Sum the magnitudes with the multiplier.
		for i, val := range item.contents[item.played] {
			multiplexed[i] += val
		}

		// If we have played all buffers, remove ourselves from the queue.
		item.played++
		if item.played >= len(item.contents) {
			finished = append(finished, e)
		}
	}
	for _, e := range finished {
		queue.Remove(e)
	}

	copy(output, multiplexed)
}
