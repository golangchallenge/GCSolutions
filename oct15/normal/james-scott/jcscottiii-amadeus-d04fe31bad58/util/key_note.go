package util

// KeyNote is a type to denote numerically which key on the piano it is.
// This number type is used for generating the sound waves.
type KeyNote int

// Enumeration of all the keys we will display. Numbering starts with their keys on a real piano.
const (
	B3 KeyNote = iota + 39
	C4
	CSharp4
	D4
	DSharp4
	E4
	F4
	FSharp4
	G4
	GSharp4
	A4
	ASharp4
	B4
	C5
)

const (
	// FirstKey holds which KeyNote is the first one. Great for iterating over notes.
	FirstKey = B3
	// LastKey holds which KeyNote is the last one. Great for iterating over notes.
	LastKey = C5
)

// blackKeyList contains which keyNotes should be marked as black keys.
var blackKeyList = []KeyNote{CSharp4, DSharp4, FSharp4, GSharp4, ASharp4}

// IsBlackKey checks if the input KeyNote has been marked as a blackkey.
// Helper function to distinguish between doing things for white keys vs. black keys.
func IsBlackKey(note KeyNote) bool {
	for _, blackKeyNote := range blackKeyList {
		if note == blackKeyNote {
			return true
		}
	}
	return false
}
