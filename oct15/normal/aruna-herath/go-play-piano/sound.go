package main

import (
	"fmt"
	"math"

	"golang.org/x/mobile/exp/audio/al"
)

var data []byte
var buffers []al.Buffer
var secBuffers []al.Buffer
var sources []al.Source
var secSources []al.Source

// InitializeSound Initializes sound
func InitializeSound(numKeys int) {
	var i float64
	data = []byte{}
	for i = 0; i < 100000; i = i + 8 {
		data = append(data, byte(128+127*math.Sin(i)))
	}

	err := al.OpenDevice()
	if err != nil {
		fmt.Println(err)
	}

	sources = al.GenSources(numKeys)
	buffers = al.GenBuffers(numKeys)
	secSources = al.GenSources(numKeys)
	secBuffers = al.GenBuffers(numKeys)

	for j := 0; j < numKeys; j++ {
		buffers[j].BufferData(al.FormatMono8, data, int32(100*(numKeys-j)))
		secBuffers[j].BufferData(al.FormatMono8, data, 2*int32(100*(numKeys-j)))
		sources[j].QueueBuffers(buffers[j : j+1])
		secSources[j].QueueBuffers(buffers[j : j+1])
	}

}

// PlaySound plays the sound of the source i
func PlaySound(i int) {
	al.PlaySources(sources[i])
}

// StopSound stops the sound playing through the source i
func StopSound(i int) {
	al.StopSources(sources[i])
}

// ReleaseSound releases allocated buffers
func ReleaseSound() {

}
