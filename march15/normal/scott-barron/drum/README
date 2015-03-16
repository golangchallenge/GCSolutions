Go Challenge #1

This is a toy drum sequencer submitted for the first Go Challenge.

## Challenge Goal

The challenge goal is to write a decoder for the drum pattern files
in the fixtures directory and get the tests to pass. This is the code
in decoder.go.

## Extras

The encoder.go file contains an encoder to generate pattern files
from a Pattern structure.

The player directory contains a program that uses libsndfile and portaudio
to play pattern files using samples in wav format. It's a little rough around
the edges because I had it in my head this was due on March 14, so there's
a little more work that can be done!

The player can be built with `go build`, though it is expecting to find the
drum package under my path.

The player expects to find wav files in a directory passed through the -d option.
It expects the wave files to be named "x.wav" where x is the track.Name. I've
included some test patterns and samples. The samples are licensed under creative
commons, found on ccmixter[1] (you can omit these from the GitHub repository).

An example run from the player directory:

`$ ./player -d sounds/ test.splice test2.splice`

This will sequence and loop test.splice then test2.splice forever.

This challenge was a lot of fun. While I have done some reverse engineering
of binary formats before, I've never really done any kind of audio programming.
I look forward to the future challenges!

[1] http://ccmixter.org/files/CarbonMonoxideMusic/23425
