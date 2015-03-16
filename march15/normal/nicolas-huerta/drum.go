// Package drum is supposed to implement the decoding of .splice drum machine files.
// See golang-challenge.com/go-challenge1/ for more information
package drum

// Constants for max length of data format and version strings
const kMaxDataFormatLength = 10
const kMaxVersionLength = 32
const kTrackPatternLength = 16

// sizeof constants, only used in this file for defining offsets
// binary.Size() is used in the decoder for decoding the binary data
const kSizeOfInt32 = 4
const kSizeOfFloat32 = 4

// Data offsets. Offsets are used in the decoder to locate the beginning of each data slice.
const kDataSizeOffset = kMaxDataFormatLength
const kVersionOffset = kDataSizeOffset + kSizeOfInt32
const kTempoOffset = kVersionOffset + kMaxVersionLength
const kTrackDataOffset = kTempoOffset + kSizeOfFloat32
