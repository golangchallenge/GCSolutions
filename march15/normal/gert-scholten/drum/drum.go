// Package drum is implements the decoding of .splice drum machine files.
package drum

// Number of steps per track
const Steps = 16

// Pattern is the high level representation of the
// drum pattern contained in a .splice file.
type Pattern struct {
	Version string
	Tempo   float32
	Tracks  []Track
}

// Track data associated with the pattern
type Track struct {
	ID   uint8
	Name string
	// 16 steps, one byte per step, value of 0 or 1 (non-zero)
	Data [Steps]byte
}
