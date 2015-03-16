// The playdrums command can be used to play splice files
// as read by the drum package.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"code.google.com/p/portaudio-go/portaudio"
	"github.com/nf/sigourney/audio"
	"github.com/rakyll/portmidi"
	"github.com/unixpickle/wav"

	"github.com/rogpeppe/misc/drum"
	"github.com/rogpeppe/misc/drum/drummachine"
)

var (
	sampleDir = flag.String("dir", "./sample", "directory containing sample files")
	showInfo  = flag.Bool("info", false, "show splice file info (do not play pattern)")
)

const usage = `usage: playdrums [flags] <file>

The playdrums command plays the drum machine pattern found
in the given file. To find a sample file, the track names from the
drum machine pattern are first lower-cased and all spaces and hyphens
removed.
`

func main() {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, usage)
		flag.PrintDefaults()
		os.Exit(2)
	}
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
	}
	p, err := drum.DecodeFile(flag.Arg(0))
	if err != nil {
		log.Fatalf("cannot decode splice file: %v", err)
	}
	if *showInfo {
		fmt.Print(p)
		return
	}
	patches, err := readPatches(p, *sampleDir)
	if err != nil {
		log.Fatalf("cannot read sample patches: %v", err)
	}
	drumMod, err := drummachine.New(p, patches)
	if err != nil {
		log.Fatalf("cannot make new drum machine: %v", err)
	}

	portaudio.Initialize()
	defer portaudio.Terminate()

	portmidi.Initialize()
	defer portmidi.Terminate()

	e := audio.NewEngine()
	e.Input("in", drumMod)
	if err := e.Start(); err != nil {
		panic(err)
	}
	select {}
}

func readPatches(p *drum.Pattern, dir string) (map[string][]audio.Sample, error) {
	m := make(map[string][]audio.Sample)
	for _, tr := range p.Tracks {
		path := filepath.Join(dir, sampleFileName(tr.Name)+".wav")
		samples, err := readWav(path)
		if err != nil {
			return nil, fmt.Errorf("cannot read %s: %v", path, err)
		}
		log.Printf("successfully read %s", tr.Name)
		m[tr.Name] = samples
	}
	return m, nil
}

func sampleFileName(name string) string {
	name = strings.ToLower(name)
	name = strings.Replace(name, " ", "", -1)
	name = strings.Replace(name, "-", "", -1)
	return name
}

// readWav reads the wav file at the given path and returns it as a set
// of audio samples.
// TODO be much less picky about what kinds
// of file we accept.
func readWav(path string) ([]audio.Sample, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cannot open file: %v", err)
	}
	defer f.Close()
	r, err := wav.NewReader(f)
	if err != nil {
		return nil, fmt.Errorf("cannot make reader: %v", err)
	}
	h := r.Header().Format
	if h.NumChannels != 1 {
		return nil, fmt.Errorf("want 1 channel, got %d", h.NumChannels)
	}
	if h.SampleRate != drummachine.SampleRate {
		return nil, fmt.Errorf("want sample rate %d, got %d", drummachine.SampleRate, h.SampleRate)
	}
	samples := make([]wav.Sample, r.Remaining())
	n, err := r.Read(samples)
	if n != len(samples) && err != nil {
		return nil, fmt.Errorf("cannot read samples (%d): %v", n, err)
	}
	if n != len(samples) {
		return nil, fmt.Errorf("only read %d/%d samples", n, len(samples))
	}
	auSamples := make([]audio.Sample, len(samples))
	for i, s := range samples {
		auSamples[i] = audio.Sample(s)
	}
	return auSamples, nil
}
