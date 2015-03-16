# SPLICE

The .splice file for decoding the drum machine patterns, is
interpreted as follows:

## Introduction

First, there comes a header-part which is in all examples of fixed
size. This is not strictly necessary though, as there is an option to
not start the interpretation of a drum pattern until four NUL-bytes in
a row has been read. The current implementation of the drum packet,
however, assumes the header to be of this fixed size.

## Header

Data is interpreted in the listed order from the table below.

>| Type     | Size | Example Representation/Value        | Discussion                                               | Endianess |
>|----------+------+-------------------------------------+----------------------------------------------------------+-----------|
>| [6]byte  |    6 | 0x[53 50 4c 49 43 45]               | States 'SPLICE' in ascii/utf-8.                          |           |
>| uint32   |    4 | 0x[00 00 00 00]                     | Might be some sort of flag / unused.                     | big       |
>| uint32   |    4 | 197                                 | Remaining bytes of the file from this point of the file. | big       |
>| [32]byte |   32 | 0x[30 2e 38 30 38 2d 61 6c 70 68..] | Contains the version string.                             |           |
>| float32  |    4 | 0x42f00000                          | The tempo as an IEEE 754 32-bit float.                   | little    |


## Tracks

For all the remaining bytes, we interpret several tracks as follows.

>| Type     |      Size | Example Representation/Value | Discussion                                  | Endianess |
>|----------+-----------+------------------------------+---------------------------------------------+-----------|
>| int32    |         4 | 0x[07 00 00 00]              | The track id.                               | little    |
>| byte     |         1 | 7                            | Strlen of the track name.                   |           |
>| []byte   | see above | "snare"                      | The track name. Maxlen = 0xff = 255         |           |
>| [16]byte |        16 | 0x[01 00 01 00, ... ]        | The 16 track steps - 0x01 = on, 0x00 = off. |           |
