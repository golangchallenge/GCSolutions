# Built Against
go1.4.2

# File Format
## PascalString format
[:1 Length as uint8] [:vary String as string]
## Track Format
[:4 ID as int32] [: Name as PascalString] [:16 Steps as [16]byte]
## Pattern Format
[:32 Version as string] [:4 Tempo as uint32] [:until-eof Tracks as []Track]
## Splice Format
[:13 "SPLICE"] [:1 filesize as byte] [: Pattern as Pattern]

# Suggestions:
## Splices are small:
  ### Problem:
    The splice format is limited to a pattern size of 256 bytes due to the filesize
    being liimted to a byte (I still think its a BigEndian int,
    but the rest of the file's encoding says otherwise).
  ### Solution:
    Shorten the length of the Magic bytes string (it only needs 6 characters,
    but I compensated for 13), cutting it down to 10 bytes, and then encoding the
    filesize as a uint32 to allow much larger patterns.

## Pattern tracks have fixed length steps:
  ### Problem:
    You only have 16 steps, no more, no less.

  ### Solution:
    With the pattern size solution, adding variable length patterns is simple
    as adding a entry alongside the tempo.

## Time Signatures
  ### Problem:
    4/4 beats only.

  ### Solution:
    As with the pattern steps size solution, adding time signatures can be
    done using 2 bytes.

## Patterns don't have names
  ### Problem:
    Title ^

  ### Solution:
    PascalString for the name of the pattern, problably some other metadata
    as well (such as Author, Date etc..)
