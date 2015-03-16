// Package drum provides a way to represent tracks in a drum machine using a
// binary file format.
//
// The drum machine plays samples of sounds, called splices, where each splice
// is one measure composed of 16 steps.  A splice file is encoded using
// LittleEndian byte ordering and containes the following:
//
//   13 byte header: Must be the word SPLICE followed by 7 null bytes (0x00)
//
//    1 byte length: length is the length of the version string (32 bytes) plus
//    the tempo (4 bytes) plus the length of all the combined tracks
//
//  32 byte version: The version string followed by null bytes (0x00)
//
//    4 byte tempo: float32
//
//  variable length tracks
//
// A track contains an id, name and sequence of steps in the following format:
//
//   4 byte id
//   1 byte name length
//   variable length name
//   16 bytes for steps
//
// Each step can activate the instrument represented by the track.  Valid values
// for steps are 0x00 for no sound and 0x01 to output the sound.
package drum
