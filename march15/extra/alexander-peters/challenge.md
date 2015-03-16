
See http://golang-challenge.com/go-challenge1/

#### The Go Challenge 1

This morning I took my daughter Giana to my secret lab to show her the various inventions I built over the years. That's when I realized that the awesome drum machine prototype I designed in the 90s had disappeared!!! The only related things I could find were printouts of the patterns I had created on the device as well as backups saved on floppy disks. Here are the printed patterns:

```
pattern_1.splice
Saved with HW Version: 0.808-alpha
Tempo: 120
(0) kick     |x---|x---|x---|x---|
(1) snare    |----|x---|----|x---|
(2) clap     |----|x-x-|----|----|
(3) hh-open  |--x-|--x-|x-x-|--x-|
(4) hh-close |x---|x---|----|x--x|
(5) cowbell  |----|----|--x-|----|
```

```
pattern_2.splice
Saved with HW Version: 0.808-alpha
Tempo: 98.4
(0) kick    |x---|----|x---|----|
(1) snare   |----|x---|----|x---|
(3) hh-open |--x-|--x-|x-x-|--x-|
(5) cowbell |----|----|x---|----|
```

```
pattern_3.splice
Saved with HW Version: 0.808-alpha
Tempo: 118
(40) kick    |x---|----|x---|----|
(1) clap     |----|x---|----|x---|
(3) hh-open  |--x-|--x-|x-x-|--x-|
(5) low-tom  |----|---x|----|----|
(12) mid-tom |----|----|x---|----|
(9) hi-tom   |----|----|-x--|----|
```

```
pattern_4.splice
Saved with HW Version: 0.909
Tempo: 240
(0) SubKick	      |----|----|----|----|
(1) Kick	      |x---|----|x---|----|
(99) Maracas	  |x-x-|x-x-|x-x-|x-x-|
(255) Low Conga	  |----|x---|----|x---|
```

```
pattern_5.splice
Saved with HW Version: 0.708-alpha
Tempo: 999
(1) Kick	|x---|----|x---|----|
(2) HiHat	|x-x-|x-x-|x-x-|x-x-|
```

I need your help to reverse engineer the binary format used by my drum machine and write a decoder so I will be able to implement a new drum machine, using Go this time!

#### To get started

You will find [attached a zip file](https://github.com/joshsoftware/golang-challenge/tree/gh-pages/data/ch1/golang-challenge-1-drum_machine.zip) containing the starting point for this challenge.
Your goal is to implement the `drum` package and make it pass the
provided tests.

#### Some information about my legacy drum machine

My drum machine loads an audio sample per track, allowing the programmer to schedule the playback of the sound. The scheduling of the playback is done using the concept of steps. A step is one of the parts of the measure that is being programmed (the programmed measure is known as a pattern). The measure (also called a bar) is divided in steps.

![measure.png](http://rubylearning.com/data/measure.png)

My drum machine only supports 16 step measure patterns played in 4/4 time. The measure is comprised of 4 quarter notes, each quarter note is comprised of 4 sixteenth notes and each sixteenth note corresponds to a step.
If all these music terms are confusing, don't worry, just know that the drum machine uses grid of 16 parts to let you trigger a sound. We have one sound per track and each track can be programmed independently. Each part is called a step. The speed of the playback is based on the tempo (aka bpm).

Taking an example from the printouts above:
```
(40) kick |x---|----|x---|----|
```

means that we have a track called "kick" (id 40) with the sound output being triggered on the first and ninth steps.

#### Goal

The goal of this challenge is to write a binary decoder that given a binary backup, outputs the same printouts as shown above. To do that you need to implement the `DecodeFile(path string) (*Pattern, error)` function inside the drum package. Note that `*Pattern` needs to be printable so it can be compared to the printouts. To help you, a test is provided. To run the test suite, use your terminal to navigate to the location of the unzip file and run `go test -v`. You should see an output similar to this:

```
go test -v
=== RUN TestDecodeFile
--- FAIL: TestDecodeFile (0.00s)
	decoder_test.go:69: decoded:
		"&{}"
	decoder_test.go:70: expected:
		"Saved with HW Version: 0.808-alpha\nTempo: 120\n(0) kick\t|x---|x---|x---|x---|\n(1) snare\t|----|x---|----|x---|\n(2) clap\t|----|x-x-|----|----|\n(3) hh-open\t|--x-|--x-|x-x-|--x-|\n(4) hh-close\t|x---|x---|----|x--x|\n(5) cowbell\t|----|----|--x-|----|\n"
	decoder_test.go:72: pattern_1.splice wasn't decoded as expect.
		Got:
		&{}
		Expected:
		Saved with HW Version: 0.808-alpha
		Tempo: 120
		(0) kick	|x---|x---|x---|x---|
		(1) snare	|----|x---|----|x---|
		(2) clap	|----|x-x-|----|----|
		(3) hh-open	|--x-|--x-|x-x-|--x-|
		(4) hh-close	|x---|x---|----|x--x|
		(5) cowbell	|----|----|--x-|----|
FAIL
exit status 1
```

#### Requirements

* Only use Go standard library. No third-party libraries may be imported.
* You are welcome to modify (improve) the included test suite or move the binaries, but the `DecodeFile` API must remain and the original test suite must still pass.

#### Hints

* Look around to see how data is usually [serialized/encoded](http://golang.org/pkg/encoding/json/).
* Become familiar with [encoding/binary package](http://golang.org/pkg/encoding/binary/), especially [binary.Read](http://golang.org/pkg/encoding/binary/#Read).
* [hex.Dump](http://golang.org/pkg/encoding/hex/#Dump) can very useful when debugging binary data (read more about [hex dump](http://en.wikipedia.org/wiki/Hex_dump))

#### I don't know where to start :(

<a href="/images/hex.png"><img src="/images/hex.png" width="627" alt="Hex Viewer" title="Hex Viewer" border=0></a>

The first step is to reverse engineer the binary file format. Look at the hex values to see if you can detect patterns. Binary data usually contains some sort of headers, then the encoded data. You should expect to find the data described in the printouts:

* version
* tempo
* tracks with each track containing
  * id
  * name
  * 16 steps

Then you need to write a decoder that takes one of the provided binary files and extracts/prints the data.

#### Go further (optional, not evaluated for the challenge)

This advanced section is not for the faint of heart, in case you were about to complain about how easy this challenge was, or if you just want to push things further, here is some more!

How about editing the cowbell track to add more steps? Reading the binary format is one thing, being able to generate/modify the data is even more fun. Take a pattern of your choosing and add more cowbell! For instance convert `pattern_2.splice` from:

```
pattern_2.splice
Saved with HW Version: 0.808-alpha
Tempo: 98.4
(0) kick    |x---|----|x---|----|
(1) snare   |----|x---|----|x---|
(3) hh-open |--x-|--x-|x-x-|--x-|
(5) cowbell |----|----|x---|----|
```

to:

```
pattern_2-morebells.splice
Saved with HW Version: 0.808-alpha
Tempo: 98.4
(0) kick    |x---|----|x---|----|
(1) snare   |----|x---|----|x---|
(3) hh-open |--x-|--x-|x-x-|--x-|
(5) cowbell |x---|x-x-|x---|x-x-|
```

Still not enough? Why not implementing the playback of the patterns
using something like [portaudio](https://godoc.org/code.google.com/p/portaudio-go/portaudio)?

---
