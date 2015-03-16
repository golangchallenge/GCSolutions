# Splice Legacy Files Decoder

This is a small decoder which decodes binary splice files into a more readable form. Its development structure & usage is pretty simple

## Testing

to run the test suite enter the main drum directory and run the test command

```bash
$ cd drum
$ go test -v
```

## Directory Structure

The Directory Structure is pretty simple

```
.
├── drum                # Contains the main drum module
│   ├── decoder.go      # Implements the decoding and main functionality
│   ├── decoder_test.go # The main test suite
│   └── drum.go
├── fixtures            # Contains 5 binary splice files used for testing
│   ├── pattern_1.splice
│   ├── pattern_2.splice
│   ├── pattern_3.splice
│   ├── pattern_4.splice
│   └── pattern_5.splice
├── main.go             # Contains the code of the command line utility
└── README.md           # This File :)

```

## Usage

The following is the usage taken from the help of the command line utility

```
Usage: go run main.go [-decode, -help] filename.splice

  -cowbell=0: adds more cowbell beats (if there is a cowbell instrument). Note: use with decode
  -decode=false: decode a splice legacy file and prints the output
  -help=false: displays this small help
```

## Examples

```bash
$ go run main.go -decode -cowbell=3 fixtures/pattern_1.splice
Saved with HW Version: 0.808-alpha
Tempo: 120
(0) kick    |x---|x---|x---|x---|
(1) snare   |----|x---|----|x---|
(2) clap    |----|x-x-|----|----|
(3) hh-open |--x-|--x-|x-x-|--x-|
(4) hh-close    |x---|x---|----|x--x|
(5) cowbell |--x-|-x--|--x-|---x|
```

```bash
$ go run main.go -decode fixtures/pattern_2.splice
Saved with HW Version: 0.808-alpha
Tempo: 98.4
(0) kick    |x---|----|x---|----|
(1) snare   |----|x---|----|x---|
(3) hh-open |--x-|--x-|x-x-|--x-|
(5) cowbell |----|----|x---|----|
```
