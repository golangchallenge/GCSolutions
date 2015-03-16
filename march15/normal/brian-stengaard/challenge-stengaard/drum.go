// Package drum implements the decoding of .splice drum machine files.
// See golang-challenge.com/go-challenge1/ for more information
//
// Little Endian is used in serialized files.
//
// The file format of a .splice file is as follows. Length is measured in
// bytes.
//
//     Name    Offset   Length    Type      Comment
//    -----------------------------------------------
//     magic   0x0      13       []byte   "SPLICE" padded with zeroes
//     n       0xd      1        uint8     Length of version + tempo + tracks
//     version 0xe      32       []byte
//     tempo   0x2e     4        float32
//     tracks  0x33     variable track
//
//
// The layout of a track is as follows.
//
//      Name     Offset       Length    Type
//     -----------------------------------------------
//      trackid  0x0          4         uin32
//      nameLen  0x4          1         uint8
//      name     0x5          nameLen   []byte
//      steps    0x5+nameLen  16        []byte
//
package drum
