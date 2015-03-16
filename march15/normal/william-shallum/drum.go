/* 
Package drum is supposed to implement the decoding of .splice drum machine files.
See golang-challenge.com/go-challenge1/ for more information

The .splice file consists of a header followed by multiple tracks. The header
format is illustrated below:

 +---------------------------------------------------+
 |   | 0| 1| 2| 3| 4| 5| 6| 7| 8| 9| A| B| C| D| E| F| 
 +---+-----------------------------------------------+
 |00 | S  P  L  I  C  E|  6 0x00 bytes   | len | hw- |
 +---+-----------------------------------------------+
 |10 | version (32 bytes)....                        |
 +---+-----------------------------------------------+
 |20 | hw-version (cont...)                    | tem-|
 +---+-----------------------------------------------+
 |30 | po  |  ...tracks...                           |
 +---+-----------------------------------------------+

The first six bytes of the header are the ASCII characters "SPLICE" followed by
six nul bytes. It is followed by an unsigned 16-bit big-endian length. This
length indicates the length of the data that follows (this length includes the
length indicator itself).

XXX There is no evidence that the length is 16-bit big-endian. It is possible
that it is a 13-byte header followed by an 8-bit length.

The hw-version is a text string right-padded with nul bytes to 32 bytes. The
value of hw-version does not appear to affect the format of the file. The
hw-version is followed by the tempo which is a little-endian IEEE-754 32-bit
float.

The header is followed by multiple tracks one immediately after the other. The
track format is as follows:

 +-----------------------------------------------+
 | 0| 1| 2| 3| 4| nl bytes  | 16 bytes           |
 +-----------------------------------------------+
 | track-id  |nl| track-name| steps              |
 +-----------------------------------------------+

An unsigned 32-bit little-endian track-id is followed by an 8-bit unsigned
integer "nl" specifying the number of bytes in the track name, then the "nl"
bytes of the track name itself. The track name is followed by 16 bytes of the
actual steps where the byte is non-zero if a sound is triggered at that step
and zero otherwise.

A track is immediately followed by another track if it exists. There is no
indication of the final track. Refer to the length in the header to know where
the data ends.

*/
package drum
