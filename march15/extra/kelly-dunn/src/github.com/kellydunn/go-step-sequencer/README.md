# go-step-sequencer

A step sequencer implemented in Golang using portaudio and libsndfile wrappers.

## Dependencies

This project requires two dependent libraries, both of which are thinly wrapped by a golang libraries.  In order for the project to successfully compile, you will need to install the development libraries of both portaudio and libsndfile for your platform.  To learn more about how to install the native libraries of these dependencies, please visit their official sites:

  - [portaudio](http://www.portaudio.com/)
  - [libsndfile](http://mega-nerd.com/libsndfile/)

## Build

This project currently has a [Makefile](./Makefile), which not only builds to go binary, but also runs the go linter and the go vet tool.

```bash
$ make
```

Upon a successful build, the binary will exist in the root level directory of this project.

## Usage

If you haven't built the tool as indicated above, you should be install the utility using the go tool:

```bash
$ go get github.com/kellydunn/go-step-sequencer
$ go install github.com/kellydunn/go-step-sequencer
```

Now you should be able to use the `go-step-sequencer` utility on the command line.  Use the `--help` flag to get more information on how to use the utility.

```bash
$ go-step-sequencer --help
Usage of go-step-sequencer:
  -kit="kits": -kit=path/to/kits
  -pattern="patterns/pattern_1.splice": -pattern=path/to/pattern.splice
```

The step sequencer was made to take a `pattern` and a `kit` as command line flags so that you can swap out different types of kits and patterns.  A typical use of the command looks like this:

```bash
$ go-step-sequencer --pattern path/to/pattern.splice --kit path/to/kits
```

The default pattern is found at `patterns/pattern_1.splice` and the default kit is located at `kits/0.808-alpha`.  Running `go-step-sequnencer` without specifying a `--pattern` or a `--kit` will run the default pattern with the default kit:   

```bash
$ go-step-sequencer
loaded sample: kits/0.808-alpha/kick.wav
loaded sample: kits/0.808-alpha/snare.wav
loaded sample: kits/0.808-alpha/clap.wav
loaded sample: kits/0.808-alpha/hh-open.wav
loaded sample: kits/0.808-alpha/hh-close.wav
loaded sample: kits/0.808-alpha/cowbell.wav
Saved with HW Version: 0.808-alpha
Tempo: 120
(0) kick        |x---|x---|x---|x---|
(1) snare       |----|x---|----|x---|
(2) clap        |----|x-x-|----|----|
(3) hh-open     |--x-|--x-|x-x-|--x-|
(4) hh-close    |x---|x---|----|x--x|
(5) cowbell     |----|----|--x-|----|
```

You should be able to hear the drum track out of your speakers now! 