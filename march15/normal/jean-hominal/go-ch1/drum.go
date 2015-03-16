// Package drum is supposed to implement the decoding of .splice drum machine files.
// See golang-challenge.com/go-challenge1/ for more information
package drum

// Format:
// SPLICE header (6 bytes)
// 8 bytes - frame size, big-endian
// 32 bytes for hardware version (ASCII)
// 4 bytes little-endian - tempo as float32
// tracks:
//   track index (4 bytes, little-endian)
//   name length (1 byte)
//   track name (ASCII)
//   track contents (01 if x, 00 otherwise)
