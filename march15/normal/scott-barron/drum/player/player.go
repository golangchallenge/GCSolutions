package main

import (
	"code.google.com/p/portaudio-go/portaudio"
	"flag"
	"github.com/rubyist/drum"
	"log"
	"time"
)

var soundDir = flag.String("d", "sounds", "directory containing samples")

func main() {
	flag.Parse()

	sequencer := NewSequencer()

	for _, file := range flag.Args() {
		pattern, err := drum.DecodeFile(file)
		if err != nil {
			log.Fatal(err)
		}
		sequencer.Add(pattern)
		log.Print(pattern.String())
	}

	portaudio.Initialize()
	defer portaudio.Terminate()
	stream, err := portaudio.OpenDefaultStream(0, 2, 44100, 0, func(o []int32) {
		sequencer.Read(o)
	})
	if err != nil {
		log.Fatal(err)
	}
	defer stream.Close()
	stream.Start()
	defer stream.Stop()

	sequencer.Start()

	for {
		time.Sleep(time.Second)
	}
}
