package drum

// This file defines some constants specific to the drum package

// The begining identifier string of a splice file
var SpliceFileHeader = "SPLICE"

// The length of the SPLICE file header at the begining of each file in bytes
var SpliceFileSize = len(SpliceFileHeader)

// The length of the body of the splice file in bytes
var FileSize = 8

// The length of the version string of the splice file in bytes
var VersionSize = 32

// The length of the encoded Tempo inside the splice file in bytes
var TempoSize = 4

// The length of a Track ID inside the splice file in bytes
var TrackIDSize = 1

// The length of a Track Name Size inside of the splice file in bytes
var TrackNameSize = 4

// The length of a Step Sequence inside the splice file in bytes
var StepSequenceSize = 16

// The full length of a single track inside the splice file in bytes
var TrackSize = TrackIDSize + TrackNameSize + StepSequenceSize

// A string representation of an Empty Byte.
// Used to trim the padding from the Version string.
var EmptyByteString = "\x00"
