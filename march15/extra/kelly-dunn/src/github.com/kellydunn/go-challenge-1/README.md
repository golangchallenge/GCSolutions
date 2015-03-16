OB# go-challenge-1

This is my implementation for the [March 2015 golang challenge](http://golang-challenge.com/go-challenge1/).

## File Protocol

Part of the challenge is the reverse engineer the `.splice` file protocol.  In this section I talk about the comoponents of the file protocol that I reverse engineered, and how they work together to create the `.splice` file format.

  - [Header](#header)
    - ["SPLICE" File Type](#splice_file_body_type)
    - [Splice File Body Size](#splice_file_body_size)
  - [Pattern](#pattern)
    - [Pattern Version String](#pattern_version_string)
    - [Pattern Tempo](#pattern_tempo)
  - [Tracks](#tracks)
    - [Track Id](#track_id)
    - [Track Name](#track_name)
    - [Track Step Sequence](#track_step_sequence)

<a name="header"></a>
### Header

The first section of any `.splice` file is the header sectoin.  It appears to contain a File Type and a File Body Size.

<a name="splice_file_body_type"></a>
#### "SPLICE" File Type

The Splice file type appears to be fairly simple: a 6-byte string with a static value of "SPLICE".  This identifies the file as a `.splice` file.

<a name="splice_file_body_size"></a>
#### Splice File Body Size

The Splice File Body Size appears to be an unsigned 64-bit integer that describes the length of the Body of the `.splice` file in bytes. It should be noted that `pattern_5.splice` is the only example file that has extra, erroneous data encoded in it that occurs after the aforementioned Splice File Body Size.  Since the tests that came with the challenge seem to require this file to be successfully decoded as a Pattern, I have decided to discard any data in the file that occurs after the Splice File Body Size.

<a name="pattern"></a>
### Pattern

The First Part of the body seems to encode the Pattern Version followed by the Pattern Tempo.

<a name="pattern_version_string"></a>
#### Pattern Version String

The Pattern Version String seems to be a 32-bytes in Length.  The unused bytes also seem to be discarded in order display the version string correctly.

<a name="pattern_tempo"></a>
#### Pattern Tempo

The Pattern Tempo sems to be encoded next in the file as a floating 32-bit number.  

<a name="tracks"></a>
### Tracks

The next substantial section of the document appears to be all of the Tracks in sequence.

<a name="track_id"></a>
#### Track Id

The Track Id appears to be a unsigned 8-bit integer.

<a name="track_name"></a>
#### Track Name

The Track Name is comprised of two parts encoded in the file.  First, appears to be an unsigned 32-bit integer which describes the size of the track name in bytes.  This value is the number of bytes in which to read off of the file in order to receive the name of the track.

<a name="track_step_sequence"></a>
#### Track Step Sequence

The final component of each Track looks to be a series of 16 bytes, each of which describes whether or not the corresponding track should be driggered for a given 16th note in the step sequencer.

