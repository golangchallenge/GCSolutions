package drum

// lSteps is an integer specifying the number of steps per track
const lSteps int = 16

// pVerEnd represents the byte position where the version info ends
const pVerEnd int = 46

// pRead holds the index of the byte having the number of
// data bytes in the binary file
const pRead int = 13

// lHeader stores the number of bytes in which the splice header is stored
const lHeader int = 14

// pTrackStart stores the position for the track information starting
const pTrackStart int = 50

// pTempo is the byte position where the tempo float32 starts
const pTempo int = 46
